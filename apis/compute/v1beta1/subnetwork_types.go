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

package v1beta1

import (
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
)

// Error strings
const (
	errResourceIsNotSubnetwork = "the managed resource is not a Subnetwork"
)

// NetworkURIReferencerForSubnetwork is an attribute referencer that resolves
// network uri from a referenced Network and assigns it to a subnetwork
type NetworkURIReferencerForSubnetwork struct {
	NetworkURIReferencer `json:",inline"`
}

// Assign assigns the retrieved network uri to a subnetwork object
func (v *NetworkURIReferencerForSubnetwork) Assign(res resource.CanReference, value string) error {
	subnetwork, ok := res.(*Subnetwork)
	if !ok {
		return errors.New(errResourceIsNotSubnetwork)
	}

	subnetwork.Spec.ForProvider.Network = &value
	return nil
}

// SubnetworkParameters define the desired state of a Google Compute Engine VPC
// Subnetwork. Most fields map directly to a Subnetwork:
// https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks
type SubnetworkParameters struct {
	// IPCIDRRange: The range of internal addresses that are owned by this
	// subnetwork. Provide this property when you create the subnetwork. For
	// example, 10.0.0.0/8 or 192.168.0.0/16. Ranges must be unique and
	// non-overlapping within a network. Only IPv4 is supported. This field
	// can be set only at resource creation time.
	// +immutable
	IPCidrRange string `json:"ipCidrRange"`

	// Network: The URL of the network to which this subnetwork belongs,
	// provided by the client when initially creating the subnetwork. Only
	// networks that are in the distributed mode can have subnetworks. This
	// field can be set only at resource creation time.
	// +immutable
	Network *string `json:"network,omitempty"`

	// NetworkRef references to a Network and retrieves its URI
	NetworkRef *NetworkURIReferencerForSubnetwork `json:"networkRef,omitempty" resource:"attributereferencer"`

	// Region: URL of the region where the Subnetwork resides. This field
	// can be set only at resource creation time.
	// +optional
	// +immutable
	Region string `json:"region"`

	// Description: An optional description of this resource. Provide this
	// property when you create the resource. This field can be set only at
	// resource creation time.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// EnableFlowLogs: Whether to enable flow logging for this subnetwork.
	// If this field is not explicitly set, it will not appear in get
	// listings. If not set the default behavior is to disable flow logging.
	// +optional
	EnableFlowLogs *bool `json:"enableFlowLogs,omitempty"`

	// PrivateIPGoogleAccess: Whether the VMs in this subnet can access
	// Google services without assigned external IP addresses. This field
	// can be both set at resource creation time and updated using
	// setPrivateIPGoogleAccess.
	// +optional
	PrivateIPGoogleAccess *bool `json:"privateIpGoogleAccess,omitempty"`

	// SecondaryIPRanges: An array of configurations for secondary IP ranges
	// for VM instances contained in this subnetwork. The primary IP of such
	// VM must belong to the primary ipCidrRange of the subnetwork. The
	// alias IPs may belong to either primary or secondary ranges. This
	// field can be updated with a patch request.
	// +optional
	SecondaryIPRanges []*SubnetworkSecondaryRange `json:"secondaryIpRanges,omitempty"`
}

// A SubnetworkObservation represents the observed state of a Google Compute
// Engine VPC Subnetwork.
type SubnetworkObservation struct {
	// CreationTimestamp: Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// Fingerprint: Fingerprint of this resource. A hash of the contents
	// stored in this object. This field is used in optimistic locking. This
	// field will be ignored when inserting a Subnetwork. An up-to-date
	// fingerprint must be provided in order to update the Subnetwork,
	// otherwise the request will fail with error 412 conditionNotMet.
	//
	// To see the latest fingerprint, make a get() request to retrieve a
	// Subnetwork.
	Fingerprint string `json:"fingerprint,omitempty"`

	// GatewayAddress: The gateway address for default routes
	// to reach destination addresses outside this subnetwork.
	GatewayAddress string `json:"gatewayAddress,omitempty"`

	// Id: The unique identifier for the resource. This
	// identifier is defined by the server.
	ID uint64 `json:"id,omitempty"`

	// SelfLink: Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`
}

// A SubnetworkSecondaryRange defines the state of a Google Compute Engine
// VPC Subnetwork secondary range.
type SubnetworkSecondaryRange struct {
	// IPCIDRRange: The range of IP addresses belonging to this subnetwork
	// secondary range. Provide this property when you create the
	// subnetwork. Ranges must be unique and non-overlapping with all
	// primary and secondary IP ranges within a network. Only IPv4 is
	// supported.
	IPCidrRange string `json:"ipCidrRange"`

	// RangeName: The name associated with this subnetwork secondary range,
	// used when adding an alias IP range to a VM instance. The name must be
	// 1-63 characters long, and comply with RFC1035. The name must be
	// unique within the subnetwork.
	RangeName string `json:"rangeName"`
}

// A SubnetworkSpec defines the desired state of a Subnetwork.
type SubnetworkSpec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           SubnetworkParameters `json:"forProvider"`
}

// A SubnetworkStatus represents the observed state of a Subnetwork.
type SubnetworkStatus struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              SubnetworkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Subnetwork is a managed resource that represents a Google Compute Engine
// VPC Subnetwork.
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type Subnetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetworkSpec   `json:"spec"`
	Status SubnetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubnetworkList contains a list of Subnetwork.
type SubnetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subnetwork `json:"items"`
}
