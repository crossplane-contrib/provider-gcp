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

package connection

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	compute "google.golang.org/api/compute/v1"
	servicenetworking "google.golang.org/api/servicenetworking/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane/provider-gcp/apis/servicenetworking/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

// PeeringName of the peering created when a service networking connection is
// added to a VPC network.
const PeeringName = "servicenetworking-googleapis-com"

// VPC Network peering states.
const (
	PeeringStateActive   = "ACTIVE"
	PeeringStateInactive = "INACTIVE"
)

// FromParameters converts the supplied ConnectionParameters into a Connection
// suitable for use with the Google Compute API.
func FromParameters(p v1beta1.ConnectionParameters) *servicenetworking.Connection {
	// Kubernetes API conventions dictate that optional, unspecified fields must
	// be nil. GCP API clients omit any field set to its zero value, using
	// NullFields and ForceSendFields to handle edge cases around unsetting
	// previously set values, or forcing zero values to be set.
	return &servicenetworking.Connection{
		Network:               gcp.StringValue(p.Network),
		ReservedPeeringRanges: p.ReservedPeeringRanges,
		ForceSendFields:       []string{"ReservedPeeringRanges"},
	}
}

// IsUpToDate returns true if the observed Connection is up to date with the
// supplied ConnectionParameters.
func IsUpToDate(p v1beta1.ConnectionParameters, observed *servicenetworking.Connection) bool {
	return cmp.Equal(p.ReservedPeeringRanges, observed.ReservedPeeringRanges, cmpopts.SortSlices(func(i, j string) bool { return i < j }))
}

// An Observation of a service networking Connection and the Network it pertains
// to. We require both to determine the Connection's availability, because a
// Connection is a thin abstraction around a Network's VPC peerings.
type Observation struct {
	Connection *servicenetworking.Connection
	Network    *compute.Network
}

// UpdateStatus updates any fields of the supplied ConnectionStatus to
// reflect the state of the supplied Address.
func UpdateStatus(s *v1beta1.ConnectionStatus, o Observation) {
	s.AtProvider.Peering = o.Connection.Peering
	s.AtProvider.Service = o.Connection.Service

	if len(o.Network.Peerings) == 0 {
		s.SetConditions(xpv1.Unavailable())
		return
	}

	for _, p := range o.Network.Peerings {
		if p.Name == o.Connection.Peering {
			switch p.State {
			case PeeringStateActive:
				s.SetConditions(xpv1.Available())
			case PeeringStateInactive:
				s.SetConditions(xpv1.Unavailable())
			}
		}
	}
}
