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
	corev1 "k8s.io/api/core/v1"
)

const (
	errCheckUpToDate = "unable to determine if external resource is up to date"
	errSecretKey   = "unable to determine secret defined in secretRef: %v"
) 
// CloudSQLOptions is wrapper type that holds all potentially passed params to the client
type CloudSQLOptions struct {
	Name     string
	Spec     *v1beta1.CloudSQLInstanceParameters
	Instance *sqladmin.DatabaseInstance
	Secret   *corev1.Secret
}
// Cyclomatic complexity test is disabled for translation methods
// because all they do is simple comparison & assignment without
// real logic. But every if statement increases the cyclomatic
// complexity rate.

// GenerateDatabaseInstance generates *sqladmin.DatabaseInstance instance from CloudSQLInstanceParameters.
func GenerateDatabaseInstance(opts CloudSQLOptions) error { // nolint:gocyclo
	opts.Instance.DatabaseVersion = gcp.StringValue(opts.Spec.DatabaseVersion)
	opts.Instance.GceZone = gcp.StringValue(opts.Spec.GceZone)
	opts.Instance.InstanceType = gcp.StringValue(opts.Spec.InstanceType)
	opts.Instance.MasterInstanceName = gcp.StringValue(opts.Spec.MasterInstanceName)
	opts.Instance.MaxDiskSize = gcp.Int64Value(opts.Spec.MaxDiskSize)
	opts.Instance.Name = opts.Name
	opts.Instance.Region = opts.Spec.Region
	opts.Instance.ReplicaNames = opts.Spec.ReplicaNames
	opts.Instance.SuspensionReason = opts.Spec.SuspensionReason
	if opts.Spec.DiskEncryptionConfiguration != nil {
		if opts.Instance.DiskEncryptionConfiguration == nil {
			opts.Instance.DiskEncryptionConfiguration = &sqladmin.DiskEncryptionConfiguration{}
		}
		opts.Instance.DiskEncryptionConfiguration.KmsKeyName = opts.Spec.DiskEncryptionConfiguration.KmsKeyName
	}
	if opts.Spec.FailoverReplica != nil {
		if opts.Instance.FailoverReplica == nil {
			opts.Instance.FailoverReplica = &sqladmin.DatabaseInstanceFailoverReplica{}
		}
		opts.Instance.FailoverReplica.Name = opts.Spec.FailoverReplica.Name
	}
	if opts.Spec.OnPremisesConfiguration != nil {
		if opts.Instance.OnPremisesConfiguration == nil {
			opts.Instance.OnPremisesConfiguration = &sqladmin.OnPremisesConfiguration{}
		}
		opts.Instance.OnPremisesConfiguration.HostPort = opts.Spec.OnPremisesConfiguration.HostPort
	}
	if opts.Instance.Settings == nil {
		opts.Instance.Settings = &sqladmin.Settings{}
	}
	opts.Instance.Settings.ActivationPolicy = gcp.StringValue(opts.Spec.Settings.ActivationPolicy)
	opts.Instance.Settings.AuthorizedGaeApplications = opts.Spec.Settings.AuthorizedGaeApplications
	opts.Instance.Settings.AvailabilityType = gcp.StringValue(opts.Spec.Settings.AvailabilityType)
	opts.Instance.Settings.CrashSafeReplicationEnabled = gcp.BoolValue(opts.Spec.Settings.CrashSafeReplicationEnabled)
	opts.Instance.Settings.DataDiskSizeGb = gcp.Int64Value(opts.Spec.Settings.DataDiskSizeGb)
	opts.Instance.Settings.DataDiskType = gcp.StringValue(opts.Spec.Settings.DataDiskType)
	opts.Instance.Settings.DatabaseReplicationEnabled = gcp.BoolValue(opts.Spec.Settings.DatabaseReplicationEnabled)
	opts.Instance.Settings.PricingPlan = gcp.StringValue(opts.Spec.Settings.PricingPlan)
	opts.Instance.Settings.ReplicationType = gcp.StringValue(opts.Spec.Settings.ReplicationType)
	opts.Instance.Settings.StorageAutoResize = opts.Spec.Settings.StorageAutoResize
	opts.Instance.Settings.StorageAutoResizeLimit = gcp.Int64Value(opts.Spec.Settings.StorageAutoResizeLimit)
	opts.Instance.Settings.Tier = opts.Spec.Settings.Tier
	opts.Instance.Settings.UserLabels = opts.Spec.Settings.UserLabels

	if opts.Spec.Settings.BackupConfiguration != nil {
		if opts.Instance.Settings.BackupConfiguration == nil {
			opts.Instance.Settings.BackupConfiguration = &sqladmin.BackupConfiguration{}
		}
		opts.Instance.Settings.BackupConfiguration.BinaryLogEnabled = gcp.BoolValue(opts.Spec.Settings.BackupConfiguration.BinaryLogEnabled)
		opts.Instance.Settings.BackupConfiguration.Enabled = gcp.BoolValue(opts.Spec.Settings.BackupConfiguration.Enabled)
		opts.Instance.Settings.BackupConfiguration.Location = gcp.StringValue(opts.Spec.Settings.BackupConfiguration.Location)
		opts.Instance.Settings.BackupConfiguration.ReplicationLogArchivingEnabled = gcp.BoolValue(opts.Spec.Settings.BackupConfiguration.ReplicationLogArchivingEnabled)
		opts.Instance.Settings.BackupConfiguration.StartTime = gcp.StringValue(opts.Spec.Settings.BackupConfiguration.StartTime)
		opts.Instance.Settings.BackupConfiguration.PointInTimeRecoveryEnabled = gcp.BoolValue(opts.Spec.Settings.BackupConfiguration.PointInTimeRecoveryEnabled)
	}
	if opts.Spec.Settings.IPConfiguration != nil {
		if opts.Instance.Settings.IpConfiguration == nil {
			opts.Instance.Settings.IpConfiguration = &sqladmin.IpConfiguration{}
		}
		opts.Instance.Settings.IpConfiguration.Ipv4Enabled = gcp.BoolValue(opts.Spec.Settings.IPConfiguration.Ipv4Enabled)
		opts.Instance.Settings.IpConfiguration.PrivateNetwork = gcp.StringValue(opts.Spec.Settings.IPConfiguration.PrivateNetwork)
		opts.Instance.Settings.IpConfiguration.RequireSsl = gcp.BoolValue(opts.Spec.Settings.IPConfiguration.RequireSsl)
		opts.Instance.Settings.IpConfiguration.ForceSendFields = []string{"Ipv4Enabled"}

		if len(opts.Spec.Settings.IPConfiguration.AuthorizedNetworks) > 0 {
			opts.Instance.Settings.IpConfiguration.AuthorizedNetworks = make([]*sqladmin.AclEntry, len(opts.Spec.Settings.IPConfiguration.AuthorizedNetworks))
		}
		for i, val := range opts.Spec.Settings.IPConfiguration.AuthorizedNetworks {
			opts.Instance.Settings.IpConfiguration.AuthorizedNetworks[i] = &sqladmin.AclEntry{
				ExpirationTime: gcp.StringValue(val.ExpirationTime),
				Name:           gcp.StringValue(val.Name),
				Value:          gcp.StringValue(val.Value),
				Kind:           "sql#aclEntry",
			}
		}
	}
	if opts.Spec.Settings.LocationPreference != nil {
		if opts.Instance.Settings.LocationPreference == nil {
			opts.Instance.Settings.LocationPreference = &sqladmin.LocationPreference{}
		}
		opts.Instance.Settings.LocationPreference.FollowGaeApplication = gcp.StringValue(opts.Spec.Settings.LocationPreference.FollowGaeApplication)
		opts.Instance.Settings.LocationPreference.Zone = gcp.StringValue(opts.Spec.Settings.LocationPreference.Zone)
	}
	if opts.Spec.Settings.MaintenanceWindow != nil {
		if opts.Instance.Settings.MaintenanceWindow == nil {
			opts.Instance.Settings.MaintenanceWindow = &sqladmin.MaintenanceWindow{}
		}
		opts.Instance.Settings.MaintenanceWindow.Day = gcp.Int64Value(opts.Spec.Settings.MaintenanceWindow.Day)
		opts.Instance.Settings.MaintenanceWindow.Hour = gcp.Int64Value(opts.Spec.Settings.MaintenanceWindow.Hour)
		opts.Instance.Settings.MaintenanceWindow.UpdateTrack = gcp.StringValue(opts.Spec.Settings.MaintenanceWindow.UpdateTrack)
	}
	if len(opts.Spec.Settings.DatabaseFlags) > 0 {
		opts.Instance.Settings.DatabaseFlags = make([]*sqladmin.DatabaseFlags, len(opts.Spec.Settings.DatabaseFlags))
	}
	for i, val := range opts.Spec.Settings.DatabaseFlags {
		opts.Instance.Settings.DatabaseFlags[i] = &sqladmin.DatabaseFlags{
			Name:  val.Name,
			Value: val.Value,
		}
	}

	if opts.Spec.ReplicaConfiguration != nil {
		if opts.Instance.ReplicaConfiguration == nil {
			opts.Instance.ReplicaConfiguration = &sqladmin.ReplicaConfiguration{
				FailoverTarget: gcp.BoolValue(opts.Spec.ReplicaConfiguration.FailoverTarget),
			}
		}
		if opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration != nil {
			if opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration == nil {
				opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration = &sqladmin.MySqlReplicaConfiguration{}
			}
			opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.DumpFilePath = gcp.StringValue(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.DumpFilePath)
			opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.ConnectRetryInterval = gcp.Int64Value(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.ConnectRetryInterval)
			opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.MasterHeartbeatPeriod = gcp.Int64Value(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.MasterHeartbeatPeriod)
			opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.VerifyServerCertificate = gcp.BoolValue(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.VerifyServerCertificate)
            return supplyReplicaConfigurationCredentials(opts.Instance, opts.Spec.ReplicaConfiguration, opts.Secret)

		}
	}
	return nil
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
func LateInitializeSpec(opts CloudSQLOptions) { // nolint:gocyclo

	// TODO(muvaf): One can marshall both objects into json and compare them as dictionaries since
	//  they both have the same key names but this may create performance problems as it'll happen in each
	//  reconcile. learn code-generation to make writing this easier and performant.
	if opts.Spec.Region == "" {
		opts.Spec.Region = opts.Instance.Region
	}
	opts.Spec.DatabaseVersion = gcp.LateInitializeString(opts.Spec.DatabaseVersion, opts.Instance.DatabaseVersion)
	opts.Spec.MasterInstanceName = gcp.LateInitializeString(opts.Spec.MasterInstanceName, opts.Instance.MasterInstanceName)
	opts.Spec.GceZone = gcp.LateInitializeString(opts.Spec.GceZone, opts.Instance.GceZone)
	opts.Spec.InstanceType = gcp.LateInitializeString(opts.Spec.InstanceType, opts.Instance.InstanceType)
	opts.Spec.MaxDiskSize = gcp.LateInitializeInt64(opts.Spec.MaxDiskSize, opts.Instance.MaxDiskSize)
	opts.Spec.ReplicaNames = gcp.LateInitializeStringSlice(opts.Spec.ReplicaNames, opts.Instance.ReplicaNames)
	opts.Spec.SuspensionReason = gcp.LateInitializeStringSlice(opts.Spec.SuspensionReason, opts.Instance.SuspensionReason)
	if opts.Instance.Settings != nil {
		if opts.Spec.Settings.Tier == "" {
			opts.Spec.Settings.Tier = opts.Instance.Settings.Tier
		}
		opts.Spec.Settings.ActivationPolicy = gcp.LateInitializeString(opts.Spec.Settings.ActivationPolicy, opts.Instance.Settings.ActivationPolicy)
		opts.Spec.Settings.AuthorizedGaeApplications = gcp.LateInitializeStringSlice(opts.Spec.Settings.AuthorizedGaeApplications, opts.Instance.Settings.AuthorizedGaeApplications)
		opts.Spec.Settings.AvailabilityType = gcp.LateInitializeString(opts.Spec.Settings.AvailabilityType, opts.Instance.Settings.AvailabilityType)
		opts.Spec.Settings.CrashSafeReplicationEnabled = gcp.LateInitializeBool(opts.Spec.Settings.CrashSafeReplicationEnabled, opts.Instance.Settings.CrashSafeReplicationEnabled)

		opts.Spec.Settings.DataDiskType = gcp.LateInitializeString(opts.Spec.Settings.DataDiskType, opts.Instance.Settings.DataDiskType)
		opts.Spec.Settings.PricingPlan = gcp.LateInitializeString(opts.Spec.Settings.PricingPlan, opts.Instance.Settings.PricingPlan)
		opts.Spec.Settings.ReplicationType = gcp.LateInitializeString(opts.Spec.Settings.ReplicationType, opts.Instance.Settings.ReplicationType)
		opts.Spec.Settings.UserLabels = gcp.LateInitializeStringMap(opts.Spec.Settings.UserLabels, opts.Instance.Settings.UserLabels)
		opts.Spec.Settings.DataDiskSizeGb = gcp.LateInitializeInt64(opts.Spec.Settings.DataDiskSizeGb, opts.Instance.Settings.DataDiskSizeGb)
		opts.Spec.Settings.DatabaseReplicationEnabled = gcp.LateInitializeBool(opts.Spec.Settings.DatabaseReplicationEnabled, opts.Instance.Settings.DatabaseReplicationEnabled)
		opts.Spec.Settings.StorageAutoResizeLimit = gcp.LateInitializeInt64(opts.Spec.Settings.StorageAutoResizeLimit, opts.Instance.Settings.StorageAutoResizeLimit)
		if opts.Spec.Settings.StorageAutoResize == nil {
			opts.Spec.Settings.StorageAutoResize = opts.Instance.Settings.StorageAutoResize
		}
		// If storage auto resize enabled, GCP does not allow setting a smaller
		// size but allows increasing it. Here, we set desired size as observed
		// if it is bigger than the current value which would allows us to get
		// in sync with the actual value but still allow us to increase it.
		if gcp.BoolValue(opts.Spec.Settings.StorageAutoResize) && gcp.Int64Value(opts.Spec.Settings.DataDiskSizeGb) < opts.Instance.Settings.DataDiskSizeGb {
			opts.Spec.Settings.DataDiskSizeGb = gcp.Int64Ptr(opts.Instance.Settings.DataDiskSizeGb)
		}
		if len(opts.Spec.Settings.DatabaseFlags) == 0 && len(opts.Instance.Settings.DatabaseFlags) != 0 {
			opts.Spec.Settings.DatabaseFlags = make([]*v1beta1.DatabaseFlags, len(opts.Instance.Settings.DatabaseFlags))
			for i, val := range opts.Instance.Settings.DatabaseFlags {
				opts.Spec.Settings.DatabaseFlags[i] = &v1beta1.DatabaseFlags{
					Name:  val.Name,
					Value: val.Value,
				}
			}
		}
		if opts.Instance.Settings.BackupConfiguration != nil {
			if opts.Spec.Settings.BackupConfiguration == nil {
				opts.Spec.Settings.BackupConfiguration = &v1beta1.BackupConfiguration{}
			}
			opts.Spec.Settings.BackupConfiguration.BinaryLogEnabled = gcp.LateInitializeBool(
				opts.Spec.Settings.BackupConfiguration.BinaryLogEnabled,
				opts.Instance.Settings.BackupConfiguration.BinaryLogEnabled)
			opts.Spec.Settings.BackupConfiguration.Enabled = gcp.LateInitializeBool(
				opts.Spec.Settings.BackupConfiguration.Enabled,
				opts.Instance.Settings.BackupConfiguration.Enabled)
			opts.Spec.Settings.BackupConfiguration.Location = gcp.LateInitializeString(
				opts.Spec.Settings.BackupConfiguration.Location,
				opts.Instance.Settings.BackupConfiguration.Location)
			opts.Spec.Settings.BackupConfiguration.ReplicationLogArchivingEnabled = gcp.LateInitializeBool(
				opts.Spec.Settings.BackupConfiguration.ReplicationLogArchivingEnabled,
				opts.Instance.Settings.BackupConfiguration.ReplicationLogArchivingEnabled)
			opts.Spec.Settings.BackupConfiguration.StartTime = gcp.LateInitializeString(
				opts.Spec.Settings.BackupConfiguration.StartTime,
				opts.Instance.Settings.BackupConfiguration.StartTime)
			opts.Spec.Settings.BackupConfiguration.PointInTimeRecoveryEnabled = gcp.LateInitializeBool(
				opts.Spec.Settings.BackupConfiguration.PointInTimeRecoveryEnabled,
				opts.Instance.Settings.BackupConfiguration.PointInTimeRecoveryEnabled)
		}
		if opts.Instance.Settings.IpConfiguration != nil {
			if opts.Spec.Settings.IPConfiguration == nil {
				opts.Spec.Settings.IPConfiguration = &v1beta1.IPConfiguration{}
			}
			opts.Spec.Settings.IPConfiguration.Ipv4Enabled = gcp.LateInitializeBool(opts.Spec.Settings.IPConfiguration.Ipv4Enabled, opts.Instance.Settings.IpConfiguration.Ipv4Enabled)
			opts.Spec.Settings.IPConfiguration.PrivateNetwork = gcp.LateInitializeString(opts.Spec.Settings.IPConfiguration.PrivateNetwork, opts.Instance.Settings.IpConfiguration.PrivateNetwork)
			opts.Spec.Settings.IPConfiguration.RequireSsl = gcp.LateInitializeBool(opts.Spec.Settings.IPConfiguration.RequireSsl, opts.Instance.Settings.IpConfiguration.RequireSsl)
			if len(opts.Instance.Settings.IpConfiguration.AuthorizedNetworks) != 0 && len(opts.Spec.Settings.IPConfiguration.AuthorizedNetworks) == 0 {
				opts.Spec.Settings.IPConfiguration.AuthorizedNetworks = make([]*v1beta1.ACLEntry, len(opts.Instance.Settings.IpConfiguration.AuthorizedNetworks))
				for i, val := range opts.Instance.Settings.IpConfiguration.AuthorizedNetworks {
					opts.Spec.Settings.IPConfiguration.AuthorizedNetworks[i] = &v1beta1.ACLEntry{
						ExpirationTime: &val.ExpirationTime,
						Name:           &val.Name,
						Value:          &val.Value,
					}
				}
			}
		}
		if opts.Instance.Settings.LocationPreference != nil {
			if opts.Spec.Settings.LocationPreference == nil {
				opts.Spec.Settings.LocationPreference = &v1beta1.LocationPreference{}
			}
			opts.Spec.Settings.LocationPreference.Zone = gcp.LateInitializeString(opts.Spec.Settings.LocationPreference.Zone, opts.Instance.Settings.LocationPreference.Zone)
			opts.Spec.Settings.LocationPreference.FollowGaeApplication = gcp.LateInitializeString(opts.Spec.Settings.LocationPreference.FollowGaeApplication, opts.Instance.Settings.LocationPreference.FollowGaeApplication)

		}
		if opts.Instance.Settings.MaintenanceWindow != nil {
			if opts.Spec.Settings.MaintenanceWindow == nil {
				opts.Spec.Settings.MaintenanceWindow = &v1beta1.MaintenanceWindow{}
			}
			opts.Spec.Settings.MaintenanceWindow.UpdateTrack = gcp.LateInitializeString(opts.Spec.Settings.MaintenanceWindow.UpdateTrack, opts.Instance.Settings.MaintenanceWindow.UpdateTrack)
			opts.Spec.Settings.MaintenanceWindow.Day = gcp.LateInitializeInt64(opts.Spec.Settings.MaintenanceWindow.Day, opts.Instance.Settings.MaintenanceWindow.Day)
			opts.Spec.Settings.MaintenanceWindow.Hour = gcp.LateInitializeInt64(opts.Spec.Settings.MaintenanceWindow.Hour, opts.Instance.Settings.MaintenanceWindow.Hour)
		}
	}
	if opts.Instance.DiskEncryptionConfiguration != nil {
		if opts.Spec.DiskEncryptionConfiguration == nil {
			opts.Spec.DiskEncryptionConfiguration = &v1beta1.DiskEncryptionConfiguration{}
		}
		if opts.Spec.DiskEncryptionConfiguration.KmsKeyName == "" {
			opts.Spec.DiskEncryptionConfiguration.KmsKeyName = opts.Instance.DiskEncryptionConfiguration.KmsKeyName
		}
	}
	if opts.Instance.FailoverReplica != nil {
		if opts.Spec.FailoverReplica == nil {
			opts.Spec.FailoverReplica = &v1beta1.DatabaseInstanceFailoverReplicaSpec{
				Name: opts.Instance.FailoverReplica.Name,
			}
		}
	}
	if opts.Instance.OnPremisesConfiguration != nil {
		if opts.Spec.OnPremisesConfiguration == nil {
			opts.Spec.OnPremisesConfiguration = &v1beta1.OnPremisesConfiguration{
				HostPort: opts.Instance.OnPremisesConfiguration.HostPort,
			}
		}
	}

	if opts.Instance.ReplicaConfiguration != nil {
		if opts.Spec.ReplicaConfiguration.FailoverTarget == nil {
			opts.Spec.ReplicaConfiguration = &v1beta1.ReplicaConfiguration{
				FailoverTarget: &opts.Instance.ReplicaConfiguration.FailoverTarget,
			}
		}
		if opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration != nil {
			if opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration == nil {
				opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration = &v1beta1.MySqlReplicaConfiguration{
					MasterHeartbeatPeriod: &opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.MasterHeartbeatPeriod,
				}
			}
			opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.DumpFilePath = gcp.LateInitializeString(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.DumpFilePath, opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.DumpFilePath)
			opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.ConnectRetryInterval = gcp.LateInitializeInt64(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.ConnectRetryInterval ,opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.ConnectRetryInterval)
			opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.MasterHeartbeatPeriod = gcp.LateInitializeInt64(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.MasterHeartbeatPeriod, opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.MasterHeartbeatPeriod)
			opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.VerifyServerCertificate = gcp.LateInitializeBool(opts.Spec.ReplicaConfiguration.MysqlReplicaConfiguration.VerifyServerCertificate, opts.Instance.ReplicaConfiguration.MysqlReplicaConfiguration.VerifyServerCertificate)
            //supplyReplicaConfigurationCredentials(opts.Instance, opts.Spec.ReplicaConfiguration, opts.Secret)
		}
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(opts CloudSQLOptions) (bool, error) {
	generated, err := copystructure.Copy(opts.Instance)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*sqladmin.DatabaseInstance)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	err = GenerateDatabaseInstance(opts)
	if err != nil {
		return false, err
	}

	return cmp.Equal(desired, opts.Instance, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(sqladmin.DatabaseInstance{}, "Settings.IpConfiguration.ForceSendFields", "ReplicaConfiguration.MysqlReplicaConfiguration")), nil
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


// supplyReplicaConfigurationCredentials Fetch Credentials from the corev1.Secret
func supplyReplicaConfigurationCredentials(in *sqladmin.DatabaseInstance, rc *v1beta1.ReplicaConfiguration, sc *corev1.Secret) error {

	caCert, err := extractValue(sc, rc.MysqlReplicaConfiguration.CaCertificateKey)
	if err != nil {
		return err
	}
	in.ReplicaConfiguration.MysqlReplicaConfiguration.CaCertificate = caCert

	clientCert, err := extractValue(sc, rc.MysqlReplicaConfiguration.ClientCertificateKey)
	if err != nil {
		return err
	}
	in.ReplicaConfiguration.MysqlReplicaConfiguration.ClientCertificate = clientCert

	clientKey, err := extractValue(sc, rc.MysqlReplicaConfiguration.ClientKey)
	if err != nil {
		return err
	}
	in.ReplicaConfiguration.MysqlReplicaConfiguration.ClientKey = clientKey

	password, err := extractValue(sc, rc.MysqlReplicaConfiguration.PasswordKey)
	if err != nil {
		return err
	}
	in.ReplicaConfiguration.MysqlReplicaConfiguration.Password = password

	username, err := extractValue(sc, rc.MysqlReplicaConfiguration.UsernameKey)
	if err != nil {
		return err
	}
	in.ReplicaConfiguration.MysqlReplicaConfiguration.Username = username

	return nil
}

// extractValue extract value from the given Secret
func extractValue(sc *corev1.Secret, key *string) (string, error) {

	if key != nil && sc != nil {
		if value, ok := sc.Data[*key]; !ok {
			return "", errors.Errorf(errSecretKey, *key) 
		} else {
			return string(value), nil
		}
	}
	return "", nil
}
