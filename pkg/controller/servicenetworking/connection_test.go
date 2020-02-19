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

package servicenetworking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	servicenetworking "google.golang.org/api/servicenetworking/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/stack-gcp/apis/servicenetworking/v1beta1"
	gcpv1alpha3 "github.com/crossplane/stack-gcp/apis/v1alpha3"
	"github.com/crossplane/stack-gcp/pkg/clients/connection"
)

var (
	errBoom           = errors.New("boom")
	errGoogleNotFound = &googleapi.Error{Code: http.StatusNotFound, Message: "boom"}
	errGoogleConflict = &googleapi.Error{Code: http.StatusConflict, Message: "boom"}
	errGoogleOther    = &googleapi.Error{Code: http.StatusInternalServerError, Message: "boom"}
)

var (
	unexpected resource.Managed

	projectID          = "myproject-id-1234"
	providerName       = "gcp-provider"
	providerSecretName = "gcp-creds"
	providerSecretKey  = "creds"
	namespace          = "test"
)

func conn() *v1beta1.Connection {
	return &v1beta1.Connection{
		Spec: v1beta1.ConnectionSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{ProviderReference: &corev1.ObjectReference{
				Name: providerName,
			}},
		},
		Status: v1beta1.ConnectionStatus{
			AtProvider: v1beta1.ConnectionObservation{
				Peering: connection.PeeringName,
			},
		},
	}
}

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
		Data:       map[string][]byte{providerSecretKey: []byte("verysecret")},
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
			conn: &connector{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newCompute: func(ctx context.Context, opts ...option.ClientOption) (*compute.Service, error) {
					return &compute.Service{}, nil
				},
				newServiceNetworking: func(ctx context.Context, opts ...option.ClientOption) (*servicenetworking.APIService, error) {
					return &servicenetworking.APIService{}, nil
				},
			},
			args: args{
				mg: conn(),
			},
			want: want{
				err: nil,
			},
		},
		"FailedToGetProvider": {
			conn: &connector{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					return errBoom
				}},
			},
			args: args{
				mg: conn(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProvider),
			},
		},
		"FailedToGetProviderSecret": {
			conn: &connector{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return errBoom
					}
					return nil
				}},
			},
			args: args{mg: conn()},
			want: want{err: errors.Wrap(errBoom, errGetProviderSecret)},
		},
		"ProviderSecretNil": {
			conn: &connector{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
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
			args: args{mg: conn()},
			want: want{err: errors.New(errProviderSecretNil)},
		},
		"FailedToCreateComputeClient": {
			conn: &connector{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newCompute: func(_ context.Context, _ ...option.ClientOption) (*compute.Service, error) { return nil, errBoom },
			},
			args: args{mg: conn()},
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
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		eo  managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotConnectionError": {
			e: &external{},
			args: args{
				ctx: context.Background(),
				mg:  unexpected,
			},
			want: want{
				err: errors.New(errNotConnection),
			},
		},
		"ErrorListConnections": {
			e: &external{
				sn: FakeServiceNetworkingService{WantMethod: http.MethodGet, ReturnError: errGoogleOther}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: want{
				err: errors.Wrap(errGoogleOther, errListConnections),
			},
		},
		"ConnectionDoesNotExist": {
			e: &external{
				sn: FakeServiceNetworkingService{
					WantMethod: http.MethodGet,
					Return: &servicenetworking.ListConnectionsResponse{Connections: []*servicenetworking.Connection{
						{Peering: connection.PeeringName + "-diff"},
					}},
				}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ErrorGetNetwork": {
			e: &external{
				sn: FakeServiceNetworkingService{
					WantMethod: http.MethodGet,
					Return: &servicenetworking.ListConnectionsResponse{Connections: []*servicenetworking.Connection{
						{Peering: connection.PeeringName},
					}},
				}.Serve(t),
				compute: FakeComputeService{WantMethod: http.MethodGet, ReturnError: errGoogleNotFound}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: want{
				err: errors.Wrap(errGoogleNotFound, errGetNetwork),
			},
		},
		"ConnectionExists": {
			e: &external{
				sn: FakeServiceNetworkingService{
					WantMethod: http.MethodGet,
					Return: &servicenetworking.ListConnectionsResponse{Connections: []*servicenetworking.Connection{
						{Peering: connection.PeeringName},
					}},
				}.Serve(t),
				compute: FakeComputeService{WantMethod: http.MethodGet}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.eo, got); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
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
		ec  managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotConnectionError": {
			e: &external{},
			args: args{
				ctx: context.Background(),
				mg:  unexpected,
			},
			want: want{
				err: errors.New(errNotConnection),
			},
		},
		"ErrorCreateConnection": {
			e: &external{
				sn: FakeServiceNetworkingService{WantMethod: http.MethodPatch, ReturnError: errGoogleOther}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: want{
				err: errors.Wrap(errGoogleOther, errCreateConnection),
			},
		},
		"ConnectionAlreadyExists": {
			e: &external{
				sn: FakeServiceNetworkingService{WantMethod: http.MethodPatch, ReturnError: errGoogleConflict}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: want{},
		},
		"ConnectionCreated": {
			e: &external{
				sn: FakeServiceNetworkingService{WantMethod: http.MethodPatch, Return: &servicenetworking.Operation{}}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.ec, got); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		eu  managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotConnectionError": {
			e: &external{},
			args: args{
				ctx: context.Background(),
				mg:  unexpected,
			},
			want: want{
				err: errors.New(errNotConnection),
			},
		},
		"ErrorUpdateConnection": {
			e: &external{
				sn: FakeServiceNetworkingService{WantMethod: http.MethodPatch, ReturnError: errGoogleOther}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: want{
				err: errors.Wrap(errGoogleOther, errUpdateConnection),
			},
		},
		"ConnectionUpdated": {
			e: &external{
				sn: FakeServiceNetworkingService{WantMethod: http.MethodPatch, Return: &servicenetworking.Operation{}}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.e.Update(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.eu, got); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want error
	}{
		"NotConnectionError": {
			e: &external{},
			args: args{
				ctx: context.Background(),
				mg:  unexpected,
			},
			want: errors.New(errNotConnection),
		},
		"ErrorDeleteConnection": {
			e: &external{
				compute: FakeComputeService{WantMethod: http.MethodPost, ReturnError: errGoogleOther}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
			want: errors.Wrap(errGoogleOther, errDeleteConnection),
		},
		"ConnectionNotFound": {
			e: &external{
				compute: FakeComputeService{WantMethod: http.MethodPost, ReturnError: errGoogleNotFound}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
		},
		"ConnectionDeleted": {
			e: &external{
				compute: FakeComputeService{WantMethod: http.MethodPost, Return: &compute.Operation{}}.Serve(t),
			},
			args: args{
				ctx: context.Background(),
				mg:  conn(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

type FakeComputeService struct {
	WantMethod string

	ReturnError error
	Return      interface{}
}

func (s FakeComputeService) Serve(t *testing.T) *compute.Service {
	// NOTE(negz): We never close this httptest.Server because returning only a
	// compute.Service makes for a simpler test fake API. We create one server
	// per test case, but they only live for the invocation of the test run.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()

		if r.Method != s.WantMethod {
			http.Error(w, fmt.Sprintf("want HTTP method %s, got %s", s.WantMethod, r.Method), http.StatusBadRequest)
			return
		}

		if gae, ok := s.ReturnError.(*googleapi.Error); ok {
			w.WriteHeader(gae.Code)
			_ = json.NewEncoder(w).Encode(struct {
				Error *googleapi.Error `json:"error"`
			}{Error: gae})
			return
		}

		if s.ReturnError != nil {
			http.Error(w, s.ReturnError.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&compute.Operation{})
	}))

	c, err := compute.NewService(context.Background(),
		option.WithEndpoint(srv.URL),
		option.WithoutAuthentication())
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type FakeServiceNetworkingService struct {
	WantMethod string

	ReturnError error
	Return      interface{}
}

func (s FakeServiceNetworkingService) Serve(t *testing.T) *servicenetworking.APIService {
	// NOTE(negz): We never close this httptest.Server because returning only a
	// servicenetworking.APIService makes for a simpler test fake API. We create
	// one server per test case, but they only live for the invocation of the
	// test run.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()

		if r.Method != s.WantMethod {
			http.Error(w, fmt.Sprintf("want HTTP method %s, got %s", s.WantMethod, r.Method), http.StatusBadRequest)
			return
		}

		if gae, ok := s.ReturnError.(*googleapi.Error); ok {
			w.WriteHeader(gae.Code)
			_ = json.NewEncoder(w).Encode(struct {
				Error *googleapi.Error `json:"error"`
			}{Error: gae})
			return
		}

		if s.ReturnError != nil {
			http.Error(w, s.ReturnError.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(s.Return)
	}))

	c, err := servicenetworking.NewService(context.Background(),
		option.WithEndpoint(srv.URL),
		option.WithoutAuthentication())
	if err != nil {
		t.Fatal(err)
	}
	return c
}
