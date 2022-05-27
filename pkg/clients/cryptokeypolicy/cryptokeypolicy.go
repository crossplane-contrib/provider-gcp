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

package cryptokeypolicy

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"google.golang.org/api/cloudkms/v1"

	"github.com/crossplane/crossplane-runtime/pkg/errors"

	iamv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane-contrib/provider-gcp/apis/kms/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// Client should be satisfied to conduct SA operations.
type Client interface {
	GetIamPolicy(resource string) *cloudkms.ProjectsLocationsKeyRingsCryptoKeysGetIamPolicyCall
	SetIamPolicy(resource string, setiampolicyrequest *cloudkms.SetIamPolicyRequest) *cloudkms.ProjectsLocationsKeyRingsCryptoKeysSetIamPolicyCall
}

// GenerateCryptoKeyPolicyInstance generates *kmsv1.Policy instance from CryptoKeyPolicyParameters.
func GenerateCryptoKeyPolicyInstance(in v1alpha1.CryptoKeyPolicyParameters, ck *cloudkms.Policy) {
	ck.Bindings = make([]*cloudkms.Binding, len(in.Policy.Bindings))
	for i, v := range in.Policy.Bindings {
		ck.Bindings[i] = &cloudkms.Binding{}
		if v.Condition != nil {
			ck.Bindings[i].Condition = &cloudkms.Expr{
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
	ck.AuditConfigs = make([]*cloudkms.AuditConfig, len(in.Policy.AuditConfigs))
	for i, v := range in.Policy.AuditConfigs {
		ck.AuditConfigs[i] = &cloudkms.AuditConfig{}
		ck.AuditConfigs[i].Service = v.Service
		ck.AuditConfigs[i].AuditLogConfigs = make([]*cloudkms.AuditLogConfig, len(v.AuditLogConfigs))
		for ai, av := range v.AuditLogConfigs {
			ck.AuditConfigs[i].AuditLogConfigs[ai] = &cloudkms.AuditLogConfig{}
			ck.AuditConfigs[i].AuditLogConfigs[ai].LogType = av.LogType
			ck.AuditConfigs[i].AuditLogConfigs[ai].ExemptedMembers = make([]string, len(av.ExemptedMembers))
			copy(ck.AuditConfigs[i].AuditLogConfigs[ai].ExemptedMembers, av.ExemptedMembers)
		}
	}
	ck.Version = iamv1alpha1.PolicyVersion
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1alpha1.CryptoKeyPolicyParameters, observed *cloudkms.Policy) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*cloudkms.Policy)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateCryptoKeyPolicyInstance(*in, desired)
	return ArePoliciesSame(desired, observed), nil
}

// ArePoliciesSame compares and returns true if two policies are same
func ArePoliciesSame(p1, p2 *cloudkms.Policy) bool {
	return cmp.Equal(p1, p2, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(cloudkms.Policy{}, "Version"),
		cmpopts.SortSlices(func(i, j *cloudkms.Binding) bool { return i.Role > j.Role }),
		cmpopts.SortSlices(func(i, j string) bool { return i > j }))
}

// IsEmpty returns if Policy is empty
func IsEmpty(in *cloudkms.Policy) bool {
	return in.Bindings == nil && in.AuditConfigs == nil
}
