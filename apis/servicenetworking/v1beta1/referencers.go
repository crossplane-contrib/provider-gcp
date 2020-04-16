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

package v1beta1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
)

// ResolveReferences of this Connection
func (mg *Connection) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.network
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Network),
		Reference:    mg.Spec.ForProvider.NetworkRef,
		Selector:     mg.Spec.ForProvider.NetworkSelector,
		To:           reference.To{Managed: &v1beta1.Network{}, List: &v1beta1.NetworkList{}},
		Extract:      v1beta1.NetworkURL(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.Network = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NetworkRef = rsp.ResolvedReference

	// Resolve spec.forProvider.reservedPeeringRanges
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.ReservedPeeringRanges,
		References:    mg.Spec.ForProvider.ReservedPeeringRangeRefs,
		Selector:      &mg.Spec.ForProvider.ReservedPeeringRangeSelector,
		To:            reference.To{Managed: &v1beta1.GlobalAddress{}, List: &v1beta1.GlobalAddressList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.ReservedPeeringRanges = mrsp.ResolvedValues
	mg.Spec.ForProvider.ReservedPeeringRangeRefs = mrsp.ResolvedReferences

	return nil
}
