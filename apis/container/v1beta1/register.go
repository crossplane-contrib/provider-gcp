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

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)

// Package type metadata.
const (
	Group   = "container.gcp.crossplane.io"
	Version = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// GKECluster type metadata.
var (
	GKEClusterKind             = reflect.TypeOf(GKECluster{}).Name()
	GKEClusterKindAPIVersion   = GKEClusterKind + "." + SchemeGroupVersion.String()
	GKEClusterGroupVersionKind = SchemeGroupVersion.WithKind(GKEClusterKind)
)

// GKEClusterClass type metadata.
var (
	GKEClusterClassKind             = reflect.TypeOf(GKEClusterClass{}).Name()
	GKEClusterClassKindAPIVersion   = GKEClusterClassKind + "." + SchemeGroupVersion.String()
	GKEClusterClassGroupVersionKind = SchemeGroupVersion.WithKind(GKEClusterClassKind)
)

// NodePool type metadata.
var (
	NodePoolKind             = reflect.TypeOf(NodePool{}).Name()
	NodePoolKindAPIVersion   = NodePoolKind + "." + SchemeGroupVersion.String()
	NodePoolGroupVersionKind = SchemeGroupVersion.WithKind(NodePoolKind)
)

// NodePoolClass type metadata.
var (
	NodePoolClassKind             = reflect.TypeOf(NodePoolClass{}).Name()
	NodePoolClassKindAPIVersion   = NodePoolClassKind + "." + SchemeGroupVersion.String()
	NodePoolClassGroupVersionKind = SchemeGroupVersion.WithKind(NodePoolClassKind)
)

func init() {
	SchemeBuilder.Register(&GKECluster{}, &GKEClusterList{})
	SchemeBuilder.Register(&GKEClusterClass{}, &GKEClusterClassList{})
	SchemeBuilder.Register(&NodePool{}, &NodePoolList{})
	SchemeBuilder.Register(&NodePoolClass{}, &NodePoolClassList{})
}
