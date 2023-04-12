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

package registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"google.golang.org/api/storage/v1"
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

	"github.com/crossplane-contrib/provider-gcp/apis/registry/v1alpha1"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
)

// Error strings.
const (
	errNewStorageClient = "cannot create new Google Storage Client"
	errNotGcr           = "managed resource is not a Google Container Registry (GCR)"
	errHandshake        = "an error occurred during handshake"
	errGetBucket        = "cannot get Bucket object"

	gcrURL                          = "gcr.io"
	testRepoName                    = "crossplane-test"
	bucketNameFormatWithLocation    = "%s.artifacts.%s.appspot.com"
	bucketNameFormatWithoutLocation = "artifacts.%s.appspot.com"
)

// SetupContainerRegistry adds a controller that reconciles ContainerRegistries.
func SetupContainerRegistry(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ContainerRegistryGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind, connection.WithTLSConfig(o.ESSOptions.TLSConfig)))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ContainerRegistryGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.ContainerRegistry{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connecter struct {
	client client.Client
}

// Connect sets up iam client using credentials from the provider
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, _, err := gcp.GetConnectionInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}

	storageService, err := storage.NewService(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errNewStorageClient)
	}

	return &external{client: c.client, storage: storageService, projectID: projectID}, nil
}

type external struct {
	client    client.Client
	storage   *storage.Service
	projectID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ContainerRegistry)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGcr)
	}

	bucket, err := e.getBucket(cr)
	if gcp.IsErrorNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetBucket)
	}

	// We skip deletion. This means that when you delete a Container Registry resource, the created Bucket will not be
	// deleted. So, after the deletion starts, if observation output still sets the ResourceExists field to true, we
	// will see that the CR could not be deleted. Therefore, it was necessary to make a deletionTimestamp check here.
	if meta.WasDeleted(cr) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.Status.AtProvider.BucketID = bucket.Id
	cr.Status.AtProvider.BucketLink = bucket.SelfLink
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ContainerRegistry)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGcr)
	}

	location := gcrURL
	if cr.Spec.ForProvider.Location != "" {
		location = fmt.Sprintf("%s.%s", strings.ToLower(cr.Spec.ForProvider.Location), location)
	}

	ref, err := name.ParseReference(fmt.Sprintf("%s/%s/%s", location, e.projectID, testRepoName))
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	auth, err := authn.DefaultKeychain.Resolve(ref.Context().Registry)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	// A Bucket is required to properly run image actions (e.g. push, pull) in the Registry. So when a ContainerRegistry MR
	// created, an external Bucket resource will be created in GCP side.
	// In order to trigger a Bucket creation mentioned above, it is sufficient to perform a handshake with the gcr.io service.
	// The following NewWithContext function performs a handshake operation.
	if _, err := transport.NewWithContext(ctx, ref.Context().Registry, auth, http.DefaultTransport,
		[]string{ref.Scope(transport.PushScope)}); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errHandshake)
	}

	return managed.ExternalCreation{}, nil
}

// Update function is skipped because, the ContainerRegistry resource only ensures that, a Bucket was created in the
// specified location.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Delete function is skipped because, deleting the created Bucket can cause data loss.
// The main aim is that prevent data loss in case this Bucket was used for other purposes.
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil
}

func (e *external) getBucket(cr *v1alpha1.ContainerRegistry) (*storage.Bucket, error) {
	bucketName := ""
	if cr.Spec.ForProvider.Location != "" {
		bucketName = fmt.Sprintf(bucketNameFormatWithLocation, strings.ToLower(cr.Spec.ForProvider.Location), e.projectID)
	} else {
		bucketName = fmt.Sprintf(bucketNameFormatWithoutLocation, e.projectID)
	}

	bucket, err := e.storage.Buckets.Get(bucketName).Do()
	if err != nil {
		return bucket, err
	}

	return bucket, nil
}
