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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
)

var _ resource.AttributeReferencer = (*GlobalAddressNameReferencerForConnection)(nil)
var _ resource.AttributeReferencer = (*NetworkURIReferencerForConnection)(nil)

func TestGlobalAddressNameReferencerForConnection(t *testing.T) {
	value := "cool"

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
		"AssignWrongType": {
			reason: "Assign should return an error when passed a CanReference that is not a *Connection.",
			args: args{
				res:   nil,
				value: value,
			},
			want: want{
				err: errors.New(errResourceIsNotConnection),
			},
		},
		"AssignSuccessful": {
			reason: "Assign should append the supplied value to ReservedPeeringRanges.",
			args: args{
				res:   &Connection{},
				value: value,
			},
			want: want{
				res: &Connection{
					Spec: ConnectionSpec{
						ForProvider: ConnectionParameters{ReservedPeeringRanges: []string{value}},
					},
				},
				err: nil,
			},
		},
		"AssignNoOp": {
			reason: "Assign should not append existing values to ReservedPeeringRanges.",
			args: args{
				res: &Connection{
					Spec: ConnectionSpec{
						ForProvider: ConnectionParameters{ReservedPeeringRanges: []string{value}},
					},
				},
				value: value,
			},
			want: want{
				res: &Connection{
					Spec: ConnectionSpec{
						ForProvider: ConnectionParameters{ReservedPeeringRanges: []string{value}},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &GlobalAddressNameReferencerForConnection{}
			err := r.Assign(tc.args.res, tc.args.value)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\nAssign(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.res, tc.args.res); diff != "" {
				t.Errorf("\nReason: %s\nAssign(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}

}

func TestNetworkURIReferencerForConnection_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &NetworkURIReferencerForConnection{}
	expectedErr := errors.New(errResourceIsNotConnection)

	err := r.Assign(nil, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestNetworkURIReferencerForConnection_AssignValidType_ReturnsExpected(t *testing.T) {
	mv := "mockValue"
	r := &NetworkURIReferencerForConnection{}
	res := &Connection{}
	var expectedErr error

	err := r.Assign(res, mv)
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.Network, &mv); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}
