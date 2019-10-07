///*
//Copyright 2019 The Crossplane Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
package v1alpha2

//
//import (
//	"testing"
//
//	"github.com/google/go-cmp/cmp"
//	sqladmin "google.golang.org/api/sqladmin/v1beta4"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
//	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
//)
//
//func TestCloudsqlInstance_ConnectionSecret(t *testing.T) {
//	tests := map[string]struct {
//		fields *CloudsqlInstance
//		want   map[string][]byte
//	}{
//		"Default": {
//			fields: &CloudsqlInstance{
//				Spec: CloudsqlInstanceSpec{
//					CloudsqlInstanceParameters: CloudsqlInstanceParameters{
//						DatabaseVersion: "POSTGRES_9_6",
//					},
//				},
//			},
//			want: map[string][]byte{
//				PublicIPKey:  []byte(""),
//				PrivateIPKey: []byte(""),
//				runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(""),
//				runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(PostgresqlDefaultUser),
//			},
//		},
//	}
//	for name, tt := range tests {
//		t.Run(name, func(t *testing.T) {
//			if diff := cmp.Diff(tt.want, tt.fields.ConnectionSecret().Data); diff != "" {
//				t.Errorf("ConnectionSecret() -want, +got: %s", diff)
//			}
//		})
//	}
//}
//
//func TestCloudsqlInstance_DatabaseInstance(t *testing.T) {
//	type fields struct {
//		Spec CloudsqlInstanceSpec
//	}
//	type args struct {
//		name string
//	}
//	tests := map[string]struct {
//		fields fields
//		args   args
//		want   *sqladmin.DatabaseInstance
//	}{
//		"Default": {
//			fields: fields{Spec: CloudsqlInstanceSpec{}},
//			args:   args{name: "foo"},
//			want: &sqladmin.DatabaseInstance{
//				Name: "foo",
//				Settings: &sqladmin.Settings{
//					IpConfiguration: &sqladmin.IpConfiguration{
//						AuthorizedNetworks: []*sqladmin.AclEntry{},
//						ForceSendFields:    []string{"Ipv4Enabled"},
//					},
//				},
//			},
//		},
//		"WithSpecs": {
//			fields: fields{
//				Spec: CloudsqlInstanceSpec{
//					CloudsqlInstanceParameters: CloudsqlInstanceParameters{
//						AuthorizedNetworks: []string{"foo", "bar"},
//						DatabaseVersion:    "test-version",
//						Labels:             map[string]string{"fooz": "booz"},
//						Region:             "test-region",
//						StorageGB:          42,
//						StorageType:        "test-storage",
//						Tier:               "test-tier",
//					},
//				},
//			},
//			args: args{name: "test-name"},
//			want: &sqladmin.DatabaseInstance{
//				DatabaseVersion: "test-version",
//				Name:            "test-name",
//				Region:          "test-region",
//				Settings: &sqladmin.Settings{
//					DataDiskSizeGb: 42,
//					DataDiskType:   "test-storage",
//					IpConfiguration: &sqladmin.IpConfiguration{
//						AuthorizedNetworks: []*sqladmin.AclEntry{
//							{Value: "foo"},
//							{Value: "bar"},
//						},
//						ForceSendFields: []string{"Ipv4Enabled"},
//					},
//					Tier:       "test-tier",
//					UserLabels: map[string]string{"fooz": "booz"},
//				},
//			},
//		},
//	}
//	for name, tt := range tests {
//		t.Run(name, func(t *testing.T) {
//			c := &CloudsqlInstance{
//				Spec: tt.fields.Spec,
//			}
//			if diff := cmp.Diff(tt.want, c.DatabaseInstance(tt.args.name)); diff != "" {
//				t.Errorf("DatabaseInstance() -want, +got: %s", diff)
//			}
//		})
//	}
//}
//
//func TestCloudsqlInstance_DatabaseUserName(t *testing.T) {
//	tests := map[string]struct {
//		spec CloudsqlInstanceSpec
//		want string
//	}{
//		"Default": {
//			spec: CloudsqlInstanceSpec{},
//			want: MysqlDefaultUser,
//		},
//		"Postgres": {
//			spec: CloudsqlInstanceSpec{
//				CloudsqlInstanceParameters: CloudsqlInstanceParameters{
//					DatabaseVersion: "POSTGRES_9_6",
//				},
//			},
//			want: PostgresqlDefaultUser,
//		},
//		"MySQL": {
//			spec: CloudsqlInstanceSpec{
//				CloudsqlInstanceParameters: CloudsqlInstanceParameters{
//					DatabaseVersion: "MYSQL_5_7",
//				},
//			},
//			want: MysqlDefaultUser,
//		},
//	}
//	for name, tt := range tests {
//		t.Run(name, func(t *testing.T) {
//			c := &CloudsqlInstance{
//				Spec: tt.spec,
//			}
//			if got := c.DatabaseUserName(); got != tt.want {
//				t.Errorf("DatabaseUserName() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestCloudsqlInstance_GetResourceName(t *testing.T) {
//	om := metav1.ObjectMeta{
//		Namespace: "foo",
//		Name:      "bar",
//		UID:       "test-uid",
//	}
//	type fields struct {
//		meta metav1.ObjectMeta
//		spec CloudsqlInstanceSpec
//	}
//	tests := map[string]struct {
//		fields fields
//		want   string
//	}{
//		"NoNameFormat": {
//			fields: fields{
//				meta: om,
//				spec: CloudsqlInstanceSpec{},
//			},
//			want: "cloudsqlinstance-test-uid",
//		},
//		"FormatString": {
//			fields: fields{
//				meta: om,
//				spec: CloudsqlInstanceSpec{
//					CloudsqlInstanceParameters: CloudsqlInstanceParameters{
//						NameFormat: "foo-%s",
//					},
//				},
//			},
//			want: "foo-test-uid",
//		},
//		"ConstantString": {
//			fields: fields{
//				meta: om,
//				spec: CloudsqlInstanceSpec{
//					CloudsqlInstanceParameters: CloudsqlInstanceParameters{
//						NameFormat: "foo-bar",
//					},
//				},
//			},
//			want: "foo-bar",
//		},
//		"InvalidMultipleSubstitutions": {
//			fields: fields{
//				meta: om,
//				spec: CloudsqlInstanceSpec{
//					CloudsqlInstanceParameters: CloudsqlInstanceParameters{
//						NameFormat: "foo-%s-bar-%s",
//					},
//				},
//			},
//			want: "foo-test-uid-bar-%!s(MISSING)",
//		},
//	}
//	for name, tt := range tests {
//		t.Run(name, func(t *testing.T) {
//			b := &CloudsqlInstance{
//				ObjectMeta: tt.fields.meta,
//				Spec:       tt.fields.spec,
//			}
//			if got := b.GetResourceName(); got != tt.want {
//				t.Errorf("CloudsqlInstance.GetResourceName() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestCloudsqlInstance_IsRunnable(t *testing.T) {
//	tests := map[string]struct {
//		status CloudsqlInstanceStatus
//		want   bool
//	}{
//		"Default": {
//			status: CloudsqlInstanceStatus{},
//		},
//		"Runnable": {
//			status: CloudsqlInstanceStatus{
//				State: StateRunnable,
//			},
//			want: true,
//		},
//		"NotRunnable": {
//			status: CloudsqlInstanceStatus{
//				State: "something-else",
//			},
//		},
//	}
//	for name, tt := range tests {
//		t.Run(name, func(t *testing.T) {
//			c := &CloudsqlInstance{
//				Status: tt.status,
//			}
//			if got := c.IsRunnable(); got != tt.want {
//				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestCloudsqlInstance_SetStatus(t *testing.T) {
//	tests := map[string]struct {
//		status CloudsqlInstanceStatus
//		args   *sqladmin.DatabaseInstance
//		want   CloudsqlInstanceStatus
//	}{
//		"Nil": {
//			status: CloudsqlInstanceStatus{},
//			args:   nil,
//			want:   CloudsqlInstanceStatus{},
//		},
//		"Default": {
//			status: CloudsqlInstanceStatus{},
//			args:   &sqladmin.DatabaseInstance{},
//			want: CloudsqlInstanceStatus{
//				ResourceStatus: runtimev1alpha1.ResourceStatus{
//					ConditionedStatus: runtimev1alpha1.ConditionedStatus{
//						Conditions: []runtimev1alpha1.Condition{
//							{
//								Type:   runtimev1alpha1.TypeReady,
//								Status: "False",
//								Reason: "Managed resource is not available for use",
//							},
//						},
//					},
//				},
//			},
//		},
//		"Available": {
//			status: CloudsqlInstanceStatus{},
//			args: &sqladmin.DatabaseInstance{
//				IpAddresses: []*sqladmin.IpMapping{
//					{
//						IpAddress: "foo",
//						Type:      PublicIPType,
//					},
//				},
//				State: StateRunnable,
//			},
//			want: CloudsqlInstanceStatus{
//				ResourceStatus: runtimev1alpha1.ResourceStatus{
//					ConditionedStatus: runtimev1alpha1.ConditionedStatus{
//						Conditions: []runtimev1alpha1.Condition{
//							{
//								Type:   runtimev1alpha1.TypeReady,
//								Status: "True",
//								Reason: "Managed resource is available for use",
//							},
//						},
//					},
//					BindingStatus: runtimev1alpha1.BindingStatus{
//						Phase: runtimev1alpha1.BindingPhaseUnbound,
//					},
//				},
//				PublicIP: "foo",
//				State:    StateRunnable,
//			},
//		},
//	}
//	for name, tt := range tests {
//		t.Run(name, func(t *testing.T) {
//			c := &CloudsqlInstance{
//				Status: tt.status,
//			}
//			c.SetStatus(tt.args)
//			if diff := cmp.Diff(tt.want, c.Status); diff != "" {
//				t.Errorf("SetStatus() -want, +got: %s", diff)
//			}
//		})
//	}
//}
