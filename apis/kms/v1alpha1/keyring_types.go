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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// KeyRingParameters defines parameters for a desired KMS KeyRing
// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings
// The name of the key ring (ie the `keyRingId` parameter of the Create
// call) is determined by the value of the `crossplane.io/external-name`
// annotation. Unless overridden by the user, this annotation is automatically
// populated with the value of the `metadata.name` attribute.
type KeyRingParameters struct {
	// The location for the KeyRing.
	// A full list of valid locations can be found by running 'gcloud kms locations list'.
	Location string `json:"location"`
}

// KeyRingObservation is used to show the observed state of the
// KeyRing resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type KeyRingObservation struct {
	// CreateTime: Output only. The time at which this KeyRing was created.
	CreateTime string `json:"createTime,omitempty"`

	// Name: Output only. The resource name for the KeyRing in the
	// format `projects/*/locations/*/keyRings/*`.
	Name string `json:"name,omitempty"`
}

// KeyRingSpec defines the desired state of a KeyRing.
type KeyRingSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       KeyRingParameters `json:"forProvider"`
}

// KeyRingStatus represents the observed state of a KeyRing.
type KeyRingStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          KeyRingObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// KeyRing is a managed resource that represents a Google KMS KeyRing
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.forProvider.location"
// +kubebuilder:resource:scope=Cluster
type KeyRing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeyRingSpec   `json:"spec"`
	Status KeyRingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeyRingList contains a list of KeyRing types
type KeyRingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeyRing `json:"items"`
}
