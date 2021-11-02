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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RouterParameters defines parameters for a desired Cloud Router
// https://cloud.google.com/compute/docs/reference/rest/v1/routers
type RouterParameters struct {
	// Name of the resource. Provided by the client when the resource is created.
	// The name must be 1-63 characters long, and comply with RFC1035. Specifically,
	// the name must be 1-63 characters long and match the regular expression
	// [a-z]([-a-z0-9]*[a-z0-9])? which means the first character must be a lowercase
	// letter, and all following characters must be a dash, lowercase letter, or digit,
	// except the last character, which cannot be a dash.
	Name string `json:"name"`

	// Optional description for this resource.
	// Provided by the client when the resource is created.
	// +optional
	Description *string `json:"description,omitempty"`

	// URI of the region where the router resides.
	Region string `json:"region"`

	// URI of the network to which this router belongs.
	Network string `json:"network"`

	// Bgp information specific to this router.
	Bgp *RouteBgp `json:"bgp,omitempty"`

	// A list of NAT services created in this router.
	Nats []*RouterNat `json:"nats,omitempty"`
}

// RouterObservation is used to show the observed state of the
// Router resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type RouterObservation struct {
	// ID is the unique identifier for the router resource.
	// This identifier is defined by the server.
	ID uint64 `json:"id,omitempty"`

	// CreationTimestamp in RFC3339 text format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	Interfaces *RouterInterface `json:"interfaces,omitempty"`

	BgpPeers *RouteBgpPeer `json:"bgpPeers,omitempty"`

	// Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Indicates if a router is dedicated for use with encrypted VLAN attachments (interconnectAttachments).
	// Not currently available publicly.
	EncryptedInterconnectRouter bool `json:"encryptedInterconnectRouter,omitempty"`

	// Type of resource. Always compute#router for routers.
	Kind string `json:"kind,omitempty"`
}

type RouteBgpPeerParameters struct {
}

type RouterInterface struct {
	// Name of this interface entry. The name must be 1-63 characters long,
	// and comply with RFC1035. Specifically, the name must be 1-63 characters
	// long and match the regular expression [a-z]([-a-z0-9]*[a-z0-9])? which
	// means the first character must be a lowercase letter, and all following
	// characters must be a dash, lowercase letter, or digit, except the last
	// character, which cannot be a dash.
	Name string `json:"name,omitempty"`

	// URI of the linked VPN tunnel, which must be in the same region as the router.
	LinkedVpnTunnel string `json:"linkedVpnTunnel,omitempty"`

	// URI of the linked Interconnect attachment. It must be in the same region as the router.
	LinkedInterconnectAttachment string `json:"linkedInterconnectAttachment,omitempty"`

	// IP address and range of the interface. The IP range must be in the
	// RFC3927 link-local IP address space. The value must be a CIDR-formatted
	// string, for example: 169.254.0.1/30.
	// NOTE: Do not truncate the address as it represents the IP address of the interface.
	IpRange string `json:"ipRange,omitempty"`

	// ManagementType: The resource that configures and manages this interface.
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
	// +kubebuilder:validation:Enum=MANAGED_BY_USER;MANAGED_BY_ATTACHMENT
	ManagementType string `json:"managementType,omitempty"`

	// The regional private internal IP address that is used to establish BGP sessions
	// to a VM instance acting as a third-party Router Appliance, such as a Next Gen
	// Firewall, a Virtual Router, or an SD-WAN VM.
	PrivateIpAddress string `json:"privateIpAddress,omitempty"`

	// Name of the interface that will be redundant with the current interface you
	// are creating. The redundantInterface must belong to the same Cloud Router as
	// the interface here. To establish the BGP session to a Router Appliance VM,
	// you must create two BGP peers. The two BGP peers must be attached to two
	// separate interfaces that are redundant with each other. The redundantInterface
	// must be 1-63 characters long, and comply with RFC1035. Specifically, the
	// redundantInterface must be 1-63 characters long and match the regular expression
	// [a-z]([-a-z0-9]*[a-z0-9])? which means the first character must be a lowercase
	// letter, and all following characters must be a dash, lowercase letter, or digit,
	// except the last character, which cannot be a dash.
	RedundantInterface string `json:"redundantInterface,omitempty"`

	// The URI of the subnetwork resource that this interface belongs to, which must be
	// in the same region as the Cloud Router. When you establish a BGP session to a VM
	// instance using this interface, the VM instance must belong to the same subnetwork
	// as the subnetwork specified here.
	Subnetwork string `json:"subnetwork,omitempty"`
}

