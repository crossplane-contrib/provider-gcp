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
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go"
	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/crossplane/provider-gcp/apis/secretversion/v1alpha1"
)

// Client is interface that lists the required functions for the reconciler
// to work.
type Client interface {
	AddSecretVersion(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	EnableSecretVersion(ctx context.Context, req *secretmanager.EnableSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	DisableSecretVersion(ctx context.Context, req *secretmanager.DisableSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	GetSecretVersion(ctx context.Context, req *secretmanager.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	DestroySecretVersion(ctx context.Context, req *secretmanager.DestroySecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
}

// NewAddSecretVersionRequest produces a Secret that is configured via given SecretParameters.
func NewAddSecretVersionRequest(projectID string, sp v1alpha1.SecretVersionParameters) *secretmanager.AddSecretVersionRequest {

	parent := fmt.Sprintf("projects/%s/secrets/%s", projectID, sp.SecretRef)

	payload := &secretmanager.SecretPayload{
		Data: []byte(sp.Payload.Data),
	}
	req := &secretmanager.AddSecretVersionRequest{
		Parent:  parent,
		Payload: payload,
	}

	return req
}

// LateInitialize fills the empty fields of SecretVersionParameters if the corresponding
// fields are given in Secret Version.
func LateInitialize(sp *v1alpha1.SecretVersionParameters, sv *secretmanager.SecretVersion, data []byte, secretRef string) {

	// if sv.State != secretmanagerpb.SecretVersion_State(sp.DesiredSecretVersionState) {
	// 	sp.DesiredSecretVersionState = v1alpha1.SecretVersionState(sv.State)
	// }
	if sp.SecretRef == "" {
		sp.SecretRef = secretRef
	}

	if sp.Payload.Data == "" && data != nil {
		sp.Payload.Data = string(data)
	}

}

// IsUpToDate checks whether Secret is configured with given SecretParameters.
func IsUpToDate(sp v1alpha1.SecretVersionParameters, sv *secretmanager.SecretVersion) bool {
	observed := &v1alpha1.SecretVersionParameters{}
	LateInitialize(observed, sv, []byte(sp.Payload.Data), sp.SecretRef)
	result := cmp.Equal(observed, &sp)
	return result
}

// Observation of a secret version
type Observation struct {
	// Name is the name of the secret version. It is a counter
	Name int

	// CreateTime is the time at which secret was created
	CreateTime string

	// DestroyTime is the time at which secret was destroyed
	DestroyTime string

	// State of the secret version
	State v1alpha1.SecretVersionState
}

// UpdateStatus updates any fields of the supplied SecretVersionStatus
func UpdateStatus(s *v1alpha1.SecretVersionStatus, o Observation) {
	s.AtProvider.CreateTime = &o.CreateTime
	s.AtProvider.DestroyTime = &o.DestroyTime
	s.AtProvider.State = o.State
}
