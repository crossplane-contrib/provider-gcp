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

// CloudSQL instance states
const (
	// StateRunnable represents a CloudSQL instance in a running, available, and ready state
	StateRunnable       = "RUNNABLE"
	StateCreating       = "PENDING_CREATE"
	StateSuspended      = "SUSPENDED"
	StateMaintenance    = "MAINTENANCE"
	StateCreationFailed = "FAILED"
	StateUnknownState   = "UNKNOWN_STATE"

	CloudSQLSecretServerCACertificateCertKey             = "serverCACertificateCert"
	CloudSQLSecretServerCACertificateCertSerialNumberKey = "serverCACertificateCertSerialNumber"
	CloudSQLSecretServerCACertificateCommonNameKey       = "serverCACertificateCommonName"
	CloudSQLSecretServerCACertificateCreateTimeKey       = "serverCACertificateCreateTime"
	CloudSQLSecretServerCACertificateExpirationTimeKey   = "serverCACertificateExpirationTime"
	CloudSQLSecretServerCACertificateInstanceKey         = "serverCACertificateInstance"
	CloudSQLSecretServerCACertificateSha1FingerprintKey  = "serverCACertificateSha1Fingerprint"

	CloudSQLSecretConnectionName = "connectionName"
)

// CloudSQL version prefixes.
const (
	MysqlDBVersionPrefix = "MYSQL"
	MysqlDefaultUser     = "root"

	PostgresqlDBVersionPrefix = "POSTGRES"
	PostgresqlDefaultUser     = "postgres"

	PrivateIPType = "PRIVATE"
	PublicIPType  = "PRIMARY"

	PrivateIPKey = "privateIP"
	PublicIPKey  = "publicIP"
)

// CloudSQLInstanceParameters define the desired state of a Google CloudSQL
// instance. Most of its fields are direct mirror of GCP DatabaseInstance object.
// See https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/instances#DatabaseInstance
type CloudSQLInstanceParameters struct {
	// Region: The geographical region. Can be us-central (FIRST_GEN
	// instances only), us-central1 (SECOND_GEN instances only), asia-east1
	// or europe-west1. Defaults to us-central or us-central1 depending on
	// the instance type (First Generation or Second Generation). The region
	// can not be changed after instance creation.
	// +immutable
	Region string `json:"region"`

	// Settings: The user settings.
	Settings Settings `json:"settings"`

	// DatabaseVersion: The database engine type and version. The
	// databaseVersion field can not be changed after instance creation.
	// MySQL Second Generation instances: MYSQL_5_7 (default) or MYSQL_5_6.
	// PostgreSQL instances: POSTGRES_9_6 (default) or POSTGRES_11 Beta.
	// MySQL First Generation instances: MYSQL_5_6 (default) or MYSQL_5_5
	// +immutable
	// +optional
	DatabaseVersion *string `json:"databaseVersion,omitempty"`

	// MasterInstanceName: The name of the instance which will act as master
	// in the replication setup.
	// +optional
	// +immutable
	MasterInstanceName *string `json:"masterInstanceName,omitempty"`

	// DiskEncryptionConfiguration: Disk encryption configuration specific
	// to an instance. Applies only to Second Generation instances.
	// +optional
	// +immutable
	DiskEncryptionConfiguration *DiskEncryptionConfiguration `json:"diskEncryptionConfiguration,omitempty"`

	// FailoverReplica: The name and status of the failover replica. This
	// property is applicable only to Second Generation instances.
	// +optional
	FailoverReplica *DatabaseInstanceFailoverReplicaSpec `json:"failoverReplica,omitempty"`

	// GceZone: The Compute Engine zone that the instance is currently
	// serving from. This value could be different from the zone that was
	// specified when the instance was created if the instance has failed
	// over to its secondary zone.
	// +optional
	GceZone *string `json:"gceZone,omitempty"`

	// InstanceType: The instance type. This can be one of the
	// following.
	// CLOUD_SQL_INSTANCE: A Cloud SQL instance that is not replicating from
	// a master.
	// ON_PREMISES_INSTANCE: An instance running on the customer's
	// premises.
	// READ_REPLICA_INSTANCE: A Cloud SQL instance configured as a
	// read-replica.
	// +optional
	// +immutable
	InstanceType *string `json:"instanceType,omitempty"`

	// MaxDiskSize: The maximum disk size of the instance in bytes.
	// +optional
	MaxDiskSize *int64 `json:"maxDiskSize,omitempty"`

	// OnPremisesConfiguration: Configuration specific to on-premises
	// instances.
	// +optional
	OnPremisesConfiguration *OnPremisesConfiguration `json:"onPremisesConfiguration,omitempty"`

	// ReplicaNames: The replicas of the instance.
	// +optional
	ReplicaNames []string `json:"replicaNames,omitempty"`

	// SuspensionReason: If the instance state is SUSPENDED, the reason for
	// the suspension.
	// +optional
	SuspensionReason []string `json:"suspensionReason,omitempty"`
}

