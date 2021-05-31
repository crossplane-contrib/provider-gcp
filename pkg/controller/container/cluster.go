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
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/container/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	gke "github.com/crossplane/provider-gcp/pkg/clients/cluster"
)

// Error strings.
const (
	errNewClient            = "cannot create new GKE container client"
	errManagedUpdateFailed  = "cannot update GKECluster custom resource"
	errNotCluster           = "managed resource is not a GKECluster"
	errGetCluster           = "cannot get GKE cluster"
	errCreateCluster        = "cannot create GKE cluster"
	errUpdateCluster        = "cannot update GKE cluster"
	errDeleteCluster        = "cannot delete GKE cluster"
	errCheckClusterUpToDate = "cannot determine if GKE cluster is up to date"
)

// SetupGKECluster adds a controller that reconciles GKECluster
// managed resources.
func SetupGKECluster(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1beta1.GKEClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1beta1.GKECluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.GKEClusterGroupVersionKind),
			managed.WithExternalConnecter(&clusterConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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
	cr, ok := mg.(*v1beta1.GKECluster)
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
	case v1beta1.ClusterStateRunning, v1beta1.ClusterStateReconciling:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta1.ClusterStateProvisioning:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta1.ClusterStateUnspecified, v1beta1.ClusterStateDegraded, v1beta1.ClusterStateError:
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
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCluster)
	}
	cr.SetConditions(xpv1.Creating())

	// Wait until creation is complete if already provisioning.
	if cr.Status.AtProvider.Status == v1beta1.ClusterStateProvisioning {
		return managed.ExternalCreation{}, nil
	}

	// Generate GKE cluster from resource spec.
	cluster := &container.Cluster{}
	gke.GenerateCluster(meta.GetExternalName(cr), cr.Spec.ForProvider, cluster)

	// Insert default node pool for bootstrapping cluster. This is required to
	// create a GKE cluster. After successful creation we delete the bootstrap
	// node pool immediately and provision any subsequent node pools using the
	// NodePool resource type.
	gke.AddNodePoolForCreate(cluster)

	create := &container.CreateClusterRequest{
		Cluster: cluster,
	}

	_, err := e.cluster.Projects.Locations.Clusters.Create(gke.GetFullyQualifiedParent(e.projectID, cr.Spec.ForProvider), create).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateCluster)
}

func (e *clusterExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCluster)
	}
	// Do not issue another update until the cluster finishes the previous one.
	if cr.Status.AtProvider.Status == v1beta1.ClusterStateReconciling || cr.Status.AtProvider.Status == v1beta1.ClusterStateProvisioning {
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
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return errors.New(errNotCluster)
	}
	cr.SetConditions(xpv1.Deleting())
	// Wait until delete is complete if already deleting.
	if cr.Status.AtProvider.Status == v1beta1.ClusterStateStopping {
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
