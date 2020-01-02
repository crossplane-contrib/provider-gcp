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

package cloudsql

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

// Cyclomatic complexity test is disabled for translation methods
// because all they do is simple comparison & assignment without
// real logic. But every if statement increases the cyclomatic
// complexity rate.

// GenerateDatabaseInstance generates *sqladmin.DatabaseInstance instance from CloudSQLInstanceParameters.
func GenerateDatabaseInstance(in v1beta1.CloudSQLInstanceParameters, name string) *sqladmin.DatabaseInstance { // nolint:gocyclo
	db := &sqladmin.DatabaseInstance{
		DatabaseVersion:    gcp.StringValue(in.DatabaseVersion),
		GceZone:            gcp.StringValue(in.GceZone),
		InstanceType:       gcp.StringValue(in.InstanceType),
		MasterInstanceName: gcp.StringValue(in.MasterInstanceName),
		MaxDiskSize:        gcp.Int64Value(in.MaxDiskSize),
		Name:               name,
		Region:             in.Region,
		ReplicaNames:       in.ReplicaNames,
		SuspensionReason:   in.SuspensionReason,
	}
	if in.DiskEncryptionConfiguration != nil {
		db.DiskEncryptionConfiguration = &sqladmin.DiskEncryptionConfiguration{
			KmsKeyName: in.DiskEncryptionConfiguration.KmsKeyName,
		}
	}
	if in.FailoverReplica != nil {
		db.FailoverReplica = &sqladmin.DatabaseInstanceFailoverReplica{
			Name: in.FailoverReplica.Name,
		}
	}
	if in.OnPremisesConfiguration != nil {
		db.OnPremisesConfiguration = &sqladmin.OnPremisesConfiguration{
			HostPort: in.OnPremisesConfiguration.HostPort,
		}
	}
	db.Settings = &sqladmin.Settings{
		ActivationPolicy:            gcp.StringValue(in.Settings.ActivationPolicy),
		AuthorizedGaeApplications:   in.Settings.AuthorizedGaeApplications,
		AvailabilityType:            gcp.StringValue(in.Settings.AvailabilityType),
		CrashSafeReplicationEnabled: gcp.BoolValue(in.Settings.CrashSafeReplicationEnabled),
		DataDiskSizeGb:              gcp.Int64Value(in.Settings.DataDiskSizeGb),
		DataDiskType:                gcp.StringValue(in.Settings.DataDiskType),
		DatabaseReplicationEnabled:  gcp.BoolValue(in.Settings.DatabaseReplicationEnabled),
		PricingPlan:                 gcp.StringValue(in.Settings.PricingPlan),
		ReplicationType:             gcp.StringValue(in.Settings.ReplicationType),
		StorageAutoResize:           in.Settings.StorageAutoResize,
		StorageAutoResizeLimit:      gcp.Int64Value(in.Settings.StorageAutoResizeLimit),
		Tier:                        in.Settings.Tier,
		UserLabels:                  in.Settings.UserLabels,
	}
	if in.Settings.BackupConfiguration != nil {
		db.Settings.BackupConfiguration = &sqladmin.BackupConfiguration{
			BinaryLogEnabled:               gcp.BoolValue(in.Settings.BackupConfiguration.BinaryLogEnabled),
			Enabled:                        gcp.BoolValue(in.Settings.BackupConfiguration.Enabled),
			Location:                       gcp.StringValue(in.Settings.BackupConfiguration.Location),
			ReplicationLogArchivingEnabled: gcp.BoolValue(in.Settings.BackupConfiguration.ReplicationLogArchivingEnabled),
			StartTime:                      gcp.StringValue(in.Settings.BackupConfiguration.StartTime),
		}
	}
	if in.Settings.IPConfiguration != nil {
		db.Settings.IpConfiguration = &sqladmin.IpConfiguration{
			Ipv4Enabled:     gcp.BoolValue(in.Settings.IPConfiguration.Ipv4Enabled),
			PrivateNetwork:  gcp.StringValue(in.Settings.IPConfiguration.PrivateNetwork),
			RequireSsl:      gcp.BoolValue(in.Settings.IPConfiguration.RequireSsl),
			ForceSendFields: []string{"ipv4Enabled"},
		}
		for _, val := range in.Settings.IPConfiguration.AuthorizedNetworks {
			acl := &sqladmin.AclEntry{
				ExpirationTime: gcp.StringValue(val.ExpirationTime),
				Name:           gcp.StringValue(val.Name),
				Value:          gcp.StringValue(val.Value),
			}
			db.Settings.IpConfiguration.AuthorizedNetworks = append(db.Settings.IpConfiguration.AuthorizedNetworks, acl)
		}
	}
	if in.Settings.LocationPreference != nil {
		db.Settings.LocationPreference = &sqladmin.LocationPreference{
			FollowGaeApplication: gcp.StringValue(in.Settings.LocationPreference.FollowGaeApplication),
			Zone:                 gcp.StringValue(in.Settings.LocationPreference.Zone),
		}
	}
	if in.Settings.MaintenanceWindow != nil {
		db.Settings.MaintenanceWindow = &sqladmin.MaintenanceWindow{
			Day:         gcp.Int64Value(in.Settings.MaintenanceWindow.Day),
			Hour:        gcp.Int64Value(in.Settings.MaintenanceWindow.Hour),
			UpdateTrack: gcp.StringValue(in.Settings.MaintenanceWindow.UpdateTrack),
		}
	}
	for _, val := range in.Settings.DatabaseFlags {
		db.Settings.DatabaseFlags = append(db.Settings.DatabaseFlags, &sqladmin.DatabaseFlags{
			Name:  val.Name,
			Value: val.Value,
		})
	}
	return db
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
	spec.ReplicaNames = gcp.LateInitializeStringSlice(spec.ReplicaNames, in.ReplicaNames)
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
func IsUpToDate(in *v1beta1.CloudSQLInstanceParameters, currentState sqladmin.DatabaseInstance) bool {
	currentParams := &v1beta1.CloudSQLInstanceParameters{}
	LateInitializeSpec(currentParams, currentState)
	return cmp.Equal(in, currentParams, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{}))
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
