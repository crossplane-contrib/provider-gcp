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

package network

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	compute "google.golang.org/api/compute/v1"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

// GenerateNetwork takes a *NetworkParameters and returns *compute.Network.
// It assigns only the fields that are writable, i.e. not labelled as [Output Only]
// in Google's reference.
func GenerateNetwork(in v1beta1.NetworkParameters, name string) *compute.Network {
	n := &compute.Network{
		Name:        name,
		Description: gcp.StringValue(in.Description),
	}
	if in.AutoCreateSubnetworks != nil {
		n.AutoCreateSubnetworks = *in.AutoCreateSubnetworks
		if !n.AutoCreateSubnetworks {
			n.ForceSendFields = []string{"AutoCreateSubnetworks"}
		}
	}
	if in.RoutingConfig != nil {
		n.RoutingConfig = &compute.NetworkRoutingConfig{
			RoutingMode: in.RoutingConfig.RoutingMode,
		}
	}
	return n
}

// GenerateNetworkObservation takes a compute.Network and returns *NetworkObservation.
func GenerateNetworkObservation(in compute.Network) v1beta1.NetworkObservation {
	gn := v1beta1.NetworkObservation{
		CreationTimestamp: in.CreationTimestamp,
		GatewayIPv4:       in.GatewayIPv4,
		ID:                in.Id,
		SelfLink:          in.SelfLink,
		Subnetworks:       in.Subnetworks,
	}
	for _, p := range in.Peerings {
		gp := &v1beta1.NetworkPeering{
			Name:                 p.Name,
			Network:              p.Network,
			State:                p.State,
			AutoCreateRoutes:     p.AutoCreateRoutes,
			ExchangeSubnetRoutes: p.ExchangeSubnetRoutes,
			StateDetails:         p.StateDetails,
		}
		gn.Peerings = append(gn.Peerings, gp)
	}
	return gn
}

// LateInitializeSpec fills unassigned fields with the values in compute.Network object.
func LateInitializeSpec(spec *v1beta1.NetworkParameters, in compute.Network) {
	spec.AutoCreateSubnetworks = gcp.LateInitializeBool(spec.AutoCreateSubnetworks, in.AutoCreateSubnetworks)
	if in.RoutingConfig != nil && spec.RoutingConfig == nil {
		spec.RoutingConfig = &v1beta1.NetworkRoutingConfig{
			RoutingMode: in.RoutingConfig.RoutingMode,
		}
	}

	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1beta1.NetworkParameters, current compute.Network) bool {
	currentParams := &v1beta1.NetworkParameters{}
	LateInitializeSpec(currentParams, current)
	return cmp.Equal(in, currentParams, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{}))
}
