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
	Policy Policy `json:"policy,omitempty"`
}

// CryptoKeyPolicyObservation is used to show the observed state of the
// CryptoKeyPolicy resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type CryptoKeyPolicyObservation struct {
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

// Policy An Identity and Access Management (IAM) policy, which
// specifies access
// controls for Google Cloud resources.
//
//
// A `Policy` is a collection of `bindings`. A `binding` binds one or
// more
// `members` to a single `role`. Members can be user accounts, service
// accounts,
// Google groups, and domains (such as G Suite). A `role` is a named
// list of
// permissions; each `role` can be an IAM predefined role or a
// user-created
// custom role.
//
// Optionally, a `binding` can specify a `condition`, which is a
// logical
// expression that allows access to a resource only if the expression
// evaluates
// to `true`. A condition can add constraints based on attributes of
// the
// request, the resource, or both.
//
// **JSON example:**
//
//     {
//       "bindings": [
//         {
//           "role": "roles/resourcemanager.organizationAdmin",
//           "members": [
//             "user:mike@example.com",
//             "group:admins@example.com",
//             "domain:google.com",
//
// "serviceAccount:my-project-id@appspot.gserviceaccount.com"
//           ]
//         },
//         {
//           "role": "roles/resourcemanager.organizationViewer",
//           "members": ["user:eve@example.com"],
//           "condition": {
//             "title": "expirable access",
//             "description": "Does not grant access after Sep 2020",
//             "expression": "request.time <
// timestamp('2020-10-01T00:00:00.000Z')",
//           }
//         }
//       ],
//       "etag": "BwWWja0YfJA=",
//       "version": 3
//     }
//
// **YAML example:**
//
//     bindings:
//     - members:
//       - user:mike@example.com
//       - group:admins@example.com
//       - domain:google.com
//       - serviceAccount:my-project-id@appspot.gserviceaccount.com
//       role: roles/resourcemanager.organizationAdmin
//     - members:
//       - user:eve@example.com
//       role: roles/resourcemanager.organizationViewer
//       condition:
//         title: expirable access
//         description: Does not grant access after Sep 2020
//         expression: request.time <
// timestamp('2020-10-01T00:00:00.000Z')
//     - etag: BwWWja0YfJA=
//     - version: 3
//
// For a description of IAM and its features, see the
// [IAM documentation](https://cloud.google.com/iam/docs/).
type Policy struct {
	// AuditConfigs: Specifies cloud audit logging configuration for this
	// policy.
	AuditConfigs []*AuditConfig `json:"auditConfigs,omitempty"`

	// Bindings: Associates a list of `members` to a `role`. Optionally, may
	// specify a
	// `condition` that determines how and when the `bindings` are applied.
	// Each
	// of the `bindings` must contain at least one member.
	Bindings []*Binding `json:"bindings,omitempty"`
}

// AuditConfig Specifies the audit configuration for a service.
// The configuration determines which permission types are logged, and
// what
// identities, if any, are exempted from logging.
// An AuditConfig must have one or more AuditLogConfigs.
//
// If there are AuditConfigs for both `allServices` and a specific
// service,
// the union of the two AuditConfigs is used for that service: the
// log_types
// specified in each AuditConfig are enabled, and the exempted_members
// in each
// AuditLogConfig are exempted.
//
// Example Policy with multiple AuditConfigs:
//
//     {
//       "audit_configs": [
//         {
//           "service": "allServices"
//           "audit_log_configs": [
//             {
//               "log_type": "DATA_READ",
//               "exempted_members": [
//                 "user:jose@example.com"
//               ]
//             },
//             {
//               "log_type": "DATA_WRITE",
//             },
//             {
//               "log_type": "ADMIN_READ",
//             }
//           ]
//         },
//         {
//           "service": "sampleservice.googleapis.com"
//           "audit_log_configs": [
//             {
//               "log_type": "DATA_READ",
//             },
//             {
//               "log_type": "DATA_WRITE",
//               "exempted_members": [
//                 "user:aliya@example.com"
//               ]
//             }
//           ]
//         }
//       ]
//     }
//
// For sampleservice, this policy enables DATA_READ, DATA_WRITE and
// ADMIN_READ
// logging. It also exempts jose@example.com from DATA_READ logging,
// and
// aliya@example.com from DATA_WRITE logging.
type AuditConfig struct {
	// AuditLogConfigs: The configuration for logging of each type of
	// permission.
	AuditLogConfigs []*AuditLogConfig `json:"auditLogConfigs,omitempty"`

	// Service: Specifies a service that will be enabled for audit
	// logging.
	// For example, `storage.googleapis.com`,
	// `cloudsql.googleapis.com`.
	// `allServices` is a special value that covers all services.
	Service string `json:"service,omitempty"`
}

