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

package v1beta1

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/stack-gcp/apis/compute/v1alpha3"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cluster states.
const (
	ClusterStateUnspecified  = "STATUS_UNSPECIFIED"
	ClusterStateProvisioning = "PROVISIONING"
	ClusterStateRunning      = "RUNNING"
	ClusterStateReconciling  = "RECONCILING"
	ClusterStateStopping     = "STOPPING"
	ClusterStateError        = "ERROR"
	ClusterStateDegraded     = "DEGRADED"
)

// Defaults for GKE resources.
const (
	DefaultReclaimPolicy = runtimev1alpha1.ReclaimRetain
	DefaultNumberOfNodes = int64(1)
)

// Error strings
const (
	errResourceIsNotGKECluster = "the managed resource is not a GKECluster"
)

// NetworkURIReferencerForGKECluster is an attribute referencer that resolves
// network uri from a referenced Network and assigns it to a GKECluster
type NetworkURIReferencerForGKECluster struct {
	v1alpha3.NetworkURIReferencer `json:",inline"`
}

// Assign assigns the retrieved network uri to GKECluster
func (v *NetworkURIReferencerForGKECluster) Assign(res resource.CanReference, value string) error {
	gke, ok := res.(*GKECluster)
	if !ok {
		return errors.New(errResourceIsNotGKECluster)
	}

	gke.Spec.ForProvider.Network = &value
	return nil
}

// SubnetworkURIReferencerForGKECluster is an attribute referencer that resolves
// subnetwork uri from a referenced Subnetwork and assigns it to a GKECluster
type SubnetworkURIReferencerForGKECluster struct {
	v1alpha3.SubnetworkURIReferencer `json:",inline"`
}

// Assign assigns the retrieved subnetwork uri to a GKECluster
func (v *SubnetworkURIReferencerForGKECluster) Assign(res resource.CanReference, value string) error {
	gke, ok := res.(*GKECluster)
	if !ok {
		return errors.New(errResourceIsNotGKECluster)
	}

	gke.Spec.ForProvider.Subnetwork = &value
	return nil
}

// GKEClusterObservation is used to show the observed state of the GKE cluster resource on GCP.
type GKEClusterObservation struct {
	// Conditions: Which conditions caused the current cluster state.
	Conditions []*StatusCondition `json:"conditions,omitempty"`

	// CreateTime: [Output only] The time the cluster was created,
	// in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text format.
	CreateTime string `json:"createTime,omitempty"`

	// CurrentMasterVersion: [Output only] The current software version of
	// the master endpoint.
	CurrentMasterVersion string `json:"currentMasterVersion,omitempty"`

	// CurrentNodeCount: [Output only]  The number of nodes currently in the
	// cluster. Deprecated.
	// Call Kubernetes API directly to retrieve node information.
	CurrentNodeCount int64 `json:"currentNodeCount,omitempty"`

	// CurrentNodeVersion: [Output only] Deprecated,
	// use
	// [NodePools.version](/kubernetes-engine/docs/reference/rest/v1/proj
	// ects.zones.clusters.nodePools)
	// instead. The current version of the node software components. If they
	// are
	// currently at multiple versions because they're in the process of
	// being
	// upgraded, this reflects the minimum version of all nodes.
	CurrentNodeVersion string `json:"currentNodeVersion,omitempty"`

	// Endpoint: [Output only] The IP address of this cluster's master
	// endpoint.
	// The endpoint can be accessed from the internet
	// at
	// `https://username:password@endpoint/`.
	//
	// See the `masterAuth` property of this resource for username
	// and
	// password information.
	Endpoint string `json:"endpoint,omitempty"`

	// ExpireTime: [Output only] The time the cluster will be
	// automatically
	// deleted in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text
	// format.
	ExpireTime string `json:"expireTime,omitempty"`

	// Location: [Output only] The name of the Google Compute
	// Engine
	// [zone](/compute/docs/regions-zones/regions-zones#available)
	// or
	// [region](/compute/docs/regions-zones/regions-zones#available) in
	// which
	// the cluster resides.
	Location string `json:"location"`

	// NodeIpv4CidrSize: [Output only] The size of the address space on each
	// node for hosting
	// containers. This is provisioned from within the
	// `container_ipv4_cidr`
	// range. This field will only be set when cluster is in route-based
	// network
	// mode.
	NodeIpv4CidrSize int64 `json:"nodeIpv4CidrSize,omitempty"`

	// NodePools: The node pools associated with this cluster.
	// This field should not be set if "node_config" or "initial_node_count"
	// are
	// specified.
	// NOTE(hasheddan): node pools are modelled in status only because
	// management of node pools is handled by the stack-gcp NodePool object.
	// TODO(hasheddan): determine if we want to reflect node pools in the
	// cluster status.
	// NodePools []*NodePoolClusterStatus `json:"nodePools,omitempty"`

	// SelfLink: [Output only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// ServicesIpv4Cidr: [Output only] The IP address range of the
	// Kubernetes services in
	// this cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `1.2.3.4/29`). Service addresses are
	// typically put in the last `/16` from the container CIDR.
	ServicesIpv4Cidr string `json:"servicesIpv4Cidr,omitempty"`

	// Status: [Output only] The current status of this cluster.
	//
	// Possible values:
	//   "STATUS_UNSPECIFIED" - Not set.
	//   "PROVISIONING" - The PROVISIONING state indicates the cluster is
	// being created.
	//   "RUNNING" - The RUNNING state indicates the cluster has been
	// created and is fully
	// usable.
	//   "RECONCILING" - The RECONCILING state indicates that some work is
	// actively being done on
	// the cluster, such as upgrading the master or node software. Details
	// can
	// be found in the `statusMessage` field.
	//   "STOPPING" - The STOPPING state indicates the cluster is being
	// deleted.
	//   "ERROR" - The ERROR state indicates the cluster may be unusable.
	// Details
	// can be found in the `statusMessage` field.
	//   "DEGRADED" - The DEGRADED state indicates the cluster requires user
	// action to restore
	// full functionality. Details can be found in the `statusMessage`
	// field.
	Status string `json:"status,omitempty"`

	// StatusMessage: [Output only] Additional information about the current
	// status of this
	// cluster, if available.
	StatusMessage string `json:"statusMessage,omitempty"`

	// TpuIpv4CidrBlock: [Output only] The IP address range of the Cloud
	// TPUs in this cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `1.2.3.4/29`).
	TpuIpv4CidrBlock string `json:"tpuIpv4CidrBlock,omitempty"`

	// Zone: [Output only] The name of the Google Compute
	// Engine
	// [zone](/compute/docs/zones#available) in which the
	// cluster
	// resides.
	// This field is deprecated, use location instead.
	Zone string `json:"zone,omitempty"`
}

