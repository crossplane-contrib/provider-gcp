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

package secretversion

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/secretsmanager/secretversion/v1alpha1"
	"github.com/crossplane/provider-gcp/pkg/clients/secretsmanager/secretversion"
)

const (
	projectID = "fooproject"
	name      = "bar"
)

var (
	errBoom = errors.New("foo")
)

type MockSecretVersionClient struct {
	MockAddSecretVersion     func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	MockAccessSecretVersion  func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	MockEnableSecretVersion  func(ctx context.Context, req *secretmanager.EnableSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	MockDisableSecretVersion func(ctx context.Context, req *secretmanager.DisableSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	MockGetSecretVersion     func(ctx context.Context, req *secretmanager.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	MockDestroySecretVersion func(ctx context.Context, req *secretmanager.DestroySecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
}

func (m *MockSecretVersionClient) AddSecretVersion(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return m.MockAddSecretVersion(ctx, req, opts...)
}

func (m *MockSecretVersionClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return m.MockAccessSecretVersion(ctx, req, opts...)
}

func (m *MockSecretVersionClient) EnableSecretVersion(ctx context.Context, req *secretmanager.EnableSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return m.MockEnableSecretVersion(ctx, req, opts...)
}

func (m *MockSecretVersionClient) DisableSecretVersion(ctx context.Context, req *secretmanager.DisableSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return m.MockDisableSecretVersion(ctx, req, opts...)
}

func (m *MockSecretVersionClient) GetSecretVersion(ctx context.Context, req *secretmanager.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanager.SecretVersion, error) {
	return m.MockGetSecretVersion(ctx, req, opts...)
}
func (m *MockSecretVersionClient) DestroySecretVersion(ctx context.Context, req *secretmanager.DestroySecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return m.MockDestroySecretVersion(ctx, req, opts...)
}

type SecretVersionOption func(*v1alpha1.SecretVersion)

func newSecretVersion(opts ...SecretVersionOption) *v1alpha1.SecretVersion {
	t := &v1alpha1.SecretVersion{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"crossplane.io/external-name": "1",
			},
		},
		Spec: v1alpha1.SecretVersionSpec{
			ForProvider: v1alpha1.SecretVersionParameters{
				SecretRef: name,
				Payload: &v1alpha1.SecretPayload{
					Data: "",
				},
				DesiredSecretVersionState: "ENABLED",
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
		sc   secretversion.Client
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
				sc: &MockSecretVersionClient{
					MockAddSecretVersion: func(_ context.Context, _ *secretmanager.AddSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, errBoom
					},
				},
				mg: newSecretVersion(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateSecretVersion),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				sc: &MockSecretVersionClient{
					MockAddSecretVersion: func(_ context.Context, _ *secretmanager.AddSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return &secretmanager.SecretVersion{}, nil
					},
				},
				mg: newSecretVersion(),
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

func TestDelete(t *testing.T) {
	type args struct {
		kube client.Client
		sc   secretversion.Client
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
				sc: &MockSecretVersionClient{
					MockDestroySecretVersion: func(_ context.Context, _ *secretmanager.DestroySecretVersionRequest, _ ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
						return nil, errBoom
					},
				},
				mg: newSecretVersion(),
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteSecretVersion),
			},
		},
		"NotFound": {
			reason: "Should not return error if resource is already gone",
			args: args{
				sc: &MockSecretVersionClient{
					MockDestroySecretVersion: func(_ context.Context, _ *secretmanager.DestroySecretVersionRequest, _ ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
						return nil, status.Error(codes.NotFound, "Error")
					},
				},
				mg: newSecretVersion(),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				sc: &MockSecretVersionClient{
					MockDestroySecretVersion: func(_ context.Context, _ *secretmanager.DestroySecretVersionRequest, _ ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
						return &secretmanagerpb.SecretVersion{}, nil
					},
				},
				mg: newSecretVersion(),
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

func TestUpdate(t *testing.T) {

	disabledSecretVersion := newSecretVersion()
	disabledSecretVersion.Spec.ForProvider.DesiredSecretVersionState = "DISABLED"

	destroyedSecretVersion := newSecretVersion()
	destroyedSecretVersion.Spec.ForProvider.DesiredSecretVersionState = "DESTROYED"
	type args struct {
		kube client.Client
		sc   secretversion.Client
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
			reason: "Should return error if GetSecretVersion fails",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, errBoom
					},
				},
				mg: newSecretVersion(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetSecretVersion),
			},
		},
		"UpdateFailed": {
			reason: "Should return error if UpdateTopic fails",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return &secretmanagerpb.SecretVersion{}, nil
					},
					MockEnableSecretVersion: func(_ context.Context, _ *secretmanager.EnableSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, errBoom
					},
				},
				mg: newSecretVersion(),
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateSecretVersion),
			},
		},
		"Enabled": {
			reason: "Should not fail if Enable call doesn't fail",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return &secretmanager.SecretVersion{}, nil
					},
					MockEnableSecretVersion: func(_ context.Context, _ *secretmanager.EnableSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, nil
					},
				},
				mg: newSecretVersion(),
			},
		},

		"Disabled": {
			reason: "Should not fail if Disable call doesn't fail",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return &secretmanager.SecretVersion{}, nil
					},
					MockDisableSecretVersion: func(_ context.Context, _ *secretmanager.DisableSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, nil
					},
				},
				mg: disabledSecretVersion,
			},
		},

		"Destroyed": {
			reason: "Should not fail if Destroy call doesn't fail",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return &secretmanager.SecretVersion{}, nil
					},
					MockDestroySecretVersion: func(_ context.Context, _ *secretmanager.DestroySecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, nil
					},
				},
				mg: destroyedSecretVersion,
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

func TestObserve(t *testing.T) {
	type args struct {
		kube client.Client
		sc   secretversion.Client
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
			reason: "Should return error if GetSecretVersion fails",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, errBoom
					},
				},
				mg: newSecretVersion(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetSecretVersion),
			},
		},
		"NotFound": {
			reason: "Should not return error if SecretVersion is not found",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return nil, status.Error(codes.NotFound, "Error")
					},
				},
				mg: newSecretVersion(),
			},
		},
		"Success": {
			reason: "Should succeed if AccessVersion call succeeds",
			args: args{
				sc: &MockSecretVersionClient{
					MockGetSecretVersion: func(_ context.Context, _ *secretmanager.GetSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.SecretVersion, error) {
						return &secretmanagerpb.SecretVersion{
							Name:  fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, name, "1"),
							State: 1,
						}, nil
					},
					MockAccessSecretVersion: func(_ context.Context, _ *secretmanager.AccessSecretVersionRequest, _ ...gax.CallOption) (*secretmanager.AccessSecretVersionResponse, error) {
						return &secretmanagerpb.AccessSecretVersionResponse{
							Payload: &secretmanagerpb.SecretPayload{
								Data: []byte(""),
							},
						}, nil
					},
				},
				mg: newSecretVersion(),
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
