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

package v1beta1

import (
	"context"
	"testing"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1beta1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
)

const (
	name      = "test-cluster"
	namespace = "mynamespace"

	projectID          = "myproject-id-1234"
	providerName       = "gcp-provider"
	providerSecretName = "gcp-creds"
	providerSecretKey  = "creds"

	password = "my_PassWord123!"
)

var errBoom = errors.New("boom")

var _ resource.ExternalConnecter = &clusterConnector{}
var _ resource.ExternalClient = &clusterExternal{}

type clusterModifier func(*v1beta1.GKECluster)

func cluster(im ...clusterModifier) *v1beta1.GKECluster {
	i := &v1beta1.GKECluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.ExternalNameAnnotationKey: name,
			},
		},
		Spec: v1beta1.GKEClusterSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
			ForProvider: v1beta1.GKEClusterParameters{},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

func TestConnect(t *testing.T) {
	provider := gcpv1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: gcpv1alpha3.ProviderSpec{
			ProjectID: projectID,
			Secret: runtimev1alpha1.SecretKeySelector{
				SecretReference: runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      providerSecretName,
				},
				Key: providerSecretKey,
			},
		},
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte("olala")},
	}

	type args struct {
		mg resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		conn resource.ExternalConnecter
		args args
		want want
	}{
		"Connected": {
			conn: &clusterConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newServiceFn: func(ctx context.Context, opts ...option.ClientOption) (*container.Service, error) {
					return &container.Service{}, nil
				},
			},
			args: args{
				mg: cluster(),
			},
			want: want{
				err: nil,
			},
		},
		"FailedToGetProvider": {
			conn: &clusterConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					return errBoom
				}},
			},
			args: args{
				mg: cluster(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProvider),
			},
		},
		"FailedToGetProviderSecret": {
			conn: &clusterConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return errBoom
					}
					return nil
				}},
			},
			args: args{mg: cluster()},
			want: want{err: errors.Wrap(errBoom, errGetProviderSecret)},
		},
		"FailedToCreateCloudSQLInstanceClient": {
			conn: &clusterConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newServiceFn: func(_ context.Context, _ ...option.ClientOption) (*container.Service, error) { return nil, errBoom },
			},
			args: args{mg: cluster()},
			want: want{err: errors.Wrap(errBoom, errNewClient)},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.conn.Connect(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}
		})
	}
}

// func TestObserve(t *testing.T) {
// 	type args struct {
// 		mg resource.Managed
// 	}
// 	type want struct {
// 		mg  resource.Managed
// 		obs resource.ExternalObservation
// 		err error
// 	}

// 	cases := map[string]struct {
// 		handler http.Handler
// 		kube    client.Client
// 		args    args
// 		want    want
// 	}{
// 		"NotFound": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff("GET", r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusNotFound)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.DatabaseInstance{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  cluster(),
// 				err: nil,
// 			},
// 		},
// 		"GetFailed": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff("GET", r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusBadRequest)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.DatabaseInstance{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  cluster(),
// 				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetFailed),
// 			},
// 		},
// 		"NotUpToDateSpecUpdateFailed": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff("GET", r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusOK)
// 				db := instance(withBackupConfigurationStartTime("22:00"))
// 				_ = json.NewEncoder(w).Encode(cloudsql.GenerateDatabaseInstance(db.Spec.ForProvider, meta.GetExternalName(db)))
// 			}),
// 			kube: &test.MockClient{
// 				MockUpdate: test.NewMockUpdateFn(errBoom),
// 			},
// 			args: args{

// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  instance(withBackupConfigurationStartTime("22:00")),
// 				err: errors.Wrap(errBoom, errManagedUpdateFailed),
// 			},
// 		},
// 		"Creating": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff("GET", r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusOK)
// 				db := cloudsql.GenerateDatabaseInstance(cluster().Spec.ForProvider, meta.GetExternalName(cluster()))
// 				db.State = v1beta1.StateCreating
// 				_ = json.NewEncoder(w).Encode(db)
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				obs: resource.ExternalObservation{
// 					ResourceExists:    true,
// 					ResourceUpToDate:  true,
// 					ConnectionDetails: connDetails("", ""),
// 				},
// 				mg: instance(withProviderState(v1beta1.StateCreating), withConditions(runtimev1alpha1.Creating())),
// 			},
// 		},
// 		"Unavailable": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff("GET", r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusOK)
// 				db := cloudsql.GenerateDatabaseInstance(cluster().Spec.ForProvider, meta.GetExternalName(cluster()))
// 				db.State = v1beta1.StateMaintenance
// 				_ = json.NewEncoder(w).Encode(db)
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				obs: resource.ExternalObservation{
// 					ResourceExists:    true,
// 					ResourceUpToDate:  true,
// 					ConnectionDetails: connDetails("", ""),
// 				},
// 				mg: instance(withProviderState(v1beta1.StateMaintenance), withConditions(runtimev1alpha1.Unavailable())),
// 			},
// 		},
// 		"RunnableUnbound": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff("GET", r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusOK)
// 				db := cloudsql.GenerateDatabaseInstance(cluster().Spec.ForProvider, meta.GetExternalName(cluster()))
// 				db.State = v1beta1.StateRunnable
// 				_ = json.NewEncoder(w).Encode(db)
// 			}),
// 			kube: &test.MockClient{
// 				MockGet: test.NewMockGetFn(nil),
// 			},
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				obs: resource.ExternalObservation{
// 					ResourceExists:    true,
// 					ResourceUpToDate:  true,
// 					ConnectionDetails: connDetails("", ""),
// 				},
// 				mg: instance(
// 					withProviderState(v1beta1.StateRunnable),
// 					withConditions(runtimev1alpha1.Available()),
// 					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound)),
// 			},
// 		},
// 	}

// 	for name, tc := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			server := httptest.NewServer(tc.handler)
// 			defer server.Close()
// 			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
// 			e := cloudsqlExternal{
// 				kube:      tc.kube,
// 				projectID: projectID,
// 				db:        s.Instances,
// 			}
// 			obs, err := e.Observe(context.Background(), tc.args.mg)
// 			if tc.want.err != nil && err != nil {
// 				// the case where our mock server returns error.
// 				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
// 					t.Errorf("Observe(...): want error string != got error string:\n%s", diff)
// 				}
// 			} else {
// 				if diff := cmp.Diff(tc.want.err, err); diff != "" {
// 					t.Errorf("Observe(...): want error != got error:\n%s", diff)
// 				}
// 			}
// 			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
// 				t.Errorf("Observe(...): -want, +got:\n%s", diff)
// 			}
// 			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
// 				t.Errorf("Observe(...): -want, +got:\n%s", diff)
// 			}
// 		})
// 	}
// }

// func TestCreate(t *testing.T) {
// 	wantRandom := "i-want-random-data-not-this-special-string"

// 	type args struct {
// 		ctx context.Context
// 		mg  resource.Managed
// 	}
// 	type want struct {
// 		mg  resource.Managed
// 		cre resource.ExternalCreation
// 		err error
// 	}

// 	cases := map[string]struct {
// 		handler http.Handler
// 		kube    client.Client
// 		args    args
// 		want    want
// 	}{
// 		"Successful": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				i := &sqladmin.DatabaseInstance{}
// 				b, err := ioutil.ReadAll(r.Body)
// 				if diff := cmp.Diff(err, nil); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				err = json.Unmarshal(b, i)
// 				if diff := cmp.Diff(err, nil); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				if len(i.RootPassword) == 0 {
// 					t.Errorf("r: wanted root password, got:%s", i.RootPassword)
// 				}
// 				w.WriteHeader(http.StatusOK)
// 				_ = r.Body.Close()
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg: instance(withConditions(runtimev1alpha1.Creating())),
// 				cre: resource.ExternalCreation{ConnectionDetails: resource.ConnectionDetails{
// 					runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(wantRandom),
// 				}},
// 				err: nil,
// 			},
// 		},
// 		"AlreadyExists": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusConflict)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  instance(withConditions(runtimev1alpha1.Creating())),
// 				err: errors.Wrap(gError(http.StatusConflict, ""), errCreateFailed),
// 			},
// 		},
// 		"Failed": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusBadRequest)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  instance(withConditions(runtimev1alpha1.Creating())),
// 				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateFailed),
// 			},
// 		},
// 	}

// 	for name, tc := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			server := httptest.NewServer(tc.handler)
// 			defer server.Close()
// 			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
// 			e := cloudsqlExternal{
// 				kube:      tc.kube,
// 				projectID: projectID,
// 				db:        s.Instances,
// 			}
// 			cre, err := e.Create(tc.args.ctx, tc.args.mg)
// 			if tc.want.err != nil && err != nil {
// 				// the case where our mock server returns error.
// 				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
// 					t.Errorf("Create(...): -want, +got:\n%s", diff)
// 				}
// 			} else {
// 				if diff := cmp.Diff(tc.want.err, err); diff != "" {
// 					t.Errorf("Create(...): -want, +got:\n%s", diff)
// 				}
// 			}
// 			if diff := cmp.Diff(tc.want.cre, cre, cmp.Comparer(func(a, b resource.ConnectionDetails) bool {
// 				// This special comparer considers two ConnectionDetails to be
// 				// equal if one has the special password value wantRandom and
// 				// the other has a non-zero password string. If neither has the
// 				// special password value it falls back to default compare
// 				// semantics.

// 				av := string(a[runtimev1alpha1.ResourceCredentialsSecretPasswordKey])
// 				bv := string(b[runtimev1alpha1.ResourceCredentialsSecretPasswordKey])

// 				if av == wantRandom {
// 					return len(bv) > 0
// 				}

// 				if bv == wantRandom {
// 					return len(av) > 0
// 				}

// 				return cmp.Equal(a, b)
// 			})); diff != "" {
// 				t.Errorf("Create(...): -want, +got:\n%s", diff)
// 			}
// 			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
// 				t.Errorf("Create(...): -want, +got:\n%s", diff)
// 			}
// 		})
// 	}
// }

// func TestDelete(t *testing.T) {
// 	type args struct {
// 		mg resource.Managed
// 	}
// 	type want struct {
// 		mg  resource.Managed
// 		err error
// 	}

// 	cases := map[string]struct {
// 		handler http.Handler
// 		kube    client.Client
// 		args    args
// 		want    want
// 	}{
// 		"Successful": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusOK)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  instance(withConditions(runtimev1alpha1.Deleting())),
// 				err: nil,
// 			},
// 		},
// 		"AlreadyGone": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusNotFound)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  instance(withConditions(runtimev1alpha1.Deleting())),
// 				err: nil,
// 			},
// 		},
// 		"Failed": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusBadRequest)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  instance(withConditions(runtimev1alpha1.Deleting())),
// 				err: errors.Wrap(gError(http.StatusBadRequest, ""), errDeleteFailed),
// 			},
// 		},
// 	}

// 	for name, tc := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			server := httptest.NewServer(tc.handler)
// 			defer server.Close()
// 			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
// 			e := cloudsqlExternal{
// 				kube:      tc.kube,
// 				projectID: projectID,
// 				db:        s.Instances,
// 			}
// 			err := e.Delete(context.Background(), tc.args.mg)
// 			if tc.want.err != nil && err != nil {
// 				// the case where our mock server returns error.
// 				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
// 					t.Errorf("Delete(...): -want, +got:\n%s", diff)
// 				}
// 			} else {
// 				if diff := cmp.Diff(tc.want.err, err); diff != "" {
// 					t.Errorf("Delete(...): -want, +got:\n%s", diff)
// 				}
// 			}
// 			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
// 				t.Errorf("Delete(...): -want, +got:\n%s", diff)
// 			}
// 		})
// 	}
// }

// func TestUpdate(t *testing.T) {
// 	type args struct {
// 		mg resource.Managed
// 	}
// 	type want struct {
// 		mg  resource.Managed
// 		upd resource.ExternalUpdate
// 		err error
// 	}

// 	cases := map[string]struct {
// 		handler http.Handler
// 		kube    client.Client
// 		args    args
// 		want    want
// 	}{
// 		"Successful": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusOK)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			kube: &test.MockClient{
// 				MockGet: test.NewMockGetFn(nil),
// 			},
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				mg:  cluster(),
// 				err: nil,
// 			},
// 		},
// 		"PatchFails": {
// 			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				_ = r.Body.Close()
// 				if diff := cmp.Diff("PATCH", r.Method); diff != "" {
// 					t.Errorf("r: -want, +got:\n%s", diff)
// 				}
// 				w.WriteHeader(http.StatusBadRequest)
// 				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
// 			}),
// 			kube: &test.MockClient{
// 				MockGet: test.NewMockGetFn(nil),
// 			},
// 			args: args{
// 				mg: cluster(),
// 			},
// 			want: want{
// 				upd: resource.ExternalUpdate{
// 					ConnectionDetails: map[string][]byte{
// 						runtimev1alpha1.ResourceCredentialsSecretUserKey: []byte(v1beta1.MysqlDefaultUser),
// 					},
// 				},
// 				mg:  cluster(),
// 				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateFailed),
// 			},
// 		},
// 	}

// 	for name, tc := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			server := httptest.NewServer(tc.handler)
// 			defer server.Close()
// 			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
// 			e := cloudsqlExternal{
// 				kube:      tc.kube,
// 				projectID: projectID,
// 				db:        s.Instances,
// 			}
// 			upd, err := e.Update(context.Background(), tc.args.mg)
// 			if tc.want.err != nil && err != nil {
// 				// the case where our mock server returns error.
// 				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
// 					t.Errorf("Update(...): -want, +got:\n%s", diff)
// 				}
// 			} else {
// 				if diff := cmp.Diff(tc.want.err, err); diff != "" {
// 					t.Errorf("Update(...): -want, +got:\n%s", diff)
// 				}
// 			}
// 			if tc.want.err == nil {
// 				if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
// 					t.Errorf("Update(...): -want, +got:\n%s", diff)
// 				}
// 				if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
// 					t.Errorf("Update(...): -want, +got:\n%s", diff)
// 				}
// 			}

// 		})
// 	}
// }
