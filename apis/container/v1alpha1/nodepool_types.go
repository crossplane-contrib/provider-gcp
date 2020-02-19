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

package v1alpha1

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/stack-gcp/apis/container/v1beta1"
)

// NodePool states.
const (
	NodePoolStateUnspecified  = "STATUS_UNSPECIFIED"
	NodePoolStateProvisioning = "PROVISIONING"
	NodePoolStateRunning      = "RUNNING"
	NodePoolStateRunningError = "RUNNING_WITH_ERROR"
	NodePoolStateReconciling  = "RECONCILING"
	NodePoolStateStopping     = "STOPPING"
	NodePoolStateError        = "ERROR"
)

// Error strings
const (
	errResourceIsNotNodePool = "the managed resource is not a NodePool"
)

// GKEClusterURIReferencerForNodePool resolves references from a NodePool to a
// GKECluster by returning the referenced GKECluster's resource link, e.g.
// projects/projectID/locations/clusterLocation/clusters/clusterName.
type GKEClusterURIReferencerForNodePool struct {
	v1beta1.GKEClusterURIReferencer `json:",inline"`
}

// Assign the PrivateGKECluster field of the supplied resource.CanReference,
// assumed to be a NodePool.
func (v *GKEClusterURIReferencerForNodePool) Assign(res resource.CanReference, value string) error {
	np, ok := res.(*NodePool)
	if !ok {
		return errors.New(errResourceIsNotNodePool)
	}

	np.Spec.ForProvider.Cluster = value
	return nil
}

// NodePoolObservation is used to show the observed state of the GKE Node Pool
// resource on GCP.
type NodePoolObservation struct {
	// Conditions: Which conditions caused the current node pool state.
	Conditions []*v1beta1.StatusCondition `json:"conditions,omitempty"`

	// InstanceGroupUrls: The resource URLs of the [managed
	// instance
	// groups](/compute/docs/instance-groups/creating-groups-of-mana
	// ged-instances)
	// associated with this node pool.
	InstanceGroupUrls []string `json:"instanceGroupUrls,omitempty"`

	// PodIpv4CidrSize: The pod CIDR block size per node in
	// this node pool.
	PodIpv4CidrSize int64 `json:"podIpv4CidrSize,omitempty"`

	// Management: NodeManagement configuration for this NodePool.
	Management *NodeManagementStatus `json:"management,omitempty"`

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
}

// NodePoolParameters define the desired state of a Google Kubernetes Engine
// node pool.
type NodePoolParameters struct {
	// NOTE(hasheddan): Cluster is marked as omitempty but is not optional. It
	// will either be assigned a value directly or set from the ClusterRef.

	// Cluster: The resource link for the GKE cluster to which the NodePool will
	// attach. Must be of format
	// projects/projectID/locations/clusterLocation/clusters/clusterName. Must
	// be supplied if ClusterRef is not.
	// +immutable
	Cluster string `json:"cluster,omitempty"`

	// ClusterRef sets the Cluster field by resolving the resource link of the
	// referenced Crossplane GKECluster managed resource. Must be supplied in
	// Cluster is not.
	// +immutable
	// +optional
	ClusterRef *GKEClusterURIReferencerForNodePool `json:"clusterRef,omitempty" resource:"attributereferencer"`

	// NOTE(hasheddan): Autoscaling can only be updated via setAutoscaling
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters.nodePools/setAutoscaling

	// NOTE(hasheddan): from GCP: If the current node pool size is lower than
	// the specified minimum or greater than the specified maximum when you
	// enable autoscaling, the autoscaler waits to take effect until a new node
	// is needed in the node pool or until a node can be safely deleted from the
	// node pool.
	// https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-autoscaler

	// Autoscaling: Autoscaler configuration for this NodePool. Autoscaler
	// is enabled
	// only if a valid configuration is present.
	Autoscaling *NodePoolAutoscaling `json:"autoscaling,omitempty"`

	// Config: The node configuration of the pool.
	Config *NodeConfig `json:"config,omitempty"`

	// NOTE(hasheddan): InitialNodeCount is only reflected in the
	// container.NodePool if it is specified on creation. If omitted at creation
	// time, it will never be reflected in container.NodePool.

	// InitialNodeCount: The initial node count for the pool. You must
	// ensure that your
	// Compute Engine <a href="/compute/docs/resource-quotas">resource
	// quota</a>
	// is sufficient for this number of instances. You must also have
	// available
	// firewall and routes quota.
	// +immutable
	// +optional
	InitialNodeCount *int64 `json:"initialNodeCount,omitempty"`

	// Locations: The list of Google Compute Engine
	// [zones](/compute/docs/zones#available)
	// in which the NodePool's nodes should be located.
	// +optional
	Locations []string `json:"locations,omitempty"`

	// NOTE(hasheddan): NodeManagement can only be updated via
	// setManagement
	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters.nodePools/setManagement

	// Management: NodeManagement configuration for this NodePool.
	Management *NodeManagementSpec `json:"management,omitempty"`

	// MaxPodsConstraint: The constraint on the maximum number of pods that
	// can be run
	// simultaneously on a node in the node pool.
	// +immutable
	MaxPodsConstraint *v1beta1.MaxPodsConstraint `json:"maxPodsConstraint,omitempty"`

	// Version: The version of the Kubernetes of this node.
	// +optional
	Version *string `json:"version,omitempty"`
}

