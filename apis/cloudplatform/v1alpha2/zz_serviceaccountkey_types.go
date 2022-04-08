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

// Code generated by terrajet. DO NOT EDIT.

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type ServiceAccountKeyObservation struct {
	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	Name *string `json:"name,omitempty" tf:"name,omitempty"`

	PublicKey *string `json:"publicKey,omitempty" tf:"public_key,omitempty"`

	ValidAfter *string `json:"validAfter,omitempty" tf:"valid_after,omitempty"`

	ValidBefore *string `json:"validBefore,omitempty" tf:"valid_before,omitempty"`
}

type ServiceAccountKeyParameters struct {

	// Arbitrary map of values that, when changed, will trigger recreation of resource.
	// +kubebuilder:validation:Optional
	Keepers map[string]string `json:"keepers,omitempty" tf:"keepers,omitempty"`

	// The algorithm used to generate the key, used only on create. KEY_ALG_RSA_2048 is the default algorithm. Valid values are: "KEY_ALG_RSA_1024", "KEY_ALG_RSA_2048".
	// +kubebuilder:validation:Optional
	KeyAlgorithm *string `json:"keyAlgorithm,omitempty" tf:"key_algorithm,omitempty"`

	// +kubebuilder:validation:Optional
	PrivateKeyType *string `json:"privateKeyType,omitempty" tf:"private_key_type,omitempty"`

	// A field that allows clients to upload their own public key. If set, use this public key data to create a service account key for given service account. Please note, the expected format for this field is a base64 encoded X509_PEM.
	// +kubebuilder:validation:Optional
	PublicKeyData *string `json:"publicKeyData,omitempty" tf:"public_key_data,omitempty"`

	// +kubebuilder:validation:Optional
	PublicKeyType *string `json:"publicKeyType,omitempty" tf:"public_key_type,omitempty"`

	// The ID of the parent service account of the key. This can be a string in the format {ACCOUNT} or projects/{PROJECT_ID}/serviceAccounts/{ACCOUNT}, where {ACCOUNT} is the email address or unique id of the service account. If the {ACCOUNT} syntax is used, the project will be inferred from the provider's configuration.
	// +crossplane:generate:reference:type=ServiceAccount
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-jet-gcp/config/common.ExtractResourceID()
	// +kubebuilder:validation:Optional
	ServiceAccountID *string `json:"serviceAccountId,omitempty" tf:"service_account_id,omitempty"`

	// +kubebuilder:validation:Optional
	ServiceAccountIDRef *v1.Reference `json:"serviceAccountIdRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	ServiceAccountIDSelector *v1.Selector `json:"serviceAccountIdSelector,omitempty" tf:"-"`
}

// ServiceAccountKeySpec defines the desired state of ServiceAccountKey
type ServiceAccountKeySpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     ServiceAccountKeyParameters `json:"forProvider"`
}

// ServiceAccountKeyStatus defines the observed state of ServiceAccountKey.
type ServiceAccountKeyStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        ServiceAccountKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountKey is the Schema for the ServiceAccountKeys API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcpjet}
type ServiceAccountKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceAccountKeySpec   `json:"spec"`
	Status            ServiceAccountKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountKeyList contains a list of ServiceAccountKeys
type ServiceAccountKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccountKey `json:"items"`
}

// Repository type metadata.
var (
	ServiceAccountKey_Kind             = "ServiceAccountKey"
	ServiceAccountKey_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ServiceAccountKey_Kind}.String()
	ServiceAccountKey_KindAPIVersion   = ServiceAccountKey_Kind + "." + CRDGroupVersion.String()
	ServiceAccountKey_GroupVersionKind = CRDGroupVersion.WithKind(ServiceAccountKey_Kind)
)

func init() {
	SchemeBuilder.Register(&ServiceAccountKey{}, &ServiceAccountKeyList{})
}
