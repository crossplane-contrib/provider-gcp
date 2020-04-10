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

package cluster

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1beta1"
	corev1 "k8s.io/api/core/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	computev1beta1 "github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	"github.com/crossplane/provider-gcp/apis/container/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const (
	name     = "my-cool-cluster"
	location = "cool-location"
	project  = "cool-project"
)

var (
	resourceLabels = map[string]string{"label": "one"}
)

func cluster(m ...func(*container.Cluster)) *container.Cluster {
	c := &container.Cluster{
		ClusterIpv4Cidr:       "0.0.0.0/0",
		Description:           "my cool description",
		EnableKubernetesAlpha: true,
		EnableTpu:             true,
		InitialClusterVersion: "1.16",
		LabelFingerprint:      "fingerprint",
		Locations:             []string{"us-central1-a", "us-central1-b"},
		LoggingService:        "logging.googleapis.com",
		MonitoringService:     "monitoring.googleapis.com",
		Name:                  name,
		Network:               "default",
		ResourceLabels:        resourceLabels,
		Subnetwork:            "default",
	}
	for _, f := range m {
		f(c)
	}

	return c
}

func params(m ...func(*v1beta1.GKEClusterParameters)) *v1beta1.GKEClusterParameters {
	p := &v1beta1.GKEClusterParameters{
		ClusterIpv4Cidr:       gcp.StringPtr("0.0.0.0/0"),
		Description:           gcp.StringPtr("my cool description"),
		EnableKubernetesAlpha: gcp.BoolPtr(true),
		EnableTpu:             gcp.BoolPtr(true),
		InitialClusterVersion: gcp.StringPtr("1.16"),
		LabelFingerprint:      gcp.StringPtr("fingerprint"),
		Locations:             []string{"us-central1-a", "us-central1-b"},
		LoggingService:        gcp.StringPtr("logging.googleapis.com"),
		MonitoringService:     gcp.StringPtr("monitoring.googleapis.com"),
		Location:              location,
		Network:               gcp.StringPtr("default"),
		ResourceLabels:        resourceLabels,
		Subnetwork:            gcp.StringPtr("default"),
	}
	for _, f := range m {
		f(p)
	}

	return p
}

