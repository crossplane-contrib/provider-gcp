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

package nodepool

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1beta1"

	"github.com/crossplaneio/stack-gcp/apis/container/v1alpha1"
	"github.com/crossplaneio/stack-gcp/apis/container/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
)

const (
	// NodePoolNameFormat is the format for the fully qualified name of a node pool.
	NodePoolNameFormat = "%s/nodePools/%s"

	errCheckUpToDate = "unable to determine if external resource is up to date"
)

// GenerateNodePool generates *container.NodePool instance from NodePoolParameters.
func GenerateNodePool(name string, in v1alpha1.NodePoolParameters, pool *container.NodePool) { // nolint:gocyclo
	pool.InitialNodeCount = gcp.Int64Value(in.InitialNodeCount)
	pool.Locations = in.Locations
	pool.Name = name
	pool.Version = gcp.StringValue(in.Version)

	GenerateAutoscaling(in.Autoscaling, pool)
	GenerateConfig(in.Config, pool)
	GenerateManagement(in.Management, pool)
	GenerateMaxPodsConstraint(in.MaxPodsConstraint, pool)
}

// GenerateAutoscaling generates *container.Autoscaling from *Autoscaling.
func GenerateAutoscaling(in *v1alpha1.NodePoolAutoscaling, pool *container.NodePool) {
	if in != nil {
		if pool.Autoscaling == nil {
			pool.Autoscaling = &container.NodePoolAutoscaling{}
		}
		pool.Autoscaling.Autoprovisioned = gcp.BoolValue(in.Autoprovisioned)
		pool.Autoscaling.Enabled = gcp.BoolValue(in.Enabled)
		pool.Autoscaling.MaxNodeCount = gcp.Int64Value(in.MaxNodeCount)
		pool.Autoscaling.MinNodeCount = gcp.Int64Value(in.MinNodeCount)
	}
}

// GenerateConfig generates *container.Config from *NodeConfig.
func GenerateConfig(in *v1alpha1.NodeConfig, pool *container.NodePool) { // nolint:gocyclo
	if in != nil {
		if pool.Config == nil {
			pool.Config = &container.NodeConfig{}
		}
		pool.Config.DiskSizeGb = gcp.Int64Value(in.DiskSizeGb)
		pool.Config.DiskType = gcp.StringValue(in.DiskType)
		pool.Config.ImageType = gcp.StringValue(in.ImageType)
		pool.Config.Labels = in.Labels
		pool.Config.LocalSsdCount = gcp.Int64Value(in.LocalSsdCount)
		pool.Config.MachineType = gcp.StringValue(in.MachineType)
		pool.Config.Metadata = in.Metadata
		pool.Config.MinCpuPlatform = gcp.StringValue(in.MinCPUPlatform)
		pool.Config.OauthScopes = in.OauthScopes
		pool.Config.Preemptible = gcp.BoolValue(in.Preemptible)
		pool.Config.ServiceAccount = gcp.StringValue(in.ServiceAccount)
		pool.Config.Tags = in.Tags

		if len(in.Accelerators) > 0 {
			pool.Config.Accelerators = make([]*container.AcceleratorConfig, len(in.Accelerators))
		}
		for i, a := range in.Accelerators {
			if a != nil {
				pool.Config.Accelerators[i] = &container.AcceleratorConfig{
					AcceleratorCount: a.AcceleratorCount,
					AcceleratorType:  a.AcceleratorType,
				}
			}
		}

		if in.SandboxConfig != nil {
			if pool.Config.SandboxConfig == nil {
				pool.Config.SandboxConfig = &container.SandboxConfig{}
			}
			pool.Config.SandboxConfig.SandboxType = in.SandboxConfig.SandboxType
		}

		if in.ShieldedInstanceConfig != nil {
			if pool.Config.ShieldedInstanceConfig == nil {
				pool.Config.ShieldedInstanceConfig = &container.ShieldedInstanceConfig{}
			}
			pool.Config.ShieldedInstanceConfig.EnableIntegrityMonitoring = gcp.BoolValue(in.ShieldedInstanceConfig.EnableIntegrityMonitoring)
			pool.Config.ShieldedInstanceConfig.EnableSecureBoot = gcp.BoolValue(in.ShieldedInstanceConfig.EnableSecureBoot)
		}

		if len(in.Taints) > 0 {
			pool.Config.Taints = make([]*container.NodeTaint, len(in.Taints))
		}
		for i, t := range in.Taints {
			if t != nil {
				pool.Config.Taints[i] = &container.NodeTaint{
					Effect: t.Effect,
					Key:    t.Key,
					Value:  t.Value,
				}
			}
		}

		if in.WorkloadMetadataConfig != nil {
			if pool.Config.WorkloadMetadataConfig == nil {
				pool.Config.WorkloadMetadataConfig = &container.WorkloadMetadataConfig{}
			}
			pool.Config.WorkloadMetadataConfig.NodeMetadata = in.WorkloadMetadataConfig.NodeMetadata
		}
	}
}

