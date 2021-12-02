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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RouterParameters define the desired state of a Google Compute Engine
// Router. Most fields map directly to a Router:
// https://cloud.google.com/compute/docs/reference/rest/v1/routers/
type RouterParameters struct {
	// Description: An optional description of this resource. Provide this
	// field when you create the resource.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// Region: URL of the region where the Subnetwork resides. This field
	// can be set only at resource creation time.
	// +immutable
	Region string `json:"region"`

	// Network: URI of the network to which this router belongs.
	// +immutable
	// +optional
	Network *string `json:"network,omitempty"`

	// NetworkRef references a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkRef *xpv1.Reference `json:"networkRef,omitempty"`

	// NetworkSelector selects a reference to a Network
	// +optional
	// +immutable
	NetworkSelector *xpv1.Selector `json:"networkSelector,omitempty"`

	// Bgp: BGP information specific to this router.
	// +optional
	Bgp *RouterBgp `json:"bgp,omitempty"`

	// BgpPeers: BGP information that must be configured into the routing
	// stack to establish BGP peering. This information must specify the
	// peer ASN and either the interface name, IP address, or peer IP
	// address. Please refer to RFC4273.
	// +optional
	BgpPeers []*RouterBgpPeer `json:"bgpPeers,omitempty"`

	// EncryptedInterconnectRouter: Field to indicate if a router is
	// dedicated to use with encrypted Interconnect Attachment
	// (IPsec-encrypted Cloud Interconnect feature).
	// Not currently available in all Interconnect locations.
	// +optional
	EncryptedInterconnectRouter *bool `json:"encryptedInterconnectRouter,omitempty"`

	// Interfaces: Router interfaces. Each interface requires either one
	// linked resource, (for example, linkedVpnTunnel), or IP address and IP
	// address range (for example, ipRange), or both.
	// +optional
	Interfaces []*RouterInterface `json:"interfaces,omitempty"`

	// Nats: A list of NAT services created in this router.
	// +optional
	Nats []*RouterNat `json:"nats,omitempty"`
}

// A RouterBgp represents the Bgp information for router.
type RouterBgp struct {
	// AdvertiseMode: User-specified flag to indicate which mode to use for
	// advertisement. The options are DEFAULT or CUSTOM.
	//
	// Possible values:
	//   "CUSTOM"
	//   "DEFAULT"
	// +optional
	// +kubebuilder:validation:Enum=CUSTOM;DEFAULT
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
	// +kubebuilder:validation:Enum=ALL_SUBNETS
	AdvertisedGroups []string `json:"advertisedGroups,omitempty"`

	// AdvertisedIpRanges: User-specified list of individual IP ranges to
	// advertise in custom mode. This field can only be populated if
	// advertise_mode is CUSTOM and is advertised to all peers of the
	// router. These IP ranges will be advertised in addition to any
	// specified groups. Leave this field blank to advertise no custom IP
	// ranges.
	// +optional
	AdvertisedIpRanges []*RouterAdvertisedIpRange `json:"advertisedIpRanges,omitempty"` // nolint

	// Asn: Local BGP Autonomous System Number (ASN). Must be an RFC6996
	// private ASN, either 16-bit or 32-bit. The value will be fixed for
	// this router resource. All VPN tunnels that link to this router will
	// have the same local ASN.
	// +optional
	Asn *int64 `json:"asn,omitempty"`
}

// A RouterAdvertisedIpRange represents the IP ranges advertised by router.
type RouterAdvertisedIpRange struct { // nolint
	// Description: User-specified description for the IP range.
	// +optional
	Description *string `json:"description,omitempty"`

	// Range: The IP range to advertise. The value must be a CIDR-formatted
	// string.
	Range string `json:"range"`
}

