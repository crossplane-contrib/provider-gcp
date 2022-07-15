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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"

	"github.com/crossplane-contrib/provider-gcp/apis/compute/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// GenerateRouter takes a *RouterParameters and returns *compute.Router.
// It assigns only the fields that are writable, i.e. not labelled as [Output Only]
// in Google's reference.
func GenerateRouter(name string, in v1alpha1.RouterParameters, router *compute.Router) { // nolint:gocyclo
	router.Name = name
	router.Description = gcp.StringValue(in.Description)
	router.Network = gcp.StringValue(in.Network)
	router.EncryptedInterconnectRouter = gcp.BoolValue(in.EncryptedInterconnectRouter)

	if in.Bgp != nil {
		router.Bgp = &compute.RouterBgp{}
		router.Bgp.AdvertiseMode = gcp.StringValue(in.Bgp.AdvertiseMode)
		router.Bgp.AdvertisedGroups = in.Bgp.AdvertisedGroups
		router.Bgp.Asn = gcp.Int64Value(in.Bgp.Asn)
		if in.Bgp.AdvertisedIpRanges != nil {
			router.Bgp.AdvertisedIpRanges = make([]*compute.RouterAdvertisedIpRange, len(in.Bgp.AdvertisedIpRanges))
			for idx, ipRange := range in.Bgp.AdvertisedIpRanges {
				router.Bgp.AdvertisedIpRanges[idx] = &compute.RouterAdvertisedIpRange{
					Description: gcp.StringValue(ipRange.Description),
					Range:       ipRange.Range,
				}
			}
		}
	}

	if in.BgpPeers != nil {
		router.BgpPeers = make([]*compute.RouterBgpPeer, len(in.BgpPeers))
		for idx, peer := range in.BgpPeers {
			router.BgpPeers[idx] = &compute.RouterBgpPeer{
				AdvertiseMode:           gcp.StringValue(peer.AdvertiseMode),
				AdvertisedGroups:        peer.AdvertisedGroups,
				AdvertisedRoutePriority: gcp.Int64Value(peer.AdvertisedRoutePriority),
				InterfaceName:           gcp.StringValue(peer.InterfaceName),
				IpAddress:               gcp.StringValue(peer.IpAddress),
				Name:                    peer.Name,
				PeerAsn:                 peer.PeerAsn,
				PeerIpAddress:           gcp.StringValue(peer.PeerIpAddress),
			}
			if peer.AdvertisedIpRanges != nil {
				router.BgpPeers[idx].AdvertisedIpRanges = make([]*compute.RouterAdvertisedIpRange, len(in.BgpPeers[idx].AdvertisedIpRanges))
				for ipIdx, ipRange := range in.BgpPeers[idx].AdvertisedIpRanges {
					router.BgpPeers[idx].AdvertisedIpRanges[ipIdx] = &compute.RouterAdvertisedIpRange{
						Description: gcp.StringValue(ipRange.Description),
						Range:       ipRange.Range,
					}
				}
			}

		}
	}

	if in.Interfaces != nil {
		router.Interfaces = make([]*compute.RouterInterface, len(in.Interfaces))
		for idx, routerInterface := range in.Interfaces {
			router.Interfaces[idx] = &compute.RouterInterface{
				IpRange:                      gcp.StringValue(routerInterface.IpRange),
				LinkedInterconnectAttachment: gcp.StringValue(routerInterface.LinkedInterconnectAttachment),
				LinkedVpnTunnel:              gcp.StringValue(routerInterface.LinkedVpnTunnel),
				Name:                         routerInterface.Name,
			}
		}
	}

	if in.Nats != nil {
		router.Nats = make([]*compute.RouterNat, len(in.Nats))
		for idx, nat := range in.Nats {
			router.Nats[idx] = &compute.RouterNat{
				DrainNatIps:                      nat.DrainNatIps,
				EnableEndpointIndependentMapping: gcp.BoolValue(nat.EnableEndpointIndependentMapping),
				IcmpIdleTimeoutSec:               gcp.Int64Value(nat.IcmpIdleTimeoutSec),
				MinPortsPerVm:                    gcp.Int64Value(nat.MinPortsPerVm),
				Name:                             gcp.StringValue(nat.Name),
				NatIpAllocateOption:              nat.NatIpAllocateOption,
				NatIps:                           nat.NatIps,
				SourceSubnetworkIpRangesToNat:    nat.SourceSubnetworkIpRangesToNat,
				TcpEstablishedIdleTimeoutSec:     gcp.Int64Value(nat.TcpEstablishedIdleTimeoutSec),
				TcpTransitoryIdleTimeoutSec:      gcp.Int64Value(nat.TcpTransitoryIdleTimeoutSec),
				UdpIdleTimeoutSec:                gcp.Int64Value(nat.UdpIdleTimeoutSec),
			}
			if nat.Subnetworks != nil {
				router.Nats[idx].Subnetworks = make([]*compute.RouterNatSubnetworkToNat, len(nat.Subnetworks))
				for subnetIdx, subnet := range nat.Subnetworks {
					router.Nats[idx].Subnetworks[subnetIdx] = &compute.RouterNatSubnetworkToNat{
						Name:                  gcp.StringValue(subnet.Name),
						SecondaryIpRangeNames: subnet.SecondaryIpRangeNames,
						SourceIpRangesToNat:   subnet.SourceIpRangesToNat,
					}
				}
			}

			if nat.LogConfig != nil {
				router.Nats[idx].LogConfig = &compute.RouterNatLogConfig{
					Enable: gcp.BoolValue(nat.LogConfig.Enable),
					Filter: gcp.StringValue(nat.LogConfig.Filter),
				}
			}
		}
	}
}

