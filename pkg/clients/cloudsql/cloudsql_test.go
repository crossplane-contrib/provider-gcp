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
	"testing"

	"github.com/google/go-cmp/cmp"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	corev1 "k8s.io/api/core/v1"

	"github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

const (
	name = "test-sql"
)

func params(m ...func(*v1beta1.CloudSQLInstanceParameters)) *v1beta1.CloudSQLInstanceParameters {
	p := &v1beta1.CloudSQLInstanceParameters{
		Region: "us-west2",
		Settings: v1beta1.Settings{
			Tier:                        "best-one-available",
			ActivationPolicy:            gcp.StringPtr("always"),
			AuthorizedGaeApplications:   []string{"my-gapp"},
			AvailabilityType:            gcp.StringPtr("time-to-time"),
			CrashSafeReplicationEnabled: gcp.BoolPtr(true),
			StorageAutoResize:           gcp.BoolPtr(false),
			DataDiskType:                gcp.StringPtr("PD_SSD"),
			PricingPlan:                 gcp.StringPtr("PER_USE"),
			ReplicationType:             gcp.StringPtr("SYNCHRONOUS"),
			UserLabels: map[string]string{
				"importance": "high",
			},
			DatabaseFlags: []*v1beta1.DatabaseFlags{
				{
					Name:  "run",
					Value: "forest",
				},
			},
			BackupConfiguration: &v1beta1.BackupConfiguration{
				BinaryLogEnabled:               gcp.BoolPtr(true),
				Enabled:                        gcp.BoolPtr(false),
				Location:                       gcp.StringPtr("us-west1"),
				ReplicationLogArchivingEnabled: gcp.BoolPtr(true),
				StartTime:                      gcp.StringPtr("20191018"),
			},
			IPConfiguration: &v1beta1.IPConfiguration{
				AuthorizedNetworks: []*v1beta1.ACLEntry{
					{
						ExpirationTime: gcp.StringPtr("20201018"),
						Name:           gcp.StringPtr("hate"),
						Value:          gcp.StringPtr("unittests"),
					},
				},
			},
			LocationPreference: &v1beta1.LocationPreference{
				FollowGaeApplication: gcp.StringPtr("my-gapp"),
				Zone:                 gcp.StringPtr("us-west1-a"),
			},
			MaintenanceWindow: &v1beta1.MaintenanceWindow{
				Day:         gcp.Int64Ptr(1),
				Hour:        gcp.Int64Ptr(2),
				UpdateTrack: gcp.StringPtr("canary"),
			},
			DataDiskSizeGb:             gcp.Int64Ptr(2),
			DatabaseReplicationEnabled: gcp.BoolPtr(true),
			StorageAutoResizeLimit:     gcp.Int64Ptr(3),
		},
		DatabaseVersion:    gcp.StringPtr("3.2"),
		MasterInstanceName: gcp.StringPtr("myFunnyMaster"),
		DiskEncryptionConfiguration: &v1beta1.DiskEncryptionConfiguration{
			KmsKeyName: "my-key",
		},
		FailoverReplica: &v1beta1.DatabaseInstanceFailoverReplicaSpec{
			Name: "my-failover",
		},
		GceZone:      gcp.StringPtr("us-west2"),
		InstanceType: gcp.StringPtr("db-standard-1"),
		MaxDiskSize:  gcp.Int64Ptr(3000000000),
		OnPremisesConfiguration: &v1beta1.OnPremisesConfiguration{
			HostPort: "3306",
		},
		ReplicaNames:     []string{"my-replica1", "and2"},
		SuspensionReason: []string{"gotta play nice with others", "or go"},
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1beta1.CloudSQLInstanceObservation)) *v1beta1.CloudSQLInstanceObservation {
	o := &v1beta1.CloudSQLInstanceObservation{
		BackendType:     "SECOND_GEN",
		CurrentDiskSize: 2000000,
		ConnectionName:  "special-conn",
		DiskEncryptionStatus: &v1beta1.DiskEncryptionStatus{
			KmsKeyVersionName: "v1.0",
		},
		IPAddresses: []*v1beta1.IPMapping{
			{
				IPAddress:    "20.0.0.1",
				TimeToRetire: "2012-11-15T16:19:00.094Z",
				Type:         "PRIVATE",
			},
		},
		FailoverReplica: &v1beta1.DatabaseInstanceFailoverReplicaStatus{
			Available: true,
		},
		IPv6Address:                "2.19sd920.2",
		Project:                    "crossplane-eats-the-cloud",
		ServiceAccountEmailAddress: "john@dontparseme.com",
		GceZone:                    "us-west2",
		State:                      "RUNNABLE",
		SettingsVersion:            23142,
		SelfLink:                   "/projects/crossplane-eats-the-cloud/database/test-sql",
	}
	for _, f := range m {
		f(o)
	}
	return o
}

