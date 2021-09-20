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
	container "google.golang.org/api/container/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/crossplane/crossplane-runtime/pkg/errors"

	"github.com/crossplane/provider-gcp/apis/container/v1beta2"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
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

// GenerateCluster generates *container.Cluster instance from ClusterParameters.
func GenerateCluster(name string, in v1beta2.ClusterParameters, cluster *container.Cluster) { // nolint:gocyclo
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
	GenerateAutopilot(in.Autopilot, cluster)
	GenerateAuthenticatorGroupsConfig(in.AuthenticatorGroupsConfig, cluster)
	GenerateAutoscaling(in.Autoscaling, cluster)
	GenerateConfidentialNodes(in.ConfidentialNodes, cluster)
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
	GenerateNotificationConfig(in.NotificationConfig, cluster)
	GeneratePrivateClusterConfig(in.PrivateClusterConfig, cluster)
	GenerateReleaseChannel(in.ReleaseChannel, cluster)
	GenerateResourceUsageExportConfig(in.ResourceUsageExportConfig, cluster)
	GenerateVerticalPodAutoscaling(in.VerticalPodAutoscaling, cluster)
	GenerateWorkloadIdentityConfig(in.WorkloadIdentityConfig, cluster)
}

// GenerateAddonsConfig generates *container.AddonsConfig from *AddonsConfig.
func GenerateAddonsConfig(in *v1beta2.AddonsConfig, cluster *container.Cluster) { // nolint:gocyclo
	if in != nil {
		if cluster.AddonsConfig == nil {
			cluster.AddonsConfig = &container.AddonsConfig{}
		}
		if in.CloudRunConfig != nil {
			if cluster.AddonsConfig.CloudRunConfig == nil {
				cluster.AddonsConfig.CloudRunConfig = &container.CloudRunConfig{}
			}
			cluster.AddonsConfig.CloudRunConfig.Disabled = in.CloudRunConfig.Disabled
			cluster.AddonsConfig.CloudRunConfig.LoadBalancerType = gcp.StringValue(in.CloudRunConfig.LoadBalancerType)
			cluster.AddonsConfig.CloudRunConfig.ForceSendFields = []string{"Disabled", "LoadBalancerType"}
		}
		if in.ConfigConnectorConfig != nil {
			if cluster.AddonsConfig.ConfigConnectorConfig == nil {
				cluster.AddonsConfig.ConfigConnectorConfig = &container.ConfigConnectorConfig{}
			}
			cluster.AddonsConfig.ConfigConnectorConfig.Enabled = in.ConfigConnectorConfig.Enabled
			cluster.AddonsConfig.ConfigConnectorConfig.ForceSendFields = []string{"Enabled"}
		}
		if in.DNSCacheConfig != nil {
			if cluster.AddonsConfig.DnsCacheConfig == nil {
				cluster.AddonsConfig.DnsCacheConfig = &container.DnsCacheConfig{}
			}
			cluster.AddonsConfig.DnsCacheConfig.Enabled = in.DNSCacheConfig.Enabled
			cluster.AddonsConfig.DnsCacheConfig.ForceSendFields = []string{"Enabled"}
		}
		if in.GCEPersistentDiskCSIDriverConfig != nil {
			if cluster.AddonsConfig.GcePersistentDiskCsiDriverConfig == nil {
				cluster.AddonsConfig.GcePersistentDiskCsiDriverConfig = &container.GcePersistentDiskCsiDriverConfig{}
			}
			cluster.AddonsConfig.GcePersistentDiskCsiDriverConfig.Enabled = in.GCEPersistentDiskCSIDriverConfig.Enabled
			cluster.AddonsConfig.GcePersistentDiskCsiDriverConfig.ForceSendFields = []string{"Enabled"}
		}
		if in.HorizontalPodAutoscaling != nil {
			if cluster.AddonsConfig.HorizontalPodAutoscaling == nil {
				cluster.AddonsConfig.HorizontalPodAutoscaling = &container.HorizontalPodAutoscaling{}
			}
			cluster.AddonsConfig.HorizontalPodAutoscaling.Disabled = in.HorizontalPodAutoscaling.Disabled
			cluster.AddonsConfig.HorizontalPodAutoscaling.ForceSendFields = []string{"Disabled"}
		}
		if in.HTTPLoadBalancing != nil {
			if cluster.AddonsConfig.HttpLoadBalancing == nil {
				cluster.AddonsConfig.HttpLoadBalancing = &container.HttpLoadBalancing{}
			}
			cluster.AddonsConfig.HttpLoadBalancing.Disabled = in.HTTPLoadBalancing.Disabled
			cluster.AddonsConfig.HttpLoadBalancing.ForceSendFields = []string{"Disabled"}
		}
		if in.KubernetesDashboard != nil {
			if cluster.AddonsConfig.KubernetesDashboard == nil {
				cluster.AddonsConfig.KubernetesDashboard = &container.KubernetesDashboard{}
			}
			cluster.AddonsConfig.KubernetesDashboard.Disabled = in.KubernetesDashboard.Disabled
			cluster.AddonsConfig.KubernetesDashboard.ForceSendFields = []string{"Disabled"}
		}
		if in.NetworkPolicyConfig != nil {
			if cluster.AddonsConfig.NetworkPolicyConfig == nil {
				cluster.AddonsConfig.NetworkPolicyConfig = &container.NetworkPolicyConfig{}
			}
			cluster.AddonsConfig.NetworkPolicyConfig.Disabled = in.NetworkPolicyConfig.Disabled
			cluster.AddonsConfig.NetworkPolicyConfig.ForceSendFields = []string{"Disabled"}
		}
	}
}

// GenerateAutopilot generates *container.Autopilot from *Autopilot.
func GenerateAutopilot(in *v1beta2.Autopilot, cluster *container.Cluster) {
	if in != nil {
		if cluster.Autopilot == nil {
			cluster.Autopilot = &container.Autopilot{}
		}
		cluster.Autopilot.Enabled = in.Enabled
	}
}

// GenerateAuthenticatorGroupsConfig generates *container.AuthenticatorGroupsConfig from *AuthenticatorGroupsConfig.
func GenerateAuthenticatorGroupsConfig(in *v1beta2.AuthenticatorGroupsConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.AuthenticatorGroupsConfig == nil {
			cluster.AuthenticatorGroupsConfig = &container.AuthenticatorGroupsConfig{}
		}
		cluster.AuthenticatorGroupsConfig.Enabled = gcp.BoolValue(in.Enabled)
		cluster.AuthenticatorGroupsConfig.SecurityGroup = gcp.StringValue(in.SecurityGroup)
	}
}