// NodePoolAutoscaling contains information
// required by cluster autoscaler to
// adjust the size of the node pool to the current cluster usage.
type NodePoolAutoscaling struct {
	// Autoprovisioned: Can this node pool be deleted automatically.
	// +optional
	Autoprovisioned *bool `json:"autoprovisioned,omitempty"`

	// Enabled: Is autoscaling enabled for this node pool.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// MaxNodeCount: Maximum number of nodes in the NodePool. Must be >=
	// min_node_count. There
	// has to enough quota to scale up the cluster.
	// +optional
	MaxNodeCount *int64 `json:"maxNodeCount,omitempty"`

	// MinNodeCount: Minimum number of nodes in the NodePool. Must be >= 1
	// and <=
	// max_node_count.
	// +optional
	MinNodeCount *int64 `json:"minNodeCount,omitempty"`
}

// NodeConfig is parameters that describe the nodes in a cluster.
type NodeConfig struct {
	// Accelerators: A list of hardware accelerators to be attached to each
	// node.
	// See https://cloud.google.com/compute/docs/gpus for more information
	// about
	// support for GPUs.
	// +immutable
	Accelerators []*AcceleratorConfig `json:"accelerators,omitempty"`

	// DiskSizeGb: Size of the disk attached to each node, specified in
	// GB.
	// The smallest allowed disk size is 10GB.
	//
	// If unspecified, the default disk size is 100GB.
	// +immutable
	// +optional
	DiskSizeGb *int64 `json:"diskSizeGb,omitempty"`

	// DiskType: Type of the disk attached to each node (e.g. 'pd-standard'
	// or 'pd-ssd')
	//
	// If unspecified, the default disk type is 'pd-standard'
	// +immutable
	// +optional
	DiskType *string `json:"diskType,omitempty"`

	// ImageType: The image type to use for this node. Note that for a given
	// image type,
	// the latest version of it will be used.
	// +optional
	ImageType *string `json:"imageType,omitempty"`

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
	// +immutable
	// +optional
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
	// +immutable
	// +optional
	LocalSsdCount *int64 `json:"localSsdCount,omitempty"`

	// MachineType: The name of a Google Compute Engine
	// [machine
	// type](/compute/docs/machine-types) (e.g.
	// `n1-standard-1`).
	//
	// If unspecified, the default machine type is
	// `n1-standard-1`.
	// +immutable
	// +optional
	MachineType *string `json:"machineType,omitempty"`

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
	// +immutable
	// +optional
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
	// +immutable
	// +optional
	MinCPUPlatform *string `json:"minCpuPlatform,omitempty"`

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
	// +immutable
	// +optional
	OauthScopes []string `json:"oauthScopes,omitempty"`

	// Preemptible: Whether the nodes are created as preemptible VM
	// instances.
	// See:
	// https://cloud.google.com/compute/docs/instances/preemptible for
	// more
	// inforamtion about preemptible VM instances.
	// +immutable
	// +optional
	Preemptible *bool `json:"preemptible,omitempty"`

	// SandboxConfig: Sandbox configuration for this node.
	// +immutable
	// +optional
	SandboxConfig *SandboxConfig `json:"sandboxConfig,omitempty"`

	// ServiceAccount: The Google Cloud Platform Service Account to be used
	// by the node VMs. If
	// no Service Account is specified, the "default" service account is
	// used.
	// +immutable
	// +optional
	ServiceAccount *string `json:"serviceAccount,omitempty"`

	// ShieldedInstanceConfig: Shielded Instance options.
	// +immutable
	// +optional
	ShieldedInstanceConfig *ShieldedInstanceConfig `json:"shieldedInstanceConfig,omitempty"`

	// Tags: The list of instance tags applied to all nodes. Tags are used
	// to identify
	// valid sources or targets for network firewalls and are specified
	// by
	// the client during cluster or node pool creation. Each tag within the
	// list
	// must comply with RFC1035.
	// +immutable
	// +optional
	Tags []string `json:"tags,omitempty"`

	// Taints: List of kubernetes taints to be applied to each node.
	//
	// For more information, including usage and the valid values,
	// see:
	// https://kubernetes.io/docs/concepts/configuration/taint-and-toler
	// ation/
	// +immutable
	// +optional
	Taints []*NodeTaint `json:"taints,omitempty"`

	// WorkloadMetadataConfig: The workload metadata configuration for this
	// node.
	// +optional
	WorkloadMetadataConfig *WorkloadMetadataConfig `json:"workloadMetadataConfig,omitempty"`
}

// AcceleratorConfig represents a Hardware Accelerator request.
type AcceleratorConfig struct {
	// AcceleratorCount: The number of the accelerator cards exposed to an
	// instance.
	AcceleratorCount int64 `json:"acceleratorCount,omitempty,string"`

	// AcceleratorType: The accelerator type resource name. List of
	// supported accelerators
	// [here](/compute/docs/gpus/#Introduction)
	AcceleratorType string `json:"acceleratorType,omitempty"`
}

