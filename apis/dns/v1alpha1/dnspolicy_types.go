/*
Copyright 2021 The Crossplane Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// The PolicyParameters define the desired state of a Policy
type PolicyParameters struct {

	// Name: User-assigned name for this policy.
	Name string `json:"name"`

	// AlternativeNameServerConfig: Sets an alternative name server for the associated networks.
	// When specified, all DNS queries are forwarded to a name server that you choose.
	// Names such as .internal are not available when an alternative name server is specified.
	// +optional
	AlternativeNameServerConfig *PolicyAlternativeNameServerConfig `json:"alternativeNameServerConfig,omitempty"`

	// Description: A mutable string of at most 1024 characters associated with this resource for the user's convenience.
	// Has no effect on the policy's function.
	// +optional
	Description *string `json:"description,omitempty"`

	// EnableInboundForwarding: Allows networks bound to this policy to receive DNS queries sent by VMs or applications over VPN connections.
	// When enabled, a virtual IP address is allocated from each of the subnetworks that are bound to this policy.
	// +optional
	EnableInboundForwarding *bool `json:"enableInboundForwarding,omitempty"`

	// EnableLogging: Controls whether logging is enabled for the networks bound to this policy.
	// Defaults to no logging if not set.
	// +optional
	EnableLogging *bool `json:"enableLogging,omitempty"`

	// Networks: List of network names specifying networks to which this policy is applied.
	// +optional
	Networks *[]PolicyNetwork `json:"networks,omitempty"`
}

// The PolicyAlternativeNameServerConfig Sets an alternative name server for the associated networks.
// When specified, all DNS queries are forwarded to a name server that you choose.
type PolicyAlternativeNameServerConfig struct {

	// TargetNameServers: Sets an alternative name server for the associated
	// networks. When specified, all DNS queries are forwarded to a name
	// server that you choose. Names such as .internal are not available
	// when an alternative name server is specified.
	TargetNameServers []PolicyAlternativeNameServerConfigTargetNameServer `json:"targetNameServers"`
}

// A PolicyAlternativeNameServerConfigTargetNameServer has the below fields.
type PolicyAlternativeNameServerConfigTargetNameServer struct {

	// ForwardingPath: Forwarding path for this TargetNameServer. If unset or set to DEFAULT, Cloud DNS makes forwarding decisions based on  address ranges; that is, RFC1918 addresses go to the VPC network, non-RFC1918 addresses go to the internet. When set to PRIVATE, Cloud
	// DNS always sends queries through the VPC network for this target. Possible values:
	// "default" - Cloud DNS makes forwarding decision based on IP address ranges; that is, RFC1918 addresses forward to the target through the VPC and non-RFC1918 addresses forward to the target through the internet
	// "private" - Cloud DNS always forwards to this target through the VPC.
	ForwardingPath *string `json:"forwardingPath,omitempty"`

	// Ipv4Address: IPv4 address to forward to.
	Ipv4Address string `json:"ipv4Address"`
}

// A PolicyNetwork struct has the field NetworkURL
type PolicyNetwork struct {

	// NetworkUrl: The fully qualified URL of the VPC network to bind to.
	// This should be formatted like https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}
	NetworkURL string `json:"networkUrl"`
}

// The PolicyObservation is used to show the observed state of the Policy
type PolicyObservation struct {

	// Id: Unique identifier for the resource; defined by the server (output only).
	ID *uint64 `json:"id,omitempty,string"`
}

// The PolicySpec defines the desired state of a DNSPolicy.
type PolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PolicyParameters `json:"forProvider"`
}

// The PolicyStatus represents the observed state of a DNSPolicy.
type PolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Policy is a collection of DNS rules applied to one or more
// Virtual Private Cloud resources.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="DNS NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec"`
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// The PolicyList contains a list of DNSPolicy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
