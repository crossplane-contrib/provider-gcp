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

package v1alpha2

import (
	"google.golang.org/genproto/googleapis/cloud/redis/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
)

// Cloud Memorystore instance states.
var (
	StateUnspecified = redis.Instance_STATE_UNSPECIFIED.String()
	StateCreating    = redis.Instance_CREATING.String()
	StateReady       = redis.Instance_READY.String()
	StateUpdating    = redis.Instance_UPDATING.String()
	StateDeleting    = redis.Instance_DELETING.String()
	StateRepairing   = redis.Instance_REPAIRING.String()
	StateMaintenance = redis.Instance_MAINTENANCE.String()
)

// Cloud Memorystore instance tiers.
var (
	TierBasic      = redis.Instance_BASIC.String()
	TierStandardHA = redis.Instance_STANDARD_HA.String()
)

// CloudMemorystoreInstanceParameters define the desired state of an Google
// Cloud Memorystore instance. Most fields map directly to an Instance:
// https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances#Instance
type CloudMemorystoreInstanceParameters struct {
	// Region in which to create this Cloud Memorystore cluster.
	Region string `json:"region"`

	// Tier specifies the replication level of the Redis cluster. BASIC provides
	// a single Redis instance with no high availability. STANDARD_HA provides a
	// cluster of two Redis instances in distinct availability zones.
	// https://cloud.google.com/memorystore/docs/redis/redis-tiers
	// +kubebuilder:validation:Enum=BASIC;STANDARD_HA
	Tier string `json:"tier"`

	// LocationID specifies the zone where the instance will be provisioned. If
	// not provided, the service will choose a zone for the instance. For
	// STANDARD_HA tier, instances will be created across two zones for
	// protection against zonal failures.
	// +optional
	LocationID string `json:"locationId,omitempty"`

	// AlternativeLocationID is only applicable to STANDARD_HA tier, which
	// protects the instance against zonal failures by provisioning it across
	// two zones. If provided, it must be a different zone from the one provided
	// in locationId.
	// +optional
	AlternativeLocationID string `json:"alternativeLocationId,omitempty"`

	// MemorySizeGB specifies the Redis memory size in GiB.
	MemorySizeGB int `json:"memorySizeGb"`

	// ReservedIPRange specifies the CIDR range of internal addresses that are
	// reserved for this instance. If not provided, the service will choose an
	// unused /29 block, for example, 10.0.0.0/29 or 192.168.0.0/29. Ranges must
	// be unique and non-overlapping with existing subnets in an authorized
	// network.
	// +optional
	ReservedIPRange string `json:"reservedIpRange,omitempty"`

	// AuthorizedNetwork specifies the full name of the Google Compute Engine
	// network to which the instance is connected. If left unspecified, the
	// default network will be used.
	// +optional
	AuthorizedNetwork string `json:"authorizedNetwork,omitempty"`

	// RedisVersion specifies the version of Redis software. If not provided,
	// latest supported version will be used. Updating the version will perform
	// an upgrade/downgrade to the new version. Currently, the supported values
	// are REDIS_3_2 for Redis 3.2, and REDIS_4_0 for Redis 4.0 (the default).
	// +kubebuilder:validation:Enum=REDIS_3_2;REDIS_4_0
	// +optional
	RedisVersion string `json:"redisVersion,omitempty"`

	// RedisConfigs specifies Redis configuration parameters, according to
	// http://redis.io/topics/config. Currently, the only supported parameters
	// are:
	// * maxmemory-policy
	// * notify-keyspace-events
	// +optional
	RedisConfigs map[string]string `json:"redisConfigs,omitempty"`
}

// A CloudMemorystoreInstanceSpec defines the desired state of a
// CloudMemorystoreInstance.
type CloudMemorystoreInstanceSpec struct {
	runtimev1alpha1.ResourceSpec       `json:",inline"`
	CloudMemorystoreInstanceParameters `json:",inline"`
}

// A CloudMemorystoreInstanceStatus represents the observed state of a
// CloudMemorystoreInstance.
type CloudMemorystoreInstanceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this instance.
	State string `json:"state,omitempty"`

	// Additional information about the current status of this instance, if
	// available.
	Message string `json:"message,omitempty"`

	// ProviderID is the external ID to identify this resource in the cloud
	// provider, e.g. 'projects/fooproj/locations/us-foo1/instances/foo'
	ProviderID string `json:"providerID,omitempty"`

	// CurrentLocationID is the current zone where the Redis endpoint is placed.
	// For Basic Tier instances, this will always be the same as the locationId
	// provided by the user at creation time. For Standard Tier instances, this
	// can be either locationId or alternativeLocationId and can change after a
	// failover event.
	CurrentLocationID string `json:"currentLocationId,omitempty"`

	// Endpoint of the Cloud Memorystore instance used in connection strings.
	Endpoint string `json:"endpoint,omitempty"`

	// Port at which the Cloud Memorystore instance endpoint is listening.
	Port int `json:"port,omitempty"`
}

// +kubebuilder:object:root=true

// A CloudMemorystoreInstance is a managed resource that represents a Google
// Cloud Memorystore instance.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.redisVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type CloudMemorystoreInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudMemorystoreInstanceSpec   `json:"spec,omitempty"`
	Status CloudMemorystoreInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudMemorystoreInstanceList contains a list of CloudMemorystoreInstance
type CloudMemorystoreInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudMemorystoreInstance `json:"items"`
}

// A CloudMemorystoreInstanceClassSpecTemplate is a template for the spec of a
// dynamically provisioned CloudMemorystoreInstance.
type CloudMemorystoreInstanceClassSpecTemplate struct {
	runtimev1alpha1.NonPortableClassSpecTemplate `json:",inline"`
	CloudMemorystoreInstanceParameters           `json:",inline"`
}

// +kubebuilder:object:root=true

// A CloudMemorystoreInstanceClass is a non-portable resource class. It defines
// the desired spec of resource claims that use it to dynamically provision a
// managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type CloudMemorystoreInstanceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// CloudMemorystoreInstance.
	SpecTemplate CloudMemorystoreInstanceClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// CloudMemorystoreInstanceClassList contains a list of cloud memorystore resource classes.
type CloudMemorystoreInstanceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudMemorystoreInstanceClass `json:"items"`
}
