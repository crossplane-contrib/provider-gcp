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
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis"
	"github.com/crossplane/provider-gcp/apis/storage/v1alpha3"
)

func init() {
	_ = apis.AddToScheme(scheme.Scheme)
}

type MockBucketCreateUpdater struct {
	MockCreate func(context.Context) (reconcile.Result, error)
	MockUpdate func(context.Context, *storage.BucketAttrs) (reconcile.Result, error)
}

func (m *MockBucketCreateUpdater) create(ctx context.Context) (reconcile.Result, error) {
	return m.MockCreate(ctx)
}

func (m *MockBucketCreateUpdater) update(ctx context.Context, a *storage.BucketAttrs) (reconcile.Result, error) {
	return m.MockUpdate(ctx, a)
}

var _ createupdater = &MockBucketCreateUpdater{}

type MockBucketSyncDeleter struct {
	MockDelete func(context.Context) (reconcile.Result, error)
	MockSync   func(context.Context) (reconcile.Result, error)
}

func newMockBucketSyncDeleter() *MockBucketSyncDeleter {
	return &MockBucketSyncDeleter{
		MockSync: func(i context.Context) (result reconcile.Result, e error) {
			return requeueOnSuccess, nil
		},
		MockDelete: func(i context.Context) (result reconcile.Result, e error) {
			return result, nil
		},
	}
}

func (m *MockBucketSyncDeleter) delete(ctx context.Context) (reconcile.Result, error) {
	return m.MockDelete(ctx)
}

func (m *MockBucketSyncDeleter) sync(ctx context.Context) (reconcile.Result, error) {
	return m.MockSync(ctx)
}

var _ syncdeleter = &MockBucketSyncDeleter{}

type MockBucketFactory struct {
	MockNew func(context.Context, *v1alpha3.Bucket) (syncdeleter, error)
}

func newMockBucketFactory(rh syncdeleter, err error) *MockBucketFactory {
	return &MockBucketFactory{
		MockNew: func(i context.Context, bucket *v1alpha3.Bucket) (handler syncdeleter, e error) {
			return rh, err
		},
	}
}

func (m *MockBucketFactory) newSyncDeleter(ctx context.Context, b *v1alpha3.Bucket) (syncdeleter, error) {
	return m.MockNew(ctx, b)
}

type bucket struct {
	*v1alpha3.Bucket
}

func newBucket(name string) *bucket {
	b := &bucket{Bucket: &v1alpha3.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Finalizers: []string{},
		},
	}}
	meta.SetExternalName(b, name)
	return b
}

func (b *bucket) withServiceAccountSecretRef(namespace, name string) *bucket {
	b.Spec.ServiceAccountSecretRef = &runtimev1alpha1.SecretReference{Namespace: namespace, Name: name}
	return b
}

func (b *bucket) withWriteConnectionSecretToReference(namespace, name string) *bucket {
	b.Spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{Namespace: namespace, Name: name}
	return b
}

func (b *bucket) withDeleteTimestamp(t metav1.Time) *bucket {
	b.Bucket.ObjectMeta.DeletionTimestamp = &t
	return b
}

func (b *bucket) withFinalizer(f string) *bucket {
	b.Bucket.ObjectMeta.Finalizers = append(b.Bucket.ObjectMeta.Finalizers, f)
	return b
}

func (b *bucket) withConditions(c ...runtimev1alpha1.Condition) *bucket {
	b.Status.SetConditions(c...)
	return b
}

const (
	testBucketName = "testBucket"
)

