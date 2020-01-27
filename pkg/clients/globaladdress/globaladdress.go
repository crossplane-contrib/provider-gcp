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

package globaladdress

import (
	compute "google.golang.org/api/compute/v1"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

// GenerateGlobalAddress converts the supplied GlobalAddressParameters into an
// Address suitable for use with the Google Compute API.
func GenerateGlobalAddress(p v1beta1.GlobalAddressParameters, name string) *compute.Address {
	// Kubernetes API conventions dictate that optional, unspecified fields must
	// be nil. GCP API clients omit any field set to its zero value, using
	// NullFields and ForceSendFields to handle edge cases around unsetting
	// previously set values, or forcing zero values to be set. The Address API
	// does not support updates, so we can safely convert any nil pointer to
	// string or int64 to their zero values.
	return &compute.Address{
		Address:      gcp.StringValue(p.Address),
		AddressType:  gcp.StringValue(p.AddressType),
		Description:  gcp.StringValue(p.Description),
		IpVersion:    gcp.StringValue(p.IPVersion),
		Name:         name,
		Network:      gcp.StringValue(p.Network),
		PrefixLength: gcp.Int64Value(p.PrefixLength),
		Purpose:      gcp.StringValue(p.Purpose),
		Subnetwork:   gcp.StringValue(p.Subnetwork),
	}
}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied GlobalAddressParameters that are set (i.e. non-zero) on the supplied
// GlobalAddress.
func LateInitializeSpec(p *v1beta1.GlobalAddressParameters, observed compute.Address) {
	p.Address = gcp.LateInitializeString(p.Address, observed.Address)
	p.AddressType = gcp.LateInitializeString(p.AddressType, observed.AddressType)
	p.Description = gcp.LateInitializeString(p.Description, observed.Description)
	p.IPVersion = gcp.LateInitializeString(p.IPVersion, observed.IpVersion)
	p.Network = gcp.LateInitializeString(p.Network, observed.Network)
	p.PrefixLength = gcp.LateInitializeInt64(p.PrefixLength, observed.PrefixLength)
	p.Purpose = gcp.LateInitializeString(p.Purpose, observed.Purpose)
	p.Subnetwork = gcp.LateInitializeString(p.Subnetwork, observed.Subnetwork)
}

// GenerateGlobalAddressObservation takes a compute.Address and returns
// *GlobalAddressObservation.
func GenerateGlobalAddressObservation(observed compute.Address) v1beta1.GlobalAddressObservation {
	return v1beta1.GlobalAddressObservation{
		CreationTimestamp: observed.CreationTimestamp,
		ID:                observed.Id,
		SelfLink:          observed.SelfLink,
		Status:            observed.Status,
		Users:             observed.Users,
	}
}