// GKEClusterParameters define the desired state of a Google Kubernetes Engine
// cluster.
type GKEClusterParameters struct {
	// AddonsConfig: Configurations for the various addons available to run
	// in the cluster.
	// +optional
	AddonsConfig *AddonsConfig `json:"addonsConfig,omitempty"`

	// AuthenticatorGroupsConfig: Configuration controlling RBAC group
	// membership information.
	// +optional
	// +immutable
	AuthenticatorGroupsConfig *AuthenticatorGroupsConfig `json:"authenticatorGroupsConfig,omitempty"`

	// Autoscaling: Cluster-level autoscaling configuration.
	// +optional
	Autoscaling *ClusterAutoscaling `json:"autoscaling,omitempty"`

	// BinaryAuthorization: Configuration for Binary Authorization.
	// +optional
	BinaryAuthorization *BinaryAuthorization `json:"binaryAuthorization,omitempty"`

	// ClusterIpv4Cidr: The IP address range of the container pods in this
	// cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `10.96.0.0/14`). Leave blank to have
	// one automatically chosen or specify a `/14` block in `10.0.0.0/8`.
	// +optional
	// +immutable
	ClusterIpv4Cidr *string `json:"clusterIpv4Cidr,omitempty"`

	// DatabaseEncryption: Configuration of etcd encryption.
	// +optional
	DatabaseEncryption *DatabaseEncryption `json:"databaseEncryption,omitempty"`

	// DefaultMaxPodsConstraint: The default constraint on the maximum
	// number of pods that can be run
	// simultaneously on a node in the node pool of this cluster. Only
	// honored
	// if cluster created with IP Alias support.
	// +optional
	// +immutable
	DefaultMaxPodsConstraint *MaxPodsConstraint `json:"defaultMaxPodsConstraint,omitempty"`

	// Description: An optional description of this cluster.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// EnableKubernetesAlpha: Kubernetes alpha features are enabled on this
	// cluster. This includes alpha API groups (e.g. v1alpha1) and features that
	// may not be production ready in the kubernetes version of the master and
	// nodes. The cluster has no SLA for uptime and master/node upgrades are
	// disabled. Alpha enabled clusters are automatically deleted thirty days
	// after creation.
	// +optional
	// +immutable
	EnableKubernetesAlpha *bool `json:"enableKubernetesAlpha,omitempty"`

	// EnableTpu: Enable the ability to use Cloud TPUs in this cluster.
	// +optional
	// +immutable
	EnableTpu *bool `json:"enableTpu,omitempty"`

	// InitialClusterVersion: The initial Kubernetes version for this
	// cluster.  Valid versions are those
	// found in validMasterVersions returned by getServerConfig.  The
	// version can
	// be upgraded over time; such upgrades are reflected
	// in
	// currentMasterVersion and currentNodeVersion.
	//
	// Users may specify either explicit versions offered by
	// Kubernetes Engine or version aliases, which have the following
	// behavior:
	//
	// - "latest": picks the highest valid Kubernetes version
	// - "1.X": picks the highest valid patch+gke.N patch in the 1.X
	// version
	// - "1.X.Y": picks the highest valid gke.N patch in the 1.X.Y version
	// - "1.X.Y-gke.N": picks an explicit Kubernetes version
	// - "","-": picks the default Kubernetes version
	// +optional
	// +immutable
	InitialClusterVersion *string `json:"initialClusterVersion,omitempty"`

	// IpAllocationPolicy: Configuration for cluster IP allocation.
	// +optional
	// +immutable
	IpAllocationPolicy *IPAllocationPolicy `json:"ipAllocationPolicy,omitempty"`

	// LabelFingerprint: The fingerprint of the set of labels for this
	// cluster.
	// +optional
	// +immutable
	LabelFingerprint *string `json:"labelFingerprint,omitempty"`

	// LegacyAbac: Configuration for the legacy ABAC authorization mode.
	// NOTE(hasheddan): this can only be updated via setLegacyAbac
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setLegacyAbac
	// +optional
	LegacyAbac *LegacyAbac `json:"legacyAbac,omitempty"`

	// Location: [Output only] The name of the Google Compute
	// Engine
	// [zone](/compute/docs/regions-zones/regions-zones#available)
	// or
	// [region](/compute/docs/regions-zones/regions-zones#available) in
	// which
	// the cluster resides.
	// NOTE(hasheddan): this is labelled as Output Only by GCP but is required
	// to create a cluster. It is not included in the actual cluster object
	// itself, but is instead passed to the create call. If a region is given
	// the cluster will be Regional, if a zone is given the cluster will be
	// Zonal.
	// +immutable
	Location string `json:"location"`

	// Locations: The list of Google Compute
	// Engine
	// [zones](/compute/docs/zones#available) in which the cluster's
	// nodes
	// should be located.
	// +optional
	Locations []string `json:"locations,omitempty"`

	// LoggingService: The logging service the cluster should use to write
	// logs.
	// Currently available options:
	//
	// * "logging.googleapis.com/kubernetes" - the Google Cloud
	// Logging
	// service with Kubernetes-native resource model in Stackdriver
	// * `logging.googleapis.com` - the Google Cloud Logging service.
	// * `none` - no logs will be exported from the cluster.
	// * if left as an empty string,`logging.googleapis.com` will be used.
	// +optional
	LoggingService *string `json:"loggingService,omitempty"`

	// MaintenancePolicy: Configure the maintenance policy for this cluster.
	// NOTE(hasheddan): this can only be updated via setMaintenancePolicy
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setMaintenancePolicy
	// +optional
	MaintenancePolicy *MaintenancePolicy `json:"maintenancePolicy,omitempty"`

	// MasterAuth: The authentication information for accessing the master
	// endpoint.
	// If unspecified, the defaults are used:
	// For clusters before v1.12, if master_auth is unspecified, `username`
	// will
	// be set to "admin", a random password will be generated, and a
	// client
	// certificate will be issued.
	// NOTE(hasheddan): this can only be updated via setMasterAuth
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setMasterAuth
	// +optional
	MasterAuth *MasterAuth `json:"masterAuth,omitempty"`

	// MasterAuthorizedNetworksConfig: The configuration options for master
	// authorized networks feature.
	// +optional
	MasterAuthorizedNetworksConfig *MasterAuthorizedNetworksConfig `json:"masterAuthorizedNetworksConfig,omitempty"`

	// MonitoringService: The monitoring service the cluster should use to
	// write metrics.
	// Currently available options:
	//
	// * `monitoring.googleapis.com` - the Google Cloud Monitoring
	// service.
	// * `none` - no metrics will be exported from the cluster.
	// * if left as an empty string, `monitoring.googleapis.com` will be
	// used.
	// +optional
	MonitoringService *string `json:"monitoringService,omitempty"`

	// Name: The name of this cluster. The name must be unique within this
	// project
	// and zone, and can be up to 40 characters with the following
	// restrictions:
	//
	// * Lowercase letters, numbers, and hyphens only.
	// * Must start with a letter.
	// * Must end with a number or a letter.
	// +immutable
	Name string `json:"name"`

	// Network: The name of the Google Compute
	// Engine
	// [network](/compute/docs/networks-and-firewalls#networks) to which
	// the
	// cluster is connected. If left unspecified, the `default` network
	// will be used.
	// +optional
	// +immutable
	Network *string `json:"network,omitempty"`

	// NetworkRef references to a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkRef *NetworkURIReferencerForGKECluster `json:"networkRef,omitempty" resource:"attributereferencer"`

	// NetworkConfig: Configuration for cluster networking.
	// NOTE(hasheddan): only intranode visibility can be updated here
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/ClusterUpdate?authuser=1#IntraNodeVisibilityConfig
	// +optional
	NetworkConfig *NetworkConfig `json:"networkConfig,omitempty"`

	// NetworkPolicy: Configuration options for the NetworkPolicy feature.
	// NOTE(hasheddan): this can only be updated via setNetworkPolicy
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setNetworkPolicy
	// +optional
	NetworkPolicy *NetworkPolicy `json:"networkPolicy,omitempty"`

	// PodSecurityPolicyConfig: Configuration for the PodSecurityPolicy
	// feature.
	// +optional
	PodSecurityPolicyConfig *PodSecurityPolicyConfig `json:"podSecurityPolicyConfig,omitempty"`

	// PrivateClusterConfig: Configuration for private cluster.
	// +optional
	PrivateClusterConfig *PrivateClusterConfig `json:"privateClusterConfig,omitempty"`

	// ResourceLabels: The resource labels for the cluster to use to
	// annotate any related
	// Google Compute Engine resources.
	// NOTE(hasheddan): this can only be updated via setResourceLabels
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setResourceLabels
	// +optional
	ResourceLabels map[string]string `json:"resourceLabels,omitempty"`

	// ResourceUsageExportConfig: Configuration for exporting resource
	// usages. Resource usage export is
	// disabled when this config is unspecified.
	// +optional
	ResourceUsageExportConfig *ResourceUsageExportConfig `json:"resourceUsageExportConfig,omitempty"`

	// Subnetwork: The name of the Google Compute
	// Engine
	// [subnetwork](/compute/docs/subnetworks) to which the
	// cluster is connected.
	// +optional
	// +immutable
	Subnetwork *string `json:"subnetwork,omitempty"`

	// SubnetworkRef references to a Subnetwork and retrieves its URI
	// +optional
	// +immutable
	SubnetworkRef *SubnetworkURIReferencerForGKECluster `json:"subnetworkRef,omitempty" resource:"attributereferencer"`

	// TierSettings: Cluster tier settings.
	// +optional
	// +immutable
	TierSettings *TierSettings `json:"tierSettings,omitempty"`

	// VerticalPodAutoscaling: Cluster-level Vertical Pod Autoscaling
	// configuration.
	// +optional
	VerticalPodAutoscaling *VerticalPodAutoscaling `json:"verticalPodAutoscaling,omitempty"`

	// WorkloadIdentityConfig: Configuration for the use of Kubernetes
	// Service Accounts in GCP IAM
	// policies.
	// +optional
	WorkloadIdentityConfig *WorkloadIdentityConfig `json:"workloadIdentityConfig,omitempty"`
}

