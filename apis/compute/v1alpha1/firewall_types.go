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

package v1alpha1

import (
	compute "google.golang.org/api/compute/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

var _ = compute.Firewall{}

// FirewallParameters define the desired state of a Google Compute Engine VPC
// Firewall. Most fields map directly to a Firewall:
// https://cloud.google.com/compute/docs/reference/rest/v1/networks
type FirewallParameters struct {
	// Allowed: The list of ALLOW rules specified by this firewall. Each
	// rule specifies a protocol and port-range tuple that describes a
	// permitted connection.
	Allowed []*FirewallAllowed `json:"allowed,omitempty"`

	// Denied: The list of DENY rules specified by this firewall. Each rule
	// specifies a protocol and port-range tuple that describes a denied
	// connection.
	Denied []*FirewallDenied `json:"denied,omitempty"`

	// Description: An optional description of this resource. Provide this
	// field when you create the resource.
	Description string `json:"description,omitempty"`

	// DestinationRanges: If destination ranges are specified, the firewall
	// rule applies only to traffic that has destination IP address in these
	// ranges. These ranges must be expressed in CIDR format. Only IPv4 is
	// supported.
	DestinationRanges []string `json:"destinationRanges,omitempty"`

	// Direction: Direction of traffic to which this firewall applies,
	// either `INGRESS` or `EGRESS`. The default is `INGRESS`. For `INGRESS`
	// traffic, you cannot specify the destinationRanges field, and for
	// `EGRESS` traffic, you cannot specify the sourceRanges or sourceTags
	// fields.
	//
	// Possible values:
	//   "EGRESS"
	//   "INGRESS"
	Direction string `json:"direction,omitempty"`

	// Disabled: Denotes whether the firewall rule is disabled. When set to
	// true, the firewall rule is not enforced and the network behaves as if
	// it did not exist. If this is unspecified, the firewall rule will be
	// enabled.
	Disabled bool `json:"disabled,omitempty"`

	// LogConfig: This field denotes the logging options for a particular
	// firewall rule. If logging is enabled, logs will be exported to
	// Stackdriver.
	LogConfig *FirewallLogConfig `json:"logConfig,omitempty"`

	// Network: URL of the network resource for this firewall rule. If not
	// specified when creating a firewall rule, the default network is
	// used:
	// global/networks/default
	// If you choose to specify this field, you can specify the network as a
	// full or partial URL. For example, the following are all valid URLs:
	//
	// -
	// https://www.googleapis.com/compute/v1/projects/myproject/global/networks/my-network
	// - projects/myproject/global/networks/my-network
	// - global/networks/default
	// +optional
	// +immutable
	Network *string `json:"network,omitempty"`

	// NetworkRef references a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkRef *xpv1.Reference `json:"networkRef,omitempty"`

	// NetworkSelector selects a reference to a Network
	// +optional
	// +immutable
	NetworkSelector *xpv1.Selector `json:"networkSelector,omitempty"`

	// Priority: Priority for this rule. This is an integer between `0` and
	// `65535`, both inclusive. The default value is `1000`. Relative
	// priorities determine which rule takes effect if multiple rules apply.
	// Lower values indicate higher priority. For example, a rule with
	// priority `0` has higher precedence than a rule with priority `1`.
	// DENY rules take precedence over ALLOW rules if they have equal
	// priority. Note that VPC networks have implied rules with a priority
	// of `65535`. To avoid conflicts with the implied rules, use a priority
	// number less than `65535`.
	Priority int64 `json:"priority,omitempty"`

	// SelfLink: [Output Only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// SourceRanges: If source ranges are specified, the firewall rule
	// applies only to traffic that has a source IP address in these ranges.
	// These ranges must be expressed in CIDR format. One or both of
	// sourceRanges and sourceTags may be set. If both fields are set, the
	// rule applies to traffic that has a source IP address within
	// sourceRanges OR a source IP from a resource with a matching tag
	// listed in the sourceTags field. The connection does not need to match
	// both fields for the rule to apply. Only IPv4 is supported.
	SourceRanges []string `json:"sourceRanges,omitempty"`

	// SourceServiceAccounts: If source service accounts are specified, the
	// firewall rules apply only to traffic originating from an instance
	// with a service account in this list. Source service accounts cannot
	// be used to control traffic to an instance's external IP address
	// because service accounts are associated with an instance, not an IP
	// address. sourceRanges can be set at the same time as
	// sourceServiceAccounts. If both are set, the firewall applies to
	// traffic that has a source IP address within the sourceRanges OR a
	// source IP that belongs to an instance with service account listed in
	// sourceServiceAccount. The connection does not need to match both
	// fields for the firewall to apply. sourceServiceAccounts cannot be
	// used at the same time as sourceTags or targetTags.
	SourceServiceAccounts []string `json:"sourceServiceAccounts,omitempty"`

	// SourceTags: If source tags are specified, the firewall rule applies
	// only to traffic with source IPs that match the primary network
	// interfaces of VM instances that have the tag and are in the same VPC
	// network. Source tags cannot be used to control traffic to an
	// instance's external IP address, it only applies to traffic between
	// instances in the same virtual network. Because tags are associated
	// with instances, not IP addresses. One or both of sourceRanges and
	// sourceTags may be set. If both fields are set, the firewall applies
	// to traffic that has a source IP address within sourceRanges OR a
	// source IP from a resource with a matching tag listed in the
	// sourceTags field. The connection does not need to match both fields
	// for the firewall to apply.
	SourceTags []string `json:"sourceTags,omitempty"`

	// TargetServiceAccounts: A list of service accounts indicating sets of
	// instances located in the network that may make network connections as
	// specified in allowed[]. targetServiceAccounts cannot be used at the
	// same time as targetTags or sourceTags. If neither
	// targetServiceAccounts nor targetTags are specified, the firewall rule
	// applies to all instances on the specified network.
	TargetServiceAccounts []string `json:"targetServiceAccounts,omitempty"`

	// TargetTags: A list of tags that controls which instances the firewall
	// rule applies to. If targetTags are specified, then the firewall rule
	// applies only to instances in the VPC network that have one of those
	// tags. If no targetTags are specified, the firewall rule applies to
	// all instances on the specified network.
	TargetTags []string `json:"targetTags,omitempty"`
}

type FirewallAllowed struct {
	// IPProtocol: The IP protocol to which this rule applies. The protocol
	// type is required when creating a firewall rule. This value can either
	// be one of the following well known protocol strings (tcp, udp, icmp,
	// esp, ah, ipip, sctp) or the IP protocol number.
	IPProtocol string `json:"IPProtocol,omitempty"`

	// Ports: An optional list of ports to which this rule applies. This
	// field is only applicable for the UDP or TCP protocol. Each entry must
	// be either an integer or a range. If not specified, this rule applies
	// to connections through any port.
	//
	// Example inputs include: ["22"], ["80","443"], and ["12345-12349"].
	Ports []string `json:"ports,omitempty"`
}

type FirewallDenied struct {
	// IPProtocol: The IP protocol to which this rule applies. The protocol
	// type is required when creating a firewall rule. This value can either
	// be one of the following well known protocol strings (tcp, udp, icmp,
	// esp, ah, ipip, sctp) or the IP protocol number.
	IPProtocol string `json:"IPProtocol,omitempty"`

	// Ports: An optional list of ports to which this rule applies. This
	// field is only applicable for the UDP or TCP protocol. Each entry must
	// be either an integer or a range. If not specified, this rule applies
	// to connections through any port.
	//
	// Example inputs include: ["22"], ["80","443"], and ["12345-12349"].
	Ports []string `json:"ports,omitempty"`
}

// FirewallLogConfig: The available logging options for a firewall rule.
type FirewallLogConfig struct {
	// Enable: This field denotes whether to enable logging for a particular
	// firewall rule.
	Enable bool `json:"enable,omitempty"`
}

// A FirewallObservation represents the observed state of a Google Compute Engine
// VPC Firewall.
type FirewallObservation struct {
	// CreationTimestamp: [Output Only] Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// Id: [Output Only] The unique identifier for the resource. This
	// identifier is defined by the server.
	ID int64 `json:"id,omitempty,string"`

	// Kind: [Output Only] Type of the resource. Always compute#firewall for
	// firewall rules.
	Kind string `json:"kind,omitempty"`
}

// A FirewallSpec defines the desired state of a Firewall.
type FirewallSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       FirewallParameters `json:"forProvider"`
}

// A FirewallStatus represents the observed state of a Firewall.
type FirewallStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          FirewallObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Firewall is a managed resource that represents a Google Compute Engine VPC
// Firewall.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Firewall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallSpec   `json:"spec"`
	Status FirewallStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallList contains a list of Firewall.
type FirewallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Firewall `json:"items"`
}
