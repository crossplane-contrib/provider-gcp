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

	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// ServiceAccountReferer defines a reference to a ServiceAccount either via its RRN,
// or via a v1alpha1.ServiceAccount object or via a selector. RRN is the
// relative resource name as defined by Google Cloud API design docs here:
// https://cloud.google.com/apis/design/resource_names#relative_resource_name
// An example value for the ServiceAccount field is as follows:
// projects/<project-name>>/serviceAccounts/perfect-test-sa@crossplane-playground.iam.gserviceaccount.com
type ServiceAccountReferer struct {
	// ServiceAccount: The RRN of the referred ServiceAccount
	// RRN is the relative resource name as defined by Google Cloud API design docs here:
	// https://cloud.google.com/apis/design/resource_names#relative_resource_name
	// An example value for the ServiceAccount field is as follows:
	// projects/<project-name>/serviceAccounts/perfect-test-sa@crossplane-playground.iam.gserviceaccount.com
	// +optional
	// +immutable
	ServiceAccount *string `json:"serviceAccount,omitempty"`

	// ServiceAccountRef references a ServiceAccount and retrieves its URI
	// +optional
	// +immutable
	ServiceAccountRef *xpv1.Reference `json:"serviceAccountRef,omitempty"`

	// ServiceAccountSelector selects a reference to a ServiceAccount
	// +optional
	ServiceAccountSelector *xpv1.Selector `json:"serviceAccountSelector,omitempty"`
}

// ServiceAccountRRN extracts the partially qualified URL of a Network.
func ServiceAccountRRN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*ServiceAccount)
		if !ok {
			return ""
		}
		return n.Status.AtProvider.Name
	}
}

// ServiceAccountMemberName returns member name for a given ServiceAccount Object.
func ServiceAccountMemberName() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*ServiceAccount)
		if !ok {
			return ""
		}
		if n.Status.AtProvider.Email == "" {
			return ""
		}
		return fmt.Sprintf("serviceAccount:%s", n.Status.AtProvider.Email)
	}
}

func (sar *ServiceAccountReferer) resolveReferences(ctx context.Context, resolver *reference.APIResolver) error {
	// Resolve spec.forProvider.serviceAccount
	rsp, err := resolver.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(sar.ServiceAccount),
		Reference:    sar.ServiceAccountRef,
		Selector:     sar.ServiceAccountSelector,
		To:           reference.To{Managed: &ServiceAccount{}, List: &ServiceAccountList{}},
		Extract:      ServiceAccountRRN(),
	})

	if err != nil {
		return err
	}

	sar.ServiceAccount = reference.ToPtrValue(rsp.ResolvedValue)
	sar.ServiceAccountRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this ServiceAccountKey
func (in *ServiceAccountKey) ResolveReferences(ctx context.Context, c client.Reader) error {
	return errors.Wrap(in.Spec.ForProvider.resolveReferences(ctx, reference.NewAPIResolver(c, in)), "spec.forProvider.serviceAccount")
}

// ResolveReferences of this ServiceAccountPolicy
func (in *ServiceAccountPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, in)

	if err := in.Spec.ForProvider.resolveReferences(ctx, r); err != nil {
		return errors.Wrap(err, "spec.forProvider.serviceAccount")
	}

	// Resolve spec.ForProvider.Policy.Bindings[*].Members
	for i := range in.Spec.ForProvider.Policy.Bindings {
		mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
			CurrentValues: in.Spec.ForProvider.Policy.Bindings[i].Members,
			References:    in.Spec.ForProvider.Policy.Bindings[i].ServiceAccountMemberRefs,
			Selector:      in.Spec.ForProvider.Policy.Bindings[i].ServiceAccountMemberSelector,
			To:            reference.To{Managed: &ServiceAccount{}, List: &ServiceAccountList{}},
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
