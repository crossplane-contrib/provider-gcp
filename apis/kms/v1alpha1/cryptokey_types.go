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

// CryptoKeyParameters defines parameters for a desired KMS CryptoKey
// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys
type CryptoKeyParameters struct {
	// KeyRing: The RRN of the KeyRing to which this CryptoKey belongs,
	// provided by the client when initially creating the CryptoKey.
	// +optional
	// +immutable
	KeyRing *string `json:"keyRing,omitempty"`

	// KeyRingRef references a KeyRing and retrieves its URI
	// +optional
	// +immutable
	KeyRingRef *xpv1.Reference `json:"keyRingRef,omitempty"`

	// KeyRingSelector selects a reference to a KeyRing
	// +optional
	// +immutable
	KeyRingSelector *xpv1.Selector `json:"keyRingSelector,omitempty"`

	// Labels: Labels with user-defined metadata. For more information,
	// see
	// [Labeling Keys](/kms/docs/labeling-keys).
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Purpose: Immutable. The immutable purpose of this CryptoKey.
	//
	// Possible values:
	//   "CRYPTO_KEY_PURPOSE_UNSPECIFIED" - Not specified.
	//   "ENCRYPT_DECRYPT" - CryptoKeys with this purpose may be used
	// with
	// Encrypt and
	// Decrypt.
	//   "ASYMMETRIC_SIGN" - CryptoKeys with this purpose may be used
	// with
	// AsymmetricSign and
	// GetPublicKey.
	//   "ASYMMETRIC_DECRYPT" - CryptoKeys with this purpose may be used
	// with
	// AsymmetricDecrypt and
	// GetPublicKey.
	// +immutable
	// +kubebuilder:validation:Enum=ENCRYPT_DECRYPT;ASYMMETRIC_SIGN;ASYMMETRIC_DECRYPT
	Purpose string `json:"purpose"`

	// RotationPeriod: next_rotation_time will be advanced by this period
	// when the service
	// automatically rotates a key. Must be at least 24 hours and at
	// most
	// 876,000 hours.
	//
	// If rotation_period is set, next_rotation_time must also be set.
	//
	// Keys with purpose
	// ENCRYPT_DECRYPT support
	// automatic rotation. For other keys, this field must be omitted.
	// +optional
	RotationPeriod *string `json:"rotationPeriod,omitempty"`

	// NextRotationTime: At next_rotation_time, the Key Management Service
	// will automatically:
	//
	// 1. Create a new version of this CryptoKey.
	// 2. Mark the new version as primary.
	//
	// Key rotations performed manually via
	// CreateCryptoKeyVersion and
	// UpdateCryptoKeyPrimaryVersion
	// do not affect next_rotation_time.
	//
	// Keys with purpose
	// ENCRYPT_DECRYPT support
	// automatic rotation. For other keys, this field must be omitted.
	// +optional
	NextRotationTime *string `json:"nextRotationTime,omitempty"`

	// VersionTemplate: A template describing settings for new
	// CryptoKeyVersion instances.
	// The properties of new CryptoKeyVersion instances created by
	// either
	// CreateCryptoKeyVersion or
	// auto-rotation are controlled by this template.
	// +optional
	VersionTemplate *CryptoKeyVersionTemplate `json:"versionTemplate,omitempty"`
}

// CryptoKeyObservation is used to show the observed state of the
// CryptoKey resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type CryptoKeyObservation struct {
	// CreateTime: Output only. The time at which this CryptoKey was
	// created.
	CreateTime string `json:"createTime,omitempty"`

	// Name: Output only. The resource name for this CryptoKey in the
	// format
	// `projects/*/locations/*/keyRings/*/cryptoKeys/*`.
	Name string `json:"name,omitempty"`

	// NextRotationTime: At next_rotation_time, the Key Management Service
	// will automatically:
	//
	// 1. Create a new version of this CryptoKey.
	// 2. Mark the new version as primary.
	//
	// Key rotations performed manually via
	// CreateCryptoKeyVersion and
	// UpdateCryptoKeyPrimaryVersion
	// do not affect next_rotation_time.
	//
	// Keys with purpose
	// ENCRYPT_DECRYPT support
	// automatic rotation. For other keys, this field must be omitted.
	NextRotationTime string `json:"nextRotationTime,omitempty"`

	// Primary: Output only. A copy of the "primary" CryptoKeyVersion that
	// will be used
	// by Encrypt when this CryptoKey is given
	// in EncryptRequest.name.
	//
	// The CryptoKey's primary version can be updated
	// via
	// UpdateCryptoKeyPrimaryVersion.
	//
	// Keys with purpose
	// ENCRYPT_DECRYPT may have a
	// primary. For other keys, this field will be omitted.
	Primary *CryptoKeyVersion `json:"primary,omitempty"`
}