// AddonsConfig is configuration for the addons that can be automatically
// spun up in the
// cluster, enabling additional functionality.
type AddonsConfig struct {
	// CloudRunConfig: Configuration for the Cloud Run addon. The
	// `IstioConfig` addon must be
	// enabled in order to enable Cloud Run addon. This option can only be
	// enabled
	// at cluster creation time.
	CloudRunConfig *CloudRunConfig `json:"cloudRunConfig,omitempty"`

	// HorizontalPodAutoscaling: Configuration for the horizontal pod
	// autoscaling feature, which
	// increases or decreases the number of replica pods a replication
	// controller
	// has based on the resource usage of the existing pods.
	HorizontalPodAutoscaling *HorizontalPodAutoscaling `json:"horizontalPodAutoscaling,omitempty"`

	// HttpLoadBalancing: Configuration for the HTTP (L7) load balancing
	// controller addon, which
	// makes it easy to set up HTTP load balancers for services in a
	// cluster.
	HttpLoadBalancing *HttpLoadBalancing `json:"httpLoadBalancing,omitempty"`

	// IstioConfig: Configuration for Istio, an open platform to connect,
	// manage, and secure
	// microservices.
	IstioConfig *IstioConfig `json:"istioConfig,omitempty"`

	// KubernetesDashboard: Configuration for the Kubernetes Dashboard.
	// This addon is deprecated, and will be disabled in 1.15. It is
	// recommended
	// to use the Cloud Console to manage and monitor your Kubernetes
	// clusters,
	// workloads and applications. For more information,
	// see:
	// https://cloud.google.com/kubernetes-engine/docs/concepts/dashboar
	// ds
	KubernetesDashboard *KubernetesDashboard `json:"kubernetesDashboard,omitempty"`

	// NetworkPolicyConfig: Configuration for NetworkPolicy. This only
	// tracks whether the addon
	// is enabled or not on the Master, it does not track whether network
	// policy
	// is enabled for the nodes.
	NetworkPolicyConfig *NetworkPolicyConfig `json:"networkPolicyConfig,omitempty"`
}

