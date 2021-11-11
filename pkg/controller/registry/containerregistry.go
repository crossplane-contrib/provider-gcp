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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"google.golang.org/api/storage/v1"

	"github.com/google/go-containerregistry/pkg/authn"

	"time"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/provider-gcp/apis/registry/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
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
func SetupContainerRegistry(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.ContainerRegistryKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.ContainerRegistry{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ContainerRegistryGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

// Connect sets up iam client using credentials from the provider
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, _, err := gcp.GetAuthInfo(ctx, c.client, mg)

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

	bucket, err := e.getBucket(e.getBucketName(cr))

	if gcp.IsErrorNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetBucket)
	}

	populateCRFromBucket(cr, bucket)

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

	if _, err := transport.NewWithContext(ctx, ref.Context().Registry, auth, http.DefaultTransport,
		[]string{ref.Scope(transport.PushScope)}); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errHandshake)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil
}

func populateCRFromBucket(cr *v1alpha1.ContainerRegistry, bucket *storage.Bucket) {
	cr.Status.AtProvider.ID = bucket.Id
	cr.Status.AtProvider.BucketLink = bucket.SelfLink
}

func (e *external) getBucketName(cr *v1alpha1.ContainerRegistry) string {
	bucketName := ""

	if cr.Spec.ForProvider.Location != "" {
		bucketName = fmt.Sprintf(bucketNameFormatWithLocation, strings.ToLower(cr.Spec.ForProvider.Location), e.projectID)
	} else {
		bucketName = fmt.Sprintf(bucketNameFormatWithoutLocation, e.projectID)
	}
	return bucketName
}

func (e *external) getBucket(bucketName string) (*storage.Bucket, error) {
	bucket, err := e.storage.Buckets.Get(bucketName).Do()

	if err != nil {
		return bucket, err
	}

	return bucket, nil
}
