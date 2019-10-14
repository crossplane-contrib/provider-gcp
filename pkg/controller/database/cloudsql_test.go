package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-gcp/apis/database/v1alpha2"
	gcpv1alpha2 "github.com/crossplaneio/stack-gcp/apis/v1alpha2"
	"github.com/crossplaneio/stack-gcp/pkg/clients/cloudsql"
)

const (
	name      = "test-sql"
	namespace = "mynamespace"
	uid       = "2320sdasd-12312-asda"

	projectID          = "myproject-id-1234"
	providerName       = "gcp-provider"
	providerSecretName = "gcp-creds"
	providerSecretKey  = "creds"
	password           = "my_PassWord123!"
)

var errBoom = errors.New("boom")

type instanceModifier func(*v1alpha2.CloudsqlInstance)

func withConditions(c ...runtimev1alpha1.Condition) instanceModifier {
	return func(i *v1alpha2.CloudsqlInstance) { i.Status.SetConditions(c...) }
}

func withProviderState(s string) instanceModifier {
	return func(i *v1alpha2.CloudsqlInstance) { i.Status.AtProvider.State = s }
}

func withBindingPhase(p runtimev1alpha1.BindingPhase) instanceModifier {
	return func(i *v1alpha2.CloudsqlInstance) { i.Status.SetBindingPhase(p) }
}

func withPublicIP(ip string) instanceModifier {
	return func(i *v1alpha2.CloudsqlInstance) {
		i.Status.AtProvider.IPAddresses = append(i.Status.AtProvider.IPAddresses, &v1alpha2.IPMapping{
			IPAddress: ip,
			Type:      v1alpha2.PublicIPType,
		})
	}
}

func withPrivateIP(ip string) instanceModifier {
	return func(i *v1alpha2.CloudsqlInstance) {
		i.Status.AtProvider.IPAddresses = append(i.Status.AtProvider.IPAddresses, &v1alpha2.IPMapping{
			IPAddress: ip,
			Type:      v1alpha2.PrivateIPType,
		})
	}
}

// Mostly used for making a spec drift.
func withBackupConfigurationStartTime(h string) instanceModifier {
	return func(i *v1alpha2.CloudsqlInstance) {
		i.Spec.ForProvider.Settings.BackupConfiguration = &v1alpha2.BackupConfiguration{
			StartTime: &h,
		}
	}
}

func instance(im ...instanceModifier) *v1alpha2.CloudsqlInstance {
	i := &v1alpha2.CloudsqlInstance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  namespace,
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.ExternalNameAnnotationKey: name,
			},
		},
		Spec: v1alpha2.CloudsqlInstanceSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference:                &corev1.ObjectReference{Namespace: namespace, Name: providerName},
				WriteConnectionSecretToReference: corev1.LocalObjectReference{Name: providerSecretName},
			},
			ForProvider: v1alpha2.CloudsqlInstanceParameters{},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

