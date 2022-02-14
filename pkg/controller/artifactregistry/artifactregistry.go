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

package artifactregistry

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/google/go-cmp/cmp"
	artifactregistry "google.golang.org/api/artifactregistry/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/artifactregistry/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	repository "github.com/crossplane/provider-gcp/pkg/clients/artifactregistry"
)

const (
	errNewClient        = "cannot create new Artifact registry Service"
	errNotRepository    = "managed resource is not of type Repository"
	errCreateRepository = "cannot create Repository"
	errGetRepository    = "cannot get Repository"
	errUpdateRepository = "cannot update Repository custom resource"
	errDeleteRepository = "cannot delete Repository"
)

// SetupRepository adds a controller that reconciles Repositories.
func SetupRepository(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RepositoryGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RepositoryGroupVersionKind),
		managed.WithExternalConnecter(&repositoryConnecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.Repository{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type repositoryConnecter struct {
	client client.Client
}

// Connect returns an ExternalClient with necessary information to talk to GCP API.
func (c *repositoryConnecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}

	s, err := artifactregistry.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &repositoryExternal{projectID: projectID, client: c.client, ps: s}, nil
}

type repositoryExternal struct {
	projectID string
	client    client.Client
	ps        *artifactregistry.Service
}

// Observe makes observation about the external resource.
func (e *repositoryExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRepository)
	}

	r, err := e.ps.Projects.Locations.Repositories.Get(repository.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider.Location, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetRepository)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	repository.LateInitialize(&cr.Spec.ForProvider, *r)

	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.client.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateRepository)
		}
	}

	cr.Status.AtProvider.CreateTime = r.CreateTime
	cr.Status.AtProvider.UpdateTime = r.UpdateTime
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: repository.IsUpToDate(e.projectID, cr.Spec.ForProvider, *r),
	}, nil
}

// Create initiates creation of external resource.
func (e *repositoryExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRepository)
	}

	cr.SetConditions(xpv1.Creating())

	parent := repository.GetFullyQualifiedParent(e.projectID, cr.Spec.ForProvider.Location)
	desired := repository.GenerateRepository(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider)
	_, err := e.ps.Projects.Locations.Repositories.Create(parent,
		desired).RepositoryId(meta.GetExternalName(cr)).Context(ctx).Do()

	return managed.ExternalCreation{}, errors.Wrap(err, errCreateRepository)
}

// Update initiates an update to the external resource.
func (e *repositoryExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRepository)
	}

	r, err := e.ps.Projects.Locations.Repositories.Get(repository.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider.Location, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetRepository)
	}

	updated, updateMask := repository.GenerateUpdateRequest(meta.GetExternalName(cr), cr.Spec.ForProvider, *r)
	_, err = e.ps.Projects.Locations.Repositories.Patch(repository.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider.Location, meta.GetExternalName(cr)),
		updated).UpdateMask(updateMask).Context(ctx).Do()

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateRepository)
}

// Delete initiates an deletion of the external resource.
func (e *repositoryExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return errors.New(errNotRepository)
	}

	_, err := e.ps.Projects.Locations.Repositories.Delete(repository.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider.Location,
		meta.GetExternalName(cr))).Context(ctx).Do()

	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteRepository)
}
