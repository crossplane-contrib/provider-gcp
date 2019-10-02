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
	googlecompute "google.golang.org/api/compute/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1alpha2"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

// TEMPORARY. This should go to crossplane core repo.
const externalResourceNameAnnotationKey = "crossplane.io/external-name"

// ValidateName checks whether given string complies with name format
// that is required by GCP Network API, which is RFC1035.
func ValidateName(name string) bool {
	return len(validation.IsDNS1035Label(name)) == 0
}

// GenerateName generates a network resource name that complies with
// GCP Network API.
func GenerateName(meta v1.ObjectMeta) string {
	return "network-" + string(meta.UID)
}

// GenerateNetwork takes a *NetworkParameters and returns *googlecompute.Network.
// It assigns only the fields that are writable, i.e. not labelled as [Output Only]
// in Google's reference.
func GenerateNetwork(in v1alpha2.Network) *googlecompute.Network {
	n := &googlecompute.Network{}
	n.IPv4Range = gcp.StringValue(in.Spec.ForProvider.IPv4Range)
	if in.Spec.ForProvider.AutoCreateSubnetworks != nil {
		n.AutoCreateSubnetworks = *in.Spec.ForProvider.AutoCreateSubnetworks
		if !n.AutoCreateSubnetworks {
			n.ForceSendFields = []string{"AutoCreateSubnetworks"}
		}
	}
	n.Description = gcp.StringValue(in.Spec.ForProvider.Description)
	n.Name = in.Annotations[externalResourceNameAnnotationKey]
	if in.Spec.ForProvider.RoutingConfig != nil {
		n.RoutingConfig = &googlecompute.NetworkRoutingConfig{
			RoutingMode: gcp.StringValue(in.Spec.ForProvider.RoutingConfig.RoutingMode),
		}
	}
	return n
}

// GenerateGCPNetworkStatus takes a googlecompute.Network and returns *GCPNetworkStatus
// It assings all the fields.
func GenerateGCPNetworkStatus(in googlecompute.Network) v1alpha2.GCPNetworkStatus {
	gn := v1alpha2.GCPNetworkStatus{
		IPv4Range:             &in.IPv4Range,
		AutoCreateSubnetworks: &in.AutoCreateSubnetworks,
		CreationTimestamp:     &in.CreationTimestamp,
		Description:           &in.Description,
		GatewayIPv4:           &in.GatewayIPv4,
		ID:                    &in.Id,
		SelfLink:              &in.SelfLink,
		Subnetworks:           in.Subnetworks,
	}
	if in.RoutingConfig != nil {
		gn.RoutingConfig = &v1alpha2.GCPNetworkRoutingConfig{
			RoutingMode: &in.RoutingConfig.RoutingMode,
		}
	}
	for _, p := range in.Peerings {
		gp := &v1alpha2.GCPNetworkPeering{
			Name:                 &p.Name,
			Network:              &p.Network,
			State:                &p.State,
			AutoCreateRoutes:     &p.AutoCreateRoutes,
			ExchangeSubnetRoutes: &p.ExchangeSubnetRoutes,
			StateDetails:         &p.StateDetails,
		}
		gn.Peerings = append(gn.Peerings, gp)
	}
	return gn
}

// PopulateMissingParameters takes values from GCP Network object and fills the fields of NetworkParameters which
// which do not have value assigned and reports whether this had been applied to any field.
func PopulateMissingParameters(p *v1alpha2.NetworkParameters, in googlecompute.Network) bool {
	initial := p.DeepCopy()
	if p.IPv4Range == nil {
		p.IPv4Range = &in.IPv4Range
	}
	if p.AutoCreateSubnetworks == nil {
		p.AutoCreateSubnetworks = &in.AutoCreateSubnetworks
	}
	if p.Description == nil {
		p.Description = &in.Description
	}
	if in.RoutingConfig != nil &&
		(p.RoutingConfig == nil || (p.RoutingConfig != nil && p.RoutingConfig.RoutingMode == nil)) {
		if in.RoutingConfig != nil {
			p.RoutingConfig = &v1alpha2.GCPNetworkRoutingConfig{
				RoutingMode: &in.RoutingConfig.RoutingMode,
			}
		}
	}
	return cmp.Diff(initial, p) != ""
}