// A CryptoKeyVersion represents an individual cryptographic key, and the
// associated key material.
//
// An ENABLED version can be used for cryptographic operations.
//
// For security reasons, the raw cryptographic key material represented
// by a
// CryptoKeyVersion can never be viewed or exported. It can only be used
// to
// encrypt, decrypt, or sign data when an authorized user or application
// invokes
// Cloud KMS.
type CryptoKeyVersion struct {
	// Algorithm: Output only. The CryptoKeyVersionAlgorithm that
	// this
	// CryptoKeyVersion supports.
	//
	// Possible values:
	//   "CRYPTO_KEY_VERSION_ALGORITHM_UNSPECIFIED" - Not specified.
	//   "GOOGLE_SYMMETRIC_ENCRYPTION" - Creates symmetric encryption keys.
	//   "RSA_SIGN_PSS_2048_SHA256" - RSASSA-PSS 2048 bit key with a SHA256
	// digest.
	//   "RSA_SIGN_PSS_3072_SHA256" - RSASSA-PSS 3072 bit key with a SHA256
	// digest.
	//   "RSA_SIGN_PSS_4096_SHA256" - RSASSA-PSS 4096 bit key with a SHA256
	// digest.
	//   "RSA_SIGN_PSS_4096_SHA512" - RSASSA-PSS 4096 bit key with a SHA512
	// digest.
	//   "RSA_SIGN_PKCS1_2048_SHA256" - RSASSA-PKCS1-v1_5 with a 2048 bit
	// key and a SHA256 digest.
	//   "RSA_SIGN_PKCS1_3072_SHA256" - RSASSA-PKCS1-v1_5 with a 3072 bit
	// key and a SHA256 digest.
	//   "RSA_SIGN_PKCS1_4096_SHA256" - RSASSA-PKCS1-v1_5 with a 4096 bit
	// key and a SHA256 digest.
	//   "RSA_SIGN_PKCS1_4096_SHA512" - RSASSA-PKCS1-v1_5 with a 4096 bit
	// key and a SHA512 digest.
	//   "RSA_DECRYPT_OAEP_2048_SHA256" - RSAES-OAEP 2048 bit key with a
	// SHA256 digest.
	//   "RSA_DECRYPT_OAEP_3072_SHA256" - RSAES-OAEP 3072 bit key with a
	// SHA256 digest.
	//   "RSA_DECRYPT_OAEP_4096_SHA256" - RSAES-OAEP 4096 bit key with a
	// SHA256 digest.
	//   "RSA_DECRYPT_OAEP_4096_SHA512" - RSAES-OAEP 4096 bit key with a
	// SHA512 digest.
	//   "EC_SIGN_P256_SHA256" - ECDSA on the NIST P-256 curve with a SHA256
	// digest.
	//   "EC_SIGN_P384_SHA384" - ECDSA on the NIST P-384 curve with a SHA384
	// digest.
	//   "EXTERNAL_SYMMETRIC_ENCRYPTION" - Algorithm representing symmetric
	// encryption by an external key manager.
	Algorithm string `json:"algorithm,omitempty"`

	// Attestation: Output only. Statement that was generated and signed by
	// the HSM at key
	// creation time. Use this statement to verify attributes of the key as
	// stored
	// on the HSM, independently of Google. Only provided for key versions
	// with
	// protection_level HSM.
	Attestation *KeyOperationAttestation `json:"attestation,omitempty"`

	// CreateTime: Output only. The time at which this CryptoKeyVersion was
	// created.
	CreateTime string `json:"createTime,omitempty"`

	// DestroyEventTime: Output only. The time this CryptoKeyVersion's key
	// material was
	// destroyed. Only present if state is
	// DESTROYED.
	DestroyEventTime string `json:"destroyEventTime,omitempty"`

	// DestroyTime: Output only. The time this CryptoKeyVersion's key
	// material is scheduled
	// for destruction. Only present if state is
	// DESTROY_SCHEDULED.
	DestroyTime string `json:"destroyTime,omitempty"`

	// ExternalProtectionLevelOptions: ExternalProtectionLevelOptions stores
	// a group of additional fields for
	// configuring a CryptoKeyVersion that are specific to the
	// EXTERNAL protection level.
	ExternalProtectionLevelOptions *ExternalProtectionLevelOptions `json:"externalProtectionLevelOptions,omitempty"`

	// GenerateTime: Output only. The time this CryptoKeyVersion's key
	// material was
	// generated.
	GenerateTime string `json:"generateTime,omitempty"`

	// ImportFailureReason: Output only. The root cause of an import
	// failure. Only present if
	// state is
	// IMPORT_FAILED.
	ImportFailureReason string `json:"importFailureReason,omitempty"`

	// ImportJob: Output only. The name of the ImportJob used to import
	// this
	// CryptoKeyVersion. Only present if the underlying key material
	// was
	// imported.
	ImportJob string `json:"importJob,omitempty"`

	// ImportTime: Output only. The time at which this CryptoKeyVersion's
	// key material
	// was imported.
	ImportTime string `json:"importTime,omitempty"`

	// Name: Output only. The resource name for this CryptoKeyVersion in the
	// format
	// `projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersio
	// ns/*`.
	Name string `json:"name,omitempty"`

	// ProtectionLevel: Output only. The ProtectionLevel describing how
	// crypto operations are
	// performed with this CryptoKeyVersion.
	//
	// Possible values:
	//   "PROTECTION_LEVEL_UNSPECIFIED" - Not specified.
	//   "SOFTWARE" - Crypto operations are performed in software.
	//   "HSM" - Crypto operations are performed in a Hardware Security
	// Module.
	//   "EXTERNAL" - Crypto operations are performed by an external key
	// manager.
	ProtectionLevel string `json:"protectionLevel,omitempty"`

	// State: The current state of the CryptoKeyVersion.
	//
	// Possible values:
	//   "CRYPTO_KEY_VERSION_STATE_UNSPECIFIED" - Not specified.
	//   "PENDING_GENERATION" - This version is still being generated. It
	// may not be used, enabled,
	// disabled, or destroyed yet. Cloud KMS will automatically mark
	// this
	// version ENABLED as soon as the version is ready.
	//   "ENABLED" - This version may be used for cryptographic operations.
	//   "DISABLED" - This version may not be used, but the key material is
	// still available,
	// and the version can be placed back into the ENABLED state.
	//   "DESTROYED" - This version is destroyed, and the key material is no
	// longer stored.
	// A version may not leave this state once entered.
	//   "DESTROY_SCHEDULED" - This version is scheduled for destruction,
	// and will be destroyed soon.
	// Call
	// RestoreCryptoKeyVersion
	// to put it back into the DISABLED state.
	//   "PENDING_IMPORT" - This version is still being imported. It may not
	// be used, enabled,
	// disabled, or destroyed yet. Cloud KMS will automatically mark
	// this
	// version ENABLED as soon as the version is ready.
	//   "IMPORT_FAILED" - This version was not imported successfully. It
	// may not be used, enabled,
	// disabled, or destroyed. The submitted key material has been
	// discarded.
	// Additional details can be found
	// in
	// CryptoKeyVersion.import_failure_reason.
	State string `json:"state,omitempty"`
}

