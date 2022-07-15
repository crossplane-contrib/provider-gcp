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

package globaladdress

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v1"

	"github.com/crossplane-contrib/provider-gcp/apis/compute/v1beta1"
)

var (
	name               = "coolName"
	description        = "coolDescription"
	addressIP          = "coolAddress"
	addressType        = "coolType"
	ipVersion          = "coolVersion"
	network            = "coolNetwork"
	purpose            = "beingCool"
	subnetwork         = "coolSubnet"
	prefixLength int64 = 3001

	timestamp        = "coolTime"
	link             = "coolLink"
	users            = []string{"coolUser", "coolerUser"}
	id        uint64 = 3001
)

func params(m ...func(*v1beta1.GlobalAddressParameters)) *v1beta1.GlobalAddressParameters {
	o := &v1beta1.GlobalAddressParameters{
		Address:      &addressIP,
		AddressType:  &addressType,
		Description:  &description,
		IPVersion:    &ipVersion,
		Network:      &network,
		PrefixLength: &prefixLength,
		Purpose:      &purpose,
		Subnetwork:   &subnetwork,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func address(m ...func(*compute.Address)) *compute.Address {
	o := &compute.Address{
		Address:      addressIP,
		AddressType:  addressType,
		Description:  description,
		IpVersion:    ipVersion,
		Name:         name,
		Network:      network,
		PrefixLength: prefixLength,
		Purpose:      purpose,
		Subnetwork:   subnetwork,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func addOutputFields(n *compute.Address) {
	n.Status = v1beta1.StatusReserving
	n.CreationTimestamp = timestamp
	n.Id = id
	n.SelfLink = link
	n.Users = users

}

func observation(m ...func(*v1beta1.GlobalAddressObservation)) *v1beta1.GlobalAddressObservation {
	o := &v1beta1.GlobalAddressObservation{
		Status:            v1beta1.StatusReserving,
		CreationTimestamp: timestamp,
		ID:                id,
		SelfLink:          link,
		Users:             users,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestGenerateGlobalAddress(t *testing.T) {
	type args struct {
		name string
		in   v1beta1.GlobalAddressParameters
	}
	cases := map[string]struct {
		args args
		want *compute.Address
	}{
		"AllFilled": {
			args: args{
				name: name,
				in:   *params(),
			},
			want: address(),
		},
		"PartialFilled": {
			args: args{
				name: name,
				in: *params(func(p *v1beta1.GlobalAddressParameters) {
					p.AddressType = nil
				}),
			},
			want: address(func(a *compute.Address) {
				a.AddressType = ""
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &compute.Address{}
			GenerateGlobalAddress(tc.args.name, tc.args.in, r)
			if diff := cmp.Diff(r, tc.want); diff != "" {
				t.Errorf("GenerateGlobalAddress(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateNetworkObservation(t *testing.T) {
	cases := map[string]struct {
		in  compute.Address
		out v1beta1.GlobalAddressObservation
	}{
		"AllFilled": {
			in:  *address(addOutputFields),
			out: *observation(),
		},
		"PartialFilled": {
			in: *address(addOutputFields, func(a *compute.Address) {
				a.CreationTimestamp = ""
			}),
			out: *observation(func(o *v1beta1.GlobalAddressObservation) {
				o.CreationTimestamp = ""
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateGlobalAddressObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateGlobalAddressObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec *v1beta1.GlobalAddressParameters
		in   compute.Address
	}
	cases := map[string]struct {
		args args
		want *v1beta1.GlobalAddressParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: params(),
				in:   *address(),
			},
			want: params(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: params(),
				in: *address(func(a *compute.Address) {
					a.Description = "some other description"
				}),
			},
			want: params(),
		},
		"PartialFilled": {
			args: args{
				spec: params(func(p *v1beta1.GlobalAddressParameters) {
					p.AddressType = nil
				}),
				in: *address(),
			},
			want: params(func(p *v1beta1.GlobalAddressParameters) {
				p.AddressType = &addressType
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
