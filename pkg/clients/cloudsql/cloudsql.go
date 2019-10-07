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
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/crossplaneio/stack-gcp/apis/database/v1alpha2"
)

// GenerateDatabaseInstance generates *sqladmin.DatabaseInstance instance from CloudsqlInstanceParameters.
func GenerateDatabaseInstance(in v1alpha2.CloudsqlInstanceParameters, name string) *sqladmin.DatabaseInstance { // nolint:gocyclo
	db := &sqladmin.DatabaseInstance{
		BackendType:                in.BackendType,
		ConnectionName:             in.ConnectionName,
		DatabaseVersion:            in.DatabaseVersion,
		Etag:                       in.Etag,
		GceZone:                    in.GceZone,
		InstanceType:               in.InstanceType,
		MasterInstanceName:         in.MasterInstanceName,
		MaxDiskSize:                in.MaxDiskSize,
		Name:                       name,
		Region:                     in.Region,
		ReplicaNames:               in.ReplicaNames,
		RootPassword:               in.RootPassword,
		ServiceAccountEmailAddress: in.ServiceAccountEmailAddress,
		SuspensionReason:           in.SuspensionReason,
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
	if in.ReplicaConfiguration != nil {
		db.ReplicaConfiguration = &sqladmin.ReplicaConfiguration{
			FailoverTarget: in.ReplicaConfiguration.FailoverTarget,
		}
		if in.ReplicaConfiguration.MySQLReplicaConfiguration != nil {
			db.ReplicaConfiguration.MysqlReplicaConfiguration = &sqladmin.MySqlReplicaConfiguration{
				CaCertificate:           in.ReplicaConfiguration.MySQLReplicaConfiguration.CaCertificate,
				ClientCertificate:       in.ReplicaConfiguration.MySQLReplicaConfiguration.ClientCertificate,
				ClientKey:               in.ReplicaConfiguration.MySQLReplicaConfiguration.ClientKey,
				ConnectRetryInterval:    in.ReplicaConfiguration.MySQLReplicaConfiguration.ConnectRetryInterval,
				DumpFilePath:            in.ReplicaConfiguration.MySQLReplicaConfiguration.DumpFilePath,
				MasterHeartbeatPeriod:   in.ReplicaConfiguration.MySQLReplicaConfiguration.MasterHeartbeatPeriod,
				Password:                in.ReplicaConfiguration.MySQLReplicaConfiguration.Password,
				SslCipher:               in.ReplicaConfiguration.MySQLReplicaConfiguration.SslCipher,
				Username:                in.ReplicaConfiguration.MySQLReplicaConfiguration.Username,
				VerifyServerCertificate: in.ReplicaConfiguration.MySQLReplicaConfiguration.VerifyServerCertificate,
			}
		}
	}
	if in.ServerCaCert != nil {
		db.ServerCaCert = &sqladmin.SslCert{
			Cert:             in.ServerCaCert.Cert,
			CertSerialNumber: in.ServerCaCert.CertSerialNumber,
			CommonName:       in.ServerCaCert.CommonName,
			CreateTime:       in.ServerCaCert.CreateTime,
			ExpirationTime:   in.ServerCaCert.ExpirationTime,
			Instance:         in.ServerCaCert.Instance,
			Sha1Fingerprint:  in.ServerCaCert.Sha1Fingerprint,
		}
	}
	if in.Settings != nil {
		db.Settings = &sqladmin.Settings{
			ActivationPolicy:            in.Settings.ActivationPolicy,
			AuthorizedGaeApplications:   in.Settings.AuthorizedGaeApplications,
			AvailabilityType:            in.Settings.AvailabilityType,
			CrashSafeReplicationEnabled: in.Settings.CrashSafeReplicationEnabled,
			DataDiskSizeGb:              in.Settings.DataDiskSizeGb,
			DataDiskType:                in.Settings.DataDiskType,
			DatabaseReplicationEnabled:  in.Settings.DatabaseReplicationEnabled,
			PricingPlan:                 in.Settings.PricingPlan,
			ReplicationType:             in.Settings.ReplicationType,
			StorageAutoResize:           in.Settings.StorageAutoResize,
			StorageAutoResizeLimit:      in.Settings.StorageAutoResizeLimit,
			Tier:                        in.Settings.Tier,
			UserLabels:                  in.Settings.UserLabels,
		}
		if in.Settings.BackupConfiguration != nil {
			db.Settings.BackupConfiguration = &sqladmin.BackupConfiguration{
				BinaryLogEnabled:               in.Settings.BackupConfiguration.BinaryLogEnabled,
				Enabled:                        in.Settings.BackupConfiguration.Enabled,
				Location:                       in.Settings.BackupConfiguration.Location,
				ReplicationLogArchivingEnabled: in.Settings.BackupConfiguration.ReplicationLogArchivingEnabled,
				StartTime:                      in.Settings.BackupConfiguration.StartTime,
			}
		}
		if in.Settings.IPConfiguration != nil {
			db.Settings.IpConfiguration = &sqladmin.IpConfiguration{
				Ipv4Enabled:    in.Settings.IPConfiguration.Ipv4Enabled,
				PrivateNetwork: in.Settings.IPConfiguration.PrivateNetwork,
				RequireSsl:     in.Settings.IPConfiguration.RequireSsl,
			}
		}
		if in.Settings.LocationPreference != nil {
			db.Settings.LocationPreference = &sqladmin.LocationPreference{
				FollowGaeApplication: in.Settings.LocationPreference.FollowGaeApplication,
				Zone:                 in.Settings.LocationPreference.Zone,
			}
		}
		if in.Settings.MaintenanceWindow != nil {
			db.Settings.MaintenanceWindow = &sqladmin.MaintenanceWindow{
				Day:         in.Settings.MaintenanceWindow.Day,
				Hour:        in.Settings.MaintenanceWindow.Hour,
				UpdateTrack: in.Settings.MaintenanceWindow.UpdateTrack,
			}
		}
	}
	for _, val := range in.Settings.DatabaseFlags {
		db.Settings.DatabaseFlags = append(db.Settings.DatabaseFlags, &sqladmin.DatabaseFlags{
			Name:  val.Name,
			Value: val.Value,
		})
	}
	for _, val := range in.Settings.IPConfiguration.AuthorizedNetworks {
		acl := &sqladmin.AclEntry{
			ExpirationTime: val.ExpirationTime,
			Name:           val.Name,
			Value:          val.Value,
		}
		db.Settings.IpConfiguration.AuthorizedNetworks = append(db.Settings.IpConfiguration.AuthorizedNetworks, acl)
	}
	return db
}

// GenerateObservation produces CloudsqlInstanceObservation object from *sqladmin.DatabaseInstance object.
func GenerateObservation(in sqladmin.DatabaseInstance) v1alpha2.CloudsqlInstanceObservation { // nolint:gocyclo
	o := v1alpha2.CloudsqlInstanceObservation{
		CurrentDiskSize: in.CurrentDiskSize,
		ConnectionName:  in.ConnectionName,
		GceZone:         in.GceZone,
		Ipv6Address:     in.Ipv6Address,
		Project:         in.Project,
		SelfLink:        in.SelfLink,
		State:           in.State,
		SettingsVersion: in.Settings.SettingsVersion,
	}
	if in.DiskEncryptionStatus != nil {
		o.DiskEncryptionStatus = &v1alpha2.DiskEncryptionStatus{
			KmsKeyVersionName: in.DiskEncryptionStatus.KmsKeyVersionName,
		}
	}
	if in.FailoverReplica != nil {
		o.FailoverReplica = &v1alpha2.DatabaseInstanceFailoverReplica{
			Available: in.FailoverReplica.Available,
			Name:      in.FailoverReplica.Name,
		}
	}
	for _, val := range in.IpAddresses {
		o.IPAddresses = append(o.IPAddresses, &v1alpha2.IPMapping{
			IPAddress:    val.IpAddress,
			TimeToRetire: val.TimeToRetire,
			Type:         val.Type,
		})
	}
	return o
}

func FillSpecWithDefaults(cr *v1alpha2.CloudsqlInstanceParameters, in sqladmin.DatabaseInstance) (changed bool) {
	// TODO(muvaf): find a way to avoid messy if statements for patching parameters with received DatabaseInstance.
	return true
}

func IsUpToDate(cr v1alpha2.CloudsqlInstanceParameters, in sqladmin.DatabaseInstance) bool {
	// TODO(muvaf): check whether spec is different than DatabaseInstance.
	return false
}