func db(m ...func(*sqladmin.DatabaseInstance)) *sqladmin.DatabaseInstance {
	db := &sqladmin.DatabaseInstance{
		Name:   "test-sql",
		Region: "us-west2",
		Settings: &sqladmin.Settings{
			Tier:                        "best-one-available",
			ActivationPolicy:            "always",
			AuthorizedGaeApplications:   []string{"my-gapp"},
			AvailabilityType:            "time-to-time",
			CrashSafeReplicationEnabled: true,
			StorageAutoResize:           gcp.BoolPtr(false),
			DataDiskType:                "PD_SSD",
			PricingPlan:                 "PER_USE",
			ReplicationType:             "SYNCHRONOUS",
			UserLabels: map[string]string{
				"importance": "high",
			},
			DatabaseFlags: []*sqladmin.DatabaseFlags{
				{
					Name:  "run",
					Value: "forest",
				},
			},
			BackupConfiguration: &sqladmin.BackupConfiguration{
				BinaryLogEnabled:               true,
				Enabled:                        false,
				Location:                       "us-west1",
				ReplicationLogArchivingEnabled: true,
				StartTime:                      "20191018",
			},
			IpConfiguration: &sqladmin.IpConfiguration{
				AuthorizedNetworks: []*sqladmin.AclEntry{
					{
						ExpirationTime: "20201018",
						Name:           "hate",
						Value:          "unittests",
						Kind:           "sql#aclEntry",
					},
				},
				ForceSendFields: []string{"Ipv4Enabled"},
			},
			LocationPreference: &sqladmin.LocationPreference{
				FollowGaeApplication: "my-gapp",
				Zone:                 "us-west1-a",
			},
			MaintenanceWindow: &sqladmin.MaintenanceWindow{
				Day:         1,
				Hour:        2,
				UpdateTrack: "canary",
			},
			DataDiskSizeGb:             2,
			DatabaseReplicationEnabled: true,
			StorageAutoResizeLimit:     3,
		},
		DatabaseVersion:    "3.2",
		MasterInstanceName: "myFunnyMaster",
		DiskEncryptionConfiguration: &sqladmin.DiskEncryptionConfiguration{
			KmsKeyName: "my-key",
		},
		FailoverReplica: &sqladmin.DatabaseInstanceFailoverReplica{
			Name: "my-failover",
		},
		GceZone:      "us-west2",
		InstanceType: "db-standard-1",
		MaxDiskSize:  int64(3000000000),
		OnPremisesConfiguration: &sqladmin.OnPremisesConfiguration{
			HostPort: "3306",
		},
		ReplicaNames:     []string{"my-replica1", "and2"},
		SuspensionReason: []string{"gotta play nice with others", "or go"},
	}
	for _, f := range m {
		f(db)
	}
	return db
}

