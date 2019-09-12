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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudsqlInstanceParameters defines the desired state of CloudsqlInstance
type CloudsqlInstanceParameters struct {
	AuthorizedNetworks []string `json:"authorizedNetworks,omitempty"`

	// PrivateNetwork: The resource link for the VPC network from which the
	// Cloud SQL instance is accessible for private IP. For example,
	// /projects/myProject/global/networks/default. This setting can be
	// updated, but it cannot be removed after it is set.
	PrivateNetwork string `json:"privateNetwork,omitempty"`

	// Ipv4Enabled: Whether the instance should be assigned an IP address or
	// not.
	Ipv4Enabled bool `json:"ipv4Enabled,omitempty"`

	// The database engine (MySQL or PostgreSQL) and its specific version to use, e.g., MYSQL_5_7 or POSTGRES_9_6.
	DatabaseVersion string `json:"databaseVersion"`

	Labels      map[string]string `json:"labels,omitempty"`
	Region      string            `json:"region"`
	StorageType string            `json:"storageType"`
	StorageGB   int64             `json:"storageGB"`

	// MySQL and PostgreSQL use different machine types.  MySQL only allows a predefined set of machine types,
	// while PostgreSQL can only use custom machine instance types and shared-core instance types. For the full
	// set of MySQL machine types, see https://cloud.google.com/sql/pricing#2nd-gen-instance-pricing. For more
	// information on custom machine types that can be used with PostgreSQL, see the examples on
	// https://cloud.google.com/sql/docs/postgres/create-instance?authuser=1#machine-types and the naming rules
	// on https://cloud.google.com/sql/docs/postgres/create-instance#create-2ndgen-curl.
	Tier string `json:"tier"`

	// TODO(illya) - this should be defined in ResourceSpec

	// NameFormat to format resource name passing it a object UID
	// If not provided, defaults to "%s", i.e. UID value
	NameFormat string `json:"nameFormat,omitempty"`
}

// CloudsqlInstanceSpec defines the desired state of CloudsqlInstance
type CloudsqlInstanceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	CloudsqlInstanceParameters   `json:",inline"`
}

// CloudsqlInstanceStatus defines the observed state of CloudsqlInstance
type CloudsqlInstanceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	State string `json:"state,omitempty"`

	// TODO(muvaf): Convert these types to *string during managed reconciler refactor because both are optional.

	// PublicIP is used to connect to this resource from other authorized networks.
	PublicIP string `json:"publicIp,omitempty"`
	// PrivateIP is used to connect to this instance from the same Network.
	PrivateIP string `json:"privateIp,omitempty"`
}

// +kubebuilder:object:root=true

// CloudsqlInstance is the Schema for the instances API
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

// SetClassReference of this CloudsqlInstance.
func (i *CloudsqlInstance) SetClassReference(r *corev1.ObjectReference) {
	i.Spec.ClassReference = r
}

// GetClassReference of this CloudsqlInstance.
func (i *CloudsqlInstance) GetClassReference() *corev1.ObjectReference {
	return i.Spec.ClassReference
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

// CloudsqlInstanceClassSpecTemplate is the Schema for the resource class
type CloudsqlInstanceClassSpecTemplate struct {
	runtimev1alpha1.ResourceClassSpecTemplate `json:",inline"`
	CloudsqlInstanceParameters                `json:",inline"`
}

var _ resource.Class = &CloudsqlInstanceClass{}

// +kubebuilder:object:root=true

// CloudsqlInstanceClass is the Schema for the resource class
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type CloudsqlInstanceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	SpecTemplate CloudsqlInstanceClassSpecTemplate `json:"specTemplate,omitempty"`
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
