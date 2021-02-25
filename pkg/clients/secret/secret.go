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

package secret

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go"
	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/crossplane/provider-gcp/apis/secretsmanager/v1alpha1"
)

// Client is interface that lists the required functions for the reconciler
// to work.
type Client interface {
	CreateSecret(ctx context.Context, req *secretmanager.CreateSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error)
	UpdateSecret(ctx context.Context, req *secretmanager.UpdateSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error)
	GetSecret(ctx context.Context, req *secretmanager.GetSecretRequest, opts ...gax.CallOption) (*secretmanager.Secret, error)
	DeleteSecret(ctx context.Context, req *secretmanager.DeleteSecretRequest, opts ...gax.CallOption) error
}

// GenerateSecret is used to convert Crossplane SecretParameters
// to GCP's Secret object.
func GenerateSecret(name string, sp v1alpha1.SecretParameters, s *secretmanager.Secret) {
	s.Labels = sp.Labels
	if sp.Replication != nil {
		if sp.Replication.ReplicationType.UserManaged.Replicas != nil {
			s.Replication = &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_UserManaged_{
					UserManaged: &secretmanagerpb.Replication_UserManaged{
						Replicas: convertCrossplaneReplicasToGCPReplicas(sp.Replication.ReplicationType.UserManaged.Replicas),
					},
				},
			}
		}

	} else {
		s.Replication = &secretmanagerpb.Replication{
			Replication: &secretmanagerpb.Replication_Automatic_{
				Automatic: &secretmanagerpb.Replication_Automatic{},
			},
		}
	}

}

// NewCreateSecretRequest produces a Secret that is configured via given SecretParameters.
func NewCreateSecretRequest(projectID, name string, sp v1alpha1.SecretParameters) *secretmanager.CreateSecretRequest {
	secret := &secretmanager.Secret{}

	GenerateSecret(name, sp, secret)
	req := &secretmanager.CreateSecretRequest{
		Parent:   sp.Parent,
		SecretId: name,
		Secret:   secret,
	}

	return req
}

// LateInitialize fills the empty fields of SecretParameters if the corresponding
// fields are given in Secret.
func LateInitialize(sp *v1alpha1.SecretParameters, s secretmanager.Secret) {
	if len(sp.Labels) == 0 && len(s.Labels) != 0 {
		sp.Labels = map[string]string{}
		for k, v := range s.Labels {
			sp.Labels[k] = v
		}
	}

}

// IsUpToDate checks whether Secret is configured with given SecretParameters.
func IsUpToDate(sp v1alpha1.SecretParameters, s secretmanager.Secret) bool {
	observed := &v1alpha1.SecretParameters{}
	LateInitialize(observed, s)
	return cmp.Equal(observed, &sp)
}

// GenerateUpdateRequest produces an UpdateTopicRequest with the difference
// between SecretParameters and Secret.
func GenerateUpdateRequest(projectID, name string, sp v1alpha1.SecretParameters, s secretmanager.Secret) *secretmanager.UpdateSecretRequest {
	observed := &v1alpha1.SecretParameters{}
	LateInitialize(observed, s)
	us := &secretmanagerpb.UpdateSecretRequest{Secret: &secretmanagerpb.Secret{Name: fmt.Sprintf("projects/%s/secrets/%s", projectID, name)}, UpdateMask: &field_mask.FieldMask{}}
	if !cmp.Equal(sp.Labels, observed.Labels) {
		us.UpdateMask.Paths = append(us.UpdateMask.Paths, "labels")
		us.Secret.Labels = sp.Labels
	}

	return us
}

// Observation of a secret and
type Observation struct {
	// CreateTime is the time at which secret was created
	CreateTime string
	// SecretID is the name of the secret represented in GCP secret manager
	SecretID string
}

// UpdateStatus updates any fields of the supplied SecretStatus
func UpdateStatus(s *v1alpha1.SecretStatus, o Observation) {
	s.AtProvider.CreateTime = o.CreateTime
	s.AtProvider.SecretID = &o.SecretID
}

func convertCrossplaneReplicasToGCPReplicas(cr []*v1alpha1.ReplicationUserManagedReplica) []*secretmanagerpb.Replication_UserManaged_Replica {
	replicas := make([]*secretmanagerpb.Replication_UserManaged_Replica, 0)

	var gcpv *secretmanagerpb.Replication_UserManaged_Replica
	for _, cv := range cr {

		gcpv = &secretmanagerpb.Replication_UserManaged_Replica{
			Location: cv.Location,
		}
		replicas = append(replicas, gcpv)
	}

	return replicas
}
