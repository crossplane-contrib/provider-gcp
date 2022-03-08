/*
Copyright 2019 The Crossplane Authors.

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

package iam

import (
	"context"

	iamv1 "google.golang.org/api/iam/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	scv1alpha1 "github.com/crossplane/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/serviceaccountpolicy"
	"github.com/crossplane/provider-gcp/pkg/features"
)

const (
	errNotServiceAccountPolicy = "managed resource is not a GCP ServiceAccountPolicy"
	errCheckUpToDate           = "cannot determine if ServiceAccountPolicy instance is up to date"

	errGetPolicy = "cannot get policy of CryptoKey"
	errSetPolicy = "cannot set policy of CryptoKey"
)

// SetupServiceAccountPolicy adds a controller that reconciles ServiceAccountPolicys.
func SetupServiceAccountPolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ServiceAccountPolicyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ServiceAccountPolicyGroupVersionKind),
		managed.WithExternalConnecter(&serviceAccountPolicyConnecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.ServiceAccountPolicy{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type serviceAccountPolicyConnecter struct {
	client client.Client
}

// Connect sets up iam client using credentials from the provider
func (c *serviceAccountPolicyConnecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := iamv1.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &serviceAccountPolicyExternal{kube: c.client, serviceaccountspolicy: iamv1.NewProjectsServiceAccountsService(s)}, nil
}

type serviceAccountPolicyExternal struct {
	kube                  client.Client
	serviceaccountspolicy serviceaccountpolicy.Client
}

func (e *serviceAccountPolicyExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccountPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotServiceAccountPolicy)
	}

	instance, err := e.serviceaccountspolicy.GetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.ServiceAccount)).OptionsRequestedPolicyVersion(v1alpha1.PolicyVersion).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetPolicy)
	}
	// Empty policy
	if serviceaccountpolicy.IsEmpty(instance) {
		return managed.ExternalObservation{}, nil
	}

	if upToDate, err := serviceaccountpolicy.IsUpToDate(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	} else if !upToDate {
		return managed.ExternalObservation{ResourceExists: true}, nil
	}

	cr.Status.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *serviceAccountPolicyExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccountPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotServiceAccountPolicy)
	}
	cr.SetConditions(xpv1.Creating())
	instance := &iamv1.Policy{}
	serviceaccountpolicy.GenerateServiceAccountPolicyInstance(cr.Spec.ForProvider, instance)

	req := &iamv1.SetIamPolicyRequest{Policy: instance}

	if _, err := e.serviceaccountspolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.ServiceAccount), req).
		Context(ctx).Do(); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSetPolicy)
	}

	return managed.ExternalCreation{}, nil
}

func (e *serviceAccountPolicyExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccountPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotServiceAccountPolicy)
	}
	instance, err := e.serviceaccountspolicy.GetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.ServiceAccount)).OptionsRequestedPolicyVersion(v1alpha1.PolicyVersion).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetPolicy)
	}

	u, err := serviceaccountpolicy.IsUpToDate(&cr.Spec.ForProvider, instance)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckUpToDate)
	}
	if u {
		return managed.ExternalUpdate{}, nil
	}

	serviceaccountpolicy.GenerateServiceAccountPolicyInstance(cr.Spec.ForProvider, instance)
	req := &iamv1.SetIamPolicyRequest{Policy: instance}

	if _, err := e.serviceaccountspolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.ServiceAccount), req).
		Context(ctx).Do(); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errSetPolicy)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *serviceAccountPolicyExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ServiceAccountPolicy)
	if !ok {
		return errors.New(errNotServiceAccountPolicy)
	}
	req := &iamv1.SetIamPolicyRequest{Policy: &iamv1.Policy{}}
	if _, err := e.serviceaccountspolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.ServiceAccount), req).
		Context(ctx).Do(); err != nil {
		return errors.Wrap(err, errSetPolicy)
	}
	return nil
}