// AuditLogConfig Provides the configuration for logging a type of
// permissions.
// Example:
//
//     {
//       "audit_log_configs": [
//         {
//           "log_type": "DATA_READ",
//           "exempted_members": [
//             "user:jose@example.com"
//           ]
//         },
//         {
//           "log_type": "DATA_WRITE",
//         }
//       ]
//     }
//
// This enables 'DATA_READ' and 'DATA_WRITE' logging, while
// exempting
// jose@example.com from DATA_READ logging.
type AuditLogConfig struct {
	// ExemptedMembers: Specifies the identities that do not cause logging
	// for this type of
	// permission.
	// Follows the same format of Binding.members.
	ExemptedMembers []string `json:"exemptedMembers,omitempty"`

	// LogType: The log type that this config enables.
	//
	// Possible values:
	//   "LOG_TYPE_UNSPECIFIED" - Default case. Should never be this.
	//   "ADMIN_READ" - Admin reads. Example: CloudIAM getIamPolicy
	//   "DATA_WRITE" - Data writes. Example: CloudSQL Users create
	//   "DATA_READ" - Data reads. Example: CloudSQL Users list
	LogType string `json:"logType,omitempty"`
}

// Binding Associates `members` with a `role`.
type Binding struct {
	// Condition: The condition that is associated with this binding.
	// NOTE: An unsatisfied condition will not allow user access via
	// current
	// binding. Different bindings, including their conditions, are
	// examined
	// independently.
	Condition *Expr `json:"condition,omitempty"`

	// Members: Specifies the identities requesting access for a Cloud
	// Platform resource.
	// `members` can have the following values:
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
	Members []string `json:"members,omitempty"`

	// ServiceAccountMemberRefs are references to ServiceAccounts used to set
	// the Members.
	// +optional
	ServiceAccountMemberRefs []xpv1.Reference `json:"serviceAccountMemberRefs,omitempty"`

	// ServiceAccountMemberRefSelector selects references to ServiceAccounts used
	// to set the Members.
	// +optional
	ServiceAccountMemberRefSelector *xpv1.Selector `json:"serviceAccountMemberRefSelector,omitempty"`

	// Role: Role that is assigned to `members`.
	// For example, `roles/viewer`, `roles/editor`, or `roles/owner`.
	Role string `json:"role,omitempty"`
}

// Expr Represents a textual expression in the Common Expression
// Language (CEL)
// syntax. CEL is a C-like expression language. The syntax and semantics
// of CEL
// are documented at https://github.com/google/cel-spec.
//
// Example (Comparison):
//
//     title: "Summary size limit"
//     description: "Determines if a summary is less than 100 chars"
//     expression: "document.summary.size() < 100"
//
// Example (Equality):
//
//     title: "Requestor is owner"
//     description: "Determines if requestor is the document owner"
//     expression: "document.owner ==
// request.auth.claims.email"
//
// Example (Logic):
//
//     title: "Public documents"
//     description: "Determine whether the document should be publicly
// visible"
//     expression: "document.type != 'private' && document.type !=
// 'internal'"
//
// Example (Data Manipulation):
//
//     title: "Notification string"
//     description: "Create a notification string with a timestamp."
//     expression: "'New message received at ' +
// string(document.create_time)"
//
// The exact variables and functions that may be referenced within an
// expression
// are determined by the service that evaluates it. See the
// service
// documentation for additional information.
type Expr struct {
	// Description: Optional. Description of the expression. This is a
	// longer text which
	// describes the expression, e.g. when hovered over it in a UI.
	Description string `json:"description,omitempty"`

	// Expression: Textual representation of an expression in Common
	// Expression Language
	// syntax.
	Expression string `json:"expression,omitempty"`

	// Location: Optional. String indicating the location of the expression
	// for error
	// reporting, e.g. a file name and a position in the file.
	Location string `json:"location,omitempty"`

	// Title: Optional. Title for the expression, i.e. a short string
	// describing
	// its purpose. This can be used e.g. in UIs which allow to enter
	// the
	// expression.
	Title string `json:"title,omitempty"`
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
	AtProvider          CryptoKeyPolicyObservation `json:"atProvider,omitempty"`
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
