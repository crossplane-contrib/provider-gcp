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

	v1beta12 "github.com/crossplane/provider-gcp/apis/classic/compute/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

// ResolveReferences of this CloudSQLInstance
func (mg *CloudSQLInstance) ResolveReferences(ctx context.Context, c client.Reader) error {

	if mg.Spec.ForProvider.Settings.IPConfiguration == nil {
		return nil
	}

	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.settings.ipConfiguration.privateNetwork
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Settings.IPConfiguration.PrivateNetwork),
		Reference:    mg.Spec.ForProvider.Settings.IPConfiguration.PrivateNetworkRef,
		Selector:     mg.Spec.ForProvider.Settings.IPConfiguration.PrivateNetworkSelector,
		To:           reference.To{Managed: &v1beta12.Network{}, List: &v1beta12.NetworkList{}},
		Extract:      v1beta12.NetworkURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.settings.ipConfiguration.privateNetwork")
	}
	mg.Spec.ForProvider.Settings.IPConfiguration.PrivateNetwork = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.Settings.IPConfiguration.PrivateNetworkRef = rsp.ResolvedReference

	return nil
}
