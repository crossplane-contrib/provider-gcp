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

var _ = compute.Router{}

// RouterParameters define the desired state of a Google Compute Engine VPC
// Router. Most fields map directly to a Router:
// https://cloud.google.com/compute/docs/reference/rest/v1/networks
type RouterParameters struct {
	// Bgp: BGP information specific to this router.
	// +optional
	Bgp *RouterBgp `json:"bgp,omitempty"`

	// BgpPeers: BGP information that must be configured into the routing
	// stack to establish BGP peering. This information must specify the
	// peer ASN and either the interface name, IP address, or peer IP
	// address. Please refer to RFC4273.
	// +optional
	BgpPeers []*RouterBgpPeer `json:"bgpPeers,omitempty"`

	// Description: An optional description of this resource. Provide this
	// property when you create the resource.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// Interfaces: Router interfaces. Each interface requires either one
	// linked resource, (for example, linkedVpnTunnel), or IP address and IP
	// address range (for example, ipRange), or both.
	// +optional
	Interfaces []*RouterInterface `json:"interfaces,omitempty"`

	// Nats: A list of NAT services created in this router.
	Nats []*RouterNat `json:"nats,omitempty"`

	// Network: URI of the network to which this router belongs.
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

	// Region: URL of the region where the Subnetwork resides. This field
	// can be set only at resource creation time.
	// +immutable
	Region string `json:"region"`
}

type RouterBgp struct {
	// AdvertiseMode: User-specified flag to indicate which mode to use for
	// advertisement. The options are DEFAULT or CUSTOM.
	//
	// Possible values:
	//   "CUSTOM"
	//   "DEFAULT"
	// +optional
	AdvertiseMode *string `json:"advertiseMode,omitempty"`

	// AdvertisedGroups: User-specified list of prefix groups to advertise
	// in custom mode. This field can only be populated if advertise_mode is
	// CUSTOM and is advertised to all peers of the router. These groups
	// will be advertised in addition to any specified prefixes. Leave this
	// field blank to advertise no custom groups.
	//
	// Possible values:
	//   "ALL_SUBNETS"
	// +optional
	AdvertisedGroups []string `json:"advertisedGroups,omitempty"`

	// AdvertisedIPRanges: User-specified list of individual IP ranges to
	// advertise in custom mode. This field can only be populated if
	// advertise_mode is CUSTOM and is advertised to all peers of the
	// router. These IP ranges will be advertised in addition to any
	// specified groups. Leave this field blank to advertise no custom IP
	// ranges.
	// +optional
	AdvertisedIPRanges []*RouterAdvertisedIPRange `json:"advertisedIpRanges,omitempty"`

	// Asn: Local BGP Autonomous System Number (ASN). Must be an RFC6996
	// private ASN, either 16-bit or 32-bit. The value will be fixed for
	// this router resource. All VPN tunnels that link to this router will
	// have the same local ASN.
	// +optional
	Asn *int64 `json:"asn,omitempty"`
}

// RouterAdvertisedIPRange is a description-tagged IP ranges for the router
// to advertise.
type RouterAdvertisedIPRange struct {
	// Description: User-specified description for the IP range.
	// +optional
	Description *string `json:"description,omitempty"`

	// Range: The IP range to advertise. The value must be a CIDR-formatted
	// string.
	Range string `json:"range"`
}

