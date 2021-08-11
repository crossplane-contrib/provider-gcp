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

package dns

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/dns/v1"

	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
)

const (
	name = "test-rrs"
)

var (
	fakeSignature = []string{"fakeSignature"}
)

func params(m ...func(*v1alpha1.ResourceRecordSetParameters)) *v1alpha1.ResourceRecordSetParameters {
	p := &v1alpha1.ResourceRecordSetParameters{
		ManagedZone:      "crossplane-zone",
		Kind:             "dns#resourceRecordSet",
		Type:             "A",
		TTL:              int64(300),
		RRDatas:          []string{"1.2.3.4"},
		SignatureRRDatas: &fakeSignature,
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func resourceRecordSet(m ...func(*dns.ResourceRecordSet)) *dns.ResourceRecordSet {
	rrs := &dns.ResourceRecordSet{
		Kind:             "dns#resourceRecordSet",
		Type:             "A",
		Ttl:              int64(300),
		Rrdatas:          []string{"1.2.3.4"},
		Name:             name,
		SignatureRrdatas: fakeSignature,
	}
	for _, f := range m {
		f(rrs)
	}
	return rrs
}

func observation(m ...func(*v1alpha1.ResourceRecordSetObservation)) *v1alpha1.ResourceRecordSetObservation {
	return &v1alpha1.ResourceRecordSetObservation{
		Name: name,
		Type: "A",
	}
}

func TestGenerateResourceRecordSet(t *testing.T) {
	type args struct {
		name   string
		params v1alpha1.ResourceRecordSetParameters
	}
	type want struct {
		resourceRecordSet *dns.ResourceRecordSet
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				name:   name,
				params: *params(),
			},
			want: want{
				resourceRecordSet: resourceRecordSet(),
			},
		},
		"MissingFields": {
			args: args{
				name: name,
				params: *params(func(p *v1alpha1.ResourceRecordSetParameters) {
					p.SignatureRRDatas = nil
				}),
			},
			want: want{
				resourceRecordSet: resourceRecordSet(func(rrs *dns.ResourceRecordSet) {
					rrs.SignatureRrdatas = nil
				}),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rrs := &dns.ResourceRecordSet{}
			GenerateResourceRecordSet(tc.args.name, tc.args.params, rrs)
			if diff := cmp.Diff(tc.want.resourceRecordSet, rrs); diff != "" {
				t.Errorf("GenerateResourceRecordSet(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec     *v1alpha1.ResourceRecordSetParameters
		external *dns.ResourceRecordSet
	}
	type want struct {
		params *v1alpha1.ResourceRecordSetParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				spec: params(func(p *v1alpha1.ResourceRecordSetParameters) {
					p.SignatureRRDatas = nil
				}),
				external: resourceRecordSet(func(rrs *dns.ResourceRecordSet) {
					rrs.SignatureRrdatas = fakeSignature
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.ResourceRecordSetParameters) {
					p.SignatureRRDatas = &fakeSignature
				}),
			},
		},
		"AllFilledAlready": {
			args: args{
				spec:     params(),
				external: resourceRecordSet(),
			},
			want: want{
				params: params(),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.spec, *tc.args.external)
			if diff := cmp.Diff(tc.want.params, tc.args.spec); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	type args struct {
		external dns.ResourceRecordSet
	}
	type want struct {
		obs v1alpha1.ResourceRecordSetObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				external: *resourceRecordSet(),
			},
			want: want{
				obs: *observation(),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o := GenerateObservation(tc.args.external)
			if diff := cmp.Diff(tc.want.obs, o); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params *v1alpha1.ResourceRecordSetParameters
		rrs    *dns.ResourceRecordSet
	}
	type want struct {
		upToDate bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"IsUpToDate": {
			args: args{
				params: params(),
				rrs:    resourceRecordSet(),
			},
			want: want{
				upToDate: true,
			},
		},
		"NeedsUpdate": {
			args: args{
				params: params(),
				rrs: resourceRecordSet(func(rrs *dns.ResourceRecordSet) {
					rrs.SignatureRrdatas = []string{"signature"}
				}),
			},
			want: want{
				upToDate: false,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate("test-rrs", tc.args.params, tc.args.rrs)
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