// A CryptoKeyVersionTemplate specifies the properties to use when creating
// a new CryptoKeyVersion, either manually with CreateCryptoKeyVersion or
// automatically as a result of auto-rotation.
type CryptoKeyVersionTemplate struct {
	// Algorithm: Required. Algorithm to use
	// when creating a CryptoKeyVersion based on this template.
	//
	// For backwards compatibility, GOOGLE_SYMMETRIC_ENCRYPTION is implied
	// if both
	// this field is omitted and CryptoKey.purpose is
	// ENCRYPT_DECRYPT.
	//
	// Possible values:
	//   "CRYPTO_KEY_VERSION_ALGORITHM_UNSPECIFIED" - Not specified.
	//   "GOOGLE_SYMMETRIC_ENCRYPTION" - Creates symmetric encryption keys.
	//   "RSA_SIGN_PSS_2048_SHA256" - RSASSA-PSS 2048 bit key with a SHA256
	// digest.
	//   "RSA_SIGN_PSS_3072_SHA256" - RSASSA-PSS 3072 bit key with a SHA256
	// digest.
	//   "RSA_SIGN_PSS_4096_SHA256" - RSASSA-PSS 4096 bit key with a SHA256
	// digest.
	//   "RSA_SIGN_PSS_4096_SHA512" - RSASSA-PSS 4096 bit key with a SHA512
	// digest.
	//   "RSA_SIGN_PKCS1_2048_SHA256" - RSASSA-PKCS1-v1_5 with a 2048 bit
	// key and a SHA256 digest.
	//   "RSA_SIGN_PKCS1_3072_SHA256" - RSASSA-PKCS1-v1_5 with a 3072 bit
	// key and a SHA256 digest.
	//   "RSA_SIGN_PKCS1_4096_SHA256" - RSASSA-PKCS1-v1_5 with a 4096 bit
	// key and a SHA256 digest.
	//   "RSA_SIGN_PKCS1_4096_SHA512" - RSASSA-PKCS1-v1_5 with a 4096 bit
	// key and a SHA512 digest.
	//   "RSA_DECRYPT_OAEP_2048_SHA256" - RSAES-OAEP 2048 bit key with a
	// SHA256 digest.
	//   "RSA_DECRYPT_OAEP_3072_SHA256" - RSAES-OAEP 3072 bit key with a
	// SHA256 digest.
	//   "RSA_DECRYPT_OAEP_4096_SHA256" - RSAES-OAEP 4096 bit key with a
	// SHA256 digest.
	//   "RSA_DECRYPT_OAEP_4096_SHA512" - RSAES-OAEP 4096 bit key with a
	// SHA512 digest.
	//   "EC_SIGN_P256_SHA256" - ECDSA on the NIST P-256 curve with a SHA256
	// digest.
	//   "EC_SIGN_P384_SHA384" - ECDSA on the NIST P-384 curve with a SHA384
	// digest.
	//   "EXTERNAL_SYMMETRIC_ENCRYPTION" - Algorithm representing symmetric
	// encryption by an external key manager.
	// +optional
	Algorithm *string `json:"algorithm,omitempty"`

	// ProtectionLevel: ProtectionLevel to use when creating a
	// CryptoKeyVersion based on
	// this template. Immutable. Defaults to SOFTWARE.
	//
	// Possible values:
	//   "PROTECTION_LEVEL_UNSPECIFIED" - Not specified.
	//   "SOFTWARE" - Crypto operations are performed in software.
	//   "HSM" - Crypto operations are performed in a Hardware Security
	// Module.
	//   "EXTERNAL" - Crypto operations are performed by an external key
	// manager.
	// +optional
	// +kubebuilder:validation:Enum=SOFTWARE;HSM;EXTERNAL
	ProtectionLevel *string `json:"protectionLevel,omitempty"`
}

