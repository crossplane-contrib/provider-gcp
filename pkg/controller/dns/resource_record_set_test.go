/*
Copyright 2021 The Crossplane Authors.

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
	rrsClient "github.com/crossplane-contrib/provider-gcp/pkg/clients/dns"
)

const (
	projectID = "myproject-id-1234"
)

var (
	unexpectedObject resource.Managed
	errBoom          = errors.New("boom")
)

type rrsOption func(*v1alpha1.ResourceRecordSet)

func newRrs(opts ...rrsOption) *v1alpha1.ResourceRecordSet {
	rrs := &v1alpha1.ResourceRecordSet{}

	for _, f := range opts {
		f(rrs)
	}

	return rrs
}

func withSignature(s string) rrsOption {
	return func(r *v1alpha1.ResourceRecordSet) {
		r.Spec.ForProvider.SignatureRRDatas = []string{s}
	}
}

func gError(code int, message string) *googleapi.Error {
	return &googleapi.Error{
		Code:    code,
		Body:    "{}\n",
		Message: message,
	}
}

func TestObserve(t *testing.T) {
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
		"NotResourceRecordSet": {
			reason: "Should return an error if the resource is not ResourceRecordSet",
			args: args{
				mg: unexpectedObject,
			},
			want: want{
				e:   managed.ExternalObservation{},
				err: errors.New(errNotResourceRecordSet),
			},
		},
		"ResourceNotFound": {
			reason: "Should not return an error if the API response is 404",
			args: args{
				mg: newRrs(),
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
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"InternalError": {
			reason: "Should return an error if the error is different than 404",
			args: args{
				mg: newRrs(),
			},
			want: want{
				e:   managed.ExternalObservation{},
				err: errors.Wrap(gError(http.StatusInternalServerError, ""), errGetFailed),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"UpdateResourceSpecFail": {
			reason: "Should return an error if the internal update fails",
			args: args{
				mg: newRrs(),
			},
			want: want{
				e:   managed.ExternalObservation{},
				err: errors.Wrap(errBoom, errManagedUpdateFailed),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				cr := newRrs(withSignature("test"))
				rrs := &dns.ResourceRecordSet{}
				rrsClient.GenerateResourceRecordSet(meta.GetExternalName(cr), cr.Spec.ForProvider, rrs)
				if err := json.NewEncoder(w).Encode(rrs); err != nil {
					t.Error(err)
				}
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(errBoom),
			},
		},
		"UpdateResourceSpecSuccess": {
			reason: "Should not return an error if the internal update succeeds",
			args: args{
				mg: newRrs(),
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
				cr := newRrs(withSignature("test"))
				rrs := &dns.ResourceRecordSet{}
				rrsClient.GenerateResourceRecordSet(meta.GetExternalName(cr), cr.Spec.ForProvider, rrs)
				if err := json.NewEncoder(w).Encode(rrs); err != nil {
					t.Error(err)
				}
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"ResourceUpToDate": {
			reason: "Should return upToDate as true if the resource is up to date",
			args: args{
				newRrs(),
			},
			want: want{
				e: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				err: nil,
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{Kind: "dns#resourceRecordSet"}); err != nil {
					t.Error(err)
				}
			}),
		},
		"ResourceNotUpToDate": {
			reason: "Should return upToDate as false if the resource is not up to date",
			args: args{
				newRrs(withSignature("test")),
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
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
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
			e := external{
				kube:      tc.kube,
				projectID: projectID,
				dns:       s.ResourceRecordSets,
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

func TestCreate(t *testing.T) {
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
		"NotResourceRecordSet": {
			reason: "Should return an error if the resource is not ResourceRecordSet",
			args: args{
				mg: unexpectedObject,
			},
			want: want{
				e:   managed.ExternalCreation{},
				err: errors.New(errNotResourceRecordSet),
			},
		},
		"Successful": {
			reason: "Should succeed if the resource creation doesn't return an error",
			args: args{
				mg: newRrs(),
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
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"Failed": {
			reason: "Should fail if the resource creation returns an error",
			args: args{
				mg: newRrs(),
			},
			want: want{
				e:   managed.ExternalCreation{},
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateCluster),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
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
			e := external{
				projectID: projectID,
				dns:       s.ResourceRecordSets,
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

func TestUpdate(t *testing.T) {
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
		"NotResourceRecordSet": {
			reason: "Should return an error if the resource is not ResourceRecordSet",
			args: args{
				mg: unexpectedObject,
			},
			want: want{
				e:   managed.ExternalUpdate{},
				err: errors.New(errNotResourceRecordSet),
			},
		},
		"Successful": {
			reason: "Should succeed if the resource update doesn't return an error",
			args: args{
				mg: newRrs(),
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
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"Failed": {
			reason: "Should fail if the resource update returns an error",
			args: args{
				mg: newRrs(),
			},
			want: want{
				e:   managed.ExternalUpdate{},
				err: gError(http.StatusBadRequest, ""),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
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
			e := external{
				projectID: projectID,
				dns:       s.ResourceRecordSets,
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

func TestDelete(t *testing.T) {
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
		"NotResourceRecordSet": {
			reason: "Should return an error if the resource is not ResourceRecordSet",
			args: args{
				mg: unexpectedObject,
			},
			want: want{
				err: errors.New(errNotResourceRecordSet),
			},
		},
		"Successful": {
			reason: "Should succeed if the resource update doesn't return an error",
			args: args{
				mg: newRrs(),
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
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"Failed": {
			reason: "Should fail if the resource update returns an error",
			args: args{
				mg: newRrs(),
			},
			want: want{
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCannotDelete),
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
					t.Error(err)
				}
			}),
		},
		"NotFound": {
			reason: "Should not return an error if the resource is not found",
			args: args{
				mg: newRrs(),
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
				if err := json.NewEncoder(w).Encode(&dns.ResourceRecordSet{}); err != nil {
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
			e := external{
				projectID: projectID,
				dns:       s.ResourceRecordSets,
			}
			err := e.Delete(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