// Settings is Cloud SQL database instance settings.
type Settings struct {
	// Tier: The tier (or machine type) for this instance, for example
	// db-n1-standard-1 (MySQL instances) or db-custom-1-3840 (PostgreSQL
	// instances). For MySQL instances, this property determines whether the
	// instance is First or Second Generation. For more information, see
	// Instance Settings.
	Tier string `json:"tier"`

	// ActivationPolicy: The activation policy specifies when the instance
	// is activated; it is applicable only when the instance state is
	// RUNNABLE. Valid values:
	// ALWAYS: The instance is on, and remains so even in the absence of
	// connection requests.
	// NEVER: The instance is off; it is not activated, even if a connection
	// request arrives.
	// ON_DEMAND: First Generation instances only. The instance responds to
	// incoming requests, and turns itself off when not in use. Instances
	// with PER_USE pricing turn off after 15 minutes of inactivity.
	// Instances with PER_PACKAGE pricing turn off after 12 hours of
	// inactivity.
	// +optional
	ActivationPolicy *string `json:"activationPolicy,omitempty"`

	// AuthorizedGaeApplications: The App Engine app IDs that can access
	// this instance. First Generation instances only.
	// +optional
	AuthorizedGaeApplications []string `json:"authorizedGaeApplications,omitempty"`

	// AvailabilityType: Availability type (PostgreSQL instances only).
	// Potential values:
	// ZONAL: The instance serves data from only one zone. Outages in that
	// zone affect data accessibility.
	// REGIONAL: The instance can serve data from more than one zone in a
	// region (it is highly available).
	// For more information, see Overview of the High Availability
	// Configuration.
	// +optional
	AvailabilityType *string `json:"availabilityType,omitempty"`

	// CrashSafeReplicationEnabled: Configuration specific to read replica
	// instances. Indicates whether database flags for crash-safe
	// replication are enabled. This property is only applicable to First
	// Generation instances.
	// +optional
	CrashSafeReplicationEnabled *bool `json:"crashSafeReplicationEnabled,omitempty"`

	// StorageAutoResize: Configuration to increase storage size
	// automatically. The default value is true. Not used for First
	// Generation instances.
	// +optional
	StorageAutoResize *bool `json:"storageAutoResize,omitempty"`

	// DataDiskType: The type of data disk: PD_SSD (default) or PD_HDD. Not
	// used for First Generation instances.
	// +optional
	DataDiskType *string `json:"dataDiskType,omitempty"`

	// PricingPlan: The pricing plan for this instance. This can be either
	// PER_USE or PACKAGE. Only PER_USE is supported for Second Generation
	// instances.
	// +optional
	PricingPlan *string `json:"pricingPlan,omitempty"`

	// ReplicationType: The type of replication this instance uses. This can
	// be either ASYNCHRONOUS or SYNCHRONOUS. This property is only
	// applicable to First Generation instances.
	// +optional
	ReplicationType *string `json:"replicationType,omitempty"`

	// UserLabels: User-provided labels, represented as a dictionary where
	// each label is a single key value pair.
	// +optional
	UserLabels map[string]string `json:"userLabels,omitempty"`

	// DatabaseFlags is the array of database flags passed to the instance at
	// startup.
	// +optional
	DatabaseFlags []*DatabaseFlags `json:"databaseFlags,omitempty"`

	// BackupConfiguration is the daily backup configuration for the instance.
	// +optional
	BackupConfiguration *BackupConfiguration `json:"backupConfiguration,omitempty"`

	// IPConfiguration: The settings for IP Management. This allows to
	// enable or disable the instance IP and manage which external networks
	// can connect to the instance. The IPv4 address cannot be disabled for
	// Second Generation instances.
	// +optional
	IPConfiguration *IPConfiguration `json:"ipConfiguration,omitempty"`

	// LocationPreference is the location preference settings. This allows the
	// instance to be located as near as possible to either an App Engine
	// app or Compute Engine zone for better performance. App Engine
	// co-location is only applicable to First Generation instances.
	// +optional
	LocationPreference *LocationPreference `json:"locationPreference,omitempty"`

	// MaintenanceWindow: The maintenance window for this instance. This
	// specifies when the instance can be restarted for maintenance
	// purposes. Not used for First Generation instances.
	// +optional
	MaintenanceWindow *MaintenanceWindow `json:"maintenanceWindow,omitempty"`

	// DataDiskSizeGb: The size of data disk, in GB. The data disk size
	// minimum is 10GB. Not used for First Generation instances.
	// +optional
	DataDiskSizeGb *int64 `json:"dataDiskSizeGb,omitempty"`

	// DatabaseReplicationEnabled: Configuration specific to read replica
	// instances. Indicates whether replication is enabled or not.
	// +optional
	DatabaseReplicationEnabled *bool `json:"databaseReplicationEnabled,omitempty"`

	// StorageAutoResizeLimit: The maximum size to which storage capacity
	// can be automatically increased. The default value is 0, which
	// specifies that there is no limit. Not used for First Generation
	// instances.
	// +optional
	StorageAutoResizeLimit *int64 `json:"storageAutoResizeLimit,omitempty"`
}

