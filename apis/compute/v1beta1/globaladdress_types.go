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

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// Known Address statuses.
const (
	StatusInUse     = "IN_USE"
	StatusReserved  = "RESERVED"
	StatusReserving = "RESERVING"
)

// Error strings
const (
	errResourceIsNotGlobalAddress = "the managed resource is not a GlobalAddress"
)

// NetworkURIReferencerForGlobalAddress is an attribute referencer that resolves
// network uri from a referenced Network and assigns it to a global address object
type NetworkURIReferencerForGlobalAddress struct {
	NetworkURIReferencer `json:",inline"`
}

// Assign assigns the retrieved network uri to a global address object
func (v *NetworkURIReferencerForGlobalAddress) Assign(res resource.CanReference, value string) error {
	ga, ok := res.(*GlobalAddress)
	if !ok {
		return errors.New(errResourceIsNotGlobalAddress)
	}

	ga.Spec.ForProvider.Network = &value
	return nil
}

// SubnetworkURIReferencerForGlobalAddress is an attribute referencer that resolves
// subnetwork uri from a referenced Subnetwork and assigns it to a global address object
type SubnetworkURIReferencerForGlobalAddress struct {
	SubnetworkURIReferencer `json:",inline"`
}

// Assign assigns the retrieved network uri to a global address object
func (v *SubnetworkURIReferencerForGlobalAddress) Assign(res resource.CanReference, value string) error {
	ga, ok := res.(*GlobalAddress)
	if !ok {
		return errors.New(errResourceIsNotGlobalAddress)
	}

	ga.Spec.ForProvider.Subnetwork = &value
	return nil
}

// GlobalAddressParameters define the desired state of a Google Compute Engine
// Global Address. Most fields map directly to an Address:
// https://cloud.google.com/compute/docs/reference/rest/v1/globalAddresses
type GlobalAddressParameters struct {
	// Address: The static IP address represented by this resource.
	// +optional
	// +immutable
	Address *string `json:"address,omitempty"`

	// AddressType: The type of address to reserve, either INTERNAL or
	// EXTERNAL. If unspecified, defaults to EXTERNAL.
	//
	// Possible values:
	//   "EXTERNAL"
	//   "INTERNAL"
	//   "UNSPECIFIED_TYPE"
	// +optional
	// +immutable
	// +kubebuilder:validation:Enum=EXTERNAL;INTERNAL;UNSPECIFIED_TYPE
	AddressType *string `json:"addressType,omitempty"`

	// Description: An optional description of this resource.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// IPVersion: The IP version that will be used by this address. Valid
	// options are IPV4 or IPV6.
	//
	// Possible values:
	//   "IPV4"
	//   "IPV6"
	//   "UNSPECIFIED_VERSION"
	// +optional
	// +immutable
	// +kubebuilder:validation:Enum=IPV6;IPV4;UNSPECIFIED_VERSION
	IPVersion *string `json:"ipVersion,omitempty"`

	// Network: The URL of the network in which to reserve the address. This
	// field can only be used with INTERNAL type with the VPC_PEERING
	// purpose.
	// +optional
	// +immutable
	Network *string `json:"network,omitempty"`

	// NetworkRef references to a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkRef *NetworkURIReferencerForGlobalAddress `json:"networkRef,omitempty" resource:"attributereferencer"`

	// PrefixLength: The prefix length if the resource represents an IP
	// range.
	// +optional
	// +immutable
	PrefixLength *int64 `json:"prefixLength,omitempty"`

	// Purpose: The purpose of this resource, which can be one of the
	// following values:
	// - `GCE_ENDPOINT` for addresses that are used by VM instances, alias
	// IP ranges, internal load balancers, and similar resources.
	// - `DNS_RESOLVER` for a DNS resolver address in a subnetwork
	// - `VPC_PEERING` for addresses that are reserved for VPC peer
	// networks.
	// - `NAT_AUTO` for addresses that are external IP addresses
	// automatically reserved for Cloud NAT.
	//
	// Possible values:
	//   "DNS_RESOLVER"
	//   "GCE_ENDPOINT"
	//   "NAT_AUTO"
	//   "VPC_PEERING"
	// +optional
	// +immutable
	// +kubebuilder:validation:Enum=DNS_RESOLVER;GCE_ENDPOINT;NAT_AUTO;VPC_PEERING
	Purpose *string `json:"purpose,omitempty"`

	// Subnetwork: The URL of the subnetwork in which to reserve the
	// address. If an IP address is specified, it must be within the
	// subnetwork's IP range. This field can only be used with INTERNAL type
	// with a GCE_ENDPOINT or DNS_RESOLVER purpose.
	// +optional
	// +immutable
	Subnetwork *string `json:"subnetwork,omitempty"`

	// SubnetworkRef references to a Subnetwork and retrieves its URI
	// +optional
	// +immutable
	SubnetworkRef *SubnetworkURIReferencerForGlobalAddress `json:"subnetworkRef,omitempty" resource:"attributereferencer"`
}

// A GlobalAddressObservation reflects the observed state of a GlobalAddress on GCP.
type GlobalAddressObservation struct {
	// CreationTimestamp in RFC3339 text format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// ID for the resource. This identifier is defined by the server.
	ID uint64 `json:"id,omitempty"`

	// SelfLink: Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Status of the address, which can be one of RESERVING, RESERVED, or
	// IN_USE. An address that is RESERVING is currently in the process of being
	// reserved. A RESERVED address is currently reserved and available to use.
	// An IN_USE address is currently being used by another resource and is not
	// available.
	//
	// Possible values:
	//   "IN_USE"
	//   "RESERVED"
	//   "RESERVING"
	Status string `json:"status,omitempty"`

	// Users that are using this address.
	Users []string `json:"users,omitempty"`
}

// A GlobalAddressSpec defines the desired state of a GlobalAddress.
type GlobalAddressSpec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           GlobalAddressParameters `json:"forProvider"`
}

// A GlobalAddressStatus represents the observed state of a GlobalAddress.
type GlobalAddressStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     GlobalAddressObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}

// A GlobalAddress is a managed resource that represents a Google Compute Engine
// Global Address.
type GlobalAddress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalAddressSpec   `json:"spec"`
	Status GlobalAddressStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlobalAddressList contains a list of GlobalAddress.
type GlobalAddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalAddress `json:"items"`
}
