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
	"strings"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"
)

// CloudSQL instance states
const (
	// StateRunnable represents a CloudSQL instance in a running, available, and ready state
	StateRunnable = "RUNNABLE"
)

// CloudSQL version prefixes.
const (
	MysqlDBVersionPrefix = "MYSQL"
	MysqlDefaultUser     = "root"

	PostgresqlDBVersionPrefix = "POSTGRES"
	PostgresqlDefaultUser     = "postgres"

	PasswordLength   = 20
	DefaultStorageGB = 10

	PrivateIPType = "PRIVATE"
	PublicIPType  = "PRIMARY"

	PrivateIPKey = "privateIP"
	PublicIPKey  = "publicIP"
)

// CloudsqlInstanceParameters define the desired state of a Google CloudSQL
// instance.
type CloudsqlInstanceParameters struct {
	// AuthorizedNetworks is the list of external networks that are allowed to
	// connect to the instance using the IP. In CIDR notation, also known as
	// 'slash' notation (e.g. 192.168.100.0/24).
	// +optional
	AuthorizedNetworks []string `json:"authorizedNetworks,omitempty"`

	// PrivateNetwork is the resource link for the VPC network from which the
	// Cloud SQL instance is accessible for private IP. For example,
	// /projects/myProject/global/networks/default. This setting can be
	// updated, but it cannot be removed after it is set.
	// +optional
	PrivateNetwork string `json:"privateNetwork,omitempty"`

	// Ipv4Enabled specifies whether the instance should be assigned an IP
	// address or not.
	// +optional
	Ipv4Enabled bool `json:"ipv4Enabled,omitempty"`

	// The database engine (MySQL or PostgreSQL) and its specific version to use, e.g., MYSQL_5_7 or POSTGRES_9_6.

	// DatabaseVersion specifies he database engine type and version. MySQL
	// Second Generation instances use MYSQL_5_7 (default) or MYSQL_5_6.
	// MySQL First Generation instances use MYSQL_5_6 (default) or MYSQL_5_5
	// PostgreSQL instances uses POSTGRES_9_6 (default) or POSTGRES_11.
	DatabaseVersion string `json:"databaseVersion"`

	// Labels to apply to this CloudSQL instance.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Region specifies the geographical region of this CloudSQL instance.
	Region string `json:"region"`

	// StorageType specifies the type of the data disk, either PD_SSD or PD_HDD.
	StorageType string `json:"storageType"`

	// StorageGB specifies the size of the data disk. The minimum is 10GB.
	StorageGB int64 `json:"storageGB"`

	// Tier (or machine type) for this instance, for example db-n1-standard-1
	// (MySQL instances) or db-custom-1-3840 (PostgreSQL instances). For MySQL
	// instances, this property determines whether the instance is First or
	// Second Generation. For more information, see
	// https://cloud.google.com/sql/docs/mysql/instance-settings
	Tier string `json:"tier"`

	// NameFormat specifies the name of the extenral CloudSQL instance. The
	// first instance of the string '%s' will be replaced with the Kubernetes
	// UID of this CloudsqlInstance.
	NameFormat string `json:"nameFormat,omitempty"`
}

// A CloudsqlInstanceSpec defines the desired state of a CloudsqlInstance.
type CloudsqlInstanceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	CloudsqlInstanceParameters   `json:",inline"`
}

// A CloudsqlInstanceStatus represents the observed state of a CloudsqlInstance.
type CloudsqlInstanceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this CloudsqlInstance.
	State string `json:"state,omitempty"`

	// TODO(muvaf): Convert these types to *string during managed reconciler
	// refactor because both are optional.
	// https://github.com/crossplaneio/crossplane/issues/741

	// PublicIP is used to connect to this resource from other authorized
	// networks.
	PublicIP string `json:"publicIp,omitempty"`

	// PrivateIP is used to connect to this instance from the same Network.
	PrivateIP string `json:"privateIp,omitempty"`
}

// +kubebuilder:object:root=true

// A CloudsqlInstance is a managed resource that represents a Google CloudSQL
// instance.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.databaseVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type CloudsqlInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudsqlInstanceSpec   `json:"spec,omitempty"`
	Status CloudsqlInstanceStatus `json:"status,omitempty"`
}

// SetBindingPhase of this CloudsqlInstance.
func (i *CloudsqlInstance) SetBindingPhase(p runtimev1alpha1.BindingPhase) {
	i.Status.SetBindingPhase(p)
}