// CloudRunConfig is configuration options for the Cloud Run feature.
type CloudRunConfig struct {
	// Disabled: Whether Cloud Run addon is enabled for this cluster.
	Disabled bool `json:"disabled"`
}

// HorizontalPodAutoscaling is configuration options for the horizontal
// pod autoscaling feature, which
// increases or decreases the number of replica pods a replication
// controller
// has based on the resource usage of the existing pods.
type HorizontalPodAutoscaling struct {
	// Disabled: Whether the Horizontal Pod Autoscaling feature is enabled
	// in the cluster.
	// When enabled, it ensures that a Heapster pod is running in the
	// cluster,
	// which is also used by the Cloud Monitoring service.
	Disabled bool `json:"disabled"`
}

// HttpLoadBalancing is configuration options for the HTTP (L7) load
// balancing controller addon,
// which makes it easy to set up HTTP load balancers for services in a
// cluster.
type HttpLoadBalancing struct {
	// Disabled: Whether the HTTP Load Balancing controller is enabled in
	// the cluster.
	// When enabled, it runs a small pod in the cluster that manages the
	// load
	// balancers.
	Disabled bool `json:"disabled"`
}

// IstioConfig is configuration options for Istio addon.
type IstioConfig struct {
	// Auth: The specified Istio auth mode, either none, or mutual TLS.
	//
	// Possible values:
	//   "AUTH_NONE" - auth not enabled
	//   "AUTH_MUTUAL_TLS" - auth mutual TLS enabled
	// +optional
	Auth *string `json:"auth,omitempty"`

	// Disabled: Whether Istio is enabled for this cluster.
	// +optional
	Disabled *bool `json:"disabled,omitempty"`
}

// KubernetesDashboard is configuration for the Kubernetes Dashboard.
type KubernetesDashboard struct {
	// Disabled: Whether the Kubernetes Dashboard is enabled for this
	// cluster.
	Disabled bool `json:"disabled"`
}

// NetworkPolicyConfig is configuration for NetworkPolicy. This only
// tracks whether the addon
// is enabled or not on the Master, it does not track whether network
// policy
// is enabled for the nodes.
type NetworkPolicyConfig struct {
	// Disabled: Whether NetworkPolicy is enabled for this cluster.
	Disabled bool `json:"disabled"`
}

// AuthenticatorGroupsConfig is configuration for returning group
// information from authenticators.
type AuthenticatorGroupsConfig struct {
	// Enabled: Whether this cluster should return group membership
	// lookups
	// during authentication using a group of security groups.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// SecurityGroup: The name of the security group-of-groups to be used.
	// Only relevant
	// if enabled = true.
	// +optional
	SecurityGroup *string `json:"securityGroup,omitempty"`

	// TODO(hasheddan): add security group ref
}

