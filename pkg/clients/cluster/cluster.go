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
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1beta1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/crossplaneio/stack-gcp/apis/container/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

const (
	// BootstrapNodePoolName is the name of the node pool that is used to
	// boostrap GKE cluster creation.
	BootstrapNodePoolName = "crossplane-bootstrap"

	// BNPNameFormat is the format for the fully qualified name of the bootstrap node pool.
	BNPNameFormat = "%s/nodePools/%s"

	// ParentFormat is the format for the fully qualified name of a cluster parent.
	ParentFormat = "projects/%s/locations/%s"

	// ClusterNameFormat is the format for the fully qualified name of a cluster.
	ClusterNameFormat = "projects/%s/locations/%s/clusters/%s"
)

const (
	errNoSecretInfo  = "missing secret information for GKE cluster"
	errCheckUpToDate = "unable to determine if external resource is up to date"
)

// AddNodePoolForCreate inserts the default node pool into *container.Cluster so
// that it can be provisioned successfully.
func AddNodePoolForCreate(in *container.Cluster) {
	pool := &container.NodePool{
		Name:             BootstrapNodePoolName,
		InitialNodeCount: 0,
	}
	in.NodePools = []*container.NodePool{pool}
}

// GenerateCluster generates *container.Cluster instance from GKEClusterParameters.
func GenerateCluster(name string, in v1beta1.GKEClusterParameters, cluster *container.Cluster) { // nolint:gocyclo
	cluster.ClusterIpv4Cidr = gcp.StringValue(in.ClusterIpv4Cidr)
	cluster.Description = gcp.StringValue(in.Description)
	cluster.EnableKubernetesAlpha = gcp.BoolValue(in.EnableKubernetesAlpha)
	cluster.EnableTpu = gcp.BoolValue(in.EnableTpu)
	cluster.InitialClusterVersion = gcp.StringValue(in.InitialClusterVersion)
	cluster.LabelFingerprint = gcp.StringValue(in.LabelFingerprint)
	cluster.Locations = in.Locations
	cluster.LoggingService = gcp.StringValue(in.LoggingService)
	cluster.MonitoringService = gcp.StringValue(in.MonitoringService)
	cluster.Name = name
	cluster.Network = gcp.StringValue(in.Network)
	cluster.ResourceLabels = in.ResourceLabels
	cluster.Subnetwork = gcp.StringValue(in.Subnetwork)

	GenerateAddonsConfig(in.AddonsConfig, cluster)
	GenerateAuthenticatorGroupsConfig(in.AuthenticatorGroupsConfig, cluster)
	GenerateAutoscaling(in.Autoscaling, cluster)
	GenerateBinaryAuthorization(in.BinaryAuthorization, cluster)
	GenerateDatabaseEncryption(in.DatabaseEncryption, cluster)
	GenerateDefaultMaxPodsConstraint(in.DefaultMaxPodsConstraint, cluster)
	GenerateIPAllocationPolicy(in.IPAllocationPolicy, cluster)
	GenerateLegacyAbac(in.LegacyAbac, cluster)
	GenerateMaintenancePolicy(in.MaintenancePolicy, cluster)
	GenerateMasterAuth(in.MasterAuth, cluster)
	GenerateMasterAuthorizedNetworksConfig(in.MasterAuthorizedNetworksConfig, cluster)
	GenerateNetworkConfig(in.NetworkConfig, cluster)
	GenerateNetworkPolicy(in.NetworkPolicy, cluster)
	GeneratePodSecurityPolicyConfig(in.PodSecurityPolicyConfig, cluster)
	GeneratePrivateClusterConfig(in.PrivateClusterConfig, cluster)
	GenerateResourceUsageExportConfig(in.ResourceUsageExportConfig, cluster)
	GenerateTierSettings(in.TierSettings, cluster)
	GenerateVerticalPodAutoscaling(in.VerticalPodAutoscaling, cluster)
	GenerateWorkloadIdentityConfig(in.WorkloadIdentityConfig, cluster)
}

// GenerateAddonsConfig generates *container.AddonsConfig from *AddonsConfig.
func GenerateAddonsConfig(in *v1beta1.AddonsConfig, cluster *container.Cluster) { // nolint:gocyclo
	if in != nil {
		if cluster.AddonsConfig == nil {
			cluster.AddonsConfig = &container.AddonsConfig{}
		}
		if in.CloudRunConfig != nil {
			if cluster.AddonsConfig.CloudRunConfig == nil {
				cluster.AddonsConfig.CloudRunConfig = &container.CloudRunConfig{}
			}
			cluster.AddonsConfig.CloudRunConfig.Disabled = gcp.BoolValue(in.CloudRunConfig.Disabled)
			cluster.AddonsConfig.CloudRunConfig.ForceSendFields = []string{"Disabled"}
		}
		if in.HorizontalPodAutoscaling != nil {
			if cluster.AddonsConfig.HorizontalPodAutoscaling == nil {
				cluster.AddonsConfig.HorizontalPodAutoscaling = &container.HorizontalPodAutoscaling{}
			}
			cluster.AddonsConfig.HorizontalPodAutoscaling.Disabled = gcp.BoolValue(in.HorizontalPodAutoscaling.Disabled)
			cluster.AddonsConfig.HorizontalPodAutoscaling.ForceSendFields = []string{"Disabled"}
		}
		if in.HTTPLoadBalancing != nil && in.HTTPLoadBalancing.Disabled != nil {
			if cluster.AddonsConfig.HttpLoadBalancing == nil {
				cluster.AddonsConfig.HttpLoadBalancing = &container.HttpLoadBalancing{}
			}
			cluster.AddonsConfig.HttpLoadBalancing.Disabled = gcp.BoolValue(in.HTTPLoadBalancing.Disabled)
			cluster.AddonsConfig.HttpLoadBalancing.ForceSendFields = []string{"Disabled"}
		}
		if in.IstioConfig != nil {
			if cluster.AddonsConfig.IstioConfig == nil {
				cluster.AddonsConfig.IstioConfig = &container.IstioConfig{}
			}
			cluster.AddonsConfig.IstioConfig.Auth = gcp.StringValue(in.IstioConfig.Auth)
			cluster.AddonsConfig.IstioConfig.Disabled = gcp.BoolValue(in.IstioConfig.Disabled)
			cluster.AddonsConfig.IstioConfig.ForceSendFields = []string{"Disabled"}
		}
		if in.KubernetesDashboard != nil {
			if cluster.AddonsConfig.KubernetesDashboard == nil {
				cluster.AddonsConfig.KubernetesDashboard = &container.KubernetesDashboard{}
			}
			cluster.AddonsConfig.KubernetesDashboard.Disabled = gcp.BoolValue(in.KubernetesDashboard.Disabled)
			cluster.AddonsConfig.KubernetesDashboard.ForceSendFields = []string{"Disabled"}
		}
		if in.NetworkPolicyConfig != nil {
			if cluster.AddonsConfig.NetworkPolicyConfig == nil {
				cluster.AddonsConfig.NetworkPolicyConfig = &container.NetworkPolicyConfig{}
			}
			cluster.AddonsConfig.NetworkPolicyConfig.Disabled = gcp.BoolValue(in.NetworkPolicyConfig.Disabled)
			cluster.AddonsConfig.NetworkPolicyConfig.ForceSendFields = []string{"Disabled"}
		}
	}
}

