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

	iamv1alpha1 "github.com/crossplane/provider-gcp/apis/classic/iam/v1alpha1"
)

// CryptoKeyPolicyParameters defines parameters for a desired KMS CryptoKeyPolicy
// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys
type CryptoKeyPolicyParameters struct {
	// CryptoKey: The RRN of the CryptoKey to which this CryptoKeyPolicy belongs.
	// +optional
	// +immutable
	CryptoKey *string `json:"cryptoKey,omitempty"`

	// CryptoKeyRef references a CryptoKey and retrieves its URI
	// +optional
	// +immutable
	CryptoKeyRef *xpv1.Reference `json:"cryptoKeyRef,omitempty"`

	// CryptoKeySelector selects a reference to a CryptoKey
	// +optional
	CryptoKeySelector *xpv1.Selector `json:"cryptoKeySelector,omitempty"`

	// Policy: An Identity and Access Management (IAM) policy, which
	// specifies access controls for Google Cloud resources.
	Policy iamv1alpha1.Policy `json:"policy"`
}

// CryptoKeyPolicySpec defines the desired state of a
// CryptoKeyPolicy.
type CryptoKeyPolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CryptoKeyPolicyParameters `json:"forProvider"`
}

// CryptoKeyPolicyStatus represents the observed state of a
// CryptoKeyPolicy.
type CryptoKeyPolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// CryptoKeyPolicy is a managed resource that represents a Google KMS Crypto Key.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type CryptoKeyPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CryptoKeyPolicySpec   `json:"spec"`
	Status CryptoKeyPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CryptoKeyPolicyList contains a list of CryptoKeyPolicy types
type CryptoKeyPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CryptoKeyPolicy `json:"items"`
}
