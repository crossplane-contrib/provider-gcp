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

package v1beta1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	computev1alpha3 "github.com/crossplaneio/stack-gcp/apis/compute/v1alpha3"
	"github.com/crossplaneio/stack-gcp/pkg/clients/connection"
)

var _ resource.AttributeReferencer = (*NetworkURIReferencerForCloudSQLInstance)(nil)

func TestNetworkURIReferencerForCloudSQLInstanceGetStatus(t *testing.T) {
	name := "coolNet"
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		res resource.CanReference
		c   client.Reader
	}
	type want struct {
		statuses []resource.ReferenceStatus
		err      error
	}

	cases := map[string]struct {
		reason string
		ref    corev1.LocalObjectReference
		args   args
		want   want
	}{
		"GetNetworkError": {
			reason: "Errors getting the Network should be wrapped and returned.",
			ref:    corev1.LocalObjectReference{Name: name},
			args: args{
				c: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errGetNetwork),
			},
		},
		"NotFound": {
			reason: "A non-existent Network should be considered not found.",
			ref:    corev1.LocalObjectReference{Name: name},
			args: args{
				c: &test.MockClient{
					MockGet: test.NewMockGetFn(kerrors.NewNotFound(schema.GroupResource{}, "")),
				},
			},
			want: want{
				statuses: []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceNotFound}},
			},
		},
		"NotReady": {
			reason: "A Network not in condition Ready should not be considered ready.",
			ref:    corev1.LocalObjectReference{Name: name},
			args: args{
				c: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						n := &computev1alpha3.Network{
							Status: computev1alpha3.NetworkStatus{
								GCPNetworkStatus: computev1alpha3.GCPNetworkStatus{
									Peerings: []*computev1alpha3.GCPNetworkPeering{
										{Name: connection.PeeringName, State: connection.PeeringStateActive},
									},
								},
							},
						}
						n.SetConditions(runtimev1alpha1.Unavailable())

						*(obj.(*computev1alpha3.Network)) = *n
						return nil
					}),
				},
			},
			want: want{
				statuses: []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceNotReady}},
			},
		},
		"NoServiceNetworkingPeering": {
			reason: "A Network in condition Ready but with no service networking peering should not be considered ready.",
			ref:    corev1.LocalObjectReference{Name: name},
			args: args{
				c: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						n := &computev1alpha3.Network{}
						n.SetConditions(runtimev1alpha1.Available())

						*(obj.(*computev1alpha3.Network)) = *n
						return nil
					}),
				},
			},
			want: want{
				statuses: []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceNotReady}},
			},
		},
		"Successful": {
			reason: "A Network in condition Ready with a service networking peering should be considered ready.",
			ref:    corev1.LocalObjectReference{Name: name},
			args: args{
				c: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						n := &computev1alpha3.Network{
							Status: computev1alpha3.NetworkStatus{
								GCPNetworkStatus: computev1alpha3.GCPNetworkStatus{
									Peerings: []*computev1alpha3.GCPNetworkPeering{
										{Name: connection.PeeringName, State: connection.PeeringStateActive},
									},
								},
							},
						}
						n.SetConditions(runtimev1alpha1.Available())

						*(obj.(*computev1alpha3.Network)) = *n
						return nil
					}),
				},
			},
			want: want{
				statuses: []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceReady}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &NetworkURIReferencerForCloudSQLInstance{LocalObjectReference: tc.ref}
			statuses, err := r.GetStatus(tc.args.ctx, tc.args.res, tc.args.c)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\n-want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.statuses, statuses, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\n-want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestNetworkURIReferencerForCloudSQLInstanceBuild(t *testing.T) {
	value := "definitely/a/network"
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		res resource.CanReference
		c   client.Reader
	}
	type want struct {
		value string
		err   error
	}
	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"GetNetworkError": {
			reason: "Errors getting the Network should be wrapped and returned.",
			args: args{
				c: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errGetNetwork),
			},
		},
		"Successful": {
			reason: "Referencing a Network should return its SelfLink with the compute URI prefix stripped.",
			args: args{
				c: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						*(obj.(*computev1alpha3.Network)) = computev1alpha3.Network{
							Status: computev1alpha3.NetworkStatus{
								GCPNetworkStatus: computev1alpha3.GCPNetworkStatus{
									SelfLink: computev1alpha3.URIPrefix + value,
								},
							},
						}
						return nil
					}),
				},
			},
			want: want{
				value: value,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &NetworkURIReferencerForCloudSQLInstance{}
			value, err := r.Build(tc.args.ctx, tc.args.res, tc.args.c)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\n-want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.value, value, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\n-want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestNetworkURIReferencerForCloudSQLInstanceAssign(t *testing.T) {
	value := "definitely/a/network"

	type args struct {
		res   resource.CanReference
		value string
	}
	type want struct {
		res resource.CanReference
		err error
	}
	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NotCloudSQLInstanceError": {
			reason: "A non-existent Network should be considered not found.",
			args:   args{},
			want: want{
				err: errors.New(errResourceIsNotCloudSQLInstance),
			},
		},
		"Successful": {
			reason: "The value should be assigned to the PrivateNetwork field, even if IPConfiguration is nil.",
			args: args{
				res:   &CloudSQLInstance{},
				value: value,
			},
			want: want{
				res: &CloudSQLInstance{
					Spec: CloudSQLInstanceSpec{ForProvider: CloudSQLInstanceParameters{
						Settings: Settings{IPConfiguration: &IPConfiguration{PrivateNetwork: &value}},
					}},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &NetworkURIReferencerForCloudSQLInstance{}
			err := r.Assign(tc.args.res, tc.args.value)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\n-want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.res, tc.args.res, test.EquateConditions()); diff != "" {
				t.Errorf("\nReason: %s\n-want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