// GenerateManagement generates *container.NodeManagement from *NodeManagementSpec.
func GenerateManagement(in *v1alpha1.NodeManagementSpec, pool *container.NodePool) {
	if in != nil {
		if pool.Management == nil {
			pool.Management = &container.NodeManagement{}
		}
		pool.Management.AutoRepair = gcp.BoolValue(in.AutoRepair)
		pool.Management.AutoUpgrade = gcp.BoolValue(in.AutoUpgrade)
	}
}

// GenerateMaxPodsConstraint generates *container.MaxPodsConstraint from *MaxPodsConstraint.
func GenerateMaxPodsConstraint(in *v1beta1.MaxPodsConstraint, pool *container.NodePool) {
	if in != nil {
		if pool.MaxPodsConstraint == nil {
			pool.MaxPodsConstraint = &container.MaxPodsConstraint{}
		}
		pool.MaxPodsConstraint.MaxPodsPerNode = in.MaxPodsPerNode
	}
}

// GenerateObservation produces NodePoolObservation object from *container.NodePool object.
func GenerateObservation(in container.NodePool) v1alpha1.NodePoolObservation { // nolint:gocyclo
	o := v1alpha1.NodePoolObservation{
		InstanceGroupUrls: in.InstanceGroupUrls,
		PodIpv4CidrSize:   in.PodIpv4CidrSize,
		SelfLink:          in.SelfLink,
		Status:            in.Status,
		StatusMessage:     in.StatusMessage,
	}

	for _, condition := range in.Conditions {
		if condition != nil {
			o.Conditions = append(o.Conditions, &v1beta1.StatusCondition{
				Code:    condition.Code,
				Message: condition.Message,
			})
		}
	}

	if in.Management != nil && in.Management.UpgradeOptions != nil {
		o.Management = &v1alpha1.NodeManagementStatus{
			UpgradeOptions: &v1alpha1.AutoUpgradeOptions{
				AutoUpgradeStartTime: in.Management.UpgradeOptions.AutoUpgradeStartTime,
				Description:          in.Management.UpgradeOptions.Description,
			},
		}
	}

	return o

}

// GenerateNodePoolUpdate produces NodePoolObservation object from *container.NodePool object.
func GenerateNodePoolUpdate(in *v1alpha1.NodePoolParameters) *container.UpdateNodePoolRequest { // nolint:gocyclo
	o := &container.UpdateNodePoolRequest{
		Locations:   in.Locations,
		NodeVersion: gcp.StringValue(in.Version),
	}

	if in.Config != nil {
		o.ImageType = gcp.StringValue(in.Config.ImageType)

		if in.Config.WorkloadMetadataConfig != nil {
			o.WorkloadMetadataConfig = &container.WorkloadMetadataConfig{
				NodeMetadata: in.Config.WorkloadMetadataConfig.NodeMetadata,
			}
		}
	}

	return o
}