func addOutputFields(db *sqladmin.DatabaseInstance) {
	db.BackendType = "SECOND_GEN"
	db.CurrentDiskSize = 2000000
	db.ConnectionName = "special-conn"
	db.DiskEncryptionStatus = &sqladmin.DiskEncryptionStatus{
		KmsKeyVersionName: "v1.0",
	}
	db.FailoverReplica = &sqladmin.DatabaseInstanceFailoverReplica{
		Name:      "my-failover",
		Available: true,
	}
	db.GceZone = "us-west2"
	db.IpAddresses = []*sqladmin.IpMapping{
		{
			IpAddress:    "20.0.0.1",
			TimeToRetire: "2012-11-15T16:19:00.094Z",
			Type:         "PRIVATE",
		},
	}
	db.Ipv6Address = "2.19sd920.2"
	db.Project = "crossplane-eats-the-cloud"
	db.SelfLink = "/projects/crossplane-eats-the-cloud/database/test-sql"
	db.ServiceAccountEmailAddress = "john@dontparseme.com"
	db.State = "RUNNABLE"
	db.Settings.SettingsVersion = 23142
}

func TestGenerateDatabaseInstance(t *testing.T) {
	type args struct {
		name   string
		params v1beta1.CloudSQLInstanceParameters
	}
	type want struct {
		db *sqladmin.DatabaseInstance
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{name: name, params: *params()},
			want: want{db: db()},
		},
		"MissingFields": {
			args: args{
				name: name,
				params: *params(func(p *v1beta1.CloudSQLInstanceParameters) {
					p.MasterInstanceName = nil
					p.GceZone = nil
				})},
			want: want{db: db(func(db *sqladmin.DatabaseInstance) {
				db.MasterInstanceName = ""
				db.GceZone = ""
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &sqladmin.DatabaseInstance{}
			GenerateDatabaseInstance(tc.args.name, tc.args.params, r)
			if diff := cmp.Diff(tc.want.db, r); diff != "" {
				t.Errorf("GenerateDatabaseInstance(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		db     *sqladmin.DatabaseInstance
		params *v1beta1.CloudSQLInstanceParameters
	}
	type want struct {
		params *v1beta1.CloudSQLInstanceParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1beta1.CloudSQLInstanceParameters) {
					p.GceZone = nil
				}),
				db: db(func(db *sqladmin.DatabaseInstance) {
					db.GceZone = "us-different-2"
					db.MasterInstanceName = "not-what-you-expect"
				}),
			},
			want: want{params: params(func(p *v1beta1.CloudSQLInstanceParameters) {
				p.GceZone = gcp.StringPtr("us-different-2")
			})},
		},
		"AllFilledAlready": {
			args: args{
				params: params(),
				db:     db(),
			},
			want: want{
				params: params(),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.params, *tc.args.db)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	type args struct {
		db *sqladmin.DatabaseInstance
	}
	type want struct {
		obs v1beta1.CloudSQLInstanceObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				db(addOutputFields),
			},
			want: want{*observation()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o := GenerateObservation(*tc.args.db)
			if diff := cmp.Diff(tc.want.obs, o); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDatabaseUserName(t *testing.T) {
	p := v1beta1.CloudSQLInstanceParameters{
		DatabaseVersion: gcp.StringPtr("POSTGRES_3.2"),
	}
	if diff := cmp.Diff(v1beta1.PostgresqlDefaultUser, DatabaseUserName(p)); diff != "" {
		t.Errorf("DatabaseUserName(...): -want, +got:\n%s", diff)
	}
	p.DatabaseVersion = gcp.StringPtr("3.2")
	if diff := cmp.Diff(v1beta1.MysqlDefaultUser, DatabaseUserName(p)); diff != "" {
		t.Errorf("DatabaseUserName(...): -want, +got:\n%s", diff)
	}
}

func TestGetServerCACertificate(t *testing.T) {
	cert := &sqladmin.SslCert{
		Cert:             "my-cert",
		CertSerialNumber: "23412342124",
		CommonName:       "my-common-name",
		CreateTime:       "2012-11-15T16:19:00.094Z",
		ExpirationTime:   "2013-11-15T16:19:00.094Z",
		Instance:         name,
		Kind:             "sql#sslCert",
		SelfLink:         "/projects/crossplane-eats-the-cloud/certificates/my-cert",
		Sha1Fingerprint:  "some-sha1",
	}

	type args struct {
		db sqladmin.DatabaseInstance
	}
	type want struct {
		r map[string][]byte
	}

	cases := map[string]struct {
		args
		want
	}{
		"NilCert": {
			args: args{db: sqladmin.DatabaseInstance{}},
			want: want{},
		},
		"FullCert": {
			args: args{db: *db(func(db *sqladmin.DatabaseInstance) {
				db.ServerCaCert = cert
			})},
			want: want{r: map[string][]byte{
				v1beta1.CloudSQLSecretServerCACertificateCertKey:             []byte(cert.Cert),
				v1beta1.CloudSQLSecretServerCACertificateCertSerialNumberKey: []byte(cert.CertSerialNumber),
				v1beta1.CloudSQLSecretServerCACertificateCommonNameKey:       []byte(cert.CommonName),
				v1beta1.CloudSQLSecretServerCACertificateCreateTimeKey:       []byte(cert.CreateTime),
				v1beta1.CloudSQLSecretServerCACertificateExpirationTimeKey:   []byte(cert.ExpirationTime),
				v1beta1.CloudSQLSecretServerCACertificateInstanceKey:         []byte(cert.Instance),
				v1beta1.CloudSQLSecretServerCACertificateSha1FingerprintKey:  []byte(cert.Sha1Fingerprint),
			},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			m := GetServerCACertificate(tc.args.db)
			if diff := cmp.Diff(tc.want.r, m); diff != "" {
				t.Errorf("GetServerCACertificate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	privateNetworkName := "a-cool-network"

	type args struct {
		params *v1beta1.CloudSQLInstanceParameters
		db     *sqladmin.DatabaseInstance
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"IsUpToDate": {
			args: args{
				params: params(),
				db:     db(),
			},
			want: true,
		},
		"IsUpToDateWithOutputFields": {
			args: args{
				params: params(),
				db:     db(addOutputFields),
			},
			want: true,
		},
		"IsUpToDateIgnoreReferences": {
			args: args{
				params: params(func(p *v1beta1.CloudSQLInstanceParameters) {
					p.Settings.IPConfiguration = &v1beta1.IPConfiguration{
						PrivateNetwork: &privateNetworkName,
						PrivateNetworkRef: &v1beta1.NetworkURIReferencerForCloudSQLInstance{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "network-ref-exists",
							},
						},
					}
				}),
				db: db(func(db *sqladmin.DatabaseInstance) {
					db.Settings.IpConfiguration = &sqladmin.IpConfiguration{
						PrivateNetwork: privateNetworkName,
					}
				}),
			},
			want: true,
		},
		"NeedsUpdate": {
			args: args{
				params: params(),
				db: db(func(db *sqladmin.DatabaseInstance) {
					db.MasterInstanceName = ""
				}),
			},
			want: false,
		},
		"NeedsUpdateBadRef": {
			args: args{
				params: params(func(p *v1beta1.CloudSQLInstanceParameters) {
					p.Settings.IPConfiguration = &v1beta1.IPConfiguration{
						PrivateNetwork: &privateNetworkName,
						PrivateNetworkRef: &v1beta1.NetworkURIReferencerForCloudSQLInstance{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "network-ref-exists",
							},
						},
					}
				}),
				db: db(func(db *sqladmin.DatabaseInstance) {
					db.Settings.IpConfiguration = &sqladmin.IpConfiguration{
						PrivateNetwork: "unexpected-network",
					}
				}),
			},
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, _ := IsUpToDate("test-sql", tc.args.params, tc.args.db)
			if diff := cmp.Diff(tc.want, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
