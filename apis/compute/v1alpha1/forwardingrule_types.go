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

var _ = compute.ForwardingRule{}

// ForwardingRuleParameters define the desired state of a Google Compute Engine VPC
// ForwardingRule. Most fields map directly to a ForwardingRule:
// https://cloud.google.com/compute/docs/reference/rest/v1/networks
type ForwardingRuleParameters struct {
	// IPAddress: IP address that this forwarding rule serves. When a client
	// sends traffic to this IP address, the forwarding rule directs the
	// traffic to the target that you specify in the forwarding rule.
	//
	// If you don't specify a reserved IP address, an ephemeral IP address
	// is assigned. Methods for specifying an IP address:
	//
	// * IPv4 dotted decimal, as in `100.1.2.3` * Full URL, as in
	// https://www.googleapis.com/compute/v1/projects/project_id/regions/region/addresses/address-name * Partial URL or by name, as in: * projects/project_id/regions/region/addresses/address-name * regions/region/addresses/address-name * global/addresses/address-name * address-name
	//
	// The loadBalancingScheme and the forwarding rule's target determine
	// the type of IP address that you can use. For detailed information,
	// refer to [IP address
	// specifications](/load-balancing/docs/forwarding-rule-concepts#ip_addre
	// ss_specifications).
	IPAddress string `json:"IPAddress,omitempty"`

	// IPProtocol: The IP protocol to which this rule applies. For protocol
	// forwarding, valid options are TCP, UDP, ESP, AH, SCTP or ICMP.
	//
	// For Internal TCP/UDP Load Balancing, the load balancing scheme is
	// INTERNAL, and one of TCP or UDP are valid. For Traffic Director, the
	// load balancing scheme is INTERNAL_SELF_MANAGED, and only TCPis valid.
	// For Internal HTTP(S) Load Balancing, the load balancing scheme is
	// INTERNAL_MANAGED, and only TCP is valid. For HTTP(S), SSL Proxy, and
	// TCP Proxy Load Balancing, the load balancing scheme is EXTERNAL and
	// only TCP is valid. For Network TCP/UDP Load Balancing, the load
	// balancing scheme is EXTERNAL, and one of TCP or UDP is valid.
	//
	// Possible values:
	//   "AH"
	//   "ESP"
	//   "ICMP"
	//   "SCTP"
	//   "TCP"
	//   "UDP"
	IPProtocol string `json:"IPProtocol,omitempty"`

	// AllPorts: This field is used along with the backend_service field for
	// internal load balancing or with the target field for internal
	// TargetInstance. This field cannot be used with port or portRange
	// fields.
	//
	// When the load balancing scheme is INTERNAL and protocol is TCP/UDP,
	// specify this field to allow packets addressed to any ports will be
	// forwarded to the backends configured with this forwarding rule.
	AllPorts bool `json:"allPorts,omitempty"`

	// AllowGlobalAccess: This field is used along with the backend_service
	// field for internal load balancing or with the target field for
	// internal TargetInstance. If the field is set to TRUE, clients can
	// access ILB from all regions. Otherwise only allows access from
	// clients in the same region as the internal load balancer.
	AllowGlobalAccess bool `json:"allowGlobalAccess,omitempty"`

	// BackendService: This field is only used for INTERNAL load
	// balancing.
	//
	// For internal load balancing, this field identifies the BackendService
	// resource to receive the matched traffic.
	// +optional
	// +immutable
	BackendService *string `json:"backendService,omitempty"`

	// BackendServiceRef references a BackendService
	// +optional
	// +immutable
	BackendServiceRef *xpv1.Reference `json:"backendServiceRef,omitempty"`

	// BackendServiceSelector selects a reference to a BackendService
	// +optional
	// +immutable
	BackendServiceSelector *xpv1.Selector `json:"backendServiceSelector,omitempty"`

	// Description: An optional description of this resource. Provide this
	// property when you create the resource.
	Description string `json:"description,omitempty"`

	// Fingerprint: Fingerprint of this resource. A hash of the contents
	// stored in this object. This field is used in optimistic locking. This
	// field will be ignored when inserting a ForwardingRule. Include the
	// fingerprint in patch request to ensure that you do not overwrite
	// changes that were applied from another concurrent request.
	//
	// To see the latest fingerprint, make a get() request to retrieve a
	// ForwardingRule.
	Fingerprint string `json:"fingerprint,omitempty"`

	// IpVersion: The IP Version that will be used by this forwarding rule.
	// Valid options are IPV4 or IPV6. This can only be specified for an
	// external global forwarding rule.
	//
	// Possible values:
	//   "IPV4"
	//   "IPV6"
	//   "UNSPECIFIED_VERSION"
	IPVersion string `json:"ipVersion,omitempty"`

	// IsMirroringCollector: Indicates whether or not this load balancer can
	// be used as a collector for packet mirroring. To prevent mirroring
	// loops, instances behind this load balancer will not have their
	// traffic mirrored even if a PacketMirroring rule applies to them. This
	// can only be set to true for load balancers that have their
	// loadBalancingScheme set to INTERNAL.
	IsMirroringCollector bool `json:"isMirroringCollector,omitempty"`

	// LoadBalancingScheme: Specifies the forwarding rule type.
	//
	//
	// - EXTERNAL is used for:
	// - Classic Cloud VPN gateways
	// - Protocol forwarding to VMs from an external IP address
	// - The following load balancers: HTTP(S), SSL Proxy, TCP Proxy, and
	// Network TCP/UDP
	// - INTERNAL is used for:
	// - Protocol forwarding to VMs from an internal IP address
	// - Internal TCP/UDP load balancers
	// - INTERNAL_MANAGED is used for:
	// - Internal HTTP(S) load balancers
	// - INTERNAL_SELF_MANAGED is used for:
	// - Traffic Director
	//
	// For more information about forwarding rules, refer to Forwarding rule
	// concepts.
	//
	// Possible values:
	//   "EXTERNAL"
	//   "INTERNAL"
	//   "INTERNAL_MANAGED"
	//   "INTERNAL_SELF_MANAGED"
	//   "INVALID"
	LoadBalancingScheme string `json:"loadBalancingScheme,omitempty"`

	// MetadataFilters: Opaque filter criteria used by Loadbalancer to
	// restrict routing configuration to a limited set of xDS compliant
	// clients. In their xDS requests to Loadbalancer, xDS clients present
	// node metadata. If a match takes place, the relevant configuration is
	// made available to those proxies. Otherwise, all the resources (e.g.
	// TargetHttpProxy, UrlMap) referenced by the ForwardingRule will not be
	// visible to those proxies.
	// For each metadataFilter in this list, if its filterMatchCriteria is
	// set to MATCH_ANY, at least one of the filterLabels must match the
	// corresponding label provided in the metadata. If its
	// filterMatchCriteria is set to MATCH_ALL, then all of its filterLabels
	// must match with corresponding labels provided in the
	// metadata.
	// metadataFilters specified here will be applifed before those
	// specified in the UrlMap that this ForwardingRule
	// references.
	// metadataFilters only applies to Loadbalancers that have their
	// loadBalancingScheme set to INTERNAL_SELF_MANAGED.
	MetadataFilters []*MetadataFilter `json:"metadataFilters,omitempty"`

	// Network: This field is not used for external load balancing.
	//
	// For INTERNAL and INTERNAL_SELF_MANAGED load balancing, this field
	// identifies the network that the load balanced IP should belong to for
	// this Forwarding Rule. If this field is not specified, the default
	// network will be used.
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

	// NetworkTier: This signifies the networking tier used for configuring
	// this load balancer and can only take the following values: PREMIUM,
	// STANDARD.
	//
	// For regional ForwardingRule, the valid values are PREMIUM and
	// STANDARD. For GlobalForwardingRule, the valid value is PREMIUM.
	//
	// If this field is not specified, it is assumed to be PREMIUM. If
	// IPAddress is specified, this value must be equal to the networkTier
	// of the Address.
	//
	// Possible values:
	//   "PREMIUM"
	//   "STANDARD"
	NetworkTier string `json:"networkTier,omitempty"`

	// PortRange: When the load balancing scheme is EXTERNAL,
	// INTERNAL_SELF_MANAGED and INTERNAL_MANAGED, you can specify a
	// port_range. Use with a forwarding rule that points to a target proxy
	// or a target pool. Do not use with a forwarding rule that points to a
	// backend service. This field is used along with the target field for
	// TargetHttpProxy, TargetHttpsProxy, TargetSslProxy, TargetTcpProxy,
	// TargetVpnGateway, TargetPool, TargetInstance.
	//
	// Applicable only when IPProtocol is TCP, UDP, or SCTP, only packets
	// addressed to ports in the specified range will be forwarded to
	// target. Forwarding rules with the same [IPAddress, IPProtocol] pair
	// must have disjoint port ranges.
	//
	// Some types of forwarding target have constraints on the acceptable
	// ports:
	// - TargetHttpProxy: 80, 8080
	// - TargetHttpsProxy: 443
	// - TargetTcpProxy: 25, 43, 110, 143, 195, 443, 465, 587, 700, 993,
	// 995, 1688, 1883, 5222
	// - TargetSslProxy: 25, 43, 110, 143, 195, 443, 465, 587, 700, 993,
	// 995, 1688, 1883, 5222
	// - TargetVpnGateway: 500, 4500
	PortRange string `json:"portRange,omitempty"`

	// Ports: This field is used along with the backend_service field for
	// internal load balancing.
	//
	// When the load balancing scheme is INTERNAL, a list of ports can be
	// configured, for example, ['80'], ['8000','9000']. Only packets
	// addressed to these ports are forwarded to the backends configured
	// with the forwarding rule.
	//
	// If the forwarding rule's loadBalancingScheme is INTERNAL, you can
	// specify ports in one of the following ways:
	//
	// * A list of up to five ports, which can be non-contiguous * Keyword
	// ALL, which causes the forwarding rule to forward traffic on any port
	// of the forwarding rule's protocol.
	Ports []string `json:"ports,omitempty"`

	// ServiceLabel: An optional prefix to the service name for this
	// Forwarding Rule. If specified, the prefix is the first label of the
	// fully qualified service name.
	//
	// The label must be 1-63 characters long, and comply with RFC1035.
	// Specifically, the label must be 1-63 characters long and match the
	// regular expression `[a-z]([-a-z0-9]*[a-z0-9])?` which means the first
	// character must be a lowercase letter, and all following characters
	// must be a dash, lowercase letter, or digit, except the last
	// character, which cannot be a dash.
	//
	// This field is only used for internal load balancing.
	ServiceLabel string `json:"serviceLabel,omitempty"`

	// Subnetwork: This field is only used for INTERNAL load balancing.
	//
	// For internal load balancing, this field identifies the subnetwork
	// that the load balanced IP should belong to for this Forwarding
	// Rule.
	//
	// If the network specified is in auto subnet mode, this field is
	// optional. However, if the network is in custom subnet mode, a
	// subnetwork must be specified.
	// +optional
	// +immutable
	Subnetwork *string `json:"subnetwork,omitempty"`

	// SubnetworkRef references a Subnetwork and retrieves its URI
	// +optional
	// +immutable
	SubnetworkRef *xpv1.Reference `json:"subnetworkRef,omitempty"`

	// SubnetworkSelector selects a reference to a Subnetwork
	// +optional
	// +immutable
	SubnetworkSelector *xpv1.Selector `json:"subnetworkSelector,omitempty"`

	// Target: The URL of the target resource to receive the matched
	// traffic. For regional forwarding rules, this target must live in the
	// same region as the forwarding rule. For global forwarding rules, this
	// target must be a global load balancing resource. The forwarded
	// traffic must be of a type appropriate to the target object. For
	// INTERNAL_SELF_MANAGED load balancing, only targetHttpProxy is valid,
	// not targetHttpsProxy.
	Target *string `json:"target,omitempty"`

	// TargetRef references a Target and retrieves its URI
	// +optional
	// +immutable
	TargetRef *xpv1.Reference `json:"targetRef,omitempty"`

	// TargetSelector selects a reference to a Target
	// +optional
	// +immutable
	TargetSelector *xpv1.Selector `json:"targetSelector,omitempty"`
}