// ClusterAutoscaling contains global, per-cluster
// information
// required by Cluster Autoscaler to automatically adjust
// the size of the cluster and create/delete
// node pools based on the current needs.
type ClusterAutoscaling struct {
	// AutoprovisioningLocations: The list of Google Compute Engine
	// [zones](/compute/docs/zones#available)
	// in which the NodePool's nodes can be created by NAP.
	AutoprovisioningLocations []string `json:"autoprovisioningLocations,omitempty"`

	// AutoprovisioningNodePoolDefaults: AutoprovisioningNodePoolDefaults
	// contains defaults for a node pool
	// created by NAP.
	AutoprovisioningNodePoolDefaults *AutoprovisioningNodePoolDefaults `json:"autoprovisioningNodePoolDefaults,omitempty"`

	// EnableNodeAutoprovisioning: Enables automatic node pool creation and
	// deletion.
	// +optional
	EnableNodeAutoprovisioning *bool `json:"enableNodeAutoprovisioning,omitempty"`

	// ResourceLimits: Contains global constraints regarding minimum and
	// maximum
	// amount of resources in the cluster.
	ResourceLimits []*ResourceLimit `json:"resourceLimits,omitempty"`
}

// AutoprovisioningNodePoolDefaults contains
// defaults for a node pool created
// by NAP.
type AutoprovisioningNodePoolDefaults struct {
	// OauthScopes: Scopes that are used by NAP when creating node pools. If
	// oauth_scopes are
	// specified, service_account should be empty.
	OauthScopes []string `json:"oauthScopes,omitempty"`

	// ServiceAccount: The Google Cloud Platform Service Account to be used
	// by the node VMs. If
	// service_account is specified, scopes should be empty.
	// +optional
	ServiceAccount *string `json:"serviceAccount,omitempty"`

	// TODO(hasheddan): add service account ref
}

// ResourceLimit contains information about amount of some resource in
// the cluster.
// For memory, value should be in GB.
type ResourceLimit struct {
	// Maximum: Maximum amount of the resource in the cluster.
	Maximum *int64 `json:"maximum,omitempty"`

	// Minimum: Minimum amount of the resource in the cluster.
	Minimum *int64 `json:"minimum,omitempty"`

	// ResourceType: Resource name "cpu", "memory" or gpu-specific string.
	ResourceType *string `json:"resourceType,omitempty"`
}

// BinaryAuthorization is configuration for Binary Authorization.
type BinaryAuthorization struct {
	// Enabled: Enable Binary Authorization for this cluster. If enabled,
	// all container
	// images will be validated by Google Binauthz.
	Enabled bool `json:"enabled,omitempty"`
}

// DatabaseEncryption is configuration of etcd encryption.
type DatabaseEncryption struct {
	// KeyName: Name of CloudKMS key to use for the encryption of secrets in
	// etcd.
	// Ex.
	// projects/my-project/locations/global/keyRings/my-ring/cryptoKeys/my-ke
	// y
	// +optional
	KeyName *string `json:"keyName,omitempty"`

	// State: Denotes the state of etcd encryption.
	//
	// Possible values:
	//   "UNKNOWN" - Should never be set
	//   "ENCRYPTED" - Secrets in etcd are encrypted.
	//   "DECRYPTED" - Secrets in etcd are stored in plain text (at etcd
	// level) - this is
	// unrelated to Google Compute Engine level full disk encryption.
	// +optional
	State *string `json:"state,omitempty"`
}

// StatusCondition describes why a cluster or a node
// pool has a certain status
// (e.g., ERROR or DEGRADED).
type StatusCondition struct {
	// Code: Machine-friendly representation of the condition
	//
	// Possible values:
	//   "UNKNOWN" - UNKNOWN indicates a generic condition.
	//   "GCE_STOCKOUT" - GCE_STOCKOUT indicates a Google Compute Engine
	// stockout.
	//   "GKE_SERVICE_ACCOUNT_DELETED" - GKE_SERVICE_ACCOUNT_DELETED
	// indicates that the user deleted their robot
	// service account.
	//   "GCE_QUOTA_EXCEEDED" - Google Compute Engine quota was exceeded.
	//   "SET_BY_OPERATOR" - Cluster state was manually changed by an SRE
	// due to a system logic error.
	// More codes TBA
	Code string `json:"code,omitempty"`

	// Message: Human-friendly representation of the condition
	Message string `json:"message,omitempty"`
}

// MaxPodsConstraint defines constraints applied to pods.
type MaxPodsConstraint struct {
	// MaxPodsPerNode: Constraint enforced on the max num of pods per node.
	MaxPodsPerNode int64 `json:"maxPodsPerNode"`
}

