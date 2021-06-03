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

package cache

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	redis "google.golang.org/api/redis/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/cache/v1beta1"
	"github.com/crossplane/provider-gcp/pkg/clients/cloudmemorystore"
)

const (
	namespace     = "cool-namespace"
	region        = "us-cool1"
	project       = "coolProject"
	instanceName  = "claimns-claimname-8sdh3"
	qualifiedName = "projects/" + project + "/locations/" + region + "/instances/" + instanceName
	memorySizeGB  = 1
	host          = "172.16.0.1"
	port          = 6379

	connectionSecretName = "cool-connection-secret"
)

var (
	authorizedNetwork = "default"
	connectMode       = "DIRECT_PEERING"
	redisConfigs      = map[string]string{"cool": "socool"}
)

func gError(code int, message string) *googleapi.Error {
	return &googleapi.Error{
		Code:    code,
		Body:    "",
		Message: message,
	}
}

type strange struct {
	resource.Managed
}

type instanceModifier func(*v1beta1.CloudMemorystoreInstance)

func withConditions(c ...xpv1.Condition) instanceModifier {
	return func(i *v1beta1.CloudMemorystoreInstance) { i.Status.SetConditions(c...) }
}

func withState(s string) instanceModifier {
	return func(i *v1beta1.CloudMemorystoreInstance) { i.Status.AtProvider.State = s }
}

func withFullName(name string) instanceModifier {
	return func(i *v1beta1.CloudMemorystoreInstance) { i.Status.AtProvider.Name = name }
}

func withHost(e string) instanceModifier {
	return func(i *v1beta1.CloudMemorystoreInstance) { i.Status.AtProvider.Host = e }
}

func withPort(p int) instanceModifier {
	return func(i *v1beta1.CloudMemorystoreInstance) { i.Status.AtProvider.Port = int64(p) }
}

func instance(im ...instanceModifier) *v1beta1.CloudMemorystoreInstance {
	i := &v1beta1.CloudMemorystoreInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       instanceName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: instanceName,
			},
		},
		Spec: v1beta1.CloudMemorystoreInstanceSpec{
			ResourceSpec: xpv1.ResourceSpec{
				WriteConnectionSecretToReference: &xpv1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
				MemorySizeGB:      memorySizeGB,
				RedisConfigs:      redisConfigs,
				AuthorizedNetwork: &authorizedNetwork,
				ConnectMode:       &connectMode,
			},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connecter{}

func TestObserve(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg          resource.Managed
		observation managed.ExternalObservation
		err         error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"ObservedInstanceAvailable": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&redis.Instance{
					State: cloudmemorystore.StateReady,
					Host:  host,
					Port:  port,
					Name:  qualifiedName,
				})
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg: instance(
					withConditions(xpv1.Available()),
					withState(cloudmemorystore.StateReady),
					withHost(host),
					withPort(port),
					withFullName(qualifiedName)),
				observation: managed.ExternalObservation{
					ResourceExists: true,
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretEndpointKey: []byte(host),
						xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
					},
				},
			},
		},
		"ObservedInstanceCreating": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&redis.Instance{
					State: cloudmemorystore.StateCreating,
					Name:  qualifiedName,
				})
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg: instance(
					withConditions(xpv1.Creating()),
					withState(cloudmemorystore.StateCreating),
					withFullName(qualifiedName)),
				observation: managed.ExternalObservation{
					ResourceExists:    true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"ObservedInstanceDeleting": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&redis.Instance{
					State: cloudmemorystore.StateDeleting,
					Name:  qualifiedName,
				})
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg: instance(
					withConditions(xpv1.Deleting()),
					withState(cloudmemorystore.StateDeleting),
					withFullName(qualifiedName)),
				observation: managed.ExternalObservation{
					ResourceExists:    true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"ObservedInstanceDoesNotExist": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg:          instance(),
				observation: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NotCloudMemorystoreInstance": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotInstance),
			},
		},
		"FailedToGetInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg:  instance(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetInstance),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := redis.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				kube:      tc.kube,
				projectID: "cool-project",
				cms:       s,
			}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.observation, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Observe(): -want, +got:\n%s", diff)
			}
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

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("resource.Managed: -want, +got:\n%s", diff)
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
		mg       resource.Managed
		creation managed.ExternalCreation
		err      error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"CreatedInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&redis.Instance{
					Name: qualifiedName,
				})
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg: instance(withConditions(xpv1.Creating())),
			},
		},
		"NotCloudMemorystoreInstance": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotInstance),
			},
		},
		"FailedToCreateInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg:  instance(withConditions(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateInstance),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := redis.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				kube:      tc.kube,
				projectID: "cool-project",
				cms:       s,
			}
			got, err := e.Create(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.creation, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Create(): -want, +got:\n%s", diff)
			}

			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Create(...): want error string != got error string:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Create(...): want error != got error:\n%s", diff)
				}
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("resource.Managed: -want, +got:\n%s", diff)
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
		mg     resource.Managed
		update managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"UpdatedInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&redis.Instance{
					Name: qualifiedName,
				})
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg: instance(withConditions()),
			},
		},
		"NotCloudMemorystoreInstance": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotInstance),
			},
		},
		"FailedToUpdateInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg:  instance(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateInstance),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := redis.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				kube:      tc.kube,
				projectID: "cool-project",
				cms:       s,
			}
			got, err := e.Update(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.update, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Update(): -want, +got:\n%s", diff)
			}

			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Update(...): want error string != got error string:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Update(...): want error != got error:\n%s", diff)
				}
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("resource.Managed: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestDelete(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
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
		"DeletedInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&redis.Instance{
					Name: qualifiedName,
				})
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg: instance(withConditions(xpv1.Deleting())),
			},
		},
		"NotCloudMemorystoreInstance": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotInstance),
			},
		},
		"FailedToDeleteInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
			}),
			args: args{
				ctx: context.Background(),
				mg:  instance(),
			},
			want: want{
				mg:  instance(withConditions(xpv1.Deleting())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errDeleteInstance),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := redis.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				kube:      tc.kube,
				projectID: "cool-project",
				cms:       s,
			}
			err := e.Delete(tc.args.ctx, tc.args.mg)

			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Delete(...): want error string != got error string:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Delete(...): want error != got error:\n%s", diff)
				}
			}
		})
	}
}
