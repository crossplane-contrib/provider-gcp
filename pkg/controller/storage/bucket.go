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

package storage

import (
	"context"
	"fmt"
	"strings"

	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"

	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	apisv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/crossplaneio/stack-gcp/pkg/clients/bucket"
)

const (
	errNotBucket                  = "managed resource is not a Bucket custom resource"
	errProviderNotRetrieved       = "provider could not be retrieved"
	errProviderSecretNotRetrieved = "secret referred in provider could not be retrieved"

	errNewClient    = "cannot create a new Storage Service"
	errCreateFailed = "cannot create a new bucket"
)

// BucketController is the controller for Bucket CRD.
type BucketController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields
// on the Controller and Start it when the Manager is Started.
func (c *BucketController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1alpha3.BucketGroupVersionKind),
		resource.WithExternalConnecter(&bucketController{kube: mgr.GetClient(), newServiceFn: storage.NewService}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha3.BucketKindAPIVersion, v1alpha3.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.Bucket{}).
		Complete(r)
}

type bucketController struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*storage.Service, error)
}

func (c *bucketController) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return nil, errors.New(errNotBucket)
	}

	provider := &apisv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), provider); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}
	secret := &v1.Secret{}
	n := types.NamespacedName{Namespace: provider.Spec.CredentialsSecretRef.Namespace, Name: provider.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, secret); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	s, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(secret.Data[provider.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(storage.CloudPlatformScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &bucketExternal{kube: c.kube, bucket: s.Buckets, projectID: provider.Spec.ProjectID}, nil
}

type bucketExternal struct {
	kube      client.Client
	bucket    *storage.BucketsService
	projectID string
}

func (b *bucketExternal) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotBucket)
	}
	instance, err := b.bucket.Get(meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, "cannot get bucket")
	}
	cr.Status.AtProvider = bucket.GenerateObservation(*instance)
	// todo: Late init
	cr.Status.SetConditions(v1alpha1.Available())
	resource.SetBindable(cr)
	return resource.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (b *bucketExternal) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotBucket)
	}
	cr.SetConditions(v1alpha1.Creating())
	instance := bucket.GenerateBucket(cr.Spec.ForProvider, meta.GetExternalName(cr))
	if _, err := b.bucket.Insert(b.projectID, instance).Context(ctx).Do(); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(resource.Ignore(gcp.IsErrorAlreadyExists, err), errCreateFailed)
	}
	return resource.ExternalCreation{}, nil
}

func (b *bucketExternal) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotBucket)
	}
	instance := bucket.GenerateBucket(cr.Spec.ForProvider, meta.GetExternalName(cr))
	if _, err := b.bucket.Patch(b.projectID, instance).Context(ctx).Do(); err != nil {
		return resource.ExternalUpdate{}, errors.Wrap(err, "cannot patch bucket")
	}
	return resource.ExternalUpdate{}, nil
}

func (b *bucketExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha3.Bucket)
	if !ok {
		return errors.New(errNotBucket)
	}
	cr.SetConditions(v1alpha1.Deleting())
	if err := b.bucket.Delete(meta.GetExternalName(cr)).Context(ctx).Do(); resource.Ignore(gcp.IsErrorNotFound, err) != nil {
		return errors.Wrap(err, "cannot delete bucket")
	}
	return nil
}
