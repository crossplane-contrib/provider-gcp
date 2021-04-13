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
	"testing"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/crossplane/provider-gcp/pkg/clients/topic"

	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go"
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

type MockPublisherClient struct {
	MockCreateTopic func(ctx context.Context, req *pubsubpb.Topic, opts ...gax.CallOption) (*pubsubpb.Topic, error)
	MockUpdateTopic func(ctx context.Context, req *pubsubpb.UpdateTopicRequest, opts ...gax.CallOption) (*pubsubpb.Topic, error)
	MockGetTopic    func(ctx context.Context, req *pubsubpb.GetTopicRequest, opts ...gax.CallOption) (*pubsubpb.Topic, error)
	MockDeleteTopic func(ctx context.Context, req *pubsubpb.DeleteTopicRequest, opts ...gax.CallOption) error
}

func (m *MockPublisherClient) CreateTopic(ctx context.Context, req *pubsubpb.Topic, opts ...gax.CallOption) (*pubsubpb.Topic, error) {
	return m.MockCreateTopic(ctx, req, opts...)
}
func (m *MockPublisherClient) UpdateTopic(ctx context.Context, req *pubsubpb.UpdateTopicRequest, opts ...gax.CallOption) (*pubsubpb.Topic, error) {
	return m.MockUpdateTopic(ctx, req, opts...)
}
func (m *MockPublisherClient) GetTopic(ctx context.Context, req *pubsubpb.GetTopicRequest, opts ...gax.CallOption) (*pubsubpb.Topic, error) {
	return m.MockGetTopic(ctx, req, opts...)
}
func (m *MockPublisherClient) DeleteTopic(ctx context.Context, req *pubsubpb.DeleteTopicRequest, opts ...gax.CallOption) error {
	return m.MockDeleteTopic(ctx, req, opts...)
}

type TopicOption func(*v1alpha1.Topic)

func newTopic(opts ...TopicOption) *v1alpha1.Topic {
	t := &v1alpha1.Topic{}

	for _, f := range opts {
		f(t)
	}
	return t
}

func TestObserve(t *testing.T) {
	type args struct {
		kube client.Client
		ps   topic.PublisherClient
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
				ps: &MockPublisherClient{
					MockGetTopic: func(_ context.Context, _ *pubsubpb.GetTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return nil, errBoom
					},
				},
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetTopic),
			},
		},
		"NotFound": {
			reason: "Should not return error if Topic is not found",
			args: args{
				ps: &MockPublisherClient{
					MockGetTopic: func(_ context.Context, _ *pubsubpb.GetTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return nil, status.Error(codes.NotFound, "olala")
					},
				},
				mg: newTopic(),
			},
		},
		"SpecUpdateFailed": {
			reason: "Should fail if spec Update failed",
			args: args{
				ps: &MockPublisherClient{
					MockGetTopic: func(_ context.Context, _ *pubsubpb.GetTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return &pubsubpb.Topic{KmsKeyName: "olala"}, nil
					},
				},
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
				ps: &MockPublisherClient{
					MockGetTopic: func(_ context.Context, _ *pubsubpb.GetTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return &pubsubpb.Topic{}, nil
					},
				},
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
			e := &external{client: tc.args.kube, ps: tc.args.ps, projectID: projectID}
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
		kube client.Client
		ps   topic.PublisherClient
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
			reason: "Should return error if GetTopic fails",
			args: args{
				ps: &MockPublisherClient{
					MockCreateTopic: func(_ context.Context, _ *pubsubpb.Topic, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return nil, errBoom
					},
				},
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateTopic),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				ps: &MockPublisherClient{
					MockCreateTopic: func(_ context.Context, _ *pubsubpb.Topic, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return &pubsubpb.Topic{}, nil
					},
				},
				mg: newTopic(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.kube, ps: tc.args.ps, projectID: projectID}
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
		ps   topic.PublisherClient
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
			reason: "Should return error if GetTopic fails",
			args: args{
				ps: &MockPublisherClient{
					MockGetTopic: func(_ context.Context, _ *pubsubpb.GetTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return nil, errBoom
					},
				},
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetTopic),
			},
		},
		"UpdateFailed": {
			reason: "Should return error if UpdateTopic fails",
			args: args{
				ps: &MockPublisherClient{
					MockGetTopic: func(_ context.Context, _ *pubsubpb.GetTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return &pubsubpb.Topic{}, nil
					},
					MockUpdateTopic: func(_ context.Context, _ *pubsubpb.UpdateTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return nil, errBoom
					},
				},
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateTopic),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				ps: &MockPublisherClient{
					MockGetTopic: func(_ context.Context, _ *pubsubpb.GetTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return &pubsubpb.Topic{}, nil
					},
					MockUpdateTopic: func(_ context.Context, _ *pubsubpb.UpdateTopicRequest, _ ...gax.CallOption) (*pubsubpb.Topic, error) {
						return nil, nil
					},
				},
				mg: newTopic(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.kube, ps: tc.args.ps, projectID: projectID}
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
		ps   topic.PublisherClient
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
			reason: "Should return error if DeleteTopic fails",
			args: args{
				ps: &MockPublisherClient{
					MockDeleteTopic: func(_ context.Context, _ *pubsubpb.DeleteTopicRequest, _ ...gax.CallOption) error {
						return errBoom
					},
				},
				mg: newTopic(),
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteTopic),
			},
		},
		"NotFound": {
			reason: "Should not return error if resource is already gone",
			args: args{
				ps: &MockPublisherClient{
					MockDeleteTopic: func(_ context.Context, _ *pubsubpb.DeleteTopicRequest, _ ...gax.CallOption) error {
						return status.Error(codes.NotFound, "olala")
					},
				},
				mg: newTopic(),
			},
		},
		"Success": {
			reason: "Should not fail if all calls succeed",
			args: args{
				ps: &MockPublisherClient{
					MockDeleteTopic: func(_ context.Context, _ *pubsubpb.DeleteTopicRequest, _ ...gax.CallOption) error {
						return nil
					},
				},
				mg: newTopic(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.kube, ps: tc.args.ps, projectID: projectID}
			err := e.Delete(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
