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

package container

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-cmp/cmp"
	gke "google.golang.org/api/container/v1beta1"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	container "google.golang.org/api/container/v1beta1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	// BootstrapNodePoolName is the name of the node pool that is used to
	// boostrap GKE cluster creation.
	BootstrapNodePoolName = "crossplane-bootstrap"
)

// GenerateNodePoolForCreate inserts the default node pool into
// *container.Cluster so that it can be provisioned successfully.
func GenerateNodePoolForCreate(in *container.Cluster) {
	in.NodePools = []*container.NodePool{
		&container.NodePool{
			Name:             BootstrapNodePoolName,
			InitialNodeCount: 0,
		},
	}
}

// GenerateCluster generates *container.Cluster instance from GKEClusterParameters.
func GenerateCluster(in v1beta1.GKEClusterParameters) *container.Cluster { // nolint:gocyclo
	cluster := &container.Cluster{
		ClusterIpv4Cidr:       gcp.StringValue(in.ClusterIpv4Cidr),
		Description:           gcp.StringValue(in.Description),
		EnableKubernetesAlpha: gcp.BoolValue(in.EnableKubernetesAlpha),
		EnableTpu:             gcp.BoolValue(in.EnableTpu),
		InitialClusterVersion: gcp.StringValue(in.InitialClusterVersion),
		LabelFingerprint:      gcp.StringValue(in.LabelFingerprint),
		Locations:             in.Locations,
		LoggingService:        gcp.StringValue(in.LoggingService),
		MonitoringService:     gcp.StringValue(in.MonitoringService),
		Name:                  in.Name,
		Network:               gcp.StringValue(in.Network),
		ResourceLabels:        in.ResourceLabels,
		Subnetwork:            gcp.StringValue(in.Subnetwork),
	}

	GenerateAddonsConfig(in.AddonsConfig, cluster.AddonsConfig)
	GenerateAuthenticatorGroupsConfig(in.AuthenticatorGroupsConfig, cluster.AuthenticatorGroupsConfig)
	GenerateAutoscaling(in.Autoscaling, cluster.Autoscaling)
	GenerateBinaryAuthorization(in.BinaryAuthorization, cluster.BinaryAuthorization)
	GenerateDatabaseEncryption(in.DatabaseEncryption, cluster.DatabaseEncryption)
	GenerateDefaultMaxPodsConstraint(in.DefaultMaxPodsConstraint, cluster.DefaultMaxPodsConstraint)
	GenerateIpAllocationPolicy(in.IpAllocationPolicy, cluster.IpAllocationPolicy)
	GenerateLegacyAbac(in.LegacyAbac, cluster.LegacyAbac)
	GenerateMaintenancePolicy(in.MaintenancePolicy, cluster.MaintenancePolicy)
	GenerateMasterAuth(in.MasterAuth, cluster.MasterAuth)
	GenerateMasterAuthorizedNetworksConfig(in.MasterAuthorizedNetworksConfig, cluster.MasterAuthorizedNetworksConfig)
	GenerateNetworkConfig(in.NetworkConfig, cluster.NetworkConfig)
	GenerateNetworkPolicy(in.NetworkPolicy, cluster.NetworkPolicy)
	GeneratePodSecurityPolicyConfig(in.PodSecurityPolicyConfig, cluster.PodSecurityPolicyConfig)
	GeneratePrivateClusterConfig(in.PrivateClusterConfig, cluster.PrivateClusterConfig)
	GenerateResourceUsageExportConfig(in.ResourceUsageExportConfig, cluster.ResourceUsageExportConfig)
	GenerateTierSettings(in.TierSettings, cluster.TierSettings)
	GenerateVerticalPodAutoscaling(in.VerticalPodAutoscaling, cluster.VerticalPodAutoscaling)
	GenerateWorkloadIdentityConfig(in.WorkloadIdentityConfig, cluster.WorkloadIdentityConfig)

	return cluster
}

