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

package v1beta2

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
)

// ClusterURL extracts the partially qualified URL of a Cluster.
func ClusterURL() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		c, ok := mg.(*Cluster)
		if !ok {
			return ""
		}
		return strings.TrimPrefix(c.Status.AtProvider.SelfLink, ContainerURIPrefix)
	}
}

// ResolveReferences of this Cluster
func (mg *Cluster) ResolveReferences(ctx context.Context, c client.Reader) error {
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

	// Resolve spec.forProvider.subnetwork
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Subnetwork),
		Reference:    mg.Spec.ForProvider.SubnetworkRef,
		Selector:     mg.Spec.ForProvider.SubnetworkSelector,
		To:           reference.To{Managed: &v1beta1.Subnetwork{}, List: &v1beta1.SubnetworkList{}},
		Extract:      v1beta1.SubnetworkURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.subnetwork")
	}
	mg.Spec.ForProvider.Subnetwork = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.SubnetworkRef = rsp.ResolvedReference

	return nil
}
