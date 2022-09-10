/*
Copyright 2022 The Crossplane Authors.

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

package managedzone

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"google.golang.org/api/dns/v1"

	"github.com/crossplane-contrib/provider-gcp/apis/dns/v1alpha1"
)

const (
	name    = "test-mz"
	dnsName = "test.local."
)

var (
	testDescription = "test description"
	testVisibility  = "private"
	fakeVisibility  = "public"
)

func params(m ...func(*v1alpha1.ManagedZoneParameters)) *v1alpha1.ManagedZoneParameters {
	p := &v1alpha1.ManagedZoneParameters{
		Description: &testDescription,
		DNSName:     dnsName,
		Visibility:  &testVisibility,
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func managedZone(m ...func(*dns.ManagedZone)) *dns.ManagedZone {
	mz := &dns.ManagedZone{
		Kind:        "dns#managedZone",
		Description: testDescription,
		DnsName:     dnsName,
		Name:        name,
		Visibility:  testVisibility,
	}

	for _, f := range m {
		f(mz)
	}
	return mz
}

func TestGenerateManagedZone(t *testing.T) {
	type args struct {
		name   string
		params v1alpha1.ManagedZoneParameters
	}
	type want struct {
		managedZone *dns.ManagedZone
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
				managedZone: managedZone(),
			},
		},
		"MissingFields": {
			args: args{
				name: name,
				params: *params(func(p *v1alpha1.ManagedZoneParameters) {
					p.PrivateVisibilityConfig = nil
				}),
			},
			want: want{
				managedZone: managedZone(func(mz *dns.ManagedZone) {
					mz.PrivateVisibilityConfig = nil
				}),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mz := &dns.ManagedZone{}
			GenerateManagedZone(tc.args.name, tc.args.params, mz)
			if diff := cmp.Diff(tc.want.managedZone, mz); diff != "" {
				t.Errorf("GenerateManagedZone(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec     *v1alpha1.ManagedZoneParameters
		external *dns.ManagedZone
	}
	type want struct {
		params *v1alpha1.ManagedZoneParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				spec: params(func(p *v1alpha1.ManagedZoneParameters) {
					p.Visibility = nil
				}),
				external: managedZone(func(mz *dns.ManagedZone) {
					mz.Visibility = fakeVisibility
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.ManagedZoneParameters) {
					p.Visibility = &fakeVisibility
				}),
			},
		},
		"AllFilledAlready": {
			args: args{
				spec:     params(),
				external: managedZone(),
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

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params *v1alpha1.ManagedZoneParameters
		mz     *dns.ManagedZone
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
				mz:     managedZone(),
			},
			want: want{
				upToDate: true,
			},
		},
		"NeedsUpdate": {
			args: args{
				params: params(),
				mz: managedZone(func(mz *dns.ManagedZone) {
					mz.Description = "new description"
				}),
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate("test-mz", tc.args.params, tc.args.mz)
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