// GetBindingPhase of this CloudsqlInstance.
func (i *CloudsqlInstance) GetBindingPhase() runtimev1alpha1.BindingPhase {
	return i.Status.GetBindingPhase()
}

// SetConditions of this CloudsqlInstance.
func (i *CloudsqlInstance) SetConditions(c ...runtimev1alpha1.Condition) {
	i.Status.SetConditions(c...)
}

// SetClaimReference of this CloudsqlInstance.
func (i *CloudsqlInstance) SetClaimReference(r *corev1.ObjectReference) {
	i.Spec.ClaimReference = r
}

// GetClaimReference of this CloudsqlInstance.
func (i *CloudsqlInstance) GetClaimReference() *corev1.ObjectReference {
	return i.Spec.ClaimReference
}

// SetNonPortableClassReference of this CloudsqlInstance.
func (i *CloudsqlInstance) SetNonPortableClassReference(r *corev1.ObjectReference) {
	i.Spec.NonPortableClassReference = r
}

// GetNonPortableClassReference of this CloudsqlInstance.
func (i *CloudsqlInstance) GetNonPortableClassReference() *corev1.ObjectReference {
	return i.Spec.NonPortableClassReference
}

// GetProviderReference of this CloudsqlInstance
func (i *CloudsqlInstance) GetProviderReference() *corev1.ObjectReference {
	return i.Spec.ProviderReference
}

// GetReclaimPolicy of this CloudsqlInstance.
func (i *CloudsqlInstance) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return i.Spec.ReclaimPolicy
}

// SetReclaimPolicy of this CloudsqlInstance.
func (i *CloudsqlInstance) SetReclaimPolicy(p runtimev1alpha1.ReclaimPolicy) {
	i.Spec.ReclaimPolicy = p
}

// SetWriteConnectionSecretToReference of this CloudsqlInstance.
func (i *CloudsqlInstance) SetWriteConnectionSecretToReference(r corev1.LocalObjectReference) {
	i.Spec.WriteConnectionSecretToReference = r
}

// GetWriteConnectionSecretToReference of this CloudsqlInstance.
func (i *CloudsqlInstance) GetWriteConnectionSecretToReference() corev1.LocalObjectReference {
	return i.Spec.WriteConnectionSecretToReference
}

// +kubebuilder:object:root=true

// CloudsqlInstanceList contains a list of CloudsqlInstance
type CloudsqlInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudsqlInstance `json:"items"`
}