// GenerateAuthenticatorGroupsConfig generates *container.AuthenticatorGroupsConfig from *AuthenticatorGroupsConfig.
func GenerateAuthenticatorGroupsConfig(in *v1beta1.AuthenticatorGroupsConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.AuthenticatorGroupsConfig == nil {
			cluster.AuthenticatorGroupsConfig = &container.AuthenticatorGroupsConfig{}
		}
		cluster.AuthenticatorGroupsConfig.Enabled = gcp.BoolValue(in.Enabled)
		cluster.AuthenticatorGroupsConfig.SecurityGroup = gcp.StringValue(in.SecurityGroup)
	}
}

// GenerateAutoscaling generates *container.ClusterAutoscaling from *ClusterAutoscaling.
func GenerateAutoscaling(in *v1beta1.ClusterAutoscaling, cluster *container.Cluster) {
	if in != nil {
		if cluster.Autoscaling == nil {
			cluster.Autoscaling = &container.ClusterAutoscaling{}
		}
		cluster.Autoscaling.AutoprovisioningLocations = in.AutoprovisioningLocations
		cluster.Autoscaling.EnableNodeAutoprovisioning = gcp.BoolValue(in.EnableNodeAutoprovisioning)

		if in.AutoprovisioningNodePoolDefaults != nil {
			if cluster.Autoscaling.AutoprovisioningNodePoolDefaults == nil {
				cluster.Autoscaling.AutoprovisioningNodePoolDefaults = &container.AutoprovisioningNodePoolDefaults{}
			}
			cluster.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes = in.AutoprovisioningNodePoolDefaults.OauthScopes
			cluster.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount = gcp.StringValue(in.AutoprovisioningNodePoolDefaults.ServiceAccount)
		}

		if len(in.ResourceLimits) > 0 {
			cluster.Autoscaling.ResourceLimits = make([]*container.ResourceLimit, len(in.ResourceLimits))
		}
		for i, limit := range in.ResourceLimits {
			if limit != nil {
				cluster.Autoscaling.ResourceLimits[i] = &container.ResourceLimit{
					Maximum:      gcp.Int64Value(limit.Maximum),
					Minimum:      gcp.Int64Value(limit.Minimum),
					ResourceType: gcp.StringValue(limit.ResourceType),
				}
			}
		}
	}
}

// GenerateBinaryAuthorization generates *container.BinaryAuthorization from *BinaryAuthorization.
func GenerateBinaryAuthorization(in *v1beta1.BinaryAuthorization, cluster *container.Cluster) {
	if in != nil {
		if cluster.BinaryAuthorization == nil {
			cluster.BinaryAuthorization = &container.BinaryAuthorization{}
		}
		cluster.BinaryAuthorization.Enabled = in.Enabled
	}
}

// GenerateDatabaseEncryption generates *container.DatabaseEncryption from *DatabaseEncryption.
func GenerateDatabaseEncryption(in *v1beta1.DatabaseEncryption, cluster *container.Cluster) {
	if in != nil {
		if cluster.DatabaseEncryption == nil {
			cluster.DatabaseEncryption = &container.DatabaseEncryption{}
		}
		cluster.DatabaseEncryption.KeyName = gcp.StringValue(in.KeyName)
		cluster.DatabaseEncryption.State = gcp.StringValue(in.State)
	}
}

// GenerateDefaultMaxPodsConstraint generates *container.MaxPodsConstraint from *DefaultMaxPodsConstraint.
func GenerateDefaultMaxPodsConstraint(in *v1beta1.MaxPodsConstraint, cluster *container.Cluster) {
	if in != nil {
		if cluster.DefaultMaxPodsConstraint == nil {
			cluster.DefaultMaxPodsConstraint = &container.MaxPodsConstraint{}
		}
		cluster.DefaultMaxPodsConstraint.MaxPodsPerNode = in.MaxPodsPerNode
	}
}

// GenerateIPAllocationPolicy generates *container.MaxPodsConstraint from *IpAllocationPolicy.
func GenerateIPAllocationPolicy(in *v1beta1.IPAllocationPolicy, cluster *container.Cluster) {
	if in != nil {
		if cluster.IpAllocationPolicy == nil {
			cluster.IpAllocationPolicy = &container.IPAllocationPolicy{}
		}
		cluster.IpAllocationPolicy.AllowRouteOverlap = gcp.BoolValue(in.AllowRouteOverlap)
		cluster.IpAllocationPolicy.ClusterIpv4CidrBlock = gcp.StringValue(in.ClusterIpv4CidrBlock)
		cluster.IpAllocationPolicy.ClusterSecondaryRangeName = gcp.StringValue(in.ClusterSecondaryRangeName)
		cluster.IpAllocationPolicy.CreateSubnetwork = gcp.BoolValue(in.CreateSubnetwork)
		cluster.IpAllocationPolicy.NodeIpv4CidrBlock = gcp.StringValue(in.NodeIpv4CidrBlock)
		cluster.IpAllocationPolicy.ServicesIpv4CidrBlock = gcp.StringValue(in.ServicesIpv4CidrBlock)
		cluster.IpAllocationPolicy.ServicesSecondaryRangeName = gcp.StringValue(in.DeepCopy().ServicesSecondaryRangeName)
		cluster.IpAllocationPolicy.SubnetworkName = gcp.StringValue(in.SubnetworkName)
		cluster.IpAllocationPolicy.TpuIpv4CidrBlock = gcp.StringValue(in.TpuIpv4CidrBlock)
		cluster.IpAllocationPolicy.UseIpAliases = gcp.BoolValue(in.UseIPAliases)
	}
}

// GenerateLegacyAbac generates *container.LegacyAbac from *LegacyAbac.
func GenerateLegacyAbac(in *v1beta1.LegacyAbac, cluster *container.Cluster) {
	if in != nil {
		if cluster.LegacyAbac == nil {
			cluster.LegacyAbac = &container.LegacyAbac{}
		}
		cluster.LegacyAbac.Enabled = in.Enabled
	}
}

// GenerateMaintenancePolicy generates *container.MaintenancePolicy from *MaintenancePolicy.
func GenerateMaintenancePolicy(in *v1beta1.MaintenancePolicySpec, cluster *container.Cluster) {
	if in != nil {
		if cluster.MaintenancePolicy == nil {
			cluster.MaintenancePolicy = &container.MaintenancePolicy{}
		}
		if cluster.MaintenancePolicy.Window == nil {
			cluster.MaintenancePolicy.Window = &container.MaintenanceWindow{}
		}
		if cluster.MaintenancePolicy.Window.DailyMaintenanceWindow == nil {
			cluster.MaintenancePolicy.Window.DailyMaintenanceWindow = &container.DailyMaintenanceWindow{}
		}
		cluster.MaintenancePolicy.Window.DailyMaintenanceWindow.StartTime = in.Window.DailyMaintenanceWindow.StartTime
	}
}

