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
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// NetworkURL extracts the partially qualified URL of a Network.
func NetworkURL() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*Network)
		if !ok {
			return ""
		}
		return strings.TrimPrefix(n.Status.AtProvider.SelfLink, ComputeURIPrefix)
	}
}

// SubnetworkURL extracts the partially qualified URL of a Subnetwork.
func SubnetworkURL() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		sn, ok := mg.(*Subnetwork)
		if !ok {
			return ""
		}
		return strings.TrimPrefix(sn.Status.AtProvider.SelfLink, ComputeURIPrefix)
	}
}

// ResolveReferences of this GlobalAddress
func (mg *GlobalAddress) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.network
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Network),
		Reference:    mg.Spec.ForProvider.NetworkRef,
		Selector:     mg.Spec.ForProvider.NetworkSelector,
		To:           reference.To{Managed: &Network{}, List: &NetworkList{}},
		Extract:      NetworkURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.network")
	}
	mg.Spec.ForProvider.Network = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NetworkRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Subnetwork
func (mg *Subnetwork) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.network
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Network),
		Reference:    mg.Spec.ForProvider.NetworkRef,
		Selector:     mg.Spec.ForProvider.NetworkSelector,
		To:           reference.To{Managed: &Network{}, List: &NetworkList{}},
		Extract:      NetworkURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.network")
	}
	mg.Spec.ForProvider.Network = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NetworkRef = rsp.ResolvedReference

	return nil
}
