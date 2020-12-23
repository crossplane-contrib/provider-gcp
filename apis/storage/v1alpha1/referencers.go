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

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane/provider-gcp/apis/storage/v1alpha3"
)

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

// ResolveReferences of this BucketPolicy
func (in *BucketPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, in)

	// Resolve spec.forProvider.bucket
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(in.Spec.ForProvider.Bucket),
		Reference:    in.Spec.ForProvider.BucketRef,
		Selector:     in.Spec.ForProvider.BucketSelector,
		To:           reference.To{Managed: &v1alpha3.Bucket{}, List: &v1alpha3.BucketList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.bucket")
	}
	in.Spec.ForProvider.Bucket = reference.ToPtrValue(rsp.ResolvedValue)
	in.Spec.ForProvider.BucketRef = rsp.ResolvedReference

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

// ResolveReferences of this BucketPolicyMember
func (in *BucketPolicyMember) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, in)

	// Resolve spec.forProvider.bucket
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(in.Spec.ForProvider.Bucket),
		Reference:    in.Spec.ForProvider.BucketRef,
		Selector:     in.Spec.ForProvider.BucketSelector,
		To:           reference.To{Managed: &v1alpha3.Bucket{}, List: &v1alpha3.BucketList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.bucket")
	}
	in.Spec.ForProvider.Bucket = reference.ToPtrValue(rsp.ResolvedValue)
	in.Spec.ForProvider.BucketRef = rsp.ResolvedReference

	// Resolve spec.forProvider.member
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(in.Spec.ForProvider.Member),
		Reference:    in.Spec.ForProvider.ServiceAccountMemberRef,
		Selector:     in.Spec.ForProvider.ServiceAccountMemberSelector,
		To:           reference.To{Managed: &v1alpha1.ServiceAccount{}, List: &v1alpha1.ServiceAccountList{}},
		Extract:      ServiceAccountMemberName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.member")
	}
	in.Spec.ForProvider.Member = reference.ToPtrValue(rsp.ResolvedValue)
	in.Spec.ForProvider.ServiceAccountMemberRef = rsp.ResolvedReference

	return nil
}
