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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	gcpstorage "github.com/crossplaneio/stack-gcp/pkg/clients/storage"
)

type operations interface {
	// Bucket object operations
	addFinalizer()
	removeFinalizer()
	isReclaimDelete() bool
	getSpecAttrs() v1alpha3.BucketUpdatableAttrs
	setSpecAttrs(*storage.BucketAttrs)
	setStatusAttrs(*storage.BucketAttrs)
	setStatusConditions(c ...runtimev1alpha1.Condition)
	setBindable()

	// Controller-runtime operations
	updateObject(ctx context.Context) error
	updateStatus(ctx context.Context) error
	updateSecret(ctx context.Context) error

	// GCP Storage Client operations
	createBucket(ctx context.Context, projectID string) error
	deleteBucket(ctx context.Context) error
	updateBucket(ctx context.Context, labels map[string]string) (*storage.BucketAttrs, error)
	getAttributes(ctx context.Context) (*storage.BucketAttrs, error)
}

type bucketHandler struct {
	*v1alpha3.Bucket
	kube client.Client
	gcp  gcpstorage.Client
}

var _ operations = &bucketHandler{}

func newBucketClients(bucket *v1alpha3.Bucket, kube client.Client, gcp gcpstorage.Client) *bucketHandler {
	return &bucketHandler{
		Bucket: bucket,
		kube:   kube,
		gcp:    gcp,
	}
}

//
// Crossplane GCP Bucket object operations
//
func (bh *bucketHandler) addFinalizer() {
	meta.AddFinalizer(bh, finalizer)
}

func (bh *bucketHandler) removeFinalizer() {
	meta.RemoveFinalizer(bh, finalizer)
}

func (bh *bucketHandler) isReclaimDelete() bool {
	return bh.Spec.ReclaimPolicy == runtimev1alpha1.ReclaimDelete
}

func (bh *bucketHandler) getSpecAttrs() v1alpha3.BucketUpdatableAttrs {
	return bh.Spec.BucketUpdatableAttrs
}

func (bh *bucketHandler) setSpecAttrs(attrs *storage.BucketAttrs) {
	bh.Spec.BucketSpecAttrs = v1alpha3.NewBucketSpecAttrs(attrs)
}

func (bh *bucketHandler) setStatusAttrs(attrs *storage.BucketAttrs) {
	bh.Status.BucketOutputAttrs = v1alpha3.NewBucketOutputAttrs(attrs)
}

func (bh *bucketHandler) setStatusConditions(c ...runtimev1alpha1.Condition) {
	bh.Status.SetConditions(c...)
}

func (bh *bucketHandler) setBindable() {
	resource.SetBindable(bh)
}

//
// Controller-runtime Client operations
//
func (bh *bucketHandler) updateObject(ctx context.Context) error {
	return bh.kube.Update(ctx, bh.Bucket)
}

func (bh *bucketHandler) updateStatus(ctx context.Context) error {
	return bh.kube.Status().Update(ctx, bh.Bucket)
}

func (bh *bucketHandler) getSecret(ctx context.Context, nn types.NamespacedName, s *corev1.Secret) error {
	return bh.kube.Get(ctx, nn, s)
}

const (
	saSecretKeyAccessKey   = "interopAccessKey"
	saSecretKeySecret      = "interopSecret"
	saSecretKeyCredentials = "credentials.json"
)

func (bh *bucketHandler) updateSecret(ctx context.Context) error {

	if bh.Spec.WriteConnectionSecretToReference == nil {
		// No connection secret is desired - don't publish one.
		return nil
	}

	s := resource.ConnectionSecretFor(bh.Bucket, v1alpha3.BucketGroupVersionKind)
	if ref := bh.Spec.ServiceAccountSecretRef; ref != nil {
		ss := &corev1.Secret{}
		nn := types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}
		if err := bh.kube.Get(ctx, nn, ss); err != nil {
			return errors.Wrapf(err, "failed to retrieve storage service account secret: %s", nn)
		}
		s.Data[runtimev1alpha1.ResourceCredentialsSecretUserKey] = ss.Data[saSecretKeyAccessKey]
		s.Data[runtimev1alpha1.ResourceCredentialsSecretPasswordKey] = ss.Data[saSecretKeySecret]
		s.Data[runtimev1alpha1.ResourceCredentialsSecretTokenKey] = ss.Data[saSecretKeyCredentials]
	}
	s.Data[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(bh.GetBucketName())

	return errors.Wrapf(apply(ctx, bh.kube, s), "failed to apply connection secret: %s/%s", s.Namespace, s.Name)
}

func apply(ctx context.Context, kube client.Client, o runtime.Object) error {
	err := kube.Create(ctx, o)
	if err != nil && kerrors.IsAlreadyExists(err) {
		return kube.Update(ctx, o)
	}
	return err
}

//
// GCP Storage Bucket operations
//
func (bh *bucketHandler) createBucket(ctx context.Context, projectID string) error {
	return bh.gcp.Create(ctx, projectID, v1alpha3.CopyBucketSpecAttrs(&bh.Spec.BucketSpecAttrs))
}

func (bh *bucketHandler) deleteBucket(ctx context.Context) error {
	return bh.gcp.Delete(ctx)
}

func (bh *bucketHandler) updateBucket(ctx context.Context, labels map[string]string) (*storage.BucketAttrs, error) {
	return bh.gcp.Update(ctx, v1alpha3.CopyToBucketUpdateAttrs(bh.Spec.BucketUpdatableAttrs, labels))
}

func (bh *bucketHandler) getAttributes(ctx context.Context) (*storage.BucketAttrs, error) {
	return bh.gcp.Attrs(ctx)
}
