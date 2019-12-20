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

	"github.com/crossplaneio/stack-gcp/apis"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
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

type mockReferenceResolver struct{}

func (*mockReferenceResolver) ResolveReferences(ctx context.Context, res resource.CanReference) (err error) {
	return nil
}

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
	return &bucket{Bucket: &v1alpha3.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Finalizers: []string{},
		},
	}}
}

func (b *bucket) withUID(uid string) *bucket {
	b.ObjectMeta.UID = types.UID(uid)
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

func (b *bucket) withProvider(name string) *bucket {
	b.Spec.ProviderReference = &corev1.ObjectReference{Name: name}
	return b
}

func (b *bucket) withConditions(c ...runtimev1alpha1.Condition) *bucket {
	b.Status.SetConditions(c...)
	return b
}

type provider struct {
	*gcpv1alpha3.Provider
}

func newProvider(name string) *provider {
	return &provider{Provider: &gcpv1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

func (p *provider) withSecret(namespace, name, key string) *provider {
	p.Spec.CredentialsSecretRef = runtimev1alpha1.SecretKeySelector{
		SecretReference: runtimev1alpha1.SecretReference{
			Namespace: namespace,
			Name:      name,
		},
		Key: key,
	}
	return p
}

type secret struct {
	*corev1.Secret
}

func newSecret(ns, name string) *secret {
	return &secret{
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		},
	}
}

func (s *secret) withKeyData(key, data string) *secret {
	if s.Data == nil {
		s.Data = make(map[string][]byte)
	}
	s.Data[key] = []byte(data)
	return s
}

const (
	testNamespace  = "default"
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
				withConditions(runtimev1alpha1.ReferenceResolutionSuccess(), runtimev1alpha1.ReconcileError(errors.New("handler-factory-error"))).
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

				ManagedReferenceResolver: &mockReferenceResolver{},
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

func Test_bucketFactory_newHandler(t *testing.T) {
	ctx := context.TODO()
	ns := testNamespace
	bucketName := testBucketName
	providerName := "test-provider"
	secretName := "test-secret"
	secretKey := "creds"
	secretData := `{
	"type": "service_account",
	"project_id": "%s",
	"private_key_id": "%s",
	"private_key": "-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----\n",
	"client_email": "%s",
	"client_id": "%s",
	"auth_uri": "https://accounts.google.com/bucket/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "%s"}`
	type want struct {
		err error
		sd  syncdeleter
	}
	tests := []struct {
		name   string
		Client client.Client
		bucket *v1alpha3.Bucket
		want   want
	}{
		{
			name:   "ErrProviderIsNotFound",
			Client: fake.NewFakeClient(),
			bucket: newBucket(bucketName).withProvider(providerName).Bucket,
			want: want{
				err: kerrors.NewNotFound(schema.GroupResource{
					Group:    gcpv1alpha3.Group,
					Resource: "providers"}, "test-provider"),
			},
		},
		{
			name:   "ProviderSecretIsNotFound",
			Client: fake.NewFakeClient(newProvider(providerName).withSecret(ns, secretName, secretKey).Provider),
			bucket: newBucket(bucketName).withProvider(providerName).Bucket,
			want: want{
				err: errors.WithStack(
					errors.Errorf("cannot get provider's secret %s/%s: secrets \"%s\" not found", ns, secretName, secretName)),
			},
		},
		{
			name: "InvalidCredentials",
			Client: fake.NewFakeClient(newProvider(providerName).
				withSecret(ns, secretName, secretKey).Provider,
				newSecret(ns, secretName).Secret),
			bucket: newBucket(bucketName).withProvider(providerName).Bucket,
			want: want{
				err: errors.WithStack(
					errors.Errorf("cannot retrieve creds from json: unexpected end of JSON input")),
			},
		},
		{
			name: "Successful",
			Client: fake.NewFakeClient(newProvider(providerName).
				withSecret(ns, secretName, secretKey).Provider,
				newSecret(ns, secretName).withKeyData(secretKey, secretData).Secret),
			bucket: newBucket(bucketName).withUID("test-uid").withProvider(providerName).Bucket,
			want: want{
				// BUG(negz): This test is broken. It appears to intend to compare
				// unexported fields, but does not. This behaviour was maintained
				// when porting the test from https://github.com/go-test/deep to cmp.
				sd: newBucketSyncDeleter(
					newBucketClients(
						newBucket(bucketName).withUID("test-uid").withProvider(providerName).Bucket,
						nil, nil), ""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &bucketFactory{
				Client: tt.Client,
			}
			got, err := m.newSyncDeleter(ctx, tt.bucket)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("bucketFactory.newSyncDeleter() -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.sd, got, cmpopts.IgnoreUnexported(bucketSyncDeleter{})); diff != "" {
				t.Errorf("bucketFactory.newSyncDeleter() -want, +got:\n%s", diff)
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
