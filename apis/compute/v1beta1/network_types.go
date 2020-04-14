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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// NetworkParameters define the desired state of a Google Compute Engine VPC
// Network. Most fields map directly to a Network:
// https://cloud.google.com/compute/docs/reference/rest/v1/networks
type NetworkParameters struct {
	// AutoCreateSubnetworks: When set to true, the VPC network is created
	// in "auto" mode. When set to false, the VPC network is created in
	// "custom" mode. When set to nil, the VPC network is created in "legacy"
	// mode which will be deprecated by GCP soon.
	//
	// An auto mode VPC network starts with one subnet per region. Each
	// subnet has a predetermined range as described in Auto mode VPC
	// network IP ranges.
	//
	// This field can only be updated from true to false after creation using
	// switchToCustomMode.
	// +optional
	AutoCreateSubnetworks *bool `json:"autoCreateSubnetworks,omitempty"`

	// Description: An optional description of this resource. Provide this
	// field when you create the resource.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// RoutingConfig: The network-level routing configuration for this
	// network. Used by Cloud Router to determine what type of network-wide
	// routing behavior to enforce.
	// +optional
	RoutingConfig *NetworkRoutingConfig `json:"routingConfig,omitempty"`
}

// A NetworkObservation represents the observed state of a Google Compute Engine
// VPC Network.
type NetworkObservation struct {
	// CreationTimestamp: Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// GatewayIPv4: The gateway address for default routing
	// out of the network, selected by GCP.
	GatewayIPv4 string `json:"gatewayIPv4,omitempty"`

	// Id: The unique identifier for the resource. This
	// identifier is defined by the server.
	ID uint64 `json:"id,omitempty"`

	// Peerings: A list of network peerings for the resource.
	Peerings []*NetworkPeering `json:"peerings,omitempty"`

	// SelfLink: Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Subnetworks: Server-defined fully-qualified URLs for
	// all subnetworks in this VPC network.
	Subnetworks []string `json:"subnetworks,omitempty"`
}

// A NetworkPeering represents the observed state of a Google Compute Engine
// VPC Network Peering.
type NetworkPeering struct {
	// AutoCreateRoutes: This field will be deprecated soon. Use the
	// exchange_subnet_routes field instead. Indicates whether full mesh
	// connectivity is created and managed automatically between peered
	// networks. Currently this field should always be true since Google
	// Compute Engine will automatically create and manage subnetwork routes
	// between two networks when peering state is ACTIVE.
	AutoCreateRoutes bool `json:"autoCreateRoutes,omitempty"`

	// ExchangeSubnetRoutes: Indicates whether full mesh connectivity is
	// created and managed automatically between peered networks. Currently
	// this field should always be true since Google Compute Engine will
	// automatically create and manage subnetwork routes between two
	// networks when peering state is ACTIVE.
	ExchangeSubnetRoutes bool `json:"exchangeSubnetRoutes,omitempty"`

	// Name: Name of this peering. Provided by the client when the peering
	// is created. The name must comply with RFC1035. Specifically, the name
	// must be 1-63 characters long and match regular expression
	// `[a-z]([-a-z0-9]*[a-z0-9])?`. The first character must be a lowercase
	// letter, and all the following characters must be a dash, lowercase
	// letter, or digit, except the last character, which cannot be a dash.
	Name string `json:"name,omitempty"`

	// Network: The URL of the peer network. It can be either full URL or
	// partial URL. The peer network may belong to a different project. If
	// the partial URL does not contain project, it is assumed that the peer
	// network is in the same project as the current network.
	Network string `json:"network,omitempty"`

	// State: State for the peering, either `ACTIVE` or
	// `INACTIVE`. The peering is `ACTIVE` when there's a matching
	// configuration in the peer network.
	//
	// Possible values:
	//   "ACTIVE"
	//   "INACTIVE"
	State string `json:"state,omitempty"`

	// StateDetails: Details about the current state of the
	// peering.
	StateDetails string `json:"stateDetails,omitempty"`
}

// A NetworkRoutingConfig specifies the desired state of a Google Compute
// Engine VPC Network Routing configuration.
type NetworkRoutingConfig struct {
	// RoutingMode: The network-wide routing mode to use. If set to
	// REGIONAL, this network's Cloud Routers will only advertise routes
	// with subnets of this network in the same region as the router. If set
	// to GLOBAL, this network's Cloud Routers will advertise routes with
	// all subnets of this network, across regions.
	//
	// Possible values:
	//   "GLOBAL"
	//   "REGIONAL"
	// +kubebuilder:validation:Enum=GLOBAL;REGIONAL
	RoutingMode string `json:"routingMode"`
}

// A NetworkSpec defines the desired state of a Network.
type NetworkSpec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           NetworkParameters `json:"forProvider"`
}

// A NetworkStatus represents the observed state of a Network.
type NetworkStatus struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              NetworkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Network is a managed resource that represents a Google Compute Engine VPC
// Network.
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec"`
	Status NetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Network.
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}