// GenerateAddonsConfig generates *container.AddonsConfig from *AddonsConfig.
func GenerateAddonsConfig(in *v1beta1.AddonsConfig, out *container.AddonsConfig) {
	if in != nil {
		out = &container.AddonsConfig{}

		if in.CloudRunConfig != nil {
			out.CloudRunConfig = &container.CloudRunConfig{
				Disabled: in.CloudRunConfig.Disabled,
			}
		}
		if in.HorizontalPodAutoscaling != nil {
			out.HorizontalPodAutoscaling = &container.HorizontalPodAutoscaling{
				Disabled: in.HorizontalPodAutoscaling.Disabled,
			}
		}
		if in.HttpLoadBalancing != nil {
			out.HttpLoadBalancing = &container.HttpLoadBalancing{
				Disabled: in.HttpLoadBalancing.Disabled,
			}
		}
		if in.IstioConfig != nil {
			out.IstioConfig = &container.IstioConfig{
				Auth:     gcp.StringValue(in.IstioConfig.Auth),
				Disabled: gcp.BoolValue(in.IstioConfig.Disabled),
			}
		}
		if in.KubernetesDashboard != nil {
			out.KubernetesDashboard = &container.KubernetesDashboard{
				Disabled: in.KubernetesDashboard.Disabled,
			}
		}
		if in.NetworkPolicyConfig != nil {
			out.NetworkPolicyConfig = &container.NetworkPolicyConfig{
				Disabled: in.NetworkPolicyConfig.Disabled,
			}
		}
	}
}

// GenerateAuthenticatorGroupsConfig generates *container.AuthenticatorGroupsConfig from *AuthenticatorGroupsConfig.
func GenerateAuthenticatorGroupsConfig(in *v1beta1.AuthenticatorGroupsConfig, out *container.AuthenticatorGroupsConfig) {
	if in != nil {
		out = &container.AuthenticatorGroupsConfig{
			Enabled:       gcp.BoolValue(in.Enabled),
			SecurityGroup: gcp.StringValue(in.SecurityGroup),
		}
	}
}

// GenerateAutoscaling generates *container.ClusterAutoscaling from *ClusterAutoscaling.
func GenerateAutoscaling(in *v1beta1.ClusterAutoscaling, out *container.ClusterAutoscaling) {
	if in != nil {
		out = &container.ClusterAutoscaling{
			AutoprovisioningLocations:  in.AutoprovisioningLocations,
			EnableNodeAutoprovisioning: gcp.BoolValue(in.EnableNodeAutoprovisioning),
		}

		if in.AutoprovisioningNodePoolDefaults != nil {
			out.AutoprovisioningNodePoolDefaults = &container.AutoprovisioningNodePoolDefaults{
				OauthScopes:    in.AutoprovisioningNodePoolDefaults.OauthScopes,
				ServiceAccount: gcp.StringValue(in.AutoprovisioningNodePoolDefaults.ServiceAccount),
			}
		}

		for _, limit := range in.ResourceLimits {
			if limit != nil {
				out.ResourceLimits = append(out.ResourceLimits, &container.ResourceLimit{
					Maximum:      gcp.Int64Value(limit.Maximum),
					Minimum:      gcp.Int64Value(limit.Minimum),
					ResourceType: gcp.StringValue(limit.ResourceType),
				})
			}
		}
	}
}

// GenerateBinaryAuthorization generates *container.BinaryAuthorization from *BinaryAuthorization.
func GenerateBinaryAuthorization(in *v1beta1.BinaryAuthorization, out *container.BinaryAuthorization) {
	if in != nil {
		out = &container.BinaryAuthorization{
			Enabled: in.Enabled,
		}
	}
}

// GenerateDatabaseEncryption generates *container.DatabaseEncryption from *DatabaseEncryption.
func GenerateDatabaseEncryption(in *v1beta1.DatabaseEncryption, out *container.DatabaseEncryption) {
	if in != nil {
		out = &container.DatabaseEncryption{
			KeyName: gcp.StringValue(in.KeyName),
			State:   gcp.StringValue(in.State),
		}
	}
}