// ExternalProtectionLevelOptions stores a group of additional fields for
// configuring a CryptoKeyVersion that are specific to the EXTERNAL protection level.
type ExternalProtectionLevelOptions struct {
	// ExternalKeyUri: The URI for an external resource that this
	// CryptoKeyVersion represents.
	ExternalKeyUri string `json:"externalKeyUri,omitempty"` // nolint:golint
}

// KeyOperationAttestation contains an HSM-generated attestation about
// a key operation. For more information, see [Verifying attestations]
// (https://cloud.google.com/kms/docs/attest-key).
type KeyOperationAttestation struct {
	// Content: Output only. The attestation data provided by the HSM when
	// the key
	// operation was performed.
	Content string `json:"content,omitempty"`

	// Format: Output only. The format of the attestation data.
	//
	// Possible values:
	//   "ATTESTATION_FORMAT_UNSPECIFIED" - Not specified.
	//   "CAVIUM_V1_COMPRESSED" - Cavium HSM attestation compressed with
	// gzip. Note that this format is
	// defined by Cavium and subject to change at any time.
	//   "CAVIUM_V2_COMPRESSED" - Cavium HSM attestation V2 compressed with
	// gzip. This is a new format
	// introduced in Cavium's version 3.2-08.
	Format string `json:"format,omitempty"`
}

// CryptoKeySpec defines the desired state of a
// CryptoKey.
type CryptoKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CryptoKeyParameters `json:"forProvider"`
}

// CryptoKeyStatus represents the observed state of a
// CryptoKey.
type CryptoKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CryptoKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// CryptoKey is a managed resource that represents a Google KMS Crypto Key.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="PURPOSE",type="string",JSONPath=".spec.forProvider.purpose"
// +kubebuilder:resource:scope=Cluster
type CryptoKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CryptoKeySpec   `json:"spec"`
	Status CryptoKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CryptoKeyList contains a list of CryptoKey types
type CryptoKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CryptoKey `json:"items"`
}
