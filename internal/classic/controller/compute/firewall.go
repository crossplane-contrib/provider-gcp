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

package compute

import (
	"context"

	v1alpha12 "github.com/crossplane/provider-gcp/apis/classic/compute/v1alpha1"

	gcp "github.com/crossplane/provider-gcp/internal/classic/clients"
	"github.com/crossplane/provider-gcp/internal/classic/clients/firewall"
	"github.com/crossplane/provider-gcp/internal/features"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	scv1alpha1 "github.com/crossplane/provider-gcp/apis/v1alpha1"
)

const (
	// Error strings.
	errNotFirewall = "managed resource is not a Firewall resource"
	errGetFirewall = "cannot get GCP Firewall"

	errFirewallUpdateFailed  = "update of Firewall resource has failed"
	errFirewallCreateFailed  = "creation of Firewall resource has failed"
	errFirewallDeleteFailed  = "deletion of Firewall resource has failed"
	errCheckFirewallUpToDate = "cannot determine if GCP Firewall is up to date"
)

// SetupFirewall adds a controller that reconciles Firewall managed
// resources.
func SetupFirewall(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha12.FirewallGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha12.FirewallGroupVersionKind),
		managed.WithExternalConnecter(&firewallConnector{kube: mgr.GetClient()}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha12.Firewall{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type firewallConnector struct {
	kube client.Client
}

func (c *firewallConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := compute.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &firewallExternal{Service: s, kube: c.kube, projectID: projectID}, nil
}

type firewallExternal struct {
	kube client.Client
	*compute.Service
	projectID string
}

func (c *firewallExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha12.Firewall)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotFirewall)
	}
	observed, err := c.Firewalls.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetFirewall)
	}

	lateIntialized := false
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	firewall.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		lateIntialized = true
	}

	cr.Status.AtProvider = firewall.GenerateFirewallObservation(*observed)

	cr.Status.SetConditions(xpv1.Available())

	u, err := firewall.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckFirewallUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: lateIntialized,
		ResourceUpToDate:        u,
	}, nil
}

func (c *firewallExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha12.Firewall)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotFirewall)
	}

	fw := &compute.Firewall{}
	firewall.GenerateFirewall(meta.GetExternalName(cr), cr.Spec.ForProvider, fw)
	_, err := c.Firewalls.Insert(c.projectID, fw).
		Context(ctx).
		Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errFirewallCreateFailed)
}

func (c *firewallExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha12.Firewall)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotFirewall)
	}

	observed, err := c.Firewalls.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetFirewall)
	}

	upToDate, err := firewall.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckFirewallUpToDate)
	}
	if upToDate {
		return managed.ExternalUpdate{}, nil
	}

	fw := &compute.Firewall{}
	firewall.GenerateFirewall(meta.GetExternalName(cr), cr.Spec.ForProvider, fw)

	_, err = c.Firewalls.Patch(c.projectID, meta.GetExternalName(cr), fw).
		Context(ctx).
		Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errFirewallUpdateFailed)
}

func (c *firewallExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha12.Firewall)
	if !ok {
		return errors.New(errNotFirewall)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := c.Firewalls.Delete(c.projectID, meta.GetExternalName(cr)).
		Context(ctx).
		Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errFirewallDeleteFailed)
}