func connDetails(password, privateIP, publicIP string) map[string][]byte {
	m := map[string][]byte{
		runtimev1alpha1.ResourceCredentialsSecretUserKey: []byte(v1alpha2.MysqlDefaultUser),
	}
	if password != "" {
		m[runtimev1alpha1.ResourceCredentialsSecretPasswordKey] = []byte(password)
	}
	if publicIP != "" {
		m[v1alpha2.PublicIPKey] = []byte(publicIP)
		m[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(publicIP)
	}
	if privateIP != "" {
		m[v1alpha2.PrivateIPKey] = []byte(privateIP)
		m[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(privateIP)
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

var _ resource.ExternalConnecter = &cloudsqlConnector{}
var _ resource.ExternalClient = &cloudsqlExternal{}

func TestConnect(t *testing.T) {
	provider := gcpv1alpha2.Provider{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerName},
		Spec: gcpv1alpha2.ProviderSpec{
			ProjectID: projectID,
			Secret: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: providerSecretName},
				Key:                  providerSecretKey,
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
			conn: &cloudsqlConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Namespace: namespace, Name: providerName}:
						*obj.(*gcpv1alpha2.Provider) = provider
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
					case client.ObjectKey{Namespace: namespace, Name: providerName}:
						*obj.(*gcpv1alpha2.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return errBoom
					}
					return nil
				}},
			},
			args: args{mg: instance()},
			want: want{err: errors.Wrap(errBoom, errProviderSecretNotRetrieved)},
		},
		"FailedToCreateCloudsqlInstanceClient": {
			conn: &cloudsqlConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Namespace: namespace, Name: providerName}:
						*obj.(*gcpv1alpha2.Provider) = provider
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
		obs resource.ExternalObservation
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
				if diff := cmp.Diff("GET", r.Method); diff != "" {
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
				if diff := cmp.Diff("GET", r.Method); diff != "" {
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
				if diff := cmp.Diff("GET", r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := instance(withBackupConfigurationStartTime("22:00"))
				_ = json.NewEncoder(w).Encode(cloudsql.GenerateDatabaseInstance(db.Spec.ForProvider, meta.GetExternalName(db)))
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
				if diff := cmp.Diff("GET", r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := cloudsql.GenerateDatabaseInstance(instance().Spec.ForProvider, meta.GetExternalName(instance()))
				db.State = v1alpha2.StateCreating
				_ = json.NewEncoder(w).Encode(db)
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				obs: resource.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: instance(withProviderState(v1alpha2.StateCreating), withConditions(runtimev1alpha1.Creating())),
			},
		},
		"Unavailable": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff("GET", r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := cloudsql.GenerateDatabaseInstance(instance().Spec.ForProvider, meta.GetExternalName(instance()))
				db.State = v1alpha2.StateMaintenance
				_ = json.NewEncoder(w).Encode(db)
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				obs: resource.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: instance(withProviderState(v1alpha2.StateMaintenance), withConditions(runtimev1alpha1.Unavailable())),
			},
		},
		"RunnableUnbound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff("GET", r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := cloudsql.GenerateDatabaseInstance(instance().Spec.ForProvider, meta.GetExternalName(instance()))
				db.State = v1alpha2.StateRunnable
				_ = json.NewEncoder(w).Encode(db)
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				obs: resource.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: connDetails("", "", ""),
				},
				mg: instance(
					withProviderState(v1alpha2.StateRunnable),
					withConditions(runtimev1alpha1.Available()),
					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound)),
			},
		},
		"RunnableConnectionGetFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff("GET", r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				db := cloudsql.GenerateDatabaseInstance(instance().Spec.ForProvider, meta.GetExternalName(instance()))
				db.State = v1alpha2.StateRunnable
				_ = json.NewEncoder(w).Encode(db)
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(errBoom),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				mg: instance(
					withProviderState(v1alpha2.StateRunnable),
					withConditions(runtimev1alpha1.Available()),
					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound)),
				err: errors.Wrap(errBoom, errConnectionNotRetrieved),
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
				user:      s.Users,
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
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg  resource.Managed
		cre resource.ExternalCreation
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
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
			}),
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(),
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
				mg: instance(),
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
				mg:  instance(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errInsertFailed),
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
				user:      s.Users,
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
			if diff := cmp.Diff(tc.want.cre, cre); diff != "" {
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
				mg:  instance(),
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
				mg:  instance(),
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
				mg:  instance(),
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
				user:      s.Users,
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
		upd resource.ExternalUpdate
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
				if strings.Contains(r.URL.Path, "/users") {
					switch r.Method {
					case http.MethodGet:
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
							Items: []*sqladmin.User{{Name: v1alpha2.MysqlDefaultUser}},
						})
					case http.MethodPut:
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
					}
				} else {
					if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				upd: resource.ExternalUpdate{
					ConnectionDetails: map[string][]byte{
						runtimev1alpha1.ResourceCredentialsSecretUserKey: []byte(v1alpha2.MysqlDefaultUser),
					},
				},
				mg:  instance(),
				err: nil,
			},
		},
		"CredentialUpdateFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if strings.Contains(r.URL.Path, "/users") {
					switch r.Method {
					case http.MethodGet:
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
							Items: []*sqladmin.User{{Name: v1alpha2.MysqlDefaultUser}},
						})
					case http.MethodPut:
						w.WriteHeader(http.StatusBadRequest)
						_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
					}
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateRootFailed),
			},
		},
		"PatchFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if strings.Contains(r.URL.Path, "/users") {
					switch r.Method {
					case http.MethodGet:
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
							Items: []*sqladmin.User{{Name: v1alpha2.MysqlDefaultUser}},
						})
					case http.MethodPut:
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
					}
				} else {
					if diff := cmp.Diff("PATCH", r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: instance(),
			},
			want: want{
				upd: resource.ExternalUpdate{
					ConnectionDetails: map[string][]byte{
						runtimev1alpha1.ResourceCredentialsSecretUserKey: []byte(v1alpha2.MysqlDefaultUser),
					},
				},
				mg:  instance(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errPatchFailed),
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
				user:      s.Users,
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
				if len(upd.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretPasswordKey]) == 0 {
					t.Errorf("Update(...): want password does not exist in connection details:\n")
				}
				delete(upd.ConnectionDetails, runtimev1alpha1.ResourceCredentialsSecretPasswordKey)
				if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
				if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}

		})
	}
}