// GenerateAutoscaling generates *container.ClusterAutoscaling from *ClusterAutoscaling.
func GenerateAutoscaling(in *v1beta2.ClusterAutoscaling, cluster *container.Cluster) { // nolint:gocyclo
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
			cluster.Autoscaling.AutoprovisioningNodePoolDefaults.BootDiskKmsKey = gcp.StringValue(in.AutoprovisioningNodePoolDefaults.BootDiskKMSKey)
			cluster.Autoscaling.AutoprovisioningNodePoolDefaults.DiskSizeGb = gcp.Int64Value(in.AutoprovisioningNodePoolDefaults.DiskSizeGb)
			cluster.Autoscaling.AutoprovisioningNodePoolDefaults.MinCpuPlatform = gcp.StringValue(in.AutoprovisioningNodePoolDefaults.MinCPUPlatform)
			cluster.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes = in.AutoprovisioningNodePoolDefaults.OauthScopes
			cluster.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount = gcp.StringValue(in.AutoprovisioningNodePoolDefaults.ServiceAccount)
			if in.AutoprovisioningNodePoolDefaults.Management != nil {
				if cluster.Autoscaling.AutoprovisioningNodePoolDefaults.Management == nil {
					cluster.Autoscaling.AutoprovisioningNodePoolDefaults.Management = &container.NodeManagement{}
				}
				cluster.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoRepair = gcp.BoolValue(in.AutoprovisioningNodePoolDefaults.Management.AutoRepair)
				cluster.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoUpgrade = gcp.BoolValue(in.AutoprovisioningNodePoolDefaults.Management.AutoUpgrade)
			}
			if in.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig != nil {
				if cluster.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig == nil {
					cluster.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig = &container.ShieldedInstanceConfig{}
				}
				cluster.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableIntegrityMonitoring = gcp.BoolValue(in.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableIntegrityMonitoring)
				cluster.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableSecureBoot = gcp.BoolValue(in.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableSecureBoot)
			}
			if in.AutoprovisioningNodePoolDefaults.UpgradeSettings != nil {
				if cluster.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings == nil {
					cluster.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings = &container.UpgradeSettings{}
				}
				cluster.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxSurge = gcp.Int64Value(in.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxSurge)
				cluster.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxUnavailable = gcp.Int64Value(in.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxUnavailable)
			}
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
func GenerateBinaryAuthorization(in *v1beta2.BinaryAuthorization, cluster *container.Cluster) {
	if in != nil {
		if cluster.BinaryAuthorization == nil {
			cluster.BinaryAuthorization = &container.BinaryAuthorization{}
		}
		cluster.BinaryAuthorization.Enabled = in.Enabled
	}
}

// GenerateConfidentialNodes generates *container.ConfidentialNodes from *ConfidentialNodes.
func GenerateConfidentialNodes(in *v1beta2.ConfidentialNodes, cluster *container.Cluster) {
	if in != nil {
		if cluster.ConfidentialNodes == nil {
			cluster.ConfidentialNodes = &container.ConfidentialNodes{}
		}
		cluster.ConfidentialNodes.Enabled = in.Enabled
	}
}

// GenerateDatabaseEncryption generates *container.DatabaseEncryption from *DatabaseEncryption.
func GenerateDatabaseEncryption(in *v1beta2.DatabaseEncryption, cluster *container.Cluster) {
	if in != nil {
		if cluster.DatabaseEncryption == nil {
			cluster.DatabaseEncryption = &container.DatabaseEncryption{}
		}
		cluster.DatabaseEncryption.KeyName = gcp.StringValue(in.KeyName)
		cluster.DatabaseEncryption.State = gcp.StringValue(in.State)
	}
}

// GenerateDefaultMaxPodsConstraint generates *container.MaxPodsConstraint from *DefaultMaxPodsConstraint.
func GenerateDefaultMaxPodsConstraint(in *v1beta2.MaxPodsConstraint, cluster *container.Cluster) {
	if in != nil {
		if cluster.DefaultMaxPodsConstraint == nil {
			cluster.DefaultMaxPodsConstraint = &container.MaxPodsConstraint{}
		}
		cluster.DefaultMaxPodsConstraint.MaxPodsPerNode = in.MaxPodsPerNode
	}
}

// GenerateIPAllocationPolicy generates *container.MaxPodsConstraint from *IpAllocationPolicy.
func GenerateIPAllocationPolicy(in *v1beta2.IPAllocationPolicy, cluster *container.Cluster) {
	if in != nil {
		if cluster.IpAllocationPolicy == nil {
			cluster.IpAllocationPolicy = &container.IPAllocationPolicy{}
		}
		cluster.IpAllocationPolicy.ClusterIpv4CidrBlock = gcp.StringValue(in.ClusterIpv4CidrBlock)
		cluster.IpAllocationPolicy.ClusterSecondaryRangeName = gcp.StringValue(in.ClusterSecondaryRangeName)
		cluster.IpAllocationPolicy.CreateSubnetwork = gcp.BoolValue(in.CreateSubnetwork)
		cluster.IpAllocationPolicy.NodeIpv4CidrBlock = gcp.StringValue(in.NodeIpv4CidrBlock)
		cluster.IpAllocationPolicy.ServicesIpv4CidrBlock = gcp.StringValue(in.ServicesIpv4CidrBlock)
		cluster.IpAllocationPolicy.ServicesSecondaryRangeName = gcp.StringValue(in.DeepCopy().ServicesSecondaryRangeName)
		cluster.IpAllocationPolicy.SubnetworkName = gcp.StringValue(in.SubnetworkName)
		cluster.IpAllocationPolicy.TpuIpv4CidrBlock = gcp.StringValue(in.TpuIpv4CidrBlock)
		cluster.IpAllocationPolicy.UseIpAliases = gcp.BoolValue(in.UseIPAliases)
		cluster.IpAllocationPolicy.UseRoutes = gcp.BoolValue(in.UseRoutes)
	}
}

// GenerateLegacyAbac generates *container.LegacyAbac from *LegacyAbac.
func GenerateLegacyAbac(in *v1beta2.LegacyAbac, cluster *container.Cluster) {
	if in != nil {
		if cluster.LegacyAbac == nil {
			cluster.LegacyAbac = &container.LegacyAbac{}
		}
		cluster.LegacyAbac.Enabled = in.Enabled
	}
}

// GenerateMaintenancePolicy generates *container.MaintenancePolicy from *MaintenancePolicy.
func GenerateMaintenancePolicy(in *v1beta2.MaintenancePolicySpec, cluster *container.Cluster) { // nolint:gocyclo
	if in != nil {
		if cluster.MaintenancePolicy == nil {
			cluster.MaintenancePolicy = &container.MaintenancePolicy{}
		}
		if cluster.MaintenancePolicy.Window == nil {
			cluster.MaintenancePolicy.Window = &container.MaintenanceWindow{}
		}
		if in.Window.DailyMaintenanceWindow != nil {
			if cluster.MaintenancePolicy.Window.DailyMaintenanceWindow == nil {
				cluster.MaintenancePolicy.Window.DailyMaintenanceWindow = &container.DailyMaintenanceWindow{}
			}
			cluster.MaintenancePolicy.Window.DailyMaintenanceWindow.StartTime = in.Window.DailyMaintenanceWindow.StartTime
		}
		if in.Window.MaintenanceExclusions != nil {
			cluster.MaintenancePolicy.Window.MaintenanceExclusions = make(map[string]container.TimeWindow, len(in.Window.MaintenanceExclusions))
			for k, v := range in.Window.MaintenanceExclusions {
				cluster.MaintenancePolicy.Window.MaintenanceExclusions[k] = container.TimeWindow{
					EndTime:   v.EndTime,
					StartTime: v.StartTime,
				}
			}
		}
		if in.Window.RecurringWindow != nil {
			if cluster.MaintenancePolicy.Window.RecurringWindow == nil {
				cluster.MaintenancePolicy.Window.RecurringWindow = &container.RecurringTimeWindow{}
			}
			if in.Window.RecurringWindow.Recurrence != nil {
				cluster.MaintenancePolicy.Window.RecurringWindow.Recurrence = *in.Window.RecurringWindow.Recurrence
			}
			if in.Window.RecurringWindow.Window != nil {
				cluster.MaintenancePolicy.Window.RecurringWindow.Window = &container.TimeWindow{
					EndTime:   in.Window.RecurringWindow.Window.EndTime,
					StartTime: in.Window.RecurringWindow.Window.StartTime,
				}
			}
		}
	}
}

// GenerateMasterAuth generates *container.MasterAuth from *MasterAuth.
func GenerateMasterAuth(in *v1beta2.MasterAuth, cluster *container.Cluster) {
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
func GenerateMasterAuthorizedNetworksConfig(in *v1beta2.MasterAuthorizedNetworksConfig, cluster *container.Cluster) {
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
func GenerateNetworkConfig(in *v1beta2.NetworkConfigSpec, cluster *container.Cluster) {
	if in != nil {
		if cluster.NetworkConfig == nil {
			cluster.NetworkConfig = &container.NetworkConfig{}
		}
		cluster.NetworkConfig.EnableIntraNodeVisibility = gcp.BoolValue(in.EnableIntraNodeVisibility)
		cluster.NetworkConfig.PrivateIpv6GoogleAccess = gcp.StringValue(in.PrivateIpv6GoogleAccess)
		cluster.NetworkConfig.DatapathProvider = gcp.StringValue(in.DatapathProvider)
		if in.DefaultSnatStatus != nil {
			if cluster.NetworkConfig.DefaultSnatStatus == nil {
				cluster.NetworkConfig.DefaultSnatStatus = &container.DefaultSnatStatus{}
			}
			cluster.NetworkConfig.DefaultSnatStatus.Disabled = in.DefaultSnatStatus.Disabled
		}
	}
}

// GenerateNetworkPolicy generates *container.NetworkPolicy from *NetworkPolicy.
func GenerateNetworkPolicy(in *v1beta2.NetworkPolicy, cluster *container.Cluster) {
	if in != nil {
		if cluster.NetworkPolicy == nil {
			cluster.NetworkPolicy = &container.NetworkPolicy{}
		}
		cluster.NetworkPolicy.Enabled = gcp.BoolValue(in.Enabled)
		cluster.NetworkPolicy.Provider = gcp.StringValue(in.Provider)
	}
}

// GenerateNotificationConfig generates *container.NotificationConfig from *NotificationConfig.
func GenerateNotificationConfig(in *v1beta2.NotificationConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.NotificationConfig == nil {
			cluster.NotificationConfig = &container.NotificationConfig{}
		}
		if cluster.NotificationConfig.Pubsub == nil {
			cluster.NotificationConfig.Pubsub = &container.PubSub{
				Enabled: in.Pubsub.Enabled,
				Topic:   in.Pubsub.Topic,
			}
		}
	}
}

// GeneratePrivateClusterConfig generates *container.PrivateClusterConfig from *PrivateClusterConfig.
func GeneratePrivateClusterConfig(in *v1beta2.PrivateClusterConfigSpec, cluster *container.Cluster) {
	if in != nil {
		if cluster.PrivateClusterConfig == nil {
			cluster.PrivateClusterConfig = &container.PrivateClusterConfig{}
		}
		cluster.PrivateClusterConfig.EnablePrivateEndpoint = gcp.BoolValue(in.EnablePrivateEndpoint)
		cluster.PrivateClusterConfig.EnablePrivateNodes = gcp.BoolValue(in.EnablePrivateNodes)
		cluster.PrivateClusterConfig.MasterIpv4CidrBlock = gcp.StringValue(in.MasterIpv4CidrBlock)
		if in.MasterGlobalAccessConfig != nil {
			cluster.PrivateClusterConfig.MasterGlobalAccessConfig = &container.PrivateClusterMasterGlobalAccessConfig{
				Enabled: in.MasterGlobalAccessConfig.Enabled,
			}
		}
	}
}

// GenerateReleaseChannel generates *container.ReleaseChannel from *ReleaseChannel.
func GenerateReleaseChannel(in *v1beta2.ReleaseChannel, cluster *container.Cluster) {
	if in != nil {
		if cluster.ReleaseChannel == nil {
			cluster.ReleaseChannel = &container.ReleaseChannel{}
		}
		cluster.ReleaseChannel.Channel = in.Channel
	}
}

// GenerateResourceUsageExportConfig generates *container.ResourceUsageExportConfig from *ResourceUsageExportConfig.
func GenerateResourceUsageExportConfig(in *v1beta2.ResourceUsageExportConfig, cluster *container.Cluster) {
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

// GenerateVerticalPodAutoscaling generates *container.VerticalPodAutoscaling from *VerticalPodAutoscaling.
func GenerateVerticalPodAutoscaling(in *v1beta2.VerticalPodAutoscaling, cluster *container.Cluster) {
	if in != nil {
		if cluster.VerticalPodAutoscaling == nil {
			cluster.VerticalPodAutoscaling = &container.VerticalPodAutoscaling{}
		}
		cluster.VerticalPodAutoscaling.Enabled = in.Enabled
	}
}

// GenerateWorkloadIdentityConfig generates *container.WorkloadIdentityConfig from *WorkloadIdentityConfig.
func GenerateWorkloadIdentityConfig(in *v1beta2.WorkloadIdentityConfig, cluster *container.Cluster) {
	if in != nil {
		if cluster.WorkloadIdentityConfig == nil {
			cluster.WorkloadIdentityConfig = &container.WorkloadIdentityConfig{}
		}
		cluster.WorkloadIdentityConfig.WorkloadPool = in.WorkloadPool
	}
}

// GenerateObservation produces ClusterObservation object from *container.Cluster object.
func GenerateObservation(in container.Cluster) v1beta2.ClusterObservation { // nolint:gocyclo
	o := v1beta2.ClusterObservation{
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
				o.MaintenancePolicy = &v1beta2.MaintenancePolicyStatus{
					Window: v1beta2.MaintenanceWindowStatus{
						DailyMaintenanceWindow: v1beta2.DailyMaintenanceWindowStatus{
							Duration: in.MaintenancePolicy.Window.DailyMaintenanceWindow.Duration,
						},
					},
				}
			}
		}
	}

	if in.NetworkConfig != nil {
		o.NetworkConfig = &v1beta2.NetworkConfigStatus{
			Network:    in.NetworkConfig.Network,
			Subnetwork: in.NetworkConfig.Subnetwork,
		}
	}

	if in.PrivateClusterConfig != nil {
		o.PrivateClusterConfig = &v1beta2.PrivateClusterConfigStatus{
			PrivateEndpoint: in.PrivateClusterConfig.PrivateEndpoint,
			PublicEndpoint:  in.PrivateClusterConfig.PublicEndpoint,
		}
	}

	for _, condition := range in.Conditions {
		if condition != nil {
			o.Conditions = append(o.Conditions, &v1beta2.StatusCondition{
				Code:    condition.Code,
				Message: condition.Message,
			})
		}
	}

	for _, nodePool := range in.NodePools {
		if nodePool != nil {
			np := &v1beta2.NodePoolClusterStatus{
				InstanceGroupUrls: nodePool.InstanceGroupUrls,
				Name:              nodePool.Name,
				PodIpv4CidrSize:   nodePool.PodIpv4CidrSize,
				SelfLink:          nodePool.SelfLink,
				Status:            nodePool.Status,
				StatusMessage:     nodePool.StatusMessage,
				Version:           nodePool.Version,
			}
			if nodePool.Autoscaling != nil {
				np.Autoscaling = &v1beta2.NodePoolAutoscalingClusterStatus{
					Autoprovisioned: nodePool.Autoscaling.Autoprovisioned,
					Enabled:         nodePool.Autoscaling.Enabled,
					MaxNodeCount:    nodePool.Autoscaling.MaxNodeCount,
					MinNodeCount:    nodePool.Autoscaling.MinNodeCount,
				}
			}
			if nodePool.Config != nil {
				np.Config = &v1beta2.NodeConfigClusterStatus{
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
						np.Config.Accelerators = append(np.Config.Accelerators, &v1beta2.AcceleratorConfigClusterStatus{
							AcceleratorCount: a.AcceleratorCount,
							AcceleratorType:  a.AcceleratorType,
						})
					}
				}
				if nodePool.Config.SandboxConfig != nil {
					np.Config.SandboxConfig = &v1beta2.SandboxConfigClusterStatus{
						Type: nodePool.Config.SandboxConfig.Type,
					}
				}
				if nodePool.Config.ShieldedInstanceConfig != nil {
					np.Config.ShieldedInstanceConfig = &v1beta2.ShieldedInstanceConfigClusterStatus{
						EnableIntegrityMonitoring: nodePool.Config.ShieldedInstanceConfig.EnableIntegrityMonitoring,
						EnableSecureBoot:          nodePool.Config.ShieldedInstanceConfig.EnableSecureBoot,
					}
				}
				for _, t := range nodePool.Config.Taints {
					if t != nil {
						np.Config.Taints = append(np.Config.Taints, &v1beta2.NodeTaintClusterStatus{
							Effect: t.Effect,
							Key:    t.Key,
							Value:  t.Value,
						})
					}
				}
			}
			for _, c := range nodePool.Conditions {
				if c != nil {
					np.Conditions = append(np.Conditions, &v1beta2.StatusCondition{
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
func LateInitializeSpec(spec *v1beta2.ClusterParameters, in container.Cluster) { // nolint:gocyclo
	if in.AddonsConfig != nil {
		if spec.AddonsConfig == nil {
			spec.AddonsConfig = &v1beta2.AddonsConfig{}
		}
		if in.AddonsConfig.CloudRunConfig != nil {
			spec.AddonsConfig.CloudRunConfig = &v1beta2.CloudRunConfig{
				Disabled: in.AddonsConfig.CloudRunConfig.Disabled,
			}
			spec.AddonsConfig.CloudRunConfig.LoadBalancerType = gcp.LateInitializeString(spec.AddonsConfig.CloudRunConfig.LoadBalancerType, in.AddonsConfig.CloudRunConfig.LoadBalancerType)
		}
		if spec.AddonsConfig.ConfigConnectorConfig == nil && in.AddonsConfig.ConfigConnectorConfig != nil {
			spec.AddonsConfig.ConfigConnectorConfig = &v1beta2.ConfigConnectorConfig{
				Enabled: in.AddonsConfig.ConfigConnectorConfig.Enabled,
			}
		}
		if spec.AddonsConfig.DNSCacheConfig == nil && in.AddonsConfig.DnsCacheConfig != nil {
			spec.AddonsConfig.DNSCacheConfig = &v1beta2.DNSCacheConfig{
				Enabled: in.AddonsConfig.DnsCacheConfig.Enabled,
			}
		}
		if spec.AddonsConfig.GCEPersistentDiskCSIDriverConfig == nil && in.AddonsConfig.GcePersistentDiskCsiDriverConfig != nil {
			spec.AddonsConfig.GCEPersistentDiskCSIDriverConfig = &v1beta2.GCEPersistentDiskCSIDriverConfig{
				Enabled: in.AddonsConfig.GcePersistentDiskCsiDriverConfig.Enabled,
			}
		}
		if spec.AddonsConfig.HorizontalPodAutoscaling == nil && in.AddonsConfig.HorizontalPodAutoscaling != nil {
			spec.AddonsConfig.HorizontalPodAutoscaling = &v1beta2.HorizontalPodAutoscaling{
				Disabled: in.AddonsConfig.HorizontalPodAutoscaling.Disabled,
			}
		}
		if spec.AddonsConfig.HTTPLoadBalancing == nil && in.AddonsConfig.HttpLoadBalancing != nil {
			spec.AddonsConfig.HTTPLoadBalancing = &v1beta2.HTTPLoadBalancing{
				Disabled: in.AddonsConfig.HttpLoadBalancing.Disabled,
			}
		}
		if spec.AddonsConfig.KubernetesDashboard == nil && in.AddonsConfig.KubernetesDashboard != nil {
			spec.AddonsConfig.KubernetesDashboard = &v1beta2.KubernetesDashboard{
				Disabled: in.AddonsConfig.KubernetesDashboard.Disabled,
			}
		}
		if spec.AddonsConfig.NetworkPolicyConfig == nil && in.AddonsConfig.NetworkPolicyConfig != nil {
			spec.AddonsConfig.NetworkPolicyConfig = &v1beta2.NetworkPolicyConfig{
				Disabled: in.AddonsConfig.NetworkPolicyConfig.Disabled,
			}
		}
	}

	if in.AuthenticatorGroupsConfig != nil {
		if spec.AuthenticatorGroupsConfig == nil {
			spec.AuthenticatorGroupsConfig = &v1beta2.AuthenticatorGroupsConfig{}
		}
		spec.AuthenticatorGroupsConfig.Enabled = gcp.LateInitializeBool(spec.AuthenticatorGroupsConfig.Enabled, in.AuthenticatorGroupsConfig.Enabled)
		spec.AuthenticatorGroupsConfig.SecurityGroup = gcp.LateInitializeString(spec.AuthenticatorGroupsConfig.SecurityGroup, in.AuthenticatorGroupsConfig.SecurityGroup)
	}

	if in.Autoscaling != nil {
		if spec.Autoscaling == nil {
			spec.Autoscaling = &v1beta2.ClusterAutoscaling{}
		}
		spec.Autoscaling.AutoprovisioningLocations = gcp.LateInitializeStringSlice(spec.Autoscaling.AutoprovisioningLocations, in.Autoscaling.AutoprovisioningLocations)
		if in.Autoscaling.AutoprovisioningNodePoolDefaults != nil {
			if spec.Autoscaling.AutoprovisioningNodePoolDefaults == nil {
				spec.Autoscaling.AutoprovisioningNodePoolDefaults = &v1beta2.AutoprovisioningNodePoolDefaults{}
			}
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.BootDiskKMSKey = gcp.LateInitializeString(spec.Autoscaling.AutoprovisioningNodePoolDefaults.BootDiskKMSKey, in.Autoscaling.AutoprovisioningNodePoolDefaults.BootDiskKmsKey)
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.DiskSizeGb = gcp.LateInitializeInt64(spec.Autoscaling.AutoprovisioningNodePoolDefaults.DiskSizeGb, in.Autoscaling.AutoprovisioningNodePoolDefaults.DiskSizeGb)
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.DiskType = gcp.LateInitializeString(spec.Autoscaling.AutoprovisioningNodePoolDefaults.DiskType, in.Autoscaling.AutoprovisioningNodePoolDefaults.DiskType)
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.MinCPUPlatform = gcp.LateInitializeString(spec.Autoscaling.AutoprovisioningNodePoolDefaults.MinCPUPlatform, in.Autoscaling.AutoprovisioningNodePoolDefaults.MinCpuPlatform)
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes = gcp.LateInitializeStringSlice(spec.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes, in.Autoscaling.AutoprovisioningNodePoolDefaults.OauthScopes)
			spec.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount = gcp.LateInitializeString(spec.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount, in.Autoscaling.AutoprovisioningNodePoolDefaults.ServiceAccount)
			if in.Autoscaling.AutoprovisioningNodePoolDefaults.Management != nil {
				if spec.Autoscaling.AutoprovisioningNodePoolDefaults.Management == nil {
					spec.Autoscaling.AutoprovisioningNodePoolDefaults.Management = &v1beta2.NodeManagement{}
				}
				spec.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoRepair = gcp.LateInitializeBool(spec.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoRepair, in.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoRepair)
				spec.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoUpgrade = gcp.LateInitializeBool(spec.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoUpgrade, in.Autoscaling.AutoprovisioningNodePoolDefaults.Management.AutoUpgrade)
			}
			if in.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig != nil {
				if spec.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig == nil {
					spec.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig = &v1beta2.ShieldedInstanceConfig{}
				}
				spec.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableIntegrityMonitoring = gcp.LateInitializeBool(spec.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableIntegrityMonitoring, in.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableIntegrityMonitoring)
				spec.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableSecureBoot = gcp.LateInitializeBool(spec.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableSecureBoot, in.Autoscaling.AutoprovisioningNodePoolDefaults.ShieldedInstanceConfig.EnableSecureBoot)
			}
			if in.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings != nil {
				if spec.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings == nil {
					spec.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings = &v1beta2.UpgradeSettings{}
				}
				spec.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxSurge = gcp.LateInitializeInt64(spec.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxSurge, in.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxSurge)
				spec.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxUnavailable = gcp.LateInitializeInt64(spec.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxUnavailable, in.Autoscaling.AutoprovisioningNodePoolDefaults.UpgradeSettings.MaxUnavailable)
			}
		}
		spec.Autoscaling.EnableNodeAutoprovisioning = gcp.LateInitializeBool(spec.Autoscaling.EnableNodeAutoprovisioning, in.Autoscaling.EnableNodeAutoprovisioning)
		if len(in.Autoscaling.ResourceLimits) != 0 && len(spec.Autoscaling.ResourceLimits) == 0 {
			spec.Autoscaling.ResourceLimits = make([]*v1beta2.ResourceLimit, len(in.Autoscaling.ResourceLimits))
			for i, limit := range in.Autoscaling.ResourceLimits {
				spec.Autoscaling.ResourceLimits[i] = &v1beta2.ResourceLimit{
					Maximum:      &limit.Maximum,
					Minimum:      &limit.Minimum,
					ResourceType: &limit.ResourceType,
				}
			}
		}
	}

	if spec.BinaryAuthorization == nil && in.BinaryAuthorization != nil {
		spec.BinaryAuthorization = &v1beta2.BinaryAuthorization{
			Enabled: in.BinaryAuthorization.Enabled,
		}
	}

	spec.ClusterIpv4Cidr = gcp.LateInitializeString(spec.ClusterIpv4Cidr, in.ClusterIpv4Cidr)

	if in.DatabaseEncryption != nil {
		if spec.DatabaseEncryption == nil {
			spec.DatabaseEncryption = &v1beta2.DatabaseEncryption{}
		}
		spec.DatabaseEncryption.KeyName = gcp.LateInitializeString(spec.DatabaseEncryption.KeyName, in.DatabaseEncryption.KeyName)
		spec.DatabaseEncryption.State = gcp.LateInitializeString(spec.DatabaseEncryption.State, in.DatabaseEncryption.State)
	}

	if spec.DefaultMaxPodsConstraint == nil && in.DefaultMaxPodsConstraint != nil {
		spec.DefaultMaxPodsConstraint = &v1beta2.MaxPodsConstraint{
			MaxPodsPerNode: in.DefaultMaxPodsConstraint.MaxPodsPerNode,
		}
	}

	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)

	spec.EnableKubernetesAlpha = gcp.LateInitializeBool(spec.EnableKubernetesAlpha, in.EnableKubernetesAlpha)
	spec.EnableTpu = gcp.LateInitializeBool(spec.EnableTpu, in.EnableTpu)
	spec.InitialClusterVersion = gcp.LateInitializeString(spec.InitialClusterVersion, in.InitialClusterVersion)

	if in.IpAllocationPolicy != nil {
		if spec.IPAllocationPolicy == nil {
			spec.IPAllocationPolicy = &v1beta2.IPAllocationPolicy{}
		}
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
		spec.LegacyAbac = &v1beta2.LegacyAbac{
			Enabled: in.LegacyAbac.Enabled,
		}
	}

	spec.Locations = gcp.LateInitializeStringSlice(spec.Locations, in.Locations)
	spec.LoggingService = gcp.LateInitializeString(spec.LoggingService, in.LoggingService)

	if spec.MaintenancePolicy == nil && in.MaintenancePolicy != nil {
		if in.MaintenancePolicy.Window != nil {
			spec.MaintenancePolicy = &v1beta2.MaintenancePolicySpec{
				Window: v1beta2.MaintenanceWindowSpec{},
			}
			if in.MaintenancePolicy.Window.DailyMaintenanceWindow != nil {
				spec.MaintenancePolicy.Window.DailyMaintenanceWindow = &v1beta2.DailyMaintenanceWindowSpec{
					StartTime: in.MaintenancePolicy.Window.DailyMaintenanceWindow.StartTime,
				}
			}
			if in.MaintenancePolicy.Window.MaintenanceExclusions != nil {
				spec.MaintenancePolicy.Window.MaintenanceExclusions = make(map[string]v1beta2.TimeWindow, len(in.MaintenancePolicy.Window.MaintenanceExclusions))
				for k, v := range in.MaintenancePolicy.Window.MaintenanceExclusions {
					spec.MaintenancePolicy.Window.MaintenanceExclusions[k] = v1beta2.TimeWindow{
						EndTime:   v.EndTime,
						StartTime: v.StartTime,
					}
				}
			}
			if in.MaintenancePolicy.Window.RecurringWindow != nil {
				spec.MaintenancePolicy.Window.RecurringWindow = &v1beta2.RecurringTimeWindow{
					Recurrence: &in.MaintenancePolicy.Window.RecurringWindow.Recurrence,
					Window: &v1beta2.TimeWindow{
						EndTime:   spec.MaintenancePolicy.Window.RecurringWindow.Window.EndTime,
						StartTime: spec.MaintenancePolicy.Window.RecurringWindow.Window.StartTime,
					},
				}
			}
		}
	}

	if in.MasterAuth != nil {
		if spec.MasterAuth == nil {
			spec.MasterAuth = &v1beta2.MasterAuth{}
		}
		if in.MasterAuth.ClientCertificateConfig != nil {
			spec.MasterAuth.ClientCertificateConfig = &v1beta2.ClientCertificateConfig{
				IssueClientCertificate: in.MasterAuth.ClientCertificateConfig.IssueClientCertificate,
			}
		}
		spec.MasterAuth.Username = gcp.LateInitializeString(spec.MasterAuth.Username, in.MasterAuth.Username)
	}

	if in.MasterAuthorizedNetworksConfig != nil {
		if spec.MasterAuthorizedNetworksConfig == nil {
			spec.MasterAuthorizedNetworksConfig = &v1beta2.MasterAuthorizedNetworksConfig{}
		}
		if len(in.MasterAuthorizedNetworksConfig.CidrBlocks) != 0 && len(spec.MasterAuthorizedNetworksConfig.CidrBlocks) == 0 {
			spec.MasterAuthorizedNetworksConfig.CidrBlocks = make([]*v1beta2.CidrBlock, len(in.MasterAuthorizedNetworksConfig.CidrBlocks))
			for i, block := range in.MasterAuthorizedNetworksConfig.CidrBlocks {
				spec.MasterAuthorizedNetworksConfig.CidrBlocks[i] = &v1beta2.CidrBlock{
					CidrBlock:   block.CidrBlock,
					DisplayName: &block.DisplayName,
				}
			}
		}
		spec.MasterAuthorizedNetworksConfig.Enabled = gcp.LateInitializeBool(spec.MasterAuthorizedNetworksConfig.Enabled, in.MasterAuthorizedNetworksConfig.Enabled)
	}

	spec.MonitoringService = gcp.LateInitializeString(spec.MonitoringService, in.MonitoringService)
	spec.Network = gcp.LateInitializeString(spec.Network, in.Network)

	if in.NetworkConfig != nil {
		if spec.NetworkConfig == nil {
			spec.NetworkConfig = &v1beta2.NetworkConfigSpec{}
		}
		spec.NetworkConfig.EnableIntraNodeVisibility = gcp.LateInitializeBool(spec.NetworkConfig.EnableIntraNodeVisibility, in.NetworkConfig.EnableIntraNodeVisibility)
		spec.NetworkConfig.PrivateIpv6GoogleAccess = gcp.LateInitializeString(spec.NetworkConfig.PrivateIpv6GoogleAccess, in.NetworkConfig.PrivateIpv6GoogleAccess)
		spec.NetworkConfig.DatapathProvider = gcp.LateInitializeString(spec.NetworkConfig.DatapathProvider, in.NetworkConfig.DatapathProvider)
		if spec.NetworkConfig.DefaultSnatStatus == nil && in.NetworkConfig.DefaultSnatStatus != nil {
			spec.NetworkConfig.DefaultSnatStatus = &v1beta2.DefaultSnatStatus{
				Disabled: in.NetworkConfig.DefaultSnatStatus.Disabled,
			}
		}
	}

	if in.NetworkPolicy != nil {
		if spec.NetworkPolicy == nil {
			spec.NetworkPolicy = &v1beta2.NetworkPolicy{}
		}
		spec.NetworkPolicy.Enabled = gcp.LateInitializeBool(spec.NetworkPolicy.Enabled, in.NetworkPolicy.Enabled)
		spec.NetworkPolicy.Provider = gcp.LateInitializeString(spec.NetworkPolicy.Provider, in.NetworkPolicy.Provider)
	}

	if in.PrivateClusterConfig != nil {
		if spec.PrivateClusterConfig == nil {
			spec.PrivateClusterConfig = &v1beta2.PrivateClusterConfigSpec{}
		}
		spec.PrivateClusterConfig.EnablePrivateEndpoint = gcp.LateInitializeBool(spec.PrivateClusterConfig.EnablePrivateEndpoint, in.PrivateClusterConfig.EnablePrivateEndpoint)
		spec.PrivateClusterConfig.EnablePrivateNodes = gcp.LateInitializeBool(spec.PrivateClusterConfig.EnablePrivateNodes, in.PrivateClusterConfig.EnablePrivateNodes)
		spec.PrivateClusterConfig.MasterIpv4CidrBlock = gcp.LateInitializeString(spec.PrivateClusterConfig.MasterIpv4CidrBlock, in.PrivateClusterConfig.MasterIpv4CidrBlock)
		if in.PrivateClusterConfig.MasterGlobalAccessConfig != nil && spec.PrivateClusterConfig.MasterGlobalAccessConfig == nil {
			spec.PrivateClusterConfig.MasterGlobalAccessConfig = &v1beta2.PrivateClusterMasterGlobalAccessConfig{
				Enabled: in.PrivateClusterConfig.MasterGlobalAccessConfig.Enabled,
			}
		}
	}

	if in.NotificationConfig != nil && in.NotificationConfig.Pubsub != nil && spec.NotificationConfig == nil {
		if spec.NotificationConfig == nil {
			spec.NotificationConfig = &v1beta2.NotificationConfig{
				Pubsub: v1beta2.PubSub{
					Enabled: in.NotificationConfig.Pubsub.Enabled,
					Topic:   in.NotificationConfig.Pubsub.Topic,
				},
			}
		}
	}

	if in.ReleaseChannel != nil && spec.ReleaseChannel == nil {
		spec.ReleaseChannel = &v1beta2.ReleaseChannel{
			Channel: in.ReleaseChannel.Channel,
		}
	}

	spec.ResourceLabels = gcp.LateInitializeStringMap(spec.ResourceLabels, in.ResourceLabels)

	if in.ResourceUsageExportConfig != nil {
		if spec.ResourceUsageExportConfig == nil {
			spec.ResourceUsageExportConfig = &v1beta2.ResourceUsageExportConfig{}
		}
		if spec.ResourceUsageExportConfig.BigqueryDestination == nil && in.ResourceUsageExportConfig.BigqueryDestination != nil {
			spec.ResourceUsageExportConfig.BigqueryDestination = &v1beta2.BigQueryDestination{
				DatasetID: in.ResourceUsageExportConfig.BigqueryDestination.DatasetId,
			}
		}
		if spec.ResourceUsageExportConfig.ConsumptionMeteringConfig == nil && in.ResourceUsageExportConfig.ConsumptionMeteringConfig != nil {
			spec.ResourceUsageExportConfig.ConsumptionMeteringConfig = &v1beta2.ConsumptionMeteringConfig{
				Enabled: in.ResourceUsageExportConfig.ConsumptionMeteringConfig.Enabled,
			}
		}
		spec.ResourceUsageExportConfig.EnableNetworkEgressMetering = gcp.LateInitializeBool(spec.ResourceUsageExportConfig.EnableNetworkEgressMetering, in.ResourceUsageExportConfig.EnableNetworkEgressMetering)
	}

	spec.Subnetwork = gcp.LateInitializeString(spec.Subnetwork, in.Subnetwork)

	if spec.VerticalPodAutoscaling == nil && in.VerticalPodAutoscaling != nil {
		spec.VerticalPodAutoscaling = &v1beta2.VerticalPodAutoscaling{
			Enabled: in.VerticalPodAutoscaling.Enabled,
		}
	}

	if spec.WorkloadIdentityConfig == nil && in.WorkloadIdentityConfig != nil {
		spec.WorkloadIdentityConfig = &v1beta2.WorkloadIdentityConfig{
			WorkloadPool: in.WorkloadIdentityConfig.WorkloadPool,
		}
	}
}

// newAddonsConfigUpdateFn returns a function that updates the AddonsConfig of a cluster.
func newAddonsConfigUpdateFn(in *v1beta2.AddonsConfig) UpdateFn {
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
func newAutoscalingUpdateFn(in *v1beta2.ClusterAutoscaling) UpdateFn {
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
func newBinaryAuthorizationUpdateFn(in *v1beta2.BinaryAuthorization) UpdateFn {
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

// newAutopilotUpdateFn returns a function that updates the Autopilot of a cluster.
func newAutopilotUpdateFn(in *v1beta2.Autopilot) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateAutopilot(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredAutopilot: out.Autopilot,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newDatabaseEncryptionUpdateFn returns a function that updates the DatabaseEncryption of a cluster.
func newDatabaseEncryptionUpdateFn(in *v1beta2.DatabaseEncryption) UpdateFn {
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
func newLegacyAbacUpdateFn(in *v1beta2.LegacyAbac) UpdateFn {
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
func newMaintenancePolicyUpdateFn(in *v1beta2.MaintenancePolicySpec) UpdateFn {
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
func newMasterAuthorizedNetworksConfigUpdateFn(in *v1beta2.MasterAuthorizedNetworksConfig) UpdateFn {
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

// newDatapathProviderUpdateFn returns a function that updates the
// DatapathProvider of a cluster.
func newDatapathProviderUpdateFn(in *string) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredDatapathProvider: gcp.StringValue(in),
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newIntraNodeVisibilityConfigUpdateFn returns a function that updates the
// IntraNodeVisibility of a cluster.
func newIntraNodeVisibilityConfigUpdateFn(in *bool) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredIntraNodeVisibilityConfig: &container.IntraNodeVisibilityConfig{
					Enabled: gcp.BoolValue(in),
				},
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newNetworkPolicyUpdateFn returns a function that updates the NetworkPolicy of a cluster.
func newNetworkPolicyUpdateFn(in *v1beta2.NetworkPolicy) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateNetworkPolicy(in, out)
		update := &container.SetNetworkPolicyRequest{
			NetworkPolicy: out.NetworkPolicy,
		}
		return s.Projects.Locations.Clusters.SetNetworkPolicy(name, update).Context(ctx).Do()
	}
}

// newNotificationConfigUpdateFn returns a function that updates the NotificationConfig of a cluster.
func newNotificationConfigUpdateFn(in *v1beta2.NotificationConfig) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateNotificationConfig(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredNotificationConfig: out.NotificationConfig,
			},
		}
		return s.Projects.Locations.Clusters.Update(name, update).Context(ctx).Do()
	}
}

// newPrivateClusterConfigUpdateFn returns a function that updates the PrivateClusterConfig of a cluster.
func newPrivateClusterConfigUpdateFn(in *v1beta2.PrivateClusterConfigSpec) UpdateFn {
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

// newReleaseChannelUpdateFn returns a function that updates the ReleaseChannel of a cluster.
func newReleaseChannelUpdateFn(in *v1beta2.ReleaseChannel) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.Cluster{}
		GenerateReleaseChannel(in, out)
		update := &container.UpdateClusterRequest{
			Update: &container.ClusterUpdate{
				DesiredReleaseChannel: out.ReleaseChannel,
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
func newResourceUsageExportConfigUpdateFn(in *v1beta2.ResourceUsageExportConfig) UpdateFn {
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
func newVerticalPodAutoscalingUpdateFn(in *v1beta2.VerticalPodAutoscaling) UpdateFn {
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
func newWorkloadIdentityConfigUpdateFn(in *v1beta2.WorkloadIdentityConfig) UpdateFn {
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
func IsUpToDate(name string, in *v1beta2.ClusterParameters, observed *container.Cluster) (bool, UpdateFn, error) { // nolint:gocyclo
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
	if !cmp.Equal(desired.AddonsConfig, observed.AddonsConfig, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "CloudRunConfig.ForceSendFields"),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "ConfigConnectorConfig.ForceSendFields"),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "DnsCacheConfig.ForceSendFields"),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "GcePersistentDiskCsiDriverConfig.ForceSendFields"),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "HorizontalPodAutoscaling.ForceSendFields"),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "HttpLoadBalancing.ForceSendFields"),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "KubernetesDashboard.ForceSendFields"),
		cmpopts.IgnoreFields(container.AddonsConfig{}, "NetworkPolicyConfig.ForceSendFields")) {
		return false, newAddonsConfigUpdateFn(in.AddonsConfig), nil
	}
	if !cmp.Equal(desired.Autopilot, observed.Autopilot, cmpopts.EquateEmpty()) {
		return false, newAutopilotUpdateFn(in.Autopilot), nil
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
	if desired.NetworkConfig != nil {
		if observed.NetworkConfig == nil {
			observed.NetworkConfig = &container.NetworkConfig{}
		}
		if !cmp.Equal(desired.NetworkConfig.EnableIntraNodeVisibility, observed.NetworkConfig.EnableIntraNodeVisibility, cmpopts.EquateEmpty()) {
			return false, newIntraNodeVisibilityConfigUpdateFn(in.NetworkConfig.EnableIntraNodeVisibility), nil
		}
		if !cmp.Equal(desired.NetworkConfig.DatapathProvider, observed.NetworkConfig.DatapathProvider, cmpopts.EquateEmpty()) {
			return false, newDatapathProviderUpdateFn(in.NetworkConfig.DatapathProvider), nil
		}
	}

	if !cmp.Equal(desired.NetworkPolicy, observed.NetworkPolicy, cmpopts.EquateEmpty()) {
		return false, newNetworkPolicyUpdateFn(in.NetworkPolicy), nil
	}
	if !cmp.Equal(desired.NotificationConfig, observed.NotificationConfig, cmpopts.EquateEmpty()) {
		return false, newNotificationConfigUpdateFn(in.NotificationConfig), nil
	}
	if !cmp.Equal(desired.PrivateClusterConfig, observed.PrivateClusterConfig, cmpopts.EquateEmpty()) {
		return false, newPrivateClusterConfigUpdateFn(in.PrivateClusterConfig), nil
	}
	if !cmp.Equal(desired.ReleaseChannel, observed.ReleaseChannel, cmpopts.EquateEmpty()) {
		return false, newReleaseChannelUpdateFn(in.ReleaseChannel), nil
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
func GetFullyQualifiedParent(project string, p v1beta2.ClusterParameters) string {
	return fmt.Sprintf(ParentFormat, project, p.Location)
}

// GetFullyQualifiedName builds the fully qualified name of the cluster.
func GetFullyQualifiedName(project string, p v1beta2.ClusterParameters, name string) string {
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