// SandboxConfig contains configurations of the sandbox to use for the node.
type SandboxConfig struct {
	// SandboxType: Type of the sandbox to use for the node (e.g. 'gvisor')
	SandboxType string `json:"sandboxType"`
}

// ShieldedInstanceConfig is a set of Shielded Instance options.
type ShieldedInstanceConfig struct {
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
	// +optional
	EnableIntegrityMonitoring *bool `json:"enableIntegrityMonitoring,omitempty"`

	// EnableSecureBoot: Defines whether the instance has Secure Boot
	// enabled.
	//
	// Secure Boot helps ensure that the system only runs authentic software
	// by
	// verifying the digital signature of all boot components, and halting
	// the
	// boot process if signature verification fails.
	// +optional
	EnableSecureBoot *bool `json:"enableSecureBoot,omitempty"`
}

// NodeTaint is a Kubernetes taint is comprised of three fields: key, value, and
// effect. Effect can only be one of three types:  NoSchedule, PreferNoSchedule
// or NoExecute.
//
// For more information, including usage and the valid values,
// see:
// https://kubernetes.io/docs/concepts/configuration/taint-and-toler
// ation/
type NodeTaint struct {
	// Effect: Effect for taint.
	//
	// Possible values:
	//   "EFFECT_UNSPECIFIED" - Not set
	//   "NO_SCHEDULE" - NoSchedule
	//   "PREFER_NO_SCHEDULE" - PreferNoSchedule
	//   "NO_EXECUTE" - NoExecute
	Effect string `json:"effect"`

	// Key: Key for taint.
	Key string `json:"key"`

	// Value: Value for taint.
	Value string `json:"value"`
}

// WorkloadMetadataConfig defines the metadata configuration to expose to
// workloads on the node pool.
type WorkloadMetadataConfig struct {
	// NodeMetadata: NodeMetadata is the configuration for how to expose
	// metadata to the
	// workloads running on the node.
	//
	// Possible values:
	//   "UNSPECIFIED" - Not set.
	//   "SECURE" - Prevent workloads not in hostGKECluster from accessing
	// certain VM metadata,
	// specifically kube-env, which contains Kubelet credentials, and
	// the
	// instance identity token.
	//
	// Metadata concealment is a temporary security solution available while
	// the
	// bootstrapping process for cluster nodes is being redesigned
	// with
	// significant security improvements.  This feature is scheduled to
	// be
	// deprecated in the future and later removed.
	//   "EXPOSE" - Expose all VM metadata to pods.
	//   "GKE_METADATA_SERVER" - Run the GKE Metadata Server on this node.
	// The GKE Metadata Server exposes
	// a metadata API to workloads that is compatible with the V1
	// Compute
	// Metadata APIs exposed by the Compute Engine and App Engine
	// Metadata
	// Servers. This feature can only be enabled if Workload Identity is
	// enabled
	// at the cluster level.
	NodeMetadata string `json:"nodeMetadata"`
}

// NodeManagementSpec defines the desired set of node management services turned on
// for the node pool.
type NodeManagementSpec struct {
	// AutoRepair: Whether the nodes will be automatically repaired.
	// +optional
	AutoRepair *bool `json:"autoRepair,omitempty"`

	// AutoUpgrade: Whether the nodes will be automatically upgraded.
	// +optional
	AutoUpgrade *bool `json:"autoUpgrade,omitempty"`
}

// NodeManagementStatus defines the observed set of node management services turned on
// for the node pool.
type NodeManagementStatus struct {
	// UpgradeOptions: Specifies the Auto Upgrade knobs for the node pool.
	UpgradeOptions *AutoUpgradeOptions `json:"upgradeOptions,omitempty"`
}

// AutoUpgradeOptions defines the set of options for the user to control how the
// Auto Upgrades will proceed.
type AutoUpgradeOptions struct {
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

// A NodePoolSpec defines the desired state of a NodePool.
type NodePoolSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  NodePoolParameters `json:"forProvider"`
}

// A NodePoolStatus represents the observed state of a NodePool.
type NodePoolStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     NodePoolObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A NodePool is a managed resource that represents a Google Kubernetes Engine
// node pool.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="CLUSTER-NAME",type="string",JSONPath=".spec.forProvider.cluster"
// +kubebuilder:printcolumn:name="NODE-POOL-CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec   `json:"spec"`
	Status NodePoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodePoolList contains a list of NodePool items
type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePool `json:"items"`
}

// A NodePoolClassSpecTemplate is a template for the spec of a dynamically
// provisioned NodePool.
type NodePoolClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	NodePoolParameters                `json:",inline"`
}

// +kubebuilder:object:root=true

// A NodePoolClass is a resource class. It defines the desired spec of
// resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type NodePoolClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// NodePool.
	SpecTemplate NodePoolClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// NodePoolClassList contains a list of cloud memorystore resource classes.
type NodePoolClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePoolClass `json:"items"`
}
