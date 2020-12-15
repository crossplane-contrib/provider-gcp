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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/pkg/errors"
	kmsv1 "google.golang.org/api/cloudkms/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/kms/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/keyring"
)

// Error strings.
const (
	errNewClient  = "cannot create new GCP KMS API client"
	errNotKeyRing = "managed resource is not a GCP KeyRing"
	errGet        = "cannot get GCP KeyRing object via KMS API"
	errCreate     = "cannot create GCP KeyRing object via KMS API"
)

// SetupKeyRing adds a controller that reconciles KeyRings.
func SetupKeyRing(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.KeyRingGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.KeyRing{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.KeyRingGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

// Connect sets up kms client using credentials from the provider
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.KeyRing)
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
	rrn := NewRelativeResourceNamer(projectID, cr.Spec.ForProvider.Location)
	return &external{keyrings: kmsv1.NewProjectsLocationsKeyRingsService(s), rrn: rrn}, errors.Wrap(err, errNewClient)
}

type external struct {
	keyrings keyring.Client
	rrn      RelativeResourceNamer
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.KeyRing)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotKeyRing)
	}

	call := e.keyrings.Get(e.rrn.ResourceName(cr))
	fromProvider, err := call.Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGet)
	}
	cr.Status.SetConditions(xpv1.Available())
	populateCRFromProvider(cr, fromProvider)

	exists := true
	// Hack to cleanup CR without deleting actual resource.
	// It is not possible to delete KMS KeyRings, there is no "delete" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings
	// Also see related faq: https://cloud.google.com/kms/docs/faq#cannot_delete
	if meta.WasDeleted(cr) {
		exists = false
	}

	return managed.ExternalObservation{
		ResourceExists:    exists,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings/create
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.KeyRing)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotKeyRing)
	}

	ckrr := &kmsv1.KeyRing{}
	call := e.keyrings.Create(e.rrn.Location(), ckrr)

	fromProvider, err := call.KeyRingId(meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	populateCRFromProvider(cr, fromProvider)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// It is not possible to update KMS KeyRings, there is no "patch" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	// It is not possible to delete KMS KeyRings, there is no "delete" method defined:
	// https://cloud.google.com/kms/docs/reference/rest#rest-resource:-v1.projects.locations.keyrings
	// Also see related faq: https://cloud.google.com/kms/docs/faq#cannot_delete
	return nil
}

// NewRelativeResourceNamer makes an instance of the RelativeResourceNamer
// which is the only type that is allowed to know how to construct GCP resource names
// for the KMS Keyring type.
func NewRelativeResourceNamer(projectName, location string) RelativeResourceNamer {
	return RelativeResourceNamer{projectName: projectName, location: location}
}

// RelativeResourceNamer allows the controller to generate the "relative resource name"
// for the service account and GCP project based on the external-name annotation.
// https://cloud.google.com/apis/design/resource_names#relative_resource_name
// The relative resource name for service accounts has the following format:
// projects/{projectName}/locations/{location}
type RelativeResourceNamer struct {
	projectName string
	location    string
}

// ProjectName yields the relative resource name for a GCP project
func (rrn RelativeResourceNamer) ProjectName() string {
	return fmt.Sprintf("projects/%s", rrn.projectName)
}

// Location yields the relative resource name for a GCP Project Location
func (rrn RelativeResourceNamer) Location() string {
	return fmt.Sprintf("projects/%s/locations/%s", rrn.projectName, rrn.location)
}

// ResourceName yields the relative resource name for the KeyRing resource
func (rrn RelativeResourceNamer) ResourceName(kr *v1alpha1.KeyRing) string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s",
		rrn.projectName, rrn.location, meta.GetExternalName(kr))
}

func populateCRFromProvider(cr *v1alpha1.KeyRing, fromProvider *kmsv1.KeyRing) {
	cr.Status.AtProvider.Name = fromProvider.Name
	cr.Status.AtProvider.CreateTime = fromProvider.CreateTime
}