// GenerateMasterAuth generates *container.MasterAuth from *MasterAuth.
func GenerateMasterAuth(in *v1beta1.MasterAuth, cluster *container.Cluster) {
	if in != nil {
		if cluster.MasterAuth == nil {
			cluster.MasterAuth = &container.MasterAuth{}
		}
		cluster.MasterAuth.Username = gcp.StringValue(in.Username)

		if in.ClientCertificateConfig != nil {
			if cluster.MasterAuth.ClientCertificateConfig == nil {
				cluster.MasterAuth.ClientCertificateConfig = &container.ClientCertificateConfig{}
			}
			cluster.MasterAuth.ClientCertificateConfig.IssueClientCertificate = in.ClientCertificateConfig.IssueClientCertificate
		}
	}
}

// GenerateMasterAuthorizedNetworksConfig generates *container.MasterAuthorizedNetworksConfig from *MasterAuthorizedNetworksConfig.
func GenerateMasterAuthorizedNetworksConfig(in *v1beta1.MasterAuthorizedNetworksConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.MasterAuthorizedNetworksConfig == nil {
			cluster.MasterAuthorizedNetworksConfig = &container.MasterAuthorizedNetworksConfig{}
		}
		cluster.MasterAuthorizedNetworksConfig.Enabled = gcp.BoolValue(in.Enabled)

		if len(in.CidrBlocks) > 0 {
			cluster.MasterAuthorizedNetworksConfig.CidrBlocks = make([]*container.CidrBlock, len(in.CidrBlocks))
		}
		for i, cidr := range in.CidrBlocks {
			if cidr != nil {
				cluster.MasterAuthorizedNetworksConfig.CidrBlocks[i] = &container.CidrBlock{
					CidrBlock:   cidr.CidrBlock,
					DisplayName: gcp.StringValue(cidr.DisplayName),
				}
			}
		}
	}
}

// GenerateNetworkConfig generates *container.NetworkConfig from *NetworkConfig.
func GenerateNetworkConfig(in *v1beta1.NetworkConfigSpec, cluster *container.Cluster) {
	if in != nil {
		if cluster.NetworkConfig == nil {
			cluster.NetworkConfig = &container.NetworkConfig{}
		}
		cluster.NetworkConfig.EnableIntraNodeVisibility = in.EnableIntraNodeVisibility
	}
}

// GenerateNetworkPolicy generates *container.NetworkPolicy from *NetworkPolicy.
func GenerateNetworkPolicy(in *v1beta1.NetworkPolicy, cluster *container.Cluster) {
	if in != nil {
		if cluster.NetworkPolicy == nil {
			cluster.NetworkPolicy = &container.NetworkPolicy{}
		}
		cluster.NetworkPolicy.Enabled = gcp.BoolValue(in.Enabled)
		cluster.NetworkPolicy.Provider = gcp.StringValue(in.Provider)
	}
}

// GeneratePodSecurityPolicyConfig generates *container.PodSecurityPolicyConfig from *PodSecurityPolicyConfig.
func GeneratePodSecurityPolicyConfig(in *v1beta1.PodSecurityPolicyConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.PodSecurityPolicyConfig == nil {
			cluster.PodSecurityPolicyConfig = &container.PodSecurityPolicyConfig{}
		}
		cluster.PodSecurityPolicyConfig.Enabled = in.Enabled
	}
}

// GeneratePrivateClusterConfig generates *container.PrivateClusterConfig from *PrivateClusterConfig.
func GeneratePrivateClusterConfig(in *v1beta1.PrivateClusterConfigSpec, cluster *container.Cluster) {
	if in != nil {
		if cluster.PrivateClusterConfig == nil {
			cluster.PrivateClusterConfig = &container.PrivateClusterConfig{}
		}
		cluster.PrivateClusterConfig.EnablePeeringRouteSharing = gcp.BoolValue(in.EnablePeeringRouteSharing)
		cluster.PrivateClusterConfig.EnablePrivateEndpoint = gcp.BoolValue(in.EnablePrivateEndpoint)
		cluster.PrivateClusterConfig.EnablePrivateNodes = gcp.BoolValue(in.EnablePrivateNodes)
		cluster.PrivateClusterConfig.MasterIpv4CidrBlock = gcp.StringValue(in.MasterIpv4CidrBlock)
	}
}

// GenerateResourceUsageExportConfig generates *container.ResourceUsageExportConfig from *ResourceUsageExportConfig.
func GenerateResourceUsageExportConfig(in *v1beta1.ResourceUsageExportConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.ResourceUsageExportConfig == nil {
			cluster.ResourceUsageExportConfig = &container.ResourceUsageExportConfig{}
		}
		cluster.ResourceUsageExportConfig.EnableNetworkEgressMetering = gcp.BoolValue(in.EnableNetworkEgressMetering)

		if in.BigqueryDestination != nil {
			if cluster.ResourceUsageExportConfig.BigqueryDestination == nil {
				cluster.ResourceUsageExportConfig.BigqueryDestination = &container.BigQueryDestination{}
			}
			cluster.ResourceUsageExportConfig.BigqueryDestination.DatasetId = in.BigqueryDestination.DatasetID
		}

		if in.ConsumptionMeteringConfig != nil {
			if cluster.ResourceUsageExportConfig.ConsumptionMeteringConfig == nil {
				cluster.ResourceUsageExportConfig.ConsumptionMeteringConfig = &container.ConsumptionMeteringConfig{}
			}
			cluster.ResourceUsageExportConfig.ConsumptionMeteringConfig.Enabled = in.ConsumptionMeteringConfig.Enabled
		}
	}
}

// GenerateTierSettings generates *container.TierSettings from *TierSettings.
func GenerateTierSettings(in *v1beta1.TierSettings, cluster *container.Cluster) {
	if in != nil {
		if cluster.TierSettings == nil {
			cluster.TierSettings = &container.TierSettings{}
		}
		cluster.TierSettings.Tier = in.Tier
	}
}

// GenerateVerticalPodAutoscaling generates *container.VerticalPodAutoscaling from *VerticalPodAutoscaling.
func GenerateVerticalPodAutoscaling(in *v1beta1.VerticalPodAutoscaling, cluster *container.Cluster) {
	if in != nil {
		if cluster.VerticalPodAutoscaling == nil {
			cluster.VerticalPodAutoscaling = &container.VerticalPodAutoscaling{}
		}
		cluster.VerticalPodAutoscaling.Enabled = in.Enabled
	}
}

// GenerateWorkloadIdentityConfig generates *container.WorkloadIdentityConfig from *WorkloadIdentityConfig.
func GenerateWorkloadIdentityConfig(in *v1beta1.WorkloadIdentityConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.WorkloadIdentityConfig == nil {
			cluster.WorkloadIdentityConfig = &container.WorkloadIdentityConfig{}
		}
		cluster.WorkloadIdentityConfig.IdentityNamespace = in.IdentityNamespace
	}
}

