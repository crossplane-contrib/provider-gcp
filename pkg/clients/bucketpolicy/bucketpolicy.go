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

package bucketpolicy

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	"google.golang.org/api/storage/v1"

	iamv1alpha1 "github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane/provider-gcp/apis/storage/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// Client should be satisfied to conduct Bucket Policy operations.
type Client interface {
	GetIamPolicy(bucket string) *storage.BucketsGetIamPolicyCall
	SetIamPolicy(bucket string, policy *storage.Policy) *storage.BucketsSetIamPolicyCall
}

// GenerateBucketPolicyInstance generates *storage.Policy instance from BucketPolicyParameters.
func GenerateBucketPolicyInstance(in v1alpha1.BucketPolicyParameters, ck *storage.Policy) {
	ck.Bindings = make([]*storage.PolicyBindings, len(in.Policy.Bindings))
	for i, v := range in.Policy.Bindings {
		ck.Bindings[i] = &storage.PolicyBindings{}
		if v.Condition != nil {
			ck.Bindings[i].Condition = &storage.Expr{
				Description: gcp.StringValue(v.Condition.Description),
				Expression:  v.Condition.Expression,
				Location:    gcp.StringValue(v.Condition.Location),
				Title:       gcp.StringValue(v.Condition.Title),
			}
		}
		ck.Bindings[i].Members = make([]string, len(v.Members))
		copy(ck.Bindings[i].Members, v.Members)
		ck.Bindings[i].Role = v.Role
	}
	ck.Version = iamv1alpha1.PolicyVersion
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1alpha1.BucketPolicyParameters, observed *storage.Policy) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*storage.Policy)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateBucketPolicyInstance(*in, desired)
	return ArePoliciesSame(desired, observed), nil
}

// ArePoliciesSame compares and returns true if two policies are same
func ArePoliciesSame(p1, p2 *storage.Policy) bool {
	return cmp.Equal(p1, p2, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(storage.Policy{}, "Version"),
		cmpopts.SortSlices(func(i, j *storage.PolicyBindings) bool { return i.Role > j.Role }),
		cmpopts.SortSlices(func(i, j string) bool { return i > j }))
}

// IsEmpty returns if Policy is empty
func IsEmpty(in *storage.Policy) bool {
	return in.Bindings == nil
}

// BindRoleToMember updates *storage.Policy instance with BucketPolicyMemberParameters.
// returns true if policy changed
func BindRoleToMember(in v1alpha1.BucketPolicyMemberParameters, ck *storage.Policy) bool {
	ck.Version = iamv1alpha1.PolicyVersion
	for _, b := range ck.Bindings {
		if b.Role == in.Role {
			for _, m := range b.Members {
				if m == gcp.StringValue(in.Member) {
					// role already bound to member, no change
					return false
				}
			}
			// role already exist, add member
			b.Members = append(b.Members, gcp.StringValue(in.Member))
			return true
		}
	}
	// role does not exist, add binding with role and member
	ck.Bindings = append(ck.Bindings, &storage.PolicyBindings{
		Role:    in.Role,
		Members: []string{gcp.StringValue(in.Member)},
	})
	return true
}

// UnbindRoleFromMember generates *storage.Policy instance from BucketPolicyMemberParameters.
// returns true if bound (i.e. policy changed)
func UnbindRoleFromMember(in v1alpha1.BucketPolicyMemberParameters, ck *storage.Policy) bool {
	for _, b := range ck.Bindings {
		if b.Role == in.Role {
			ix := -1
			for i, m := range b.Members {
				if m == gcp.StringValue(in.Member) {
					// found member binding in role
					ix = i
					break
				}
			}
			if ix >= 0 {
				// remove member located at index ix
				b.Members = append(b.Members[:ix], b.Members[ix+1:]...)
				return true
			}
			return false
		}
	}
	return false
}
