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

package database

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/crossplaneio/stack-gcp/pkg/clients/cloudsql"
)

const (
	name      = "test-sql"
	namespace = "mynamespace"

	projectID          = "myproject-id-1234"
	providerName       = "gcp-provider"
	providerSecretName = "gcp-creds"
	providerSecretKey  = "creds"

	connectionName = "some:connection:name"
)

var errBoom = errors.New("boom")

type instanceModifier func(*v1beta1.CloudSQLInstance)

func withConditions(c ...runtimev1alpha1.Condition) instanceModifier {
	return func(i *v1beta1.CloudSQLInstance) { i.Status.SetConditions(c...) }
}

func withProviderState(s string) instanceModifier {
	return func(i *v1beta1.CloudSQLInstance) { i.Status.AtProvider.State = s }
}

func withBindingPhase(p runtimev1alpha1.BindingPhase) instanceModifier {
	return func(i *v1beta1.CloudSQLInstance) { i.Status.SetBindingPhase(p) }
}

func withPublicIP(ip string) instanceModifier {
	return func(i *v1beta1.CloudSQLInstance) {
		i.Status.AtProvider.IPAddresses = append(i.Status.AtProvider.IPAddresses, &v1beta1.IPMapping{
			IPAddress: ip,
			Type:      v1beta1.PublicIPType,
		})
	}
}

func withPrivateIP(ip string) instanceModifier {
	return func(i *v1beta1.CloudSQLInstance) {
		i.Status.AtProvider.IPAddresses = append(i.Status.AtProvider.IPAddresses, &v1beta1.IPMapping{
			IPAddress: ip,
			Type:      v1beta1.PrivateIPType,
		})
	}
}

func withConnectionName(cn string) instanceModifier {
	return func(i *v1beta1.CloudSQLInstance) {
		i.Status.AtProvider.ConnectionName = cn
	}
}

// Mostly used for making a spec drift.
func withBackupConfigurationStartTime(h string) instanceModifier {
	return func(i *v1beta1.CloudSQLInstance) {
		i.Spec.ForProvider.Settings.BackupConfiguration = &v1beta1.BackupConfiguration{
			StartTime: &h,
		}
	}
}

func instance(im ...instanceModifier) *v1beta1.CloudSQLInstance {
	i := &v1beta1.CloudSQLInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.ExternalNameAnnotationKey: name,
			},
		},
		Spec: v1beta1.CloudSQLInstanceSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
			ForProvider: v1beta1.CloudSQLInstanceParameters{},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

func connDetails(privateIP, publicIP string, additions ...map[string][]byte) managed.ConnectionDetails {
	m := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretUserKey: []byte(v1beta1.MysqlDefaultUser),
		v1beta1.CloudSQLSecretConnectionName:             []byte(""),
	}
	if publicIP != "" {
		m[v1beta1.PublicIPKey] = []byte(publicIP)
		m[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(publicIP)
	}
	if privateIP != "" {
		m[v1beta1.PrivateIPKey] = []byte(privateIP)
		m[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(privateIP)
	}
	for _, a := range additions {
		for k, v := range a {
			m[k] = v
		}
	}
	return m
}

func gError(code int, message string) *googleapi.Error {
	return &googleapi.Error{
		Code:    code,
		Body:    "{}\n",
		Message: message,
	}
}

var _ managed.ExternalConnecter = &cloudsqlConnector{}
var _ managed.ExternalClient = &cloudsqlExternal{}

func TestConnect(t *testing.T) {
	provider := gcpv1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: gcpv1alpha3.ProviderSpec{
			ProjectID: projectID,
			ProviderSpec: runtimev1alpha1.ProviderSpec{
				CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
					SecretReference: runtimev1alpha1.SecretReference{
						Namespace: namespace,
						Name:      providerSecretName,
					},
					Key: providerSecretKey,
				},
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
		conn managed.ExternalConnecter
		args args
		want want
	}{
		"Connected": {
			conn: &cloudsqlConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newServiceFn: func(ctx context.Context, opts ...option.ClientOption) (*sqladmin.Service, error) {
					return &sqladmin.Service{}, nil
				},
			},
			args: args{
				mg: instance(),
			},
			want: want{
				err: nil,
			},
		},
		"FailedToGetProvider": {
			conn: &cloudsqlConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					return errBoom
				}},
			},
			args: args{
				mg: instance(),
			},
			want: want{
				err: errors.Wrap(errBoom, errProviderNotRetrieved),
			},
		},
		"FailedToGetProviderSecret": {
			conn: &cloudsqlConnector{
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
			args: args{mg: instance()},
			want: want{err: errors.Wrap(errBoom, errProviderSecretNotRetrieved)},
		},
		"ProviderSecretNil": {
			conn: &cloudsqlConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						nilSecretProvider := provider
						nilSecretProvider.SetCredentialsSecretReference(nil)
						*obj.(*gcpv1alpha3.Provider) = nilSecretProvider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return errBoom
					}
					return nil
				}},
			},
			args: args{mg: instance()},
			want: want{err: errors.New(errProviderSecretNil)},
		},
		"FailedToCreateCloudSQLInstanceClient": {
			conn: &cloudsqlConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newServiceFn: func(_ context.Context, _ ...option.ClientOption) (*sqladmin.Service, error) { return nil, errBoom },
			},
			args: args{mg: instance()},
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

func TestObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"NotFound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(&sqladmin.DatabaseInstance{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(),
				err: nil,
			},
		},
		"GetFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&sqladmin.DatabaseInstance{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetFailed),
			},
		},
		"NotUpToDateSpecUpdateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				instance := instance(withBackupConfigurationStartTime("22:00"))
				db := &sqladmin.DatabaseInstance{}
				cloudsql.GenerateDatabaseInstance(meta.GetExternalName(instance), instance.Spec.ForProvider, db)
				_ = json.NewEncoder(w).Encode(db)
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(errBoom),
			},
			args: args{

				mg: instance(),
			},
			want: want{
				mg:  instance(withBackupConfigurationStartTime("22:00")),
				err: errors.Wrap(errBoom, errManagedUpdateFailed),
			},
		},
		"Creating": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := &sqladmin.DatabaseInstance{}
				cloudsql.GenerateDatabaseInstance(meta.GetExternalName(instance()), instance().Spec.ForProvider, db)
				db.State = v1beta1.StateCreating
				_ = json.NewEncoder(w).Encode(db)
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: connDetails("", ""),
				},
				mg: instance(withProviderState(v1beta1.StateCreating), withConditions(runtimev1alpha1.Creating())),
			},
		},
		"Unavailable": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := &sqladmin.DatabaseInstance{}
				cloudsql.GenerateDatabaseInstance(meta.GetExternalName(instance()), instance().Spec.ForProvider, db)
				db.State = v1beta1.StateMaintenance
				_ = json.NewEncoder(w).Encode(db)
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: connDetails("", ""),
				},
				mg: instance(withProviderState(v1beta1.StateMaintenance), withConditions(runtimev1alpha1.Unavailable())),
			},
		},
		"RunnableUnbound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := &sqladmin.DatabaseInstance{}
				cloudsql.GenerateDatabaseInstance(meta.GetExternalName(instance()), instance().Spec.ForProvider, db)
				db.ConnectionName = connectionName
				db.State = v1beta1.StateRunnable
				_ = json.NewEncoder(w).Encode(db)
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: connDetails("", "", map[string][]byte{v1beta1.CloudSQLSecretConnectionName: []byte(connectionName)}),
				},
				mg: instance(
					withProviderState(v1beta1.StateRunnable),
					withConditions(runtimev1alpha1.Available()),
					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound),
					withConnectionName(connectionName)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := cloudsqlExternal{
				kube:      tc.kube,
				projectID: projectID,
				db:        s.Instances,
			}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Observe(...): want error string != got error string:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Observe(...): want error != got error:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	wantRandom := "i-want-random-data-not-this-special-string"

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg  resource.Managed
		cre managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				i := &sqladmin.DatabaseInstance{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				if len(i.RootPassword) == 0 {
					t.Errorf("r: wanted root password, got:%s", i.RootPassword)
				}
				w.WriteHeader(http.StatusOK)
				_ = r.Body.Close()
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg: instance(withConditions(runtimev1alpha1.Creating())),
				cre: managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{
					runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(wantRandom),
				}},
				err: nil,
			},
		},
		"AlreadyExists": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(gError(http.StatusConflict, ""), errNameInUse),
			},
		},
		"Failed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := cloudsqlExternal{
				kube:      tc.kube,
				projectID: projectID,
				db:        s.Instances,
			}
			cre, err := e.Create(tc.args.ctx, tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.cre, cre, cmp.Comparer(func(a, b managed.ConnectionDetails) bool {
				// This special comparer considers two ConnectionDetails to be
				// equal if one has the special password value wantRandom and
				// the other has a non-zero password string. If neither has the
				// special password value it falls back to default compare
				// semantics.

				av := string(a[runtimev1alpha1.ResourceCredentialsSecretPasswordKey])
				bv := string(b[runtimev1alpha1.ResourceCredentialsSecretPasswordKey])

				if av == wantRandom {
					return len(bv) > 0
				}

				if bv == wantRandom {
					return len(av) > 0
				}

				return cmp.Equal(a, b)
			})); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(withConditions(runtimev1alpha1.Deleting())),
				err: nil,
			},
		},
		"AlreadyGone": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(withConditions(runtimev1alpha1.Deleting())),
				err: nil,
			},
		},
		"Failed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := cloudsqlExternal{
				kube:      tc.kube,
				projectID: projectID,
				db:        s.Instances,
			}
			err := e.Delete(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		upd managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(),
				err: nil,
			},
		},
		"NoUpdateNecessary": {
			args: args{
				mg: instance(withProviderState(v1beta1.StateCreating)),
			},
			want: want{
				mg:  instance(withProviderState(v1beta1.StateCreating)),
				err: nil,
			},
		},
		"PatchFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				upd: managed.ExternalUpdate{
					ConnectionDetails: map[string][]byte{
						runtimev1alpha1.ResourceCredentialsSecretUserKey: []byte(v1beta1.MysqlDefaultUser),
					},
				},
				mg:  instance(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := sqladmin.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := cloudsqlExternal{
				kube:      tc.kube,
				projectID: projectID,
				db:        s.Instances,
			}
			upd, err := e.Update(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}
			if tc.want.err == nil {
				if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
				if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}

		})
	}
}