// +kubebuilder:validation:Enum=ALL_SUBNETS;ALL_VPC_SUBNETS
type RouterAdvertisedGroup string

// RouteBgpPeer information that must be configured into the routing stack to establish BGP peering.
type RouteBgpPeer struct {
	// Name of this BGP peer. The name must be 1-63 characters long, and comply with RFC1035.
	// Specifically, the name must be 1-63 characters long and match the regular expression
	// [a-z]([-a-z0-9]*[a-z0-9])? which means the first character must be a lowercase letter,
	// and all following characters must be a dash, lowercase letter, or digit, except the last
	// character, which cannot be a dash.
	Name string `json:"name,omitempty"`

	// Name of the interface the BGP peer is associated with.
	InterfaceName string `json:"interfaceName,omitempty"`

	// IP address of the interface inside Google Cloud Platform. Only IPv4 is supported.
	IpAddress string `json:"ipAddress,omitempty"`

	// IP address of the BGP interface outside Google Cloud Platform. Only IPv4 is supported.
	PeerIpAddress string `json:"peerIpAddress,omitempty"`

	// Peer BGP Autonomous System Number (ASN). Each BGP interface may use a different value.
	PeerAsn uint32 `json:"peerAsn,omitempty"`

	// The priority of routes advertised to this BGP peer. Where there is more than one matching
	// route of maximum length, the routes with the lowest priority value win.
	AdvertisedRoutePriority uint32 `json:"advertisedRoutePriority,omitempty"`

	// User-specified flag to indicate which mode to use for advertisement.
	// +kubebuilder:validation:Enum=DEFAULT;CUSTOM
	AdvertiseMode string `json:"advertiseMode,omitempty"`

	// User-specified list of prefix groups to advertise in custom mode, which can take one of the
	// following options:
	// ALL_SUBNETS: Advertises all available subnets, including peer VPC subnets.
	// ALL_VPC_SUBNETS: Advertises the router's own VPC subnets.
	// Note that this field can only be populated if advertiseMode is CUSTOM and overrides the list
	// defined for the router (in the "bgp" message). These groups are advertised in addition to any
	// specified prefixes. Leave this field blank to advertise no custom groups.
	AdvertisedGroups []*RouterAdvertisedGroup `json:"advertisedGroups,omitempty"`

	// User-specified list of individual IP ranges to advertise in custom mode. This field can only be
	// populated if advertiseMode is CUSTOM and overrides the list defined for the router
	// (in the "bgp" message). These IP ranges are advertised in addition to any specified groups. Leave
	// this field blank to advertise no custom IP ranges.
	AdvertisedIpRanges []*RouterAdvertisedIpRange `json:"advertisedIpRanges,omitempty"`

	// The resource that configures and manages this BGP peer.
	// MANAGED_BY_USER is the default value and can be managed by you or other users
	// MANAGED_BY_ATTACHMENT is a BGP peer that is configured and managed by Cloud Interconnect,
	// specifically by an InterconnectAttachment of type PARTNER. Google automatically creates,
	// updates, and deletes this type of BGP peer when the PARTNER InterconnectAttachment is created,
	// updated, or deleted.
	// +kubebuilder:validation:Enum=MANAGED_BY_USER;MANAGED_BY_ATTACHMENT
	ManagementType string `json:"managementType,omitempty"`

	// The status of the BGP peer connection.
	// If set to FALSE, any active session with the peer is terminated and all associated routing
	// information is removed. If set to TRUE, the peer connection can be established with routing
	// information. The default is TRUE.
	// +kubebuilder:validation:Enum=TRUE;FALSE
	Enable string `json:"enable,omitempty"`

	// BFD configuration for the BGP peering.
	Bfd *RouterBfd `json:"bfd,omitempty"`

	// URI of the VM instance that is used as third-party router appliances such as Next Gen Firewalls,
	// Virtual Routers, or Router Appliances. The VM instance must be located in zones contained in the
	// same region as this Cloud Router. The VM instance is the peer side of the BGP session.
	RouterApplianceInstance string `json:"routerApplianceInstance,omitempty"`
}

type RouterAdvertisedIpRange struct {
	// The IP range to advertise. The value must be a CIDR-formatted string.
	Range string `json:"range,omitempty"`

	// User-specified description for the IP range.
	Description string `json:"description,omitempty"`
}

