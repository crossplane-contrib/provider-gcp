/*
Copyright 2020 The Crossplane Authors.

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

package pubsub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/pubsub/v1alpha1"
)

const (
	projectID = "fooproject"
)

var (
	errBoom = errors.New("foo")
)

type TopicOption func(*v1alpha1.Topic)

func newTopic(opts ...TopicOption) *v1alpha1.Topic {
	t := &v1alpha1.Topic{}

	for _, f := range opts {
		f(t)
	}
	return t
}

func gError(code int, message string) *googleapi.Error {
	return &googleapi.Error{
		Code:    code,
		Body:    "",
		Message: message,
	}
}

func TestObserve(t *testing.T) {
	type args struct {
		handler http.Handler
		kube    client.Client
		mg      resource.Managed
	}

	type want struct {
		eo  managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"GetFailed": {
			reason: "Should return error if GetTopic fails",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusBadRequest)
				}),
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetTopic),
			},
		},
		"NotFound": {
			reason: "Should not return error if Topic is not found",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusNotFound)
				}),
				mg: newTopic(),
			},
		},
		"SpecUpdateFailed": {
			reason: "Should fail if spec Update failed",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&pubsub.Topic{
						KmsKeyName: "cool-key",
					})
				}),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(errBoom, errKubeUpdateTopic),
			},
		},
		"Success": {
			reason: "Should succeed",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&pubsub.Topic{})
				}),
				mg: newTopic(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: managed.ConnectionDetails{
						v1alpha1.ConnectionSecretKeyTopic:       []byte(""),
						v1alpha1.ConnectionSecretKeyProjectName: []byte(projectID),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.args.handler)
			defer server.Close()
			s, _ := pubsub.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				client:    tc.args.kube,
				projectID: projectID,
				ps:        s,
			}
			got, err := e.Observe(context.Background(), tc.args.mg)
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
		handler http.Handler
		kube    client.Client
		mg      resource.Managed
	}

	type want struct {
		eo  managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"CreateFailed": {
			reason: "Should return error if GetTopic fails",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusBadRequest)
				}),
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateTopic),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&pubsub.Topic{
						KmsKeyName: "cool-key",
					})
				}),
				mg: newTopic(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.args.handler)
			defer server.Close()
			s, _ := pubsub.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				client:    tc.args.kube,
				projectID: projectID,
				ps:        s,
			}
			got, err := e.Create(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.eo, got); diff != "" {
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
		handler http.Handler
		kube    client.Client
		mg      resource.Managed
	}

	type want struct {
		eo  managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"GetFailed": {
			reason: "Should return error if GetTopic fails",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusBadRequest)
				}),
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetTopic),
			},
		},
		"UpdateFailed": {
			reason: "Should return error if UpdateTopic fails",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if r.Method == http.MethodPatch {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&pubsub.Topic{
						KmsKeyName: "cool-key",
					})
				}),
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateTopic),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&pubsub.Topic{
						KmsKeyName: "cool-key",
					})
				}),
				mg: newTopic(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.args.handler)
			defer server.Close()
			s, _ := pubsub.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				client:    tc.args.kube,
				projectID: projectID,
				ps:        s,
			}
			got, err := e.Update(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.eo, got); diff != "" {
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
		ctx     context.Context
		handler http.Handler
		kube    client.Client
		mg      resource.Managed
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"DeleteFailed": {
			reason: "Should return error if DeleteTopic fails",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusBadRequest)
				}),
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errDeleteTopic),
			},
		},
		"NotFound": {
			reason: "Should not return error if resource is already gone",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusNotFound)
				}),
				mg: newTopic(),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = r.Body.Close()
					if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&pubsub.Topic{
						Name: "cool-name",
					})
				}),
				mg: newTopic(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.args.handler)
			defer server.Close()
			s, _ := pubsub.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := external{
				client:    tc.args.kube,
				projectID: projectID,
				ps:        s,
			}
			err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
