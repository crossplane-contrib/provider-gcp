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

package cloudsql

import (
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/crossplane/provider-gcp/apis/database/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// Cyclomatic complexity test is disabled for translation methods
// because all they do is simple comparison & assignment without
// real logic. But every if statement increases the cyclomatic
// complexity rate.

// GenerateDatabaseInstance generates *sqladmin.DatabaseInstance instance from CloudSQLInstanceParameters.
func GenerateDatabaseInstance(name string, in v1beta1.CloudSQLInstanceParameters, db *sqladmin.DatabaseInstance) { // nolint:gocyclo
	db.DatabaseVersion = gcp.StringValue(in.DatabaseVersion)
	db.GceZone = gcp.StringValue(in.GceZone)
	db.InstanceType = gcp.StringValue(in.InstanceType)
	db.MasterInstanceName = gcp.StringValue(in.MasterInstanceName)
	db.MaxDiskSize = gcp.Int64Value(in.MaxDiskSize)
	db.Name = name
	db.Region = in.Region
	db.ReplicaNames = in.ReplicaNames
	db.SuspensionReason = in.SuspensionReason
	if in.DiskEncryptionConfiguration != nil {
		if db.DiskEncryptionConfiguration == nil {
			db.DiskEncryptionConfiguration = &sqladmin.DiskEncryptionConfiguration{}
		}
		db.DiskEncryptionConfiguration.KmsKeyName = in.DiskEncryptionConfiguration.KmsKeyName
	}
	if in.FailoverReplica != nil {
		if db.FailoverReplica == nil {
			db.FailoverReplica = &sqladmin.DatabaseInstanceFailoverReplica{}
		}
		db.FailoverReplica.Name = in.FailoverReplica.Name
	}
	if in.OnPremisesConfiguration != nil {
		if db.OnPremisesConfiguration == nil {
			db.OnPremisesConfiguration = &sqladmin.OnPremisesConfiguration{}
		}
		db.OnPremisesConfiguration.HostPort = in.OnPremisesConfiguration.HostPort
	}
	if db.Settings == nil {
		db.Settings = &sqladmin.Settings{}
	}
	db.Settings.ActivationPolicy = gcp.StringValue(in.Settings.ActivationPolicy)
	db.Settings.AuthorizedGaeApplications = in.Settings.AuthorizedGaeApplications
	db.Settings.AvailabilityType = gcp.StringValue(in.Settings.AvailabilityType)
	db.Settings.CrashSafeReplicationEnabled = gcp.BoolValue(in.Settings.CrashSafeReplicationEnabled)
	db.Settings.DataDiskSizeGb = gcp.Int64Value(in.Settings.DataDiskSizeGb)
	db.Settings.DataDiskType = gcp.StringValue(in.Settings.DataDiskType)
	db.Settings.DatabaseReplicationEnabled = gcp.BoolValue(in.Settings.DatabaseReplicationEnabled)
	db.Settings.PricingPlan = gcp.StringValue(in.Settings.PricingPlan)
	db.Settings.ReplicationType = gcp.StringValue(in.Settings.ReplicationType)
	db.Settings.StorageAutoResize = in.Settings.StorageAutoResize
	db.Settings.StorageAutoResizeLimit = gcp.Int64Value(in.Settings.StorageAutoResizeLimit)
	db.Settings.Tier = in.Settings.Tier
	db.Settings.UserLabels = in.Settings.UserLabels

	if in.Settings.BackupConfiguration != nil {
		if db.Settings.BackupConfiguration == nil {
			db.Settings.BackupConfiguration = &sqladmin.BackupConfiguration{}
		}
		db.Settings.BackupConfiguration.BinaryLogEnabled = gcp.BoolValue(in.Settings.BackupConfiguration.BinaryLogEnabled)
		db.Settings.BackupConfiguration.Enabled = gcp.BoolValue(in.Settings.BackupConfiguration.Enabled)
		db.Settings.BackupConfiguration.Location = gcp.StringValue(in.Settings.BackupConfiguration.Location)
		db.Settings.BackupConfiguration.ReplicationLogArchivingEnabled = gcp.BoolValue(in.Settings.BackupConfiguration.ReplicationLogArchivingEnabled)
		db.Settings.BackupConfiguration.StartTime = gcp.StringValue(in.Settings.BackupConfiguration.StartTime)
		db.Settings.BackupConfiguration.PointInTimeRecoveryEnabled = gcp.BoolValue(in.Settings.BackupConfiguration.PointInTimeRecoveryEnabled)
	}
	if in.Settings.IPConfiguration != nil {
		if db.Settings.IpConfiguration == nil {
			db.Settings.IpConfiguration = &sqladmin.IpConfiguration{}
		}
		db.Settings.IpConfiguration.Ipv4Enabled = gcp.BoolValue(in.Settings.IPConfiguration.Ipv4Enabled)
		db.Settings.IpConfiguration.PrivateNetwork = gcp.StringValue(in.Settings.IPConfiguration.PrivateNetwork)
		db.Settings.IpConfiguration.RequireSsl = gcp.BoolValue(in.Settings.IPConfiguration.RequireSsl)
		db.Settings.IpConfiguration.ForceSendFields = []string{"Ipv4Enabled"}

		if len(in.Settings.IPConfiguration.AuthorizedNetworks) > 0 {
			db.Settings.IpConfiguration.AuthorizedNetworks = make([]*sqladmin.AclEntry, len(in.Settings.IPConfiguration.AuthorizedNetworks))
		}
		for i, val := range in.Settings.IPConfiguration.AuthorizedNetworks {
			db.Settings.IpConfiguration.AuthorizedNetworks[i] = &sqladmin.AclEntry{
				ExpirationTime: gcp.StringValue(val.ExpirationTime),
				Name:           gcp.StringValue(val.Name),
				Value:          gcp.StringValue(val.Value),
				Kind:           "sql#aclEntry",
			}
		}
	}
	if in.Settings.LocationPreference != nil {
		if db.Settings.LocationPreference == nil {
			db.Settings.LocationPreference = &sqladmin.LocationPreference{}
		}
		db.Settings.LocationPreference.FollowGaeApplication = gcp.StringValue(in.Settings.LocationPreference.FollowGaeApplication)
		db.Settings.LocationPreference.Zone = gcp.StringValue(in.Settings.LocationPreference.Zone)
	}
	if in.Settings.MaintenanceWindow != nil {
		if db.Settings.MaintenanceWindow == nil {
			db.Settings.MaintenanceWindow = &sqladmin.MaintenanceWindow{}
		}
		db.Settings.MaintenanceWindow.Day = gcp.Int64Value(in.Settings.MaintenanceWindow.Day)
		db.Settings.MaintenanceWindow.Hour = gcp.Int64Value(in.Settings.MaintenanceWindow.Hour)
		db.Settings.MaintenanceWindow.UpdateTrack = gcp.StringValue(in.Settings.MaintenanceWindow.UpdateTrack)
	}
	if len(in.Settings.DatabaseFlags) > 0 {
		db.Settings.DatabaseFlags = make([]*sqladmin.DatabaseFlags, len(in.Settings.DatabaseFlags))
	}
	for i, val := range in.Settings.DatabaseFlags {
		db.Settings.DatabaseFlags[i] = &sqladmin.DatabaseFlags{
			Name:  val.Name,
			Value: val.Value,
		}
	}
}

// GenerateObservation produces CloudSQLInstanceObservation object from *sqladmin.DatabaseInstance object.
func GenerateObservation(in sqladmin.DatabaseInstance) v1beta1.CloudSQLInstanceObservation { // nolint:gocyclo
	o := v1beta1.CloudSQLInstanceObservation{
		BackendType:                in.BackendType,
		CurrentDiskSize:            in.CurrentDiskSize,
		ConnectionName:             in.ConnectionName,
		GceZone:                    in.GceZone,
		IPv6Address:                in.Ipv6Address,
		Project:                    in.Project,
		SelfLink:                   in.SelfLink,
		ServiceAccountEmailAddress: in.ServiceAccountEmailAddress,
		State:                      in.State,
		SettingsVersion:            in.Settings.SettingsVersion,
	}
	if in.DiskEncryptionStatus != nil {
		o.DiskEncryptionStatus = &v1beta1.DiskEncryptionStatus{
			KmsKeyVersionName: in.DiskEncryptionStatus.KmsKeyVersionName,
		}
	}
	if in.FailoverReplica != nil {
		o.FailoverReplica = &v1beta1.DatabaseInstanceFailoverReplicaStatus{
			Available: in.FailoverReplica.Available,
		}
	}
	for _, val := range in.IpAddresses {
		o.IPAddresses = append(o.IPAddresses, &v1beta1.IPMapping{
			IPAddress:    val.IpAddress,
			TimeToRetire: val.TimeToRetire,
			Type:         val.Type,
		})
	}
	return o
}

// LateInitializeSpec fills unassigned fields with the values in sqladmin.DatabaseInstance object.
func LateInitializeSpec(spec *v1beta1.CloudSQLInstanceParameters, in sqladmin.DatabaseInstance) { // nolint:gocyclo

	// TODO(muvaf): One can marshall both objects into json and compare them as dictionaries since
	//  they both have the same key names but this may create performance problems as it'll happen in each
	//  reconcile. learn code-generation to make writing this easier and performant.
	if spec.Region == "" {
		spec.Region = in.Region
	}
	spec.DatabaseVersion = gcp.LateInitializeString(spec.DatabaseVersion, in.DatabaseVersion)
	spec.MasterInstanceName = gcp.LateInitializeString(spec.MasterInstanceName, in.MasterInstanceName)
	spec.GceZone = gcp.LateInitializeString(spec.GceZone, in.GceZone)
	spec.InstanceType = gcp.LateInitializeString(spec.InstanceType, in.InstanceType)
	spec.MaxDiskSize = gcp.LateInitializeInt64(spec.MaxDiskSize, in.MaxDiskSize)
	spec.ReplicaNames = in.ReplicaNames
	spec.SuspensionReason = gcp.LateInitializeStringSlice(spec.SuspensionReason, in.SuspensionReason)
	if in.Settings != nil {
		if spec.Settings.Tier == "" {
			spec.Settings.Tier = in.Settings.Tier
		}
		spec.Settings.ActivationPolicy = gcp.LateInitializeString(spec.Settings.ActivationPolicy, in.Settings.ActivationPolicy)
		spec.Settings.AuthorizedGaeApplications = gcp.LateInitializeStringSlice(spec.Settings.AuthorizedGaeApplications, in.Settings.AuthorizedGaeApplications)
		spec.Settings.AvailabilityType = gcp.LateInitializeString(spec.Settings.AvailabilityType, in.Settings.AvailabilityType)
		spec.Settings.CrashSafeReplicationEnabled = gcp.LateInitializeBool(spec.Settings.CrashSafeReplicationEnabled, in.Settings.CrashSafeReplicationEnabled)

		spec.Settings.DataDiskType = gcp.LateInitializeString(spec.Settings.DataDiskType, in.Settings.DataDiskType)
		spec.Settings.PricingPlan = gcp.LateInitializeString(spec.Settings.PricingPlan, in.Settings.PricingPlan)
		spec.Settings.ReplicationType = gcp.LateInitializeString(spec.Settings.ReplicationType, in.Settings.ReplicationType)
		spec.Settings.UserLabels = gcp.LateInitializeStringMap(spec.Settings.UserLabels, in.Settings.UserLabels)
		spec.Settings.DataDiskSizeGb = gcp.LateInitializeInt64(spec.Settings.DataDiskSizeGb, in.Settings.DataDiskSizeGb)
		spec.Settings.DatabaseReplicationEnabled = gcp.LateInitializeBool(spec.Settings.DatabaseReplicationEnabled, in.Settings.DatabaseReplicationEnabled)
		spec.Settings.StorageAutoResizeLimit = gcp.LateInitializeInt64(spec.Settings.StorageAutoResizeLimit, in.Settings.StorageAutoResizeLimit)
		if spec.Settings.StorageAutoResize == nil {
			spec.Settings.StorageAutoResize = in.Settings.StorageAutoResize
		}
		// If storage auto resize enabled, GCP does not allow setting a smaller
		// size but allows increasing it. Here, we set desired size as observed
		// if it is bigger than the current value which would allows us to get
		// in sync with the actual value but still allow us to increase it.
		if gcp.BoolValue(spec.Settings.StorageAutoResize) && gcp.Int64Value(spec.Settings.DataDiskSizeGb) < in.Settings.DataDiskSizeGb {
			spec.Settings.DataDiskSizeGb = gcp.Int64Ptr(in.Settings.DataDiskSizeGb)
		}
		if len(spec.Settings.DatabaseFlags) == 0 && len(in.Settings.DatabaseFlags) != 0 {
			spec.Settings.DatabaseFlags = make([]*v1beta1.DatabaseFlags, len(in.Settings.DatabaseFlags))
			for i, val := range in.Settings.DatabaseFlags {
				spec.Settings.DatabaseFlags[i] = &v1beta1.DatabaseFlags{
					Name:  val.Name,
					Value: val.Value,
				}
			}
		}
		if in.Settings.BackupConfiguration != nil {
			if spec.Settings.BackupConfiguration == nil {
				spec.Settings.BackupConfiguration = &v1beta1.BackupConfiguration{}
			}
			spec.Settings.BackupConfiguration.BinaryLogEnabled = gcp.LateInitializeBool(
				spec.Settings.BackupConfiguration.BinaryLogEnabled,
				in.Settings.BackupConfiguration.BinaryLogEnabled)
			spec.Settings.BackupConfiguration.Enabled = gcp.LateInitializeBool(
				spec.Settings.BackupConfiguration.Enabled,
				in.Settings.BackupConfiguration.Enabled)
			spec.Settings.BackupConfiguration.Location = gcp.LateInitializeString(
				spec.Settings.BackupConfiguration.Location,
				in.Settings.BackupConfiguration.Location)
			spec.Settings.BackupConfiguration.ReplicationLogArchivingEnabled = gcp.LateInitializeBool(
				spec.Settings.BackupConfiguration.ReplicationLogArchivingEnabled,
				in.Settings.BackupConfiguration.ReplicationLogArchivingEnabled)
			spec.Settings.BackupConfiguration.StartTime = gcp.LateInitializeString(
				spec.Settings.BackupConfiguration.StartTime,
				in.Settings.BackupConfiguration.StartTime)
			spec.Settings.BackupConfiguration.PointInTimeRecoveryEnabled = gcp.LateInitializeBool(
				spec.Settings.BackupConfiguration.PointInTimeRecoveryEnabled,
				in.Settings.BackupConfiguration.PointInTimeRecoveryEnabled)
		}
		if in.Settings.IpConfiguration != nil {
			if spec.Settings.IPConfiguration == nil {
				spec.Settings.IPConfiguration = &v1beta1.IPConfiguration{}
			}
			spec.Settings.IPConfiguration.Ipv4Enabled = gcp.LateInitializeBool(spec.Settings.IPConfiguration.Ipv4Enabled, in.Settings.IpConfiguration.Ipv4Enabled)
			spec.Settings.IPConfiguration.PrivateNetwork = gcp.LateInitializeString(spec.Settings.IPConfiguration.PrivateNetwork, in.Settings.IpConfiguration.PrivateNetwork)
			spec.Settings.IPConfiguration.RequireSsl = gcp.LateInitializeBool(spec.Settings.IPConfiguration.RequireSsl, in.Settings.IpConfiguration.RequireSsl)
			if len(in.Settings.IpConfiguration.AuthorizedNetworks) != 0 && len(spec.Settings.IPConfiguration.AuthorizedNetworks) == 0 {
				spec.Settings.IPConfiguration.AuthorizedNetworks = make([]*v1beta1.ACLEntry, len(in.Settings.IpConfiguration.AuthorizedNetworks))
				for i, val := range in.Settings.IpConfiguration.AuthorizedNetworks {
					spec.Settings.IPConfiguration.AuthorizedNetworks[i] = &v1beta1.ACLEntry{
						ExpirationTime: &val.ExpirationTime,
						Name:           &val.Name,
						Value:          &val.Value,
					}
				}
			}
		}
		if in.Settings.LocationPreference != nil {
			if spec.Settings.LocationPreference == nil {
				spec.Settings.LocationPreference = &v1beta1.LocationPreference{}
			}
			spec.Settings.LocationPreference.Zone = gcp.LateInitializeString(spec.Settings.LocationPreference.Zone, in.Settings.LocationPreference.Zone)
			spec.Settings.LocationPreference.FollowGaeApplication = gcp.LateInitializeString(spec.Settings.LocationPreference.FollowGaeApplication, in.Settings.LocationPreference.FollowGaeApplication)

		}
		if in.Settings.MaintenanceWindow != nil {
			if spec.Settings.MaintenanceWindow == nil {
				spec.Settings.MaintenanceWindow = &v1beta1.MaintenanceWindow{}
			}
			spec.Settings.MaintenanceWindow.UpdateTrack = gcp.LateInitializeString(spec.Settings.MaintenanceWindow.UpdateTrack, in.Settings.MaintenanceWindow.UpdateTrack)
			spec.Settings.MaintenanceWindow.Day = gcp.LateInitializeInt64(spec.Settings.MaintenanceWindow.Day, in.Settings.MaintenanceWindow.Day)
			spec.Settings.MaintenanceWindow.Hour = gcp.LateInitializeInt64(spec.Settings.MaintenanceWindow.Hour, in.Settings.MaintenanceWindow.Hour)
		}
	}
	if in.DiskEncryptionConfiguration != nil {
		if spec.DiskEncryptionConfiguration == nil {
			spec.DiskEncryptionConfiguration = &v1beta1.DiskEncryptionConfiguration{}
		}
		if spec.DiskEncryptionConfiguration.KmsKeyName == "" {
			spec.DiskEncryptionConfiguration.KmsKeyName = in.DiskEncryptionConfiguration.KmsKeyName
		}
	}
	if in.FailoverReplica != nil {
		if spec.FailoverReplica == nil {
			spec.FailoverReplica = &v1beta1.DatabaseInstanceFailoverReplicaSpec{
				Name: in.FailoverReplica.Name,
			}
		}
	}
	if in.OnPremisesConfiguration != nil {
		if spec.OnPremisesConfiguration == nil {
			spec.OnPremisesConfiguration = &v1beta1.OnPremisesConfiguration{
				HostPort: in.OnPremisesConfiguration.HostPort,
			}
		}
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, in *v1beta1.CloudSQLInstanceParameters, observed *sqladmin.DatabaseInstance) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*sqladmin.DatabaseInstance)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateDatabaseInstance(name, *in, desired)
	return cmp.Equal(desired, observed, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(sqladmin.DatabaseInstance{}, "Settings.IpConfiguration.ForceSendFields")), nil
}

