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

package dns

import (
	"context"

	dns "google.golang.org/api/dns/v1"
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

	"github.com/crossplane-contrib/provider-gcp/apis/dns/v1alpha1"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	dnsclient "github.com/crossplane-contrib/provider-gcp/pkg/clients/dns"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
)

const (
	errGetPolicyFailed    = "cannot get the DNSPolicy"
	errNotDNSPolicy       = "managed resource is not a DNSPolicy custom resource"
	errCannotDeletePolicy = "cannot delete new DNSPolicy"
	errCreatePolicy       = "cannot create DNSPolicy"
	errCannotUpdate       = "Cannot update DNSPolicy"
)

func SetupPolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.PolicyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PolicyGroupVersionKind),
		managed.WithExternalConnecter(&Connector{kube: mgr.GetClient()}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.Policy{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type Connector struct {
	kube client.Client
}

func (c *Connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetConnectionInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}

	d, err := dns.NewService(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &policyExternal{
		kube:      c.kube,
		dns:       d.Policies,
		projectID: projectID,
	}, nil
}

type policyExternal struct {
	kube      client.Client
	dns       *dns.PoliciesService
	projectID string
}

func (e *policyExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDNSPolicy)
	}

	policy, err := e.dns.Get(
		e.projectID,
		meta.GetExternalName(cr),
	).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(
			resource.Ignore(gcp.IsErrorNotFound, err),
			errGetPolicyFailed,
		)
	}

	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider.ID = &policy.Id

	UpToDate, err := dnsclient.IsUptoDate(
		meta.GetExternalName(cr),
		&cr.Spec.ForProvider,
		policy)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: UpToDate,
	}, nil

}

func (e *policyExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDNSPolicy)
	}

	args := &dns.Policy{}

	dnsclient.GenerateDNSPolicy(
		meta.GetExternalName(cr),
		cr.Spec.ForProvider,
		args,
	)

	_, err := e.dns.Create(
		e.projectID,
		args,
	).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreatePolicy)
}

func (e *policyExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDNSPolicy)
	}

	args := &dns.Policy{}
	dnsclient.GenerateDNSPolicy(
		meta.GetExternalName(cr),
		cr.Spec.ForProvider,
		args,
	)
	_, err := e.dns.Patch(
		e.projectID,
		meta.GetExternalName(cr),
		args,
	).Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errCannotUpdate)
}

func (e *policyExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return errors.New(errNotDNSPolicy)
	}

	err := e.dns.Delete(
		e.projectID,
		meta.GetExternalName(cr),
	).Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errCannotDelete)

}
