/*
Copyright 2022 The Crossplane Authors.

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

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	computev1beta1 "github.com/crossplane-contrib/provider-gcp/apis/compute/v1beta1"
)

// ResolveReferences of ManagedZone
func (mg *ManagedZone) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.privateVisibilityConfig.networks[*].NetworkURL
	for i := range mg.Spec.ForProvider.PrivateVisibilityConfig.Networks {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.PrivateVisibilityConfig.Networks[i].NetworkURL),
			Reference:    mg.Spec.ForProvider.PrivateVisibilityConfig.Networks[i].NetworkRef,
			Selector:     mg.Spec.ForProvider.PrivateVisibilityConfig.Networks[i].NetworkSelector,
			To:           reference.To{Managed: &computev1beta1.Network{}, List: &computev1beta1.NetworkList{}},
			Extract:      computev1beta1.NetworkSelfLink(),
		})
		if err != nil {
			return errors.Wrapf(err, "spec.forProvider.PrivateVisibilityConfig.Networks[%d].NetworkURL", i)
		}
		mg.Spec.ForProvider.PrivateVisibilityConfig.Networks[i].NetworkURL = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.PrivateVisibilityConfig.Networks[i].NetworkRef = rsp.ResolvedReference
	}

	return nil
}
