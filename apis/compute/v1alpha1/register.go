/*
Copyright 2021 The Crossplane Authors.

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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "compute.gcp.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Router type metadata.
var (
	RouterKind             = reflect.TypeOf(Router{}).Name()
	RouterGroupKind        = schema.GroupKind{Group: Group, Kind: RouterKind}.String()
	RouterKindAPIVersion   = RouterKind + "." + SchemeGroupVersion.String()
	RouterGroupVersionKind = SchemeGroupVersion.WithKind(RouterKind)
)

// HealthCheck type metadata.
var (
	HealthCheckKind             = reflect.TypeOf(HealthCheck{}).Name()
	HealthCheckGroupKind        = schema.GroupKind{Group: Group, Kind: HealthCheckKind}.String()
	HealthCheckKindAPIVersion   = HealthCheckKind + "." + SchemeGroupVersion.String()
	HealthCheckGroupVersionKind = SchemeGroupVersion.WithKind(HealthCheckKind)
)

// BackendService type metadata.
var (
	BackendServiceKind             = reflect.TypeOf(BackendService{}).Name()
	BackendServiceGroupKind        = schema.GroupKind{Group: Group, Kind: BackendServiceKind}.String()
	BackendServiceKindAPIVersion   = BackendServiceKind + "." + SchemeGroupVersion.String()
	BackendServiceGroupVersionKind = SchemeGroupVersion.WithKind(BackendServiceKind)
)

// TargetTcpProxy type metadata.
var (
	TargetTcpProxyKind             = reflect.TypeOf(TargetTcpProxy{}).Name()
	TargetTcpProxyGroupKind        = schema.GroupKind{Group: Group, Kind: TargetTcpProxyKind}.String()
	TargetTcpProxyKindAPIVersion   = TargetTcpProxyKind + "." + SchemeGroupVersion.String()
	TargetTcpProxyGroupVersionKind = SchemeGroupVersion.WithKind(TargetTcpProxyKind)
)

// ForwardingRule type metadata.
var (
	ForwardingRuleKind             = reflect.TypeOf(ForwardingRule{}).Name()
	ForwardingRuleGroupKind        = schema.GroupKind{Group: Group, Kind: ForwardingRuleKind}.String()
	ForwardingRuleKindAPIVersion   = ForwardingRuleKind + "." + SchemeGroupVersion.String()
	ForwardingRuleGroupVersionKind = SchemeGroupVersion.WithKind(ForwardingRuleKind)
)

// InstanceGroup type metadata.
var (
	InstanceGroupKind             = reflect.TypeOf(InstanceGroup{}).Name()
	InstanceGroupGroupKind        = schema.GroupKind{Group: Group, Kind: InstanceGroupKind}.String()
	InstanceGroupKindAPIVersion   = InstanceGroupKind + "." + SchemeGroupVersion.String()
	InstanceGroupGroupVersionKind = SchemeGroupVersion.WithKind(InstanceGroupKind)
)

// Firewall type metadata.
var (
	FirewallKind             = reflect.TypeOf(Firewall{}).Name()
	FirewallGroupKind        = schema.GroupKind{Group: Group, Kind: FirewallKind}.String()
	FirewallKindAPIVersion   = FirewallKind + "." + SchemeGroupVersion.String()
	FirewallGroupVersionKind = SchemeGroupVersion.WithKind(FirewallKind)
)

func init() {
	SchemeBuilder.Register(&Router{}, &RouterList{})
	SchemeBuilder.Register(&HealthCheck{}, &HealthCheckList{})
	SchemeBuilder.Register(&BackendService{}, &BackendServiceList{})
	SchemeBuilder.Register(&TargetTcpProxy{}, &TargetTcpProxyList{})
	SchemeBuilder.Register(&ForwardingRule{}, &ForwardingRuleList{})
	SchemeBuilder.Register(&InstanceGroup{}, &InstanceGroupList{})
	SchemeBuilder.Register(&Firewall{}, &FirewallList{})
}