// RouterBgpPeer information that must be configured into the routing
// stack to establish BGP peering. This information must specify the
// peer ASN and either the interface name, IP address, or peer IP
// address. Please refer to RFC4273.
type RouterBgpPeer struct {
	// AdvertiseMode: User-specified flag to indicate which mode to use for
	// advertisement.
	//
	// Possible values:
	//   "CUSTOM"
	//   "DEFAULT"
	// +optional
	AdvertiseMode *string `json:"advertiseMode,omitempty"`

	// AdvertisedGroups: User-specified list of prefix groups to advertise
	// in custom mode, which can take one of the following options:
	// - ALL_SUBNETS: Advertises all available subnets, including peer VPC
	// subnets.
	// - ALL_VPC_SUBNETS: Advertises the router's own VPC subnets.
	// - ALL_PEER_VPC_SUBNETS: Advertises peer subnets of the router's VPC
	// network. Note that this field can only be populated if advertise_mode
	// is CUSTOM and overrides the list defined for the router (in the "bgp"
	// message). These groups are advertised in addition to any specified
	// prefixes. Leave this field blank to advertise no custom groups.
	//
	// Possible values:
	//   "ALL_SUBNETS"
	// +optional
	AdvertisedGroups []string `json:"advertisedGroups,omitempty"`

	// AdvertisedIPRanges: User-specified list of individual IP ranges to
	// advertise in custom mode. This field can only be populated if
	// advertise_mode is CUSTOM and overrides the list defined for the
	// router (in the "bgp" message). These IP ranges are advertised in
	// addition to any specified groups. Leave this field blank to advertise
	// no custom IP ranges.
	// +optional
	AdvertisedIPRanges []*RouterAdvertisedIPRange `json:"advertisedIpRanges,omitempty"`

	// AdvertisedRoutePriority: The priority of routes advertised to this
	// BGP peer. Where there is more than one matching route of maximum
	// length, the routes with the lowest priority value win.
	// +optional
	AdvertisedRoutePriority *int64 `json:"advertisedRoutePriority,omitempty"`

	// InterfaceName: Name of the interface the BGP peer is associated with.
	// +optional
	InterfaceName *string `json:"interfaceName,omitempty"`

	// IPAddress: IP address of the interface inside Google Cloud Platform.
	// Only IPv4 is supported.
	// +optional
	IPAddress *string `json:"ipAddress,omitempty"`

	// Name: Name of this BGP peer. The name must be 1-63 characters long,
	// and comply with RFC1035. Specifically, the name must be 1-63
	// characters long and match the regular expression
	// `[a-z]([-a-z0-9]*[a-z0-9])?` which means the first character must be
	// a lowercase letter, and all following characters must be a dash,
	// lowercase letter, or digit, except the last character, which cannot
	// be a dash.
	Name string `json:"name"`

	// PeerAsn: Peer BGP Autonomous System Number (ASN). Each BGP interface
	// may use a different value.
	PeerAsn int64 `json:"peerAsn"`

	// PeerIPAddress: IP address of the BGP interface outside Google Cloud
	// Platform. Only IPv4 is supported.
	// +optional
	PeerIPAddress *string `json:"peerIpAddress,omitempty"`
}

// RouterBgpPeerObservation about BGP peering.
type RouterBgpPeerObservation struct {
	// ManagementType: [Output Only] The resource that configures and
	// manages this BGP peer.
	// - MANAGED_BY_USER is the default value and can be managed by you or
	// other users
	// - MANAGED_BY_ATTACHMENT is a BGP peer that is configured and managed
	// by Cloud Interconnect, specifically by an InterconnectAttachment of
	// type PARTNER. Google automatically creates, updates, and deletes this
	// type of BGP peer when the PARTNER InterconnectAttachment is created,
	// updated, or deleted.
	//
	// Possible values:
	//   "MANAGED_BY_ATTACHMENT"
	//   "MANAGED_BY_USER"
	ManagementType string `json:"managementType,omitempty"`
}

type RouterInterface struct {
	// IPRange: IP address and range of the interface. The IP range must be
	// in the RFC3927 link-local IP address space. The value must be a
	// CIDR-formatted string, for example: 169.254.0.1/30. NOTE: Do not
	// truncate the address as it represents the IP address of the
	// interface.
	// +optional
	IPRange *string `json:"ipRange,omitempty"`

	// LinkedInterconnectAttachment: URI of the linked Interconnect
	// attachment. It must be in the same region as the router. Each
	// interface can have one linked resource, which can be either be a VPN
	// tunnel or an Interconnect attachment.
	// +optional
	LinkedInterconnectAttachment *string `json:"linkedInterconnectAttachment,omitempty"`

	// LinkedVpnTunnel: URI of the linked VPN tunnel, which must be in the
	// same region as the router. Each interface can have one linked
	// resource, which can be either a VPN tunnel or an Interconnect
	// attachment.
	// +optional
	LinkedVpnTunnel *string `json:"linkedVpnTunnel,omitempty"`

	// Name: Name of this interface entry. The name must be 1-63 characters
	// long, and comply with RFC1035. Specifically, the name must be 1-63
	// characters long and match the regular expression
	// `[a-z]([-a-z0-9]*[a-z0-9])?` which means the first character must be
	// a lowercase letter, and all following characters must be a dash,
	// lowercase letter, or digit, except the last character, which cannot
	// be a dash.
	Name string `json:"name"`
}