// A RouterBgpPeer represents the BgpPeer configuration for the router.
type RouterBgpPeer struct {
	// AdvertiseMode: User-specified flag to indicate which mode to use for
	// advertisement.
	//
	// Possible values:
	//   "CUSTOM"
	//   "DEFAULT"
	// +optional
	// +kubebuilder:validation:Enum=CUSTOM;DEFAULT
	AdvertiseMode *string `json:"advertiseMode,omitempty"`

	// AdvertisedGroups: User-specified list of prefix groups to advertise
	// in custom mode, which can take one of the following options:
	// - ALL_SUBNETS: Advertises all available subnets, including peer VPC
	// subnets.
	// - ALL_VPC_SUBNETS: Advertises the router's own VPC subnets. Note that
	// this field can only be populated if advertise_mode is CUSTOM and
	// overrides the list defined for the router (in the "bgp" message).
	// These groups are advertised in addition to any specified prefixes.
	// Leave this field blank to advertise no custom groups.
	//
	// Possible values:
	//   "ALL_SUBNETS"
	// +optional
	// +kubebuilder:validation:Enum=ALL_SUBNETS
	AdvertisedGroups []string `json:"advertisedGroups,omitempty"`

	// AdvertisedIpRanges: User-specified list of individual IP ranges to
	// advertise in custom mode. This field can only be populated if
	// advertise_mode is CUSTOM and overrides the list defined for the
	// router (in the "bgp" message). These IP ranges are advertised in
	// addition to any specified groups. Leave this field blank to advertise
	// no custom IP ranges.
	// +optional
	AdvertisedIpRanges []*RouterAdvertisedIpRange `json:"advertisedIpRanges,omitempty"` // nolint

	// AdvertisedRoutePriority: The priority of routes advertised to this
	// BGP peer. Where there is more than one matching route of maximum
	// length, the routes with the lowest priority value win.
	// +optional
	AdvertisedRoutePriority *int64 `json:"advertisedRoutePriority,omitempty"`

	// InterfaceName: Name of the interface the BGP peer is associated with.
	// +optional
	InterfaceName *string `json:"interfaceName,omitempty"`

	// IpAddress: IP address of the interface inside Google Cloud Platform.
	// Only IPv4 is supported.
	// +optional
	IpAddress *string `json:"ipAddress,omitempty"` // nolint

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

	// PeerIpAddress: IP address of the BGP interface outside Google Cloud
	// Platform. Only IPv4 is supported.
	// +optional
	PeerIpAddress *string `json:"peerIpAddress,omitempty"` // nolint
}

// RouterNat represents the Nat Service for the router.
type RouterNat struct {
	// DrainNatIps: A list of URLs of the IP resources to be drained. These
	// IPs must be valid static external IPs that have been assigned to the
	// NAT. These IPs should be used for updating/patching a NAT only.
	// +optional
	DrainNatIps []string `json:"drainNatIps,omitempty"`

	// +optional
	EnableEndpointIndependentMapping *bool `json:"enableEndpointIndependentMapping,omitempty"`

	// IcmpIdleTimeoutSec: Timeout (in seconds) for ICMP connections.
	// Defaults to 30s if not set.
	// +optional
	IcmpIdleTimeoutSec *int64 `json:"icmpIdleTimeoutSec,omitempty"`

	// LogConfig: Configure logging on this NAT.
	// +optional
	LogConfig *RouterNatLogConfig `json:"logConfig,omitempty"`

	// MinPortsPerVm: Minimum number of ports allocated to a VM from this
	// NAT config. If not set, a default number of ports is allocated to a
	// VM. This is rounded up to the nearest power of 2. For example, if the
	// value of this field is 50, at least 64 ports are allocated to a VM.
	// +optional
	MinPortsPerVm *int64 `json:"minPortsPerVm,omitempty"` // nolint

	// Name: Unique name of this Nat service. The name must be 1-63
	// characters long and comply with RFC1035.
	// +optional
	Name *string `json:"name,omitempty"`

	// NatIpAllocateOption: Specify the NatIpAllocateOption, which can take
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
	// +kubebuilder:validation:Enum=AUTO_ONLY;MANUAL_ONLY
	NatIpAllocateOption string `json:"natIpAllocateOption,omitempty"` // nolint

	// NatIps: A list of URLs of the IP resources used for this Nat service.
	// These IP addresses must be valid static external IP addresses
	// assigned to the project.
	// +optional
	NatIps []string `json:"natIps"`

	// SourceSubnetworkIpRangesToNat: Specify the Nat option, which can take
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
	// +kubebuilder:validation:Enum=ALL_SUBNETWORKS_ALL_IP_RANGES;ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES;LIST_OF_SUBNETWORKS
	SourceSubnetworkIpRangesToNat string `json:"sourceSubnetworkIpRangesToNat"` // nolint

	// Subnetworks: A list of Subnetwork resources whose traffic should be
	// translated by NAT Gateway. It is used only when LIST_OF_SUBNETWORKS
	// is selected for the SubnetworkIpRangeToNatOption above.
	// +optional
	Subnetworks []*RouterNatSubnetworkToNat `json:"subnetworks,omitempty"`

	// TcpEstablishedIdleTimeoutSec: Timeout (in seconds) for TCP
	// established connections. Defaults to 1200s if not set.
	// +optional
	TcpEstablishedIdleTimeoutSec *int64 `json:"tcpEstablishedIdleTimeoutSec,omitempty"` // nolint

	// TcpTransitoryIdleTimeoutSec: Timeout (in seconds) for TCP transitory
	// connections. Defaults to 30s if not set.
	// +optional
	TcpTransitoryIdleTimeoutSec *int64 `json:"tcpTransitoryIdleTimeoutSec,omitempty"` // nolint

	// UdpIdleTimeoutSec: Timeout (in seconds) for UDP connections. Defaults
	// to 30s if not set.
	// +optional
	UdpIdleTimeoutSec *int64 `json:"udpIdleTimeoutSec,omitempty"` // nolint
}