// GenerateDefaultMaxPodsConstraint generates *container.MaxPodsConstraint from *DefaultMaxPodsConstraint.
func GenerateDefaultMaxPodsConstraint(in *v1beta1.MaxPodsConstraint, out *container.MaxPodsConstraint) {
	if in != nil {
		out = &container.MaxPodsConstraint{
			MaxPodsPerNode: in.MaxPodsPerNode,
		}
	}
}

// GenerateIpAllocationPolicy generates *container.MaxPodsConstraint from *IpAllocationPolicy.
func GenerateIpAllocationPolicy(in *v1beta1.IPAllocationPolicy, out *container.IPAllocationPolicy) {
	if in != nil {
		out = &container.IPAllocationPolicy{
			AllowRouteOverlap:          gcp.BoolValue(in.AllowRouteOverlap),
			ClusterIpv4CidrBlock:       gcp.StringValue(in.ClusterIpv4CidrBlock),
			ClusterSecondaryRangeName:  gcp.StringValue(in.ClusterSecondaryRangeName),
			CreateSubnetwork:           gcp.BoolValue(in.CreateSubnetwork),
			NodeIpv4CidrBlock:          gcp.StringValue(in.NodeIpv4CidrBlock),
			ServicesIpv4CidrBlock:      gcp.StringValue(in.ServicesIpv4CidrBlock),
			ServicesSecondaryRangeName: gcp.StringValue(in.DeepCopy().ServicesSecondaryRangeName),
			SubnetworkName:             gcp.StringValue(in.SubnetworkName),
			TpuIpv4CidrBlock:           gcp.StringValue(in.TpuIpv4CidrBlock),
			UseIpAliases:               gcp.BoolValue(in.UseIpAliases),
		}
	}
}

// GenerateLegacyAbac generates *container.LegacyAbac from *LegacyAbac.
func GenerateLegacyAbac(in *v1beta1.LegacyAbac, out *container.LegacyAbac) {
	if in != nil {
		out = &container.LegacyAbac{
			Enabled: in.Enabled,
		}
	}
}

// GenerateMaintenancePolicy generates *container.MaintenancePolicy from *MaintenancePolicy.
func GenerateMaintenancePolicy(in *v1beta1.MaintenancePolicy, out *container.MaintenancePolicy) {
	if in != nil {
		out = &container.MaintenancePolicy{
			Window: &container.MaintenanceWindow{
				DailyMaintenanceWindow: &container.DailyMaintenanceWindow{
					StartTime: in.Window.DailyMaintenanceWindow.StartTime,
				},
			},
		}
	}
}

// GenerateMasterAuth generates *container.MasterAuth from *MasterAuth.
func GenerateMasterAuth(in *v1beta1.MasterAuth, out *container.MasterAuth) {
	if in != nil {
		out = &container.MasterAuth{
			Password: gcp.StringValue(in.Password),
			Username: gcp.StringValue(in.Username),
		}

		if in.ClientCertificateConfig != nil {
			out.ClientCertificateConfig = &container.ClientCertificateConfig{
				IssueClientCertificate: in.ClientCertificateConfig.IssueClientCertificate,
			}
		}
	}
}

// GenerateMasterAuthorizedNetworksConfig generates *container.MasterAuthorizedNetworksConfig from *MasterAuthorizedNetworksConfig.
func GenerateMasterAuthorizedNetworksConfig(in *v1beta1.MasterAuthorizedNetworksConfig, out *container.MasterAuthorizedNetworksConfig) {
	if in != nil {
		out = &container.MasterAuthorizedNetworksConfig{
			Enabled: gcp.BoolValue(in.Enabled),
		}

		for _, cidr := range in.CidrBlocks {
			if cidr != nil {
				out.CidrBlocks = append(out.CidrBlocks, &container.CidrBlock{
					CidrBlock:   cidr.CidrBlock,
					DisplayName: gcp.StringValue(cidr.DisplayName),
				})
			}
		}
	}
}

