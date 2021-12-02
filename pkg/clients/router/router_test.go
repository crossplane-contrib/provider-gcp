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

package router

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"

	"github.com/crossplane/provider-gcp/apis/compute/v1alpha1"
)

const (
	testName              = "some-name"
	testCreationTimestamp = "10/10/2023"
	testSelfLink          = "/link/to/self"
	testRegion            = "us-west1"
)

var (
	testNetwork             = "test-network"
	testAsn           int64 = 65550
	testRoutePriority int64 = 1000
	testDescription         = "some desc"
	testAdvertiseMode       = "DEFAULT"
)

func params(m ...func(*v1alpha1.RouterParameters)) *v1alpha1.RouterParameters {
	o := &v1alpha1.RouterParameters{
		Description: &testDescription,
		Network:     &testNetwork,
		Region:      testRegion,
		Bgp: &v1alpha1.RouterBgp{
			AdvertiseMode: &testAdvertiseMode,
			Asn:           &testAsn,
		},
		BgpPeers: []*v1alpha1.RouterBgpPeer{
			{
				AdvertiseMode:           &testAdvertiseMode,
				AdvertisedRoutePriority: &testRoutePriority,
				Name:                    "test-bgp-peer",
				PeerAsn:                 65551,
			},
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func router(m ...func(*compute.Router)) *compute.Router {
	o := &compute.Router{
		Description: testDescription,
		Network:     testNetwork,
		Bgp: &compute.RouterBgp{
			AdvertiseMode: testAdvertiseMode,
			Asn:           testAsn,
		},
		BgpPeers: []*compute.RouterBgpPeer{
			{
				AdvertiseMode:           testAdvertiseMode,
				AdvertisedRoutePriority: testRoutePriority,
				Name:                    "test-bgp-peer",
				PeerAsn:                 65551,
			},
		},
		Name: testName,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func addOutputFields(n *compute.Router) {
	n.CreationTimestamp = testCreationTimestamp
	n.Id = 2029819203
	n.SelfLink = testSelfLink
}

func observation(m ...func(*v1alpha1.RouterObservation)) *v1alpha1.RouterObservation {
	o := &v1alpha1.RouterObservation{
		CreationTimestamp: testCreationTimestamp,
		ID:                2029819203,
		SelfLink:          testSelfLink,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestGenerateRouter(t *testing.T) {
	type args struct {
		name string
		in   v1alpha1.RouterParameters
	}
	cases := map[string]struct {
		args args
		want *compute.Router
	}{
		"BgpAsnNil": {
			args: args{
				name: testName,
				in: *params(func(p *v1alpha1.RouterParameters) {
					p.Bgp.Asn = nil
				}),
			},
			want: router(func(n *compute.Router) {
				n.Bgp.Asn = 0
			}),
		},
		"SpecifyBgpAsn": {
			args: args{
				name: testName,
				in: *params(func(p *v1alpha1.RouterParameters) {
					p.Bgp.Asn = &testAsn
				}),
			},
			want: router(func(n *compute.Router) {
				n.Bgp.Asn = testAsn
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &compute.Router{}
			GenerateRouter(tc.args.name, tc.args.in, r)
			if diff := cmp.Diff(r, tc.want); diff != "" {
				t.Errorf("GenerateRouter(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRouterObservation(t *testing.T) {
	cases := map[string]struct {
		in  compute.Router
		out v1alpha1.RouterObservation
	}{
		"AllFilled": {
			in:  *router(addOutputFields),
			out: *observation(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRouterObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateRouterObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec *v1alpha1.RouterParameters
		in   compute.Router
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.RouterParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: params(),
				in:   *router(),
			},
			want: params(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: params(),
				in: *router(func(n *compute.Router) {
					n.Description = "some other description"
				}),
			},
			want: params(),
		},
		"PartialFilled": {
			args: args{
				spec: params(func(p *v1alpha1.RouterParameters) {
					p.Bgp = nil
				}),
				in: *router(),
			},
			want: params(func(p *v1alpha1.RouterParameters) {
				p.Bgp = &v1alpha1.RouterBgp{
					AdvertiseMode: &testAdvertiseMode,
					Asn:           &testAsn,
				}
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.spec, tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		in      *v1alpha1.RouterParameters
		current *compute.Router
	}
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"UpToDate": {
			args: args{
				in:      params(),
				current: router(),
			},
			want: want{upToDate: true, isErr: false},
		},
		"UpToDateWithOutputFields": {
			args: args{
				in:      params(),
				current: router(addOutputFields),
			},
			want: want{upToDate: true, isErr: false},
		},
		"NotUpToDate": {
			args: args{
				in: params(func(p *v1alpha1.RouterParameters) {
					p.Description = nil
				}),
				current: router(),
			},
			want: want{upToDate: false, isErr: false},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			u, err := IsUpToDate(testName, tc.args.in, tc.args.current)
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, u); diff != "" {
				t.Errorf("IsUpToDate(...) UpToDate: -want, +got:\n%s", diff)
			}
		})
	}
}
