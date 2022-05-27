/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

import (
	"context"

	"github.com/google/go-cmp/cmp"
	container "google.golang.org/api/container/v1"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gcp/apis/container/v1beta2"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	gke "github.com/crossplane-contrib/provider-gcp/pkg/clients/cluster"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
)

// Error strings.
const (
	errNewClient            = "cannot create new GKE container client"
	errManagedUpdateFailed  = "cannot update Cluster custom resource"
	errNotCluster           = "managed resource is not a Cluster"
	errGetCluster           = "cannot get GKE cluster"
	errCreateCluster        = "cannot create GKE cluster"
	errUpdateCluster        = "cannot update GKE cluster"
	errDeleteCluster        = "cannot delete GKE cluster"
	errCheckClusterUpToDate = "cannot determine if GKE cluster is up to date"
)

// SetupCluster adds a controller that reconciles Cluster
// managed resources.
func SetupCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta2.ClusterGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta2.ClusterGroupVersionKind),
		managed.WithExternalConnecter(&clusterConnector{kube: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta2.Cluster{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type clusterConnector struct {
	kube client.Client
}

func (c *clusterConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := container.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &clusterExternal{cluster: s, projectID: projectID, kube: c.kube}, errors.Wrap(err, errNewClient)
}

type clusterExternal struct {
	kube      client.Client
	cluster   *container.Service
	projectID string
}

func (e *clusterExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta2.Cluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCluster)
	}

	existing, err := e.cluster.Projects.Locations.Clusters.Get(gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetCluster)
	}

	cr.Status.AtProvider = gke.GenerateObservation(*existing)
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	gke.LateInitializeSpec(&cr.Spec.ForProvider, *existing)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	switch cr.Status.AtProvider.Status {
	case v1beta2.ClusterStateRunning, v1beta2.ClusterStateReconciling:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta2.ClusterStateProvisioning:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta2.ClusterStateUnspecified, v1beta2.ClusterStateDegraded, v1beta2.ClusterStateError:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	u, _, err := gke.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, existing)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckClusterUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  u,
		ConnectionDetails: connectionDetails(existing),
	}, nil
}

func (e *clusterExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta2.Cluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCluster)
	}
	cr.SetConditions(xpv1.Creating())

	// Wait until creation is complete if already provisioning.
	if cr.Status.AtProvider.Status == v1beta2.ClusterStateProvisioning {
		return managed.ExternalCreation{}, nil
	}

	// Generate GKE cluster from resource spec.
	cluster := &container.Cluster{}
	gke.GenerateCluster(meta.GetExternalName(cr), cr.Spec.ForProvider, cluster)

	// When autopilot is enabled, node pools cannot be specified.
	if cluster.Autopilot == nil || !cluster.Autopilot.Enabled {
		// Insert default node pool for bootstrapping cluster. This is required
		// to create a GKE cluster. After successful creation we delete the
		// bootstrap node pool immediately and provision any subsequent node
		// pools using the NodePool resource type.
		gke.AddNodePoolForCreate(cluster)
	}

	create := &container.CreateClusterRequest{
		Cluster: cluster,
	}

	_, err := e.cluster.Projects.Locations.Clusters.Create(gke.GetFullyQualifiedParent(e.projectID, cr.Spec.ForProvider), create).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateCluster)
}

func (e *clusterExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta2.Cluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCluster)
	}
	// Do not issue another update until the cluster finishes the previous one.
	if cr.Status.AtProvider.Status == v1beta2.ClusterStateReconciling || cr.Status.AtProvider.Status == v1beta2.ClusterStateProvisioning {
		return managed.ExternalUpdate{}, nil
	}
	// We have to get the cluster again here to determine how to update.
	existing, err := e.cluster.Projects.Locations.Clusters.Get(gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetCluster)
	}

	u, fn, err := gke.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, existing)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckClusterUpToDate)
	}
	if u {
		return managed.ExternalUpdate{}, nil
	}

	// GKE uses different update methods depending on the field that is being
	// changed. gke.IsUpToDate returns the appropriate update operation based on
	// the difference in the desired and existing spec. Only one field can be
	// updated at a time, so if there are multiple diffs, the next one will be
	// handled after the current one is completed.
	_, err = fn(ctx, e.cluster, gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr)))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateCluster)
}

func (e *clusterExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta2.Cluster)
	if !ok {
		return errors.New(errNotCluster)
	}
	cr.SetConditions(xpv1.Deleting())
	// Wait until delete is complete if already deleting.
	if cr.Status.AtProvider.Status == v1beta2.ClusterStateStopping {
		return nil
	}

	_, err := e.cluster.Projects.Locations.Clusters.Delete(gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteCluster)
}

// connectionSecret return secret object for cluster instance
func connectionDetails(cluster *container.Cluster) managed.ConnectionDetails {
	config, err := gke.GenerateClientConfig(cluster)
	if err != nil {
		return nil
	}
	rawConfig, err := clientcmd.Write(config)
	if err != nil {
		return nil
	}
	cd := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey:   []byte(config.Clusters[cluster.Name].Server),
		xpv1.ResourceCredentialsSecretUserKey:       []byte(config.AuthInfos[cluster.Name].Username),
		xpv1.ResourceCredentialsSecretPasswordKey:   []byte(config.AuthInfos[cluster.Name].Password),
		xpv1.ResourceCredentialsSecretCAKey:         config.Clusters[cluster.Name].CertificateAuthorityData,
		xpv1.ResourceCredentialsSecretClientCertKey: config.AuthInfos[cluster.Name].ClientCertificateData,
		xpv1.ResourceCredentialsSecretClientKeyKey:  config.AuthInfos[cluster.Name].ClientKeyData,
		xpv1.ResourceCredentialsSecretKubeconfigKey: rawConfig,
	}
	return cd
}
