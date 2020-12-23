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
	"fmt"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"

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

// ResolveReferences of this CryptoKey
func (in *CryptoKey) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, in)

	// Resolve spec.forProvider.keyRing
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(in.Spec.ForProvider.KeyRing),
		Reference:    in.Spec.ForProvider.KeyRingRef,
		Selector:     in.Spec.ForProvider.KeyRingSelector,
		To:           reference.To{Managed: &KeyRing{}, List: &KeyRingList{}},
		Extract:      KeyRingRRN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.keyRing")
	}

	in.Spec.ForProvider.KeyRing = reference.ToPtrValue(rsp.ResolvedValue)
	in.Spec.ForProvider.KeyRingRef = rsp.ResolvedReference

	return nil
}

// CryptoKeyRRN extracts the partially qualified URL of a Network.
func CryptoKeyRRN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*CryptoKey)
		if !ok {
			return ""
		}
		return n.Status.AtProvider.Name
	}
}

// ServiceAccountMemberName returns member name for a given ServiceAccount Object.
func ServiceAccountMemberName() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*v1alpha1.ServiceAccount)
		if !ok {
			return ""
		}
		if n.Status.AtProvider.Email == "" {
			return ""
		}
		return fmt.Sprintf("serviceAccount:%s", n.Status.AtProvider.Email)
	}
}

// ResolveReferences of this CryptoKeyPolicy
func (in *CryptoKeyPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, in)

	// Resolve spec.forProvider.keyRing
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(in.Spec.ForProvider.CryptoKey),
		Reference:    in.Spec.ForProvider.CryptoKeyRef,
		Selector:     in.Spec.ForProvider.CryptoKeySelector,
		To:           reference.To{Managed: &CryptoKey{}, List: &CryptoKeyList{}},
		Extract:      CryptoKeyRRN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.cryptoKey")
	}
	in.Spec.ForProvider.CryptoKey = reference.ToPtrValue(rsp.ResolvedValue)
	in.Spec.ForProvider.CryptoKeyRef = rsp.ResolvedReference

	// Resolve spec.ForProvider.Policy.Bindings[*].Members
	for i := range in.Spec.ForProvider.Policy.Bindings {
		mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
			CurrentValues: in.Spec.ForProvider.Policy.Bindings[i].Members,
			References:    in.Spec.ForProvider.Policy.Bindings[i].ServiceAccountMemberRefs,
			Selector:      in.Spec.ForProvider.Policy.Bindings[i].ServiceAccountMemberSelector,
			To:            reference.To{Managed: &v1alpha1.ServiceAccount{}, List: &v1alpha1.ServiceAccountList{}},
			Extract:       ServiceAccountMemberName(),
		})
		if err != nil {
			return errors.Wrapf(err, "spec.forProvider.Policy.Bindings[%d].Members", i)
		}
		in.Spec.ForProvider.Policy.Bindings[i].Members = mrsp.ResolvedValues
		in.Spec.ForProvider.Policy.Bindings[i].ServiceAccountMemberRefs = mrsp.ResolvedReferences
	}

	return nil
}