func TestReconciler_Reconcile(t *testing.T) {
	name := testBucketName
	key := types.NamespacedName{Name: name}
	req := reconcile.Request{NamespacedName: key}
	ctx := context.TODO()
	rsDone := reconcile.Result{}

	type fields struct {
		client  client.Client
		factory factory
	}
	tests := []struct {
		name    string
		fields  fields
		wantRs  reconcile.Result
		wantErr error
		wantObj *v1alpha3.Bucket
	}{
		{
			name:    "GetErrNotFound",
			fields:  fields{fake.NewFakeClient(), nil},
			wantRs:  rsDone,
			wantErr: nil,
		},
		{
			name: "GetErrorOther",
			fields: fields{
				client: &test.MockClient{
					MockGet: func(context.Context, client.ObjectKey, runtime.Object) error {
						return errors.New("test-get-error")
					},
				},
				factory: nil},
			wantRs:  rsDone,
			wantErr: errors.New("test-get-error"),
		},
		{
			name: "BucketHandlerError",
			fields: fields{
				client:  fake.NewFakeClient(newBucket(name).withFinalizer("foo.bar").Bucket),
				factory: newMockBucketFactory(nil, errors.New("handler-factory-error")),
			},
			wantRs:  resultRequeue,
			wantErr: nil,
			wantObj: newBucket(name).
				withConditions(runtimev1alpha1.ReconcileError(errors.New("handler-factory-error"))).
				withFinalizer("foo.bar").Bucket,
		},
		{
			name: "ReconcileDelete",
			fields: fields{
				client: fake.NewFakeClient(newBucket(name).
					withDeleteTimestamp(metav1.NewTime(time.Now())).Bucket),
				factory: newMockBucketFactory(newMockBucketSyncDeleter(), nil),
			},
			wantRs:  rsDone,
			wantErr: nil,
		},
		{
			name: "ReconcileSync",
			fields: fields{
				client:  fake.NewFakeClient(newBucket(name).Bucket),
				factory: newMockBucketFactory(newMockBucketSyncDeleter(), nil),
			},
			wantRs:  requeueOnSuccess,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client:  tt.fields.client,
				factory: tt.fields.factory,

				log:         logging.NewNopLogger(),
				initializer: managed.NewNameAsExternalName(tt.fields.client),
			}
			got, err := r.Reconcile(req)
			if diff := cmp.Diff(tt.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("Reconciler.Reconcile() -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantRs, got); diff != "" {
				t.Errorf("Reconciler.Reconcile() -want result, +got result:\n%s", diff)
			}
			if tt.wantObj != nil {
				b := &v1alpha3.Bucket{}
				if err := r.Get(ctx, key, b); err != nil {
					t.Errorf("Reconciler.Reconcile() bucket error: %s", err)
				}
				// NOTE(muvaf): Get call fills TypeMeta and ObjectMeta.ResourceVersion
				// but these are not our concern in these tests.
				tt.wantObj.TypeMeta = b.TypeMeta
				tt.wantObj.ResourceVersion = b.ResourceVersion

				if diff := cmp.Diff(tt.wantObj, b, test.EquateConditions()); diff != "" {
					t.Errorf("Reconciler.Reconcile() -want bucket, +got bucket:\n%s", diff)
				}
			}
		})
	}
}

