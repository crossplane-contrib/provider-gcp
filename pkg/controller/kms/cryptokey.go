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
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	kmsv1 "google.golang.org/api/cloudkms/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/kms/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/cryptokey"
)

const (
	errNotCryptoKey        = "managed resource is not a GCP CryptoKey"
	errManagedUpdateFailed = "cannot update CryptoKey custom resource"
	errCheckUpToDate       = "cannot determine if CryptoKey instance is up to date"
)

// SetupCryptoKey adds a controller that reconciles CryptoKeys.
func SetupCryptoKey(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CryptoKeyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CryptoKey{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CryptoKeyGroupVersionKind),
			managed.WithExternalConnecter(&cryptoKeyConnecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type cryptoKeyConnecter struct {
	client client.Client
}

// Connect sets up kms client using credentials from the provider
func (c *cryptoKeyConnecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := kmsv1.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &cryptoKeyExternal{kube: c.client, cryptokeys: kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)}, nil
}

type cryptoKeyExternal struct {
	kube       client.Client
	cryptokeys cryptokey.Client
}

func (e *cryptoKeyExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CryptoKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCryptoKey)
	}

	// Hack to cleanup CR without deleting actual resource.
	// It is not possible to delete KMS CryptoKeys, there is no "delete" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings.cryptokeys
	// Also see related faq: https://cloud.google.com/kms/docs/faq#cannot_delete
	if meta.WasDeleted(cr) {
		return managed.ExternalObservation{
			ResourceExists:    false,
			ResourceUpToDate:  false,
			ConnectionDetails: managed.ConnectionDetails{},
		}, nil
	}

	instance, err := e.cryptokeys.Get(cryptoKeyRRN(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGet)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	cryptokey.LateInitializeSpec(&cr.Spec.ForProvider, *instance)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider = cryptokey.GenerateObservation(*instance)
	cr.Status.SetConditions(xpv1.Available())

	upToDate, _, err := cryptokey.IsUpToDate(&cr.Spec.ForProvider, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *cryptoKeyExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CryptoKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCryptoKey)
	}
	cr.SetConditions(xpv1.Creating())
	instance := &kmsv1.CryptoKey{}
	cryptokey.GenerateCryptoKeyInstance(cr.Spec.ForProvider, instance)

	if _, err := e.cryptokeys.Create(gcp.StringValue(cr.Spec.ForProvider.KeyRing), instance).
		CryptoKeyId(meta.GetExternalName(cr)).Context(ctx).Do(); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	return managed.ExternalCreation{}, nil
}

func (e *cryptoKeyExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CryptoKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCryptoKey)
	}
	// We have to get the cluster again here to calculate update mask (what to patch).
	instance, err := e.cryptokeys.Get(cryptoKeyRRN(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGet)
	}

	u, um, err := cryptokey.IsUpToDate(&cr.Spec.ForProvider, instance)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckUpToDate)
	}
	if u {
		return managed.ExternalUpdate{}, nil
	}

	cryptokey.GenerateCryptoKeyInstance(cr.Spec.ForProvider, instance)
	if _, err := e.cryptokeys.Patch(cryptoKeyRRN(cr), instance).UpdateMask(um).
		Context(ctx).Do(); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *cryptoKeyExternal) Delete(ctx context.Context, mg resource.Managed) error {
	// It is not possible to delete KMS CryptoKeys, there is no "delete" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings.cryptokeys
	// Also see related faq: https://cloud.google.com/kms/docs/faq#cannot_delete
	return nil
}

func cryptoKeyRRN(cr *v1alpha1.CryptoKey) string {
	return fmt.Sprintf("%s/cryptoKeys/%s", gcp.StringValue(cr.Spec.ForProvider.KeyRing), meta.GetExternalName(cr))
}
