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
	gke "github.com/crossplaneio/stack-gcp/pkg/clients/container"
)

// Error strings.
const (
	errGetProvider         = "cannot get Provider"
	errGetProviderSecret   = "cannot get Provider Secret"
	errNewClient           = "cannot create new GKE cluster client"
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
	case v1beta1.ClusterStateRunning:
		cr.Status.SetConditions(v1alpha1.Available())
		resource.SetBindable(cr)
	case v1beta1.ClusterStateProvisioning:
		cr.Status.SetConditions(v1alpha1.Creating())
	case v1beta1.ClusterStateUnspecified, v1beta1.ClusterStateDegraded, v1beta1.ClusterStateError, v1beta1.ClusterStateReconciling:
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

	// We have to get the cluster again here to determine how to update.
	existing, err := e.cluster.Projects.Locations.Clusters.Get(gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return resource.ExternalUpdate{}, errors.Wrap(err, errGetCluster)
	}

	u, kind := gke.IsUpToDate(&cr.Spec.ForProvider, *existing)
	if u {
		return resource.ExternalUpdate{}, nil
	}

	// GKE uses different update methods depending on the field that is being
	// changed. updateFactory returns the appropriate update operation based on
	// the difference in the desired and existing spec. Only one field can be
	// updated at a time, so if there are multiple diffs, the next one will be
	// handled after the current one is completed.
	_, err = updateFactory(kind, &cr.Spec.ForProvider)(ctx, e.cluster, gke.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr)))
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateCluster)
}

func (e *clusterExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.GKECluster)
	if !ok {
		return errors.New(errNotCluster)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())

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

func updateFactory(kind gke.ClusterUpdate, update *v1beta1.GKEClusterParameters) updateFn { // nolint:gocyclo
	switch kind {
	case gke.NodePoolUpdate:
		return deleteBootstrapNodePool()
	case gke.AddonsConfigUpdate:
		return newAddonsConfigUpdate(update.AddonsConfig)
	case gke.AutoscalingUpdate:
		return newAutoscalingUpdate(update.Autoscaling)
	case gke.BinaryAuthorizationUpdate:
		return newBinaryAuthorizationUpdate(update.BinaryAuthorization)
	case gke.DatabaseEncryptionUpdate:
		return newDatabaseEncryptionUpdate(update.DatabaseEncryption)
	case gke.LegacyAbacUpdate:
		return newLegacyAbacUpdate(update.LegacyAbac)
	case gke.LocationsUpdate:
		return newLocationsUpdate(update.Locations)
	case gke.LoggingServiceUpdate:
		return newLoggingServiceUpdate(update.LoggingService)
	case gke.MaintenancePolicyUpdate:
		return newMaintenancePolicyUpdate(update.MaintenancePolicy)
	case gke.MasterAuthorizedNetworksConfigUpdate:
		return newMasterAuthorizedNetworksConfigUpdate(update.MasterAuthorizedNetworksConfig)
	case gke.MonitoringServiceUpdate:
		return newMonitoringServiceUpdate(update.MonitoringService)
	case gke.NetworkConfigUpdate:
		return newNetworkConfigUpdate(update.NetworkConfig)
	case gke.NetworkPolicyUpdate:
		return newNetworkPolicyUpdate(update.NetworkPolicy)
	case gke.PodSecurityPolicyConfigUpdate:
		return newPodSecurityPolicyConfigUpdate(update.PodSecurityPolicyConfig)
	case gke.PrivateClusterConfigUpdate:
		return newPrivateClusterConfigUpdate(update.PrivateClusterConfig)
	case gke.ResourceLabelsUpdate:
		return newResourceLabelsUpdate(update.ResourceLabels)
	case gke.ResourceUsageExportConfigUpdate:
		return newResourceUsageExportConfigUpdate(update.ResourceUsageExportConfig)
	case gke.VerticalPodAutoscalingUpdate:
		return newVerticalPodAutoscalingUpdate(update.VerticalPodAutoscaling)
	case gke.WorkloadIdentityConfigUpdate:
		return newWorkloadIdentityConfigUpdate(update.WorkloadIdentityConfig)
	}
	return noOpUpdate
}

// updateFn returns a function that updates a cluster.
type updateFn func(context.Context, *container.Service, string) (*container.Operation, error)

func noOpUpdate(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
	return nil, nil
}