// LocationPreference is preferred location. This specifies where a Cloud
// SQL instance should preferably be located, either in a specific
// Compute Engine zone, or co-located with an App Engine application.
// Note that if the preferred location is not available, the instance
// will be located as close as possible within the region. Only one
// location may be specified.
type LocationPreference struct {
	// FollowGaeApplication: The AppEngine application to follow, it must be
	// in the same region as the Cloud SQL instance.
	// +optional
	FollowGaeApplication *string `json:"followGaeApplication,omitempty"`

	// Zone: The preferred Compute Engine zone (e.g. us-central1-a,
	// us-central1-b, etc.).
	// +optional
	Zone *string `json:"zone,omitempty"`
}

// MaintenanceWindow specifies when a v2 Cloud SQL instance should preferably
// be restarted for system maintenance purposes.
type MaintenanceWindow struct {
	// Day: day of week (1-7), starting on Monday.
	// +optional
	Day *int64 `json:"day,omitempty"`

	// Hour: hour of day - 0 to 23.
	// +optional
	Hour *int64 `json:"hour,omitempty"`

	// UpdateTrack: Maintenance timing setting: canary (Earlier) or stable
	// (Later).
	// +optional
	UpdateTrack *string `json:"updateTrack,omitempty"`
}

// BackupConfiguration is database instance backup configuration.
type BackupConfiguration struct {
	// BinaryLogEnabled: Whether binary log is enabled. If backup
	// configuration is disabled, binary log must be disabled as well.
	// +optional
	BinaryLogEnabled *bool `json:"binaryLogEnabled,omitempty"`

	// Enabled: Whether this configuration is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Location: The location of the backup.
	// +optional
	Location *string `json:"location,omitempty"`

	// ReplicationLogArchivingEnabled: Reserved for future use.
	// +optional
	ReplicationLogArchivingEnabled *bool `json:"replicationLogArchivingEnabled,omitempty"`

	// StartTime: Start time for the daily backup configuration in UTC
	// timezone in the 24 hour format - HH:MM.
	// +optional
	StartTime *string `json:"startTime,omitempty"`
}

// DatabaseFlags are database flags for Cloud SQL instances.
type DatabaseFlags struct {
	// Name: The name of the flag. These flags are passed at instance
	// startup, so include both server options and system variables for
	// MySQL. Flags should be specified with underscores, not hyphens. For
	// more information, see Configuring Database Flags in the Cloud SQL
	// documentation.
	Name string `json:"name"`

	// Value: The value of the flag. Booleans should be set to on for true
	// and off for false. This field must be omitted if the flag doesn't
	// take a value.
	Value string `json:"value"`
}

