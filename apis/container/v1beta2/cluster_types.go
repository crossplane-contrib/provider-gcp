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

// nolint:gocritic,golint // Deprecation comment format false positives.
package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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
	DefaultNumberOfNodes = int64(1)
)

// ClusterParameters define the desired state of a Google Kubernetes Engine
// cluster. Most of its fields are direct mirror of GCP Cluster object.
// See https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.locations.clusters#Cluster
type ClusterParameters struct {
	// NOTE(hasheddan): Location is labelled as Output Only by GCP but is required
	// to create a cluster. It is not included in the actual cluster object
	// itself, but is instead passed to the create call. If a region is given
	// the cluster will be Regional, if a zone is given the cluster will be
	// Zonal.

	// Location: The name of the Google Compute
	// Engine
	// [zone](https://cloud.google.com/compute/docs/regions-zones/regions-zones#available)
	// or
	// [region](https://cloud.google.com/compute/docs/regions-zones/regions-zones#available) in
	// which
	// the cluster resides.
	// +immutable
	Location string `json:"location"`

	// AddonsConfig: Configurations for the various addons available to run
	// in the cluster.
	// +optional
	AddonsConfig *AddonsConfig `json:"addonsConfig,omitempty"`

	// AuthenticatorGroupsConfig: Configuration controlling RBAC group
	// membership information.
	// +optional
	// +immutable
	AuthenticatorGroupsConfig *AuthenticatorGroupsConfig `json:"authenticatorGroupsConfig,omitempty"`

	// Autopilot: Autopilot configuration for the cluster.
	// +optional
	// +immutable
	Autopilot *Autopilot `json:"autopilot,omitempty"`

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

	// ConfidentialNodes: Configuration of Confidential Nodes
	// +optional
	// +immutable
	ConfidentialNodes *ConfidentialNodes `json:"confidentialNodes,omitempty"`

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

	// IPAllocationPolicy: Configuration for cluster IP allocation.
	// +optional
	// +immutable
	IPAllocationPolicy *IPAllocationPolicy `json:"ipAllocationPolicy,omitempty"`

	// LabelFingerprint: The fingerprint of the set of labels for this
	// cluster.
	// +optional
	// +immutable
	LabelFingerprint *string `json:"labelFingerprint,omitempty"`

	// NOTE(hasheddan): LegacyAbac can only be updated via setLegacyAbac
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setLegacyAbac

	// LegacyAbac: Configuration for the legacy ABAC authorization mode.
	// +optional
	LegacyAbac *LegacyAbac `json:"legacyAbac,omitempty"`

	// Locations: The list of Google Compute
	// Engine
	// [zones](https://cloud.google.com/compute/docs/zones#available) in which the cluster's
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

	// NOTE(hasheddan): MaintenancePolciy can only be updated via
	// setMaintenancePolicy
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setMaintenancePolicy

	// MaintenancePolicy: Configure the maintenance policy for this cluster.
	// +optional
	MaintenancePolicy *MaintenancePolicySpec `json:"maintenancePolicy,omitempty"`

	// NOTE(hasheddan): MasterAuth can only be updated via setMasterAuth
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setMasterAuth

	// MasterAuth: The authentication information for accessing the master
	// endpoint.
	// If unspecified, the defaults are used:
	// For clusters before v1.12, if master_auth is unspecified, `username`
	// will
	// be set to "admin", a random password will be generated, and a
	// client
	// certificate will be issued.
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

	// Network: The name of the Google Compute
	// Engine
	// [network](https://cloud.google.com/vpc/docs/vpc#vpc_networks_and_subnets) to which
	// the
	// cluster is connected. If left unspecified, the `default` network
	// will be used.
	// +optional
	// +immutable
	Network *string `json:"network,omitempty"`

	// NetworkRef references to a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkRef *xpv1.Reference `json:"networkRef,omitempty"`

	// NetworkSelector selects a reference to a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkSelector *xpv1.Selector `json:"networkSelector,omitempty"`

	// NetworkConfig: Configuration for cluster networking.
	// +optional
	NetworkConfig *NetworkConfigSpec `json:"networkConfig,omitempty"`

	// NOTE(hasheddan): NetworkPolicy can only be updated via setNetworkPolicy
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setNetworkPolicy

	// NetworkPolicy: Configuration options for the NetworkPolicy feature.
	// +optional
	NetworkPolicy *NetworkPolicy `json:"networkPolicy,omitempty"`

	// NotificationConfig: Notification configuration of the cluster.
	NotificationConfig *NotificationConfig `json:"notificationConfig,omitempty"`

	// PrivateClusterConfig: Configuration for private cluster.
	// +optional
	PrivateClusterConfig *PrivateClusterConfigSpec `json:"privateClusterConfig,omitempty"`

	// ReleaseChannel: Release channel configuration.
	ReleaseChannel *ReleaseChannel `json:"releaseChannel,omitempty"`

	// NOTE(hasheddan): ResourceLabels can only be updated via setResourceLabels
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters/setResourceLabels

	// ResourceLabels: The resource labels for the cluster to use to
	// annotate any related
	// Google Compute Engine resources.
	// +optional
	ResourceLabels map[string]string `json:"resourceLabels,omitempty"`

	// ResourceUsageExportConfig: Configuration for exporting resource
	// usages. Resource usage export is
	// disabled when this config is unspecified.
	// +optional
	ResourceUsageExportConfig *ResourceUsageExportConfig `json:"resourceUsageExportConfig,omitempty"`

	// Subnetwork: The name of the Google Compute
	// Engine
	// [subnetwork](https://cloud.google.com/vpc/docs/subnets) to which the
	// cluster is connected.
	// +optional
	// +immutable
	Subnetwork *string `json:"subnetwork,omitempty"`

	// SubnetworkRef references to a Subnetwork and retrieves its URI
	// +optional
	// +immutable
	SubnetworkRef *xpv1.Reference `json:"subnetworkRef,omitempty"`

	// SubnetworkSelector selects a reference to a Subnetwork and retrieves its
	// URI
	// +optional
	// +immutable
	SubnetworkSelector *xpv1.Selector `json:"subnetworkSelector,omitempty"`

	// VerticalPodAutoscaling: Cluster-level Vertical Pod Autoscaling
	// configuration.
	// +optional
	VerticalPodAutoscaling *VerticalPodAutoscaling `json:"verticalPodAutoscaling,omitempty"`

	// WorkloadIdentityConfig: Configuration for the use of Kubernetes
	// Service Accounts in GCP IAM
	// policies.
	// +optional
	WorkloadIdentityConfig *WorkloadIdentityConfig `json:"workloadIdentityConfig,omitempty"`

	// IdentityServiceConfig: Cluster-level external identity providers
	// configuration.
	// +optional
	IdentityServiceConfig *IdentityServiceConfig `json:"identityServiceConfig,omitempty"`
}