// newAddonsConfigUpdate returns a function that updates the AddonsConfig of a cluster.
func newAddonsConfigUpdate(in *v1beta1.AddonsConfig) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateAddonsConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredAddonsConfig: out.AddonsConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newAutoscalingUpdate returns a function that updates the Autoscaling of a cluster.
func newAutoscalingUpdate(in *v1beta1.ClusterAutoscaling) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateAutoscaling(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredClusterAutoscaling: out.Autoscaling,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newBinaryAuthorizationUpdate returns a function that updates the BinaryAuthorization of a cluster.
func newBinaryAuthorizationUpdate(in *v1beta1.BinaryAuthorization) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateBinaryAuthorization(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredBinaryAuthorization: out.BinaryAuthorization,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newDatabaseEncryptionUpdate returns a function that updates the DatabaseEncryption of a cluster.
func newDatabaseEncryptionUpdate(in *v1beta1.DatabaseEncryption) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateDatabaseEncryption(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredDatabaseEncryption: out.DatabaseEncryption,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newLegacyAbacUpdate returns a function that updates the LegacyAbac of a cluster.
func newLegacyAbacUpdate(in *v1beta1.LegacyAbac) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateLegacyAbac(in, out)
		update := &container.SetLegacyAbacRequest{
			Enabled: out.LegacyAbac.Enabled,
		}
		return s.Projects.Locations.Clusters.SetLegacyAbac(name, update).Context(ctx).Do()
	}
}

// newLocationsUpdate returns a function that updates the Locations of a cluster.
func newLocationsUpdate(in []string) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredLocations: in,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newLoggingServiceUpdate returns a function that updates the LoggingService of a cluster.
func newLoggingServiceUpdate(in *string) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredLoggingService: gcp.StringValue(in),
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newMaintenancePolicyUpdate returns a function that updates the MaintenancePolicy of a cluster.
func newMaintenancePolicyUpdate(in *v1beta1.MaintenancePolicySpec) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateMaintenancePolicy(in, out)
		update := &container.SetMaintenancePolicyRequest{
			MaintenancePolicy: out.MaintenancePolicy,
		}
		return s.Projects.Locations.Clusters.SetMaintenancePolicy(name, update).Context(ctx).Do()
	}
}

// newMasterAuthorizedNetworksConfigUpdate returns a function that updates the MasterAuthorizedNetworksConfig of a cluster.
func newMasterAuthorizedNetworksConfigUpdate(in *v1beta1.MasterAuthorizedNetworksConfig) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateMasterAuthorizedNetworksConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredMasterAuthorizedNetworksConfig: out.MasterAuthorizedNetworksConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newMonitoringServiceUpdate returns a function that updates the MonitoringService of a cluster.
func newMonitoringServiceUpdate(in *string) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredMonitoringService: gcp.StringValue(in),
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newNetworkConfigUpdate returns a function that updates the NetworkConfig of a cluster.
func newNetworkConfigUpdate(in *v1beta1.NetworkConfigSpec) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateNetworkConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredIntraNodeVisibilityConfig: &container.IntraNodeVisibilityConfig{
					Enabled: out.NetworkConfig.EnableIntraNodeVisibility,
				},
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newNetworkPolicyUpdate returns a function that updates the NetworkPolicy of a cluster.
func newNetworkPolicyUpdate(in *v1beta1.NetworkPolicy) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateNetworkPolicy(in, out)
		update := &container.SetNetworkPolicyRequest{
			NetworkPolicy: out.NetworkPolicy,
		}
		return s.Projects.Locations.Clusters.SetNetworkPolicy(name, update).Context(ctx).Do()
	}
}

// newPodSecurityPolicyConfigUpdate returns a function that updates the PodSecurityPolicyConfig of a cluster.
func newPodSecurityPolicyConfigUpdate(in *v1beta1.PodSecurityPolicyConfig) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GeneratePodSecurityPolicyConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredPodSecurityPolicyConfig: out.PodSecurityPolicyConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newPrivateClusterConfigUpdate returns a function that updates the PrivateClusterConfig of a cluster.
func newPrivateClusterConfigUpdate(in *v1beta1.PrivateClusterConfigSpec) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GeneratePrivateClusterConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredPrivateClusterConfig: out.PrivateClusterConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newResourceLabelsUpdate returns a function that updates the ResourceLabels of a cluster.
func newResourceLabelsUpdate(in map[string]string) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.SetLabelsRequest{
			ResourceLabels: in,
		}
		return s.Projects.Locations.Clusters.SetResourceLabels(name, update).Context(ctx).Do()
	}
}

// newResourceUsageExportConfigUpdate returns a function that updates the ResourceUsageExportConfig of a cluster.
func newResourceUsageExportConfigUpdate(in *v1beta1.ResourceUsageExportConfig) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateResourceUsageExportConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredResourceUsageExportConfig: out.ResourceUsageExportConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newVerticalPodAutoscalingUpdate returns a function that updates the VerticalPodAutoscaling of a cluster.
func newVerticalPodAutoscalingUpdate(in *v1beta1.VerticalPodAutoscaling) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateVerticalPodAutoscaling(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredVerticalPodAutoscaling: out.VerticalPodAutoscaling,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newWorkloadIdentityConfigUpdate returns a function that updates the WorkloadIdentityConfig of a cluster.
func newWorkloadIdentityConfigUpdate(in *v1beta1.WorkloadIdentityConfig) updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		gke.GenerateWorkloadIdentityConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredWorkloadIdentityConfig: out.WorkloadIdentityConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// deleteBootstrapNodePool returns a function to delete the bootstrap node pool.
func deleteBootstrapNodePool() updateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		return s.Projects.Locations.Clusters.NodePools.Delete(gke.GetFullyQualifiedBNP(name)).Context(ctx).Do()
	}
}
