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
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	gcpstorage "github.com/crossplaneio/stack-gcp/pkg/clients/storage"
	storagefake "github.com/crossplaneio/stack-gcp/pkg/clients/storage/fake"
)

type mockOperations struct {
	mockIsReclaimDelete     func() bool
	mockAddFinalizer        func()
	mockRemoveFinalizer     func()
	mockGetSpecAttrs        func() v1alpha3.BucketUpdatableAttrs
	mockSetSpecAttrs        func(*storage.BucketAttrs)
	mockSetStatusAttrs      func(*storage.BucketAttrs)
	mockSetStatusConditions func(...runtimev1alpha1.Condition)
	mockSetBindable         func()

	mockUpdateObject func(ctx context.Context) error
	mockUpdateStatus func(ctx context.Context) error
	mockUpdateSecret func(ctx context.Context) error

	mockCreateBucket  func(ctx context.Context, projectID string) error
	mockDeleteBucket  func(ctx context.Context) error
	mockUpdateBucket  func(ctx context.Context, labels map[string]string) (*storage.BucketAttrs, error)
	mockGetAttributes func(ctx context.Context) (*storage.BucketAttrs, error)
}

var _ operations = &mockOperations{}

func (o *mockOperations) isReclaimDelete() bool {
	return o.mockIsReclaimDelete()
}

func (o *mockOperations) addFinalizer() {
	o.mockAddFinalizer()
}

func (o *mockOperations) removeFinalizer() {
	o.mockRemoveFinalizer()
}

func (o *mockOperations) getSpecAttrs() v1alpha3.BucketUpdatableAttrs {
	return o.mockGetSpecAttrs()
}

func (o *mockOperations) setSpecAttrs(attrs *storage.BucketAttrs) {
	o.mockSetSpecAttrs(attrs)
}

func (o *mockOperations) setStatusAttrs(attrs *storage.BucketAttrs) {
	o.mockSetStatusAttrs(attrs)
}

func (o *mockOperations) setStatusConditions(c ...runtimev1alpha1.Condition) {
	o.mockSetStatusConditions(c...)
}

func (o *mockOperations) setBindable() {
	o.mockSetBindable()
}

//
//
func (o *mockOperations) updateObject(ctx context.Context) error {
	return o.mockUpdateObject(ctx)
}

func (o *mockOperations) updateStatus(ctx context.Context) error {
	return o.mockUpdateStatus(ctx)
}

func (o *mockOperations) updateSecret(ctx context.Context) error {
	return o.mockUpdateSecret(ctx)
}

//
//
func (o *mockOperations) createBucket(ctx context.Context, projectID string) error {
	return o.mockCreateBucket(ctx, projectID)
}

func (o *mockOperations) deleteBucket(ctx context.Context) error {
	return o.mockDeleteBucket(ctx)
}

func (o *mockOperations) updateBucket(ctx context.Context, labels map[string]string) (*storage.BucketAttrs, error) {
	return o.mockUpdateBucket(ctx, labels)
}

func (o *mockOperations) getAttributes(ctx context.Context) (*storage.BucketAttrs, error) {
	return o.mockGetAttributes(ctx)
}