// DatabaseUserName returns default database user name base on database version
func DatabaseUserName(p v1beta1.CloudSQLInstanceParameters) string {
	if strings.HasPrefix(gcp.StringValue(p.DatabaseVersion), v1beta1.PostgresqlDBVersionPrefix) {
		return v1beta1.PostgresqlDefaultUser
	}
	return v1beta1.MysqlDefaultUser
}

// GetServerCACertificate takes sqladmin.DatabaseInstance and returns the server CA certificate
// in a form that can be embedded directly into a connection secret.
func GetServerCACertificate(in sqladmin.DatabaseInstance) map[string][]byte {
	if in.ServerCaCert == nil {
		return nil
	}
	return map[string][]byte{
		v1beta1.CloudSQLSecretServerCACertificateCertKey:             []byte(in.ServerCaCert.Cert),
		v1beta1.CloudSQLSecretServerCACertificateCertSerialNumberKey: []byte(in.ServerCaCert.CertSerialNumber),
		v1beta1.CloudSQLSecretServerCACertificateCommonNameKey:       []byte(in.ServerCaCert.CommonName),
		v1beta1.CloudSQLSecretServerCACertificateCreateTimeKey:       []byte(in.ServerCaCert.CreateTime),
		v1beta1.CloudSQLSecretServerCACertificateExpirationTimeKey:   []byte(in.ServerCaCert.ExpirationTime),
		v1beta1.CloudSQLSecretServerCACertificateInstanceKey:         []byte(in.ServerCaCert.Instance),
		v1beta1.CloudSQLSecretServerCACertificateSha1FingerprintKey:  []byte(in.ServerCaCert.Sha1Fingerprint),
	}
}