// GenerateObservation produces GKEClusterObservation object from *container.Cluster object.
func GenerateObservation(in container.Cluster) v1beta1.GKEClusterObservation { // nolint:gocyclo
	o := v1beta1.GKEClusterObservation{
		CreateTime:           in.CreateTime,
		CurrentMasterVersion: in.CurrentMasterVersion,
		CurrentNodeCount:     in.CurrentNodeCount,
		CurrentNodeVersion:   in.CurrentNodeVersion,
		Endpoint:             in.Endpoint,
		ExpireTime:           in.ExpireTime,
		Location:             in.Location,
		NodeIpv4CidrSize:     in.NodeIpv4CidrSize,
		SelfLink:             in.SelfLink,
		ServicesIpv4Cidr:     in.ServicesIpv4Cidr,
		Status:               in.Status,
		StatusMessage:        in.StatusMessage,
		TpuIpv4CidrBlock:     in.TpuIpv4CidrBlock,
		Zone:                 in.Zone,
	}

	if in.MaintenancePolicy != nil {
		if in.MaintenancePolicy.Window != nil {
			if in.MaintenancePolicy.Window.DailyMaintenanceWindow != nil {
				o.MaintenancePolicy = &v1beta1.MaintenancePolicyStatus{
					Window: v1beta1.MaintenanceWindowStatus{
						DailyMaintenanceWindow: v1beta1.DailyMaintenanceWindowStatus{
							Duration: in.MaintenancePolicy.Window.DailyMaintenanceWindow.Duration,
						},
					},
				}
			}
		}
	}

	if in.NetworkConfig != nil {
		o.NetworkConfig = &v1beta1.NetworkConfigStatus{
			Network:    in.NetworkConfig.Network,
			Subnetwork: in.NetworkConfig.Subnetwork,
		}
	}

	if in.PrivateClusterConfig != nil {
		o.PrivateClusterConfig = &v1beta1.PrivateClusterConfigStatus{
			PrivateEndpoint: in.PrivateClusterConfig.PrivateEndpoint,
			PublicEndpoint:  in.PrivateClusterConfig.PublicEndpoint,
		}
	}

	for _, condition := range in.Conditions {
		if condition != nil {
			o.Conditions = append(o.Conditions, &v1beta1.StatusCondition{
				Code:    condition.Code,
				Message: condition.Message,
			})
		}
	}

	for _, nodePool := range in.NodePools {
		if nodePool != nil {
			np := &v1beta1.NodePoolClusterStatus{
				InstanceGroupUrls: nodePool.InstanceGroupUrls,
				Name:              nodePool.Name,
				PodIpv4CidrSize:   nodePool.PodIpv4CidrSize,
				SelfLink:          nodePool.SelfLink,
				Status:            nodePool.Status,
				StatusMessage:     nodePool.StatusMessage,
				Version:           nodePool.Version,
			}
			if nodePool.Autoscaling != nil {
				np.Autoscaling = &v1beta1.NodePoolAutoscalingClusterStatus{
					Autoprovisioned: nodePool.Autoscaling.Autoprovisioned,
					Enabled:         nodePool.Autoscaling.Enabled,
					MaxNodeCount:    nodePool.Autoscaling.MaxNodeCount,
					MinNodeCount:    nodePool.Autoscaling.MinNodeCount,
				}
			}
			if nodePool.Config != nil {
				np.Config = &v1beta1.NodeConfigClusterStatus{
					DiskSizeGb:     nodePool.Config.DiskSizeGb,
					DiskType:       nodePool.Config.DiskType,
					ImageType:      nodePool.Config.ImageType,
					Labels:         nodePool.Config.Labels,
					LocalSsdCount:  nodePool.Config.LocalSsdCount,
					MachineType:    nodePool.Config.MachineType,
					Metadata:       nodePool.Config.Metadata,
					MinCPUPlatform: nodePool.Config.MinCpuPlatform,
					OauthScopes:    nodePool.Config.OauthScopes,
					Preemptible:    nodePool.Config.Preemptible,
					ServiceAccount: nodePool.Config.ServiceAccount,
					Tags:           nodePool.Config.Tags,
				}
				for _, a := range nodePool.Config.Accelerators {
					if a != nil {
						np.Config.Accelerators = append(np.Config.Accelerators, &v1beta1.AcceleratorConfigClusterStatus{
							AcceleratorCount: a.AcceleratorCount,
							AcceleratorType:  a.AcceleratorType,
						})
					}
				}
				if nodePool.Config.SandboxConfig != nil {
					np.Config.SandboxConfig = &v1beta1.SandboxConfigClusterStatus{
						SandboxType: nodePool.Config.SandboxConfig.SandboxType,
					}
				}
				if nodePool.Config.ShieldedInstanceConfig != nil {
					np.Config.ShieldedInstanceConfig = &v1beta1.ShieldedInstanceConfigClusterStatus{
						EnableIntegrityMonitoring: nodePool.Config.ShieldedInstanceConfig.EnableIntegrityMonitoring,
						EnableSecureBoot:          nodePool.Config.ShieldedInstanceConfig.EnableSecureBoot,
					}
				}
				for _, t := range nodePool.Config.Taints {
					if t != nil {
						np.Config.Taints = append(np.Config.Taints, &v1beta1.NodeTaintClusterStatus{
							Effect: t.Effect,
							Key:    t.Key,
							Value:  t.Value,
						})
					}
				}
				if nodePool.Config.WorkloadMetadataConfig != nil {
					np.Config.WorkloadMetadataConfig = &v1beta1.WorkloadMetadataConfigClusterStatus{
						NodeMetadata: nodePool.Config.WorkloadMetadataConfig.NodeMetadata,
					}
				}
			}
			for _, c := range nodePool.Conditions {
				if c != nil {
					np.Conditions = append(np.Conditions, &v1beta1.StatusCondition{
						Code:    c.Code,
						Message: c.Message,
					})
				}
			}
			o.NodePools = append(o.NodePools, np)
		}
	}

	return o
}

