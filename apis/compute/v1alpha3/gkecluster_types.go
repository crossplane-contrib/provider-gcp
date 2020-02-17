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

package v1alpha3

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
)

// Cluster states.
const (
	ClusterStateProvisioning = "PROVISIONING"
	ClusterStateRunning      = "RUNNING"
	ClusterStateError        = "ERROR"
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
	v1beta1.NetworkURIReferencer `json:",inline"`
}

// Assign assigns the retrieved network uri to GKECluster
func (v *NetworkURIReferencerForGKECluster) Assign(res resource.CanReference, value string) error {
	gke, ok := res.(*GKECluster)
	if !ok {
		return errors.New(errResourceIsNotGKECluster)
	}

	gke.Spec.Network = value
	return nil
}

// SubnetworkURIReferencerForGKECluster is an attribute referencer that resolves
// subnetwork uri from a referenced Subnetwork and assigns it to a GKECluster
type SubnetworkURIReferencerForGKECluster struct {
	v1beta1.SubnetworkURIReferencer `json:",inline"`
}

// Assign assigns the retrieved subnetwork uri to a GKECluster
func (v *SubnetworkURIReferencerForGKECluster) Assign(res resource.CanReference, value string) error {
	gke, ok := res.(*GKECluster)
	if !ok {
		return errors.New(errResourceIsNotGKECluster)
	}

	gke.Spec.Subnetwork = value
	return nil
}

// GKEClusterParameters define the desired state of a Google Kubernetes Engine
// cluster.
type GKEClusterParameters struct {
	// ClusterVersion is the initial Kubernetes version for this cluster.
	// Users may specify either explicit versions offered by Kubernetes Engine
	// or version aliases, for example "latest", "1.X", or "1.X.Y". Leave unset
	// to use the default version.
	// +optional
	ClusterVersion string `json:"clusterVersion,omitempty"`

	// Labels for the cluster to use to annotate any related Google Compute
	// Engine resources.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// MachineType is the name of a Google Compute Engine machine type (e.g.
	// n1-standard-1). If unspecified the default machine type is n1-standard-1.
	// +optional
	MachineType string `json:"machineType,omitempty"`

	// NumNodes is the number of nodes to create in this cluster. You must
	// ensure that your Compute Engine resource quota is sufficient for this
	// number of instances. You must also have available firewall and routes
	// quota.
	NumNodes int64 `json:"numNodes"`

	// TODO(negz): Does setting the zone even do anything? The Google API docs
	// state that the field is output only.

	// Zone specifies the name of the Google Compute Engine zone in which this
	// cluster resides.
	// +optional
	Zone string `json:"zone,omitempty"`

	// Scopes are the set of Google API scopes to be made available on all of
	// the node VMs under the "default" service account.
	// +optional
	Scopes []string `json:"scopes,omitempty"`

	// Network is the name of the Google Compute Engine network to which the
	// cluster is connected. If left unspecified, the default network will be
	// used.
	// +optional
	Network string `json:"network,omitempty"`

	// NetworkRef references to a Network and retrieves its URI
	NetworkRef *NetworkURIReferencerForGKECluster `json:"networkRef,omitempty" resource:"attributereferencer"`

	// Subnetwork is the name of the Google Compute Engine subnetwork to which
	// the cluster is connected.
	// +optional
	Subnetwork string `json:"subnetwork,omitempty"`

	// SubnetworkRef references to a Subnetwork and retrieves its URI
	SubnetworkRef *SubnetworkURIReferencerForGKECluster `json:"subnetworkRef,omitempty" resource:"attributereferencer"`

	// EnableIPAlias determines whether Alias IPs will be used for pod IPs in
	// the cluster.
	// +optional
	EnableIPAlias bool `json:"enableIPAlias,omitempty"`

	// CreateSubnetwork determines whether a new subnetwork will be created
	// automatically for the cluster. Only applicable when EnableIPAlias is
	// true.
	// +optional
	CreateSubnetwork bool `json:"createSubnetwork,omitempty"`

	// NodeIPV4CIDR specifies the IP address range of the instance IPs in this
	// cluster. This is applicable only if CreateSubnetwork is true. Omit this
	// field to have a range chosen with the default size. Set it to a netmask
	// (e.g. /24) to have a range chosen with a specific netmask.
	// +optional
	NodeIPV4CIDR string `json:"nodeIPV4CIDR,omitempty"`

	// ClusterIPV4CIDR specifies the IP address range of the pod IPs in this
	// cluster. This is applicable only if EnableIPAlias is true. Omit this
	// field to have a range chosen with the default size. Set it to a netmask
	// (e.g. /24) to have a range chosen with a specific netmask.
	// +optional
	ClusterIPV4CIDR string `json:"clusterIPV4CIDR,omitempty"`

	// ClusterSecondaryRangeName specifies the name of the secondary range to be
	// used for the cluster CIDR block. The secondary range will be used for pod
	// IP addresses. This must be an existing secondary range associated with
	// the cluster subnetwork.
	// +optional
	ClusterSecondaryRangeName string `json:"clusterSecondaryRangeName,omitempty"`

	// ServiceIPV4CIDR specifies the IP address range of service IPs in this
	// cluster. This is applicable only if EnableIPAlias is true. Omit this
	// field to have a range chosen with the default size. Set it to a netmask
	// (e.g. /24) to have a range chosen with a specific netmask.
	// +optional
	ServiceIPV4CIDR string `json:"serviceIPV4CIDR,omitempty"`

	// ServicesSecondaryRangeName specifies the name of the secondary range to
	// be used as for the services CIDR block. The secondary range will be used
	// for service ClusterIPs. This must be an existing secondary range
	// associated with the cluster subnetwork.
	ServicesSecondaryRangeName string `json:"servicesSecondaryRangeName,omitempty"`
}

// A GKEClusterSpec defines the desired state of a GKECluster.
type GKEClusterSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	GKEClusterParameters         `json:",inline"`
}

// A GKEClusterStatus represents the observed state of a GKECluster.
type GKEClusterStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// ClusterName is the name of this GKE cluster. The name is automatically
	// generated by Crossplane.
	ClusterName string `json:"clusterName,omitempty"`

	// Endpoint of the GKE cluster used in connection strings.
	Endpoint string `json:"endpoint,omitempty"`

	// State of this GKE cluster.
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true

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

	Spec   GKEClusterSpec   `json:"spec"`
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