func Test_bucketSyncDeleter_delete(t *testing.T) {
	ctx := context.TODO()
	type fields struct {
		ops operations
	}
	type want struct {
		err error
		res reconcile.Result
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "RetainPolicy",
			fields: fields{
				ops: &mockOperations{
					mockIsReclaimDelete:     func() bool { return false },
					mockRemoveFinalizer:     func() {},
					mockUpdateObject:        func(ctx context.Context) error { return nil },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
				},
			},
			want: want{
				res: reconcile.Result{},
			},
		},
		{
			name: "DeleteSuccessful",
			fields: fields{
				ops: &mockOperations{
					mockIsReclaimDelete:     func() bool { return true },
					mockDeleteBucket:        func(ctx context.Context) error { return nil },
					mockRemoveFinalizer:     func() {},
					mockUpdateObject:        func(ctx context.Context) error { return nil },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
				},
			},
			want: want{
				err: nil,
				res: reconcile.Result{},
			},
		},
		{
			name: "DeleteFailedNotFound",
			fields: fields{
				ops: &mockOperations{
					mockIsReclaimDelete: func() bool { return true },
					mockDeleteBucket: func(ctx context.Context) error {
						return storage.ErrBucketNotExist
					},
					mockRemoveFinalizer:     func() {},
					mockUpdateObject:        func(ctx context.Context) error { return nil },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
			},
			want: want{
				err: nil,
				res: reconcile.Result{},
			},
		},
		{
			name: "DeleteFailedOther",
			fields: fields{
				ops: &mockOperations{
					mockIsReclaimDelete: func() bool { return true },
					mockDeleteBucket: func(ctx context.Context) error {
						return errors.New("test-error")
					},
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
			},
			want: want{
				res: resultRequeue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bsd := newBucketSyncDeleter(tt.fields.ops, "")
			got, err := bsd.delete(ctx)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("bucketSyncDeleter.delete() -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("bucketSyncDeleter.delete() -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_bucketSyncDeleter_sync(t *testing.T) {
	ctx := context.TODO()

	secretError := errors.New("test-update-secret-error")
	bucket404 := storage.ErrBucketNotExist
	getAttrsError := errors.New("test-get-attributes-error")

	type fields struct {
		ops operations
		cu  createupdater
	}
	type want struct {
		err error
		res reconcile.Result
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "FailedToUpdateConnectionSecret",
			fields: fields{
				ops: &mockOperations{
					mockUpdateSecret:        func(ctx context.Context) error { return secretError },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
			},
			want: want{res: resultRequeue},
		},
		{
			name: "AttrsErrorOther",
			fields: fields{
				ops: &mockOperations{
					mockUpdateSecret:        func(ctx context.Context) error { return nil },
					mockGetAttributes:       func(ctx context.Context) (*storage.BucketAttrs, error) { return nil, getAttrsError },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
			},
			want: want{res: resultRequeue},
		},
		{
			name: "CreateBucket",
			fields: fields{
				ops: &mockOperations{
					mockUpdateSecret: func(ctx context.Context) error { return nil },
					mockGetAttributes: func(ctx context.Context) (*storage.BucketAttrs, error) {
						return nil, bucket404
					},
				},
				cu: &MockBucketCreateUpdater{
					MockCreate: func(ctx context.Context) (reconcile.Result, error) {
						return reconcile.Result{}, nil
					},
				},
			},
			want: want{res: reconcile.Result{}},
		},
		{
			name: "UpdateBucket",
			fields: fields{
				ops: &mockOperations{
					mockUpdateSecret: func(ctx context.Context) error { return nil },
					mockGetAttributes: func(ctx context.Context) (*storage.BucketAttrs, error) {
						return &storage.BucketAttrs{}, bucket404
					},
				},
				cu: &MockBucketCreateUpdater{
					MockUpdate: func(ctx context.Context, attrs *storage.BucketAttrs) (reconcile.Result, error) {
						return requeueOnSuccess, nil
					},
				},
			},
			want: want{res: requeueOnSuccess},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketSyncDeleter{
				operations:    tt.fields.ops,
				createupdater: tt.fields.cu,
			}

			got, err := bh.sync(ctx)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("bucketSyncDeleter.sync() -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("bucketSyncDeleter.sync() -want, +got\n%s", diff)
				return
			}
		})
	}
}

func Test_bucketCreateUpdater_create(t *testing.T) {
	ctx := context.TODO()
	testError := errors.New("test-error")

	type fields struct {
		ops       operations
		projectID string
	}
	type want struct {
		err error
		res reconcile.Result
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "FailureToCreate",
			fields: fields{
				ops: &mockOperations{
					mockAddFinalizer:        func() {},
					mockCreateBucket:        func(ctx context.Context, projectID string) error { return testError },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
			},
			want: want{
				res: resultRequeue,
			},
		},
		{
			name: "FailureToGetAttributes",
			fields: fields{
				ops: &mockOperations{
					mockAddFinalizer:        func() {},
					mockCreateBucket:        func(ctx context.Context, projectID string) error { return nil },
					mockGetAttributes:       func(ctx context.Context) (*storage.BucketAttrs, error) { return nil, testError },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
			},
			want: want{
				res: resultRequeue,
			},
		},
		{
			name: "FailureToUpdateObject",
			fields: fields{
				ops: &mockOperations{
					mockAddFinalizer:        func() {},
					mockCreateBucket:        func(ctx context.Context, projectID string) error { return nil },
					mockGetAttributes:       func(ctx context.Context) (*storage.BucketAttrs, error) { return nil, nil },
					mockSetSpecAttrs:        func(attrs *storage.BucketAttrs) {},
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateObject:        func(ctx context.Context) error { return testError },
				},
			},
			want: want{
				err: testError,
				res: resultRequeue,
			},
		},
		{
			name: "Success",
			fields: fields{
				ops: &mockOperations{
					mockAddFinalizer:        func() {},
					mockCreateBucket:        func(ctx context.Context, projectID string) error { return nil },
					mockGetAttributes:       func(ctx context.Context) (*storage.BucketAttrs, error) { return nil, nil },
					mockSetSpecAttrs:        func(attrs *storage.BucketAttrs) {},
					mockUpdateObject:        func(ctx context.Context) error { return nil },
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockSetBindable:         func() {},
					mockSetStatusAttrs:      func(attrs *storage.BucketAttrs) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
			},
			want: want{
				res: requeueOnSuccess,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketCreateUpdater{
				operations: tt.fields.ops,
				projectID:  tt.fields.projectID,
			}
			got, err := bh.create(ctx)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("bucketCreateUpdater.create() -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("bucketCreateUpdater.create() -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_bucketCreateUpdater_update(t *testing.T) {
	ctx := context.TODO()
	testError := errors.New("test-error")

	type fields struct {
		ops       operations
		projectID string
	}
	type want struct {
		err error
		res reconcile.Result
	}
	tests := []struct {
		name   string
		fields fields
		args   *storage.BucketAttrs
		want   want
	}{
		{
			name: "NoChanges",
			fields: fields{
				ops: &mockOperations{
					mockGetSpecAttrs: func() v1alpha3.BucketUpdatableAttrs {
						return v1alpha3.BucketUpdatableAttrs{}
					},
				},
				projectID: "",
			},
			args: &storage.BucketAttrs{},
			want: want{res: requeueOnSuccess},
		},
		{
			name: "FailureToUpdateBucket",
			fields: fields{
				ops: &mockOperations{
					mockGetSpecAttrs: func() v1alpha3.BucketUpdatableAttrs {
						return v1alpha3.BucketUpdatableAttrs{RequesterPays: true}
					},
					mockUpdateBucket: func(ctx context.Context, labels map[string]string) (*storage.BucketAttrs, error) {
						return nil, testError
					},
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
				projectID: "",
			},
			args: &storage.BucketAttrs{},
			want: want{res: resultRequeue},
		},
		{
			name: "FailureToUpdateObject",
			fields: fields{
				ops: &mockOperations{
					mockGetSpecAttrs: func() v1alpha3.BucketUpdatableAttrs {
						return v1alpha3.BucketUpdatableAttrs{RequesterPays: true}
					},
					mockUpdateBucket: func(ctx context.Context, labels map[string]string) (*storage.BucketAttrs, error) {
						return nil, nil
					},
					mockSetSpecAttrs:        func(attrs *storage.BucketAttrs) {},
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateObject:        func(ctx context.Context) error { return testError },
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
				projectID: "",
			},
			args: &storage.BucketAttrs{},
			want: want{
				err: testError,
				res: resultRequeue,
			},
		},
		{
			name: "Successful",
			fields: fields{
				ops: &mockOperations{
					mockGetSpecAttrs: func() v1alpha3.BucketUpdatableAttrs {
						return v1alpha3.BucketUpdatableAttrs{RequesterPays: true}
					},
					mockUpdateBucket: func(ctx context.Context, labels map[string]string) (*storage.BucketAttrs, error) {
						return nil, nil
					},
					mockSetSpecAttrs:        func(attrs *storage.BucketAttrs) {},
					mockSetStatusConditions: func(_ ...runtimev1alpha1.Condition) {},
					mockUpdateObject:        func(ctx context.Context) error { return nil },
					mockUpdateStatus:        func(ctx context.Context) error { return nil },
				},
				projectID: "",
			},
			args: &storage.BucketAttrs{},
			want: want{
				res: requeueOnSuccess,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &bucketCreateUpdater{
				operations: tt.fields.ops,
				projectID:  tt.fields.projectID,
			}
			got, err := bh.update(ctx, tt.args)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("bucketCreateUpdater.update() -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("bucketCreateUpdater.update() -want, +got:\n%s", diff)
			}
		})
	}
}