// GenerateNetworkConfig generates *container.NetworkConfig from *NetworkConfig.
func GenerateNetworkConfig(in *v1beta1.NetworkConfig, out *container.NetworkConfig) {
	if in != nil {
		out = &container.NetworkConfig{
			EnableIntraNodeVisibility: in.EnableIntraNodeVisibility,
		}
	}
}

// GenerateNetworkPolicy generates *container.NetworkPolicy from *NetworkPolicy.
func GenerateNetworkPolicy(in *v1beta1.NetworkPolicy, out *container.NetworkPolicy) {
	if in != nil {
		out = &container.NetworkPolicy{
			Enabled:  gcp.BoolValue(in.Enabled),
			Provider: gcp.StringValue(in.Provider),
		}
	}
}

// GeneratePodSecurityPolicyConfig generates *container.PodSecurityPolicyConfig from *PodSecurityPolicyConfig.
func GeneratePodSecurityPolicyConfig(in *v1beta1.PodSecurityPolicyConfig, out *container.PodSecurityPolicyConfig) {
	if in != nil {
		out = &container.PodSecurityPolicyConfig{
			Enabled: in.Enabled,
		}
	}
}

// GeneratePrivateClusterConfig generates *container.PrivateClusterConfig from *PrivateClusterConfig.
func GeneratePrivateClusterConfig(in *v1beta1.PrivateClusterConfig, out *container.PrivateClusterConfig) {
	if in != nil {
		out = &container.PrivateClusterConfig{
			EnablePeeringRouteSharing: gcp.BoolValue(in.EnablePeeringRouteSharing),
			EnablePrivateEndpoint:     gcp.BoolValue(in.EnablePrivateEndpoint),
			EnablePrivateNodes:        gcp.BoolValue(in.EnablePrivateNodes),
			MasterIpv4CidrBlock:       gcp.StringValue(in.MasterIpv4CidrBlock),
		}
	}
}

// GenerateResourceUsageExportConfig generates *container.ResourceUsageExportConfig from *ResourceUsageExportConfig.
func GenerateResourceUsageExportConfig(in *v1beta1.ResourceUsageExportConfig, out *container.ResourceUsageExportConfig) {
	if in != nil {
		out = &container.ResourceUsageExportConfig{
			EnableNetworkEgressMetering: gcp.BoolValue(in.EnableNetworkEgressMetering),
		}

		if in.BigqueryDestination != nil {
			out.BigqueryDestination = &container.BigQueryDestination{
				DatasetId: in.BigqueryDestination.DatasetId,
			}
		}

		if in.ConsumptionMeteringConfig != nil {
			out.ConsumptionMeteringConfig = &container.ConsumptionMeteringConfig{
				Enabled: in.ConsumptionMeteringConfig.Enabled,
			}
		}
	}
}

// GenerateTierSettings generates *container.TierSettings from *TierSettings.
func GenerateTierSettings(in *v1beta1.TierSettings, out *container.TierSettings) {
	if in != nil {
		out = &container.TierSettings{
			Tier: in.Tier,
		}
	}
}

// GenerateVerticalPodAutoscaling generates *container.VerticalPodAutoscaling from *VerticalPodAutoscaling.
func GenerateVerticalPodAutoscaling(in *v1beta1.VerticalPodAutoscaling, out *container.VerticalPodAutoscaling) {
	if in != nil {
		out = &container.VerticalPodAutoscaling{
			Enabled: in.Enabled,
		}
	}
}

// GenerateWorkloadIdentityConfig generates *container.WorkloadIdentityConfig from *WorkloadIdentityConfig.
func GenerateWorkloadIdentityConfig(in *v1beta1.WorkloadIdentityConfig, out *container.WorkloadIdentityConfig) {
	if in != nil {
		out = &container.WorkloadIdentityConfig{
			IdentityNamespace: in.IdentityNamespace,
		}
	}
}