// GenerateRouterObservation takes a compute.Router and returns *RouterObservation.
func GenerateRouterObservation(in compute.Router) v1alpha1.RouterObservation {
	rt := v1alpha1.RouterObservation{
		CreationTimestamp: in.CreationTimestamp,
		ID:                in.Id,
		SelfLink:          in.SelfLink,
	}
	return rt
}

// LateInitializeSpec fills unassigned fields with the values in compute.Router object.
func LateInitializeSpec(spec *v1alpha1.RouterParameters, in compute.Router) { // nolint:gocyclo
	spec.Network = gcp.LateInitializeString(spec.Network, in.Network)
	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)
	spec.EncryptedInterconnectRouter = gcp.LateInitializeBool(spec.EncryptedInterconnectRouter, in.EncryptedInterconnectRouter)

	if in.Bgp != nil {
		spec.Bgp = &v1alpha1.RouterBgp{}
		spec.Bgp.AdvertiseMode = gcp.LateInitializeString(spec.Bgp.AdvertiseMode, in.Bgp.AdvertiseMode)
		spec.Bgp.AdvertisedGroups = gcp.LateInitializeStringSlice(spec.Bgp.AdvertisedGroups, in.Bgp.AdvertisedGroups)
		spec.Bgp.Asn = gcp.LateInitializeInt64(spec.Bgp.Asn, in.Bgp.Asn)
		if len(in.Bgp.AdvertisedIpRanges) != 0 && len(spec.Bgp.AdvertisedIpRanges) == 0 {
			spec.Bgp.AdvertisedIpRanges = make([]*v1alpha1.RouterAdvertisedIpRange, len(in.Bgp.AdvertisedIpRanges))
			for idx, ipRange := range in.Bgp.AdvertisedIpRanges {
				spec.Bgp.AdvertisedIpRanges[idx] = &v1alpha1.RouterAdvertisedIpRange{
					Description: &ipRange.Description,
					Range:       ipRange.Range,
				}
			}
		}
	}

	if len(in.BgpPeers) != 0 && len(spec.BgpPeers) == 0 {
		spec.BgpPeers = make([]*v1alpha1.RouterBgpPeer, len(in.BgpPeers))
		for idx, peer := range in.BgpPeers {
			spec.BgpPeers[idx] = &v1alpha1.RouterBgpPeer{
				AdvertiseMode:           &peer.AdvertiseMode,
				AdvertisedGroups:        peer.AdvertisedGroups,
				AdvertisedRoutePriority: &peer.AdvertisedRoutePriority,
				InterfaceName:           &peer.InterfaceName,
				IpAddress:               &peer.IpAddress,
				Name:                    peer.Name,
				PeerAsn:                 peer.PeerAsn,
				PeerIpAddress:           &peer.PeerIpAddress,
			}
			if len(peer.AdvertisedIpRanges) != 0 {
				spec.BgpPeers[idx].AdvertisedIpRanges = make([]*v1alpha1.RouterAdvertisedIpRange, len(peer.AdvertisedIpRanges))
				for ipIdx, ipRange := range peer.AdvertisedIpRanges {
					spec.BgpPeers[idx].AdvertisedIpRanges[ipIdx] = &v1alpha1.RouterAdvertisedIpRange{
						Description: &ipRange.Description,
						Range:       ipRange.Range,
					}
				}
			}
		}
	}

	if len(in.Nats) != 0 && len(spec.Nats) == 0 {
		spec.Nats = make([]*v1alpha1.RouterNat, len(in.Nats))
		for idx, nat := range in.Nats {
			spec.Nats[idx] = &v1alpha1.RouterNat{
				DrainNatIps:                      nat.DrainNatIps,
				EnableEndpointIndependentMapping: &nat.EnableEndpointIndependentMapping,
				IcmpIdleTimeoutSec:               &nat.IcmpIdleTimeoutSec,
				MinPortsPerVm:                    &nat.MinPortsPerVm,
				Name:                             &nat.Name,
				NatIpAllocateOption:              nat.NatIpAllocateOption,
				NatIps:                           nat.NatIps,
				SourceSubnetworkIpRangesToNat:    nat.SourceSubnetworkIpRangesToNat,
				TcpEstablishedIdleTimeoutSec:     &nat.TcpEstablishedIdleTimeoutSec,
				TcpTransitoryIdleTimeoutSec:      &nat.TcpTransitoryIdleTimeoutSec,
				UdpIdleTimeoutSec:                &nat.UdpIdleTimeoutSec,
			}
			if nat.LogConfig != nil {
				spec.Nats[idx].LogConfig = &v1alpha1.RouterNatLogConfig{
					Enable: &nat.LogConfig.Enable,
					Filter: &nat.LogConfig.Filter,
				}
			}

			if nat.Subnetworks != nil {
				spec.Nats[idx].Subnetworks = make([]*v1alpha1.RouterNatSubnetworkToNat, len(nat.Subnetworks))
				for subnetIdx, subnet := range nat.Subnetworks {
					spec.Nats[idx].Subnetworks[subnetIdx] = &v1alpha1.RouterNatSubnetworkToNat{
						Name:                  &subnet.Name,
						SecondaryIpRangeNames: subnet.SecondaryIpRangeNames,
						SourceIpRangesToNat:   subnet.SourceIpRangesToNat,
					}
				}
			}
		}
	}

	if len(in.Interfaces) != 0 && len(spec.Interfaces) == 0 {
		spec.Interfaces = make([]*v1alpha1.RouterInterface, len(in.Interfaces))
		for idx, routerInterface := range in.Interfaces {
			spec.Interfaces[idx] = &v1alpha1.RouterInterface{
				IpRange:                      &routerInterface.IpRange,
				LinkedInterconnectAttachment: &routerInterface.LinkedInterconnectAttachment,
				LinkedVpnTunnel:              &routerInterface.LinkedVpnTunnel,
				Name:                         routerInterface.Name,
			}
		}
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, in *v1alpha1.RouterParameters, observed *compute.Router) (upTodate bool, err error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*compute.Router)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateRouter(name, *in, desired)
	return cmp.Equal(desired, observed, cmpopts.EquateEmpty(), gcp.EquateComputeURLs(), cmpopts.IgnoreFields(compute.Router{}, "ForceSendFields")), nil
}