// IPAllocationPolicy is configuration for controlling how IPs are
// allocated in the cluster.
type IPAllocationPolicy struct {
	// AllowRouteOverlap: If true, allow allocation of cluster CIDR ranges
	// that overlap with certain
	// kinds of network routes. By default we do not allow cluster CIDR
	// ranges to
	// intersect with any user declared routes. With allow_route_overlap ==
	// true,
	// we allow overlapping with CIDR ranges that are larger than the
	// cluster CIDR
	// range.
	//
	// If this field is set to true, then cluster and services CIDRs must
	// be
	// fully-specified (e.g. `10.96.0.0/14`, but not `/14`), which means:
	// 1) When `use_ip_aliases` is true, `cluster_ipv4_cidr_block` and
	//    `services_ipv4_cidr_block` must be fully-specified.
	// 2) When `use_ip_aliases` is false, `cluster.cluster_ipv4_cidr` muse
	// be
	//    fully-specified.
	AllowRouteOverlap *bool `json:"allowRouteOverlap,omitempty"`

	// ClusterIpv4CidrBlock: The IP address range for the cluster pod IPs. If
	// this field is set, then `cluster.cluster_ipv4_cidr` must be left blank.
	//
	// This field is only applicable when `use_ip_aliases` is true.
	//
	// Set to blank to have a range chosen with the default size.
	//
	// Set to /netmask (e.g. `/14`) to have a range chosen with a specific
	// netmask.
	//
	// Set to a
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `10.96.0.0/14`) from the RFC-1918 private networks (e.g.
	// `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) to pick a specific range
	// to use.
	// +optional
	ClusterIpv4CidrBlock *string `json:"clusterIpv4CidrBlock,omitempty"`

	// ClusterSecondaryRangeName: The name of the secondary range to be used
	// for the cluster CIDR
	// block.  The secondary range will be used for pod IP
	// addresses. This must be an existing secondary range associated
	// with the cluster subnetwork.
	//
	// This field is only applicable with use_ip_aliases is true
	// and
	// create_subnetwork is false.
	// +optional
	ClusterSecondaryRangeName *string `json:"clusterSecondaryRangeName,omitempty"`

	// CreateSubnetwork: Whether a new subnetwork will be created
	// automatically for the cluster.
	//
	// This field is only applicable when `use_ip_aliases` is true.
	// TODO(hasheddan): should this be removed?
	// +optional
	CreateSubnetwork *bool `json:"createSubnetwork,omitempty"`

	// NodeIpv4CidrBlock: The IP address range of the instance IPs in this
	// cluster.
	//
	// This is applicable only if `create_subnetwork` is true.
	//
	// Set to blank to have a range chosen with the default size.
	//
	// Set to /netmask (e.g. `/14`) to have a range chosen with a specific
	// netmask.
	//
	// Set to a
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `10.96.0.0/14`) from the RFC-1918 private networks (e.g.
	// `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) to pick a specific range
	// to use.
	// TODO(hasheddan): should this be removed?
	// +optional
	NodeIpv4CidrBlock *string `json:"nodeIpv4CidrBlock,omitempty"`

	// ServicesIpv4CidrBlock: The IP address range of the services IPs in this
	// cluster. If blank, a range will be automatically chosen with the default
	// size.
	//
	// This field is only applicable when `use_ip_aliases` is true.
	//
	// Set to blank to have a range chosen with the default size.
	//
	// Set to /netmask (e.g. `/14`) to have a range chosen with a specific
	// netmask.
	//
	// Set to a
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `10.96.0.0/14`) from the RFC-1918 private networks (e.g.
	// `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) to pick a specific range
	// to use.
	// +optional
	ServicesIpv4CidrBlock *string `json:"servicesIpv4CidrBlock,omitempty"`

	// ServicesSecondaryRangeName: The name of the secondary range to be
	// used as for the services
	// CIDR block.  The secondary range will be used for service
	// ClusterIPs. This must be an existing secondary range associated
	// with the cluster subnetwork.
	//
	// This field is only applicable with use_ip_aliases is true
	// and
	// create_subnetwork is false.
	// +optional
	ServicesSecondaryRangeName *string `json:"servicesSecondaryRangeName,omitempty"`

	// SubnetworkName: A custom subnetwork name to be used if
	// `create_subnetwork` is true.  If
	// this field is empty, then an automatic name will be chosen for the
	// new
	// subnetwork.
	// TODO(hasheddan): should this be removed?
	SubnetworkName *string `json:"subnetworkName,omitempty"`

	// TpuIpv4CidrBlock: The IP address range of the Cloud TPUs in this cluster.
	// If unspecified, a range will be automatically chosen with the default
	// size.
	//
	// This field is only applicable when `use_ip_aliases` is true.
	//
	// If unspecified, the range will use the default size.
	//
	// Set to /netmask (e.g. `/14`) to have a range chosen with a specific
	// netmask.
	//
	// Set to a
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `10.96.0.0/14`) from the RFC-1918 private networks (e.g.
	// `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) to pick a specific range
	// to use.
	// +optional
	TpuIpv4CidrBlock *string `json:"tpuIpv4CidrBlock,omitempty"`

	// UseIpAliases: Whether alias IPs will be used for pod IPs in the
	// cluster.
	// +optional
	UseIpAliases *bool `json:"useIpAliases,omitempty"`
}

// LegacyAbac is configuration for the legacy Attribute Based Access
// Control authorization
// mode.
type LegacyAbac struct {
	// Enabled: Whether the ABAC authorizer is enabled for this cluster.
	// When enabled,
	// identities in the system, including service accounts, nodes,
	// and
	// controllers, will have statically granted permissions beyond
	// those
	// provided by the RBAC configuration or IAM.
	Enabled bool `json:"enabled,omitempty"`
}

// MaintenancePolicy defines the maintenance policy
// to be used for the cluster.
type MaintenancePolicy struct {
	// Window: Specifies the maintenance window in which maintenance may be
	// performed.
	Window MaintenanceWindow `json:"window"`
}

// MaintenanceWindow defines the maintenance window
// to be used for the cluster.
type MaintenanceWindow struct {
	// DailyMaintenanceWindow: DailyMaintenanceWindow specifies a daily
	// maintenance operation window.
	DailyMaintenanceWindow DailyMaintenanceWindow `json:"dailyMaintenanceWindow"`
}

