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

package secret

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/crossplane/provider-gcp/apis/secretsmanager/secret/v1alpha1"
)

const (
	projectID = "fooproject"
	name      = "barname"
	location  = "us-east1"
)

func automaticParams() *v1alpha1.SecretParameters {
	return &v1alpha1.SecretParameters{
		Parent: fmt.Sprintf("projects/%s", projectID),
		Labels: map[string]string{
			"foo": "bar",
		},
	}
}

func managedParams() *v1alpha1.SecretParameters {
	replicas := make([]*v1alpha1.ReplicationUserManagedReplica, 0)
	replica := &v1alpha1.ReplicationUserManagedReplica{
		Location: location,
	}
	replicas = append(replicas, replica)
	replication := v1alpha1.Replication{
		ReplicationType: &v1alpha1.ReplicationType{
			UserManaged: &v1alpha1.ReplicationUserManaged{
				Replicas: replicas,
			},
		},
	}
	return &v1alpha1.SecretParameters{
		Parent: fmt.Sprintf("projects/%s", projectID),
		Labels: map[string]string{
			"foo": "bar",
		},
		Replication: &replication,
	}
}

func secretAutomatic() *secretmanager.Secret {
	return &secretmanagerpb.Secret{
		Labels: map[string]string{
			"foo": "bar",
		},
		Replication: &secretmanagerpb.Replication{
			Replication: &secretmanagerpb.Replication_Automatic_{
				Automatic: &secretmanagerpb.Replication_Automatic{},
			},
		},
	}
}

func secretManaged() *secretmanager.Secret {
	replicas := make([]*secretmanagerpb.Replication_UserManaged_Replica, 0)
	replica := &secretmanagerpb.Replication_UserManaged_Replica{
		Location: location,
	}

	replicas = append(replicas, replica)
	return &secretmanagerpb.Secret{
		Labels: map[string]string{
			"foo": "bar",
		},
		Replication: &secretmanagerpb.Replication{
			Replication: &secretmanagerpb.Replication_UserManaged_{
				UserManaged: &secretmanagerpb.Replication_UserManaged{
					Replicas: replicas,
				},
			},
		},
	}
}

func TestGenerateSecret(t *testing.T) {
	type args struct {
		sp v1alpha1.SecretParameters
	}
	cases := map[string]struct {
		args
		out *secretmanager.Secret
	}{
		"Automatic": {
			args: args{
				sp: *automaticParams(),
			},
			out: secretAutomatic(),
		},
		"Managed": {
			args: args{
				sp: *managedParams(),
			},
			out: secretManaged(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out := GenerateSecret(tc.sp)
			if diff := cmp.Diff(tc.out, out); diff != "" {
				t.Errorf("GenerateSecret(...): -want, +got:\n%s", diff)
			}
		})
	}

}

func automaticCreateSecretRequestParams() *secretmanager.CreateSecretRequest {

	return &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: name,
		Secret:   secretAutomatic(),
	}

}

func managedCreateSecretRequestParams() *secretmanager.CreateSecretRequest {

	return &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: name,
		Secret:   secretManaged(),
	}

}

func TestNewSecretRequest(t *testing.T) {
	type args struct {
		projectID string
		name      string
		sp        v1alpha1.SecretParameters
	}

	cases := map[string]struct {
		args
		out *secretmanager.CreateSecretRequest
	}{
		"Automatic": {
			args: args{
				projectID: projectID,
				name:      name,
				sp:        *automaticParams(),
			},
			out: automaticCreateSecretRequestParams(),
		},
		"Managed": {
			args: args{
				projectID: projectID,
				name:      name,
				sp:        *managedParams(),
			},
			out: managedCreateSecretRequestParams(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out := NewCreateSecretRequest(tc.projectID, tc.name, tc.sp)
			if diff := cmp.Diff(tc.out, out); diff != "" {
				t.Errorf("NewCreateSecretRequest(...): -want, +got:\n%s", diff)
			}
		})
	}

}

func TestLateInitalize(t *testing.T) {
	type args struct {
		sp *v1alpha1.SecretParameters
		s  secretmanager.Secret
	}
	cases := map[string]struct {
		args
		out *v1alpha1.SecretParameters
	}{
		"Managed": {
			args: args{
				sp: &v1alpha1.SecretParameters{
					Parent: fmt.Sprintf("projects/%s", projectID),
				},
				s: *secretManaged(),
			},
			out: managedParams(),
		},
		"Automatic": {
			args: args{
				sp: &v1alpha1.SecretParameters{
					Parent: fmt.Sprintf("projects/%s", projectID),
				},
				s: *secretAutomatic(),
			},
			out: automaticParams(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.sp, tc.s)
			if diff := cmp.Diff(tc.args.sp, tc.out); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUptoDate(t *testing.T) {
	// Setting name explicitly since these are output only fields
	managedSecret := *secretManaged()
	managedSecret.Name = fmt.Sprintf("projects/%s/secrets/%s", projectID, name)

	automaticSecret := *secretAutomatic()
	automaticSecret.Name = fmt.Sprintf("projects/%s/secrets/%s", projectID, name)

	type args struct {
		sp *v1alpha1.SecretParameters
		s  secretmanager.Secret
	}
	cases := map[string]struct {
		args
		result bool
	}{
		"ManagedFalse": {
			args: args{
				sp: &v1alpha1.SecretParameters{
					Parent: fmt.Sprintf("projects/%s", projectID),
				},
				s: managedSecret,
			},
			result: false,
		},
		"AutomaticFalse": {
			args: args{
				sp: &v1alpha1.SecretParameters{
					Parent: fmt.Sprintf("projects/%s", projectID),
				},
				s: automaticSecret,
			},
			result: false,
		},
		"ManagedTrue": {
			args: args{
				sp: managedParams(),
				s:  managedSecret,
			},
			result: true,
		},
		"AutomaticTrue": {
			args: args{
				sp: automaticParams(),
				s:  automaticSecret,
			},
			result: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res := IsUpToDate(*tc.sp, tc.s)
			if diff := cmp.Diff(tc.result, res); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