// ConnectionSecret returns a connection secret for this instance
func (i *CloudsqlInstance) ConnectionSecret() *corev1.Secret {
	s := resource.ConnectionSecretFor(i, CloudsqlInstanceGroupVersionKind)
	s.Data[PublicIPKey] = []byte(i.Status.PublicIP)
	s.Data[PrivateIPKey] = []byte(i.Status.PrivateIP)
	s.Data[runtimev1alpha1.ResourceCredentialsSecretUserKey] = []byte(i.DatabaseUserName())
	// NOTE: this is for backward compatibility. Please use PublicIPKey and PrivateIPKey going forward.
	s.Data[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(i.Status.PublicIP)
	// TODO(muvaf): we explicitly enforce use of private IP if it's available. But this should be configured
	// by resource class or claim.
	if i.Status.PrivateIP != "" {
		s.Data[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(i.Status.PrivateIP)
	}
	return s
}

// DatabaseInstance representing spec of this instance
func (i *CloudsqlInstance) DatabaseInstance(name string) *sqladmin.DatabaseInstance {
	authnets := make([]*sqladmin.AclEntry, len(i.Spec.AuthorizedNetworks))
	for i, v := range i.Spec.AuthorizedNetworks {
		authnets[i] = &sqladmin.AclEntry{Value: v}
	}

	return &sqladmin.DatabaseInstance{
		Name:            name,
		Region:          i.Spec.Region,
		DatabaseVersion: i.Spec.DatabaseVersion,
		Settings: &sqladmin.Settings{
			Tier:           i.Spec.Tier,
			DataDiskType:   i.Spec.StorageType,
			DataDiskSizeGb: i.Spec.StorageGB,
			IpConfiguration: &sqladmin.IpConfiguration{
				AuthorizedNetworks: authnets,
				PrivateNetwork:     i.Spec.PrivateNetwork,
				Ipv4Enabled:        i.Spec.Ipv4Enabled,
				// NOTE: if we don't send false value explicitly, the default on GCP is true as opposed to
				// golang zero value of this type.
				ForceSendFields: []string{"Ipv4Enabled"},
			},
			UserLabels: i.Spec.Labels,
		},
	}
}

// DatabaseUserName returns default database user name base on database version
func (i *CloudsqlInstance) DatabaseUserName() string {
	if strings.HasPrefix(i.Spec.DatabaseVersion, PostgresqlDBVersionPrefix) {
		return PostgresqlDefaultUser
	}
	return MysqlDefaultUser
}

// GetResourceName based on the NameFormat spec value,
// If name format is not provided, resource name defaults to {{kind}}-UID
// If name format provided with '%s' value, resource name will result in formatted string + UID,
//   NOTE: only single %s substitution is supported
// If name format does not contain '%s' substitution, i.e. a constant string, the
// constant string value is returned back
//
// Examples:
//   For all examples assume "UID" = "test-uid",
//   and assume that "{{kind}}" = "mykind"
//   1. NameFormat = "", ResourceName = "mykind-test-uid"
//   2. NameFormat = "%s", ResourceName = "test-uid"
//   3. NameFormat = "foo", ResourceName = "foo"
//   4. NameFormat = "foo-%s", ResourceName = "foo-test-uid"
//   5. NameFormat = "foo-%s-bar-%s", ResourceName = "foo-test-uid-bar-%!s(MISSING)"
//
// Note that CloudSQL instance names must begin with a letter, per:
// https://cloud.google.com/sql/docs/mysql/instance-settings
func (i *CloudsqlInstance) GetResourceName() string {
	instanceNameFormatString := i.Spec.NameFormat

	if instanceNameFormatString == "" {
		instanceNameFormatString = strings.ToLower(CloudsqlInstanceKind) + "-%s"
	}

	return util.ConditionalStringFormat(instanceNameFormatString, string(i.GetUID()))
}

// IsRunnable returns true if instance is in Runnable state
func (i *CloudsqlInstance) IsRunnable() bool {
	return i.Status.State == StateRunnable
}

// SetStatus and Available condition, and other fields base on the provided database instance
func (i *CloudsqlInstance) SetStatus(inst *sqladmin.DatabaseInstance) {
	if inst == nil {
		return
	}
	i.Status.State = inst.State
	if i.IsRunnable() {
		i.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(i)
	} else {
		i.Status.SetConditions(runtimev1alpha1.Unavailable())
	}
	// TODO(muvaf): There might be cases where more than 1 private and/or public IP address has been assigned. We should
	// somehow show all addresses that are possible to use.
	for _, mapping := range inst.IpAddresses {
		switch mapping.Type {
		case PrivateIPType:
			i.Status.PrivateIP = mapping.IpAddress
		case PublicIPType:
			i.Status.PublicIP = mapping.IpAddress
		}
	}
}

// A CloudsqlInstanceClassSpecTemplate is a template for the spec of a
// dynamically provisioned CloudsqlInstance.
type CloudsqlInstanceClassSpecTemplate struct {
	runtimev1alpha1.NonPortableClassSpecTemplate `json:",inline"`
	CloudsqlInstanceParameters                   `json:",inline"`
}

// All non-portable classes must implement the NonPortableClass interface.
var _ resource.NonPortableClass = &CloudsqlInstanceClass{}

// +kubebuilder:object:root=true

// A CloudsqlInstanceClass is a non-portable resource class. It defines the
// desired spec of resource claims that use it to dynamically provision a
// managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type CloudsqlInstanceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// CloudsqlInstance.
	SpecTemplate CloudsqlInstanceClassSpecTemplate `json:"specTemplate"`
}

// GetReclaimPolicy of this CloudsqlInstanceClass.
func (i *CloudsqlInstanceClass) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return i.SpecTemplate.ReclaimPolicy
}

// SetReclaimPolicy of this CloudsqlInstanceClass.
func (i *CloudsqlInstanceClass) SetReclaimPolicy(p runtimev1alpha1.ReclaimPolicy) {
	i.SpecTemplate.ReclaimPolicy = p
}

// +kubebuilder:object:root=true

// CloudsqlInstanceClassList contains a list of cloud memorystore resource classes.
type CloudsqlInstanceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudsqlInstanceClass `json:"items"`
}
