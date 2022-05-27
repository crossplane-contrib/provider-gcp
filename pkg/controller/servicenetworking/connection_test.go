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
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	servicenetworking "google.golang.org/api/servicenetworking/v1"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-gcp/apis/servicenetworking/v1beta1"
	"github.com/crossplane-contrib/provider-gcp/pkg/clients/connection"
)

var (
	errGoogleNotFound = &googleapi.Error{Code: http.StatusNotFound, Message: "boom"}
	errGoogleConflict = &googleapi.Error{Code: http.StatusConflict, Message: "boom"}
	errGoogleOther    = &googleapi.Error{Code: http.StatusInternalServerError, Message: "boom"}

	unexpected resource.Managed
)

func conn() *v1beta1.Connection {
	return &v1beta1.Connection{
		Spec: v1beta1.ConnectionSpec{},
		Status: v1beta1.ConnectionStatus{
			AtProvider: v1beta1.ConnectionObservation{
				Peering: connection.PeeringName,
			},
		},
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