// GenerateObservation produces GKEClusterObservation object from *sqladmin.DatabaseInstance object.
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

	for _, condition := range in.Conditions {
		if condition != nil {
			o.Conditions = append(o.Conditions, &v1beta1.StatusCondition{
				Code:    condition.Code,
				Message: condition.Message,
			})
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
				Disabled: in.AddonsConfig.CloudRunConfig.Disabled,
			}
		}
		if spec.AddonsConfig.HorizontalPodAutoscaling == nil && in.AddonsConfig.HorizontalPodAutoscaling != nil {
			spec.AddonsConfig.HorizontalPodAutoscaling = &v1beta1.HorizontalPodAutoscaling{
				Disabled: in.AddonsConfig.HorizontalPodAutoscaling.Disabled,
			}
		}
		if spec.AddonsConfig.HttpLoadBalancing == nil && in.AddonsConfig.HttpLoadBalancing != nil {
			spec.AddonsConfig.HttpLoadBalancing = &v1beta1.HttpLoadBalancing{
				Disabled: in.AddonsConfig.HttpLoadBalancing.Disabled,
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
				Disabled: in.AddonsConfig.KubernetesDashboard.Disabled,
			}
		}
		if spec.AddonsConfig.NetworkPolicyConfig == nil && in.AddonsConfig.NetworkPolicyConfig != nil {
			spec.AddonsConfig.NetworkPolicyConfig = &v1beta1.NetworkPolicyConfig{
				Disabled: in.AddonsConfig.NetworkPolicyConfig.Disabled,
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

	if spec.Description == nil {
		spec.Description = &in.Description
	}

	spec.EnableKubernetesAlpha = gcp.LateInitializeBool(spec.EnableKubernetesAlpha, in.EnableKubernetesAlpha)
	spec.EnableTpu = gcp.LateInitializeBool(spec.EnableTpu, in.EnableTpu)
	spec.InitialClusterVersion = gcp.LateInitializeString(spec.InitialClusterVersion, in.InitialClusterVersion)

	if in.IpAllocationPolicy != nil {
		if spec.IpAllocationPolicy == nil {
			spec.IpAllocationPolicy = &v1beta1.IPAllocationPolicy{}
		}
		spec.IpAllocationPolicy.AllowRouteOverlap = gcp.LateInitializeBool(spec.IpAllocationPolicy.AllowRouteOverlap, in.IpAllocationPolicy.AllowRouteOverlap)
		spec.IpAllocationPolicy.ClusterIpv4CidrBlock = gcp.LateInitializeString(spec.IpAllocationPolicy.ClusterIpv4CidrBlock, in.IpAllocationPolicy.ClusterIpv4Cidr)
		spec.IpAllocationPolicy.ClusterSecondaryRangeName = gcp.LateInitializeString(spec.IpAllocationPolicy.ClusterSecondaryRangeName, in.IpAllocationPolicy.ClusterSecondaryRangeName)
		spec.IpAllocationPolicy.CreateSubnetwork = gcp.LateInitializeBool(spec.IpAllocationPolicy.CreateSubnetwork, in.IpAllocationPolicy.CreateSubnetwork)
		spec.IpAllocationPolicy.NodeIpv4CidrBlock = gcp.LateInitializeString(spec.IpAllocationPolicy.NodeIpv4CidrBlock, in.IpAllocationPolicy.NodeIpv4CidrBlock)
		spec.IpAllocationPolicy.ServicesIpv4CidrBlock = gcp.LateInitializeString(spec.IpAllocationPolicy.ServicesIpv4CidrBlock, in.IpAllocationPolicy.ServicesIpv4CidrBlock)
		spec.IpAllocationPolicy.SubnetworkName = gcp.LateInitializeString(spec.IpAllocationPolicy.SubnetworkName, in.IpAllocationPolicy.SubnetworkName)
		spec.IpAllocationPolicy.TpuIpv4CidrBlock = gcp.LateInitializeString(spec.IpAllocationPolicy.TpuIpv4CidrBlock, in.IpAllocationPolicy.TpuIpv4CidrBlock)
		spec.IpAllocationPolicy.UseIpAliases = gcp.LateInitializeBool(spec.IpAllocationPolicy.UseIpAliases, in.IpAllocationPolicy.UseIpAliases)
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
				spec.MaintenancePolicy = &v1beta1.MaintenancePolicy{
					Window: v1beta1.MaintenanceWindow{
						DailyMaintenanceWindow: v1beta1.DailyMaintenanceWindow{
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
		if spec.MasterAuth.ClientCertificateConfig == nil && in.MasterAuth.ClientCertificateConfig != nil {
			spec.MasterAuth.ClientCertificateConfig = &v1beta1.ClientCertificateConfig{
				IssueClientCertificate: in.MasterAuth.ClientCertificateConfig.IssueClientCertificate,
			}
		}
		spec.MasterAuth.Password = gcp.LateInitializeString(spec.MasterAuth.Password, in.MasterAuth.Password)
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
		spec.NetworkConfig = &v1beta1.NetworkConfig{
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
			spec.PrivateClusterConfig = &v1beta1.PrivateClusterConfig{}
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
				DatasetId: in.ResourceUsageExportConfig.BigqueryDestination.DatasetId,
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

// UpdateFn returns a function that updates a cluster.
type UpdateFn func(gke.Service, context.Context, string) (*container.Operation, error)

// NewAddonsConfigUpdate returns a function that updates the AddonsConfig of a cluster.
func NewAddonsConfigUpdate(in *v1beta1.AddonsConfig) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.AddonsConfig{}
		GenerateAddonsConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredAddonsConfig: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewAutoscalingUpdate returns a function that updates the Autoscaling of a cluster.
func NewAutoscalingUpdate(in *v1beta1.ClusterAutoscaling) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.ClusterAutoscaling{}
		GenerateAutoscaling(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredClusterAutoscaling: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewBinaryAuthorizationUpdate returns a function that updates the BinaryAuthorization of a cluster.
func NewBinaryAuthorizationUpdate(in *v1beta1.BinaryAuthorization) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.BinaryAuthorization{}
		GenerateBinaryAuthorization(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredBinaryAuthorization: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewDatabaseEncryptionUpdate returns a function that updates the DatabaseEncryption of a cluster.
func NewDatabaseEncryptionUpdate(in *v1beta1.DatabaseEncryption) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.DatabaseEncryption{}
		GenerateDatabaseEncryption(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredDatabaseEncryption: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewLegacyAbacUpdate returns a function that updates the LegacyAbac of a cluster.
func NewLegacyAbacUpdate(in *v1beta1.LegacyAbac) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.LegacyAbac{}
		GenerateLegacyAbac(in, out)
		update := &container.SetLegacyAbacRequest{
			Enabled: out.Enabled,
		}
		return s.Projects.Locations.Clusters.SetLegacyAbac(name, update).Context(ctx).Do()
	}
}

// NewLocationsUpdate returns a function that updates the Locations of a cluster.
func NewLocationsUpdate(in []string) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredLocations: in,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewLoggingServiceUpdate returns a function that updates the LoggingService of a cluster.
func NewLoggingServiceUpdate(in *string) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredLoggingService: gcp.StringValue(in),
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewMaintenancePolicyUpdate returns a function that updates the MaintenancePolicy of a cluster.
func NewMaintenancePolicyUpdate(in *v1beta1.MaintenancePolicy) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.MaintenancePolicy{}
		GenerateMaintenancePolicy(in, out)
		update := &container.SetMaintenancePolicyRequest{
			MaintenancePolicy: out,
		}
		return s.Projects.Locations.Clusters.SetMaintenancePolicy(name, update).Context(ctx).Do()
	}
}

// NewMasterAuthUpdate returns a function that updates the MasterAuth of a cluster.
func NewMasterAuthUpdate(in *v1beta1.MasterAuth) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.MasterAuth{}
		GenerateMasterAuth(in, out)
		update := &container.SetMasterAuthRequest{
			// TODO(hasheddan): need to set Action here?
			Update: out,
		}
		return s.Projects.Locations.Clusters.SetMasterAuth(name, update).Context(ctx).Do()
	}
}

// NewMasterAuthorizedNetworksConfigUpdate returns a function that updates the MasterAuthorizedNetworksConfig of a cluster.
func NewMasterAuthorizedNetworksConfigUpdate(in *v1beta1.MasterAuthorizedNetworksConfig) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.MasterAuthorizedNetworksConfig{}
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredMasterAuthorizedNetworksConfig: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewMonitoringServiceUpdate returns a function that updates the MonitoringService of a cluster.
func NewMonitoringServiceUpdate(in *string) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredMonitoringService: gcp.StringValue(in),
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewNetworkConfigUpdate returns a function that updates the NetworkConfig of a cluster.
func NewNetworkConfigUpdate(in *v1beta1.NetworkConfig) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.NetworkConfig{}
		GenerateNetworkConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredIntraNodeVisibilityConfig: &container.IntraNodeVisibilityConfig{
					Enabled: out.EnableIntraNodeVisibility,
				},
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewNetworkPolicyUpdate returns a function that updates the NetworkPolicy of a cluster.
func NewNetworkPolicyUpdate(in *v1beta1.NetworkPolicy) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.NetworkPolicy{}
		GenerateNetworkPolicy(in, out)
		update := &container.SetNetworkPolicyRequest{
			NetworkPolicy: out,
		}
		return s.Projects.Locations.Clusters.SetNetworkPolicy(name, update).Context(ctx).Do()
	}
}

// NewPodSecurityPolicyConfigUpdate returns a function that updates the PodSecurityPolicyConfig of a cluster.
func NewPodSecurityPolicyConfigUpdate(in *v1beta1.PodSecurityPolicyConfig) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.PodSecurityPolicyConfig{}
		GeneratePodSecurityPolicyConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredPodSecurityPolicyConfig: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewPrivateClusterConfigUpdate returns a function that updates the PrivateClusterConfig of a cluster.
func NewPrivateClusterConfigUpdate(in *v1beta1.PrivateClusterConfig) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.PrivateClusterConfig{}
		GeneratePrivateClusterConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredPrivateClusterConfig: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewResourceLabelsUpdate returns a function that updates the ResourceLabels of a cluster.
func NewResourceLabelsUpdate(in map[string]string) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		update := &container.SetLabelsRequest{
			ResourceLabels: in,
		}
		return s.Projects.Locations.Clusters.SetResourceLabels(name, update).Context(ctx).Do()
	}
}

