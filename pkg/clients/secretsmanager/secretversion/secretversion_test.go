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

package secretversion

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/crossplane/provider-gcp/apis/secretsmanager/secretversion/v1alpha1"
)

const (
	projectID = "fooproject"
	name      = "barname"
	location  = "us-east1"
	payload   = "data"
)

func paramsAddSecretVersionRequest() *secretmanager.AddSecretVersionRequest {

	return &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", projectID, name),
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(payload),
		},
	}

}

func paramsSecretVersion() *v1alpha1.SecretVersionParameters {

	return &v1alpha1.SecretVersionParameters{
		SecretRef: name,
		Payload: &v1alpha1.SecretPayload{
			Data: payload,
		},
		DesiredSecretVersionState: "ENABLED",
	}
}

func TestNewAddSecretVersionRequest(t *testing.T) {
	type args struct {
		projectID string
		sp        v1alpha1.SecretVersionParameters
	}

	cases := map[string]struct {
		args
		out *secretmanager.AddSecretVersionRequest
	}{
		"Full": {
			args: args{
				projectID: projectID,
				sp:        *paramsSecretVersion(),
			},
			out: paramsAddSecretVersionRequest(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out := NewAddSecretVersionRequest(tc.projectID, tc.sp)
			if diff := cmp.Diff(tc.out, out); diff != "" {
				t.Errorf("NewAddSecretVersionRequest(...): -want, +got:\n%s", diff)
			}
		})
	}

}

func secretVersion() *secretmanager.SecretVersion {
	return &secretmanagerpb.SecretVersion{
		Name:  fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, name, "1"),
		State: 1, // 1 implies state is ENABLED
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		sp   *v1alpha1.SecretVersionParameters
		sv   *secretmanager.SecretVersion
		data []byte
	}

	cases := map[string]struct {
		args
		out *v1alpha1.SecretVersionParameters
	}{
		"Full": {
			args: args{
				sp: &v1alpha1.SecretVersionParameters{
					Payload: &v1alpha1.SecretPayload{
						Data: "",
					},
				},
				sv:   secretVersion(),
				data: []byte(payload),
			},
			out: paramsSecretVersion(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.sp, tc.sv, tc.data)
			if diff := cmp.Diff(tc.out, tc.sp); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUptoData(t *testing.T) {
	type args struct {
		sp v1alpha1.SecretVersionParameters
		sv *secretmanager.SecretVersion
	}

	cases := map[string]struct {
		args
		res bool
	}{
		"True": {
			args: args{
				sp: *paramsSecretVersion(),
				sv: secretVersion(),
			},
			res: true,
		},
		"False": {args: args{
			sp: v1alpha1.SecretVersionParameters{
				Payload: &v1alpha1.SecretPayload{
					Data: "",
				},
			},
			sv: secretVersion(),
		},
			res: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res := IsUpToDate(tc.sp, tc.sv)
			if diff := cmp.Diff(tc.res, res); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