// ClusterObservation is used to show the observed state of the GKE cluster resource on GCP.
type ClusterObservation struct {
	// Conditions: Which conditions caused the current cluster state.
	Conditions []*StatusCondition `json:"conditions,omitempty"`

	// CreateTime: The time the cluster was created,
	// in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text format.
	CreateTime string `json:"createTime,omitempty"`

	// CurrentMasterVersion: The current software version of
	// the master endpoint.
	CurrentMasterVersion string `json:"currentMasterVersion,omitempty"`

	// CurrentNodeCount:  The number of nodes currently in the
	// cluster. Deprecated.
	// Call Kubernetes API directly to retrieve node information.
	CurrentNodeCount int64 `json:"currentNodeCount,omitempty"`

	// CurrentNodeVersion: Deprecated,
	// use
	// [NodePools.version](/kubernetes-engine/docs/reference/rest/v1/projects.zones.clusters.nodePools)
	// instead. The current version of the node software components. If they
	// are
	// currently at multiple versions because they're in the process of
	// being
	// upgraded, this reflects the minimum version of all nodes.
	CurrentNodeVersion string `json:"currentNodeVersion,omitempty"`

	// Endpoint: The IP address of this cluster's master
	// endpoint.
	// The endpoint can be accessed from the internet
	// at
	// `https://username:password@endpoint/`.
	//
	// See the `masterAuth` property of this resource for username
	// and
	// password information.
	Endpoint string `json:"endpoint,omitempty"`

	// ExpireTime: The time the cluster will be
	// automatically
	// deleted in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text
	// format.
	ExpireTime string `json:"expireTime,omitempty"`

	// Location: The name of the Google Compute
	// Engine
	// [zone](https://cloud.google.com/compute/docs/regions-zones/regions-zones#available)
	// or
	// [region](https://cloud.google.com/compute/docs/regions-zones/regions-zones#available) in
	// which
	// the cluster resides.
	Location string `json:"location"`

	// MaintenancePolicy: Configure the maintenance policy for this cluster.
	MaintenancePolicy *MaintenancePolicyStatus `json:"maintenancePolicy,omitempty"`

	// NetworkConfig: Configuration for cluster networking.
	NetworkConfig *NetworkConfigStatus `json:"networkConfig,omitempty"`

	// NodeIpv4CidrSize: The size of the address space on each
	// node for hosting
	// containers. This is provisioned from within the
	// `container_ipv4_cidr`
	// range. This field will only be set when cluster is in route-based
	// network
	// mode.
	NodeIpv4CidrSize int64 `json:"nodeIpv4CidrSize,omitempty"`

	// PrivateClusterConfig: Configuration for private cluster.
	PrivateClusterConfig *PrivateClusterConfigStatus `json:"privateClusterConfig,omitempty"`

	// NOTE(hasheddan): node pools are modelled in status only because
	// management of node pools is handled by the provider-gcp NodePool object.

	// NodePools: The node pools associated with this cluster.
	// This field should not be set if "node_config" or "initial_node_count"
	// are
	// specified.
	NodePools []*NodePoolClusterStatus `json:"nodePools,omitempty"`

	// SelfLink: Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// ServicesIpv4Cidr: The IP address range of the
	// Kubernetes services in
	// this cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `1.2.3.4/29`). Service addresses are
	// typically put in the last `/16` from the container CIDR.
	ServicesIpv4Cidr string `json:"servicesIpv4Cidr,omitempty"`

	// Status: The current status of this cluster.
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

	// StatusMessage: Additional information about the current
	// status of this
	// cluster, if available.
	StatusMessage string `json:"statusMessage,omitempty"`

	// TpuIpv4CidrBlock: The IP address range of the Cloud
	// TPUs in this cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `1.2.3.4/29`).
	TpuIpv4CidrBlock string `json:"tpuIpv4CidrBlock,omitempty"`

	// Zone: The name of the Google Compute
	// Engine
	// [zone](https://cloud.google.com/compute/docs/zones#available) in which the
	// cluster
	// resides.
	// This field is deprecated, use location instead.
	Zone string `json:"zone,omitempty"`
}

// AddonsConfig is configuration for the addons that can be automatically
// spun up in the
// cluster, enabling additional functionality.
type AddonsConfig struct {
	// CloudRunConfig: Configuration for the Cloud Run addon. The
	// `IstioConfig` addon must be
	// enabled in order to enable Cloud Run addon. This option can only be
	// enabled at cluster creation time.
	// +optional
	CloudRunConfig *CloudRunConfig `json:"cloudRunConfig,omitempty"`

	// ConfigConnectorConfig: Configuration for the ConfigConnector add-on,
	// a Kubernetes extension to manage hosted GCP services through the
	// Kubernetes API
	// +optional
	ConfigConnectorConfig *ConfigConnectorConfig `json:"configConnectorConfig,omitempty"`

	// DNSCacheConfig: Configuration for NodeLocalDNS, a dns cache running
	// on cluster nodes
	// +optional
	DNSCacheConfig *DNSCacheConfig `json:"dnsCacheConfig,omitempty"`

	// GCEPersistentDiskCSIDriverConfig: Configuration for the GCP Compute
	// Persistent Disk CSI driver.
	// +optional
	GCEPersistentDiskCSIDriverConfig *GCEPersistentDiskCSIDriverConfig `json:"gcePersistentDiskCsiDriverConfig,omitempty"`

	// HorizontalPodAutoscaling: Configuration for the horizontal pod
	// autoscaling feature, which increases or decreases the number of replica
	// pods a replication controller has based on the resource usage of the
	// existing pods.
	// +optional
	HorizontalPodAutoscaling *HorizontalPodAutoscaling `json:"horizontalPodAutoscaling,omitempty"`

	// HTTpLoadBalancing: Configuration for the HTTP (L7) load balancing
	// controller addon, which
	// makes it easy to set up HTTP load balancers for services in a
	// cluster.
	// +optional
	HTTPLoadBalancing *HTTPLoadBalancing `json:"httpLoadBalancing,omitempty"`

	// KubernetesDashboard: Configuration for the Kubernetes Dashboard.
	// This addon is deprecated, and will be disabled in 1.15. It is
	// recommended
	// to use the Cloud Console to manage and monitor your Kubernetes
	// clusters,
	// workloads and applications. For more information,
	// see:
	// https://cloud.google.com/kubernetes-engine/docs/concepts/dashboards
	// +optional
	KubernetesDashboard *KubernetesDashboard `json:"kubernetesDashboard,omitempty"`

	// NetworkPolicyConfig: Configuration for NetworkPolicy. This only
	// tracks whether the addon is enabled or not on the Master, it does not
	// track whether network policy is enabled for the nodes.
	// +optional
	NetworkPolicyConfig *NetworkPolicyConfig `json:"networkPolicyConfig,omitempty"`
}

