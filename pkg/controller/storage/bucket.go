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

package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/imdario/mergo"
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

	"github.com/crossplane-contrib/provider-gcp/apis/storage/v1alpha3"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
)

// Error strings.
const (
	errNewClient = "cannot create new GCP storage client"
	errNotBucket = "managed resource is not a GCP bucket"
	errAttrs     = "cannot get GCP bucket attributes"
	errLateInit  = "cannot late initialize GCP bucket"
	errCreate    = "cannot create GCP bucket"
	errUpdate    = "cannot update GCP bucket"
	errDelete    = "cannot delete GCP bucket"
)

// SetupBucket adds a controller that reconciles Buckets.
func SetupBucket(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha3.BucketGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha3.BucketGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha3.Bucket{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A BucketClient produces a BucketHandler for the named bucket.
type BucketClient interface {
	Bucket(name string) BucketHandler
}

// A GCSBucketClient wraps the GCS storage.Client as a BucketClient.
type GCSBucketClient struct {
	c *storage.Client
}

// Bucket produces a BucketHandler for the named bucket.
func (sbc *GCSBucketClient) Bucket(name string) BucketHandler {
	return sbc.c.Bucket(name)
}

// A BucketHandler handles requests to interact with buckets.
type BucketHandler interface {
	Attrs(context.Context) (*storage.BucketAttrs, error)
	Create(context.Context, string, *storage.BucketAttrs) error
	Update(context.Context, storage.BucketAttrsToUpdate) (*storage.BucketAttrs, error)
	Delete(context.Context) error
}

type connecter struct {
	client client.Client
}

// Connect sets up iam client using credentials from the provider
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetConnectionInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}

	s, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &external{handle: &GCSBucketClient{c: s}, projectID: projectID, client: c.client}, errors.Wrap(err, errNewClient)
}

type external struct {
	handle    BucketClient
	projectID string
	client    client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBucket)
	}

	a, err := e.handle.Bucket(meta.GetExternalName(cr)).Attrs(ctx)
	// NOTE(negz): The storage client appears to intercept the typical GCP API
	// error that we check for with gcp.IsErrorNotFound and return this error
	// instead, but only when getting bucket attributes.
	if err == storage.ErrBucketNotExist {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errAttrs)
	}

	proposed := cr.Spec.BucketSpecAttrs.DeepCopy()
	if err := mergo.Merge(proposed, v1alpha3.NewBucketSpecAttrs(a)); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errLateInit)
	}
	if !cmp.Equal(*proposed, cr.Spec.BucketSpecAttrs) {
		cr.Spec.BucketSpecAttrs = *proposed
		if err := e.client.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errLateInit)
		}
	}

	cr.Status.BucketOutputAttrs = v1alpha3.NewBucketOutputAttrs(a)
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: cmp.Equal(v1alpha3.NewBucketUpdatableAttrs(a), &cr.Spec.BucketUpdatableAttrs),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBucket)
	}

	err := e.handle.Bucket(meta.GetExternalName(cr)).Create(ctx, e.projectID, v1alpha3.CopyBucketSpecAttrs(&cr.Spec.BucketSpecAttrs))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBucket)
	}

	current, err := e.handle.Bucket(meta.GetExternalName(cr)).Attrs(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errAttrs)
	}
	ua := v1alpha3.CopyToBucketUpdateAttrs(cr.Spec.BucketUpdatableAttrs, current.Labels)
	_, err = e.handle.Bucket(meta.GetExternalName(cr)).Update(ctx, ua)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return errors.New(errNotBucket)
	}

	err := e.handle.Bucket(meta.GetExternalName(cr)).Delete(ctx)
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDelete)
}
