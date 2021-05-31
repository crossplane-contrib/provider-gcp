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

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-gcp/apis/container/v1beta2"
)

// ResolveReferences of this NodePool
func (mg *NodePool) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.cluster
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.Cluster,
		Reference:    mg.Spec.ForProvider.ClusterRef,
		Selector:     mg.Spec.ForProvider.ClusterSelector,
		To:           reference.To{Managed: &v1beta2.GKECluster{}, List: &v1beta2.GKEClusterList{}},
		Extract:      v1beta2.GKEClusterURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.cluster")
	}
	mg.Spec.ForProvider.Cluster = rsp.ResolvedValue
	mg.Spec.ForProvider.ClusterRef = rsp.ResolvedReference

	return nil
}
