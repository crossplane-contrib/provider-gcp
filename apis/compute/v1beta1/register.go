/*
Copyright 2020 The Crossplane Authors.

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

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "compute.gcp.crossplane.io"
	Version = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Network type metadata.
var (
	NetworkKind             = reflect.TypeOf(Network{}).Name()
	NetworkGroupKind        = schema.GroupKind{Group: Group, Kind: NetworkKind}.String()
	NetworkKindAPIVersion   = NetworkKind + "." + SchemeGroupVersion.String()
	NetworkGroupVersionKind = SchemeGroupVersion.WithKind(NetworkKind)
)

// Subnetwork type metadata.
var (
	SubnetworkKind             = reflect.TypeOf(Subnetwork{}).Name()
	SubnetworkGroupKind        = schema.GroupKind{Group: Group, Kind: SubnetworkKind}.String()
	SubnetworkKindAPIVersion   = SubnetworkKind + "." + SchemeGroupVersion.String()
	SubnetworkGroupVersionKind = SchemeGroupVersion.WithKind(SubnetworkKind)
)

// GlobalAddress type metadata.
var (
	GlobalAddressKind             = reflect.TypeOf(GlobalAddress{}).Name()
	GlobalAddressGroupKind        = schema.GroupKind{Group: Group, Kind: GlobalAddressKind}.String()
	GlobalAddressKindAPIVersion   = GlobalAddressKind + "." + SchemeGroupVersion.String()
	GlobalAddressGroupVersionKind = SchemeGroupVersion.WithKind(GlobalAddressKind)
)

// GlobalAddress type metadata.
var (
	NetworkEndpointGroupKind             = reflect.TypeOf(NetworkEndpointGroup{}).Name()
	NetworkEndpointGroupGroupKind        = schema.GroupKind{Group: Group, Kind: NetworkEndpointGroupKind}.String()
	NetworkEndpointGroupKindAPIVersion   = NetworkEndpointGroupKind + "." + SchemeGroupVersion.String()
	NetworkEndpointGroupGroupVersionKind = SchemeGroupVersion.WithKind(NetworkEndpointGroupKind)
)

func init() {
	SchemeBuilder.Register(&Network{}, &NetworkList{})
	SchemeBuilder.Register(&Subnetwork{}, &SubnetworkList{})
	SchemeBuilder.Register(&GlobalAddress{}, &GlobalAddressList{})
	SchemeBuilder.Register(&NetworkEndpointGroup{}, &NetworkEndpointGroupList{})
}