// DailyMaintenanceWindow is the time window specified for daily maintenance
// operations.
type DailyMaintenanceWindow struct {
	// Duration: [Output only] Duration of the time window, automatically
	// chosen to be
	// smallest possible in the given scenario.
	// Duration will be in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)
	// format "PTnHnMnS".
	// Duration string `json:"duration,omitempty"`
	// TODO(hasheddan): move to status

	// StartTime: Time within the maintenance window to start the
	// maintenance operations.
	// Time format should be in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)
	// format "HH:MM", where HH : [00-23] and MM : [00-59] GMT.
	StartTime string `json:"startTime"`
}

// MasterAuth is the authentication information for accessing the master endpoint.
// Authentication can be done using HTTP basic auth or using client
// certificates.
type MasterAuth struct {
	// ClientCertificate: [Output only] Base64-encoded public certificate
	// used by clients to
	// authenticate to the cluster endpoint.
	// ClientCertificate string `json:"clientCertificate,omitempty"`
	// TODO(hasheddan): move to status

	// ClientCertificateConfig: Configuration for client certificate
	// authentication on the cluster. For
	// clusters before v1.12, if no configuration is specified, a
	// client
	// certificate is issued.
	// +optional
	ClientCertificateConfig *ClientCertificateConfig `json:"clientCertificateConfig,omitempty"`

	// ClientKey: [Output only] Base64-encoded private key used by clients
	// to authenticate
	// to the cluster endpoint.
	// ClientKey string `json:"clientKey,omitempty"`
	// TODO(hasheddan): move to status

	// ClusterCaCertificate: [Output only] Base64-encoded public certificate
	// that is the root of
	// trust for the cluster.
	// ClusterCaCertificate string `json:"clusterCaCertificate,omitempty"`
	// TODO(hasheddan): move to status

	// Password: The password to use for HTTP basic authentication to the
	// master endpoint.
	// Because the master endpoint is open to the Internet, you should
	// create a
	// strong password.  If a password is provided for cluster creation,
	// username
	// must be non-empty.
	// +optional
	Password *string `json:"password,omitempty"`

	// Username: The username to use for HTTP basic authentication to the
	// master endpoint.
	// For clusters v1.6.0 and later, basic authentication can be disabled
	// by
	// leaving username unspecified (or setting it to the empty string).
	// +optional
	Username *string `json:"username,omitempty"`
}

// ClientCertificateConfig is configuration for client certificates on the
// cluster.
type ClientCertificateConfig struct {
	// IssueClientCertificate: Issue a client certificate.
	IssueClientCertificate bool `json:"issueClientCertificate"`
}

