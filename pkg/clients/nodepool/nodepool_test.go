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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	container "google.golang.org/api/container/v1beta1"

	"github.com/crossplane/provider-gcp/apis/container/v1alpha1"
	"github.com/crossplane/provider-gcp/apis/container/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const (
	name                  = "my-cool-pool"
	cluster               = "/projects/cool-proj/locations/us-central1/clusters/cool-cluster"
	zonalCluster          = "/projects/cool-proj/zones/us-central1-a/clusters/cool-cluster"
	zonalClusterFormatted = "/projects/cool-proj/locations/us-central1-a/clusters/cool-cluster"
)

func nodePool(m ...func(*container.NodePool)) *container.NodePool {
	n := &container.NodePool{
		Name:             name,
		InitialNodeCount: 3,
		Locations:        []string{"us-central1-a"},
	}
	for _, f := range m {
		f(n)
	}

	return n
}

func params(m ...func(*v1alpha1.NodePoolParameters)) *v1alpha1.NodePoolParameters {
	p := &v1alpha1.NodePoolParameters{
		Cluster:          cluster,
		InitialNodeCount: gcp.Int64Ptr(3),
		Locations:        []string{"us-central1-a"},
	}
	for _, f := range m {
		f(p)
	}

	return p
}

func observation(m ...func(*v1alpha1.NodePoolObservation)) *v1alpha1.NodePoolObservation {
	o := &v1alpha1.NodePoolObservation{
		Conditions: []*v1beta1.StatusCondition{
			{
				Code:    "UNKNOWN",
				Message: "Condition is unknown.",
			},
		},
		InstanceGroupUrls: []string{
			"cool-group-1",
			"cool-group-2",
		},
		PodIpv4CidrSize: 24,
		Management: &v1alpha1.NodeManagementStatus{
			UpgradeOptions: &v1alpha1.AutoUpgradeOptions{
				AutoUpgradeStartTime: "13:13",
				Description:          "Time to upgrade.",
			},
		},
		SelfLink:      "/link/to/myself",
		Status:        "RUNNING",
		StatusMessage: "I am running.",
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func addOutputFields(n *container.NodePool) {
	n.Conditions = []*container.StatusCondition{
		{
			Code:    "UNKNOWN",
			Message: "Condition is unknown.",
		},
	}
	n.InstanceGroupUrls = []string{
		"cool-group-1",
		"cool-group-2",
	}
	n.PodIpv4CidrSize = 24
	n.SelfLink = "/link/to/myself"
	n.Management = &container.NodeManagement{
		UpgradeOptions: &container.AutoUpgradeOptions{
			AutoUpgradeStartTime: "13:13",
			Description:          "Time to upgrade.",
		},
	}
	n.Status = "RUNNING"
	n.StatusMessage = "I am running."
}

func TestGenerateObservation(t *testing.T) {
	type args struct {
		nodePool *container.NodePool
	}

	tests := map[string]struct {
		args args
		want *v1alpha1.NodePoolObservation
	}{
		"Successful": {
			args: args{
				nodePool: nodePool(addOutputFields),
			},
			want: observation(),
		},
		"SuccessfulWithChange": {
			args: args{
				nodePool: nodePool(addOutputFields, func(n *container.NodePool) {
					n.PodIpv4CidrSize = 16
				}),
			},
			want: observation(func(p *v1alpha1.NodePoolObservation) {
				p.PodIpv4CidrSize = 16
			}),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			observation := GenerateObservation(*tc.args.nodePool)
			if diff := cmp.Diff(*tc.want, observation); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateNodePool(t *testing.T) {
	type args struct {
		nodePool *container.NodePool
		params   *v1alpha1.NodePoolParameters
		name     string
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: nodePool(),
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.Autoscaling = &v1alpha1.NodePoolAutoscaling{
						Enabled:      gcp.BoolPtr(true),
						MaxNodeCount: gcp.Int64Ptr(3),
						MinNodeCount: gcp.Int64Ptr(3),
					}
				}),
				name: name,
			},
			want: nodePool(func(n *container.NodePool) {
				n.Autoscaling = &container.NodePoolAutoscaling{
					Enabled:      true,
					MaxNodeCount: 3,
					MinNodeCount: 3,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				nodePool: nodePool(),
				params:   params(),
				name:     name,
			},
			want: nodePool(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateNodePool(tc.args.name, *tc.args.params, tc.args.nodePool)
			if diff := cmp.Diff(tc.want, tc.args.nodePool); diff != "" {
				t.Errorf("GenerateNodePool(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAutoscaling(t *testing.T) {
	enable := true
	count := int64(3)

	type args struct {
		nodePool *container.NodePool
		params   *v1alpha1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.Autoscaling = &v1alpha1.NodePoolAutoscaling{
						Autoprovisioned: &enable,
						Enabled:         &enable,
						MaxNodeCount:    &count,
						MinNodeCount:    &count,
					}
				}),
			},
			want: nodePool(func(n *container.NodePool) {
				n.Autoscaling = &container.NodePoolAutoscaling{
					Autoprovisioned: enable,
					Enabled:         enable,
					MaxNodeCount:    count,
					MinNodeCount:    count,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				nodePool: &container.NodePool{},
				params:   params(),
			},
			want: nodePool(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateAutoscaling(tc.args.params.Autoscaling, tc.args.nodePool)
			if diff := cmp.Diff(tc.want.Autoscaling, tc.args.nodePool.Autoscaling); diff != "" {
				t.Errorf("GenerateAutoscaling(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	diskSizeGb := int64(10)
	diskType := "SSD"
	enable := true
	imageType := "CO"
	labels := map[string]string{
		"cool": "value",
	}
	localSsdCount := int64(3)
	machineType := "n1-standard"
	metadata := map[string]string{
		"cool": "metadata",
	}
	minCPUPlatform := "mincpu"
	oauthScopes := []string{"scope-1"}
	preemptible := true
	serviceAccount := "my-cool-account"
	tags := []string{"tag"}
	accConf := &v1alpha1.AcceleratorConfig{
		AcceleratorCount: 3,
		AcceleratorType:  "nvidia-tesla-t4",
	}
	taint := &v1alpha1.NodeTaint{
		Effect: "NO_SCHEDULE",
		Key:    "cool-key",
		Value:  "cool-val",
	}
	gcpAccConf := &container.AcceleratorConfig{
		AcceleratorCount: 3,
		AcceleratorType:  "nvidia-tesla-t4",
	}
	gcpTaint := &container.NodeTaint{
		Effect: "NO_SCHEDULE",
		Key:    "cool-key",
		Value:  "cool-val",
	}

	type args struct {
		nodePool *container.NodePool
		params   *v1alpha1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.Config = &v1alpha1.NodeConfig{
						Accelerators:   []*v1alpha1.AcceleratorConfig{accConf},
						DiskSizeGb:     gcp.Int64Ptr(diskSizeGb),
						DiskType:       gcp.StringPtr(diskType),
						ImageType:      gcp.StringPtr(imageType),
						Labels:         labels,
						LocalSsdCount:  gcp.Int64Ptr(localSsdCount),
						MachineType:    gcp.StringPtr(machineType),
						Metadata:       metadata,
						MinCPUPlatform: gcp.StringPtr(minCPUPlatform),
						OauthScopes:    oauthScopes,
						Preemptible:    gcp.BoolPtr(preemptible),
						SandboxConfig: &v1alpha1.SandboxConfig{
							SandboxType: "gvisor",
						},
						ServiceAccount: gcp.StringPtr(serviceAccount),
						ShieldedInstanceConfig: &v1alpha1.ShieldedInstanceConfig{
							EnableIntegrityMonitoring: &enable,
							EnableSecureBoot:          &enable,
						},
						Taints: []*v1alpha1.NodeTaint{taint},
						Tags:   tags,
					}

				}),
			},
			want: nodePool(func(n *container.NodePool) {
				n.Config = &container.NodeConfig{
					Accelerators:   []*container.AcceleratorConfig{gcpAccConf},
					DiskSizeGb:     diskSizeGb,
					DiskType:       diskType,
					ImageType:      imageType,
					Labels:         labels,
					LocalSsdCount:  localSsdCount,
					MachineType:    machineType,
					Metadata:       metadata,
					MinCpuPlatform: minCPUPlatform,
					OauthScopes:    oauthScopes,
					Preemptible:    preemptible,
					SandboxConfig: &container.SandboxConfig{
						SandboxType: "gvisor",
					},
					ServiceAccount: serviceAccount,
					ShieldedInstanceConfig: &container.ShieldedInstanceConfig{
						EnableIntegrityMonitoring: enable,
						EnableSecureBoot:          enable,
					},
					Taints: []*container.NodeTaint{gcpTaint},
					Tags:   tags,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				nodePool: &container.NodePool{},
				params:   params(),
			},
			want: nodePool(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateConfig(tc.args.params.Config, tc.args.nodePool)
			if diff := cmp.Diff(tc.want.Config, tc.args.nodePool.Config); diff != "" {
				t.Errorf("GenerateConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateManagement(t *testing.T) {
	enable := true

	type args struct {
		nodePool *container.NodePool
		params   *v1alpha1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.Management = &v1alpha1.NodeManagementSpec{
						AutoRepair:  &enable,
						AutoUpgrade: &enable,
					}
				}),
			},
			want: nodePool(func(n *container.NodePool) {
				n.Management = &container.NodeManagement{
					AutoRepair:  enable,
					AutoUpgrade: enable,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				nodePool: &container.NodePool{},
				params:   params(),
			},
			want: nodePool(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateManagement(tc.args.params.Management, tc.args.nodePool)
			if diff := cmp.Diff(tc.want.Management, tc.args.nodePool.Management); diff != "" {
				t.Errorf("GenerateManagement(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateMaxPodsConstraint(t *testing.T) {
	max := int64(3)

	type args struct {
		nodePool *container.NodePool
		params   *v1alpha1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.MaxPodsConstraint = &v1beta1.MaxPodsConstraint{
						MaxPodsPerNode: max,
					}
				}),
			},
			want: nodePool(func(n *container.NodePool) {
				n.MaxPodsConstraint = &container.MaxPodsConstraint{
					MaxPodsPerNode: max,
				}
			}),
		},
		"SuccessfulNil": {
			args: args{
				nodePool: &container.NodePool{},
				params:   params(),
			},
			want: nodePool(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			GenerateMaxPodsConstraint(tc.args.params.MaxPodsConstraint, tc.args.nodePool)
			if diff := cmp.Diff(tc.want.MaxPodsConstraint, tc.args.nodePool.MaxPodsConstraint); diff != "" {
				t.Errorf("GenerateMaxPodsConstraint(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		nodePool *container.NodePool
		params   *v1alpha1.NodePoolParameters
	}
	type want struct {
		params *v1alpha1.NodePoolParameters
	}
	tests := map[string]struct {
		args args
		want want
	}{
		"SomeFilled": {
			args: args{
				nodePool: nodePool(func(n *container.NodePool) {
					n.Autoscaling = &container.NodePoolAutoscaling{
						Autoprovisioned: true,
						Enabled:         true,
						MaxNodeCount:    3,
						MinNodeCount:    3,
					}
				}),
				params: params(),
			},
			want: want{
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.Autoscaling = &v1alpha1.NodePoolAutoscaling{
						Autoprovisioned: gcp.BoolPtr(true),
						Enabled:         gcp.BoolPtr(true),
						MaxNodeCount:    gcp.Int64Ptr(3),
						MinNodeCount:    gcp.Int64Ptr(3),
					}
				}),
			},
		},
		"NoneFilled": {
			args: args{
				nodePool: nodePool(),
				params:   params(),
			},
			want: want{
				params: params(),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.params, *tc.args.nodePool)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	falseVal := false

	type args struct {
		name     string
		nodePool *container.NodePool
		params   *v1alpha1.NodePoolParameters
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
				name:     name,
				nodePool: nodePool(),
				params:   params(),
			},
			want: want{
				upToDate: true,
				isErr:    false,
			},
		},
		"UpToDateWithOutputFields": {
			args: args{
				name:     name,
				nodePool: nodePool(addOutputFields),
				params:   params(),
			},
			want: want{
				upToDate: true,
				isErr:    false,
			},
		},
		"UpToDateIgnoreGVisor": {
			args: args{
				name: name,
				nodePool: nodePool(addOutputFields, func(n *container.NodePool) {
					n.Config = &container.NodeConfig{
						Labels: map[string]string{
							"cool-key": "cool-value",
							runtimeKey: "any value here ignored",
						},
						Taints: []*container.NodeTaint{
							{
								Key:   "cool-key",
								Value: "cool-value",
							},
							{
								Key:   runtimeKey,
								Value: "any value here ignored",
							},
						},
					}
				}),
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.Config = &v1alpha1.NodeConfig{
						Labels: map[string]string{"cool-key": "cool-value"},
						Taints: []*v1alpha1.NodeTaint{
							{
								Key:   "cool-key",
								Value: "cool-value",
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
		"NeedsUpdate": {
			args: args{
				name: name,
				nodePool: nodePool(func(n *container.NodePool) {
					n.Autoscaling = &container.NodePoolAutoscaling{
						Autoprovisioned: true,
						Enabled:         true,
						MaxNodeCount:    3,
						MinNodeCount:    3,
					}
				}),
				params: params(func(p *v1alpha1.NodePoolParameters) {
					p.Autoscaling = &v1alpha1.NodePoolAutoscaling{
						Autoprovisioned: &falseVal,
					}
				}),
			},
			want: want{
				upToDate: false,
				isErr:    false,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r, _, err := IsUpToDate(tc.args.name, tc.args.params, tc.args.nodePool)
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want upToDate, +got upToDate:\n%s", diff)
			}
		})
	}
}

func TestGetFullyQualifiedName(t *testing.T) {
	type args struct {
		params v1alpha1.NodePoolParameters
		name   string
	}
	tests := map[string]struct {
		args args
		want string
	}{
		"Successful": {
			args: args{
				params: *params(),
				name:   name,
			},
			want: fmt.Sprintf(NodePoolNameFormat, cluster, name),
		},
		"SuccessfulZonal": {
			args: args{
				params: *params(func(p *v1alpha1.NodePoolParameters) {
					p.Cluster = zonalCluster
				}),
				name: name,
			},
			want: fmt.Sprintf(NodePoolNameFormat, zonalClusterFormatted, name),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := GetFullyQualifiedName(tc.args.params, tc.args.name)
			if diff := cmp.Diff(tc.want, s); diff != "" {
				t.Errorf("GetFullyQualifiedName(...): -want, +got:\n%s", diff)
			}
		})
	}
}