// NewResourceUsageExportConfigUpdate returns a function that updates the ResourceUsageExportConfig of a cluster.
func NewResourceUsageExportConfigUpdate(in *v1beta1.ResourceUsageExportConfig) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.ResourceUsageExportConfig{}
		GenerateResourceUsageExportConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredResourceUsageExportConfig: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewVerticalPodAutoscalingUpdate returns a function that updates the VerticalPodAutoscaling of a cluster.
func NewVerticalPodAutoscalingUpdate(in *v1beta1.VerticalPodAutoscaling) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.VerticalPodAutoscaling{}
		GenerateVerticalPodAutoscaling(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredVerticalPodAutoscaling: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// NewWorkloadIdentityConfigUpdate returns a function that updates the WorkloadIdentityConfig of a cluster.
func NewWorkloadIdentityConfigUpdate(in *v1beta1.WorkloadIdentityConfig) UpdateFn {
	return func(s gke.Service, ctx context.Context, name string) (*container.Operation, error) {
		out := &container.WorkloadIdentityConfig{}
		GenerateWorkloadIdentityConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredWorkloadIdentityConfig: out,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1beta1.GKEClusterParameters, currentState container.Cluster) (bool, UpdateFn) {
	currentParams := &v1beta1.GKEClusterParameters{}
	LateInitializeSpec(currentParams, currentState)
	// TODO(hasheddan): fix "fullname" params below
	if !cmp.Equal(in.AddonsConfig, currentParams.AddonsConfig) {
		return false, NewAddonsConfigUpdate(in.AddonsConfig)
	}
	if !cmp.Equal(in.Autoscaling, currentParams.Autoscaling) {
		return false, NewAutoscalingUpdate(in.Autoscaling)
	}
	if !cmp.Equal(in.BinaryAuthorization, currentParams.BinaryAuthorization) {
		return false, NewBinaryAuthorizationUpdate(in.BinaryAuthorization)
	}
	if !cmp.Equal(in.DatabaseEncryption, currentParams.DatabaseEncryption) {
		return false, NewDatabaseEncryptionUpdate(in.DatabaseEncryption)
	}
	if !cmp.Equal(in.LegacyAbac, currentParams.LegacyAbac) {
		return false, NewLegacyAbacUpdate(in.LegacyAbac)
	}
	if !cmp.Equal(in.Locations, currentParams.Locations) {
		return false, NewLocationsUpdate(in.Locations)
	}
	if !cmp.Equal(in.LoggingService, currentParams.LoggingService) {
		return false, NewLoggingServiceUpdate(in.LoggingService)
	}
	if !cmp.Equal(in.MaintenancePolicy, currentParams.MaintenancePolicy) {
		return false, NewMaintenancePolicyUpdate(in.MaintenancePolicy)
	}
	if !cmp.Equal(in.MasterAuth, currentParams.MasterAuth) {
		return false, NewMasterAuthUpdate(in.MasterAuth)
	}
	if !cmp.Equal(in.MasterAuthorizedNetworksConfig, currentParams.MasterAuthorizedNetworksConfig) {
		return false, NewMasterAuthorizedNetworksConfigUpdate(in.MasterAuthorizedNetworksConfig)
	}
	if !cmp.Equal(in.MonitoringService, currentParams.MonitoringService) {
		return false, NewMonitoringServiceUpdate(in.MonitoringService)
	}
	if !cmp.Equal(in.NetworkConfig, currentParams.NetworkConfig) {
		return false, NewNetworkConfigUpdate(in.NetworkConfig)
	}
	if !cmp.Equal(in.NetworkPolicy, currentParams.NetworkPolicy) {
		return false, NewNetworkPolicyUpdate(in.NetworkPolicy)
	}
	if !cmp.Equal(in.PodSecurityPolicyConfig, currentParams.PodSecurityPolicyConfig) {
		return false, NewPodSecurityPolicyConfigUpdate(in.PodSecurityPolicyConfig)
	}
	if !cmp.Equal(in.PrivateClusterConfig, currentParams.PrivateClusterConfig) {
		return false, NewPrivateClusterConfigUpdate(in.PrivateClusterConfig)
	}
	if !cmp.Equal(in.ResourceLabels, currentParams.ResourceLabels) {
		return false, NewResourceLabelsUpdate(in.ResourceLabels)
	}
	if !cmp.Equal(in.ResourceUsageExportConfig, currentParams.ResourceUsageExportConfig) {
		return false, NewResourceUsageExportConfigUpdate(in.ResourceUsageExportConfig)
	}
	if !cmp.Equal(in.VerticalPodAutoscaling, currentParams.VerticalPodAutoscaling) {
		return false, NewVerticalPodAutoscalingUpdate(in.VerticalPodAutoscaling)
	}
	if !cmp.Equal(in.WorkloadIdentityConfig, currentParams.WorkloadIdentityConfig) {
		return false, NewWorkloadIdentityConfigUpdate(in.WorkloadIdentityConfig)
	}
	return true, nil
}

// GenerateClientConfig generates a clientcmdapi.Config that can be used by any
// kubernetes client.
func GenerateClientConfig(cluster *container.Cluster) (clientcmdapi.Config, error) {
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
