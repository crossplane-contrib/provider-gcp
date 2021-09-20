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

package compute

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	"github.com/crossplane/provider-gcp/pkg/clients/subnetwork"
)

const (
	testSubnetworkName = "test-subnetwork"
)

var _ managed.ExternalConnecter = &subnetworkConnector{}
var _ managed.ExternalClient = &subnetworkExternal{}

type subnetworkModifier func(*v1beta1.Subnetwork)

func subnetworkWithConditions(c ...xpv1.Condition) subnetworkModifier {
	return func(i *v1beta1.Subnetwork) { i.Status.SetConditions(c...) }
}

func subnetworkWithDescription(d string) subnetworkModifier {
	return func(i *v1beta1.Subnetwork) { i.Spec.ForProvider.Description = &d }
}

func subnetworkWithPrivateAccess(p bool) subnetworkModifier {
	return func(i *v1beta1.Subnetwork) { i.Spec.ForProvider.PrivateIPGoogleAccess = &p }
}

func subnetworkObj(im ...subnetworkModifier) *v1beta1.Subnetwork {
	i := &v1beta1.Subnetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testSubnetworkName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: testSubnetworkName,
			},
		},
		Spec: v1beta1.SubnetworkSpec{
			ForProvider: v1beta1.SubnetworkParameters{},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

func TestSubnetworkObserve(t *testing.T) {
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
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"NotFound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(),
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
				_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetSubnetwork),
			},
		},
		"NotUpToDateSpecUpdateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				c := subnetworkObj()
				gn := &compute.Subnetwork{}
				subnetwork.GenerateSubnetwork(testSubnetworkName, c.Spec.ForProvider, gn)
				gn.Description = "a very interesting description"
				_ = json.NewEncoder(w).Encode(gn)
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(errBoom),
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithDescription("a very interesting description")),
				err: errors.Wrap(errBoom, errManagedSubnetworkUpdate),
			},
		},
		"RunnableUnbound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				c := &compute.Subnetwork{}
				subnetwork.GenerateSubnetwork(testSubnetworkName, subnetworkObj().Spec.ForProvider, c)
				_ = json.NewEncoder(w).Encode(c)
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: subnetworkObj(subnetworkWithConditions(xpv1.Available())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
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

func TestSubnetworkCreate(t *testing.T) {
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
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				i := &compute.Subnetwork{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = r.Body.Close()
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(xpv1.Creating())),
				cre: managed.ExternalCreation{},
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
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusConflict, ""), errCreateSubnetworkFailed),
			},
		},
		"Failed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateSubnetworkFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			_, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSubnetworkDelete(t *testing.T) {
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
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(xpv1.Deleting())),
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
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(xpv1.Deleting())),
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
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(xpv1.Deleting())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errDeleteSubnetworkFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			err := e.Delete(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSubnetworkUpdate(t *testing.T) {
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
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"GetExternalFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetSubnetwork),
			},
		},
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{
						Description: "not the one I want",
					})
				case http.MethodPatch:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(subnetworkWithDescription("a new description")),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithDescription("a new description")),
				err: nil,
			},
		},
		"SuccessfulPrivateAccess": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{
						PrivateIpGoogleAccess: false,
					})
				case http.MethodPost:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(subnetworkWithPrivateAccess(true)),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithPrivateAccess(true)),
				err: nil,
			},
		},
		"UpdateGeneralFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
				case http.MethodPatch:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				// Must include field that causes update.
				mg: subnetworkObj(subnetworkWithDescription("a new description")),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithDescription("a new description")),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateSubnetworkFailed),
			},
		},
		"UpdatePrivateAccessFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{
						PrivateIpGoogleAccess: false,
					})
				case http.MethodPost:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(subnetworkWithPrivateAccess(true)),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithPrivateAccess(true)),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateSubnetworkPAFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			upd, err := e.Update(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}

		})
	}
}