func observation(m ...func(*v1beta1.GKEClusterObservation)) *v1beta1.GKEClusterObservation {
	o := &v1beta1.GKEClusterObservation{
		CreateTime: "13:13",
		Conditions: []*v1beta1.StatusCondition{
			{
				Code:    "UNKNOWN",
				Message: "Condition is unknown.",
			},
		},
		CurrentMasterVersion: "1.16",
		CurrentNodeCount:     5,
		CurrentNodeVersion:   "1.16",
		Endpoint:             "12.12.12.12",
		ExpireTime:           "13:13",
		Location:             "us-central1",
		NodeIpv4CidrSize:     8,
		SelfLink:             "/link/to/myself",
		ServicesIpv4Cidr:     "0.0.0.0/0",
		Status:               "RUNNING",
		StatusMessage:        "I am running.",
		TpuIpv4CidrBlock:     "0.0.0.0/0",
		Zone:                 "us-central1-a",

		MaintenancePolicy: &v1beta1.MaintenancePolicyStatus{
			Window: v1beta1.MaintenanceWindowStatus{
				DailyMaintenanceWindow: v1beta1.DailyMaintenanceWindowStatus{
					Duration: "1h",
				},
			},
		},

		NetworkConfig: &v1beta1.NetworkConfigStatus{
			Network:    "my-cool-network",
			Subnetwork: "my-cool-subnetwork",
		},

		PrivateClusterConfig: &v1beta1.PrivateClusterConfigStatus{
			PrivateEndpoint: "12.12.12.12",
			PublicEndpoint:  "12.12.12.12",
		},
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func addOutputFields(c *container.Cluster) {
	c.CreateTime = "13:13"
	c.Conditions = []*container.StatusCondition{
		{
			Code:    "UNKNOWN",
			Message: "Condition is unknown.",
		},
	}
	c.CurrentMasterVersion = "1.16"
	c.CurrentNodeCount = 5
	c.CurrentNodeVersion = "1.16"
	c.Endpoint = "12.12.12.12"
	c.ExpireTime = "13:13"
	c.Location = "us-central1"
	c.NodeIpv4CidrSize = 8
	c.SelfLink = "/link/to/myself"
	c.ServicesIpv4Cidr = "0.0.0.0/0"
	c.Status = "RUNNING"
	c.StatusMessage = "I am running."
	c.TpuIpv4CidrBlock = "0.0.0.0/0"
	c.Zone = "us-central1-a"

	c.MaintenancePolicy = &container.MaintenancePolicy{
		Window: &container.MaintenanceWindow{
			DailyMaintenanceWindow: &container.DailyMaintenanceWindow{
				Duration: "1h",
			},
		},
	}

	c.NetworkConfig = &container.NetworkConfig{
		Network:    "my-cool-network",
		Subnetwork: "my-cool-subnetwork",
	}

	c.PrivateClusterConfig = &container.PrivateClusterConfig{
		PrivateEndpoint: "12.12.12.12",
		PublicEndpoint:  "12.12.12.12",
	}
}

func TestGenerateObservation(t *testing.T) {
	type args struct {
		cluster *container.Cluster
	}

	tests := map[string]struct {
		args args
		want *v1beta1.GKEClusterObservation
	}{
		"Successful": {
			args: args{
				cluster: cluster(addOutputFields),
			},
			want: observation(),
		},
		"SuccessfulWithNodePool": {
			args: args{
				cluster(addOutputFields, func(c *container.Cluster) {
					sc := &container.StatusCondition{
						Code:    "cool-code",
						Message: "cool-message",
					}
					ac := &container.AcceleratorConfig{
						AcceleratorCount: 5,
					}
					np := &container.NodePool{
						Conditions: []*container.StatusCondition{sc},
						Config: &container.NodeConfig{
							Accelerators: []*container.AcceleratorConfig{ac},
						},
						Name: "cool-node-pool",
					}
					c.NodePools = []*container.NodePool{np}
				}),
			},
			want: observation(func(p *v1beta1.GKEClusterObservation) {
				sc := &v1beta1.StatusCondition{
					Code:    "cool-code",
					Message: "cool-message",
				}
				ac := &v1beta1.AcceleratorConfigClusterStatus{
					AcceleratorCount: 5,
				}
				np := &v1beta1.NodePoolClusterStatus{
					Conditions: []*v1beta1.StatusCondition{sc},
					Config: &v1beta1.NodeConfigClusterStatus{
						Accelerators: []*v1beta1.AcceleratorConfigClusterStatus{ac},
					},
					Name: "cool-node-pool",
				}
				p.NodePools = []*v1beta1.NodePoolClusterStatus{np}
			}),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			observation := GenerateObservation(*tc.args.cluster)
			if diff := cmp.Diff(*tc.want, observation); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCluster(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
		name    string
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.AddonsConfig = &v1beta1.AddonsConfig{
						HorizontalPodAutoscaling: &v1beta1.HorizontalPodAutoscaling{
							Disabled: gcp.BoolPtr(true),
						},
					}
					p.DatabaseEncryption = &v1beta1.DatabaseEncryption{
						KeyName: gcp.StringPtr("cool-key"),
						State:   gcp.StringPtr("UNKNOWN"),
					}
				}),
				name: name,
			},
			want: cluster(func(c *container.Cluster) {
				c.AddonsConfig = &container.AddonsConfig{
					HorizontalPodAutoscaling: &container.HorizontalPodAutoscaling{
						Disabled:        true,
						ForceSendFields: []string{"Disabled"},
					},
				}
				c.DatabaseEncryption = &container.DatabaseEncryption{
					KeyName: "cool-key",
					State:   "UNKNOWN",
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
				name:    name,
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateCluster(tc.args.name, *tc.args.params, tc.args.cluster)
			if diff := cmp.Diff(tc.args.cluster, tc.want); diff != "" {
				t.Errorf("GenerateCluster(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAddNodePoolForCreate(t *testing.T) {
	pool := &container.NodePool{
		Name:             BootstrapNodePoolName,
		InitialNodeCount: 0,
	}
	tests := map[string]struct {
		args *container.Cluster
		want *container.Cluster
	}{
		"Successful": {
			args: cluster(),
			want: cluster(func(c *container.Cluster) {
				c.NodePools = []*container.NodePool{pool}
			}),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			AddNodePoolForCreate(tc.args)
			if diff := cmp.Diff(tc.want, tc.args); diff != "" {
				t.Errorf("AddNodePoolForCreate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAddonsConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: &container.Cluster{},
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.AddonsConfig = &v1beta1.AddonsConfig{
						HorizontalPodAutoscaling: &v1beta1.HorizontalPodAutoscaling{
							Disabled: gcp.BoolPtr(true),
						},
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.AddonsConfig = &container.AddonsConfig{
					HorizontalPodAutoscaling: &container.HorizontalPodAutoscaling{
						Disabled:        true,
						ForceSendFields: []string{"Disabled"},
					},
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: &container.Cluster{},
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateAddonsConfig(tc.args.params.AddonsConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.AddonsConfig, tc.args.cluster.AddonsConfig); diff != "" {
				t.Errorf("GenerateAddonsConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAuthenticatorGroupsConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.AuthenticatorGroupsConfig = &v1beta1.AuthenticatorGroupsConfig{
						Enabled:       gcp.BoolPtr(true),
						SecurityGroup: gcp.StringPtr("my-group"),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.AuthenticatorGroupsConfig = &container.AuthenticatorGroupsConfig{
					Enabled:       true,
					SecurityGroup: "my-group",
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateAuthenticatorGroupsConfig(tc.args.params.AuthenticatorGroupsConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.AuthenticatorGroupsConfig, tc.args.cluster.AuthenticatorGroupsConfig); diff != "" {
				t.Errorf("GenerateAuthenticatorGroupsConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAutoscaling(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.Autoscaling = &v1beta1.ClusterAutoscaling{
						AutoprovisioningLocations:  []string{"here", "there"},
						EnableNodeAutoprovisioning: gcp.BoolPtr(true),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.Autoscaling = &container.ClusterAutoscaling{
					AutoprovisioningLocations:  []string{"here", "there"},
					EnableNodeAutoprovisioning: true,
				}
			}),
		},
		"SuccessfulWithResourceLimits": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.Autoscaling = &v1beta1.ClusterAutoscaling{
						AutoprovisioningLocations:  []string{"here", "there"},
						EnableNodeAutoprovisioning: gcp.BoolPtr(true),
						ResourceLimits: []*v1beta1.ResourceLimit{
							{
								Maximum:      gcp.Int64Ptr(20),
								ResourceType: gcp.StringPtr("cpu"),
							},
						},
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.Autoscaling = &container.ClusterAutoscaling{
					AutoprovisioningLocations:  []string{"here", "there"},
					EnableNodeAutoprovisioning: true,
					ResourceLimits: []*container.ResourceLimit{
						{
							Maximum:      20,
							Minimum:      0,
							ResourceType: "cpu",
						},
					},
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateAutoscaling(tc.args.params.Autoscaling, tc.args.cluster)
			if diff := cmp.Diff(tc.want.Autoscaling, tc.args.cluster.Autoscaling); diff != "" {
				t.Errorf("GenerateAutoscaling(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateBinaryAuthorization(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.BinaryAuthorization = &v1beta1.BinaryAuthorization{
						Enabled: true,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.BinaryAuthorization = &container.BinaryAuthorization{
					Enabled: true,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateBinaryAuthorization(tc.args.params.BinaryAuthorization, tc.args.cluster)
			if diff := cmp.Diff(tc.want.BinaryAuthorization, tc.args.cluster.BinaryAuthorization); diff != "" {
				t.Errorf("GenerateBinaryAuthorization(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateDatabaseEncryption(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.DatabaseEncryption = &v1beta1.DatabaseEncryption{
						KeyName: gcp.StringPtr("cool-key"),
						State:   gcp.StringPtr("UNKNOWN"),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.DatabaseEncryption = &container.DatabaseEncryption{
					KeyName: "cool-key",
					State:   "UNKNOWN",
				}
			}),
		},
		"SuccessfulPartial": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.DatabaseEncryption = &v1beta1.DatabaseEncryption{
						KeyName: gcp.StringPtr("cool-key"),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.DatabaseEncryption = &container.DatabaseEncryption{
					KeyName: "cool-key",
					State:   "",
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateDatabaseEncryption(tc.args.params.DatabaseEncryption, tc.args.cluster)
			if diff := cmp.Diff(tc.want.DatabaseEncryption, tc.args.cluster.DatabaseEncryption); diff != "" {
				t.Errorf("GenerateDatabaseEncryption(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateDefaultMaxPodsConstraint(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.DefaultMaxPodsConstraint = &v1beta1.MaxPodsConstraint{
						MaxPodsPerNode: 5,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.DefaultMaxPodsConstraint = &container.MaxPodsConstraint{
					MaxPodsPerNode: 5,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateDefaultMaxPodsConstraint(tc.args.params.DefaultMaxPodsConstraint, tc.args.cluster)
			if diff := cmp.Diff(tc.want.DefaultMaxPodsConstraint, tc.args.cluster.DefaultMaxPodsConstraint); diff != "" {
				t.Errorf("GenerateDefaultMaxPodsConstraint(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIpAllocationPolicy(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.IPAllocationPolicy = &v1beta1.IPAllocationPolicy{
						AllowRouteOverlap:    gcp.BoolPtr(true),
						ClusterIpv4CidrBlock: gcp.StringPtr("0.0.0.0/0"),
						UseIPAliases:         gcp.BoolPtr(true),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.IpAllocationPolicy = &container.IPAllocationPolicy{
					AllowRouteOverlap:    true,
					ClusterIpv4CidrBlock: "0.0.0.0/0",
					UseIpAliases:         true,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateIPAllocationPolicy(tc.args.params.IPAllocationPolicy, tc.args.cluster)
			if diff := cmp.Diff(tc.want.IpAllocationPolicy, tc.args.cluster.IpAllocationPolicy); diff != "" {
				t.Errorf("GenerateIpAllocationPolicy(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateLegacyAbac(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.LegacyAbac = &v1beta1.LegacyAbac{
						Enabled: true,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.LegacyAbac = &container.LegacyAbac{
					Enabled: true,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateLegacyAbac(tc.args.params.LegacyAbac, tc.args.cluster)
			if diff := cmp.Diff(tc.want.LegacyAbac, tc.args.cluster.LegacyAbac); diff != "" {
				t.Errorf("GenerateLegacyAbac(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateMaintenancePolicy(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.MaintenancePolicy = &v1beta1.MaintenancePolicySpec{
						Window: v1beta1.MaintenanceWindowSpec{
							DailyMaintenanceWindow: v1beta1.DailyMaintenanceWindowSpec{
								StartTime: "13:13",
							},
						},
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.MaintenancePolicy = &container.MaintenancePolicy{
					Window: &container.MaintenanceWindow{
						DailyMaintenanceWindow: &container.DailyMaintenanceWindow{
							StartTime: "13:13",
						},
					},
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateMaintenancePolicy(tc.args.params.MaintenancePolicy, tc.args.cluster)
			if diff := cmp.Diff(tc.want.MaintenancePolicy, tc.args.cluster.MaintenancePolicy); diff != "" {
				t.Errorf("GenerateMaintenancePolicy(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateMasterAuth(t *testing.T) {
	var adminUser = "admin"

	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.MasterAuth = &v1beta1.MasterAuth{
						ClientCertificateConfig: &v1beta1.ClientCertificateConfig{
							IssueClientCertificate: true,
						},
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.MasterAuth = &container.MasterAuth{
					ClientCertificateConfig: &container.ClientCertificateConfig{
						IssueClientCertificate: true,
					},
				}
			}),
		},
		"SuccessfulFalseWithUsername": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.MasterAuth = &v1beta1.MasterAuth{
						ClientCertificateConfig: &v1beta1.ClientCertificateConfig{
							IssueClientCertificate: false,
						},
						Username: &adminUser,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.MasterAuth = &container.MasterAuth{
					ClientCertificateConfig: &container.ClientCertificateConfig{
						IssueClientCertificate: false,
					},
					Username: adminUser,
				}
			}),
		},
		"SuccessfulOnlyUsername": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.MasterAuth = &v1beta1.MasterAuth{
						Username: &adminUser,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.MasterAuth = &container.MasterAuth{
					Username: adminUser,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateMasterAuth(tc.args.params.MasterAuth, tc.args.cluster)
			if diff := cmp.Diff(tc.want.MasterAuth, tc.args.cluster.MasterAuth); diff != "" {
				t.Errorf("GenerateMasterAuth(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateMasterAuthorizedNetworksConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.MasterAuthorizedNetworksConfig = &v1beta1.MasterAuthorizedNetworksConfig{
						Enabled: gcp.BoolPtr(true),
						CidrBlocks: []*v1beta1.CidrBlock{
							{
								CidrBlock: "0.0.0.0/0",
							},
						},
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.MasterAuthorizedNetworksConfig = &container.MasterAuthorizedNetworksConfig{
					Enabled: true,
					CidrBlocks: []*container.CidrBlock{
						{
							CidrBlock: "0.0.0.0/0",
						},
					},
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateMasterAuthorizedNetworksConfig(tc.args.params.MasterAuthorizedNetworksConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.MasterAuthorizedNetworksConfig, tc.args.cluster.MasterAuthorizedNetworksConfig); diff != "" {
				t.Errorf("GenerateMasterAuthorizedNetworksConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateNetworkConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.NetworkConfig = &v1beta1.NetworkConfigSpec{
						EnableIntraNodeVisibility: true,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.NetworkConfig = &container.NetworkConfig{
					EnableIntraNodeVisibility: true,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateNetworkConfig(tc.args.params.NetworkConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.NetworkConfig, tc.args.cluster.NetworkConfig); diff != "" {
				t.Errorf("GenerateNetworkConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateNetworkPolicy(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.NetworkPolicy = &v1beta1.NetworkPolicy{
						Enabled:  gcp.BoolPtr(true),
						Provider: gcp.StringPtr("CALICO"),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.NetworkPolicy = &container.NetworkPolicy{
					Enabled:  true,
					Provider: "CALICO",
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateNetworkPolicy(tc.args.params.NetworkPolicy, tc.args.cluster)
			if diff := cmp.Diff(tc.want.NetworkPolicy, tc.args.cluster.NetworkPolicy); diff != "" {
				t.Errorf("GenerateNetworkPolicy(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGeneratePodSecurityPolicyConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.PodSecurityPolicyConfig = &v1beta1.PodSecurityPolicyConfig{
						Enabled: true,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.PodSecurityPolicyConfig = &container.PodSecurityPolicyConfig{
					Enabled: true,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GeneratePodSecurityPolicyConfig(tc.args.params.PodSecurityPolicyConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.PodSecurityPolicyConfig, tc.args.cluster.PodSecurityPolicyConfig); diff != "" {
				t.Errorf("GeneratePodSecurityPolicyConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGeneratePrivateClusterConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.PrivateClusterConfig = &v1beta1.PrivateClusterConfigSpec{
						EnablePeeringRouteSharing: gcp.BoolPtr(true),
						EnablePrivateEndpoint:     gcp.BoolPtr(true),
						EnablePrivateNodes:        gcp.BoolPtr(true),
						MasterIpv4CidrBlock:       gcp.StringPtr("0.0.0.0/0"),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.PrivateClusterConfig = &container.PrivateClusterConfig{
					EnablePeeringRouteSharing: true,
					EnablePrivateEndpoint:     true,
					EnablePrivateNodes:        true,
					MasterIpv4CidrBlock:       "0.0.0.0/0",
				}
			}),
		},
		"SuccessfulPartial": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.PrivateClusterConfig = &v1beta1.PrivateClusterConfigSpec{
						EnablePeeringRouteSharing: gcp.BoolPtr(true),
						MasterIpv4CidrBlock:       gcp.StringPtr("0.0.0.0/0"),
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.PrivateClusterConfig = &container.PrivateClusterConfig{
					EnablePeeringRouteSharing: true,
					EnablePrivateEndpoint:     false,
					EnablePrivateNodes:        false,
					MasterIpv4CidrBlock:       "0.0.0.0/0",
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GeneratePrivateClusterConfig(tc.args.params.PrivateClusterConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.PrivateClusterConfig, tc.args.cluster.PrivateClusterConfig); diff != "" {
				t.Errorf("GeneratePrivateClusterConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateResourceUsageExportConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.ResourceUsageExportConfig = &v1beta1.ResourceUsageExportConfig{
						EnableNetworkEgressMetering: gcp.BoolPtr(true),
						BigqueryDestination: &v1beta1.BigQueryDestination{
							DatasetID: "cool-id",
						},
						ConsumptionMeteringConfig: &v1beta1.ConsumptionMeteringConfig{
							Enabled: true,
						},
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.ResourceUsageExportConfig = &container.ResourceUsageExportConfig{
					EnableNetworkEgressMetering: true,
					BigqueryDestination: &container.BigQueryDestination{
						DatasetId: "cool-id",
					},
					ConsumptionMeteringConfig: &container.ConsumptionMeteringConfig{
						Enabled: true,
					},
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateResourceUsageExportConfig(tc.args.params.ResourceUsageExportConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.ResourceUsageExportConfig, tc.args.cluster.ResourceUsageExportConfig); diff != "" {
				t.Errorf("GenerateResourceUsageExportConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateTierSettings(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.TierSettings = &v1beta1.TierSettings{
						Tier: "STANDARD",
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.TierSettings = &container.TierSettings{
					Tier: "STANDARD",
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateTierSettings(tc.args.params.TierSettings, tc.args.cluster)
			if diff := cmp.Diff(tc.want.TierSettings, tc.args.cluster.TierSettings); diff != "" {
				t.Errorf("GenerateTierSettings(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateVerticalPodAutoscaling(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.VerticalPodAutoscaling = &v1beta1.VerticalPodAutoscaling{
						Enabled: true,
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.VerticalPodAutoscaling = &container.VerticalPodAutoscaling{
					Enabled: true,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateVerticalPodAutoscaling(tc.args.params.VerticalPodAutoscaling, tc.args.cluster)
			if diff := cmp.Diff(tc.want.VerticalPodAutoscaling, tc.args.cluster.VerticalPodAutoscaling); diff != "" {
				t.Errorf("GenerateVerticalPodAutoscaling(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateWorkloadIdentityConfig(t *testing.T) {
	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}

	tests := map[string]struct {
		args args
		want *container.Cluster
	}{
		"Successful": {
			args: args{
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.WorkloadIdentityConfig = &v1beta1.WorkloadIdentityConfig{
						IdentityNamespace: "cool-namespace",
					}
				}),
			},
			want: cluster(func(c *container.Cluster) {
				c.WorkloadIdentityConfig = &container.WorkloadIdentityConfig{
					IdentityNamespace: "cool-namespace",
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: cluster(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateWorkloadIdentityConfig(tc.args.params.WorkloadIdentityConfig, tc.args.cluster)
			if diff := cmp.Diff(tc.want.WorkloadIdentityConfig, tc.args.cluster.WorkloadIdentityConfig); diff != "" {
				t.Errorf("GenerateWorkloadIdentityConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	var adminUser = "admin"

	type args struct {
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}
	type want struct {
		params *v1beta1.GKEClusterParameters
	}
	tests := map[string]struct {
		args args
		want want
	}{
		"SomeFilled": {
			args: args{
				cluster: cluster(func(c *container.Cluster) {
					c.AddonsConfig = &container.AddonsConfig{
						HttpLoadBalancing: &container.HttpLoadBalancing{
							Disabled: true,
						},
					}
					c.IpAllocationPolicy = &container.IPAllocationPolicy{
						ClusterIpv4CidrBlock: "0.0.0.0/0",
					}
				}),
				params: params(),
			},
			want: want{
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.AddonsConfig = &v1beta1.AddonsConfig{
						HTTPLoadBalancing: &v1beta1.HTTPLoadBalancing{
							Disabled: gcp.BoolPtr(true),
						},
					}
					p.IPAllocationPolicy = &v1beta1.IPAllocationPolicy{
						ClusterIpv4CidrBlock: gcp.StringPtr("0.0.0.0/0"),
					}
				}),
			},
		},
		"SomeFilledOverride": {
			args: args{
				cluster: cluster(func(c *container.Cluster) {
					c.AddonsConfig = &container.AddonsConfig{
						HttpLoadBalancing: &container.HttpLoadBalancing{
							Disabled: true,
						},
					}
					c.IpAllocationPolicy = &container.IPAllocationPolicy{
						ClusterIpv4CidrBlock: "0.0.0.0/0",
					}
					c.MasterAuth = &container.MasterAuth{
						Username: "someUser",
					}
				}),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.MasterAuth = &v1beta1.MasterAuth{
						Username: &adminUser,
					}
				}),
			},
			want: want{
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.AddonsConfig = &v1beta1.AddonsConfig{
						HTTPLoadBalancing: &v1beta1.HTTPLoadBalancing{
							Disabled: gcp.BoolPtr(true),
						},
					}
					p.IPAllocationPolicy = &v1beta1.IPAllocationPolicy{
						ClusterIpv4CidrBlock: gcp.StringPtr("0.0.0.0/0"),
					}
					p.MasterAuth = &v1beta1.MasterAuth{
						Username: &adminUser,
					}
				}),
			},
		},
		"NoneFilled": {
			args: args{
				cluster: cluster(),
				params:  params(),
			},
			want: want{
				params: params(),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.params, *tc.args.cluster)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	falseVal := false

	type args struct {
		name    string
		cluster *container.Cluster
		params  *v1beta1.GKEClusterParameters
	}
	type want struct {
		upToDate bool
		isErr    bool
	}
	tests := map[string]struct {
		args args
		want want
	}{
		"UpToDate": {
			args: args{
				name:    name,
				cluster: cluster(),
				params:  params(),
			},
			want: want{
				upToDate: true,
				isErr:    false,
			},
		},
		"UpToDateWithOutputFields": {
			args: args{
				name:    name,
				cluster: cluster(addOutputFields),
				params:  params(),
			},
			want: want{
				upToDate: true,
				isErr:    false,
			},
		},
		"UpToDateIgnoreRefs": {
			args: args{
				name:    name,
				cluster: cluster(),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.NetworkRef = &v1beta1.NetworkURIReferencerForGKECluster{
						NetworkURIReferencer: computev1beta1.NetworkURIReferencer{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-network",
							},
						},
					}
				}),
			},
			want: want{
				upToDate: true,
				isErr:    false,
			},
		},
		"UpToDateIgnoreForceSendFields": {
			args: args{
				name: name,
				cluster: cluster(func(c *container.Cluster) {
					c.AddonsConfig = &container.AddonsConfig{
						KubernetesDashboard: &container.KubernetesDashboard{
							Disabled: true,
						},
					}
				}),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.AddonsConfig = &v1beta1.AddonsConfig{
						KubernetesDashboard: &v1beta1.KubernetesDashboard{
							Disabled: gcp.BoolPtr(true),
						},
					}
				}),
			},
			want: want{
				upToDate: true,
				isErr:    false,
			},
		},
		"NeedsUpdate": {
			args: args{
				name: name,
				cluster: cluster(func(c *container.Cluster) {
					c.AddonsConfig = &container.AddonsConfig{
						HttpLoadBalancing: &container.HttpLoadBalancing{
							Disabled: true,
						},
					}
					c.IpAllocationPolicy = &container.IPAllocationPolicy{
						ClusterIpv4CidrBlock: "0.0.0.0/0",
					}
				}),
				params: params(func(p *v1beta1.GKEClusterParameters) {
					p.AddonsConfig = &v1beta1.AddonsConfig{
						HTTPLoadBalancing: &v1beta1.HTTPLoadBalancing{
							Disabled: &falseVal,
						},
					}
				}),
			},
			want: want{
				upToDate: false,
				isErr:    false,
			},
		},
		"NoUpdateNotBootstrapNodePool": {
			args: args{
				name: name,
				cluster: cluster(func(c *container.Cluster) {
					sc := &container.StatusCondition{
						Code:    "cool-code",
						Message: "cool-message",
					}
					np := &container.NodePool{
						Conditions: []*container.StatusCondition{sc},
						Name:       "cool-node-pool",
					}
					c.NodePools = []*container.NodePool{np}
				}),
				params: params(),
			},
			want: want{
				upToDate: true,
				isErr:    false,
			},
		},
		"NeedsUpdateBootstrapNodePool": {
			args: args{
				name: name,
				cluster: cluster(func(c *container.Cluster) {
					sc := &container.StatusCondition{
						Code:    "cool-code",
						Message: "cool-message",
					}
					np := &container.NodePool{
						Conditions: []*container.StatusCondition{sc},
						Name:       BootstrapNodePoolName,
					}
					c.NodePools = []*container.NodePool{np}
				}),
				params: params(),
			},
			want: want{
				upToDate: false,
				isErr:    false,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r, _, err := IsUpToDate(tc.args.name, tc.args.params, tc.args.cluster)
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want upToDate, +got upToDate:\n%s", diff)
			}
		})
	}
}

func TestGetFullyQualifiedParent(t *testing.T) {
	type args struct {
		project string
		params  v1beta1.GKEClusterParameters
	}
	tests := map[string]struct {
		args args
		want string
	}{
		"Successful": {
			args: args{
				project: project,
				params:  *params(),
			},
			want: fmt.Sprintf(ParentFormat, project, location),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := GetFullyQualifiedParent(tc.args.project, tc.args.params)
			if diff := cmp.Diff(tc.want, s); diff != "" {
				t.Errorf("GetFullyQualifiedParent(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetFullyQualifiedName(t *testing.T) {
	type args struct {
		project string
		params  v1beta1.GKEClusterParameters
		name    string
	}
	tests := map[string]struct {
		args args
		want string
	}{
		"Successful": {
			args: args{
				project: project,
				params:  *params(),
				name:    name,
			},
			want: fmt.Sprintf(ClusterNameFormat, project, location, name),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := GetFullyQualifiedName(tc.args.project, tc.args.params, tc.args.name)
			if diff := cmp.Diff(tc.want, s); diff != "" {
				t.Errorf("GetFullyQualifiedName(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetFullyQualifiedBNP(t *testing.T) {
	clusterName := fmt.Sprintf(ClusterNameFormat, project, location, name)
	tests := map[string]struct {
		name string
		want string
	}{
		"Successful": {
			name: clusterName,
			want: fmt.Sprintf(BNPNameFormat, clusterName, BootstrapNodePoolName),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := GetFullyQualifiedBNP(tc.name)
			if diff := cmp.Diff(tc.want, s); diff != "" {
				t.Errorf("GetFullyQualifiedBNP(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateClientConfig(t *testing.T) {
	name := "gke-cluster"
	endpoint := "endpoint"
	username := "username"
	password := "password"
	clusterCA, _ := base64.StdEncoding.DecodeString("clusterCA")
	clientCert, _ := base64.StdEncoding.DecodeString("clientCert")
	clientKey, _ := base64.StdEncoding.DecodeString("clientKey")

	type want struct {
		out clientcmdapi.Config
		err error
	}
	cases := map[string]struct {
		in   *container.Cluster
		want want
	}{
		"Full": {
			in: &container.Cluster{
				Name:     name,
				Endpoint: endpoint,
				MasterAuth: &container.MasterAuth{
					Username:             username,
					Password:             password,
					ClusterCaCertificate: base64.StdEncoding.EncodeToString(clusterCA),
					ClientCertificate:    base64.StdEncoding.EncodeToString(clientCert),
					ClientKey:            base64.StdEncoding.EncodeToString(clientKey),
				},
			},
			want: want{
				out: clientcmdapi.Config{
					Clusters: map[string]*clientcmdapi.Cluster{
						name: {
							Server:                   fmt.Sprintf("https://%s", endpoint),
							CertificateAuthorityData: clusterCA,
						},
					},
					Contexts: map[string]*clientcmdapi.Context{
						name: {
							Cluster:  name,
							AuthInfo: name,
						},
					},
					AuthInfos: map[string]*clientcmdapi.AuthInfo{
						name: {
							Username:              username,
							Password:              password,
							ClientKeyData:         clientKey,
							ClientCertificateData: clientCert,
						},
					},
					CurrentContext: name,
				},
			},
		},
		"Empty": {
			in: &container.Cluster{},
			want: want{
				out: clientcmdapi.Config{},
				err: errors.New(errNoSecretInfo),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := GenerateClientConfig(tc.in)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("GenerateClientConfig(...): -want error, +got error:\n%s", diff)
				return
			}
			if diff := cmp.Diff(tc.want.out, got); diff != "" {
				t.Errorf("GenerateClientConfig(...): -want config, +got config:\n%s", diff)
			}
		})
	}

}
