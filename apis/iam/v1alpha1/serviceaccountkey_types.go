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

// ServiceAccountKeyParameters defines parameters for a desired IAM ServiceAccountKey
// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts.keys
//
type ServiceAccountKeyParameters struct {
	// KeyAlgorithm is an optional user-specified string that specifies the type of key and algorithm
	// to use for the key. The default is currently a 2048-bit RSA key. However this may change in the future.
	// Possible values:
	//   "KEY_ALG_UNSPECIFIED" - Not specified.
	//   "KEY_ALG_RSA_1024" - 1024-bit RSA key
	//   "KEY_ALG_RSA_2048" - 2048-bit RSA key
	// +optional
	// +immutable
	KeyAlgorithm *string `json:"keyAlgorithm,omitempty"`

	// PrivateKeyType is an optional specification of the output format of the generated private key.
	// The default value is TYPE_GOOGLE_CREDENTIALS_FILE, which corresponds to the Google Credentials File Format.
	// Possible values:
	//   "TYPE_UNSPECIFIED" - Not specified. Equivalent to TYPE_GOOGLE_CREDENTIALS_FILE.
	//   "TYPE_PKCS12_FILE" - Private key stored in a RFC7292 PKCS #12 document. Password for the PKCS #12 document is "notasecret".
	//   "TYPE_GOOGLE_CREDENTIALS_FILE" - Google Credentials File format.
	// +optional
	// +immutable
	PrivateKeyType *string `json:"privateKeyType,omitempty"`

	// PublicKeyType is an optional specification of the output format for the associated public key.
	// The default value is TYPE_RAW_PUBLIC_KEY.
	// Possible values:
	//   "TYPE_NONE" - Not specified. Public key is not retrieved via Google Cloud API.
	//   "TYPE_X509_PEM_FILE" - X509 PEM format.
	//   "TYPE_RAW_PUBLIC_KEY" - Raw public key.
	// +optional
	// +kubebuilder:default=TYPE_RAW_PUBLIC_KEY
	PublicKeyType *string `json:"publicKeyType,omitempty"`

	// ServiceAccountRef is a reference to a ServiceAccount which this policy is associated with
	ServiceAccountReferer `json:",inline"`
}

// ServiceAccountKeyObservation is used to show the observed state of the
// ServiceAccountKey resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type ServiceAccountKeyObservation struct {
	// Name is the resource name of the service account key in the following format:
	// projects/{PROJECT_ID}/serviceAccounts/{ACCOUNT}/keys/{external-name}.
	// part of https://godoc.org/google.golang.org/genproto/googleapis/iam/admin/v1#ServiceAccountKey
	Name string `json:"name,omitempty"`

	// KeyID is the generated unique & stable key id for the service account key.
	KeyID string `json:"keyId,omitempty"`

	// PrivateKeyType is the output format for the generated private key. Only set in keys.create responses.
	// Determines the encoding for the private key stored in the "connection" secret.
	PrivateKeyType string `json:"privateKeyType,omitempty"`

	// KeyAlgorithm is the key algorithm & possibly key size used for public/private key pair generation.
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`

	// ValidAfterTime is the timestamp after which this key can be used in RFC3339 UTC "Zulu" format.
	ValidAfterTime string `json:"validAfterTime,omitempty"`

	// ValidBeforeTime is the timestamp before which this key can be used in RFC3339 UTC "Zulu" format.
	ValidBeforeTime string `json:"validBeforeTime,omitempty"`

	// KeyOrigin is the origin of the key.
	// Possible values:
	//   "ORIGIN_UNSPECIFIED" - Unspecified key origin.
	//   "USER_PROVIDED" - Key is provided by user.
	//   "GOOGLE_PROVIDED" - Key is provided by Google.
	KeyOrigin string `json:"keyOrigin,omitempty"`

	// KeyType is the type of the key.
	// Possible values:
	//   "KEY_TYPE_UNSPECIFIED" - Unspecified key type.
	//   "USER_MANAGED" - User-managed key (managed and rotated by the user).
	//   "SYSTEM_MANAGED" - System-managed key (managed and rotated by Google).
	KeyType string `json:"keyType,omitempty"`
}

// ServiceAccountKeySpec defines the desired state of a ServiceAccountKey.
type ServiceAccountKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Todo(turkenh): Move to crossplane runtime ResourceSpec
	// PublishConnectionDetailsTo specifies the connection secret config which
	// contains a name, metadata and a reference to secret store config to
	// which any connection details for this managed resource should be written.
	// Connection details frequently include the endpoint, username,
	// and password required to connect to the managed resource.
	// +optional
	PublishConnectionDetailsTo *xpv1.PublishConnectionDetailsTo `json:"publishConnectionDetailsTo,omitempty"`

	ForProvider ServiceAccountKeyParameters `json:"forProvider"`
}

// ServiceAccountKeyStatus represents the observed state of a ServiceAccountKey.
type ServiceAccountKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ServiceAccountKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountKey is a managed resource that represents a Google IAM Service Account Key.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="KEY_ID",type="string",JSONPath=".status.atProvider.keyId"
// +kubebuilder:printcolumn:name="CREATED_AT",type="string",JSONPath=".status.atProvider.validAfterTime"
// +kubebuilder:printcolumn:name="EXPIRES_AT",type="boolean",JSONPath=".status.atProvider.validBeforeTime"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type ServiceAccountKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceAccountKeySpec   `json:"spec"`
	Status ServiceAccountKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountKeyList contains a list of ServiceAccountKey types
type ServiceAccountKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccountKey `json:"items"`
}

// Todo(turkenh): To be generated with AngryJet

// SetPublishConnectionDetailsTo sets PublishConnectionDetailsTo
func (mg *ServiceAccountKey) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// GetPublishConnectionDetailsTo returns xpv1.PublishConnectionDetailsTo
func (mg *ServiceAccountKey) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}
