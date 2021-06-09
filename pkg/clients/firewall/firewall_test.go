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

package firewall

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
)

const (
	testName              = "some-name"
	testCreationTimestamp = "10/10/2023"
	testSelfLink          = "/link/to/self"
)

var (
	testNetwork           = "test-network"
	testPriority    int64 = 9000
	testDirection         = "INGRESS"
	trueVal               = true
	falseVal              = false
	testDescription       = "some desc"
)

func params(m ...func(*v1beta1.FirewallParameters)) *v1beta1.FirewallParameters {
	o := &v1beta1.FirewallParameters{
		Description:  &testDescription,
		Network:      &testNetwork,
		Priority:     &testPriority,
		SourceRanges: []string{"10.0.0.0/24"},
		Direction:    &testDirection,
		Disabled:     &trueVal,
		Allowed: []*v1beta1.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"80", "443"},
			},
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func firewall(m ...func(*compute.Firewall)) *compute.Firewall {
	o := &compute.Firewall{
		Description:  testDescription,
		Network:      testNetwork,
		Priority:     testPriority,
		SourceRanges: []string{"10.0.0.0/24"},
		Direction:    testDirection,
		Disabled:     trueVal,
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"80", "443"},
			},
		},
		Name: testName,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func addOutputFields(n *compute.Firewall) {
	n.CreationTimestamp = testCreationTimestamp
	n.Id = 2029819203
	n.SelfLink = testSelfLink
}

func observation(m ...func(*v1beta1.FirewallObservation)) *v1beta1.FirewallObservation {
	o := &v1beta1.FirewallObservation{
		CreationTimestamp: testCreationTimestamp,
		ID:                2029819203,
		SelfLink:          testSelfLink,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestGenerateFirewall(t *testing.T) {
	type args struct {
		name string
		in   v1beta1.FirewallParameters
	}
	cases := map[string]struct {
		args args
		want *compute.Firewall
	}{
		"DisabledNil": {
			args: args{
				name: testName,
				in: *params(func(p *v1beta1.FirewallParameters) {
					p.Disabled = nil
				}),
			},
			want: firewall(func(n *compute.Firewall) {
				n.Disabled = false
			}),
		},
		"DisabledFalse": {
			args: args{
				name: testName,
				in: *params(func(p *v1beta1.FirewallParameters) {
					p.Disabled = &falseVal
				}),
			},
			want: firewall(func(n *compute.Firewall) {
				n.Disabled = false
			}),
		},
		"DisabledTrue": {
			args: args{
				name: testName,
				in: *params(func(p *v1beta1.FirewallParameters) {
					p.Disabled = &trueVal
				}),
			},
			want: firewall(func(n *compute.Firewall) {
				n.Disabled = true
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &compute.Firewall{}
			GenerateFirewall(tc.args.name, tc.args.in, r)
			if diff := cmp.Diff(r, tc.want); diff != "" {
				t.Errorf("GenerateFirewall(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateFirewallObservation(t *testing.T) {
	cases := map[string]struct {
		in  compute.Firewall
		out v1beta1.FirewallObservation
	}{
		"AllFilled": {
			in:  *firewall(addOutputFields),
			out: *observation(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateFirewallObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateFirewallObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec *v1beta1.FirewallParameters
		in   compute.Firewall
	}
	cases := map[string]struct {
		args args
		want *v1beta1.FirewallParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: params(),
				in:   *firewall(),
			},
			want: params(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: params(),
				in: *firewall(func(n *compute.Firewall) {
					n.Description = "some other description"
				}),
			},
			want: params(),
		},
		"PartialFilled": {
			args: args{
				spec: params(func(p *v1beta1.FirewallParameters) {
					p.Direction = nil
				}),
				in: *firewall(),
			},
			want: params(func(p *v1beta1.FirewallParameters) {
				p.Direction = &testDirection
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
		in      *v1beta1.FirewallParameters
		current *compute.Firewall
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
				current: firewall(),
			},
			want: want{upToDate: true, isErr: false},
		},
		"UpToDateWithOutputFields": {
			args: args{
				in:      params(),
				current: firewall(addOutputFields),
			},
			want: want{upToDate: true, isErr: false},
		},
		"NotUpToDate": {
			args: args{
				in: params(func(p *v1beta1.FirewallParameters) {
					p.Description = nil
				}),
				current: firewall(),
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
