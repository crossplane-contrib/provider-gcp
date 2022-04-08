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

package subnetwork

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/compute/v1"

	v1beta12 "github.com/crossplane/provider-gcp/apis/classic/compute/v1beta1"
	gcp "github.com/crossplane/provider-gcp/internal/classic/clients"
)

const (
	testName              = "some-name"
	testIPCIDRRange       = "10.0.0.0/9"
	testRegion            = "test-region"
	testSelfLink          = "/link/to/self"
	testFingerprint       = "averycoolfingerprinthash"
	testCreationTimestamp = "10/10/2023"
	testGatewayAddress    = "10.0.0.0"
)

var equateSecondaryRange = func(i, j *v1beta12.SubnetworkSecondaryRange) bool { return i.RangeName > j.RangeName }

var (
	trueVal         = true
	testDescription = "some desc"
	testNetwork     = "test-network"
)

func params(m ...func(*v1beta12.SubnetworkParameters)) *v1beta12.SubnetworkParameters {
	o := &v1beta12.SubnetworkParameters{
		Description:           &testDescription,
		EnableFlowLogs:        &trueVal,
		IPCidrRange:           testIPCIDRRange,
		Network:               &testNetwork,
		PrivateIPGoogleAccess: &trueVal,
		Region:                testRegion,
		SecondaryIPRanges: []*v1beta12.SubnetworkSecondaryRange{
			{
				RangeName:   "zzaa",
				IPCidrRange: "10.1.0.0/9",
			},
			{
				RangeName:   "aazz",
				IPCidrRange: "10.0.2.1/9",
			},
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func subnetwork(m ...func(*compute.Subnetwork)) *compute.Subnetwork {
	o := &compute.Subnetwork{
		Description:           testDescription,
		Name:                  testName,
		EnableFlowLogs:        trueVal,
		IpCidrRange:           testIPCIDRRange,
		Network:               v1beta12.ComputeURIPrefix + testNetwork,
		PrivateIpGoogleAccess: trueVal,
		Region:                v1beta12.ComputeURIPrefix + testRegion,
		SecondaryIpRanges: []*compute.SubnetworkSecondaryRange{
			{
				RangeName:   "aazz",
				IpCidrRange: "10.0.2.1/9",
			},
			{
				RangeName:   "zzaa",
				IpCidrRange: "10.1.0.0/9",
			},
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func addOutputFields(n *compute.Subnetwork) {
	n.CreationTimestamp = testCreationTimestamp
	n.GatewayAddress = testGatewayAddress
	n.Fingerprint = testFingerprint
	n.Id = 12345678
	n.SelfLink = testSelfLink
}

func observation(m ...func(*v1beta12.SubnetworkObservation)) *v1beta12.SubnetworkObservation {
	o := &v1beta12.SubnetworkObservation{
		CreationTimestamp: testCreationTimestamp,
		GatewayAddress:    testGatewayAddress,
		Fingerprint:       testFingerprint,
		ID:                12345678,
		SelfLink:          testSelfLink,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestGenerateSubnetwork(t *testing.T) {
	type args struct {
		name string
		in   v1beta12.SubnetworkParameters
	}
	cases := map[string]struct {
		args args
		want *compute.Subnetwork
	}{
		"FilledGeneration": {
			args: args{
				name: testName,
				in:   *params(),
			},
			want: subnetwork(),
		},
		"NoSecondary": {
			args: args{
				name: testName,
				in: *params(func(p *v1beta12.SubnetworkParameters) {
					p.SecondaryIPRanges = nil
				}),
			},
			want: subnetwork(func(s *compute.Subnetwork) {
				s.SecondaryIpRanges = nil
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &compute.Subnetwork{}
			GenerateSubnetwork(tc.args.name, tc.args.in, r)
			if diff := cmp.Diff(tc.want, r, equateSecondaryRanges(), gcp.EquateComputeURLs()); diff != "" {
				t.Errorf("GenerateSubnetwork(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateSubnetworkObservation(t *testing.T) {
	cases := map[string]struct {
		in  compute.Subnetwork
		out v1beta12.SubnetworkObservation
	}{
		"FullObservation": {
			in:  *subnetwork(addOutputFields),
			out: *observation(),
		},
		"PartialObservation": {
			in: *subnetwork(addOutputFields, func(s *compute.Subnetwork) {
				s.GatewayAddress = ""
			}),
			out: *observation(func(o *v1beta12.SubnetworkObservation) {
				o.GatewayAddress = ""
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateSubnetworkObservation(tc.in)
			if diff := cmp.Diff(tc.out, r, cmpopts.SortSlices(equateSecondaryRange)); diff != "" {
				t.Errorf("GenerateSubnetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec *v1beta12.SubnetworkParameters
		in   compute.Subnetwork
	}
	cases := map[string]struct {
		args args
		want *v1beta12.SubnetworkParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: params(),
				in:   *subnetwork(),
			},
			want: params(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: params(),
				in: *subnetwork(func(n *compute.Subnetwork) {
					n.Description = "some other description"
				}),
			},
			want: params(),
		},
		"PartialFilled": {
			args: args{
				spec: params(func(p *v1beta12.SubnetworkParameters) {
					p.EnableFlowLogs = nil
				}),
				in: *subnetwork(),
			},
			want: params(func(p *v1beta12.SubnetworkParameters) {
				p.EnableFlowLogs = &trueVal
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
		name    string
		in      *v1beta12.SubnetworkParameters
		current *compute.Subnetwork
	}
	type want struct {
		upToDate bool
		privAcc  bool
		isErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"UpToDate": {
			args: args{
				name:    testName,
				in:      params(),
				current: subnetwork(),
			},
			want: want{upToDate: true, privAcc: false},
		},
		"UpToDateWithOutputFields": {
			args: args{
				name:    testName,
				in:      params(),
				current: subnetwork(addOutputFields),
			},
			want: want{upToDate: true, privAcc: false},
		},
		"NotUpToDate": {
			args: args{
				name: testName,
				in:   params(),
				current: subnetwork(func(s *compute.Subnetwork) {
					s.Description = "some other description"
				}),
			},
			want: want{upToDate: false, privAcc: false},
		},
		"NotUpToDatePrivateAccess": {
			args: args{
				name: testName,
				in:   params(),
				current: subnetwork(func(s *compute.Subnetwork) {
					s.PrivateIpGoogleAccess = false
				}),
			},
			want: want{upToDate: false, privAcc: true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			u, p, err := IsUpToDate(tc.args.name, tc.args.in, tc.args.current)
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, u); diff != "" {
				t.Errorf("IsUpToDate(...) Up To Date: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.privAcc, p); diff != "" {
				t.Errorf("IsUpToDate(...) Private Access: -want, +got:\n%s", diff)
			}
		})
	}
}
