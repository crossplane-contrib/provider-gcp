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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

const (
	// LabelKeyProviderName is added to ProviderConfigUsages to relate them to
	// their ProviderConfig.
	LabelKeyProviderName = "crossplane.io/provider-config"
)

// A ProviderConfigSpec defines the desired state of a ProviderConfig.
type ProviderConfigSpec struct {
	// CredentialsSecretRef references a specific secret's key that contains
	// the credentials that are used to connect to the GCP API.
	CredentialsSecretRef v1alpha1.SecretKeySelector `json:"credentialsSecretRef"`

	// ProjectID is the project name (not numerical ID) of this GCP ProviderConfig.
	ProjectID string `json:"projectID"`
}

// ProviderConfigStatus represents the observed state of a ProviderConfig.
type ProviderConfigStatus struct {
	// Usages of this ProviderConfig.
	Usages *int64 `json:"usages,omitempty"`
}

// +kubebuilder:object:root=true

// A ProviderConfig configures how GCP controller should connect to GCP API.
// +kubebuilder:printcolumn:name="PROJECT-ID",type="string",JSONPath=".spec.projectID"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="USAGES",type="integer",JSONPath=".status.usages"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentialsSecretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,gcp}
// +kubebuilder:subresource:status
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}

// +kubebuilder:object:root=true

// A ProviderConfigUsage indicates that a resource is using a ProviderConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="PROVIDER-NAME",type="string",JSONPath=".providerRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,gcp}
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ProviderRef v1alpha1.Reference `json:"providerRef"`

	ResourceRef v1alpha1.TypedReference `json:"resourceRef"`
}

// +kubebuilder:object:root=true

// ProviderConfigUsageList contains a list of ProviderConfigUsage
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}
