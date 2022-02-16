/*
Copyright 2022 The Crossplane Authors.

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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RepositoryParameters define the desired state of a Repository
type RepositoryParameters struct {
	// Location: The name of the Google Compute zone or region
	// +immutable
	Location string `json:"location"`

	// Description: The user-provided description of the repository.
	// +optional
	Description string `json:"description,omitempty"`

	// Format: The format of packages that are stored in the repository.
	//
	// Possible values:
	//   "FORMAT_UNSPECIFIED" - Unspecified package format.
	//   "DOCKER" - Docker package format.
	//   "MAVEN" - Maven package format.
	//   "NPM" - NPM package format.
	//   "PYPI" - PyPI package format. Deprecated, use PYTHON instead.
	//   "APT" - APT package format.
	//   "YUM" - YUM package format.
	//   "PYTHON" - Python package format.
	// +immutable
	// +kubebuilder:validation:Enum=FORMAT_UNSPECIFIED;DOCKER;MAVEN;NPM;APT;YUM;PYTHON
	Format string `json:"format,omitempty"`

	// Labels: Labels with user-defined metadata. This field may contain up
	// to 64 entries. Label keys and values may be no longer than 63
	// characters. Label keys must begin with a lowercase letter and may
	// only contain lowercase letters, numeric characters, underscores, and
	// dashes.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// MavenConfig: Maven repository config contains repository level
	// configuration for the repositories of maven type.
	// +optional
	MavenConfig *MavenRepositoryConfig `json:"mavenConfig,omitempty"`

	// KmsKeyName is the resource name of the Cloud KMS CryptoKey to be used to
	// protect access to messages published on this topic.
	//
	// The expected format is `projects/*/locations/*/keyRings/*/cryptoKeys/*`.
	// +optional
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane/provider-gcp/apis/kms/v1alpha1.CryptoKey
	// +crossplane:generate:reference:extractor=github.com/crossplane/provider-gcp/apis/kms/v1alpha1.CryptoKeyRRN()
	KmsKeyName *string `json:"kmsKeyName,omitempty"`

	// KmsKeyNameRef allows you to specify custom resource name of the KMS Key
	// to fill KmsKeyName field.
	KmsKeyNameRef *xpv1.Reference `json:"kmsKeyNameRef,omitempty"`

	// KmsKeyNameSelector allows you to use selector constraints to select a
	// KMS Key.
	KmsKeyNameSelector *xpv1.Selector `json:"kmsKeyNameSelector,omitempty"`
}

// MavenRepositoryConfig is maven related
// repository details. Provides additional configuration details for
// repositories of the maven format type.
type MavenRepositoryConfig struct {
	// AllowSnapshotOverwrites: The repository with this flag will allow
	// publishing the same snapshot versions.
	AllowSnapshotOverwrites bool `json:"allowSnapshotOverwrites,omitempty"`

	// VersionPolicy: Version policy defines the versions that the registry
	// will accept.
	//
	// Possible values:
	//   "VERSION_POLICY_UNSPECIFIED" - VERSION_POLICY_UNSPECIFIED - the
	// version policy is not defined. When the version policy is not
	// defined, no validation is performed for the versions.
	//   "RELEASE" - RELEASE - repository will accept only Release versions.
	//   "SNAPSHOT" - SNAPSHOT - repository will accept only Snapshot
	// versions.
	VersionPolicy *string `json:"versionPolicy,omitempty"`
}

// RepositoryObservation is used to show the observed state of the Repository
type RepositoryObservation struct {

	// CreateTime: The time when the repository was created.
	CreateTime string `json:"createTime,omitempty"`

	// UpdateTime: The time when the repository was last updated.
	UpdateTime string `json:"updateTime,omitempty"`
}

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RepositoryParameters `json:"forProvider"`
}

// RepositoryStatus defines the observed state of Repository
type RepositoryStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RepositoryObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}

// Repository is a managed resource that represents a Google Artifact registry Repository.
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RepositoryList contains a list of Repository
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repository{}, &RepositoryList{})
}