// MetadataFilter: Opaque filter criteria used by loadbalancers to
// restrict routing configuration to a limited set of loadbalancing
// proxies. Proxies and sidecars involved in loadbalancing would
// typically present metadata to the loadbalancers which need to match
// criteria specified here. If a match takes place, the relevant
// configuration is made available to those proxies.
// For each metadataFilter in this list, if its filterMatchCriteria is
// set to MATCH_ANY, at least one of the filterLabels must match the
// corresponding label provided in the metadata. If its
// filterMatchCriteria is set to MATCH_ALL, then all of its filterLabels
// must match with corresponding labels provided in the metadata.
// An example for using metadataFilters would be: if loadbalancing
// involves  Envoys, they will only receive routing configuration when
// values in metadataFilters match values supplied in <a
// href="https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/core/b
// ase.proto#envoy-api-msg-core-node" Node metadata of their XDS
// requests to loadbalancers.
type MetadataFilter struct {
	// FilterLabels: The list of label value pairs that must match labels in
	// the provided metadata based on filterMatchCriteria
	// This list must not be empty and can have at the most 64 entries.
	FilterLabels []*MetadataFilterLabelMatch `json:"filterLabels,omitempty"`

	// FilterMatchCriteria: Specifies how individual filterLabel matches
	// within the list of filterLabels contribute towards the overall
	// metadataFilter match.
	// Supported values are:
	// - MATCH_ANY: At least one of the filterLabels must have a matching
	// label in the provided metadata.
	// - MATCH_ALL: All filterLabels must have matching labels in the
	// provided metadata.
	//
	// Possible values:
	//   "MATCH_ALL"
	//   "MATCH_ANY"
	//   "NOT_SET"
	FilterMatchCriteria string `json:"filterMatchCriteria,omitempty"`
}

