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

	"github.com/crossplaneio/stack-gcp/apis/container/v1beta1"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	"github.com/google/go-cmp/cmp"
	container "google.golang.org/api/container/v1beta1"
)

const (
	name    = "my-cool-pool"
	cluster = "/projects/cool-proj/locations/us-central1/clusters/cool-cluster"
)

var (
	resourceLabels = map[string]string{"label": "one"}
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

func params(m ...func(*v1beta1.NodePoolParameters)) *v1beta1.NodePoolParameters {
	p := &v1beta1.NodePoolParameters{
		Cluster:          cluster,
		InitialNodeCount: gcp.Int64Ptr(3),
		Locations:        []string{"us-central1-a"},
	}
	for _, f := range m {
		f(p)
	}

	return p
}

func observation(m ...func(*v1beta1.NodePoolObservation)) *v1beta1.NodePoolObservation {
	o := &v1beta1.NodePoolObservation{
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
		Management: &v1beta1.NodeManagementStatus{
			UpgradeOptions: &v1beta1.AutoUpgradeOptions{
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
		want *v1beta1.NodePoolObservation
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
			want: observation(func(p *v1beta1.NodePoolObservation) {
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
		params   *v1beta1.NodePoolParameters
		name     string
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: nodePool(),
				params: params(func(p *v1beta1.NodePoolParameters) {
					p.Autoscaling = &v1beta1.NodePoolAutoscaling{
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
			nodePool := GenerateNodePool(*tc.args.params, tc.args.name)
			if diff := cmp.Diff(tc.want, nodePool); diff != "" {
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
		params   *v1beta1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1beta1.NodePoolParameters) {
					p.Autoscaling = &v1beta1.NodePoolAutoscaling{
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

	type args struct {
		nodePool *container.NodePool
		params   *v1beta1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1beta1.NodePoolParameters) {
					p.Config = &v1beta1.NodeConfig{
						Accelerators: []*v1beta1.AcceleratorConfig{
							&v1beta1.AcceleratorConfig{
								AcceleratorCount: 3,
								AcceleratorType:  "nvidia-tesla-t4",
							},
						},
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
						SandboxConfig: &v1beta1.SandboxConfig{
							SandboxType: "gvisor",
						},
						ServiceAccount: gcp.StringPtr(serviceAccount),
						ShieldedInstanceConfig: &v1beta1.ShieldedInstanceConfig{
							EnableIntegrityMonitoring: &enable,
							EnableSecureBoot:          &enable,
						},
						Taints: []*v1beta1.NodeTaint{
							&v1beta1.NodeTaint{
								Effect: "NO_SCHEDULE",
								Key:    "cool-key",
								Value:  "cool-val",
							},
						},
						Tags: tags,
					}

				}),
			},
			want: nodePool(func(n *container.NodePool) {
				n.Config = &container.NodeConfig{
					Accelerators: []*container.AcceleratorConfig{
						&container.AcceleratorConfig{
							AcceleratorCount: 3,
							AcceleratorType:  "nvidia-tesla-t4",
						},
					},
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
					Taints: []*container.NodeTaint{
						&container.NodeTaint{
							Effect: "NO_SCHEDULE",
							Key:    "cool-key",
							Value:  "cool-val",
						},
					},
					Tags: tags,
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
		params   *v1beta1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1beta1.NodePoolParameters) {
					p.Management = &v1beta1.NodeManagementSpec{
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
		params   *v1beta1.NodePoolParameters
	}

	tests := map[string]struct {
		args args
		want *container.NodePool
	}{
		"Successful": {
			args: args{
				nodePool: &container.NodePool{},
				params: params(func(p *v1beta1.NodePoolParameters) {
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
		params   *v1beta1.NodePoolParameters
	}
	type want struct {
		params *v1beta1.NodePoolParameters
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
				params: params(func(p *v1beta1.NodePoolParameters) {
					p.Autoscaling = &v1beta1.NodePoolAutoscaling{
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

// func TestIsUpToDate(t *testing.T) {
// 	type args struct {
// 		cluster *container.NodePool
// 		params  *v1beta1.NodePoolParameters
// 	}
// 	type want struct {
// 		upToDate bool
// 		kind     UpdateKind
// 	}
// 	tests := map[string]struct {
// 		args args
// 		want want
// 	}{
// 		"UpToDate": {
// 			args: args{
// 				cluster: cluster(),
// 				params:  params(),
// 			},
// 			want: want{
// 				upToDate: true,
// 			},
// 		},
// 		"UpToDateIgnoreRefs": {
// 			args: args{
// 				cluster: cluster(),
// 				params: params(func(p *v1beta1.NodePoolParameters) {
// 					p.NetworkRef = &v1beta1.NetworkURIReferencerForGKECluster{
// 						NetworkURIReferencer: v1alpha3.NetworkURIReferencer{
// 							LocalObjectReference: corev1.LocalObjectReference{
// 								Name: "my-network",
// 							},
// 						},
// 					}
// 				}),
// 			},
// 			want: want{
// 				upToDate: true,
// 			},
// 		},
// 		"NeedsUpdate": {
// 			args: args{
// 				cluster: cluster(func(n *container.NodePool) {
// 					c.AddonsConfig = &container.AddonsConfig{
// 						HttpLoadBalancing: &container.HttpLoadBalancing{
// 							Disabled: true,
// 						},
// 					}
// 					c.IpAllocationPolicy = &container.IPAllocationPolicy{
// 						ClusterIpv4CidrBlock: "0.0.0.0/0",
// 					}
// 				}),
// 				params: params(),
// 			},
// 			want: want{
// 				upToDate: false,
// 				kind:     AddonsConfigUpdate,
// 			},
// 		},
// 		"NoUpdateNotBootstrapNodePool": {
// 			args: args{
// 				cluster: cluster(func(n *container.NodePool) {
// 					sc := &container.StatusCondition{
// 						Code:    "cool-code",
// 						Message: "cool-message",
// 					}
// 					np := &container.NodePool{
// 						Conditions: []*container.StatusCondition{sc},
// 						Name:       "cool-node-pool",
// 					}
// 					c.NodePools = []*container.NodePool{np}
// 				}),
// 				params: params(),
// 			},
// 			want: want{
// 				upToDate: true,
// 			},
// 		},
// 		"NeedsUpdateBootstrapNodePool": {
// 			args: args{
// 				cluster: cluster(func(n *container.NodePool) {
// 					sc := &container.StatusCondition{
// 						Code:    "cool-code",
// 						Message: "cool-message",
// 					}
// 					np := &container.NodePool{
// 						Conditions: []*container.StatusCondition{sc},
// 						Name:       BootstrapNodePoolName,
// 					}
// 					c.NodePools = []*container.NodePool{np}
// 				}),
// 				params: params(),
// 			},
// 			want: want{
// 				upToDate: false,
// 				kind:     NodePoolUpdate,
// 			},
// 		},
// 	}
// 	for name, tc := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			r, k := IsUpToDate(tc.args.params, *tc.args.cluster)
// 			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
// 				t.Errorf("IsUpToDate(...): -want upToDate, +got upToDate:\n%s", diff)
// 			}
// 			if diff := cmp.Diff(tc.want.kind, k); diff != "" {
// 				t.Errorf("IsUpToDate(...): -want kind, +got kind:\n%s", diff)
// 			}
// 		})
// 	}
// }

func TestGetFullyQualifiedName(t *testing.T) {
	type args struct {
		params v1beta1.NodePoolParameters
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