// A RouterNatSubnetworkToNat represent the Subnetwork information for Router Nat Service.
type RouterNatSubnetworkToNat struct {
	// Name: URL for the subnetwork resource that will use NAT.
	// +optional
	Name *string `json:"name,omitempty"`

	// SecondaryIpRangeNames: A list of the secondary ranges of the
	// Subnetwork that are allowed to use NAT. This can be populated only if
	// "LIST_OF_SECONDARY_IP_RANGES" is one of the values in
	// source_ip_ranges_to_nat.
	// +optional
	SecondaryIpRangeNames []string `json:"secondaryIpRangeNames,omitempty"` // nolint

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
	// +kubebuilder:validation:Enum=ALL_IP_RANGES;LIST_OF_SECONDARY_IP_RANGES;PRIMARY_IP_RANGE
	SourceIpRangesToNat []string `json:"sourceIpRangesToNat,omitempty"` // nolint
}

// A RouterNatLogConfig represent the Log config Router Nat service.
type RouterNatLogConfig struct {
	// Enable: Indicates whether or not to export logs. This is false by
	// default.
	// +optional
	Enable *bool `json:"enable,omitempty"`

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
	// +optional
	// +kubebuilder:validation:Enum=ALL;ERRORS_ONLY;TRANSLATIONS_ONLY
	Filter *string `json:"filter,omitempty"`
}

// A RouterInterface represent the Interface information for router.
type RouterInterface struct {
	// IpRange: IP address and range of the interface. The IP range must be
	// in the RFC3927 link-local IP address space. The value must be a
	// CIDR-formatted string, for example: 169.254.0.1/30. NOTE: Do not
	// truncate the address as it represents the IP address of the
	// interface.
	// +optional
	IpRange *string `json:"ipRange,omitempty"` // nolint

	// LinkedInterconnectAttachment: URI of the linked Interconnect
	// attachment. It must be in the same region as the router. Each
	// interface can have one linked resource, which can be a VPN tunnel, an
	// Interconnect attachment, or a virtual machine instance.
	// +optional
	LinkedInterconnectAttachment *string `json:"linkedInterconnectAttachment,omitempty"`

	// LinkedVpnTunnel: URI of the linked VPN tunnel, which must be in the
	// same region as the router. Each interface can have one linked
	// resource, which can be a VPN tunnel, an Interconnect attachment, or a
	// virtual machine instance.
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

// A RouterObservation represents the observed state of a Google Compute Engine
// Router.
type RouterObservation struct {
	// CreationTimestamp: Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// Id: The unique identifier for the resource. This
	// identifier is defined by the server.
	ID uint64 `json:"id,omitempty"`

	// SelfLink: Server-defined URL for the resource.
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

// A Router is a managed resource that represents a Google Compute Engine Router
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

// RouterList contains a list of Routers.
type RouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Router `json:"items"`
}
