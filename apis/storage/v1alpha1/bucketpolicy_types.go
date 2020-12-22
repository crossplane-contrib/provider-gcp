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

	iamv1alpha1 "github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
)

// BucketPolicyParameters defines parameters for a desired KMS BucketPolicy
type BucketPolicyParameters struct {
	// Bucket: The RRN of the Bucket to which this BucketPolicy belongs.
	// +optional
	// +immutable
	Bucket *string `json:"bucket,omitempty"`

	// BucketRef references a Bucket and retrieves its URI
	// +optional
	// +immutable
	BucketRef *xpv1.Reference `json:"bucketRef,omitempty"`

	// BucketSelector selects a reference to a Bucket
	// +optional
	BucketSelector *xpv1.Selector `json:"bucketSelector,omitempty"`

	// Policy: An Identity and Access Management (IAM) policy, which
	// specifies access controls for Google Cloud resources.
	Policy iamv1alpha1.Policy `json:"policy"`
}

// BucketPolicyObservation is used to show the observed state of the
// BucketPolicy resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type BucketPolicyObservation struct {
	// Version: Specifies the format of the policy.
	//
	// Valid values are `0`, `1`, and `3`. Requests that specify an invalid
	// value
	// are rejected.
	//
	// Any operation that affects conditional role bindings must specify
	// version
	// `3`. This requirement applies to the following operations:
	//
	// * Getting a policy that includes a conditional role binding
	// * Adding a conditional role binding to a policy
	// * Changing a conditional role binding in a policy
	// * Removing any role binding, with or without a condition, from a
	// policy
	//   that includes conditions
	//
	// **Important:** If you use IAM Conditions, you must include the `etag`
	// field
	// whenever you call `setIamPolicy`. If you omit this field, then IAM
	// allows
	// you to overwrite a version `3` policy with a version `1` policy, and
	// all of
	// the conditions in the version `3` policy are lost.
	//
	// If a policy does not include any conditions, operations on that
	// policy may
	// specify any valid version or leave the field unset.
	Version int64 `json:"version,omitempty"`
}

// BucketPolicySpec defines the desired state of a
// BucketPolicy.
type BucketPolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       BucketPolicyParameters `json:"forProvider"`
}

// BucketPolicyStatus represents the observed state of a
// BucketPolicy.
type BucketPolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          BucketPolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// BucketPolicy is a managed resource that represents a Google KMS Crypto Key.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type BucketPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketPolicySpec   `json:"spec"`
	Status BucketPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketPolicyList contains a list of BucketPolicy types
type BucketPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketPolicy `json:"items"`
}
