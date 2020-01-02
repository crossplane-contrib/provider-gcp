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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
)

// CloudMemorystoreInstanceParameters define the desired state of an Google
// Cloud Memorystore instance. Most fields map directly to an Instance:
// https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances#Instance
type CloudMemorystoreInstanceParameters struct {
	// Region in which to create this Cloud Memorystore cluster.
	// +immutable
	Region string `json:"region"`

	// Tier specifies the replication level of the Redis cluster. BASIC provides
	// a single Redis instance with no high availability. STANDARD_HA provides a
	// cluster of two Redis instances in distinct availability zones.
	// https://cloud.google.com/memorystore/docs/redis/redis-tiers
	// +kubebuilder:validation:Enum=BASIC;STANDARD_HA
	// +immutable
	Tier string `json:"tier"`

	// Redis memory size in GiB.
	MemorySizeGB int32 `json:"memorySizeGb"`

	// An arbitrary and optional user-provided name for the instance.
	// +optional
	DisplayName *string `json:"displayName,omitempty"`

	// Resource labels to represent user provided metadata
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// The zone where the instance will be provisioned. If not provided,
	// the service will choose a zone for the instance. For STANDARD_HA tier,
	// instances will be created across two zones for protection against zonal
	// failures. If [alternative_location_id] is also provided, it must be
	// different from [location_id].
	// +optional
	// +immutable
	LocationID *string `json:"locationId,omitempty"`

	// Only applicable to STANDARD_HA tier which protects the instance
	// against zonal failures by provisioning it across two zones. If provided, it
	// must be a different zone from the one provided in [location_id].
	// +optional
	// +immutable
	AlternativeLocationID *string `json:"alternativeLocationId,omitempty"`

	// The version of Redis software.
	// If not provided, latest supported version will be used. Updating the
	// version will perform an upgrade/downgrade to the new version. Currently,
	// the supported values are:
	//
	//  *   `REDIS_4_0` for Redis 4.0 compatibility (default)
	//  *   `REDIS_3_2` for Redis 3.2 compatibility
	// +optional
	// +immutable
	RedisVersion *string `json:"redisVersion,omitempty"`

	// The CIDR range of internal addresses that are reserved for this
	// instance. If not provided, the service will choose an unused /29 block,
	// for example, 10.0.0.0/29 or 192.168.0.0/29. Ranges must be unique
	// and non-overlapping with existing subnets in an authorized network.
	// +optional
	// +immutable
	ReservedIPRange *string `json:"reservedIpRange,omitempty"`

	// Redis configuration parameters, according to
	// http://redis.io/topics/config. Currently, the only supported parameters
	// are:
	//
	//  Redis 3.2 and above:
	//
	//  *   maxmemory-policy
	//  *   notify-keyspace-events
	//
	//  Redis 4.0 and above:
	//
	//  *   activedefrag
	//  *   lfu-log-factor
	//  *   lfu-decay-time
	// +optional
	RedisConfigs map[string]string `json:"redisConfigs,omitempty"`

	// The full name of the Google Compute Engine
	// [network](/compute/docs/networks-and-firewalls#networks) to which the
	// instance is connected. If left unspecified, the `default` network
	// will be used.
	// +optional
	// +immutable
	AuthorizedNetwork *string `json:"authorizedNetwork,omitempty"`
}

// CloudMemorystoreInstanceObservation is used to show the observed state of the
// CloudMemorystore resource on GCP.
type CloudMemorystoreInstanceObservation struct {
	// Unique name of the resource in this scope including project and
	// location using the form:
	//     `projects/{project_id}/locations/{location_id}/instances/{instance_id}`
	//
	// Note: Redis instances are managed and addressed at regional level so
	// location_id here refers to a GCP region; however, users may choose which
	// specific zone (or collection of zones for cross-zone instances) an instance
	// should be provisioned in. Refer to [location_id] and
	// [alternative_location_id] fields for more details.
	Name string `json:"name,omitempty"`

	// Hostname or IP address of the exposed Redis endpoint used by
	// clients to connect to the service.
	Host string `json:"host,omitempty"`

	// The port number of the exposed Redis endpoint.
	Port int32 `json:"port,omitempty"`

	// The current zone where the Redis endpoint is placed. For Basic
	// Tier instances, this will always be the same as the [location_id]
	// provided by the user at creation time. For Standard Tier instances,
	// this can be either [location_id] or [alternative_location_id] and can
	// change after a failover event.
	CurrentLocationID string `json:"currentLocationId,omitempty"`

	// The time the instance was created.
	CreateTime *metav1.Time `json:"createTime,omitempty"`

	// The current state of this instance.
	State string `json:"state,omitempty"`

	// Additional information about the current status of this
	// instance, if available.
	StatusMessage string `json:"statusMessage,omitempty"`

	// Cloud IAM identity used by import / export operations to
	// transfer data to/from Cloud Storage. Format is
	// "serviceAccount:<service_account_email>". The value may change over time
	// for a given instance so should be checked before each import/export
	// operation.
	PersistenceIAMIdentity string `json:"persistenceIamIdentity,omitempty"`
}

// A CloudMemorystoreInstanceSpec defines the desired state of a
// CloudMemorystoreInstance.
type CloudMemorystoreInstanceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CloudMemorystoreInstanceParameters `json:"forProvider"`
}

// A CloudMemorystoreInstanceStatus represents the observed state of a
// CloudMemorystoreInstance.
type CloudMemorystoreInstanceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     CloudMemorystoreInstanceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CloudMemorystoreInstance is a managed resource that represents a Google
// Cloud Memorystore instance.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.redisVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type CloudMemorystoreInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudMemorystoreInstanceSpec   `json:"spec"`
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
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	ForProvider                       CloudMemorystoreInstanceParameters `json:"forProvider"`
}

// +kubebuilder:object:root=true

// A CloudMemorystoreInstanceClass is a resource class. It defines the desired
// spec of resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
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