type RouterBfd struct {
	// The BFD session initialization mode for this BGP peer.
	// If set to ACTIVE, the Cloud Router will initiate the BFD session for this BGP peer.
	// If set to PASSIVE, the Cloud Router will wait for the peer router to initiate the
	// BFD session for this BGP peer. If set to DISABLED, BFD is disabled for this BGP peer.
	// The default is PASSIVE.
	// +kubebuilder:validation:Enum=PASSIVE;ACTIVE;DISABLED
	SessionInitializationMode string `json:"sessionInitializationMode,omitempty"`

	// The minimum interval, in milliseconds, between BFD control packets transmitted to the
	// peer router. The actual value is negotiated between the two routers and is equal to
	// the greater of this value and the corresponding receive interval of the other router.
	// If set, this value must be between 1000 and 30000.
	// The default is 1000.
	MinTransmitInterval uint32 `json:"minTransmitInterval,omitempty"`

	// The minimum interval, in milliseconds, between BFD control packets received from the
	// peer router. The actual value is negotiated between the two routers and is equal to
	// the greater of this value and the transmit interval of the other router.
	// If set, this value must be between 1000 and 30000.
	// The default is 1000.
	MinReceiveInterval uint32 `json:"minReceiveInterval,omitempty"`

	// The number of consecutive BFD packets that must be missed before BFD declares that a peer is unavailable.
	// If set, the value must be a value between 5 and 16.
	// The default is 5.
	Multiplier uint32 `json:"multiplier,omitempty"`
}

type RouteBgp struct {
	// Local BGP Autonomous System Number (ASN). Must be an RFC6996 private ASN, either
	// 16-bit or 32-bit. The value will be fixed for this router resource. All VPN tunnels
	// that link to this router will have the same local ASN.
	Asn uint32 `json:"asn,omitempty"`

	// User-specified flag to indicate which mode to use for advertisement.
	// The options are DEFAULT or CUSTOM.
	// +kubebuilder:validation:Enum=DEFAULT;CUSTOM
	AdvertiseMode string `json:"advertiseMode,omitempty"`

	// User-specified list of prefix groups to advertise in custom mode. This field can only
	// be populated if advertiseMode is CUSTOM and is advertised to all peers of the router.
	// These groups will be advertised in addition to any specified prefixes. Leave this field
	// blank to advertise no custom groups.
	AdvertisedGroups []*RouterAdvertisedGroup `json:"advertisedGroups,omitempty"`

	// User-specified list of individual IP ranges to advertise in custom mode. This field can
	// only be populated if advertiseMode is CUSTOM and is advertised to all peers of the router.
	// These IP ranges will be advertised in addition to any specified groups. Leave this field
	// blank to advertise no custom IP ranges.
	AdvertisedIpRanges []*RouterAdvertisedIpRange `json:"advertisedIpRanges,omitempty"`

	// The interval in seconds between BGP keepalive messages that are sent to the peer.
	// Hold time is three times the interval at which keepalive messages are sent, and the hold
	// time is the maximum number of seconds allowed to elapse between successive keepalive
	// messages that BGP receives from a peer.
	// BGP will use the smaller of either the local hold time value or the peer's hold time value
	// as the hold time for the BGP connection between the two peers.
	// If set, this value must be between 20 and 60. The default is 20.
	KeepaliveInterval uint32 `json:"keepaliveInterval,omitempty"`
}