// LateInitializeSpec fills unassigned fields with the values in container.Cluster object.
func LateInitializeSpec(spec *v1beta1.GKEClusterParameters, in container.Cluster) { // nolint:gocyclo
	if in.AddonsConfig != nil {
		if spec.AddonsConfig == nil {
			spec.AddonsConfig = &v1beta1.AddonsConfig{}
		}
		if spec.AddonsConfig.CloudRunConfig == nil && in.AddonsConfig.CloudRunConfig != nil {
			spec.AddonsConfig.CloudRunConfig = &v1beta1.CloudRunConfig{
				Disabled: gcp.BoolPtr(in.AddonsConfig.CloudRunConfig.Disabled),
			}
		}
		if spec.AddonsConfig.HorizontalPodAutoscaling == nil && in.AddonsConfig.HorizontalPodAutoscaling != nil {
			spec.AddonsConfig.HorizontalPodAutoscaling = &v1beta1.HorizontalPodAutoscaling{
				Disabled: gcp.BoolPtr(in.AddonsConfig.HorizontalPodAutoscaling.Disabled),
			}
		}
		if spec.AddonsConfig.HTTPLoadBalancing == nil && in.AddonsConfig.HttpLoadBalancing != nil {
			spec.AddonsConfig.HTTPLoadBalancing = &v1beta1.HTTPLoadBalancing{
				Disabled: gcp.BoolPtr(in.AddonsConfig.HttpLoadBalancing.Disabled),
			}
		}
		if in.AddonsConfig.IstioConfig != nil {
			if spec.AddonsConfig.IstioConfig == nil {
				spec.AddonsConfig.IstioConfig = &v1beta1.IstioConfig{}
			}
			spec.AddonsConfig.IstioConfig.Auth = gcp.LateInitializeString(spec.AddonsConfig.IstioConfig.Auth, in.AddonsConfig.IstioConfig.Auth)
			spec.AddonsConfig.IstioConfig.Disabled = gcp.LateInitializeBool(spec.AddonsConfig.IstioConfig.Disabled, in.AddonsConfig.IstioConfig.Disabled)
		}
		if spec.AddonsConfig.KubernetesDashboard == nil && in.AddonsConfig.KubernetesDashboard != nil {
			spec.AddonsConfig.KubernetesDashboard = &v1beta1.KubernetesDashboard{
				Disabled: gcp.BoolPtr(in.AddonsConfig.KubernetesDashboard.Disabled),
			}
		}
		if spec.AddonsConfig.NetworkPolicyConfig == nil && in.AddonsConfig.NetworkPolicyConfig != nil {
			spec.AddonsConfig.NetworkPolicyConfig = &v1beta1.NetworkPolicyConfig{
				Disabled: gcp.BoolPtr(in.AddonsConfig.NetworkPolicyConfig.Disabled),
			}
		}
	}

	if in.AuthenticatorGroupsConfig != nil {
		if spec.AuthenticatorGroupsConfig == nil {
			spec.AuthenticatorGroupsConfig = &v1beta1.AuthenticatorGroupsConfig{}
		}
		spec.AuthenticatorGroupsConfig.Enabled = gcp.LateInitializeBool(spec.AuthenticatorGroupsConfig.Enabled, in.AuthenticatorGroupsConfig.Enabled)
		spec.AuthenticatorGroupsConfig.SecurityGroup = gcp.LateInitializeString(spec.AuthenticatorGroupsConfig.SecurityGroup, in.AuthenticatorGroupsConfig.SecurityGroup)
	}

	if in.Autoscaling != nil {
		if spec.Autoscaling == nil {
			spec.Autoscaling = &v1beta1.ClusterAutoscaling{}
		}
		spec.Autoscaling.AutoprovisioningLocations = gcp.LateInitializeStringSlice(spec.Autoscaling.AutoprovisioningLocations, in.Autoscaling.AutoprovisioningLocations)
		if in.Autoscaling.AutoprovisioningNodePoolDefaults != nil {
			if spec.Autoscaling.AutoprovisioningNodePoolDefaults == nil {
				spec.Autoscaling.AutoprovisioningNodePoolDefaults = &v1beta1.AutoprovisioningNodePoolDefaults{}
			}
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes = gcp.LateInitializeStringSlice(spec.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes, in.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes)
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount = gcp.LateInitializeString(spec.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount, in.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount)
		}
		spec.Autoscaling.EnableNodeAutoprovisioning = gcp.LateInitializeBool(spec.Autoscaling.EnableNodeAutoprovisioning, in.Autoscaling.EnableNodeAutoprovisioning)
		if len(in.Autoscaling.ResourceLimits) != 0 && len(spec.Autoscaling.ResourceLimits) == 0 {
			spec.Autoscaling.ResourceLimits = make([]*v1beta1.ResourceLimit, len(in.Autoscaling.ResourceLimits))
			for i, limit := range in.Autoscaling.ResourceLimits {
				spec.Autoscaling.ResourceLimits[i] = &v1beta1.ResourceLimit{
					Maximum:      &limit.Maximum,
					Minimum:      &limit.Minimum,
					ResourceType: &limit.ResourceType,
				}
			}
		}
	}

	if spec.BinaryAuthorization == nil && in.BinaryAuthorization != nil {
		spec.BinaryAuthorization = &v1beta1.BinaryAuthorization{
			Enabled: in.BinaryAuthorization.Enabled,
		}
	}

	spec.ClusterIpv4Cidr = gcp.LateInitializeString(spec.ClusterIpv4Cidr, in.ClusterIpv4Cidr)

	if in.DatabaseEncryption != nil {
		if spec.DatabaseEncryption == nil {
			spec.DatabaseEncryption = &v1beta1.DatabaseEncryption{}
		}
		spec.DatabaseEncryption.KeyName = gcp.LateInitializeString(spec.DatabaseEncryption.KeyName, in.DatabaseEncryption.KeyName)
		spec.DatabaseEncryption.State = gcp.LateInitializeString(spec.DatabaseEncryption.State, in.DatabaseEncryption.State)
	}

	if spec.DefaultMaxPodsConstraint == nil && in.DefaultMaxPodsConstraint != nil {
		spec.DefaultMaxPodsConstraint = &v1beta1.MaxPodsConstraint{
			MaxPodsPerNode: in.DefaultMaxPodsConstraint.MaxPodsPerNode,
		}
	}

	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)

	spec.EnableKubernetesAlpha = gcp.LateInitializeBool(spec.EnableKubernetesAlpha, in.EnableKubernetesAlpha)
	spec.EnableTpu = gcp.LateInitializeBool(spec.EnableTpu, in.EnableTpu)
	spec.InitialClusterVersion = gcp.LateInitializeString(spec.InitialClusterVersion, in.InitialClusterVersion)

	if in.IpAllocationPolicy != nil {
		if spec.IPAllocationPolicy == nil {
			spec.IPAllocationPolicy = &v1beta1.IPAllocationPolicy{}
		}
		spec.IPAllocationPolicy.AllowRouteOverlap = gcp.LateInitializeBool(spec.IPAllocationPolicy.AllowRouteOverlap, in.IpAllocationPolicy.AllowRouteOverlap)
		spec.IPAllocationPolicy.ClusterIpv4CidrBlock = gcp.LateInitializeString(spec.IPAllocationPolicy.ClusterIpv4CidrBlock, in.IpAllocationPolicy.ClusterIpv4CidrBlock)
		spec.IPAllocationPolicy.ClusterSecondaryRangeName = gcp.LateInitializeString(spec.IPAllocationPolicy.ClusterSecondaryRangeName, in.IpAllocationPolicy.ClusterSecondaryRangeName)
		spec.IPAllocationPolicy.CreateSubnetwork = gcp.LateInitializeBool(spec.IPAllocationPolicy.CreateSubnetwork, in.IpAllocationPolicy.CreateSubnetwork)
		spec.IPAllocationPolicy.NodeIpv4CidrBlock = gcp.LateInitializeString(spec.IPAllocationPolicy.NodeIpv4CidrBlock, in.IpAllocationPolicy.NodeIpv4CidrBlock)
		spec.IPAllocationPolicy.ServicesIpv4CidrBlock = gcp.LateInitializeString(spec.IPAllocationPolicy.ServicesIpv4CidrBlock, in.IpAllocationPolicy.ServicesIpv4CidrBlock)
		spec.IPAllocationPolicy.SubnetworkName = gcp.LateInitializeString(spec.IPAllocationPolicy.SubnetworkName, in.IpAllocationPolicy.SubnetworkName)
		spec.IPAllocationPolicy.TpuIpv4CidrBlock = gcp.LateInitializeString(spec.IPAllocationPolicy.TpuIpv4CidrBlock, in.IpAllocationPolicy.TpuIpv4CidrBlock)
		spec.IPAllocationPolicy.UseIPAliases = gcp.LateInitializeBool(spec.IPAllocationPolicy.UseIPAliases, in.IpAllocationPolicy.UseIpAliases)
	}

	spec.LabelFingerprint = gcp.LateInitializeString(spec.LabelFingerprint, in.LabelFingerprint)

	if spec.LegacyAbac == nil && in.LegacyAbac != nil {
		spec.LegacyAbac = &v1beta1.LegacyAbac{
			Enabled: in.LegacyAbac.Enabled,
		}
	}

	spec.Locations = gcp.LateInitializeStringSlice(spec.Locations, in.Locations)
	spec.LoggingService = gcp.LateInitializeString(spec.LoggingService, in.LoggingService)

	if spec.MaintenancePolicy == nil && in.MaintenancePolicy != nil {
		if in.MaintenancePolicy.Window != nil {
			if in.MaintenancePolicy.Window.DailyMaintenanceWindow != nil {
				spec.MaintenancePolicy = &v1beta1.MaintenancePolicySpec{
					Window: v1beta1.MaintenanceWindowSpec{
						DailyMaintenanceWindow: v1beta1.DailyMaintenanceWindowSpec{
							StartTime: in.MaintenancePolicy.Window.DailyMaintenanceWindow.StartTime,
						},
					},
				}
			}
		}
	}

	if in.MasterAuth != nil {
		if spec.MasterAuth == nil {
			spec.MasterAuth = &v1beta1.MasterAuth{}
		}
		if in.MasterAuth.ClientCertificateConfig != nil {
			spec.MasterAuth.ClientCertificateConfig = &v1beta1.ClientCertificateConfig{
				IssueClientCertificate: in.MasterAuth.ClientCertificateConfig.IssueClientCertificate,
			}
		}
		spec.MasterAuth.Username = gcp.LateInitializeString(spec.MasterAuth.Username, in.MasterAuth.Username)
	}

	if in.MasterAuthorizedNetworksConfig != nil {
		if spec.MasterAuthorizedNetworksConfig == nil {
			spec.MasterAuthorizedNetworksConfig = &v1beta1.MasterAuthorizedNetworksConfig{}
		}
		if len(in.MasterAuthorizedNetworksConfig.CidrBlocks) != 0 && len(spec.MasterAuthorizedNetworksConfig.CidrBlocks) == 0 {
			spec.MasterAuthorizedNetworksConfig.CidrBlocks = make([]*v1beta1.CidrBlock, len(in.MasterAuthorizedNetworksConfig.CidrBlocks))
			for i, block := range in.MasterAuthorizedNetworksConfig.CidrBlocks {
				spec.MasterAuthorizedNetworksConfig.CidrBlocks[i] = &v1beta1.CidrBlock{
					CidrBlock:   block.CidrBlock,
					DisplayName: &block.DisplayName,
				}
			}
		}
		spec.MasterAuthorizedNetworksConfig.Enabled = gcp.LateInitializeBool(spec.MasterAuthorizedNetworksConfig.Enabled, in.MasterAuthorizedNetworksConfig.Enabled)
	}

	spec.MonitoringService = gcp.LateInitializeString(spec.MonitoringService, in.MonitoringService)
	spec.Network = gcp.LateInitializeString(spec.Network, in.Network)

	if spec.NetworkConfig == nil && in.NetworkConfig != nil {
		spec.NetworkConfig = &v1beta1.NetworkConfigSpec{
			EnableIntraNodeVisibility: in.NetworkConfig.EnableIntraNodeVisibility,
		}
	}

	if in.NetworkPolicy != nil {
		if spec.NetworkPolicy == nil {
			spec.NetworkPolicy = &v1beta1.NetworkPolicy{}
		}
		spec.NetworkPolicy.Enabled = gcp.LateInitializeBool(spec.NetworkPolicy.Enabled, in.NetworkPolicy.Enabled)
		spec.NetworkPolicy.Provider = gcp.LateInitializeString(spec.NetworkPolicy.Provider, in.NetworkPolicy.Provider)
	}

	if spec.PodSecurityPolicyConfig == nil && in.PodSecurityPolicyConfig != nil {
		spec.PodSecurityPolicyConfig = &v1beta1.PodSecurityPolicyConfig{
			Enabled: in.PodSecurityPolicyConfig.Enabled,
		}
	}

	if in.PrivateClusterConfig != nil {
		if spec.PrivateClusterConfig == nil {
			spec.PrivateClusterConfig = &v1beta1.PrivateClusterConfigSpec{}
		}
		spec.PrivateClusterConfig.EnablePeeringRouteSharing = gcp.LateInitializeBool(spec.PrivateClusterConfig.EnablePeeringRouteSharing, in.PrivateClusterConfig.EnablePeeringRouteSharing)
		spec.PrivateClusterConfig.EnablePrivateEndpoint = gcp.LateInitializeBool(spec.PrivateClusterConfig.EnablePrivateEndpoint, in.PrivateClusterConfig.EnablePrivateEndpoint)
		spec.PrivateClusterConfig.EnablePrivateNodes = gcp.LateInitializeBool(spec.PrivateClusterConfig.EnablePrivateNodes, in.PrivateClusterConfig.EnablePrivateNodes)
		spec.PrivateClusterConfig.MasterIpv4CidrBlock = gcp.LateInitializeString(spec.PrivateClusterConfig.MasterIpv4CidrBlock, in.PrivateClusterConfig.MasterIpv4CidrBlock)
	}

	spec.ResourceLabels = gcp.LateInitializeStringMap(spec.ResourceLabels, in.ResourceLabels)

	if in.ResourceUsageExportConfig != nil {
		if spec.ResourceUsageExportConfig == nil {
			spec.ResourceUsageExportConfig = &v1beta1.ResourceUsageExportConfig{}
		}
		if spec.ResourceUsageExportConfig.BigqueryDestination == nil && in.ResourceUsageExportConfig.BigqueryDestination != nil {
			spec.ResourceUsageExportConfig.BigqueryDestination = &v1beta1.BigQueryDestination{
				DatasetID: in.ResourceUsageExportConfig.BigqueryDestination.DatasetId,
			}
		}
		if spec.ResourceUsageExportConfig.ConsumptionMeteringConfig == nil && in.ResourceUsageExportConfig.ConsumptionMeteringConfig != nil {
			spec.ResourceUsageExportConfig.ConsumptionMeteringConfig = &v1beta1.ConsumptionMeteringConfig{
				Enabled: in.ResourceUsageExportConfig.ConsumptionMeteringConfig.Enabled,
			}
		}
		spec.ResourceUsageExportConfig.EnableNetworkEgressMetering = gcp.LateInitializeBool(spec.ResourceUsageExportConfig.EnableNetworkEgressMetering, in.ResourceUsageExportConfig.EnableNetworkEgressMetering)
	}

	spec.Subnetwork = gcp.LateInitializeString(spec.Subnetwork, in.Subnetwork)

	if spec.TierSettings == nil && in.TierSettings != nil {
		spec.TierSettings = &v1beta1.TierSettings{
			Tier: in.TierSettings.Tier,
		}
	}

	if spec.VerticalPodAutoscaling == nil && in.VerticalPodAutoscaling != nil {
		spec.VerticalPodAutoscaling = &v1beta1.VerticalPodAutoscaling{
			Enabled: in.VerticalPodAutoscaling.Enabled,
		}
	}

	if spec.WorkloadIdentityConfig == nil && in.WorkloadIdentityConfig != nil {
		spec.WorkloadIdentityConfig = &v1beta1.WorkloadIdentityConfig{
			IdentityNamespace: in.WorkloadIdentityConfig.IdentityNamespace,
		}
	}

}

