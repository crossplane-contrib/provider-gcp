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

package router

import (
	compute "google.golang.org/api/compute/v1"

	"github.com/crossplane/provider-gcp/apis/compute/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// GenerateRouter takes a *RouterParameters and returns *compute.Router.
// It assigns only the fields that are writable, i.e. not labelled as [Output Only]
// in Google's reference.
func GenerateRouter(name string, in v1alpha1.RouterParameters, router *compute.Router) {
	router.Name = name
	router.Description = gcp.StringValue(in.Description)

	if in.Bgp != nil {
		router.Bgp = &compute.RouterBgp{}
		router.Bgp.AdvertiseMode = gcp.StringValue(in.Bgp.AdvertiseMode)
		if len(in.Bgp.AdvertisedGroups) > 0 {
			router.Bgp.AdvertisedGroups = in.Bgp.AdvertisedGroups
		}
		if len(in.Bgp.AdvertisedGroups) > 0 {
			router.Bgp.AdvertisedGroups = in.Bgp.AdvertisedGroups
		}
		router.Bgp.Asn = gcp.Int64Value(in.Bgp.Asn)
	}

	if len(in.BgpPeers) > 0 {
		router.BgpPeers = make([]*compute.RouterBgpPeer, len(in.BgpPeers))
		for i, p := range in.BgpPeers {
			peer := &compute.RouterBgpPeer{
				Name:    p.Name,
				PeerAsn: p.PeerAsn,
			}
			peer.AdvertiseMode = gcp.StringValue(p.AdvertiseMode)
			if len(p.AdvertisedGroups) > 0 {
				peer.AdvertisedGroups = p.AdvertisedGroups
			}
			if len(p.AdvertisedIPRanges) > 0 {
				peer.AdvertisedIpRanges = make([]*compute.RouterAdvertisedIpRange, len(p.AdvertisedIPRanges))
				for i, r := range p.AdvertisedIPRanges {
					peer.AdvertisedIpRanges[i] = &compute.RouterAdvertisedIpRange{
						Range:       r.Range,
						Description: gcp.StringValue(r.Description),
					}
				}
			}
			peer.AdvertisedRoutePriority = gcp.Int64Value(p.AdvertisedRoutePriority)
			peer.InterfaceName = gcp.StringValue(p.InterfaceName)
			peer.IpAddress = gcp.StringValue(p.IPAddress)
			peer.PeerIpAddress = gcp.StringValue(p.PeerIPAddress)
			router.BgpPeers[i] = peer
		}
	}
	if len(in.Interfaces) > 0 {
		router.Interfaces = make([]*compute.RouterInterface, len(in.Interfaces))
		for i, f := range in.Interfaces {
			inter := &compute.RouterInterface{
				Name: f.Name,
			}
			inter.IpRange = gcp.StringValue(f.IPRange)
			inter.LinkedInterconnectAttachment = gcp.StringValue(f.LinkedInterconnectAttachment)
			inter.LinkedVpnTunnel = gcp.StringValue(f.LinkedVpnTunnel)
			router.Interfaces[i] = inter
		}
	}
	if len(in.Nats) > 0 {
		router.Nats = make([]*compute.RouterNat, len(in.Nats))
		for i, n := range in.Nats {
			nat := &compute.RouterNat{
				Name:        n.Name,
				DrainNatIps: n.DrainNatIps,
			}
			nat.IcmpIdleTimeoutSec = gcp.Int64Value(n.IcmpIdleTimeoutSec)
			if n.LogConfig != nil {
				nat.LogConfig = &compute.RouterNatLogConfig{
					Enable: n.LogConfig.Enable,
				}
				nat.LogConfig.Filter = gcp.StringValue(n.LogConfig.Filter)
			}
			nat.MinPortsPerVm = gcp.Int64Value(n.MinPortsPerVM)
			nat.NatIpAllocateOption = gcp.StringValue(n.NatIPAllocateOption)
			nat.NatIps = n.NatIPs
			nat.SourceSubnetworkIpRangesToNat = gcp.StringValue(n.SourceSubnetworkIPRangesToNat)
			nat.Subnetworks = make([]*compute.RouterNatSubnetworkToNat, len(n.Subnetworks))
			for i, sub := range n.Subnetworks {
				subnet := &compute.RouterNatSubnetworkToNat{}
				subnet.Name = gcp.StringValue(sub.Name)
				subnet.SecondaryIpRangeNames = sub.SecondaryIpRangeNames
				subnet.SourceIpRangesToNat = sub.SourceIpRangesToNat
				nat.Subnetworks[i] = subnet
			}
			nat.TcpEstablishedIdleTimeoutSec = gcp.Int64Value(n.TCPEstablishedIdleTimeoutSec)
			nat.TcpTransitoryIdleTimeoutSec = gcp.Int64Value(n.TCPTransitoryIdleTimeoutSec)
			nat.UdpIdleTimeoutSec = gcp.Int64Value(n.UDPIdleTimeoutSec)
			router.Nats[i] = nat
		}
	}
	router.Network = gcp.StringValue(in.Network)
}

// // GenerateRouterObservation takes a compute.Router and returns *RouterObservation.
// func GenerateRouterObservation(in compute.Router) v1alpha1.RouterObservation {
// 	ro := v1alpha1.RouterObservation{
// 		CreationTimestamp: in.CreationTimestamp,
// 		GatewayIPv4:       in.GatewayIPv4,
// 		ID:                in.Id,
// 		SelfLink:          in.SelfLink,
// 		SubRouters:        in.SubRouters,
// 	}
// 	return ro
// }

// // LateInitializeSpec fills unassigned fields with the values in compute.Router object.
// func LateInitializeSpec(spec *v1beta1.RouterParameters, in compute.Router) {
// 	spec.AutoCreateSubRouters = gcp.LateInitializeBool(spec.AutoCreateSubRouters, in.AutoCreateSubRouters)
// 	if in.RoutingConfig != nil && spec.RoutingConfig == nil {
// 		spec.RoutingConfig = &v1beta1.RouterRoutingConfig{
// 			RoutingMode: in.RoutingConfig.RoutingMode,
// 		}
// 	}

// 	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)
// }

// // IsUpToDate checks whether current state is up-to-date compared to the given
// // set of parameters.
// func IsUpToDate(name string, in *v1beta1.RouterParameters, observed *compute.Router) (upTodate bool, switchToCustom bool, err error) {
// 	generated, err := copystructure.Copy(observed)
// 	if err != nil {
// 		return true, false, errors.Wrap(err, errCheckUpToDate)
// 	}
// 	desired, ok := generated.(*compute.Router)
// 	if !ok {
// 		return true, false, errors.New(errCheckUpToDate)
// 	}
// 	GenerateRouter(name, *in, desired)
// 	if !desired.AutoCreateSubRouters && observed.AutoCreateSubRouters {
// 		return false, true, nil
// 	}
// 	return cmp.Equal(desired, observed, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(compute.Router{}, "ForceSendFields")), false, nil
// }