type RouterNat struct {
	// Unique name of this Nat service.
	// The name must be 1-63 characters long and comply with RFC1035.
	Name string `json:"name,omitempty"`

	// Specify the Nat option, which can take one of the following values:
	// ALL_SUBNETWORKS_ALL_IP_RANGES: All of the IP ranges in every Subnetwork are allowed to Nat.
	// ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES: All of the primary IP ranges in every Subnetwork
	// are allowed to Nat.
	// LIST_OF_SUBNETWORKS: A list of Subnetworks are allowed to Nat
	// (specified in the field subnetwork below)
	// The default is SUBNETWORK_IP_RANGE_TO_NAT_OPTION_UNSPECIFIED. Note that if this field contains
	// ALL_SUBNETWORKS_ALL_IP_RANGES or ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES, then there should not
	// be any other Router.Nat section in any Router for this network in this region.
	// +kubebuilder:validation:Enum=ALL_SUBNETWORKS_ALL_IP_RANGES;ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES;LIST_OF_SUBNETWORKS
	SourceSubnetworkIpRangesToNat string `json:"sourceSubnetworkIpRangesToNat,omitempty"`

	// A list of Subnetwork resources whose traffic should be translated by NAT Gateway.
	// It is used only when LIST_OF_SUBNETWORKS is selected for the SubnetworkIpRangeToNatOption above.
	Subnetworks *RouterNatSubnetworks `json:"subnetworks,omitempty"`

	// A list of URLs of the IP resources used for this Nat service. These IP addresses must be valid static
	// external IP addresses assigned to the project.
	NatIps []string `json:"natIps,omitempty"`

	// A list of URLs of the IP resources to be drained. These IPs must be valid static external IPs that
	// have been assigned to the NAT. These IPs should be used for updating/patching a NAT only.
	DrainNatIps []string `json:"drainNatIps,omitempty"`

	// Specify the NatIpAllocateOption, which can take one of the following values:
	// MANUAL_ONLY: Uses only Nat IP addresses provided by customers. When there are not enough specified
	// Nat IPs, the Nat service fails for new VMs.
	// AUTO_ONLY: Nat IPs are allocated by Google Cloud Platform; customers can't specify any Nat IPs.
	// When choosing AUTO_ONLY, then natIp should be empty.
	// +kubebuilder:validation:Enum=MANUAL_ONLY;AUTO_ONLY
	NatIpAllocateOption string `json:"natIpAllocateOption,omitempty"`

	// Minimum number of ports allocated to a VM from this NAT config. If not set, a default number of
	// ports is allocated to a VM. This is rounded up to the nearest power of 2. For example, if the
	// value of this field is 50, at least 64 ports are allocated to a VM.
	MinPortsPerVm int `json:"minPortsPerVm,omitempty"`

	// Timeout (in seconds) for UDP connections. Defaults to 30s if not set.
	UdpIdleTimeoutSec int `json:"udpIdleTimeoutSec,omitempty"`

	// Timeout (in seconds) for ICMP connections. Defaults to 30s if not set.
	IcmpIdleTimeoutSec int `json:"icmpIdleTimeoutSec,omitempty"`

	// Timeout (in seconds) for TCP established connections. Defaults to 1200s if not set.
	TcpEstablishedIdleTimeoutSec int `json:"tcpEstablishedIdleTimeoutSec,omitempty"`

	// Timeout (in seconds) for TCP transitory connections. Defaults to 30s if not set.
	TcpTransitoryIdleTimeoutSec int `json:"tcpTransitoryIdleTimeoutSec,omitempty"`

	// Timeout (in seconds) for TCP connections that are in TIME_WAIT state. Defaults to 120s if not set.
	TcpTimeWaitTimeoutSec int `json:"tcpTimeWaitTimeoutSec,omitempty"`

	// Configure logging on this NAT.
	LogConfig *RouterNatLogConfig `json:"logConfig,omitempty"`

	EnableEndpointIndependentMapping bool `json:"enableEndpointIndependentMapping,omitempty"`
}

type RouterNatSubnetworks struct {
	// URL for the subnetwork resource that will use NAT.
	Name string `json:"name,omitempty"`

	// Specify the options for NAT ranges in the Subnetwork. All options of a single value
	// are valid except NAT_IP_RANGE_OPTION_UNSPECIFIED.
	// The only valid option with multiple values is: ["PRIMARY_IP_RANGE", "LIST_OF_SECONDARY_IP_RANGES"]
	// Default: [ALL_IP_RANGES]
	// +kubebuilder:validation:Enum=ALL_IP_RANGES;PRIMARY_IP_RANGE;LIST_OF_SECONDARY_IP_RANGES
	SourceIpRangesToNat []string `json:"sourceIpRangesToNat,omitempty"`

	// A list of the secondary ranges of the Subnetwork that are allowed to use NAT. This can be
	// populated only if "LIST_OF_SECONDARY_IP_RANGES" is one of the values in sourceIpRangesToNat.
	SecondaryIpRangeNames []string `json:"secondaryIpRangeNames,omitempty"`
}

type RouterNatLogConfig struct {
	// Indicates whether or not to export logs. This is false by default.
	Enable bool `json:"enable,omitempty"`

	// Specify the desired filtering of logs on this NAT. If unspecified, logs are exported for all
	// connections handled by this NAT. This option can take one of the following values:
	// ERRORS_ONLY: instantSnapshots.export logs only for connection failures.
	// TRANSLATIONS_ONLY: instantSnapshots.export logs only for successful connections.
	// ALL: instantSnapshots.export logs for all connections, successful and unsuccessful.
	// +kubebuilder:validation:Enum=ERRORS_ONLY;TRANSLATIONS_ONLY;ALL
	Filter string `json:"filter,omitempty"`
}

// RouterSpec defines the desired state of Router
type RouterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RouterParameters `json:"forProvider"`
}

// RouterStatus defines the observed state of Router
type RouterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RouterObservation `json:"atProvider,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Router is the Schema for the routers API
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterSpec   `json:"spec,omitempty"`
	Status RouterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RouterList contains a list of Router
type RouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Router `json:"items"`
}
