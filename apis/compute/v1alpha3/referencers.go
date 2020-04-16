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

package v1alpha3

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
)

// ResolveReferences of this GKECluster
func (mg *GKECluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.network
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.Network,
		Reference:    mg.Spec.NetworkRef,
		Selector:     mg.Spec.NetworkSelector,
		To:           reference.To{Managed: &v1beta1.Network{}, List: &v1beta1.NetworkList{}},
		Extract:      v1beta1.NetworkURL(),
	})
	if err != nil {
		return err
	}
	mg.Spec.Network = rsp.ResolvedValue
	mg.Spec.NetworkRef = rsp.ResolvedReference

	// Resolve spec.subnetwork
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.Subnetwork,
		Reference:    mg.Spec.SubnetworkRef,
		Selector:     mg.Spec.SubnetworkSelector,
		To:           reference.To{Managed: &v1beta1.Subnetwork{}, List: &v1beta1.SubnetworkList{}},
		Extract:      v1beta1.SubnetworkURL(),
	})
	if err != nil {
		return err
	}
	mg.Spec.Subnetwork = rsp.ResolvedValue
	mg.Spec.SubnetworkRef = rsp.ResolvedReference

	return nil
}