// newAddonsConfigUpdateFn returns a function that updates the AddonsConfig of a cluster.
func newAddonsConfigUpdateFn(in *v1beta1.AddonsConfig) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateAddonsConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredAddonsConfig: out.AddonsConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newAutoscalingUpdateFn returns a function that updates the Autoscaling of a cluster.
func newAutoscalingUpdateFn(in *v1beta1.ClusterAutoscaling) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateAutoscaling(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredClusterAutoscaling: out.Autoscaling,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newBinaryAuthorizationUpdateFn returns a function that updates the BinaryAuthorization of a cluster.
func newBinaryAuthorizationUpdateFn(in *v1beta1.BinaryAuthorization) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateBinaryAuthorization(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredBinaryAuthorization: out.BinaryAuthorization,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newDatabaseEncryptionUpdateFn returns a function that updates the DatabaseEncryption of a cluster.
func newDatabaseEncryptionUpdateFn(in *v1beta1.DatabaseEncryption) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateDatabaseEncryption(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredDatabaseEncryption: out.DatabaseEncryption,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newLegacyAbacUpdateFn returns a function that updates the LegacyAbac of a cluster.
func newLegacyAbacUpdateFn(in *v1beta1.LegacyAbac) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateLegacyAbac(in, out)
		update := &container.SetLegacyAbacRequest{
			Enabled: out.LegacyAbac.Enabled,
		}
		return s.Projects.Locations.Clusters.SetLegacyAbac(name, update).Context(ctx).Do()
	}
}

// newLocationsUpdateFn returns a function that updates the Locations of a cluster.
func newLocationsUpdateFn(in []string) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredLocations: in,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newLoggingServiceUpdateFn returns a function that updates the LoggingService of a cluster.
func newLoggingServiceUpdateFn(in *string) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredLoggingService: gcp.StringValue(in),
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newMaintenancePolicyUpdateFn returns a function that updates the MaintenancePolicy of a cluster.
func newMaintenancePolicyUpdateFn(in *v1beta1.MaintenancePolicySpec) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateMaintenancePolicy(in, out)
		update := &container.SetMaintenancePolicyRequest{
			MaintenancePolicy: out.MaintenancePolicy,
		}
		return s.Projects.Locations.Clusters.SetMaintenancePolicy(name, update).Context(ctx).Do()
	}
}