// IPConfiguration is the IP Management configuration.
type IPConfiguration struct {
	// AuthorizedNetworks: The list of external networks that are allowed to
	// connect to the instance using the IP. In CIDR notation, also known as
	// 'slash' notation (e.g. 192.168.100.0/24).
	// +optional
	AuthorizedNetworks []*ACLEntry `json:"authorizedNetworks,omitempty"`

	// Ipv4Enabled: Whether the instance should be assigned an IP address or
	// not.
	// +optional
	Ipv4Enabled *bool `json:"ipv4Enabled,omitempty"`

	// PrivateNetwork: The resource link for the VPC network from which the
	// Cloud SQL instance is accessible for private IP. For example,
	// /projects/myProject/global/networks/default. This setting can be updated,
	// but it cannot be removed after it is set. The Network must have an active
	// Service Networking connection peering before resolution will proceed.
	// https://cloud.google.com/vpc/docs/configure-private-services-access
	// +optional
	PrivateNetwork *string `json:"privateNetwork,omitempty"`

	// PrivateNetworkRef sets the PrivateNetwork field by resolving the resource
	// link of the referenced Crossplane Network managed resource.
	// +optional
	PrivateNetworkRef *xpv1.Reference `json:"privateNetworkRef,omitempty"`

	// PrivateNetworkSelector selects a PrivateNetworkRef.
	// +optional
	PrivateNetworkSelector *xpv1.Selector `json:"privateNetworkSelector,omitempty"`

	// RequireSsl: Whether SSL connections over IP should be enforced or
	// not.
	// +optional
	RequireSsl *bool `json:"requireSsl,omitempty"`
}

// ACLEntry is an entry for an Access Control list.
type ACLEntry struct {
	// ExpirationTime: The time when this access control entry expires in
	// RFC 3339 format, for example 2012-11-15T16:19:00.094Z.
	// +optional
	ExpirationTime *string `json:"expirationTime,omitempty"`

	// Name: An optional label to identify this entry.
	// +optional
	Name *string `json:"name,omitempty"`

	// Value: The whitelisted value for the access control list.
	// +optional
	Value *string `json:"value,omitempty"`
}

// OnPremisesConfiguration is on-premises instance configuration.
type OnPremisesConfiguration struct {
	// HostPort: The host and port of the on-premises instance in host:port
	// format
	HostPort string `json:"hostPort"`
}

// CloudSQLInstanceObservation is used to show the observed state of the Cloud SQL resource on GCP.
type CloudSQLInstanceObservation struct {
	// BackendType: FIRST_GEN: First Generation instance. MySQL
	// only.
	// SECOND_GEN: Second Generation instance or PostgreSQL
	// instance.
	// EXTERNAL: A database server that is not managed by Google.
	// This property is read-only; use the tier property in the settings
	// object to determine the database type and Second or First Generation.
	BackendType string `json:"backendType,omitempty"`

	// CurrentDiskSize: The current disk usage of the instance in bytes.
	// This property has been deprecated. Users should use the
	// "cloudsql.googleapis.com/database/disk/bytes_used" metric in Cloud
	// Monitoring API instead. Please see this announcement for details.
	CurrentDiskSize int64 `json:"currentDiskSize,omitempty"`

	// ConnectionName: Connection name of the Cloud SQL instance used in
	// connection strings.
	ConnectionName string `json:"connectionName,omitempty"`

	// DiskEncryptionStatus: Disk encryption status specific to an instance.
	// Applies only to Second Generation instances.
	DiskEncryptionStatus *DiskEncryptionStatus `json:"diskEncryptionStatus,omitempty"`

	// FailoverReplica: The name and status of the failover replica. This
	// property is applicable only to Second Generation instances.
	FailoverReplica *DatabaseInstanceFailoverReplicaStatus `json:"failoverReplica,omitempty"`

	// GceZone: The Compute Engine zone that the instance is currently
	// serving from. This value could be different from the zone that was
	// specified when the instance was created if the instance has failed
	// over to its secondary zone.
	GceZone string `json:"gceZone,omitempty"`

	// IPAddresses: The assigned IP addresses for the instance.
	IPAddresses []*IPMapping `json:"ipAddresses,omitempty"`

	// IPv6Address: The IPv6 address assigned to the instance. This property
	// is applicable only to First Generation instances.
	IPv6Address string `json:"ipv6Address,omitempty"`

	// Project: The project ID of the project containing the Cloud SQL
	// instance. The Google apps domain is prefixed if applicable.
	Project string `json:"project,omitempty"`

	// SelfLink: The URI of this resource.
	SelfLink string `json:"selfLink,omitempty"`

	// ServiceAccountEmailAddress: The service account email address
	// assigned to the instance. This property is applicable only to Second
	// Generation instances.
	ServiceAccountEmailAddress string `json:"serviceAccountEmailAddress,omitempty"`

	// State: The current serving state of the Cloud SQL instance. This can
	// be one of the following.
	// RUNNABLE: The instance is running, or is ready to run when
	// accessed.
	// SUSPENDED: The instance is not available, for example due to problems
	// with billing.
	// PENDING_CREATE: The instance is being created.
	// MAINTENANCE: The instance is down for maintenance.
	// FAILED: The instance creation failed.
	// UNKNOWN_STATE: The state of the instance is unknown.
	State string `json:"state,omitempty"`

	// NOTE(muvaf): This comes from Settings sub-struct, not directly from
	// DatabaseInstance struct.

	// SettingsVersion: The version of instance settings. This is a required
	// field for update method to make sure concurrent updates are handled
	// properly. During update, use the most recent settingsVersion value
	// for this instance and do not try to update this value.
	SettingsVersion int64 `json:"settingsVersion,omitempty"`
}

