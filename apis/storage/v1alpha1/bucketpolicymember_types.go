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

// BucketPolicyMemberParameters defines parameters for a desired KMS BucketPolicyMember
type BucketPolicyMemberParameters struct {
	// Bucket: The RRN of the Bucket to which this BucketPolicyMember belongs.
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

	// Role: Role that is assigned to `members`.
	// For example, `roles/viewer`, `roles/editor`, or `roles/owner`.
	// +immutable
	Role string `json:"role"`

	// Member: Specifies the identity requesting access for a Cloud
	// Platform resource.
	// `member` can have the following values:
	//
	// * `allUsers`: A special identifier that represents anyone who is
	//    on the internet; with or without a Google account.
	//
	// * `allAuthenticatedUsers`: A special identifier that represents
	// anyone
	//    who is authenticated with a Google account or a service
	// account.
	//
	// * `user:{emailid}`: An email address that represents a specific
	// Google
	//    account. For example, `alice@example.com` .
	//
	//
	// * `serviceAccount:{emailid}`: An email address that represents a
	// service
	//    account. For example,
	// `my-other-app@appspot.gserviceaccount.com`.
	//
	// * `group:{emailid}`: An email address that represents a Google
	// group.
	//    For example, `admins@example.com`.
	//
	// * `deleted:user:{emailid}?uid={uniqueid}`: An email address (plus
	// unique
	//    identifier) representing a user that has been recently deleted.
	// For
	//    example, `alice@example.com?uid=123456789012345678901`. If the
	// user is
	//    recovered, this value reverts to `user:{emailid}` and the
	// recovered user
	//    retains the role in the binding.
	//
	// * `deleted:serviceAccount:{emailid}?uid={uniqueid}`: An email address
	// (plus
	//    unique identifier) representing a service account that has been
	// recently
	//    deleted. For example,
	//
	// `my-other-app@appspot.gserviceaccount.com?uid=123456789012345678901`.
	//
	//    If the service account is undeleted, this value reverts to
	//    `serviceAccount:{emailid}` and the undeleted service account
	// retains the
	//    role in the binding.
	//
	// * `deleted:group:{emailid}?uid={uniqueid}`: An email address (plus
	// unique
	//    identifier) representing a Google group that has been recently
	//    deleted. For example,
	// `admins@example.com?uid=123456789012345678901`. If
	//    the group is recovered, this value reverts to `group:{emailid}`
	// and the
	//    recovered group retains the role in the binding.
	//
	//
	// * `domain:{domain}`: The G Suite domain (primary) that represents all
	// the
	//    users of that domain. For example, `google.com` or
	// `example.com`.
	//
	//
	// +optional
	// +immutable
	Member *string `json:"member,omitempty"`

	// ServiceAccountMemberRef is reference to ServiceAccount used to set
	// the Member.
	// +optional
	// +immutable
	ServiceAccountMemberRef *xpv1.Reference `json:"serviceAccountMemberRef,omitempty"`

	// ServiceAccountMemberSelector selects reference to ServiceAccount used
	// to set the Member.
	// +optional
	// +immutable
	ServiceAccountMemberSelector *xpv1.Selector `json:"serviceAccountMemberSelector,omitempty"`
}

// BucketPolicyMemberSpec defines the desired state of a
// BucketPolicyMember.
type BucketPolicyMemberSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       BucketPolicyMemberParameters `json:"forProvider"`
}

// BucketPolicyMemberStatus represents the observed state of a
// BucketPolicyMember.
type BucketPolicyMemberStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// BucketPolicyMember is a managed resource that represents a Google KMS Crypto Key.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type BucketPolicyMember struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketPolicyMemberSpec   `json:"spec"`
	Status BucketPolicyMemberStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketPolicyMemberList contains a list of BucketPolicyMember types
type BucketPolicyMemberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketPolicyMember `json:"items"`
}
