/*
Copyright 2020 The Crossplane Authors.

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
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1beta1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/container/v1beta1"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	gke "github.com/crossplaneio/stack-gcp/pkg/clients/cluster"
)

// Error strings.
const (
	errGetProvider         = "cannot get Provider"
	errGetProviderSecret   = "cannot get Provider Secret"
	errNewClient           = "cannot create new GKE container client"
	errManagedUpdateFailed = "cannot update GKECluster custom resource"
	errNotCluster          = "managed resource is not a GKECluster"
	errGetCluster          = "cannot get GKE cluster"
	errCreateCluster       = "cannot create GKE cluster"
	errUpdateCluster       = "cannot update GKE cluster"
	errDeleteCluster       = "cannot delete GKE cluster"
)

// GKEClusterController is responsible for adding the GKECluster
// controller and its corresponding reconciler to the manager with any runtime configuration.
type GKEClusterController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func (c *GKEClusterController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1beta1.GKEClusterGroupVersionKind),
		resource.WithExternalConnecter(&clusterConnector{kube: mgr.GetClient(), newServiceFn: container.NewService}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.GKEClusterKindAPIVersion, v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.GKECluster{}).
		Complete(r)
}

type clusterConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*container.Service, error)
}

func (c *clusterConnector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	i, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return nil, errors.New(errNotCluster)
	}

	p := &gcpv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(i.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	client, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(s.Data[p.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(container.CloudPlatformScope))
	return &clusterExternal{cluster: client, projectID: p.Spec.ProjectID, kube: c.kube}, errors.Wrap(err, errNewClient)
}

type clusterExternal struct {
	kube      client.Client
	cluster   *container.Service
	projectID string
}

func (e *clusterExternal) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotCluster)
	}

	existing, err := e.cluster.Projects.Locations.Clusters.Get(gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetCluster)
	}

	cr.Status.AtProvider = gke.GenerateObservation(*existing)
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	gke.LateInitializeSpec(&cr.Spec.ForProvider, *existing)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	switch cr.Status.AtProvider.Status {
	case v1beta1.ClusterStateRunning, v1beta1.ClusterStateReconciling:
		cr.Status.SetConditions(v1alpha1.Available())
		resource.SetBindable(cr)
	case v1beta1.ClusterStateProvisioning:
		cr.Status.SetConditions(v1alpha1.Creating())
	case v1beta1.ClusterStateUnspecified, v1beta1.ClusterStateDegraded, v1beta1.ClusterStateError:
		cr.Status.SetConditions(v1alpha1.Unavailable())
	}

	u, _ := gke.IsUpToDate(&cr.Spec.ForProvider, *existing)

	return resource.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  u,
		ConnectionDetails: connectionDetails(existing),
	}, nil
}

func (e *clusterExternal) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotCluster)
	}
	cr.SetConditions(v1alpha1.Creating())

	// Wait until creation is complete if already provisioning.
	if cr.Status.AtProvider.Status == v1beta1.ClusterStateProvisioning {
		return resource.ExternalCreation{}, nil
	}

	// Generate GKE cluster from resource spec.
	cluster := gke.GenerateCluster(cr.Spec.ForProvider, meta.GetExternalName(cr))

	// Insert default node pool for bootstrapping cluster. This is required to
	// create a GKE cluster. After successful creation we delete the bootstrap
	// node pool immediately and provision any subsequent node pools using the
	// NodePool resource type.
	gke.AddNodePoolForCreate(cluster)

	create := &container.CreateClusterRequest{
		Cluster: cluster,
	}

	if _, err := e.cluster.Projects.Locations.Clusters.Create(gke.GetFullyQualifiedParent(e.projectID, cr.Spec.ForProvider), create).Context(ctx).Do(); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errCreateCluster)
	}

	return resource.ExternalCreation{}, nil
}

func (e *clusterExternal) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotCluster)
	}
	// Do not issue another update until the cluster finishes the previous one.
	if cr.Status.AtProvider.Status == v1beta1.ClusterStateReconciling || cr.Status.AtProvider.Status == v1beta1.ClusterStateProvisioning {
		return resource.ExternalUpdate{}, nil
	}
	// We have to get the cluster again here to determine how to update.
	existing, err := e.cluster.Projects.Locations.Clusters.Get(gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return resource.ExternalUpdate{}, errors.Wrap(err, errGetCluster)
	}

	u, fn := gke.IsUpToDate(&cr.Spec.ForProvider, *existing)
	if u {
		return resource.ExternalUpdate{}, nil
	}

	// GKE uses different update methods depending on the field that is being
	// changed. gke.IsUpToDate returns the appropriate update operation based on
	// the difference in the desired and existing spec. Only one field can be
	// updated at a time, so if there are multiple diffs, the next one will be
	// handled after the current one is completed.
	_, err = fn(ctx, e.cluster, gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr)))
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateCluster)
}

func (e *clusterExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return errors.New(errNotCluster)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	// Wait until delete is complete if already deleting.
	if cr.Status.AtProvider.Status == v1beta1.ClusterStateStopping {
		return nil
	}

	_, err := e.cluster.Projects.Locations.Clusters.Delete(gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteCluster)
}

// connectionSecret return secret object for cluster instance
func connectionDetails(cluster *container.Cluster) resource.ConnectionDetails {
	config, err := gke.GenerateClientConfig(cluster)
	if err != nil {
		return nil
	}
	rawConfig, err := clientcmd.Write(config)
	if err != nil {
		return nil
	}
	cd := resource.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey:   []byte(config.Clusters[cluster.Name].Server),
		runtimev1alpha1.ResourceCredentialsSecretUserKey:       []byte(config.AuthInfos[cluster.Name].Username),
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey:   []byte(config.AuthInfos[cluster.Name].Password),
		runtimev1alpha1.ResourceCredentialsSecretCAKey:         config.Clusters[cluster.Name].CertificateAuthorityData,
		runtimev1alpha1.ResourceCredentialsSecretClientCertKey: config.AuthInfos[cluster.Name].ClientCertificateData,
		runtimev1alpha1.ResourceCredentialsSecretClientKeyKey:  config.AuthInfos[cluster.Name].ClientKeyData,
		runtimev1alpha1.ResourceCredentialsSecretKubeconfigKey: rawConfig,
	}
	return cd
}
