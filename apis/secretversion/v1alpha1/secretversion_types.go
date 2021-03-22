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

// SecretVersionState gives the state of a [SecretVersion][google.cloud.secretmanager.v1.SecretVersion], indicating if it can be accessed.
type SecretVersionState int32

const (
	// SecretVersionSTATEUNSPECIFIED represents that SecretVersion state is not specified. This value is unused and invalid.
	SecretVersionSTATEUNSPECIFIED SecretVersionState = 0
	// SecretVersionENABLED represents that SecretVersion state  may be accessed.
	SecretVersionENABLED SecretVersionState = 1
	// SecretVersionDISABLED represents that SecretVersion state may not be accessed, but the secret data
	// is still available and can be placed back into the [ENABLED] state.
	SecretVersionDISABLED SecretVersionState = 2
	// SecretVersionDESTROYED represents that SecretVersion state is destroyed and the secret data is no longer
	// stored. A version may not leave this state once entered.
	SecretVersionDESTROYED SecretVersionState = 3
)

// SecretVersionStateName is to map integers with states
var SecretVersionStateName = map[int32]string{
	0: "STATE_UNSPECIFIED",
	1: "ENABLED",
	2: "DISABLED",
	3: "DESTROYED",
}

// SecretVersionStateValue is to map states with integers
var SecretVersionStateValue = map[string]int32{
	"STATE_UNSPECIFIED": 0,
	"ENABLED":           1,
	"DISABLED":          2,
	"DESTROYED":         3,
}

// SecretVersionParameters defines parameters for a desired Secret Manager's secret version.
type SecretVersionParameters struct {
	// SecretRef refers to the secret object(GCP) created in Kubernetes
	SecretRef string `json:"secretref,omitempty"`

	// Payload is the secret payload of the [SecretVersion][google.cloud.secretmanager.v1.SecretVersion].
	Payload SecretPayload `json:"payload,omitempty"`

	// DesiredSecretVersionState is the desired state the end user wants for the secret version
	DesiredSecretVersionState SecretVersionState `json:"desiredsecretversionstate,omitempty"`
}

// SecretPayload is a secret payload resource in the Secret Manager API. This contains the
// sensitive secret data that is associated with a [SecretVersion][google.cloud.secretmanager.v1.SecretVersion].
type SecretPayload struct {
	// Data is the secret data. Must be no larger than 64KiB.
	Data string `json:"data,omitempty"`
}

// SecretVersionObservation is used to show the observed state of the
// Secret resource on GCP Secrets Manager. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type SecretVersionObservation struct {

	// Output only. The time at which the [SecretVersion][google.cloud.secretmanager.v1.SecretVersion] was created.
	CreateTime *string `json:"create_time,omitempty"`

	// Output only. The time this [SecretVersion][google.cloud.secretmanager.v1.SecretVersion] was destroyed.
	// Only present if [state][google.cloud.secretmanager.v1.SecretVersion.state] is
	// [DESTROYED][google.cloud.secretmanager.v1.SecretVersion.State.DESTROYED].
	DestroyTime *string `json:"destroy_time,omitempty"`

	// Output only. This must be unique within the project. External name of the object is set to this field, hence making it optional
	// +optional
	SecretID *string `json:"secretid,omitempty"`

	// Output only. The current state of the [SecretVersion][google.cloud.secretmanager.v1.SecretVersion].
	State SecretVersionState `json:"state,omitempty"`
}

// SecretVersionSpec defines the desired state of a
// Secret.
type SecretVersionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretVersionParameters `json:"forProvider"`
}

// SecretVersionStatus represents the observed state of a
// Secret.
type SecretVersionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretVersionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// SecretVersion is a managed resource that represents a version of  Google Secret Manger's Secret.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type SecretVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretVersionSpec   `json:"spec"`
	Status SecretVersionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretVersionList contains a list of Secret Version types
type SecretVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretVersion `json:"items"`
}