// LateInitializeSpec fills unassigned fields with the values in container.NodePool object.
func LateInitializeSpec(spec *v1alpha1.NodePoolParameters, in container.NodePool) { // nolint:gocyclo
	if in.Autoscaling != nil {
		if spec.Autoscaling == nil {
			spec.Autoscaling = &v1alpha1.NodePoolAutoscaling{}
		}

		spec.Autoscaling.Autoprovisioned = gcp.LateInitializeBool(spec.Autoscaling.Autoprovisioned, in.Autoscaling.Autoprovisioned)
		spec.Autoscaling.Enabled = gcp.LateInitializeBool(spec.Autoscaling.Enabled, in.Autoscaling.Enabled)
		spec.Autoscaling.MaxNodeCount = gcp.LateInitializeInt64(spec.Autoscaling.MaxNodeCount, in.Autoscaling.MaxNodeCount)
		spec.Autoscaling.MinNodeCount = gcp.LateInitializeInt64(spec.Autoscaling.MinNodeCount, in.Autoscaling.MinNodeCount)
	}

	if in.Config != nil {
		if spec.Config == nil {
			spec.Config = &v1alpha1.NodeConfig{}
		}

		if len(in.Config.Accelerators) != 0 && len(spec.Config.Accelerators) == 0 {
			spec.Config.Accelerators = make([]*v1alpha1.AcceleratorConfig, len(in.Config.Accelerators))
			for i, a := range in.Config.Accelerators {
				spec.Config.Accelerators[i] = &v1alpha1.AcceleratorConfig{
					AcceleratorCount: a.AcceleratorCount,
					AcceleratorType:  a.AcceleratorType,
				}
			}
		}

		spec.Config.DiskSizeGb = gcp.LateInitializeInt64(spec.Config.DiskSizeGb, in.Config.DiskSizeGb)
		spec.Config.DiskType = gcp.LateInitializeString(spec.Config.DiskType, in.Config.DiskType)
		spec.Config.ImageType = gcp.LateInitializeString(spec.Config.ImageType, in.Config.ImageType)
		spec.Config.Labels = gcp.LateInitializeStringMap(spec.Config.Labels, in.Config.Labels)
		spec.Config.LocalSsdCount = gcp.LateInitializeInt64(spec.Config.LocalSsdCount, in.Config.LocalSsdCount)
		spec.Config.MachineType = gcp.LateInitializeString(spec.Config.MachineType, in.Config.MachineType)
		spec.Config.Metadata = gcp.LateInitializeStringMap(spec.Config.Metadata, in.Config.Metadata)
		spec.Config.MinCPUPlatform = gcp.LateInitializeString(spec.Config.MinCPUPlatform, in.Config.MinCpuPlatform)
		spec.Config.OauthScopes = gcp.LateInitializeStringSlice(spec.Config.OauthScopes, in.Config.OauthScopes)
		spec.Config.Preemptible = gcp.LateInitializeBool(spec.Config.Preemptible, in.Config.Preemptible)

		if in.Config.SandboxConfig != nil && spec.Config.SandboxConfig == nil {
			spec.Config.SandboxConfig = &v1alpha1.SandboxConfig{
				SandboxType: in.Config.SandboxConfig.SandboxType,
			}
		}

		spec.Config.ServiceAccount = gcp.LateInitializeString(spec.Config.ServiceAccount, in.Config.ServiceAccount)

		if in.Config.ShieldedInstanceConfig != nil {
			if spec.Config.ShieldedInstanceConfig == nil {
				spec.Config.ShieldedInstanceConfig = &v1alpha1.ShieldedInstanceConfig{}
			}
			spec.Config.ShieldedInstanceConfig.EnableIntegrityMonitoring = gcp.LateInitializeBool(spec.Config.ShieldedInstanceConfig.EnableIntegrityMonitoring, in.Config.ShieldedInstanceConfig.EnableIntegrityMonitoring)
			spec.Config.ShieldedInstanceConfig.EnableSecureBoot = gcp.LateInitializeBool(spec.Config.ShieldedInstanceConfig.EnableSecureBoot, in.Config.ShieldedInstanceConfig.EnableSecureBoot)
		}

		spec.Config.Tags = gcp.LateInitializeStringSlice(spec.Config.Tags, in.Config.Tags)

		if len(in.Config.Taints) != 0 && len(spec.Config.Taints) == 0 {
			spec.Config.Taints = make([]*v1alpha1.NodeTaint, len(in.Config.Taints))
			for i, t := range in.Config.Taints {
				spec.Config.Taints[i] = &v1alpha1.NodeTaint{
					Effect: t.Effect,
					Key:    t.Key,
					Value:  t.Value,
				}
			}
		}

		if in.Config.WorkloadMetadataConfig != nil && spec.Config.WorkloadMetadataConfig == nil {
			spec.Config.WorkloadMetadataConfig = &v1alpha1.WorkloadMetadataConfig{
				NodeMetadata: in.Config.WorkloadMetadataConfig.NodeMetadata,
			}
		}
	}

	spec.InitialNodeCount = gcp.LateInitializeInt64(spec.InitialNodeCount, in.InitialNodeCount)
	spec.Locations = gcp.LateInitializeStringSlice(spec.Locations, in.Locations)

	if in.Management != nil {
		if spec.Management == nil {
			spec.Management = &v1alpha1.NodeManagementSpec{}
		}

		spec.Management.AutoRepair = gcp.LateInitializeBool(spec.Management.AutoRepair, in.Management.AutoRepair)
		spec.Management.AutoUpgrade = gcp.LateInitializeBool(spec.Management.AutoUpgrade, in.Management.AutoUpgrade)
	}

	if in.MaxPodsConstraint != nil && spec.MaxPodsConstraint == nil {
		spec.MaxPodsConstraint = &v1beta1.MaxPodsConstraint{
			MaxPodsPerNode: in.MaxPodsConstraint.MaxPodsPerNode,
		}
	}

	spec.Version = gcp.LateInitializeString(spec.Version, in.Version)
}

