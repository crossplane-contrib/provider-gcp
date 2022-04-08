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

	scv1alpha1 "github.com/crossplane/provider-gcp/apis/classic/v1alpha1"

	v1alpha12 "github.com/crossplane/provider-gcp/apis/classic/kms/v1alpha1"

	gcp "github.com/crossplane/provider-gcp/internal/classic/clients"
	"github.com/crossplane/provider-gcp/internal/classic/clients/keyring"
	"github.com/crossplane/provider-gcp/internal/features"

	kmsv1 "google.golang.org/api/cloudkms/v1"
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
)

// Error strings.
const (
	errNewClient  = "cannot create new GCP KMS API client"
	errNotKeyRing = "managed resource is not a GCP KeyRing"
	errGet        = "cannot get GCP object via KMS API"
	errCreate     = "cannot create GCP object via KMS API"
	errUpdate     = "cannot update GCP object via KMS API"
)

// SetupKeyRing adds a controller that reconciles KeyRings.
func SetupKeyRing(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha12.KeyRingGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha12.KeyRingGroupVersionKind),
		managed.WithExternalConnecter(&keyRingConnecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha12.KeyRing{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type keyRingConnecter struct {
	client client.Client
}

// Connect sets up kms client using credentials from the provider
func (c *keyRingConnecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha12.KeyRing)
	if !ok {
		return nil, errors.New(errNotKeyRing)
	}

	projectID, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := kmsv1.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	rrn := NewRelativeResourceNamerKeyRing(projectID, cr.Spec.ForProvider.Location)
	return &keyRingExternal{keyrings: kmsv1.NewProjectsLocationsKeyRingsService(s), rrn: rrn}, nil
}

type keyRingExternal struct {
	keyrings keyring.Client
	rrn      RelativeResourceNamerKeyRing
}

func (e *keyRingExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha12.KeyRing)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotKeyRing)
	}

	// Hack to cleanup CR without deleting actual resource.
	// It is not possible to delete KMS KeyRings, there is no "delete" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings
	// Also see related faq: https://cloud.google.com/kms/docs/faq#cannot_delete
	if meta.WasDeleted(cr) {
		return managed.ExternalObservation{}, nil
	}

	call := e.keyrings.Get(e.rrn.ResourceName(cr))
	instance, err := call.Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGet)
	}
	cr.Status.SetConditions(xpv1.Available())
	cr.Status.AtProvider = keyring.GenerateObservation(*instance)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings/create
func (e *keyRingExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha12.KeyRing)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotKeyRing)
	}
	cr.SetConditions(xpv1.Creating())
	instance := &kmsv1.KeyRing{}

	if _, err := e.keyrings.Create(e.rrn.LocationRRN(), instance).
		KeyRingId(meta.GetExternalName(cr)).Context(ctx).Do(); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	return managed.ExternalCreation{}, nil
}

func (e *keyRingExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// It is not possible to update KMS KeyRings, there is no "patch" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings
	return managed.ExternalUpdate{}, nil
}

func (e *keyRingExternal) Delete(ctx context.Context, mg resource.Managed) error {
	// It is not possible to delete KMS KeyRings, there is no "delete" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings
	// Also see related faq: https://cloud.google.com/kms/docs/faq#cannot_delete
	return nil
}

// NewRelativeResourceNamerKeyRing makes an instance of the RelativeResourceNamerKeyRing
// which is the only type that is allowed to know how to construct GCP resource names
// for the KMS Keyring type.
func NewRelativeResourceNamerKeyRing(projectName, location string) RelativeResourceNamerKeyRing {
	return RelativeResourceNamerKeyRing{projectName: projectName, location: location}
}

// RelativeResourceNamerKeyRing allows the controller to generate the "relative resource name"
// for the KeyRing and GCP project based on the keyRing external-name annotation.
// https://cloud.google.com/apis/design/resource_names#relative_resource_name
// The relative resource name for KeyRing has the following format:
// projects/{projectName}/locations/{location}/keyRings/{keyRingName}
type RelativeResourceNamerKeyRing struct {
	projectName string
	location    string
}

// ProjectRRN yields the relative resource name for a GCP project
func (rrn RelativeResourceNamerKeyRing) ProjectRRN() string {
	return fmt.Sprintf("projects/%s", rrn.projectName)
}

// LocationRRN yields the relative resource name for a GCP Project Location
func (rrn RelativeResourceNamerKeyRing) LocationRRN() string {
	return fmt.Sprintf("%s/locations/%s", rrn.ProjectRRN(), rrn.location)
}

// ResourceName yields the relative resource name for the KeyRing resource
func (rrn RelativeResourceNamerKeyRing) ResourceName(kr *v1alpha12.KeyRing) string {
	return fmt.Sprintf("%s/keyRings/%s",
		rrn.LocationRRN(), meta.GetExternalName(kr))
}