// IPMapping is database instance IP Mapping.
type IPMapping struct {
	// IPAddress: The IP address assigned.
	IPAddress string `json:"ipAddress,omitempty"`

	// TimeToRetire: The due time for this IP to be retired in RFC 3339
	// format, for example 2012-11-15T16:19:00.094Z. This field is only
	// available when the IP is scheduled to be retired.
	TimeToRetire string `json:"timeToRetire,omitempty"`

	// Type: The type of this IP address. A PRIMARY address is a public
	// address that can accept incoming connections. A PRIVATE address is a
	// private address that can accept incoming connections. An OUTGOING
	// address is the source address of connections originating from the
	// instance, if supported.
	Type string `json:"type,omitempty"`
}

// DiskEncryptionConfiguration is disk encryption configuration.
type DiskEncryptionConfiguration struct {
	// KmsKeyName: KMS key resource name
	KmsKeyName string `json:"kmsKeyName"`
}

// DiskEncryptionStatus is disk encryption status.
type DiskEncryptionStatus struct {
	// KmsKeyVersionName: KMS key version used to encrypt the Cloud SQL
	// instance disk
	KmsKeyVersionName string `json:"kmsKeyVersionName"`
}

// DatabaseInstanceFailoverReplicaSpec is where you can specify a name
// for the failover replica.
type DatabaseInstanceFailoverReplicaSpec struct {
	// Name: The name of the failover replica. If specified at instance
	// creation, a failover replica is created for the instance. The name
	// doesn't include the project ID. This property is applicable only to
	// Second Generation instances.
	Name string `json:"name"`
}

// DatabaseInstanceFailoverReplicaStatus is status of the failover
// replica.
type DatabaseInstanceFailoverReplicaStatus struct {
	// Available: The availability status of the failover replica. A false
	// status indicates that the failover replica is out of sync. The master
	// can only failover to the failover replica when the status is true.
	Available bool `json:"available"`
}

// A CloudSQLInstanceSpec defines the desired state of a CloudSQLInstance.
type CloudSQLInstanceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CloudSQLInstanceParameters `json:"forProvider"`
}

// A CloudSQLInstanceStatus represents the observed state of a CloudSQLInstance.
type CloudSQLInstanceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CloudSQLInstanceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CloudSQLInstance is a managed resource that represents a Google CloudSQL
// instance.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.databaseVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type CloudSQLInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudSQLInstanceSpec   `json:"spec"`
	Status CloudSQLInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudSQLInstanceList contains a list of CloudSQLInstance
type CloudSQLInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudSQLInstance `json:"items"`
}
