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
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BackendServiceURL extracts the partially qualified URL of a BackendService.
func BackendServiceURL() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*BackendService)
		if !ok {
			return ""
		}
		return strings.TrimPrefix(n.Status.AtProvider.SelfLink, v1beta1.ComputeURIPrefix)
	}
}

// TargetTcpProxyURL extracts the partially qualified URL of a TargetTcpProxy.
func TargetTcpProxyURL() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		n, ok := mg.(*TargetTcpProxy)
		if !ok {
			return ""
		}
		return strings.TrimPrefix(n.Status.AtProvider.SelfLink, v1beta1.ComputeURIPrefix)
	}
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

// ResolveReferences of this BackendService
func (mg *BackendService) ResolveReferences(ctx context.Context, c client.Reader) error {
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

	// TODO(hasheddan): resolve references to health check and instance group
	return nil
}

// ResolveReferences of this TargetTCPProxy
func (mg *TargetTcpProxy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.service
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Service),
		Reference:    mg.Spec.ForProvider.ServiceRef,
		Selector:     mg.Spec.ForProvider.ServiceSelector,
		To:           reference.To{Managed: &BackendService{}, List: &BackendServiceList{}},
		Extract:      BackendServiceURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.service")
	}
	mg.Spec.ForProvider.Service = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ServiceRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this ForwardingRule
func (mg *ForwardingRule) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

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

	// Resolve spec.forProvider.backendService
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.BackendService),
		Reference:    mg.Spec.ForProvider.BackendServiceRef,
		Selector:     mg.Spec.ForProvider.BackendServiceSelector,
		To:           reference.To{Managed: &BackendService{}, List: &BackendServiceList{}},
		Extract:      BackendServiceURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.service")
	}
	mg.Spec.ForProvider.BackendService = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.BackendServiceRef = rsp.ResolvedReference

	// TODO(hasheddan): this can reference multiple resource types. This is a
	// hack for CAPI to ref the target TCP proxy.

	// Resolve spec.forProvider.target
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Target),
		Reference:    mg.Spec.ForProvider.TargetRef,
		Selector:     mg.Spec.ForProvider.TargetSelector,
		To:           reference.To{Managed: &TargetTcpProxy{}, List: &TargetTcpProxyList{}},
		Extract:      TargetTcpProxyURL(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.target")
	}
	mg.Spec.ForProvider.Target = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.TargetRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this InstanceGroup
func (mg *InstanceGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

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

// ResolveReferences of this Firewall
func (mg *Firewall) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

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
