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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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
	MemorySizeGB int64 `json:"memorySizeGb"`

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
	//  *   `REDIS_5_0` for Redis 5.0 compatibility
	//  *   `REDIS_6_X` for Redis 6.x compatibility
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

	// ConnectMode: Optional. The network connect mode of the Redis
	// instance. If not provided, the connect mode defaults to
	// DIRECT_PEERING.
	//
	// Possible values:
	//   "CONNECT_MODE_UNSPECIFIED" - Not set.
	//   "DIRECT_PEERING" - Connect via direct peering to the Memorystore
	// for Redis hosted service.
	//   "PRIVATE_SERVICE_ACCESS" - Connect your Memorystore for Redis
	// instance using Private Services Access. Private services access
	// provides an IP address range for multiple Google Cloud services,
	// including Memorystore.
	// +kubebuilder:validation:Enum=DIRECT_PEERING;PRIVATE_SERVICE_ACCESS
	// +optional
	// +immutable
	ConnectMode *string `json:"connectMode,omitempty"`

	// AuthEnabled: Optional. Indicates whether OSS Redis AUTH is enabled
	// for the instance. If set to "true" AUTH is enabled on the instance.
	// Default value is "false" meaning AUTH is disabled.
	// +optional
	AuthEnabled *bool `json:"authEnabled,omitempty"`
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
	Port int64 `json:"port,omitempty"`

	// The current zone where the Redis endpoint is placed. For Basic
	// Tier instances, this will always be the same as the [location_id]
	// provided by the user at creation time. For Standard Tier instances,
	// this can be either [location_id] or [alternative_location_id] and can
	// change after a failover event.
	CurrentLocationID string `json:"currentLocationId,omitempty"`

	// The time the instance was created.
	CreateTime *metav1.Time `json:"createTime,omitempty"`

	// State: Output only. The current state of this instance.
	//
	// Possible values:
	//   "STATE_UNSPECIFIED" - Not set.
	//   "CREATING" - Redis instance is being created.
	//   "READY" - Redis instance has been created and is fully usable.
	//   "UPDATING" - Redis instance configuration is being updated. Certain
	// kinds of updates may cause the instance to become unusable while the
	// update is in progress.
	//   "DELETING" - Redis instance is being deleted.
	//   "REPAIRING" - Redis instance is being repaired and may be unusable.
	//   "MAINTENANCE" - Maintenance is being performed on this Redis
	// instance.
	//   "IMPORTING" - Redis instance is importing data (availability may be
	// affected).
	//   "FAILING_OVER" - Redis instance is failing over (availability may
	// be affected).
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
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CloudMemorystoreInstanceParameters `json:"forProvider"`
}

// A CloudMemorystoreInstanceStatus represents the observed state of a
// CloudMemorystoreInstance.
type CloudMemorystoreInstanceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CloudMemorystoreInstanceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CloudMemorystoreInstance is a managed resource that represents a Google
// Cloud Memorystore instance.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.redisVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
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