// newMasterAuthorizedNetworksConfigUpdateFn returns a function that updates the MasterAuthorizedNetworksConfig of a cluster.
func newMasterAuthorizedNetworksConfigUpdateFn(in *v1beta1.MasterAuthorizedNetworksConfig) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateMasterAuthorizedNetworksConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredMasterAuthorizedNetworksConfig: out.MasterAuthorizedNetworksConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newMonitoringServiceUpdateFn returns a function that updates the MonitoringService of a cluster.
func newMonitoringServiceUpdateFn(in *string) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredMonitoringService: gcp.StringValue(in),
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newNetworkConfigUpdateFn returns a function that updates the NetworkConfig of a cluster.
func newNetworkConfigUpdateFn(in *v1beta1.NetworkConfigSpec) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateNetworkConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredIntraNodeVisibilityConfig: &container.IntraNodeVisibilityConfig{
					Enabled: out.NetworkConfig.EnableIntraNodeVisibility,
				},
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newNetworkPolicyUpdateFn returns a function that updates the NetworkPolicy of a cluster.
func newNetworkPolicyUpdateFn(in *v1beta1.NetworkPolicy) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateNetworkPolicy(in, out)
		update := &container.SetNetworkPolicyRequest{
			NetworkPolicy: out.NetworkPolicy,
		}
		return s.Projects.Locations.Clusters.SetNetworkPolicy(name, update).Context(ctx).Do()
	}
}

// newPodSecurityPolicyConfigUpdateFn returns a function that updates the PodSecurityPolicyConfig of a cluster.
func newPodSecurityPolicyConfigUpdateFn(in *v1beta1.PodSecurityPolicyConfig) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GeneratePodSecurityPolicyConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredPodSecurityPolicyConfig: out.PodSecurityPolicyConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newPrivateClusterConfigUpdateFn returns a function that updates the PrivateClusterConfig of a cluster.
func newPrivateClusterConfigUpdateFn(in *v1beta1.PrivateClusterConfigSpec) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GeneratePrivateClusterConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredPrivateClusterConfig: out.PrivateClusterConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newResourceLabelsUpdateFn returns a function that updates the ResourceLabels of a cluster.
func newResourceLabelsUpdateFn(in map[string]string) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.SetLabelsRequest{
			ResourceLabels: in,
		}
		return s.Projects.Locations.Clusters.SetResourceLabels(name, update).Context(ctx).Do()
	}
}

// newResourceUsageExportConfigUpdateFn returns a function that updates the ResourceUsageExportConfig of a cluster.
func newResourceUsageExportConfigUpdateFn(in *v1beta1.ResourceUsageExportConfig) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateResourceUsageExportConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredResourceUsageExportConfig: out.ResourceUsageExportConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newVerticalPodAutoscalingUpdateFn returns a function that updates the VerticalPodAutoscaling of a cluster.
func newVerticalPodAutoscalingUpdateFn(in *v1beta1.VerticalPodAutoscaling) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateVerticalPodAutoscaling(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredVerticalPodAutoscaling: out.VerticalPodAutoscaling,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newWorkloadIdentityConfigUpdateFn returns a function that updates the WorkloadIdentityConfig of a cluster.
func newWorkloadIdentityConfigUpdateFn(in *v1beta1.WorkloadIdentityConfig) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateWorkloadIdentityConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredWorkloadIdentityConfig: out.WorkloadIdentityConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// deleteBootstrapNodePoolFn returns a function to delete the bootstrap node pool.
func deleteBootstrapNodePoolFn() UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		return s.Projects.Locations.Clusters.NodePools.Delete(GetFullyQualifiedBNP(name)).Context(ctx).Do()
	}
}

// UpdateFn returns a function that updates a node pool.
type UpdateFn func(context.Context, *container.Service, string) (*container.Operation, error)

