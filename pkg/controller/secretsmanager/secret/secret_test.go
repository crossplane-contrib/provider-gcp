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

package secret

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/secretsmanager/secret/v1alpha1"
	"github.com/crossplane/provider-gcp/pkg/clients/secretsmanager/secret"
)

const (
	projectID = "fooproject"
	name      = "bar"
)

var (
	errBoom = errors.New("foo")
)

type MockSecretClient struct {
	MockCreateSecret func(ctx context.Context, req *secretmanager.CreateSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error)
	MockUpdateSecret func(ctx context.Context, req *secretmanager.UpdateSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error)
	MockGetSecret    func(ctx context.Context, req *secretmanager.GetSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error)
	MockDeleteSecret func(ctx context.Context, req *secretmanager.DeleteSecretRequest, opts ...gax.CallOption) error
}

func (m *MockSecretClient) CreateSecret(ctx context.Context, req *secretmanager.CreateSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error) {
	return m.MockCreateSecret(ctx, req, opts...)
}
func (m *MockSecretClient) UpdateSecret(ctx context.Context, req *secretmanager.UpdateSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error) {
	return m.MockUpdateSecret(ctx, req, opts...)
}
func (m *MockSecretClient) GetSecret(ctx context.Context, req *secretmanager.GetSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error) {
	return m.MockGetSecret(ctx, req, opts...)
}
func (m *MockSecretClient) DeleteSecret(ctx context.Context, req *secretmanager.DeleteSecretRequest, opts ...gax.CallOption) error {
	return m.MockDeleteSecret(ctx, req, opts...)
}

type SecretOption func(*v1alpha1.Secret)

func newSecret(opts ...SecretOption) *v1alpha1.Secret {
	t := &v1alpha1.Secret{
		Spec: v1alpha1.SecretSpec{
			ForProvider: v1alpha1.SecretParameters{
				Parent: fmt.Sprintf("projects/%s", projectID),
			},
		},
	}

	for _, f := range opts {
		f(t)
	}
	return t
}

func TestCreate(t *testing.T) {

	type args struct {
		kube client.Client
		sc   secret.Client
		mg   resource.Managed
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
		"GetFailed": {
			reason: "Should return error if GetSecret fails",
			args: args{
				sc: &MockSecretClient{
					MockCreateSecret: func(_ context.Context, _ *secretmanager.CreateSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return nil, errBoom
					},
				},
				mg: newSecret(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateSecret),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				sc: &MockSecretClient{
					MockCreateSecret: func(_ context.Context, _ *secretmanager.CreateSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return &secretmanager.Secret{}, nil
					},
				},
				mg: newSecret(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.kube, sc: tc.args.sc, projectID: projectID}
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
		kube client.Client
		sc   secret.Client
		mg   resource.Managed
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
			reason: "Should return error if GetSecret fails",
			args: args{
				sc: &MockSecretClient{
					MockGetSecret: func(_ context.Context, _ *secretmanager.GetSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return nil, errBoom
					},
				},
				mg: newSecret(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetSecret),
			},
		},
		"UpdateFailed": {
			reason: "Should return error if UpdateSecret fails",
			args: args{
				sc: &MockSecretClient{
					MockGetSecret: func(_ context.Context, _ *secretmanager.GetSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return &secretmanager.Secret{
							Name: fmt.Sprintf("projects/%s/secrets/%s", projectID, name),
						}, nil
					},
					MockUpdateSecret: func(_ context.Context, _ *secretmanager.UpdateSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return nil, errBoom
					},
				},
				mg: newSecret(),
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateSecret),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				sc: &MockSecretClient{
					MockGetSecret: func(_ context.Context, _ *secretmanager.GetSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return &secretmanager.Secret{
							Name: fmt.Sprintf("projects/%s/secrets/%s", projectID, name),
						}, nil
					},
					MockUpdateSecret: func(_ context.Context, _ *secretmanager.UpdateSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return nil, nil
					},
				},
				mg: newSecret(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.kube, sc: tc.args.sc, projectID: projectID}
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
		kube client.Client
		sc   secret.Client
		mg   resource.Managed
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
			reason: "Should return error if DeleteSecret fails",
			args: args{
				sc: &MockSecretClient{
					MockDeleteSecret: func(_ context.Context, _ *secretmanager.DeleteSecretRequest, _ ...gax.CallOption) error {
						return errBoom
					},
				},
				mg: newSecret(),
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteSecret),
			},
		},
		"NotFound": {
			reason: "Should not return error if resource is already gone",
			args: args{
				sc: &MockSecretClient{
					MockDeleteSecret: func(_ context.Context, _ *secretmanager.DeleteSecretRequest, _ ...gax.CallOption) error {
						return status.Error(codes.NotFound, "Error")
					},
				},
				mg: newSecret(),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				sc: &MockSecretClient{
					MockDeleteSecret: func(_ context.Context, _ *secretmanager.DeleteSecretRequest, _ ...gax.CallOption) error {
						return nil
					},
				},
				mg: newSecret(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.kube, sc: tc.args.sc, projectID: projectID}
			err := e.Delete(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type args struct {
		kube client.Client
		sc   secret.Client
		mg   resource.Managed
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
				sc: &MockSecretClient{
					MockGetSecret: func(_ context.Context, _ *secretmanager.GetSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return nil, errBoom
					},
				},
				mg: newSecret(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetSecret),
			},
		},
		"NotFound": {
			reason: "Should not return error if Topic is not found",
			args: args{
				sc: &MockSecretClient{
					MockGetSecret: func(_ context.Context, _ *secretmanager.GetSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return nil, status.Error(codes.NotFound, "Error")
					},
				},
				mg: newSecret(),
			},
		},
		"Success": {
			reason: "Should succeed",
			args: args{
				sc: &MockSecretClient{
					MockGetSecret: func(_ context.Context, _ *secretmanager.GetSecretRequest, _ ...gax.CallOption) (*secretmanager.Secret, error) {
						return &secretmanager.Secret{
							Name: fmt.Sprintf("projects/%s/secrets/%s", projectID, name),
						}, nil
					},
				},
				mg: newSecret(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: managed.ConnectionDetails{
						v1alpha1.ConnectionSecretKeyName:        []byte(""),
						v1alpha1.ConnectionSecretKeyProjectName: []byte(projectID),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.kube, sc: tc.args.sc, projectID: projectID}
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