// MetadataFilterLabelMatch is a MetadataFilter label name value pairs that
// are expected to match corresponding labels presented as metadata to
// the loadbalancer.
type MetadataFilterLabelMatch struct {
	// Name: Name of metadata label.
	// The name can have a maximum length of 1024 characters and must be at
	// least 1 character long.
	Name string `json:"name,omitempty"`

	// Value: The value of the label must match the specified value.
	// value can have a maximum length of 1024 characters.
	Value string `json:"value,omitempty"`
}

// A ForwardingRuleObservation represents the observed state of a Google Compute Engine
// VPC ForwardingRule.
type ForwardingRuleObservation struct {
	// CreationTimestamp: [Output Only] Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// Id: [Output Only] The unique identifier for the resource. This
	// identifier is defined by the server.
	ID int64 `json:"id,omitempty"`

	// Kind: [Output Only] Type of the resource. Always
	// compute#forwardingRule for Forwarding Rule resources.
	Kind string `json:"kind,omitempty"`

	// Region: [Output Only] URL of the region where the regional forwarding
	// rule resides. This field is not applicable to global forwarding
	// rules. You must specify this field as part of the HTTP request URL.
	// It is not settable as a field in the request body.
	Region string `json:"region,omitempty"`

	// SelfLink: [Output Only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// ServiceName: [Output Only] The internal fully qualified service name
	// for this Forwarding Rule.
	//
	// This field is only used for internal load balancing.
	ServiceName string `json:"serviceName,omitempty"`
}

// A ForwardingRuleSpec defines the desired state of a ForwardingRule.
type ForwardingRuleSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ForwardingRuleParameters `json:"forProvider"`
}

// A ForwardingRuleStatus represents the observed state of a ForwardingRule.
type ForwardingRuleStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ForwardingRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ForwardingRule is a managed resource that represents a Google Compute Engine VPC
// ForwardingRule.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type ForwardingRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ForwardingRuleSpec   `json:"spec"`
	Status ForwardingRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ForwardingRuleList contains a list of ForwardingRule.
type ForwardingRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ForwardingRule `json:"items"`
}
