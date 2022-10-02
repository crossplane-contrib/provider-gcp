/*
Copyright 2022 The Crossplane Authors.

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

package dns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	dns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-gcp/apis/dns/v1alpha1"
	mzclient "github.com/crossplane-contrib/provider-gcp/pkg/clients/managedzone"
)

const (
	managedZoneProjectID = "myproject-id-1234"
)

var (
	nonManagedZone    resource.Managed
	managedZoneLabels = map[string]string{"foo": "bar"}
)

type ManagedZoneOption func(*v1alpha1.ManagedZone)

func newManagedZone(opts ...ManagedZoneOption) *v1alpha1.ManagedZone {
	mz := &v1alpha1.ManagedZone{}

	for _, f := range opts {
		f(mz)
	}

	return mz
}

func withLabels(l map[string]string) ManagedZoneOption {
	return func(mz *v1alpha1.ManagedZone) {
		mz.Spec.ForProvider.Labels = l
	}
}

func managedZoneGError(code int, message string) *googleapi.Error {
	return &googleapi.Error{
		Code:    code,
		Body:    "{}\n",
		Message: message,
	}
}

func TestManagedZoneObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		e   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason  string
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"NotManagedZone": {
			reason: "Should return an error if the resource is not ManagedZone",
			args: args{
				mg: nonManagedZone,
			},
			want: want{
				e:   managed.ExternalObservation{},
				err: errors.New(errNotManagedZone),
			},
		},
		"ResourceNotFound": {
			reason: "Should not return an error if the API response is 404",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e:   managed.ExternalObservation{},
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"InternalError": {
			reason: "Should return an error if the error is different than 404",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e:   managed.ExternalObservation{},
				err: errors.Wrap(managedZoneGError(http.StatusInternalServerError, ""), errGetManagedZone),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"UpdateResourceSpecSuccess": {
			reason: "Should not return an error if the internal update succeeds",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e: managed.ExternalObservation{
					ResourceLateInitialized: true,
					ResourceExists:          true,
					ResourceUpToDate:        true,
				},
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				cr := newManagedZone(withLabels(managedZoneLabels))
				mz := &dns.ManagedZone{}
				mzclient.GenerateManagedZone(meta.GetExternalName(cr), cr.Spec.ForProvider, mz)
				if err := json.NewEncoder(w).Encode(mz); err != nil {
					t.Error(err)
				}
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"ResourceNotUpToDate": {
			reason: "Should return upToDate as false if the resource is not up to date",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := dns.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := managedZoneExternal{
				kube:      tc.kube,
				projectID: managedZoneProjectID,
				dns:       s.ManagedZones,
			}
			got, err := e.Observe(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.e, got); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestManagedZoneCreate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		e   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason  string
		handler http.Handler
		args    args
		want    want
	}{
		"NotManagedZone": {
			reason: "Should return an error if the resource is not ManagedZone",
			args: args{
				mg: nonManagedZone,
			},
			want: want{
				e:   managed.ExternalCreation{},
				err: errors.New(errNotManagedZone),
			},
		},
		"Successful": {
			reason: "Should succeed if the resource creation doesn't return an error",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e:   managed.ExternalCreation{},
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"Failed": {
			reason: "Should fail if the resource creation returns an error",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e:   managed.ExternalCreation{},
				err: errors.Wrap(managedZoneGError(http.StatusBadRequest, ""), errCreateManagedZone),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := dns.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := managedZoneExternal{
				projectID: managedZoneProjectID,
				dns:       s.ManagedZones,
			}
			got, err := e.Create(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.e, got); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestManagedZoneUpdate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		e   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason  string
		handler http.Handler
		args    args
		want    want
	}{
		"NotManagedZone": {
			reason: "Should return an error if the resource is not ManagedZone",
			args: args{
				mg: nonManagedZone,
			},
			want: want{
				e:   managed.ExternalUpdate{},
				err: errors.New(errNotManagedZone),
			},
		},
		"Successful": {
			reason: "Should succeed if the resource update doesn't return an error",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e:   managed.ExternalUpdate{},
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"Failed": {
			reason: "Should fail if the resource update returns an error",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				e:   managed.ExternalUpdate{},
				err: managedZoneGError(http.StatusBadRequest, ""),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := dns.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := managedZoneExternal{
				projectID: managedZoneProjectID,
				dns:       s.ManagedZones,
			}
			got, err := e.Update(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.e, got); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestManagedZoneDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		reason  string
		handler http.Handler
		args    args
		want    want
	}{
		"NotManagedZone": {
			reason: "Should return an error if the resource is not ManagedZone",
			args: args{
				mg: nonManagedZone,
			},
			want: want{
				err: errors.New(errNotManagedZone),
			},
		},
		"Successful": {
			reason: "Should succeed if the resource update doesn't return an error",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"Failed": {
			reason: "Should fail if the resource update returns an error",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				err: errors.Wrap(managedZoneGError(http.StatusBadRequest, ""), errDeleteManagedZone),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"NotFound": {
			reason: "Should not return an error if the resource is not found",
			args: args{
				mg: newManagedZone(),
			},
			want: want{
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				if err := json.NewEncoder(w).Encode(&dns.ManagedZone{}); err != nil {
					t.Error(err)
				}
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := dns.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := managedZoneExternal{
				projectID: managedZoneProjectID,
				dns:       s.ManagedZones,
			}
			err := e.Delete(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
