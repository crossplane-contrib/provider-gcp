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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Error strings
const (
	errResourceIsNotNodePool = "the managed resource is not a NodePool"
)

// NodePoolObservation is used to show the observed state of the node pool resource on GCP.
type NodePoolObservation struct {
	// Conditions: Which conditions caused the current node pool state.
	Conditions []*StatusCondition `json:"conditions,omitempty"`

	// InstanceGroupUrls: [Output only] The resource URLs of the [managed
	// instance
	// groups](/compute/docs/instance-groups/creating-groups-of-mana
	// ged-instances)
	// associated with this node pool.
	InstanceGroupUrls []string `json:"instanceGroupUrls,omitempty"`

	// PodIpv4CidrSize: [Output only] The pod CIDR block size per node in
	// this node pool.
	PodIpv4CidrSize int64 `json:"podIpv4CidrSize,omitempty"`

	// SelfLink: [Output only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Status: [Output only] The status of the nodes in this pool instance.
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

	// StatusMessage: [Output only] Additional information about the current
	// status of this
	// node pool instance, if available.
	StatusMessage string `json:"statusMessage,omitempty"`
}

// NodePoolParameters define the desired state of a GCP node pool.
type NodePoolParameters struct {
	// Autoscaling: Autoscaler configuration for this NodePool. Autoscaler
	// is enabled
	// only if a valid configuration is present.
	Autoscaling *NodePoolAutoscaling `json:"autoscaling,omitempty"`

	// Config: The node configuration of the pool.
	Config *NodeConfig `json:"config,omitempty"`

	// InitialNodeCount: The initial node count for the pool. You must
	// ensure that your
	// Compute Engine <a href="/compute/docs/resource-quotas">resource
	// quota</a>
	// is sufficient for this number of instances. You must also have
	// available
	// firewall and routes quota.
	InitialNodeCount int64 `json:"initialNodeCount,omitempty"`

	// Management: NodeManagement configuration for this NodePool.
	Management *NodeManagement `json:"management,omitempty"`

	// MaxPodsConstraint: The constraint on the maximum number of pods that
	// can be run
	// simultaneously on a node in the node pool.
	MaxPodsConstraint *MaxPodsConstraint `json:"maxPodsConstraint,omitempty"`

	// Name: The name of the node pool.
	Name string `json:"name,omitempty"`

	// Version: The version of the Kubernetes of this node.
	Version string `json:"version,omitempty"`
}

// NodeConfig is parameters that describe the nodes in a cluster.
type NodeConfig struct {
	// Accelerators: A list of hardware accelerators to be attached to each
	// node.
	// See https://cloud.google.com/compute/docs/gpus for more information
	// about
	// support for GPUs.
	Accelerators []*AcceleratorConfig `json:"accelerators,omitempty"`

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
	// type](/compute/docs/machine-types) (e.g.
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
	//  "enable-os-login"
	//  "gci-update-strategy"
	//  "gci-ensure-gke-docker"
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
	//
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
	// information about preemptible VM instances.
	Preemptible bool `json:"preemptible,omitempty"`

	// ServiceAccount: The Google Cloud Platform Service Account to be used
	// by the node VMs. If
	// no Service Account is specified, the "default" service account is
	// used.
	ServiceAccount string `json:"serviceAccount,omitempty"`

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
	Taints []*NodeTaint `json:"taints,omitempty"`
}

// AcceleratorConfig represents a Hardware
// Accelerator request.
type AcceleratorConfig struct {
	// AcceleratorCount: The number of the accelerator cards exposed to an
	// instance.
	AcceleratorCount int64 `json:"acceleratorCount,omitempty,string"`

	// AcceleratorType: The accelerator type resource name. List of
	// supported accelerators
	// [here](/compute/docs/gpus/#Introduction)
	AcceleratorType string `json:"acceleratorType,omitempty"`
}

// NodeTaint is a Kubernetes taint is comprised of three fields: key, value, and
// effect. Effect can only be one of three types:  NoSchedule, PreferNoSchedule
// or NoExecute.
//
// For more information, including usage and the valid values, see:
// https://kubernetes.io/docs/concepts/configuration/taint-and-toler ation/
type NodeTaint struct {
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

// NodePoolAutoscaling contains information
// required by cluster autoscaler to
// adjust the size of the node pool to the current cluster usage.
type NodePoolAutoscaling struct {
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

// NodeManagement defines the set of node management
// services turned on for the
// node pool.
type NodeManagement struct {
	// AutoRepair: A flag that specifies whether the node auto-repair is
	// enabled for the node
	// pool. If enabled, the nodes in this node pool will be monitored and,
	// if
	// they fail health checks too many times, an automatic repair action
	// will be
	// triggered.
	AutoRepair bool `json:"autoRepair,omitempty"`

	// AutoUpgrade: A flag that specifies whether node auto-upgrade is
	// enabled for the node
	// pool. If enabled, node auto-upgrade helps keep the nodes in your node
	// pool
	// up to date with the latest release version of Kubernetes.
	AutoUpgrade bool `json:"autoUpgrade,omitempty"`

	// UpgradeOptions: Specifies the Auto Upgrade knobs for the node pool.
	UpgradeOptions *AutoUpgradeOptions `json:"upgradeOptions,omitempty"`
}

// AutoUpgradeOptions defines the set of options for
// the user to control how
// the Auto Upgrades will proceed.
type AutoUpgradeOptions struct {
	// AutoUpgradeStartTime: [Output only] This field is set when upgrades
	// are about to commence
	// with the approximate start time for the upgrades,
	// in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text format.
	AutoUpgradeStartTime string `json:"autoUpgradeStartTime,omitempty"`

	// Description: [Output only] This field is set when upgrades are about
	// to commence
	// with the description of the upgrade.
	Description string `json:"description,omitempty"`
}

// A NodePoolSpec defines the desired state of a NodePool.
type NodePoolSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  NodePoolParameters `json:"forProvider,omitempty"`
}

// A NodePoolStatus represents the observed state of a NodePool.
type NodePoolStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     NodePoolObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A NodePool is a managed resource that represents a Google Kubernetes Engine
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
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec   `json:"spec,omitempty"`
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
