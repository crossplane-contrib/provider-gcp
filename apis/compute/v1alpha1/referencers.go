/*
Copyright 2021 The Crossplane Authors.

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

	"github.com/crossplane-contrib/provider-gcp/apis/compute/v1beta1"
)

// ResolveReferences of this Firewall
func (mg *Firewall) ResolveReferences(ctx context.Context, c client.Reader) error {
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
		return errors.Wrap(err, "spec.forProvider.network")
	}
	mg.Spec.ForProvider.Network = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NetworkRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Router
func (mg *Router) ResolveReferences(ctx context.Context, c client.Reader) error {
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
		return errors.Wrap(err, "spec.forProvider.network")
	}
	mg.Spec.ForProvider.Network = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NetworkRef = rsp.ResolvedReference

	return nil
}
