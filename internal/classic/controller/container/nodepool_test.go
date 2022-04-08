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

package container

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crossplane/provider-gcp/apis/classic/container/v1beta1"

	np "github.com/crossplane/provider-gcp/internal/classic/clients/nodepool"

	"github.com/google/go-cmp/cmp"
	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

type nodePoolModifier func(*v1beta1.NodePool)

func npWithConditions(c ...xpv1.Condition) nodePoolModifier {
	return func(i *v1beta1.NodePool) { i.Status.SetConditions(c...) }
}

func npWithProviderStatus(s string) nodePoolModifier {
	return func(i *v1beta1.NodePool) { i.Status.AtProvider.Status = s }
}

func npWithLocations(l []string) nodePoolModifier {
	return func(i *v1beta1.NodePool) { i.Spec.ForProvider.Locations = l }
}

func nodePool(im ...nodePoolModifier) *v1beta1.NodePool {
	i := &v1beta1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: name,
			},
		},
		Spec: v1beta1.NodePoolSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderReference: &xpv1.Reference{Name: providerName},
			},
			ForProvider: v1beta1.NodePoolParameters{},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

func TestNodePoolObserve(t *testing.T) {
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
				_ = json.NewEncoder(w).Encode(&container.NodePool{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(),
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
				_ = json.NewEncoder(w).Encode(&container.NodePool{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetNodePool),
			},
		},
		"NotUpToDateSpecUpdateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				n := nodePool()
				gn := &container.NodePool{}
				np.GenerateNodePool(name, n.Spec.ForProvider, gn)
				gn.Locations = []string{"loc-1"}
				_ = json.NewEncoder(w).Encode(gn)
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(errBoom),
			},
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(npWithLocations([]string{"loc-1"})),
				err: errors.Wrap(errBoom, errManagedNodePoolUpdateFailed),
			},
		},
		"Creating": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				n := &container.NodePool{}
				np.GenerateNodePool(name, nodePool().Spec.ForProvider, n)
				n.Status = v1beta1.NodePoolStateProvisioning
				_ = json.NewEncoder(w).Encode(n)
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: nodePool(npWithProviderStatus(v1beta1.NodePoolStateProvisioning), npWithConditions(xpv1.Creating())),
			},
		},
		"Unavailable": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				c := &container.NodePool{}
				np.GenerateNodePool(name, nodePool().Spec.ForProvider, c)
				c.Status = v1beta1.NodePoolStateError
				_ = json.NewEncoder(w).Encode(c)
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: nodePool(npWithProviderStatus(v1beta1.NodePoolStateError), npWithConditions(xpv1.Unavailable())),
			},
		},
		"RunnableUnbound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				c := &container.NodePool{}
				np.GenerateNodePool(name, nodePool().Spec.ForProvider, c)
				c.Status = v1beta1.NodePoolStateRunning
				_ = json.NewEncoder(w).Encode(c)
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: nodePool(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: nodePool(
					npWithProviderStatus(v1beta1.NodePoolStateRunning),
					npWithConditions(xpv1.Available())),
			},
		},
		"BoundUnavailable": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				n := &container.NodePool{}
				np.GenerateNodePool(name, nodePool().Spec.ForProvider, n)
				n.Status = v1beta1.NodePoolStateError
				_ = json.NewEncoder(w).Encode(n)
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: nodePool(
					npWithProviderStatus(v1beta1.NodePoolStateRunning),
					npWithConditions(xpv1.Available()),
				),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: nodePool(
					npWithProviderStatus(v1beta1.NodePoolStateError),
					npWithConditions(xpv1.Unavailable())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := container.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := nodePoolExternal{
				kube:      tc.kube,
				projectID: projectID,
				container: s,
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

func TestNodePoolCreate(t *testing.T) {
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
				i := &container.NodePool{}
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
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(npWithConditions(xpv1.Creating())),
				cre: managed.ExternalCreation{},
				err: nil,
			},
		},
		"SuccessfulSkipCreate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				i := &container.NodePool{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// Return bad request for post to demonstrate that
				// http call is never made.
				w.WriteHeader(http.StatusBadRequest)
				_ = r.Body.Close()
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(npWithProviderStatus(v1beta1.NodePoolStateProvisioning)),
			},
			want: want{
				mg: nodePool(
					npWithConditions(xpv1.Creating()),
					npWithProviderStatus(v1beta1.NodePoolStateProvisioning),
				),
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
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(npWithConditions(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusConflict, ""), errCreateNodePool),
			},
		},
		"Failed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(npWithConditions(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateNodePool),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := container.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := nodePoolExternal{
				kube:      tc.kube,
				projectID: projectID,
				container: s,
			}
			_, err := e.Create(tc.args.ctx, tc.args.mg)
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
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNodePoolDelete(t *testing.T) {
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
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(npWithConditions(xpv1.Deleting())),
				err: nil,
			},
		},
		"SuccessfulSkipDelete": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// Return bad request for delete to demonstrate that
				// http call is never made.
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(npWithProviderStatus(v1beta1.NodePoolStateStopping)),
			},
			want: want{
				mg: nodePool(
					npWithConditions(xpv1.Deleting()),
					npWithProviderStatus(v1beta1.NodePoolStateStopping),
				),
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
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(npWithConditions(xpv1.Deleting())),
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
				_ = json.NewEncoder(w).Encode(&container.Operation{})
			}),
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(npWithConditions(xpv1.Deleting())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errDeleteNodePool),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := container.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := nodePoolExternal{
				kube:      tc.kube,
				projectID: projectID,
				container: s,
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

func TestNodePoolUpdate(t *testing.T) {
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
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&container.NodePool{})
				case http.MethodPut:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: nodePool(npWithLocations([]string{"loc-1"})),
			},
			want: want{
				mg:  nodePool(npWithLocations([]string{"loc-1"})),
				err: nil,
			},
		},
		"SuccessfulSkipWhileReconciling": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					// Return bad request for get to demonstrate that
					// http call is never made.
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.NodePool{})
				case http.MethodPut:
					// Return bad request for put to demonstrate that
					// http call is never made.
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: nodePool(
					npWithLocations([]string{"loc-1"}),
					npWithProviderStatus(v1beta1.NodePoolStateReconciling),
				),
			},
			want: want{
				mg: nodePool(
					npWithLocations([]string{"loc-1"}),
					npWithProviderStatus(v1beta1.NodePoolStateReconciling),
				),
				err: nil,
			},
		},
		"SuccessfulSkipWhileProvisioning": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					// Return bad request for get to demonstrate that
					// http call is never made.
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.NodePool{})
				case http.MethodPut:
					// Return bad request for put to demonstrate that
					// http call is never made.
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: nodePool(
					npWithLocations([]string{"loc-1"}),
					npWithProviderStatus(v1beta1.NodePoolStateProvisioning),
				),
			},
			want: want{
				mg: nodePool(
					npWithLocations([]string{"loc-1"}),
					npWithProviderStatus(v1beta1.NodePoolStateProvisioning),
				),
				err: nil,
			},
		},
		"SuccessfulNoopUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&container.NodePool{
						Name: name,
					})
				case http.MethodPut:
					// Return bad request for update to demonstrate that
					// underlying update is not making any http call.
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(),
				err: nil,
			},
		},
		"GetFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.NodePool{})
				case http.MethodPut:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				default:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				// No need to actually require an update. We won't get that far.
				mg: nodePool(),
			},
			want: want{
				mg:  nodePool(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetNodePool),
			},
		},
		"UpdateFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					// Must return successful get of node pool that does not match spec.
					_ = json.NewEncoder(w).Encode(&container.NodePool{})
				case http.MethodPut:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				default:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&container.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				// Must include field that causes update.
				mg: nodePool(npWithLocations([]string{"loc-1"})),
			},
			want: want{
				mg:  nodePool(npWithLocations([]string{"loc-1"})),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateNodePool),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := container.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := nodePoolExternal{
				kube:      tc.kube,
				projectID: projectID,
				container: s,
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