// GCEPersistentDiskCSIDriverConfig is configuration for the GCE PD CSI driver.
// This option can only be enabled at cluster creation time.
type GCEPersistentDiskCSIDriverConfig struct {
	// Enabled: Whether the GCE PD CSI driver is enabled for this cluster.
	Enabled bool `json:"enabled"`
}

// DNSCacheConfig is configuration for NodeLocal DNSCache.
type DNSCacheConfig struct {
	// Enabled: Whether NodeLocal DNSCache is enabled for this cluster.
	Enabled bool `json:"enabled"`
}

// CloudRunConfig is configuration options for the Cloud Run feature.
type CloudRunConfig struct {
	// Disabled: Whether Cloud Run addon is enabled for this cluster.
	Disabled bool `json:"disabled"`

	// LoadBalancerType: Which load balancer type is installed for Cloud
	// Run.
	//
	// Possible values:
	//   "LOAD_BALANCER_TYPE_UNSPECIFIED" - Load balancer type for Cloud Run
	// is unspecified.
	//   "LOAD_BALANCER_TYPE_EXTERNAL" - Install external load balancer for
	// Cloud Run.
	//   "LOAD_BALANCER_TYPE_INTERNAL" - Install internal load balancer for
	// Cloud Run.
	// +optional
	LoadBalancerType *string `json:"loadBalancerType,omitempty"`
}

