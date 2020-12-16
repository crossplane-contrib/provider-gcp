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

package v1alpha1

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// KeyRingRRN extracts the partially qualified URL of a Network.
func KeyRingRRN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*KeyRing)
		if !ok {
			return ""
		}
		return n.Status.AtProvider.Name
	}
}

// ResolveReferences of this Subnetwork
func (in *CryptoKey) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, in)

	// Resolve spec.forProvider.network
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(in.Spec.ForProvider.KeyRing),
		Reference:    in.Spec.ForProvider.KeyRingRef,
		Selector:     in.Spec.ForProvider.KeyRingSelector,
		To:           reference.To{Managed: &KeyRing{}, List: &KeyRingList{}},
		Extract:      KeyRingRRN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.network")
	}

	in.Spec.ForProvider.KeyRing = reference.ToPtrValue(rsp.ResolvedValue)
	in.Spec.ForProvider.KeyRingRef = rsp.ResolvedReference

	return nil
}