// MasterAuthorizedNetworksConfig is configuration options for the master
// authorized networks feature. Enabled
// master authorized networks will disallow all external traffic to
// access
// Kubernetes master through HTTPS except traffic from the given CIDR
// blocks,
// Google Compute Engine Public IPs and Google Prod IPs.
type MasterAuthorizedNetworksConfig struct {
	// CidrBlocks: cidr_blocks define up to 50 external networks that could
	// access
	// Kubernetes master through HTTPS.
	// +optional
	CidrBlocks []*CidrBlock `json:"cidrBlocks,omitempty"`

	// Enabled: Whether or not master authorized networks is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// CidrBlock contains an optional name and one CIDR block.
type CidrBlock struct {
	// CidrBlock: cidr_block must be specified in CIDR notation.
	CidrBlock string `json:"cidrBlock"`

	// DisplayName: display_name is an optional field for users to identify
	// CIDR blocks.
	// +optional
	DisplayName *string `json:"displayName,omitempty"`
}

// NetworkConfig reports the relative names of network &
// subnetwork.
type NetworkConfig struct {
	// EnableIntraNodeVisibility: Whether Intra-node visibility is enabled
	// for this cluster.
	// This makes same node pod to pod traffic visible for VPC network.
	EnableIntraNodeVisibility bool `json:"enableIntraNodeVisibility"`

	// Network: Output only. The relative name of the Google Compute
	// Engine
	// network(/compute/docs/networks-and-firewalls#networks) to which
	// the cluster is connected.
	// Example: projects/my-project/global/networks/my-network
	// Network string `json:"network,omitempty"`
	// TODO(hasheddan): move to status

	// Subnetwork: Output only. The relative name of the Google Compute
	// Engine
	// [subnetwork](/compute/docs/vpc) to which the cluster is
	// connected.
	// Example:
	// projects/my-project/regions/us-central1/subnetworks/my-subnet
	// Subnetwork string `json:"subnetwork,omitempty"`
	// TODO(hasheddan): move to status
}

// NetworkPolicy is configuration options for the NetworkPolicy
// feature.
// https://kubernetes.io/docs/concepts/services-networking/netwo
// rkpolicies/
type NetworkPolicy struct {
	// Enabled: Whether network policy is enabled on the cluster.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Provider: The selected network policy provider.
	//
	// Possible values:
	//   "PROVIDER_UNSPECIFIED" - Not set
	//   "CALICO" - Tigera (Calico Felix).
	// +optional
	Provider *string `json:"provider,omitempty"`
}

// PodSecurityPolicyConfig is configuration for the PodSecurityPolicy
// feature.
type PodSecurityPolicyConfig struct {
	// Enabled: Enable the PodSecurityPolicy controller for this cluster. If
	// enabled, pods
	// must be valid under a PodSecurityPolicy to be created.
	Enabled bool `json:"enabled,omitempty"`
}

// PrivateClusterConfig is configuration options for private clusters.
type PrivateClusterConfig struct {
	// EnablePeeringRouteSharing: Whether to enable route sharing over the
	// network peering.
	EnablePeeringRouteSharing *bool `json:"enablePeeringRouteSharing,omitempty"`

	// EnablePrivateEndpoint: Whether the master's internal IP address is
	// used as the cluster endpoint.
	// +optional
	EnablePrivateEndpoint *bool `json:"enablePrivateEndpoint,omitempty"`

	// EnablePrivateNodes: Whether nodes have internal IP addresses only. If
	// enabled, all nodes are
	// given only RFC 1918 private addresses and communicate with the master
	// via
	// private networking.
	// +optional
	EnablePrivateNodes *bool `json:"enablePrivateNodes,omitempty"`

	// MasterIpv4CidrBlock: The IP range in CIDR notation to use for the
	// hosted master network. This
	// range will be used for assigning internal IP addresses to the master
	// or
	// set of masters, as well as the ILB VIP. This range must not overlap
	// with
	// any other ranges in use within the cluster's network.
	// +optional
	MasterIpv4CidrBlock *string `json:"masterIpv4CidrBlock,omitempty"`

	// PrivateEndpoint: Output only. The internal IP address of this
	// cluster's master endpoint.
	// PrivateEndpoint string `json:"privateEndpoint,omitempty"`
	// TODO(hasheddan): move to status

	// PublicEndpoint: Output only. The external IP address of this
	// cluster's master endpoint.
	// PublicEndpoint string `json:"publicEndpoint,omitempty"`
	// TODO(hasheddan): move to status
}

// ResourceUsageExportConfig is configuration for exporting cluster
// resource usages.
type ResourceUsageExportConfig struct {
	// BigqueryDestination: Configuration to use BigQuery as usage export
	// destination.
	// +optional
	BigqueryDestination *BigQueryDestination `json:"bigqueryDestination,omitempty"`

	// ConsumptionMeteringConfig: Configuration to enable resource
	// consumption metering.
	// +optional
	ConsumptionMeteringConfig *ConsumptionMeteringConfig `json:"consumptionMeteringConfig,omitempty"`

	// EnableNetworkEgressMetering: Whether to enable network egress
	// metering for this cluster. If enabled, a
	// daemonset will be created in the cluster to meter network egress
	// traffic.
	// +optional
	EnableNetworkEgressMetering *bool `json:"enableNetworkEgressMetering,omitempty"`
}

// BigQueryDestination is parameters for using BigQuery as the destination
// of resource usage export.
type BigQueryDestination struct {
	// DatasetId: The ID of a BigQuery Dataset.
	DatasetId string `json:"datasetId,omitempty"`
}

// ConsumptionMeteringConfig is parameters for controlling consumption
// metering.
type ConsumptionMeteringConfig struct {
	// Enabled: Whether to enable consumption metering for this cluster. If
	// enabled, a
	// second BigQuery table will be created to hold resource
	// consumption
	// records.
	Enabled bool `json:"enabled,omitempty"`
}

// TierSettings is cluster tier settings.
type TierSettings struct {
	// Tier: Cluster tier.
	//
	// Possible values:
	//   "UNSPECIFIED" - UNSPECIFIED is the default value. If this value is
	// set during create or
	// update, it defaults to the project level tier setting.
	//   "STANDARD" - Represents the standard tier or base Google Kubernetes
	// Engine offering.
	//   "ADVANCED" - Represents the advanced tier.
	Tier string `json:"tier,omitempty"`
}

// VerticalPodAutoscaling contains global,
// per-cluster information
// required by Vertical Pod Autoscaler to automatically adjust
// the resources of pods controlled by it.
type VerticalPodAutoscaling struct {
	// Enabled: Enables vertical pod autoscaling.
	Enabled bool `json:"enabled,omitempty"`
}

// WorkloadIdentityConfig is configuration for the use of Kubernetes
// Service Accounts in GCP IAM
// policies.
type WorkloadIdentityConfig struct {
	// IdentityNamespace: IAM Identity Namespace to attach all Kubernetes
	// Service Accounts to.
	IdentityNamespace string `json:"identityNamespace,omitempty"`
}

// A GKEClusterSpec defines the desired state of a GKECluster.
type GKEClusterSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  GKEClusterParameters `json:"forProvider,omitempty"`
}

// A GKEClusterStatus represents the observed state of a GKECluster.
type GKEClusterStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     GKEClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// A GKECluster is a managed resource that represents a Google Kubernetes Engine
// cluster.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLUSTER-NAME",type="string",JSONPath=".status.clusterName"
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=".status.endpoint"
// +kubebuilder:printcolumn:name="CLUSTER-CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.zone"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type GKECluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GKEClusterSpec   `json:"spec,omitempty"`
	Status GKEClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GKEClusterList contains a list of GKECluster items
type GKEClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GKECluster `json:"items"`
}

// A GKEClusterClassSpecTemplate is a template for the spec of a dynamically
// provisioned GKECluster.
type GKEClusterClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	GKEClusterParameters              `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// A GKEClusterClass is a resource class. It defines the desired spec of
// resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type GKEClusterClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// GKECluster.
	SpecTemplate GKEClusterClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// GKEClusterClassList contains a list of cloud memorystore resource classes.
type GKEClusterClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GKEClusterClass `json:"items"`
}