// ConfigConnectorConfig is configuration options for the Config Connector
// add-on.
type ConfigConnectorConfig struct {
	// Enabled: Whether Cloud Connector is enabled for this cluster.
	Enabled bool `json:"enabled"`
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

// HTTPLoadBalancing is configuration options for the HTTP (L7) load
// balancing controller addon,
// which makes it easy to set up HTTP load balancers for services in a
// cluster.
type HTTPLoadBalancing struct {
	// Disabled: Whether the HTTP Load Balancing controller is enabled in
	// the cluster.
	// When enabled, it runs a small pod in the cluster that manages the
	// load
	// balancers.
	Disabled bool `json:"disabled"`
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
}

// Autopilot is configuration for Autopilot mode.
type Autopilot struct {
	// Enabled: Enable Autopilot
	Enabled bool `json:"enabled"`
}

// ClusterAutoscaling contains global, per-cluster
// information
// required by Cluster Autoscaler to automatically adjust
// the size of the cluster and create/delete
// node pools based on the current needs.
type ClusterAutoscaling struct {
	// AutoprovisioningLocations: The list of Google Compute Engine
	// [zones](https://cloud.google.com/compute/docs/zones#available)
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
	// BootDiskKmsKey: The Customer Managed Encryption Key used to encrypt
	// the boot disk attached to each node in the node pool. This should be
	// of the form
	// projects/[KEY_PROJECT_ID]/locations/[LOCATION]/keyRings/[RING_NAME]/cr
	// yptoKeys/[KEY_NAME]. For more information about protecting resources
	// with Cloud KMS Keys please see:
	// https://cloud.google.com/compute/docs/disks/customer-managed-encryption
	// +optional
	BootDiskKMSKey *string `json:"bootDiskKmsKey,omitempty"`

	// DiskSizeGb: Size of the disk attached to each node, specified in GB.
	// The smallest allowed disk size is 10GB. If unspecified, the default
	// disk size is 100GB.
	// +optional
	DiskSizeGb *int64 `json:"diskSizeGb,omitempty"`

	// DiskType: Type of the disk attached to each node (e.g. 'pd-standard',
	// 'pd-ssd' or 'pd-balanced') If unspecified, the default disk type is
	// 'pd-standard'
	// +optional
	DiskType *string `json:"diskType,omitempty"`

	// Management: Specifies the node management options for NAP created
	// node-pools.
	Management *NodeManagement `json:"management,omitempty"`

	// MinCpuPlatform: Minimum CPU platform to be used for NAP created node
	// pools. The instance may be scheduled on the specified or newer CPU
	// platform. Applicable values are the friendly names of CPU platforms,
	// such as minCpuPlatform: Intel Haswell or minCpuPlatform: Intel Sandy
	// Bridge. For more information, read how to specify min CPU platform
	// (https://cloud.google.com/compute/docs/instances/specify-min-cpu-platform)
	// To unset the min cpu platform field pass "automatic" as field value.
	// +optional
	MinCPUPlatform *string `json:"minCpuPlatform,omitempty"`

	// OauthScopes: Scopes that are used by NAP when creating node pools. If
	// oauth_scopes are
	// specified, service_account should be empty.
	OauthScopes []string `json:"oauthScopes,omitempty"`

	// ServiceAccount: The Google Cloud Platform Service Account to be used
	// by the node VMs. If
	// service_account is specified, scopes should be empty.
	// +optional
	ServiceAccount *string `json:"serviceAccount,omitempty"`

	// ShieldedInstanceConfig: Shielded Instance options.
	// +optional
	ShieldedInstanceConfig *ShieldedInstanceConfig `json:"shieldedInstanceConfig,omitempty"`

	// UpgradeSettings: Specifies the upgrade settings for NAP created node
	// pools
	// +optional
	UpgradeSettings *UpgradeSettings `json:"upgradeSettings,omitempty"`
}

// NodeManagement defines the set of node management services turned on for the
// node pool.
type NodeManagement struct {
	// AutoRepair: A flag that specifies whether the node auto-repair is
	// enabled for the node pool. If enabled, the nodes in this node pool
	// will be monitored and, if they fail health checks too many times, an
	// automatic repair action will be triggered.
	// +optional
	AutoRepair *bool `json:"autoRepair,omitempty"`

	// AutoUpgrade: A flag that specifies whether node auto-upgrade is
	// enabled for the node pool. If enabled, node auto-upgrade helps keep
	// the nodes in your node pool up to date with the latest release
	// version of Kubernetes.
	// +optional
	AutoUpgrade *bool `json:"autoUpgrade,omitempty"`

	// UpgradeOptions: Specifies the Auto Upgrade knobs for the node pool.
	// NOTE(hasheddan): this field is excluded because it is output only and is
	// set when upgrades are actually taking place.
}

// ShieldedInstanceConfig is a set of Shielded Instance options.
type ShieldedInstanceConfig struct {
	// EnableIntegrityMonitoring: Defines whether the instance has integrity
	// monitoring enabled. Enables monitoring and attestation of the boot
	// integrity of the instance. The attestation is performed against the
	// integrity policy baseline. This baseline is initially derived from
	// the implicitly trusted boot image when the instance is created.
	// +optional
	EnableIntegrityMonitoring *bool `json:"enableIntegrityMonitoring,omitempty"`

	// EnableSecureBoot: Defines whether the instance has Secure Boot
	// enabled. Secure Boot helps ensure that the system only runs authentic
	// software by verifying the digital signature of all boot components,
	// and halting the boot process if signature verification fails.
	// +optional
	EnableSecureBoot *bool `json:"enableSecureBoot,omitempty"`
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

// UpgradeSettings control the level of parallelism and the level of disruption
// caused by an upgrade. maxUnavailable controls the number of nodes that can be
// simultaneously unavailable. maxSurge controls the number of additional nodes
// that can be added to the node pool temporarily for the time of the upgrade to
// increase the number of available nodes. (maxUnavailable + maxSurge)
// determines the level of parallelism (how many nodes are being upgraded at the
// same time). Note: upgrades inevitably introduce some disruption since
// workloads need to be moved from old nodes to new, upgraded ones. Even if
// maxUnavailable=0, this holds true. (Disruption stays within the limits of
// PodDisruptionBudget, if it is configured.) Consider a hypothetical node pool
// with 5 nodes having maxSurge=2, maxUnavailable=1. This means the upgrade
// process upgrades 3 nodes simultaneously. It creates 2 additional (upgraded)
// nodes, then it brings down 3 old (not yet upgraded) nodes at the same time.
// This ensures that there are always at least 4 nodes available.
type UpgradeSettings struct {
	// MaxSurge: The maximum number of nodes that can be created beyond the
	// current size of the node pool during the upgrade process.
	// +optional
	MaxSurge *int64 `json:"maxSurge,omitempty"`

	// MaxUnavailable: The maximum number of nodes that can be
	// simultaneously unavailable during the upgrade process. A node is
	// considered available if its status is Ready.
	// +optional
	MaxUnavailable *int64 `json:"maxUnavailable,omitempty"`
}

// BinaryAuthorization is configuration for Binary Authorization.
type BinaryAuthorization struct {
	// Enabled: Enable Binary Authorization for this cluster. If enabled,
	// all container
	// images will be validated by Google Binauthz.
	Enabled bool `json:"enabled"`
}

// ConfidentialNodes is configuration for Confidential Nodes.
type ConfidentialNodes struct {
	// Enabled: Whether Confidential Nodes feature is enabled for all nodes
	// in this cluster.
	Enabled bool `json:"enabled"`
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

// ReleaseChannel indicates which release channel a cluster is subscribed to.
// Release channels are arranged in order of risk. When a cluster is subscribed
// to a release channel, Google maintains both the master version and the node
// version. Node auto-upgrade defaults to true and cannot be disabled.
type ReleaseChannel struct {
	// Channel: channel specifies which release channel the cluster is
	// subscribed to.
	//
	// Possible values:
	//   "UNSPECIFIED" - No channel specified.
	//   "RAPID" - RAPID channel is offered on an early access basis for
	// customers who want to test new releases. WARNING: Versions available in
	// the RAPID Channel may be subject to unresolved issues with no known
	// workaround and are not subject to any SLAs.
	//   "REGULAR" - Clusters subscribed to REGULAR receive versions that
	// are considered GA quality. REGULAR is intended for production users who
	// want to take advantage of new features.
	//   "STABLE" - Clusters subscribed to STABLE receive versions that are
	// known to be stable and reliable in production.
	Channel string `json:"channel"`
}

// NotificationConfig is the configuration of notifications.
type NotificationConfig struct {
	// Pubsub: Notification config for Pub/Sub.
	Pubsub PubSub `json:"pubsub"`
}

// PubSub specific notification config.
type PubSub struct {
	// Enabled: Enable notifications for Pub/Sub.
	Enabled bool `json:"enabled"`

	// Topic: The desired Pub/Sub topic to which notifications will be sent
	// by GKE. Format is `projects/{project}/topics/{topic}`.
	Topic string `json:"topic"`
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
	// +optional
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

	// UseIPAliases: Whether alias IPs will be used for pod IPs in the
	// cluster.
	// +optional
	UseIPAliases *bool `json:"useIpAliases,omitempty"`

	// UseRoutes: Whether routes will be used for pod IPs in the cluster.
	// This is used in conjunction with use_ip_aliases. It cannot be true if
	// use_ip_aliases is true. If both use_ip_aliases and use_routes are
	// false, then the server picks the default IP allocation mode
	UseRoutes *bool `json:"useRoutes,omitempty"`
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
	Enabled bool `json:"enabled"`
}

// MaintenancePolicySpec defines the maintenance policy
// to be used for the cluster.
type MaintenancePolicySpec struct {
	// Window: Specifies the maintenance window in which maintenance may be
	// performed.
	Window MaintenanceWindowSpec `json:"window"`
}

// MaintenanceWindowSpec defines the maintenance window
// to be used for the cluster.
type MaintenanceWindowSpec struct {
	// DailyMaintenanceWindow: DailyMaintenanceWindow specifies a daily
	// maintenance operation window.
	// +optional
	DailyMaintenanceWindow *DailyMaintenanceWindowSpec `json:"dailyMaintenanceWindow,omitempty"`

	// MaintenanceExclusions: Exceptions to maintenance window.
	// Non-emergency maintenance should not occur in these windows.
	// +optional
	MaintenanceExclusions map[string]TimeWindow `json:"maintenanceExclusions,omitempty"`

	// RecurringWindow: RecurringWindow specifies some number of recurring
	// time periods for maintenance to occur. The time windows may be
	// overlapping. If no maintenance windows are set, maintenance can occur
	// at any time.
	// +optional
	RecurringWindow *RecurringTimeWindow `json:"recurringWindow,omitempty"`
}

// DailyMaintenanceWindowSpec is the time window specified for daily maintenance
// operations.
type DailyMaintenanceWindowSpec struct {
	// StartTime: Time within the maintenance window to start the
	// maintenance operations.
	// Time format should be in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)
	// format "HH:MM", where HH : [00-23] and MM : [00-59] GMT.
	StartTime string `json:"startTime"`
}

// TimeWindow is a window of time.
type TimeWindow struct {
	// EndTime: The time that the window ends. The end time should take
	// place after the start time.
	EndTime string `json:"endTime,omitempty"`

	// StartTime: The time that the window first starts.
	StartTime string `json:"startTime,omitempty"`
}

// RecurringTimeWindow is a recurring window of time.
type RecurringTimeWindow struct {
	// Recurrence: An RRULE
	// (https://tools.ietf.org/html/rfc5545#section-3.8.5.3) for how this
	// window reccurs. They go on for the span of time between the start and
	// end time. For example, to have something repeat every weekday, you'd
	// use: `FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR` To repeat some window daily
	// (equivalent to the DailyMaintenanceWindow): `FREQ=DAILY` For the
	// first weekend of every month: `FREQ=MONTHLY;BYSETPOS=1;BYDAY=SA,SU`
	// This specifies how frequently the window starts. Eg, if you wanted to
	// have a 9-5 UTC-4 window every weekday, you'd use something like: “`
	// start time = 2019-01-01T09:00:00-0400 end time =
	// 2019-01-01T17:00:00-0400 recurrence =
	// FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR “` Windows can span multiple days.
	// Eg, to make the window encompass every weekend from midnight Saturday
	// till the last minute of Sunday UTC: “` start time =
	// 2019-01-05T00:00:00Z end time = 2019-01-07T23:59:00Z recurrence =
	// FREQ=WEEKLY;BYDAY=SA “` Note the start and end time's specific dates
	// are largely arbitrary except to specify duration of the window and
	// when it first starts. The FREQ values of HOURLY, MINUTELY, and
	// SECONDLY are not supported.
	Recurrence *string `json:"recurrence,omitempty"`

	// Window: The window of the first recurrence.
	Window *TimeWindow `json:"window,omitempty"`
}

// MaintenancePolicyStatus defines the maintenance policy
// to be used for the cluster.
type MaintenancePolicyStatus struct {
	// Window: Specifies the maintenance window in which maintenance may be
	// performed.
	Window MaintenanceWindowStatus `json:"window,omitempty"`
}

// MaintenanceWindowStatus defines the maintenance window
// to be used for the cluster.
type MaintenanceWindowStatus struct {
	// DailyMaintenanceWindow: DailyMaintenanceWindow specifies a daily
	// maintenance operation window.
	DailyMaintenanceWindow DailyMaintenanceWindowStatus `json:"dailyMaintenanceWindow,omitempty"`
}

// DailyMaintenanceWindowStatus is the observed time window for daily
// maintenance operations.
type DailyMaintenanceWindowStatus struct {
	// Duration: Duration of the time window, automatically
	// chosen to be
	// smallest possible in the given scenario.
	// Duration will be in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)
	// format "PTnHnMnS".
	Duration string `json:"duration,omitempty"`
}

// MasterAuth is the authentication information for accessing the master endpoint.
// Authentication can be done using HTTP basic auth or using client
// certificates.
type MasterAuth struct {
	// ClientCertificateConfig: Configuration for client certificate
	// authentication on the cluster. For
	// clusters before v1.12, if no configuration is specified, a
	// client
	// certificate is issued.
	// +optional
	ClientCertificateConfig *ClientCertificateConfig `json:"clientCertificateConfig,omitempty"`

	// NOTE(hasheddan): password field is omitted and auto-generated.
	// TODO(hasheddan): support providing password via Secret.

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
	// +immutable
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

// NetworkConfigSpec reports the relative names of network &
// subnetwork.
type NetworkConfigSpec struct {
	// DatapathProvider: The desired datapath provider for this cluster. By
	// default, uses the IPTables-based kube-proxy implementation.
	//
	// Possible values:
	//   "DATAPATH_PROVIDER_UNSPECIFIED" - Default value.
	//   "LEGACY_DATAPATH" - Use the IPTables implementation based on
	// kube-proxy.
	//   "ADVANCED_DATAPATH" - Use the eBPF based GKE Dataplane V2 with
	// additional features. See the [GKE Dataplane V2
	// documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/dataplane-v2) for more.
	// +optional
	DatapathProvider *string `json:"datapathProvider,omitempty"`

	// DefaultSnatStatus: Whether the cluster disables default in-node sNAT
	// rules. In-node sNAT rules will be disabled when default_snat_status
	// is disabled. When disabled is set to false, default IP masquerade
	// rules will be applied to the nodes to prevent sNAT on cluster
	// internal traffic.
	// +optional
	DefaultSnatStatus *DefaultSnatStatus `json:"defaultSnatStatus,omitempty"`

	// EnableIntraNodeVisibility: Whether Intra-node visibility is enabled
	// for this cluster.
	// This makes same node pod to pod traffic visible for VPC network.
	// +optional
	EnableIntraNodeVisibility *bool `json:"enableIntraNodeVisibility"`

	// PrivateIpv6GoogleAccess: The desired state of IPv6 connectivity to
	// Google Services. By default, no private IPv6 access to or from Google
	// Services (all access will be via IPv4)
	//
	// Possible values:
	//   "PRIVATE_IPV6_GOOGLE_ACCESS_UNSPECIFIED" - Default value. Same as
	// DISABLED
	//   "PRIVATE_IPV6_GOOGLE_ACCESS_DISABLED" - No private access to or
	// from Google Services
	//   "PRIVATE_IPV6_GOOGLE_ACCESS_TO_GOOGLE" - Enables private IPv6
	// access to Google Services from GKE
	//   "PRIVATE_IPV6_GOOGLE_ACCESS_BIDIRECTIONAL" - Enables priate IPv6
	// access to and from Google Services
	// +optional
	PrivateIpv6GoogleAccess *string `json:"privateIpv6GoogleAccess,omitempty"`

	// DNSConfig contains the desired set of options for configuring clusterDNS.
	// +optional
	DnsConfig *DnsConfig `json:"dnsConfig,omitempty"`
}

// DefaultSnatStatus contains the desired state of whether default sNAT should
// be disabled on the cluster.
type DefaultSnatStatus struct {
	// Disabled: Disables cluster default sNAT rules.
	Disabled bool `json:"disabled"`
}

// DnsConfig contains the desired set of options for configuring clusterDNS.
type DnsConfig struct {
	// ClusterDns indicates which in-cluster DNS provider should be used.
	// Possible values:
	//   "PROVIDER_UNSPECIFIED" - indicates the default value.
	//   "PLATFORM_DEFAULT" - indicates using the GKE default DNS
	//  provider(kube-dns) for DNS resolution.
	//   "CLOUD_DNS" - indicates using CloudDNS for DNS resolution.
	// +kubebuilder:validation:Enum=PROVIDER_UNSPECIFIED;PLATFORM_DEFAULT;CLOUD_DNS
	// +optional
	ClusterDns *string `json:"clusterDns,omitempty"`

	// ClusterDnsScope indicates the scope of access to cluster DNS records.
	// Possible Values:
	//   "DNS_SCOPE_UNSPECIFIED" - indicates the default value,
	//  will be inferred as cluster scope.
	//   "VPC_SCOPE" - indicates that DNS records are accessible from within the VPC.
	// +kubebuilder:validation:Enum=DNS_SCOPE_UNSPECIFIED;VPC_SCOPE
	// +optional
	ClusterDnsScope *string `json:"clusterDnsScope,omitempty"`

	// ClusterDnsDomain is the suffix used for all cluster service records.
	// +optional
	ClusterDnsDomain *string `json:"clusterDnsDomain,omitempty"`
}

// NetworkConfigStatus reports the relative names of network &
// subnetwork.
type NetworkConfigStatus struct {
	// Network: The relative name of the Google Compute
	// Engine
	// network(https://cloud.google.com/vpc/docs/vpc#vpc_networks_and_subnets) to which
	// the cluster is connected.
	// Example: projects/my-project/global/networks/my-network
	Network string `json:"network,omitempty"`

	// Subnetwork: The relative name of the Google Compute
	// Engine
	// [subnetwork](https://cloud.google.com/vpc/docs/vpc) to which the cluster is
	// connected.
	// Example:
	// projects/my-project/regions/us-central1/subnetworks/my-subnet
	Subnetwork string `json:"subnetwork,omitempty"`
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

// PrivateClusterConfigSpec is configuration options for private clusters.
type PrivateClusterConfigSpec struct {
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

	// MasterGlobalAccessConfig: Controls master global access settings.
	// +optional
	MasterGlobalAccessConfig *PrivateClusterMasterGlobalAccessConfig `json:"masterGlobalAccessConfig,omitempty"`

	// MasterIpv4CidrBlock: The IP range in CIDR notation to use for the
	// hosted master network. This
	// range will be used for assigning internal IP addresses to the master
	// or
	// set of masters, as well as the ILB VIP. This range must not overlap
	// with
	// any other ranges in use within the cluster's network.
	// +optional
	MasterIpv4CidrBlock *string `json:"masterIpv4CidrBlock,omitempty"`
}

// PrivateClusterMasterGlobalAccessConfig is Configuration for controlling
// master global access settings.
type PrivateClusterMasterGlobalAccessConfig struct {
	// Enabled: Whenever master is accessible globally or not.
	Enabled bool `json:"enabled"`
}

// PrivateClusterConfigStatus is configuration options for private clusters.
type PrivateClusterConfigStatus struct {
	// PrivateEndpoint: The internal IP address of this
	// cluster's master endpoint.
	PrivateEndpoint string `json:"privateEndpoint,omitempty"`

	// PublicEndpoint: The external IP address of this
	// cluster's master endpoint.
	PublicEndpoint string `json:"publicEndpoint,omitempty"`
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
	DatasetID string `json:"datasetId"`
}

// ConsumptionMeteringConfig is parameters for controlling consumption
// metering.
type ConsumptionMeteringConfig struct {
	// Enabled: Whether to enable consumption metering for this cluster. If
	// enabled, a
	// second BigQuery table will be created to hold resource
	// consumption
	// records.
	Enabled bool `json:"enabled"`
}

// VerticalPodAutoscaling contains global,
// per-cluster information
// required by Vertical Pod Autoscaler to automatically adjust
// the resources of pods controlled by it.
type VerticalPodAutoscaling struct {
	// Enabled: Enables vertical pod autoscaling.
	Enabled bool `json:"enabled"`
}

// WorkloadIdentityConfig is configuration for the use of Kubernetes
// Service Accounts in GCP IAM
// policies.
type WorkloadIdentityConfig struct {
	// WorkloadPool: The workload pool to attach all Kubernetes service
	// accounts to.
	WorkloadPool string `json:"workloadPool,omitempty"`
}

// IdentityServiceConfig is the configuration to external identity providers
// with the K8S API
type IdentityServiceConfig struct {
	// Enabled: Enables external identity providers.
	Enabled bool `json:"enabled"`
}

// NOTE(hasheddan): the following structs are meant to be utilized to model Node
// Pools in the status of Cluster objects. They are not to be used to define
// configurable fields for NodePool objects.

// NodePoolClusterStatus is a subset of information about NodePools associated
// with a GKE cluster.
type NodePoolClusterStatus struct {
	// Autoscaling: Autoscaler configuration for this NodePool. Autoscaler
	// is enabled
	// only if a valid configuration is present.
	Autoscaling *NodePoolAutoscalingClusterStatus `json:"autoscaling,omitempty"`

	// Conditions: Which conditions caused the current node pool state.
	Conditions []*StatusCondition `json:"conditions,omitempty"`

	// Config: The node configuration of the pool.
	Config *NodeConfigClusterStatus `json:"config,omitempty"`

	// InitialNodeCount: The initial node count for the pool. You must
	// ensure that your
	// Compute Engine <a href="https://cloud.google.com/compute/quotas">resource
	// quota</a>
	// is sufficient for this number of instances. You must also have
	// available
	// firewall and routes quota.
	InitialNodeCount int64 `json:"initialNodeCount,omitempty"`

	// InstanceGroupUrls: The resource URLs of the [managed
	// instance
	// groups](https://cloud.google.com/compute/docs/instance-groups/creating-groups-of-managed-instances)
	// associated with this node pool.
	InstanceGroupUrls []string `json:"instanceGroupUrls,omitempty"`

	// Locations: The list of Google Compute Engine
	// [zones](https://cloud.google.com/compute/docs/regions-zones#available)
	// in which the NodePool's nodes should be located.
	Locations []string `json:"locations,omitempty"`

	// Management: NodeManagement configuration for this NodePool.
	Management *NodeManagementClusterStatus `json:"management,omitempty"`

	// MaxPodsConstraint: The constraint on the maximum number of pods that
	// can be run
	// simultaneously on a node in the node pool.
	MaxPodsConstraint *MaxPodsConstraint `json:"maxPodsConstraint,omitempty"`

	// Name: The name of the node pool.
	Name string `json:"name,omitempty"`

	// PodIpv4CidrSize: The pod CIDR block size per node in
	// this node pool.
	PodIpv4CidrSize int64 `json:"podIpv4CidrSize,omitempty"`

	// SelfLink: Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Status: The status of the nodes in this pool instance.
	//
	// Possible values:
	//   "STATUS_UNSPECIFIED" - Not set.
	//   "PROVISIONING" - The PROVISIONING state indicates the node pool is
	// being created.
	//   "RUNNING" - The RUNNING state indicates the node pool has been
	// created
	// and is fully usable.
	//   "RUNNING_WITH_ERROR" - The RUNNING_WITH_ERROR state indicates the
	// node pool has been created
	// and is partially usable. Some error state has occurred and
	// some
	// functionality may be impaired. Customer may need to reissue a
	// request
	// or trigger a new update.
	//   "RECONCILING" - The RECONCILING state indicates that some work is
	// actively being done on
	// the node pool, such as upgrading node software. Details can
	// be found in the `statusMessage` field.
	//   "STOPPING" - The STOPPING state indicates the node pool is being
	// deleted.
	//   "ERROR" - The ERROR state indicates the node pool may be unusable.
	// Details
	// can be found in the `statusMessage` field.
	Status string `json:"status,omitempty"`

	// StatusMessage: Additional information about the current
	// status of this
	// node pool instance, if available.
	StatusMessage string `json:"statusMessage,omitempty"`

	// Version: The version of the Kubernetes of this node.
	Version string `json:"version,omitempty"`
}

// NodePoolAutoscalingClusterStatus contains information required by cluster
// autoscaler to adjust the size of the node pool to the current cluster usage.
type NodePoolAutoscalingClusterStatus struct {
	// Autoprovisioned: Can this node pool be deleted automatically.
	Autoprovisioned bool `json:"autoprovisioned,omitempty"`

	// Enabled: Is autoscaling enabled for this node pool.
	Enabled bool `json:"enabled,omitempty"`

	// MaxNodeCount: Maximum number of nodes in the NodePool. Must be >=
	// min_node_count. There
	// has to enough quota to scale up the cluster.
	MaxNodeCount int64 `json:"maxNodeCount,omitempty"`

	// MinNodeCount: Minimum number of nodes in the NodePool. Must be >= 1
	// and <=
	// max_node_count.
	MinNodeCount int64 `json:"minNodeCount,omitempty"`
}

// NodeConfigClusterStatus is the configuration of the node pool.
type NodeConfigClusterStatus struct {
	// Accelerators: A list of hardware accelerators to be attached to each
	// node.
	// See https://cloud.google.com/compute/docs/gpus for more information
	// about
	// support for GPUs.
	Accelerators []*AcceleratorConfigClusterStatus `json:"accelerators,omitempty"`

	// DiskSizeGb: Size of the disk attached to each node, specified in
	// GB.
	// The smallest allowed disk size is 10GB.
	//
	// If unspecified, the default disk size is 100GB.
	DiskSizeGb int64 `json:"diskSizeGb,omitempty"`

	// DiskType: Type of the disk attached to each node (e.g. 'pd-standard'
	// or 'pd-ssd')
	//
	// If unspecified, the default disk type is 'pd-standard'
	DiskType string `json:"diskType,omitempty"`

	// ImageType: The image type to use for this node. Note that for a given
	// image type,
	// the latest version of it will be used.
	ImageType string `json:"imageType,omitempty"`

	// Labels: The map of Kubernetes labels (key/value pairs) to be applied
	// to each node.
	// These will added in addition to any default label(s) that
	// Kubernetes may apply to the node.
	// In case of conflict in label keys, the applied set may differ
	// depending on
	// the Kubernetes version -- it's best to assume the behavior is
	// undefined
	// and conflicts should be avoided.
	// For more information, including usage and the valid values,
	// see:
	// https://kubernetes.io/docs/concepts/overview/working-with-objects
	// /labels/
	Labels map[string]string `json:"labels,omitempty"`

	// LocalSsdCount: The number of local SSD disks to be attached to the
	// node.
	//
	// The limit for this value is dependant upon the maximum number
	// of
	// disks available on a machine per zone.
	// See:
	// https://cloud.google.com/compute/docs/disks/local-ssd#local_ssd_l
	// imits
	// for more information.
	LocalSsdCount int64 `json:"localSsdCount,omitempty"`

	// MachineType: The name of a Google Compute Engine
	// [machine
	// type](https://cloud.google.com/compute/docs/machine-types) (e.g.
	// `n1-standard-1`).
	//
	// If unspecified, the default machine type is
	// `n1-standard-1`.
	MachineType string `json:"machineType,omitempty"`

	// Metadata: The metadata key/value pairs assigned to instances in the
	// cluster.
	//
	// Keys must conform to the regexp [a-zA-Z0-9-_]+ and be less than 128
	// bytes
	// in length. These are reflected as part of a URL in the metadata
	// server.
	// Additionally, to avoid ambiguity, keys must not conflict with any
	// other
	// metadata keys for the project or be one of the reserved keys:
	//  "cluster-location"
	//  "cluster-name"
	//  "cluster-uid"
	//  "configure-sh"
	//  "containerd-configure-sh"
	//  "enable-oslogin"
	//  "gci-ensure-gke-docker"
	//  "gci-update-strategy"
	//  "instance-template"
	//  "kube-env"
	//  "startup-script"
	//  "user-data"
	//  "disable-address-manager"
	//  "windows-startup-script-ps1"
	//  "common-psm1"
	//  "k8s-node-setup-psm1"
	//  "install-ssh-psm1"
	//  "user-profile-psm1"
	//  "serial-port-logging-enable"
	// Values are free-form strings, and only have meaning as interpreted
	// by
	// the image running in the instance. The only restriction placed on
	// them is
	// that each value's size must be less than or equal to 32 KB.
	//
	// The total size of all keys and values must be less than 512 KB.
	Metadata map[string]string `json:"metadata,omitempty"`

	// MinCpuPlatform: Minimum CPU platform to be used by this instance. The
	// instance may be
	// scheduled on the specified or newer CPU platform. Applicable values
	// are the
	// friendly names of CPU platforms, such as
	// <code>minCpuPlatform: &quot;Intel Haswell&quot;</code>
	// or
	// <code>minCpuPlatform: &quot;Intel Sandy Bridge&quot;</code>. For
	// more
	// information, read [how to specify min
	// CPU
	// platform](https://cloud.google.com/compute/docs/instances/specify-
	// min-cpu-platform)
	MinCPUPlatform string `json:"minCpuPlatform,omitempty"`

	// OauthScopes: The set of Google API scopes to be made available on all
	// of the
	// node VMs under the "default" service account.
	//
	// The following scopes are recommended, but not required, and by
	// default are
	// not included:
	//
	// * `https://www.googleapis.com/auth/compute` is required for
	// mounting
	// persistent storage on your nodes.
	// * `https://www.googleapis.com/auth/devstorage.read_only` is required
	// for
	// communicating with **gcr.io**
	// (the [Google Container Registry](/container-registry/)).
	//
	// If unspecified, no scopes are added, unless Cloud Logging or
	// Cloud
	// Monitoring are enabled, in which case their required scopes will be
	// added.
	OauthScopes []string `json:"oauthScopes,omitempty"`

	// Preemptible: Whether the nodes are created as preemptible VM
	// instances.
	// See:
	// https://cloud.google.com/compute/docs/instances/preemptible for
	// more
	// inforamtion about preemptible VM instances.
	Preemptible bool `json:"preemptible,omitempty"`

	// SandboxConfig: Sandbox configuration for this node.
	SandboxConfig *SandboxConfigClusterStatus `json:"sandboxConfig,omitempty"`

	// ServiceAccount: The Google Cloud Platform Service Account to be used
	// by the node VMs. If
	// no Service Account is specified, the "default" service account is
	// used.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// ShieldedInstanceConfig: Shielded Instance options.
	ShieldedInstanceConfig *ShieldedInstanceConfigClusterStatus `json:"shieldedInstanceConfig,omitempty"`

	// Tags: The list of instance tags applied to all nodes. Tags are used
	// to identify
	// valid sources or targets for network firewalls and are specified
	// by
	// the client during cluster or node pool creation. Each tag within the
	// list
	// must comply with RFC1035.
	Tags []string `json:"tags,omitempty"`

	// Taints: List of kubernetes taints to be applied to each node.
	//
	// For more information, including usage and the valid values,
	// see:
	// https://kubernetes.io/docs/concepts/configuration/taint-and-toler
	// ation/
	Taints []*NodeTaintClusterStatus `json:"taints,omitempty"`
}

// AcceleratorConfigClusterStatus represents a Hardware
// Accelerator request.
type AcceleratorConfigClusterStatus struct {
	// AcceleratorCount: The number of the accelerator cards exposed to an
	// instance.
	AcceleratorCount int64 `json:"acceleratorCount,omitempty"`

	// AcceleratorType: The accelerator type resource name. List of
	// supported accelerators
	// [here](https://cloud.google.com/compute/docs/gpus/#Introduction)
	AcceleratorType string `json:"acceleratorType,omitempty"`
}

// SandboxConfigClusterStatus contains configurations of the sandbox to use for
// the node.
type SandboxConfigClusterStatus struct {
	// Type: Type of the sandbox to use for the node (e.g. 'gvisor')
	Type string `json:"type,omitempty"`
}

// ShieldedInstanceConfigClusterStatus is a set of Shielded Instance options.
type ShieldedInstanceConfigClusterStatus struct {
	// EnableIntegrityMonitoring: Defines whether the instance has integrity
	// monitoring enabled.
	//
	// Enables monitoring and attestation of the boot integrity of the
	// instance.
	// The attestation is performed against the integrity policy baseline.
	// This
	// baseline is initially derived from the implicitly trusted boot image
	// when
	// the instance is created.
	EnableIntegrityMonitoring bool `json:"enableIntegrityMonitoring,omitempty"`

	// EnableSecureBoot: Defines whether the instance has Secure Boot
	// enabled.
	//
	// Secure Boot helps ensure that the system only runs authentic software
	// by
	// verifying the digital signature of all boot components, and halting
	// the
	// boot process if signature verification fails.
	EnableSecureBoot bool `json:"enableSecureBoot,omitempty"`
}

// NodeTaintClusterStatus is a Kubernetes taint is comprised of three fields:
// key, value, and effect. Effect can only be one of three types:  NoSchedule,
// PreferNoSchedule or NoExecute.
type NodeTaintClusterStatus struct {
	// Effect: Effect for taint.
	//
	// Possible values:
	//   "EFFECT_UNSPECIFIED" - Not set
	//   "NO_SCHEDULE" - NoSchedule
	//   "PREFER_NO_SCHEDULE" - PreferNoSchedule
	//   "NO_EXECUTE" - NoExecute
	Effect string `json:"effect,omitempty"`

	// Key: Key for taint.
	Key string `json:"key,omitempty"`

	// Value: Value for taint.
	Value string `json:"value,omitempty"`
}

// NodeManagementClusterStatus defines the set of node management services
// turned on for the node pool.
type NodeManagementClusterStatus struct {
	// AutoRepair: Whether the nodes will be automatically repaired.
	AutoRepair bool `json:"autoRepair,omitempty"`

	// AutoUpgrade: Whether the nodes will be automatically upgraded.
	AutoUpgrade bool `json:"autoUpgrade,omitempty"`

	// UpgradeOptions: Specifies the Auto Upgrade knobs for the node pool.
	UpgradeOptions *AutoUpgradeOptionsClusterStatus `json:"upgradeOptions,omitempty"`
}

// AutoUpgradeOptionsClusterStatus defines the set of options for the user to
// control how the Auto Upgrades will proceed.
type AutoUpgradeOptionsClusterStatus struct {
	// AutoUpgradeStartTime: This field is set when upgrades
	// are about to commence
	// with the approximate start time for the upgrades,
	// in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text format.
	AutoUpgradeStartTime string `json:"autoUpgradeStartTime,omitempty"`

	// Description: This field is set when upgrades are about
	// to commence
	// with the description of the upgrade.
	Description string `json:"description,omitempty"`
}

// A ClusterSpec defines the desired state of a Cluster.
type ClusterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ClusterParameters `json:"forProvider"`
}

// A ClusterStatus represents the observed state of a Cluster.
type ClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Cluster is a managed resource that represents a Google Kubernetes Engine
// cluster.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=".status.atProvider.endpoint"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.forProvider.location"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster items
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}