type RouterInterfaceObservation struct {
	// ManagementType: [Output Only] The resource that configures and
	// manages this interface.
	// - MANAGED_BY_USER is the default value and can be managed directly by
	// users.
	// - MANAGED_BY_ATTACHMENT is an interface that is configured and
	// managed by Cloud Interconnect, specifically, by an
	// InterconnectAttachment of type PARTNER. Google automatically creates,
	// updates, and deletes this type of interface when the PARTNER
	// InterconnectAttachment is created, updated, or deleted.
	//
	// Possible values:
	//   "MANAGED_BY_ATTACHMENT"
	//   "MANAGED_BY_USER"
	ManagementType string `json:"managementType,omitempty"`
}

// RouterNat represents a Nat resource. It enables the VMs within the
// specified subnetworks to access Internet without external IP
// addresses. It specifies a list of subnetworks (and the ranges within)
// that want to use NAT. Customers can also provide the external IPs
// that would be used for NAT. GCP would auto-allocate ephemeral IPs if
// no external IPs are provided.
type RouterNat struct {
	// DrainNatIps: A list of URLs of the IP resources to be drained. These
	// IPs must be valid static external IPs that have been assigned to the
	// NAT. These IPs should be used for updating/patching a NAT only.
	// +optional
	DrainNatIps []string `json:"drainNatIps,omitempty"`

	// IcmpIdleTimeoutSec: Timeout (in seconds) for ICMP connections.
	// Defaults to 30s if not set.
	// +optional
	IcmpIdleTimeoutSec *int64 `json:"icmpIdleTimeoutSec,omitempty"`

	// LogConfig: Configure logging on this NAT.
	// +optional
	LogConfig *RouterNatLogConfig `json:"logConfig,omitempty"`

	// MinPortsPerVM: Minimum number of ports allocated to a VM from this
	// NAT config. If not set, a default number of ports is allocated to a
	// VM. This is rounded up to the nearest power of 2. For example, if the
	// value of this field is 50, at least 64 ports are allocated to a VM.
	// +optional
	MinPortsPerVM *int64 `json:"minPortsPerVm,omitempty"`

	// Name: Unique name of this Nat service. The name must be 1-63
	// characters long and comply with RFC1035.
	Name string `json:"name,omitempty"`

	// NatIPAllocateOption: Specify the NatIpAllocateOption, which can take
	// one of the following values:
	// - MANUAL_ONLY: Uses only Nat IP addresses provided by customers. When
	// there are not enough specified Nat IPs, the Nat service fails for new
	// VMs.
	// - AUTO_ONLY: Nat IPs are allocated by Google Cloud Platform;
	// customers can't specify any Nat IPs. When choosing AUTO_ONLY, then
	// nat_ip should be empty.
	//
	// Possible values:
	//   "AUTO_ONLY"
	//   "MANUAL_ONLY"
	// +optional
	NatIPAllocateOption *string `json:"natIpAllocateOption,omitempty"`

	// NatIPs: A list of URLs of the IP resources used for this Nat service.
	// These IP addresses must be valid static external IP addresses
	// assigned to the project.
	// +optional
	NatIPs []string `json:"natIps,omitempty"`

	// SourceSubnetworkIPRangesToNat: Specify the Nat option, which can take
	// one of the following values:
	// - ALL_SUBNETWORKS_ALL_IP_RANGES: All of the IP ranges in every
	// Subnetwork are allowed to Nat.
	// - ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES: All of the primary IP ranges
	// in every Subnetwork are allowed to Nat.
	// - LIST_OF_SUBNETWORKS: A list of Subnetworks are allowed to Nat
	// (specified in the field subnetwork below) The default is
	// SUBNETWORK_IP_RANGE_TO_NAT_OPTION_UNSPECIFIED. Note that if this
	// field contains ALL_SUBNETWORKS_ALL_IP_RANGES or
	// ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES, then there should not be any
	// other Router.Nat section in any Router for this network in this
	// region.
	//
	// Possible values:
	//   "ALL_SUBNETWORKS_ALL_IP_RANGES"
	//   "ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES"
	//   "LIST_OF_SUBNETWORKS"
	// +optional
	SourceSubnetworkIPRangesToNat *string `json:"sourceSubnetworkIpRangesToNat,omitempty"`

	// Subnetworks: A list of Subnetwork resources whose traffic should be
	// translated by NAT Gateway. It is used only when LIST_OF_SUBNETWORKS
	// is selected for the SubnetworkIpRangeToNatOption above.
	// +optional
	Subnetworks []*RouterNatSubnetworkToNat `json:"subnetworks,omitempty"`

	// TCPEstablishedIdleTimeoutSec: Timeout (in seconds) for TCP
	// established connections. Defaults to 1200s if not set.
	// +optional
	TCPEstablishedIdleTimeoutSec *int64 `json:"tcpEstablishedIdleTimeoutSec,omitempty"`

	// TCPTransitoryIdleTimeoutSec: Timeout (in seconds) for TCP transitory
	// connections. Defaults to 30s if not set.
	// +optional
	TCPTransitoryIdleTimeoutSec *int64 `json:"tcpTransitoryIdleTimeoutSec,omitempty"`

	// UDPIdleTimeoutSec: Timeout (in seconds) for UDP connections. Defaults
	// to 30s if not set.
	// +optional
	UDPIdleTimeoutSec *int64 `json:"udpIdleTimeoutSec,omitempty"`
}