//
//
func Test_bucketHandler_addFinalizer(t *testing.T) {
	type fields struct {
		bucket *v1alpha3.Bucket
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "Test",
			fields: fields{bucket: &v1alpha3.Bucket{}},
			want:   []string{finalizer},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &bucketHandler{
				Bucket: tt.fields.bucket,
			}
			bc.addFinalizer()
			got := tt.fields.bucket.Finalizers
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("bucketHandler.addFinalizer(): -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_bucketHandler_removeFinalizer(t *testing.T) {
	type fields struct {
		bucket *v1alpha3.Bucket
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Test",
			fields: fields{bucket: &v1alpha3.Bucket{
				ObjectMeta: metav1.ObjectMeta{Finalizers: []string{finalizer}},
			}},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &bucketHandler{
				Bucket: tt.fields.bucket,
			}
			bc.removeFinalizer()
			got := tt.fields.bucket.Finalizers
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("bucketHandler.removeFinalizer(): -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_bucketHandler_isReclaimDelete(t *testing.T) {
	type fields struct {
		bucket *v1alpha3.Bucket
		kube   client.Client
		gcp    gcpstorage.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Default",
			fields: fields{bucket: &v1alpha3.Bucket{}},
			want:   false,
		},
		{
			name: "Delete",
			fields: fields{bucket: &v1alpha3.Bucket{
				Spec: v1alpha3.BucketSpec{
					ResourceSpec: runtimev1alpha1.ResourceSpec{
						ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
					},
				},
			},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &bucketHandler{
				Bucket: tt.fields.bucket,
				kube:   tt.fields.kube,
				gcp:    tt.fields.gcp,
			}
			if got := bc.isReclaimDelete(); got != tt.want {
				t.Errorf("bucketHandler.isReclaimDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bucketHandler_getSpecAttrs(t *testing.T) {
	testBucketSpecAttrs := v1alpha3.BucketUpdatableAttrs{RequesterPays: true}
	type fields struct {
		bucket *v1alpha3.Bucket
	}
	tests := []struct {
		name   string
		fields fields
		want   v1alpha3.BucketUpdatableAttrs
	}{
		{
			name: "Test",
			fields: fields{bucket: &v1alpha3.Bucket{
				Spec: v1alpha3.BucketSpec{
					BucketParameters: v1alpha3.BucketParameters{
						BucketSpecAttrs: v1alpha3.BucketSpecAttrs{BucketUpdatableAttrs: testBucketSpecAttrs},
					},
				},
			}},
			want: testBucketSpecAttrs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketHandler{
				Bucket: tt.fields.bucket,
			}
			got := bh.getSpecAttrs()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("bucketHandler.getSpecAttrs(): -want, +got\n%s", diff)
			}
		})
	}
}

func Test_bucketHandler_setSpecAttrs(t *testing.T) {
	testSpecAttrs := v1alpha3.BucketSpecAttrs{Location: "foo"}
	type fields struct {
		bucket *v1alpha3.Bucket
	}
	tests := []struct {
		name   string
		fields fields
		args   *storage.BucketAttrs
		want   v1alpha3.BucketSpecAttrs
	}{
		{
			name:   "Test",
			fields: fields{bucket: &v1alpha3.Bucket{}},
			args:   &storage.BucketAttrs{Location: "foo"},
			want:   testSpecAttrs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketHandler{
				Bucket: tt.fields.bucket,
			}
			bh.setSpecAttrs(tt.args)
			got := tt.fields.bucket.Spec.BucketSpecAttrs
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("bucketHandler.setSpecAttrs(): -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_bucketHandler_setStatusAttrs(t *testing.T) {
	type fields struct {
		bucket *v1alpha3.Bucket
	}
	tests := []struct {
		name   string
		fields fields
		args   *storage.BucketAttrs
		want   v1alpha3.BucketOutputAttrs
	}{
		{
			name:   "Test",
			fields: fields{bucket: &v1alpha3.Bucket{}},
			args:   &storage.BucketAttrs{Name: "foo"},
			want:   v1alpha3.BucketOutputAttrs{Name: "foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketHandler{
				Bucket: tt.fields.bucket,
			}
			bh.setStatusAttrs(tt.args)
			got := tt.fields.bucket.Status.BucketOutputAttrs
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("bucketHandler.setStatusAttrs(): -want, +got\n%s", diff)
			}
		})
	}
}

func Test_bucketHandler_updateObject(t *testing.T) {
	ctx := context.TODO()
	bucket := &v1alpha3.Bucket{}
	bc := &bucketHandler{
		Bucket: bucket,
		kube: &test.MockClient{
			MockUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
				if _, ok := obj.(*v1alpha3.Bucket); !ok {
					t.Errorf("bucketHandler.updateObject() unexpected type %T, want %T", obj, bucket)
				}
				return nil
			},
		},
	}
	if err := bc.updateObject(ctx); err != nil {
		t.Errorf("bucketHandler.updateObject() unexpected error %v", err)
	}
}

func Test_bucketHandler_updateStatus(t *testing.T) {
	ctx := context.TODO()
	bucket := &v1alpha3.Bucket{}
	bc := &bucketHandler{
		Bucket: bucket,
		kube: &test.MockClient{
			MockStatusUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
				if _, ok := obj.(*v1alpha3.Bucket); !ok {
					t.Errorf("bucketHandler.updateStatus() unexpected type %T, want %T", obj, bucket)
				}
				return nil
			},
		},
	}
	if err := bc.updateStatus(ctx); err != nil {
		t.Errorf("bucketHandler.updateStatus() unexpected error %v", err)
	}
}

func Test_bucketHandler_getSecret(t *testing.T) {
	ctx := context.TODO()
	nn := types.NamespacedName{Namespace: "foo", Name: "bar"}
	s := &corev1.Secret{}
	bc := &bucketHandler{
		kube: &test.MockClient{
			MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
				if key != nn {
					t.Errorf("bucketHandler.getSecret() key = %v, want %v", key, nn)
				}
				if _, ok := obj.(*corev1.Secret); !ok {
					t.Errorf("bucketHandler.getSecret() type = %T, want %T", obj, &corev1.Secret{})
				}
				return nil
			},
		},
	}
	if err := bc.getSecret(ctx, nn, s); err != nil {
		t.Errorf("bucketHandler.getSecret() unexpected error %v", err)
	}
}

func Test_bucketHandler_updateSecret(t *testing.T) {
	ctx := context.TODO()
	testError := errors.New("test-error")
	testNamespace := "test-sa-namespace"
	bucketUID := "test-uid"
	saSecretName := "test-sa-secret"
	saSecretUser := "test-user"
	saSecretPass := "test-pass"
	saSecretCreds := "test-creds"

	assertSecretData := func(data map[string][]byte, key, want string) {
		if v := data[key]; string(v) != want {
			t.Errorf("bucketHandler.updateSecret() data = %v, want %v", v, want)
		}
	}

	type fields struct {
		Bucket *v1alpha3.Bucket
		kube   client.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "WithoutServiceAccountSecretReference",
			fields: fields{
				Bucket: newBucket(testBucketName).Bucket,
				kube:   test.NewMockClient(),
			},
		},
		{
			name: "FailureToRetrieveSecret",
			fields: fields{
				Bucket: newBucket(testBucketName).
					withWriteConnectionSecretToReference(testNamespace, testBucketName).
					withServiceAccountSecretRef(testNamespace, saSecretName).Bucket,
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
						return testError
					},
				},
			},
			want: errors.Wrapf(testError,
				"failed to retrieve storage service account secret: %s/%s", testNamespace, saSecretName),
		},
		{
			name: "FailureToUpdateSecret",
			fields: fields{
				Bucket: newBucket(testBucketName).
					withWriteConnectionSecretToReference(testNamespace, testBucketName).
					withServiceAccountSecretRef(testNamespace, saSecretName).
					withUID(bucketUID).
					Bucket,
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
						s, ok := obj.(*corev1.Secret)
						if !ok {
							t.Errorf("bucketHandler.updateSecret() invalid type = %T, want %T",
								obj, &corev1.Secret{})
						}
						s.Name = saSecretName
						s.Data = map[string][]byte{
							saSecretKeyAccessKey:   []byte(saSecretUser),
							saSecretKeySecret:      []byte(saSecretPass),
							saSecretKeyCredentials: []byte(saSecretCreds),
						}
						return nil
					},
					MockCreate: func(ctx context.Context, obj runtime.Object, _ ...client.CreateOption) error {
						// assert secret
						s, ok := obj.(*corev1.Secret)
						if !ok {
							t.Errorf("bucketHandler.updateSecret() invalid type = %T, want %T",
								obj, &corev1.Secret{})
						}
						// assert secret data
						assertSecretData(s.Data, runtimev1alpha1.ResourceCredentialsSecretEndpointKey, bucketUID)
						assertSecretData(s.Data, runtimev1alpha1.ResourceCredentialsSecretUserKey, saSecretUser)
						assertSecretData(s.Data, runtimev1alpha1.ResourceCredentialsSecretPasswordKey, saSecretPass)
						assertSecretData(s.Data, runtimev1alpha1.ResourceCredentialsSecretTokenKey, saSecretCreds)
						return testError
					},
				},
			},
			want: errors.Wrapf(testError,
				"failed to apply connection secret: %s/%s", testNamespace, testBucketName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketHandler{
				Bucket: tt.fields.Bucket,
				kube:   tt.fields.kube,
			}
			err := bh.updateSecret(ctx)
			if diff := cmp.Diff(tt.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("bucketHandler.updateSecret(): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func Test_bucketHandler_createBucket(t *testing.T) {
	ctx := context.TODO()
	testProjectID := "foo"
	actualProjectID := "bar"
	bc := &bucketHandler{
		Bucket: &v1alpha3.Bucket{},
		gcp: &storagefake.MockBucketClient{
			MockCreate: func(ctx context.Context, s string, attrs *storage.BucketAttrs) error {
				actualProjectID = s
				return nil
			},
		},
	}
	if err := bc.createBucket(ctx, testProjectID); err != nil {
		t.Errorf("bucketHandler.createBucket() unexpected error %v", err)
	}
	if actualProjectID != testProjectID {
		t.Errorf("bucketHandler.createBucket() projectID = %s, want %v", actualProjectID, testProjectID)
	}
}

func Test_bucketHandler_deleteBucket(t *testing.T) {
	ctx := context.TODO()
	bc := &bucketHandler{
		Bucket: &v1alpha3.Bucket{},
		gcp: &storagefake.MockBucketClient{
			MockDelete: func(ctx context.Context) error { return nil },
		},
	}
	if err := bc.deleteBucket(ctx); err != nil {
		t.Errorf("bucketHandler.deleteBucket() unexpected error %v", err)
	}
}

func Test_bucketHandler_updateBucket(t *testing.T) {
	ctx := context.TODO()
	bc := &bucketHandler{
		Bucket: &v1alpha3.Bucket{},
		gcp: &storagefake.MockBucketClient{
			MockUpdate: func(ctx context.Context, update storage.BucketAttrsToUpdate) (attrs *storage.BucketAttrs, e error) {
				return &storage.BucketAttrs{}, nil
			},
		},
	}
	labels := map[string]string{"Foo": "bar"}
	want := &storage.BucketAttrs{}
	got, err := bc.updateBucket(ctx, labels)
	if err != nil {
		t.Errorf("bucketHandler.updateBucket() unexpected error %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("bucketHandler.updateBucket(): -want, +got:\n%s", diff)
	}
}

func Test_bucketHandler_getAttributes(t *testing.T) {
	ctx := context.TODO()
	type fields struct {
		gcp gcpstorage.Client
	}
	type want struct {
		err   error
		attrs *storage.BucketAttrs
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Test",
			fields: fields{
				gcp: &storagefake.MockBucketClient{
					MockAttrs: func(ctx context.Context) (*storage.BucketAttrs, error) { return nil, nil },
				},
			},
			want: want{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketHandler{
				gcp: tt.fields.gcp,
			}
			got, err := bh.getAttributes(ctx)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("bucketHandler.getAttributes(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.attrs, got); diff != "" {
				t.Errorf("bucketHandler.getAttributes(): -want, +got:\n%s", diff)
			}
		})
	}
}
