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

package serviceaccountpolicy

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	"google.golang.org/api/iam/v1"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
)

const (
	// https://cloud.google.com/iam/docs/reference/rest/v1/Policy
	// Specifies the format of the policy.
	// Any operation that affects conditional role bindings must specify version 3.
	// Our CR supports conditional role bindings.
	policyVersion = 3
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// Client should be satisfied to conduct SA operations.
type Client interface {
	GetIamPolicy(resource string) *iam.ProjectsServiceAccountsGetIamPolicyCall
	SetIamPolicy(resource string, setiampolicyrequest *iam.SetIamPolicyRequest) *iam.ProjectsServiceAccountsSetIamPolicyCall
}

// GenerateServiceAccountPolicyInstance generates *iam.Policy instance from ServiceAccountPolicyParameters.
func GenerateServiceAccountPolicyInstance(in v1alpha1.ServiceAccountPolicyParameters, p *iam.Policy) {
	p.Bindings = make([]*iam.Binding, len(in.Policy.Bindings))
	for i, v := range in.Policy.Bindings {
		p.Bindings[i] = &iam.Binding{}
		if v.Condition != nil {
			p.Bindings[i].Condition = &iam.Expr{
				Description: v.Condition.Description,
				Expression:  v.Condition.Expression,
				Location:    v.Condition.Location,
				Title:       v.Condition.Title,
			}
		}
		p.Bindings[i].Members = make([]string, len(v.Members))
		copy(p.Bindings[i].Members, v.Members)
		p.Bindings[i].Role = v.Role
	}
	p.AuditConfigs = make([]*iam.AuditConfig, len(in.Policy.AuditConfigs))
	for i, v := range in.Policy.AuditConfigs {
		p.AuditConfigs[i] = &iam.AuditConfig{}
		p.AuditConfigs[i].Service = v.Service
		p.AuditConfigs[i].AuditLogConfigs = make([]*iam.AuditLogConfig, len(v.AuditLogConfigs))
		for ai, av := range v.AuditLogConfigs {
			p.AuditConfigs[i].AuditLogConfigs[ai] = &iam.AuditLogConfig{}
			p.AuditConfigs[i].AuditLogConfigs[ai].LogType = av.LogType
			p.AuditConfigs[i].AuditLogConfigs[ai].ExemptedMembers = make([]string, len(av.ExemptedMembers))
			copy(p.AuditConfigs[i].AuditLogConfigs[ai].ExemptedMembers, av.ExemptedMembers)
		}
	}
	p.Version = policyVersion
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1alpha1.ServiceAccountPolicyParameters, observed *iam.Policy) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*iam.Policy)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateServiceAccountPolicyInstance(*in, desired)
	return ArePoliciesSame(desired, observed), nil
}

// ArePoliciesSame compares and returns true if two policies are same
func ArePoliciesSame(p1, p2 *iam.Policy) bool {
	return cmp.Equal(p1, p2, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(iam.Policy{}, "Version"),
		cmpopts.SortSlices(func(i, j *iam.Binding) bool { return i.Role > j.Role }),
		cmpopts.SortSlices(func(i, j string) bool { return i > j }))
}

// IsEmpty returns if Policy is empty
func IsEmpty(in *iam.Policy) bool {
	return in.Bindings == nil && in.AuditConfigs == nil
}