// RouterNatLogConfig: Configuration of logging on a NAT.
type RouterNatLogConfig struct {
	// Enable: Indicates whether or not to export logs. This is false by
	// default.
	Enable bool `json:"enable"`

	// Filter: Specify the desired filtering of logs on this NAT. If
	// unspecified, logs are exported for all connections handled by this
	// NAT. This option can take one of the following values:
	// - ERRORS_ONLY: Export logs only for connection failures.
	// - TRANSLATIONS_ONLY: Export logs only for successful connections.
	// - ALL: Export logs for all connections, successful and unsuccessful.
	//
	// Possible values:
	//   "ALL"
	//   "ERRORS_ONLY"
	//   "TRANSLATIONS_ONLY"
	Filter *string `json:"filter,omitempty"`
}

// RouterNatSubnetworkToNat defines the IP ranges that want to use NAT
// for a subnetwork.
type RouterNatSubnetworkToNat struct {
	// Name: URL for the subnetwork resource that will use NAT.
	// +optional
	// +immutable
	Name *string `json:"name,omitempty"`

	// NameRef references a Subnetwork and retrieves its URI
	// +optional
	// +immutable
	NameRef *xpv1.Reference `json:"nameRef,omitempty"`

	// NameSelector selects a reference to a Subnetwork
	// +optional
	// +immutable
	NameSelector *xpv1.Selector `json:"nameSelector,omitempty"`

	// SecondaryIpRangeNames: A list of the secondary ranges of the
	// Subnetwork that are allowed to use NAT. This can be populated only if
	// "LIST_OF_SECONDARY_IP_RANGES" is one of the values in
	// source_ip_ranges_to_nat.
	// +optional
	SecondaryIpRangeNames []string `json:"secondaryIpRangeNames,omitempty"`

	// SourceIpRangesToNat: Specify the options for NAT ranges in the
	// Subnetwork. All options of a single value are valid except
	// NAT_IP_RANGE_OPTION_UNSPECIFIED. The only valid option with multiple
	// values is: ["PRIMARY_IP_RANGE", "LIST_OF_SECONDARY_IP_RANGES"]
	// Default: [ALL_IP_RANGES]
	//
	// Possible values:
	//   "ALL_IP_RANGES"
	//   "LIST_OF_SECONDARY_IP_RANGES"
	//   "PRIMARY_IP_RANGE"
	// +optional
	SourceIpRangesToNat []string `json:"sourceIpRangesToNat,omitempty"`
}

// A RouterObservation represents the observed state of a Google Compute Engine
// VPC Router.
type RouterObservation struct {
	// CreationTimestamp: [Output Only] Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// ID: [Output Only] The unique identifier for the resource. This
	// identifier is defined by the server.
	ID int64 `json:"id,omitempty,string"`

	// Kind: [Output Only] Type of resource. Always compute#router for
	// routers.
	Kind string `json:"kind,omitempty"`

	// Region: [Output Only] URI of the region where the router resides. You
	// must specify this field as part of the HTTP request URL. It is not
	// settable as a field in the request body.
	Region string `json:"region,omitempty"`

	RouterBgpPeerObservation `json:"routingBgpPeerObservation,omitempty"`

	RouterInterfaceObservation `json:"routerInterfaceObservation,omitempty"`

	// SelfLink: [Output Only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`
}

// A RouterSpec defines the desired state of a Router.
type RouterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RouterParameters `json:"forProvider"`
}

// A RouterStatus represents the observed state of a Router.
type RouterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RouterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Router is a managed resource that represents a Google Compute Engine VPC
// Router.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterSpec   `json:"spec"`
	Status RouterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouterList contains a list of Router.
type RouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Router `json:"items"`
}
