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

package cryptokey

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/cloudkms/v1"

	"github.com/crossplane-contrib/provider-gcp/apis/kms/v1alpha1"
)

func TestGenerateObservation(t *testing.T) {
	createTime := "2020-12-15T11:31:46.958565764Z"
	rotationTime := "2021-01-14T21:00:00Z"
	testCryptoKey := "test-crypto-key"
	type args struct {
		in cloudkms.CryptoKey
	}
	type want struct {
		out v1alpha1.CryptoKeyObservation
	}
	cases := map[string]struct {
		args
		want
	}{
		"Empty": {
			args: args{
				in: cloudkms.CryptoKey{},
			},
			want: want{
				out: v1alpha1.CryptoKeyObservation{},
			},
		},
		"Valid": {
			args: args{
				in: cloudkms.CryptoKey{
					CreateTime:       createTime,
					Name:             testCryptoKey,
					NextRotationTime: rotationTime,
				},
			},
			want: want{
				out: v1alpha1.CryptoKeyObservation{
					CreateTime:       createTime,
					Name:             testCryptoKey,
					NextRotationTime: rotationTime,
				},
			},
		},
		"WithPrimaryKey": {
			args: args{
				in: cloudkms.CryptoKey{
					CreateTime:       createTime,
					Name:             testCryptoKey,
					NextRotationTime: rotationTime,
					Primary: &cloudkms.CryptoKeyVersion{
						Algorithm:       "GOOGLE_SYMMETRIC_ENCRYPTION",
						CreateTime:      createTime,
						Name:            "latest-key",
						ProtectionLevel: "HSM",
						State:           "Enabled",
					},
				},
			},
			want: want{
				out: v1alpha1.CryptoKeyObservation{
					CreateTime:       createTime,
					Name:             testCryptoKey,
					NextRotationTime: rotationTime,
					Primary: &v1alpha1.CryptoKeyVersion{
						Algorithm:       "GOOGLE_SYMMETRIC_ENCRYPTION",
						CreateTime:      createTime,
						Name:            "latest-key",
						ProtectionLevel: "HSM",
						State:           "Enabled",
					},
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.args.in)
			if diff := cmp.Diff(tc.want.out, got); diff != "" {
				t.Errorf("#TODO(...): -want result, +got result: %s", diff)
			}
		})
	}
}
