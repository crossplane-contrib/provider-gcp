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

package network

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
)

const (
	testName              = "some-name"
	testRoutingMode       = "GLOBAL"
	testCreationTimestamp = "10/10/2023"
	testGatewayIPv4       = "10.0.0.0"
	testSelfLink          = "/link/to/self"

	testPeeringName         = "some-peering-name"
	testPeeringNetwork      = "name"
	testPeeringState        = "ACTIVE"
	testPeeringStateDetails = "more-detailed than ACTIVE"
)

var (
	trueVal         = true
	falseVal        = false
	testDescription = "some desc"
)

func params(m ...func(*v1beta1.NetworkParameters)) *v1beta1.NetworkParameters {
	o := &v1beta1.NetworkParameters{
		AutoCreateSubnetworks: &trueVal,
		Description:           &testDescription,
		RoutingConfig: &v1beta1.NetworkRoutingConfig{
			RoutingMode: testRoutingMode,
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func network(m ...func(*compute.Network)) *compute.Network {
	o := &compute.Network{
		AutoCreateSubnetworks: true,
		Description:           testDescription,
		Name:                  testName,
		RoutingConfig: &compute.NetworkRoutingConfig{
			RoutingMode: testRoutingMode,
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func addOutputFields(n *compute.Network) {
	n.CreationTimestamp = testCreationTimestamp
	n.GatewayIPv4 = testGatewayIPv4
	n.Id = 2029819203
	n.Peerings = []*compute.NetworkPeering{
		{
			AutoCreateRoutes:     true,
			ExchangeSubnetRoutes: true,
			Name:                 testPeeringName,
			Network:              testPeeringNetwork,
			State:                testPeeringState,
			StateDetails:         testPeeringStateDetails,
		},
	}
	n.SelfLink = testSelfLink
	n.Subnetworks = []string{
		"my-subnetwork",
	}
}

func observation(m ...func(*v1beta1.NetworkObservation)) *v1beta1.NetworkObservation {
	o := &v1beta1.NetworkObservation{
		CreationTimestamp: testCreationTimestamp,
		GatewayIPv4:       testGatewayIPv4,
		ID:                2029819203,
		Peerings: []*v1beta1.NetworkPeering{
			{
				AutoCreateRoutes:     true,
				ExchangeSubnetRoutes: true,
				Name:                 testPeeringName,
				Network:              testPeeringNetwork,
				State:                testPeeringState,
				StateDetails:         testPeeringStateDetails,
			},
		},
		SelfLink: testSelfLink,
		Subnetworks: []string{
			"my-subnetwork",
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestGenerateNetwork(t *testing.T) {
	type args struct {
		name string
		in   v1beta1.NetworkParameters
	}
	cases := map[string]struct {
		args args
		want *compute.Network
	}{
		"AutoCreateSubnetworksNil": {
			args: args{
				name: testName,
				in: *params(func(p *v1beta1.NetworkParameters) {
					p.AutoCreateSubnetworks = nil
				}),
			},
			want: network(func(n *compute.Network) {
				n.AutoCreateSubnetworks = false
			}),
		},
		"AutoCreateSubnetworksFalse": {
			args: args{
				name: testName,
				in: *params(func(p *v1beta1.NetworkParameters) {
					p.AutoCreateSubnetworks = &falseVal
				}),
			},
			want: network(func(n *compute.Network) {
				n.AutoCreateSubnetworks = false
				n.ForceSendFields = []string{"AutoCreateSubnetworks"}
			}),
		},
		"AutoCreateSubnetworksTrue": {
			args: args{
				name: testName,
				in:   *params(),
			},
			want: network(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &compute.Network{}
			GenerateNetwork(tc.args.name, tc.args.in, r)
			if diff := cmp.Diff(r, tc.want); diff != "" {
				t.Errorf("GenerateNetwork(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateNetworkObservation(t *testing.T) {
	cases := map[string]struct {
		in  compute.Network
		out v1beta1.NetworkObservation
	}{
		"AllFilled": {
			in:  *network(addOutputFields),
			out: *observation(),
		},
		"NoPeerings": {
			in: *network(addOutputFields, func(n *compute.Network) {
				n.Peerings = []*compute.NetworkPeering{}
			}),
			out: *observation(func(o *v1beta1.NetworkObservation) {
				o.Peerings = nil
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateNetworkObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec *v1beta1.NetworkParameters
		in   compute.Network
	}
	cases := map[string]struct {
		args args
		want *v1beta1.NetworkParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: params(),
				in:   *network(),
			},
			want: params(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: params(),
				in: *network(func(n *compute.Network) {
					n.Description = "some other description"
				}),
			},
			want: params(),
		},
		"PartialFilled": {
			args: args{
				spec: params(func(p *v1beta1.NetworkParameters) {
					p.RoutingConfig = nil
				}),
				in: *network(),
			},
			want: params(func(p *v1beta1.NetworkParameters) {
				p.RoutingConfig = &v1beta1.NetworkRoutingConfig{
					RoutingMode: testRoutingMode,
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
		in      *v1beta1.NetworkParameters
		current *compute.Network
	}
	type want struct {
		upToDate     bool
		switchCustom bool
		isErr        bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"UpToDate": {
			args: args{
				in:      params(),
				current: network(),
			},
			want: want{upToDate: true, switchCustom: false, isErr: false},
		},
		"UpToDateWithOutputFields": {
			args: args{
				in:      params(),
				current: network(addOutputFields),
			},
			want: want{upToDate: true, switchCustom: false, isErr: false},
		},
		"NotUpToDate": {
			args: args{
				in: params(func(p *v1beta1.NetworkParameters) {
					p.Description = nil
				}),
				current: network(),
			},
			want: want{upToDate: false, switchCustom: false, isErr: false},
		},
		"NotUpToDateSwitchToCustom": {
			args: args{
				in: params(func(p *v1beta1.NetworkParameters) {
					p.AutoCreateSubnetworks = &falseVal
				}),
				current: network(func(n *compute.Network) {
					n.AutoCreateSubnetworks = true
				}),
			},
			want: want{upToDate: false, switchCustom: true, isErr: false},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			u, s, err := IsUpToDate(testName, tc.args.in, tc.args.current)
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, u); diff != "" {
				t.Errorf("IsUpToDate(...) UpToDate: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.switchCustom, s); diff != "" {
				t.Errorf("IsUpToDate(...) SwitchToCustoms: -want, +got:\n%s", diff)
			}
		})
	}
}
