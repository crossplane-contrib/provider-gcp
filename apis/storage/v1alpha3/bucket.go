/*
Copyright 2019 The Crossplane Authors.

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

package v1alpha3

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BucketParameters define the desired state of the Google Cloud Storage Bucket.
type BucketParameters struct {
	// Location: The location of the bucket. Object data for objects in the
	// bucket resides in physical storage within this region. Defaults to
	// US. See the developer's guide for the authoritative list.
	Location string `json:"location"`

	// TODO: different resource
	//Acl []*BucketAccessControl `json:"acl,omitempty"`

	// Billing: The bucket's billing configuration.
	// +optional
	Billing *BucketBilling `json:"billing,omitempty"`

	// Cors: The bucket's Cross-Origin Resource Sharing (CORS)
	// configuration.
	// +optional
	Cors []*BucketCORS `json:"cors,omitempty"`

	// DefaultEventBasedHold: The default value for event-based hold on
	// newly created objects in this bucket. Event-based hold is a way to
	// retain objects indefinitely until an event occurs, signified by the
	// hold's release. After being released, such objects will be subject to
	// bucket-level retention (if any). One sample use case of this flag is
	// for banks to hold loan documents for at least 3 years after loan is
	// paid in full. Here, bucket-level retention is 3 years and the event
	// is loan being paid in full. In this example, these objects will be
	// held intact for any number of years until the event has occurred
	// (event-based hold on the object is released) and then 3 more years
	// after that. That means retention duration of the objects begins from
	// the moment event-based hold transitioned from true to false. Objects
	// under event-based hold cannot be deleted, overwritten or archived
	// until the hold is removed.
	// +optional
	DefaultEventBasedHold *bool `json:"defaultEventBasedHold,omitempty"`

	// TODO: different resource
	// DefaultObjectAcl: Default access controls to apply to new objects
	// when no ACL is provided.
	// DefaultObjectAcl []*ObjectAccessControl `json:"defaultObjectAcl,omitempty"`

	// Encryption: Encryption configuration for a bucket.
	// +optional
	Encryption *BucketEncryption `json:"encryption,omitempty"`

	// IamConfiguration: The bucket's IAM configuration.
	// +optional
	IamConfiguration *BucketIamConfiguration `json:"iamConfiguration,omitempty"`

	// Labels: User-provided labels, in key/value pairs.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Lifecycle: The bucket's lifecycle configuration. See lifecycle
	// management for more information.
	// +optional
	Lifecycle *BucketLifecycle `json:"lifecycle,omitempty"`

	// LocationType: The type of the bucket location.
	// +optional
	// +immutable
	LocationType *string `json:"locationType,omitempty"`

	// Logging: The bucket's logging configuration, which defines the
	// destination bucket and optional name prefix for the current bucket's
	// logs.
	// +optional
	Logging *BucketLogging `json:"logging,omitempty"`

	// RetentionPolicy: The bucket's retention policy. The retention policy
	// enforces a minimum retention time for all objects contained in the
	// bucket, based on their creation time. Any attempt to overwrite or
	// delete objects younger than the retention period will result in a
	// PERMISSION_DENIED error. An unlocked retention policy can be modified
	// or removed from the bucket via a storage.buckets.update operation. A
	// locked retention policy cannot be removed or shortened in duration
	// for the lifetime of the bucket. Attempting to remove or decrease
	// period of a locked retention policy will result in a
	// PERMISSION_DENIED error.
	// +optional
	RetentionPolicy *BucketRetentionPolicy `json:"retentionPolicy,omitempty"`

	// StorageClass: The bucket's default storage class, used whenever no
	// storageClass is specified for a newly-created object. This defines
	// how objects in the bucket are stored and determines the SLA and the
	// cost of storage. Values include MULTI_REGIONAL, REGIONAL, STANDARD,
	// NEARLINE, COLDLINE, and DURABLE_REDUCED_AVAILABILITY. If this value
	// is not specified when the bucket is created, it will default to
	// STANDARD. For more information, see storage classes.
	// +optional
	// +immutable
	StorageClass *string `json:"storageClass,omitempty"`

	// Versioning: The bucket's versioning configuration.
	// +optional
	Versioning *BucketVersioning `json:"versioning,omitempty"`

	// Website: The bucket's website configuration, controlling how the
	// service behaves when accessing bucket contents as a web site. See the
	// Static Website Examples for more information.
	// +optional
	Website *BucketWebsite `json:"website,omitempty"`
}

// BucketObservation represent the observed status of a Google Cloud Storage Bucket.
type BucketObservation struct {
	// SelfLink: The URI of this bucket.
	SelfLink string `json:"selfLink,omitempty"`

	// TimeCreated: The creation time of the bucket in RFC 3339 format.
	TimeCreated string `json:"timeCreated,omitempty"`

	// Updated: The modification time of the bucket in RFC 3339 format.
	Updated string `json:"updated,omitempty"`

	// ProjectNumber: The project number of the project the bucket belongs
	// to.
	ProjectNumber int64 `json:"projectNumber,omitempty"`

	// Owner: The owner of the bucket. This is always the project team's
	// owner group.
	Owner *BucketOwner `json:"owner,omitempty"`

	// Metageneration: The metadata generation of this bucket.
	Metageneration int64 `json:"metageneration,omitempty"`
}

// BucketLogging: The bucket's logging configuration, which defines the
// destination bucket and optional name prefix for the current bucket's
// logs.
type BucketLogging struct {
	// LogBucket: The destination bucket where the current bucket's logs
	// should be placed.
	// +optional
	LogBucket *string `json:"logBucket,omitempty"`

	// LogObjectPrefix: A prefix for log object names.
	// +optional
	LogObjectPrefix *string `json:"logObjectPrefix,omitempty"`
}

// BucketLifecycle: The bucket's lifecycle configuration. See lifecycle
// management for more information.
type BucketLifecycle struct {
	// Rule: A lifecycle management rule, which is made of an action to take
	// and the condition(s) under which the action will be taken.
	Rule []BucketLifecycleRule `json:"rule"`
}

type BucketLifecycleRule struct {
	// Action: The action to take.
	Action *BucketLifecycleRuleAction `json:"action,omitempty"`

	// Condition: The condition(s) under which the action will be taken.
	Condition *BucketLifecycleRuleCondition `json:"condition,omitempty"`
}

// BucketLifecycleRuleAction: The action to take.
type BucketLifecycleRuleAction struct {
	// StorageClass: Target storage class. Required iff the type of the
	// action is SetStorageClass.
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Type: Type of the action. Currently, only Delete and SetStorageClass
	// are supported.
	Type string `json:"type"`
}

// BucketLifecycleRuleCondition: The condition(s) under which the action
// will be taken.
type BucketLifecycleRuleCondition struct {
	// Age: Age of an object (in days). This condition is satisfied when an
	// object reaches the specified age.
	// +optional
	Age *int64 `json:"age,omitempty"`

	// CreatedBefore: A date in RFC 3339 format with only the date part (for
	// instance, "2013-01-15"). This condition is satisfied when an object
	// is created before midnight of the specified date in UTC.
	// +optional
	CreatedBefore *string `json:"createdBefore,omitempty"`

	// IsLive: Relevant only for versioned objects. If the value is true,
	// this condition matches live objects; if the value is false, it
	// matches archived objects.
	// +optional
	IsLive *bool `json:"isLive,omitempty"`

	// MatchesPattern: A regular expression that satisfies the RE2 syntax.
	// This condition is satisfied when the name of the object matches the
	// RE2 pattern. Note: This feature is currently in the "Early Access"
	// launch stage and is only available to a whitelisted set of users;
	// that means that this feature may be changed in backward-incompatible
	// ways and that it is not guaranteed to be released.
	// +optional
	MatchesPattern *string `json:"matchesPattern,omitempty"`

	// MatchesStorageClass: Objects having any of the storage classes
	// specified by this condition will be matched. Values include
	// MULTI_REGIONAL, REGIONAL, NEARLINE, COLDLINE, STANDARD, and
	// DURABLE_REDUCED_AVAILABILITY.
	// +optional
	MatchesStorageClass []string `json:"matchesStorageClass,omitempty"`

	// NumNewerVersions: Relevant only for versioned objects. If the value
	// is N, this condition is satisfied when there are at least N versions
	// (including the live version) newer than this version of the object.
	// +optional
	NumNewerVersions *int64 `json:"numNewerVersions,omitempty"`
}

// BucketIamConfiguration: The bucket's IAM configuration.
type BucketIamConfiguration struct {
	// BucketPolicyOnly: The bucket's Bucket Policy Only configuration.
	// +optional
	BucketPolicyOnly *BucketIamConfigurationBucketPolicyOnly `json:"bucketPolicyOnly,omitempty"`

	// UniformBucketLevelAccess: The bucket's uniform bucket-level access
	// configuration.
	// +optional
	UniformBucketLevelAccess *BucketIamConfigurationUniformBucketLevelAccess `json:"uniformBucketLevelAccess,omitempty"`
}

// BucketIamConfigurationBucketPolicyOnly: The bucket's Bucket Policy
// Only configuration.
type BucketIamConfigurationBucketPolicyOnly struct {
	// Enabled: If set, access is controlled only by bucket-level or above
	// IAM policies.
	Enabled bool `json:"enabled"`

	// LockedTime: The deadline for changing
	// iamConfiguration.bucketPolicyOnly.enabled from true to false in RFC
	// 3339 format. iamConfiguration.bucketPolicyOnly.enabled may be changed
	// from true to false until the locked time, after which the field is
	// immutable.
	// +optional
	LockedTime *string `json:"lockedTime,omitempty"`
}

// BucketIamConfigurationUniformBucketLevelAccess: The bucket's uniform
// bucket-level access configuration.
type BucketIamConfigurationUniformBucketLevelAccess struct {
	// Enabled: If set, access is controlled only by bucket-level or above
	// IAM policies.
	Enabled bool `json:"enabled"`

	// LockedTime: The deadline for changing
	// iamConfiguration.uniformBucketLevelAccess.enabled from true to false
	// in RFC 3339  format.
	// iamConfiguration.uniformBucketLevelAccess.enabled may be changed from
	// true to false until the locked time, after which the field is
	// immutable.
	// +optional
	LockedTime *string `json:"lockedTime,omitempty"`
}

// BucketEncryption: Encryption configuration for a bucket.
type BucketEncryption struct {
	// DefaultKmsKeyName: A Cloud KMS key that will be used to encrypt
	// objects inserted into this bucket, if no encryption method is
	// specified.
	DefaultKmsKeyName string `json:"defaultKmsKeyName,omitempty"`
}

// BucketBilling: The bucket's billing configuration.
type BucketBilling struct {
	// RequesterPays: When set to true, Requester Pays is enabled for this
	// bucket.
	RequesterPays bool `json:"requesterPays,omitempty"`
}

type BucketCORS struct {
	// MaxAgeSeconds: The value, in seconds, to return in the
	// Access-Control-Max-Age header used in preflight responses.
	MaxAgeSeconds int64 `json:"maxAgeSeconds,omitempty"`

	// Method: The list of HTTP methods on which to include CORS response
	// headers, (GET, OPTIONS, POST, etc) Note: "*" is permitted in the list
	// of methods, and means "any method".
	Method []string `json:"method,omitempty"`

	// Origin: The list of Origins eligible to receive CORS response
	// headers. Note: "*" is permitted in the list of origins, and means
	// "any Origin".
	Origin []string `json:"origin,omitempty"`

	// ResponseHeader: The list of HTTP headers other than the simple
	// response headers to give permission for the user-agent to share
	// across domains.
	ResponseHeader []string `json:"responseHeader,omitempty"`
}

// BucketOwner: The owner of the bucket. This is always the project
// team's owner group.
type BucketOwner struct {
	// Entity: The entity, in the form project-owner-projectId.
	Entity string `json:"entity,omitempty"`

	// EntityId: The ID for the entity.
	EntityId string `json:"entityId,omitempty"`
}

// BucketRetentionPolicy: The bucket's retention policy. The retention
// policy enforces a minimum retention time for all objects contained in
// the bucket, based on their creation time. Any attempt to overwrite or
// delete objects younger than the retention period will result in a
// PERMISSION_DENIED error. An unlocked retention policy can be modified
// or removed from the bucket via a storage.buckets.update operation. A
// locked retention policy cannot be removed or shortened in duration
// for the lifetime of the bucket. Attempting to remove or decrease
// period of a locked retention policy will result in a
// PERMISSION_DENIED error.
type BucketRetentionPolicy struct {
	// EffectiveTime: Server-determined value that indicates the time from
	// which policy was enforced and effective. This value is in RFC 3339
	// format.
	// +optional
	EffectiveTime *string `json:"effectiveTime,omitempty"`

	// IsLocked: Once locked, an object retention policy cannot be modified.
	// +optional
	IsLocked *bool `json:"isLocked,omitempty"`

	// RetentionPeriod: The duration in seconds that objects need to be
	// retained. Retention duration must be greater than zero and less than
	// 100 years. Note that enforcement of retention periods less than a day
	// is not guaranteed. Such periods should only be used for testing
	// purposes.
	// +optional
	RetentionPeriod *int64 `json:"retentionPeriod,omitempty"`
}

// BucketVersioning: The bucket's versioning configuration.
type BucketVersioning struct {
	// Enabled: While set to true, versioning is fully enabled for this
	// bucket.
	Enabled bool `json:"enabled"`
}

// BucketWebsite: The bucket's website configuration, controlling how
// the service behaves when accessing bucket contents as a web site. See
// the Static Website Examples for more information.
type BucketWebsite struct {
	// MainPageSuffix: If the requested object path is missing, the service
	// will ensure the path has a trailing '/', append this suffix, and
	// attempt to retrieve the resulting object. This allows the creation of
	// index.html objects to represent directory pages.
	// +optional
	MainPageSuffix *string `json:"mainPageSuffix,omitempty"`

	// NotFoundPage: If the requested object path is missing, and any
	// mainPageSuffix object is missing, if applicable, the service will
	// return the named object from this bucket as the content for a 404 Not
	// Found result.
	// +optional
	NotFoundPage *string `json:"notFoundPage,omitempty"`
}

// A BucketSpec defines the desired state of a Bucket.
type BucketSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  BucketParameters `json:"forProvider"`
}

// A BucketStatus represents the observed state of a Bucket.
type BucketStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     BucketObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// A Bucket is a managed resource that represents a Google Cloud Storage bucket.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec"`
	Status BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketList contains a list of GCPBuckets
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

// BucketClassSpecTemplate is the Schema for the resource class

// A BucketClassSpecTemplate is a template for the spec of a dynamically
// provisioned Bucket.
type BucketClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	ForProvider                       BucketParameters `json:"forProvider"`
}

// +kubebuilder:object:root=true

// A BucketClass is a resource class. It defines the desired spec of resource
// claims that use it to dynamically provision a managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type BucketClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// Bucket.
	SpecTemplate BucketClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// BucketClassList contains a list of cloud memorystore resource classes.
type BucketClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketClass `json:"items"`
}
