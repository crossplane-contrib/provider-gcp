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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// Keys used in connection secret.
const (
	ConnectionSecretKeyName        = "secret"
	ConnectionSecretKeyProjectName = "projectName"
)

// SecretParameters defines parameters for a desired Secret Manager's secret.
type SecretParameters struct {

	// Required. The resource name of the project to associate with the
	// [Secret][google.cloud.secretmanager.v1.Secret], in the format `projects/*`.
	Parent string `json:"parent,omitempty"`

	// Labels are used as additional metadata on Topic.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// The replication policy of the secret data attached to the Secret. Can be automatic or userManaged.
	// Ref: https://cloud.google.com/secret-manager/docs/reference/rest/v1/projects.secrets
	// +optional
	// +immutable
	Replication *Replication `json:"replication,omitempty"`
}

// Replication policy that defines the replication configuration of data.
type Replication struct {
	// The replication policy for this secret.
	// This devaites from the internal representation of the google API since it's an interface.
	// Ref: google/cloud/secretmanager/v1/resources.proto
	// There are two possible values : usermanaged and automatic
	// +optional
	// +immutable
	ReplicationType *ReplicationType `json:"ReplicationType"`
}

// ReplicationType refers to type of replication of the secrets
type ReplicationType struct {
	// +optional
	// +immutable
	// +nullable
	Automatic *ReplicationAutomatic `json:"ReplicationAutomatic"`

	// +optional
	// +immutable
	UserManaged *ReplicationUserManaged `json:"ReplicationUserManaged"`
}

// ReplicationAutomatic has fields for automatic replication of secrets
type ReplicationAutomatic struct {
	// intentionally empty since the GCP SDK of current version has no fields.
	// TODO: Add new fields when GCP SDK is updated
}

// ReplicationUserManaged has fields for user managed replication of secrets
type ReplicationUserManaged struct {

	// Locations of the secret.
	// +immutable
	Replicas []*ReplicationUserManagedReplica `json:"Replicas"`
}

// ReplicationUserManagedReplica has fields of each replica of user managed replica type
type ReplicationUserManagedReplica struct {
	// Location of the secret.
	// +immutable
	Location string `json:"location,omitempty"`
}

// SecretObservation is used to show the observed state of the
// Secret resource on GCP Secrets Manager. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type SecretObservation struct {
	// CreateTime: Output only. The time at which this Secret was created.
	CreateTime string `json:"createTime,omitempty"`
}

// SecretSpec defines the desired state of a
// Secret.
type SecretSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretParameters `json:"forProvider"`
}

// SecretStatus represents the observed state of a
// Secret.
type SecretStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Secret is a managed resource that represents a Google Secret Manger's Secret.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Secret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretSpec   `json:"spec"`
	Status SecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretList contains a list of Secret types
type SecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secret `json:"items"`
}
