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

package subnetwork

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/compute/v1"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

var equateSecondaryRange = func(i, j *v1beta1.SubnetworkSecondaryRange) bool { return i.RangeName > j.RangeName }

// GenerateSubnetwork creates a *googlecompute.Subnetwork object using SubnetworkParameters.
func GenerateSubnetwork(s v1beta1.SubnetworkParameters, name string) *compute.Subnetwork {
	sn := &compute.Subnetwork{
		Name:                  name,
		Description:           gcp.StringValue(s.Description),
		EnableFlowLogs:        gcp.BoolValue(s.EnableFlowLogs),
		IpCidrRange:           s.IPCidrRange,
		Network:               gcp.StringValue(s.Network),
		PrivateIpGoogleAccess: gcp.BoolValue(s.PrivateIPGoogleAccess),
		Region:                s.Region,
	}
	for _, val := range s.SecondaryIPRanges {
		obj := &compute.SubnetworkSecondaryRange{
			IpCidrRange: val.IPCidrRange,
			RangeName:   val.RangeName,
		}
		sn.SecondaryIpRanges = append(sn.SecondaryIpRanges, obj)
	}
	return sn
}

// GenerateSubnetworkForUpdate creates a *googlecompute.Subnetwork object using
// SubnetworkParameters with fields disallowed by the GCP API removed. If a
// field can be included in the GCP API but will result in an error if the value
// is changed, it will still be included here such that users are notified of
// invalid updates.
func GenerateSubnetworkForUpdate(s v1beta1.Subnetwork, name string) *compute.Subnetwork {
	sn := &compute.Subnetwork{
		Name:                  name,
		Description:           gcp.StringValue(s.Spec.ForProvider.Description),
		EnableFlowLogs:        gcp.BoolValue(s.Spec.ForProvider.EnableFlowLogs),
		IpCidrRange:           s.Spec.ForProvider.IPCidrRange,
		PrivateIpGoogleAccess: gcp.BoolValue(s.Spec.ForProvider.PrivateIPGoogleAccess),
		Fingerprint:           s.Status.AtProvider.Fingerprint,
	}
	for _, val := range s.Spec.ForProvider.SecondaryIPRanges {
		obj := &compute.SubnetworkSecondaryRange{
			IpCidrRange: val.IPCidrRange,
			RangeName:   val.RangeName,
		}
		sn.SecondaryIpRanges = append(sn.SecondaryIpRanges, obj)
	}
	return sn
}

// GenerateSubnetworkObservation creates a SubnetworkObservation object using *googlecompute.Subnetwork.
func GenerateSubnetworkObservation(in compute.Subnetwork) v1beta1.SubnetworkObservation {
	return v1beta1.SubnetworkObservation{
		CreationTimestamp: in.CreationTimestamp,
		Fingerprint:       in.Fingerprint,
		GatewayAddress:    in.GatewayAddress,
		ID:                in.Id,
		SelfLink:          in.SelfLink,
	}
}

// LateInitializeSpec fills unassigned fields with the values in compute.Subnetwork object.
func LateInitializeSpec(spec *v1beta1.SubnetworkParameters, in compute.Subnetwork) {
	if spec.IPCidrRange == "" {
		spec.IPCidrRange = in.IpCidrRange
	}

	if spec.Region == "" {
		spec.Region = in.Region
	}

	spec.Network = gcp.LateInitializeString(spec.Network, in.Network)
	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)
	spec.EnableFlowLogs = gcp.LateInitializeBool(spec.EnableFlowLogs, in.EnableFlowLogs)
	spec.PrivateIPGoogleAccess = gcp.LateInitializeBool(spec.PrivateIPGoogleAccess, in.PrivateIpGoogleAccess)
	if len(in.SecondaryIpRanges) != 0 && len(spec.SecondaryIPRanges) == 0 {
		spec.SecondaryIPRanges = make([]*v1beta1.SubnetworkSecondaryRange, len(in.SecondaryIpRanges))
		for i, r := range in.SecondaryIpRanges {
			spec.SecondaryIPRanges[i] = &v1beta1.SubnetworkSecondaryRange{
				IPCidrRange: r.IpCidrRange,
				RangeName:   r.RangeName,
			}
		}
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1beta1.SubnetworkParameters, current compute.Subnetwork) (upToDate bool, privateAccess bool) {
	currentParams := &v1beta1.SubnetworkParameters{}
	LateInitializeSpec(currentParams, current)
	if !cmp.Equal(in.PrivateIPGoogleAccess, currentParams.PrivateIPGoogleAccess) {
		return false, true
	}
	return cmp.Equal(in, currentParams, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{}), cmpopts.SortSlices(equateSecondaryRange)), false
}