func noOpUpdate(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
	return nil, nil
}

// checkForBootstrapNodePool checks if the bootstrap node pool exists for the
// cluster.
func checkForBootstrapNodePool(c *container.Cluster) bool {
	for _, pool := range c.NodePools {
		if pool == nil || pool.Name != BootstrapNodePoolName {
			continue
		}
		return true
	}
	return false
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
// NOTE(hasheddan): This function is significantly above our cyclomatic
// complexity limit, but is necessary due to the fact that the GKE API only
// allows for update of one field at a time.
func IsUpToDate(name string, in *v1beta1.GKEClusterParameters, observed *container.Cluster) (bool, UpdateFn, error) { // nolint:gocyclo
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, noOpUpdate, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*container.Cluster)
	if !ok {
		return true, noOpUpdate, errors.New(errCheckUpToDate)
	}
	GenerateCluster(name, *in, desired)
	if checkForBootstrapNodePool(observed) {
		return false, deleteBootstrapNodePoolFn(), nil
	}
	if !cmp.Equal(desired.AddonsConfig, observed.AddonsConfig, cmpopts.EquateEmpty()) {
		return false, newAddonsConfigUpdateFn(in.AddonsConfig), nil
	}
	if !cmp.Equal(desired.Autoscaling, observed.Autoscaling, cmpopts.EquateEmpty()) {
		return false, newAutoscalingUpdateFn(in.Autoscaling), nil
	}
	if !cmp.Equal(desired.BinaryAuthorization, observed.BinaryAuthorization, cmpopts.EquateEmpty()) {
		return false, newBinaryAuthorizationUpdateFn(in.BinaryAuthorization), nil
	}
	if !cmp.Equal(desired.DatabaseEncryption, observed.DatabaseEncryption, cmpopts.EquateEmpty()) {
		return false, newDatabaseEncryptionUpdateFn(in.DatabaseEncryption), nil
	}
	if !cmp.Equal(desired.LegacyAbac, observed.LegacyAbac, cmpopts.EquateEmpty()) {
		return false, newLegacyAbacUpdateFn(in.LegacyAbac), nil
	}
	if !cmp.Equal(desired.Locations, observed.Locations, cmpopts.EquateEmpty()) {
		return false, newLocationsUpdateFn(in.Locations), nil
	}
	if !cmp.Equal(desired.LoggingService, observed.LoggingService, cmpopts.EquateEmpty()) {
		return false, newLoggingServiceUpdateFn(in.LoggingService), nil
	}
	if !cmp.Equal(desired.MaintenancePolicy, observed.MaintenancePolicy, cmpopts.EquateEmpty()) {
		return false, newMaintenancePolicyUpdateFn(in.MaintenancePolicy), nil
	}
	if !cmp.Equal(desired.MasterAuthorizedNetworksConfig, observed.MasterAuthorizedNetworksConfig, cmpopts.EquateEmpty()) {
		return false, newMasterAuthorizedNetworksConfigUpdateFn(in.MasterAuthorizedNetworksConfig), nil
	}
	if !cmp.Equal(desired.MonitoringService, observed.MonitoringService, cmpopts.EquateEmpty()) {
		return false, newMonitoringServiceUpdateFn(in.MonitoringService), nil
	}
	if !cmp.Equal(desired.NetworkConfig, observed.NetworkConfig, cmpopts.EquateEmpty()) {
		return false, newNetworkConfigUpdateFn(in.NetworkConfig), nil
	}
	if !cmp.Equal(desired.NetworkPolicy, observed.NetworkPolicy, cmpopts.EquateEmpty()) {
		return false, newNetworkPolicyUpdateFn(in.NetworkPolicy), nil
	}
	if !cmp.Equal(desired.PodSecurityPolicyConfig, observed.PodSecurityPolicyConfig, cmpopts.EquateEmpty()) {
		return false, newPodSecurityPolicyConfigUpdateFn(in.PodSecurityPolicyConfig), nil
	}
	if !cmp.Equal(desired.PrivateClusterConfig, observed.PrivateClusterConfig, cmpopts.EquateEmpty()) {
		return false, newPrivateClusterConfigUpdateFn(in.PrivateClusterConfig), nil
	}
	if !cmp.Equal(desired.ResourceLabels, observed.ResourceLabels, cmpopts.EquateEmpty()) {
		return false, newResourceLabelsUpdateFn(in.ResourceLabels), nil
	}
	if !cmp.Equal(desired.ResourceUsageExportConfig, observed.ResourceUsageExportConfig, cmpopts.EquateEmpty()) {
		return false, newResourceUsageExportConfigUpdateFn(in.ResourceUsageExportConfig), nil
	}
	if !cmp.Equal(desired.VerticalPodAutoscaling, observed.VerticalPodAutoscaling, cmpopts.EquateEmpty()) {
		return false, newVerticalPodAutoscalingUpdateFn(in.VerticalPodAutoscaling), nil
	}
	if !cmp.Equal(desired.WorkloadIdentityConfig, observed.WorkloadIdentityConfig, cmpopts.EquateEmpty()) {
		return false, newWorkloadIdentityConfigUpdateFn(in.WorkloadIdentityConfig), nil
	}
	return true, noOpUpdate, nil
}

// GetFullyQualifiedParent builds the fully qualified name of the cluster
// parent.
func GetFullyQualifiedParent(project string, p v1beta1.GKEClusterParameters) string {
	return fmt.Sprintf(ParentFormat, project, p.Location)
}

// GetFullyQualifiedName builds the fully qualified name of the cluster.
func GetFullyQualifiedName(project string, p v1beta1.GKEClusterParameters, name string) string {
	return fmt.Sprintf(ClusterNameFormat, project, p.Location, name)
}

// GetFullyQualifiedBNP build the fully qualified name of the bootstrap node
// pool.
func GetFullyQualifiedBNP(clusterName string) string {
	return fmt.Sprintf(BNPNameFormat, clusterName, BootstrapNodePoolName)
}

// GenerateClientConfig generates a clientcmdapi.Config that can be used by any
// kubernetes client.
func GenerateClientConfig(cluster *container.Cluster) (clientcmdapi.Config, error) {
	if cluster.MasterAuth == nil {
		return clientcmdapi.Config{}, errors.New(errNoSecretInfo)
	}
	c := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			cluster.Name: {
				Server: fmt.Sprintf("https://%s", cluster.Endpoint),
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			cluster.Name: {
				Cluster:  cluster.Name,
				AuthInfo: cluster.Name,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			cluster.Name: {
				Username: cluster.MasterAuth.Username,
				Password: cluster.MasterAuth.Password,
			},
		},
		CurrentContext: cluster.Name,
	}

	val, err := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.Clusters[cluster.Name].CertificateAuthorityData = val

	val, err = base64.StdEncoding.DecodeString(cluster.MasterAuth.ClientCertificate)
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.AuthInfos[cluster.Name].ClientCertificateData = val

	val, err = base64.StdEncoding.DecodeString(cluster.MasterAuth.ClientKey)
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.AuthInfos[cluster.Name].ClientKeyData = val

	return c, nil
}