func TestGetConnectionDetails(t *testing.T) {
	privateIP := "10.0.0.2"
	publicIP := "243.2.220.2"
	cert := "My-precious-cert"
	commonName := "And-its-precious-common-name"

	type args struct {
		cr *v1beta1.CloudSQLInstance
		i  *sqladmin.DatabaseInstance
	}
	type want struct {
		conn managed.ConnectionDetails
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				cr: instance(
					withPublicIP(publicIP),
					withPrivateIP(privateIP),
				),
				i: &sqladmin.DatabaseInstance{
					ServerCaCert: &sqladmin.SslCert{
						Cert:       cert,
						CommonName: commonName,
					},
				},
			},
			want: want{
				conn: connDetails(privateIP, publicIP, map[string][]byte{
					v1beta1.CloudSQLSecretServerCACertificateCertKey:             []byte(cert),
					v1beta1.CloudSQLSecretServerCACertificateCommonNameKey:       []byte(commonName),
					v1beta1.CloudSQLSecretServerCACertificateCertSerialNumberKey: []byte(""),
					v1beta1.CloudSQLSecretServerCACertificateExpirationTimeKey:   []byte(""),
					v1beta1.CloudSQLSecretServerCACertificateCreateTimeKey:       []byte(""),
					v1beta1.CloudSQLSecretServerCACertificateInstanceKey:         []byte(""),
					v1beta1.CloudSQLSecretServerCACertificateSha1FingerprintKey:  []byte(""),
					v1beta1.CloudSQLSecretConnectionName:                         []byte(""),
				}),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			conn := getConnectionDetails(tc.args.cr, tc.args.i)
			if diff := cmp.Diff(tc.want.conn, conn); diff != "" {
				t.Errorf("getConnectionDetails(...): -want, +got:\n%s", diff)
			}
		})
	}
}