func TestGetConnectionDetails(t *testing.T) {
	privateIP := "10.0.0.2"
	publicIP := "243.2.220.2"
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data: map[string][]byte{
			runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password),
		},
	}

	type args struct {
		cr *v1alpha2.CloudsqlInstance
	}
	type want struct {
		conn resource.ConnectionDetails
		err  error
	}

	cases := map[string]struct {
		kube client.Client
		args args
		want want
	}{
		"Successful": {
			kube: &test.MockClient{
				MockGet: func(_ context.Context, _ client.ObjectKey, obj runtime.Object) error {
					secret.DeepCopyInto(obj.(*corev1.Secret))
					return nil
				},
			},
			args: args{
				cr: instance(
					withPublicIP(publicIP),
					withPrivateIP(privateIP),
				),
			},
			want: want{
				conn: connDetails(password, privateIP, publicIP),
			},
		},
		"SecretGetFailed": {
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(errBoom),
			},
			args: args{
				cr: instance(),
			},
			want: want{
				err: errBoom,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := cloudsqlExternal{kube: tc.kube}
			conn, err := e.getConnectionDetails(context.TODO(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("getConnectionDetails(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.conn, conn); diff != "" {
				t.Errorf("getConnectionDetails(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdateRootCredentials(t *testing.T) {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data: map[string][]byte{
			runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password),
		},
	}
	type args struct {
		cr *v1alpha2.CloudsqlInstance
	}
	type want struct {
		conn resource.ConnectionDetails
		err  error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"SuccessfulGeneratedPassword": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
						Items: []*sqladmin.User{{Name: v1alpha2.MysqlDefaultUser}},
					})
				case http.MethodPut:
					user := &sqladmin.User{}
					_ = json.NewDecoder(r.Body).Decode(user)
					if user.Password == "" {
						t.Errorf("updateRootCredentials(...): no password is sent over\n")
					}
					if diff := cmp.Diff(user.Name, v1alpha2.MysqlDefaultUser); diff != "" {
						t.Errorf("updateRootCredentials(...): -want user name, +got user name:\n%s", diff)
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
				}
				_ = r.Body.Close()
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				cr: instance(),
			},
		},
		"SuccessfulExistingPassword": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
						Items: []*sqladmin.User{{Name: v1alpha2.MysqlDefaultUser}},
					})
				case http.MethodPut:
					user := &sqladmin.User{}
					_ = json.NewDecoder(r.Body).Decode(user)
					if diff := cmp.Diff(password, user.Password); diff != "" {
						t.Errorf("updateRootCredentials(...): -want password, +got password:\n%s", diff)
					}
					if diff := cmp.Diff(v1alpha2.MysqlDefaultUser, user.Name); diff != "" {
						t.Errorf("updateRootCredentials(...): -want user name, +got user name:\n%s", diff)
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&sqladmin.Operation{})
				}
				_ = r.Body.Close()
			}),
			kube: &test.MockClient{
				MockGet: func(_ context.Context, _ client.ObjectKey, obj runtime.Object) error {
					secret.DeepCopyInto(obj.(*corev1.Secret))
					return nil
				},
			},
			args: args{
				cr: instance(),
			},
			want: want{
				conn: connDetails(password, "", ""),
			},
		},
		"UserNotFound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{})
				}
				_ = r.Body.Close()
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				cr: instance(),
			},
			want: want{
				err: gError(http.StatusNotFound, fmt.Sprintf("user: %s is not found", v1alpha2.MysqlDefaultUser)),
			},
		},
		"GetConnectionDetailsFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
						Items: []*sqladmin.User{{Name: v1alpha2.MysqlDefaultUser}},
					})
				}
				_ = r.Body.Close()
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(errBoom),
			},
			args: args{
				cr: instance(),
			},
			want: want{
				err: errBoom,
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
				user:      s.Users,
			}
			_, err := e.updateRootCredentials(context.Background(), tc.args.cr)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("updateRootCredentials(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("updateRootCredentials(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
