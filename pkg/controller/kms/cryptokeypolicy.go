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

package kms

import (
	"context"

	"github.com/pkg/errors"
	kmsv1 "google.golang.org/api/cloudkms/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	iamv1alpha1 "github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane/provider-gcp/apis/kms/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/cryptokeypolicy"
)

const (
	errNotCryptoKeyPolicy = "managed resource is not a GCP CryptoKeyPolicy"
	errGetPolicy          = "cannot get policy of CryptoKey"
	errSetPolicy          = "cannot set policy of CryptoKey"
)

// SetupCryptoKeyPolicy adds a controller that reconciles CryptoKeyPolicys.
func SetupCryptoKeyPolicy(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CryptoKeyPolicyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CryptoKeyPolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CryptoKeyPolicyGroupVersionKind),
			managed.WithExternalConnecter(&cryptoKeyPolicyConnecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type cryptoKeyPolicyConnecter struct {
	client client.Client
}

// Connect sets up kms client using credentials from the provider
func (c *cryptoKeyPolicyConnecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := kmsv1.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &cryptoKeyPolicyExternal{kube: c.client, cryptokeyspolicy: kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)}, nil
}

type cryptoKeyPolicyExternal struct {
	kube             client.Client
	cryptokeyspolicy cryptokeypolicy.Client
}

func (e *cryptoKeyPolicyExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CryptoKeyPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCryptoKeyPolicy)
	}

	instance, err := e.cryptokeyspolicy.GetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.CryptoKey)).OptionsRequestedPolicyVersion(iamv1alpha1.PolicyVersion).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetPolicy)
	}

	if cryptokeypolicy.IsEmpty(instance) {
		// Empty policy
		return managed.ExternalObservation{}, nil
	}

	if upToDate, err := cryptokeypolicy.IsUpToDate(&cr.Spec.ForProvider, instance); err != nil {
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

func (e *cryptoKeyPolicyExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CryptoKeyPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCryptoKeyPolicy)
	}
	cr.SetConditions(xpv1.Creating())
	instance := &kmsv1.Policy{}
	cryptokeypolicy.GenerateCryptoKeyPolicyInstance(cr.Spec.ForProvider, instance)

	req := &kmsv1.SetIamPolicyRequest{Policy: instance}

	if _, err := e.cryptokeyspolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.CryptoKey), req).
		Context(ctx).Do(); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSetPolicy)
	}

	return managed.ExternalCreation{}, nil
}

func (e *cryptoKeyPolicyExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CryptoKeyPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCryptoKeyPolicy)
	}
	instance, err := e.cryptokeyspolicy.GetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.CryptoKey)).OptionsRequestedPolicyVersion(iamv1alpha1.PolicyVersion).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetPolicy)
	}

	u, err := cryptokeypolicy.IsUpToDate(&cr.Spec.ForProvider, instance)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckUpToDate)
	}
	if u {
		return managed.ExternalUpdate{}, nil
	}

	cryptokeypolicy.GenerateCryptoKeyPolicyInstance(cr.Spec.ForProvider, instance)
	req := &kmsv1.SetIamPolicyRequest{Policy: instance}

	if _, err := e.cryptokeyspolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.CryptoKey), req).
		Context(ctx).Do(); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errSetPolicy)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *cryptoKeyPolicyExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CryptoKeyPolicy)
	if !ok {
		return errors.New(errNotCryptoKeyPolicy)
	}
	req := &kmsv1.SetIamPolicyRequest{Policy: &kmsv1.Policy{}}
	if _, err := e.cryptokeyspolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.CryptoKey), req).
		Context(ctx).Do(); err != nil {
		return errors.Wrap(err, errSetPolicy)
	}
	return nil
}
