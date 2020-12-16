package keyring

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/cloudkms/v1"

	"github.com/crossplane/provider-gcp/apis/kms/v1alpha1"
)

func TestGenerateObservation(t *testing.T) {
	createTime := "2020-12-15T11:31:46.958565764Z"
	testKeyRing := "test-keyring"
	type args struct {
		in cloudkms.KeyRing
	}
	type want struct {
		out v1alpha1.KeyRingObservation
	}
	cases := map[string]struct {
		args
		want
	}{
		"Empty": {
			args: args{
				in: cloudkms.KeyRing{
					CreateTime: "",
					Name:       "",
				},
			},
			want: want{
				out: v1alpha1.KeyRingObservation{
					CreateTime: "",
					Name:       "",
				},
			},
		},
		"Valid": {
			args: args{
				in: cloudkms.KeyRing{
					CreateTime: createTime,
					Name:       testKeyRing,
				},
			},
			want: want{
				out: v1alpha1.KeyRingObservation{
					CreateTime: createTime,
					Name:       testKeyRing,
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
