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

package storage

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/storage/v1alpha3"
)

type MockBucketClient struct {
	h BucketHandler
}

func (c *MockBucketClient) Bucket(name string) BucketHandler {
	return c.h
}

type MockBucketHandler struct {
	MockAttrs  func(context.Context) (*storage.BucketAttrs, error)
	MockCreate func(context.Context, string, *storage.BucketAttrs) error
	MockUpdate func(context.Context, storage.BucketAttrsToUpdate) (*storage.BucketAttrs, error)
	MockDelete func(context.Context) error
}

func (m *MockBucketHandler) Attrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return m.MockAttrs(ctx)
}

func (m *MockBucketHandler) Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) error {
	return m.MockCreate(ctx, projectID, attrs)
}

func (m *MockBucketHandler) Update(ctx context.Context, attrs storage.BucketAttrsToUpdate) (*storage.BucketAttrs, error) {
	return m.MockUpdate(ctx, attrs)
}

func (m *MockBucketHandler) Delete(ctx context.Context) error {
	return m.MockDelete(ctx)
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		handle    BucketClient
		projectID string
		client    client.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"NotABucket": {
			reason: "We should return an error if the supplied managed resource is not a bucket",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotBucket),
			},
		},
		"BucketNotFound": {
			reason: "We should report a non-existent bucket via our observation",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockAttrs: func(context.Context) (*storage.BucketAttrs, error) { return nil, storage.ErrBucketNotExist },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"AttrsError": {
			reason: "Errors updating a bucket should be returned",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockAttrs: func(context.Context) (*storage.BucketAttrs, error) { return nil, errBoom },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				err: errors.Wrap(errBoom, errAttrs),
			},
		},
		"UpdateError": {
			reason: "Observing a bucket successfully should return an ExternalObservation and nil error",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockAttrs: func(context.Context) (*storage.BucketAttrs, error) {
						return &storage.BucketAttrs{
							// This should trigger a 'late-init' because the
							// associated spec field is the empty string.
							Location: "over-there",
						}, nil
					},
				}},
				client: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				err: errors.Wrap(errBoom, errLateInit),
			},
		},
		"Success": {
			reason: "Observing a bucket successfully should return an ExternalObservation and nil error",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockAttrs: func(context.Context) (*storage.BucketAttrs, error) { return &storage.BucketAttrs{}, nil },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				o:   managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{handle: tc.fields.handle, projectID: tc.fields.projectID, client: tc.fields.client}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		handle    BucketClient
		projectID string
		client    client.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		c   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"NotABucket": {
			reason: "We should return an error if the supplied managed resource is not a bucket",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotBucket),
			},
		},
		"CreateError": {
			reason: "Errors creating a bucket should be returned",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockCreate: func(context.Context, string, *storage.BucketAttrs) error { return errBoom },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				err: errors.Wrap(errBoom, errCreate),
			},
		},
		"Success": {
			reason: "Creating a bucket successfully should return an empty ExternalCreation and nil error",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockCreate: func(context.Context, string, *storage.BucketAttrs) error { return nil },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{handle: tc.fields.handle, projectID: tc.fields.projectID, client: tc.fields.client}
			got, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.c, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		handle    BucketClient
		projectID string
		client    client.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		u   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"NotABucket": {
			reason: "We should return an error if the supplied managed resource is not a bucket",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotBucket),
			},
		},
		"AttrsError": {
			reason: "Errors updating a bucket should be returned",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockAttrs: func(context.Context) (*storage.BucketAttrs, error) { return nil, errBoom },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				err: errors.Wrap(errBoom, errAttrs),
			},
		},
		"UpdateError": {
			reason: "Errors updating a bucket should be returned",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockAttrs:  func(context.Context) (*storage.BucketAttrs, error) { return &storage.BucketAttrs{}, nil },
					MockUpdate: func(context.Context, storage.BucketAttrsToUpdate) (*storage.BucketAttrs, error) { return nil, errBoom },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
		"Success": {
			reason: "Updating a bucket successfully should return an empty ExternalUpdate and nil error",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockAttrs:  func(context.Context) (*storage.BucketAttrs, error) { return &storage.BucketAttrs{}, nil },
					MockUpdate: func(context.Context, storage.BucketAttrsToUpdate) (*storage.BucketAttrs, error) { return nil, nil },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: want{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{handle: tc.fields.handle, projectID: tc.fields.projectID, client: tc.fields.client}
			got, err := e.Update(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.u, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		handle    BucketClient
		projectID string
		client    client.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   error
	}{
		"NotABucket": {
			reason: "We should return an error if the supplied managed resource is not a bucket",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotBucket),
		},
		"DeleteError": {
			reason: "Errors deleting a bucket should be returned",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockDelete: func(context.Context) error { return errBoom },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: errors.Wrap(errBoom, errDelete),
		},
		"Success": {
			reason: "Deleting a bucket successfully should return a nil error",
			fields: fields{
				handle: &MockBucketClient{&MockBucketHandler{
					MockDelete: func(context.Context) error { return nil },
				}},
			},
			args: args{
				mg: &v1alpha3.Bucket{},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{handle: tc.fields.handle, projectID: tc.fields.projectID, client: tc.fields.client}
			err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