// newAutoscalingUpdateFn returns a function that updates the Autoscaling of a node pool.
func newAutoscalingUpdateFn(in *v1alpha1.NodePoolAutoscaling) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.NodePool{}
		GenerateAutoscaling(in, out)
		update := &container.SetNodePoolAutoscalingRequest{
			Autoscaling: out.Autoscaling,
		}
		return s.Projects.Locations.Clusters.NodePools.SetAutoscaling(name, update).Context(ctx).Do()
	}
}

// newManagementUpdateFn returns a function that updates the Management of a node pool.
func newManagementUpdateFn(in *v1alpha1.NodeManagementSpec) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		out := &container.NodePool{}
		GenerateManagement(in, out)
		update := &container.SetNodePoolManagementRequest{
			Management: out.Management,
		}
		return s.Projects.Locations.Clusters.NodePools.SetManagement(name, update).Context(ctx).Do()
	}
}

// newGeneralUpdateFn returns a function that updates a node pool.
func newGeneralUpdateFn(in *v1alpha1.NodePoolParameters) UpdateFn {
	return func(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
		return s.Projects.Locations.Clusters.NodePools.Update(name, GenerateNodePoolUpdate(in)).Context(ctx).Do()
	}
}

func noOpUpdate(ctx context.Context, s *container.Service, name string) (*container.Operation, error) {
	return nil, nil
}

// UpdateFn returns a function that updates a node pool.
type UpdateFn func(context.Context, *container.Service, string) (*container.Operation, error)

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, in *v1alpha1.NodePoolParameters, observed *container.NodePool) (bool, UpdateFn, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, noOpUpdate, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*container.NodePool)
	if !ok {
		return true, noOpUpdate, errors.New(errCheckUpToDate)
	}
	GenerateNodePool(name, *in, desired)
	if !cmp.Equal(desired.Autoscaling, observed.Autoscaling) {
		return false, newAutoscalingUpdateFn(in.Autoscaling), nil
	}
	if !cmp.Equal(desired.Management, observed.Management) {
		return false, newManagementUpdateFn(in.Management), nil
	}
	if !cmp.Equal(desired, observed) {
		return false, newGeneralUpdateFn(in), nil
	}
	return true, noOpUpdate, nil
}

// GetFullyQualifiedName builds the fully qualified name of the cluster.
func GetFullyQualifiedName(p v1alpha1.NodePoolParameters, name string) string {
	// Zonal clusters use /zones/ in their path instead of /locations/. We
	// manage node pools using the locations API endpoint so we must modify the
	// path.
	return strings.Replace(fmt.Sprintf(NodePoolNameFormat, p.Cluster, name), "/zones/", "/locations/", -1)
}
