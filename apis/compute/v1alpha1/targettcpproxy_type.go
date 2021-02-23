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

package v1alpha1

import (
	compute "google.golang.org/api/compute/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

var _ = compute.TargetTcpProxy{}

// TargetTcpProxyParameters define the desired state of a Google Compute Engine VPC
// TargetTcpProxy. Most fields map directly to a TargetTcpProxy:
// https://cloud.google.com/compute/docs/reference/rest/v1/networks
type TargetTcpProxyParameters struct {
	// Description: An optional description of this resource. Provide this
	// property when you create the resource.
	Description string `json:"description,omitempty"`

	// ProxyHeader: Specifies the type of proxy header to append before
	// sending data to the backend, either NONE or PROXY_V1. The default is
	// NONE.
	//
	// Possible values:
	//   "NONE"
	//   "PROXY_V1"
	ProxyHeader string `json:"proxyHeader,omitempty"`

	// Service: URL of the network to which this router belongs.
	// +optional
	// +immutable
	Service *string `json:"service,omitempty"`

	// ServiceRef references a BackendService
	// +optional
	// +immutable
	ServiceRef *xpv1.Reference `json:"serviceRef,omitempty"`

	// ServiceSelector selects a reference to a BackendService
	// +optional
	// +immutable
	ServiceSelector *xpv1.Selector `json:"serviceSelector,omitempty"`
}

// A TargetTcpProxyObservation represents the observed state of a Google Compute Engine
// VPC TargetTcpProxy.
type TargetTcpProxyObservation struct {
	// CreationTimestamp: [Output Only] Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// Id: [Output Only] The unique identifier for the resource. This
	// identifier is defined by the server.
	ID int64 `json:"id,omitempty,string"`

	// Kind: [Output Only] Type of the resource. Always
	// compute#targetTcpProxy for target TCP proxies.
	Kind string `json:"kind,omitempty"`

	// SelfLink: [Output Only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`
}

// A TargetTcpProxySpec defines the desired state of a TargetTcpProxy.
type TargetTcpProxySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       TargetTcpProxyParameters `json:"forProvider"`
}

// A TargetTcpProxyStatus represents the observed state of a TargetTcpProxy.
type TargetTcpProxyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          TargetTcpProxyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A TargetTcpProxy is a managed resource that represents a Google Compute Engine VPC
// TargetTcpProxy.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type TargetTcpProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetTcpProxySpec   `json:"spec"`
	Status TargetTcpProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TargetTcpProxyList contains a list of TargetTcpProxy.
type TargetTcpProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetTcpProxy `json:"items"`
}
