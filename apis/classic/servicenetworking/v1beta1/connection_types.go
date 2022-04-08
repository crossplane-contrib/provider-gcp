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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ConnectionParameters define the desired state of a Google Cloud Service
// Networking Connection. Most fields map direct to a Connection:
// https://cloud.google.com/service-infrastructure/docs/service-networking/reference/rest/v1/services.connections#Connection
type ConnectionParameters struct {
	// Parent: The service that is managing peering connectivity for a service
	// producer's organization. For Google services that support this
	// functionality, this value is services/servicenetworking.googleapis.com.
	// +immutable
	Parent string `json:"parent"`

	// Network: The name of service consumer's VPC network that's connected
	// with service producer network, in the following format:
	// `projects/{project}/global/networks/{network}`.
	// `{project}` is a project number, such as in `12345` that includes
	// the VPC service consumer's VPC network. `{network}` is the name of
	// the service consumer's VPC network.
	// +optional
	Network *string `json:"network,omitempty"`

	// NetworkRef references a Network and retrieves its URI
	// +optional
	NetworkRef *xpv1.Reference `json:"networkRef,omitempty"`

	// NetworkSelector selects a reference to a Network and retrieves its URI
	// +optional
	NetworkSelector *xpv1.Selector `json:"networkSelector,omitempty"`

	// ReservedPeeringRanges: The name of one or more allocated IP address
	// ranges for this service producer of type `PEERING`.
	// +optional
	ReservedPeeringRanges []string `json:"reservedPeeringRanges,omitempty"`

	// ReservedPeeringRangeRefs is a set of references to GlobalAddress objects
	// +optional
	ReservedPeeringRangeRefs []xpv1.Reference `json:"reservedPeeringRangeRefs,omitempty"`

	// ReservedPeeringRangeSelector selects a set of references to GlobalAddress
	// objects.
	// +optional
	ReservedPeeringRangeSelector xpv1.Selector `json:"reservedPeeringRangeSelector,omitempty"`
}

// ConnectionObservation is used to show the observed state of the Connection.
type ConnectionObservation struct {
	// Peering: The name of the VPC Network Peering connection that was created
	// by the service producer.
	Peering string `json:"peering,omitempty"`

	// Service: The name of the peering service that's associated with this
	// connection, in the following format: `services/{service name}`.
	Service string `json:"service,omitempty"`
}

// A ConnectionSpec defines the desired state of a Connection.
type ConnectionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ConnectionParameters `json:"forProvider"`
}

// A ConnectionStatus represents the observed state of a Connection.
type ConnectionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ConnectionObservation `json:"atProvider,omitempty"`
}

// A Connection is a managed resource that represents a Google Cloud Service
// Networking Connection.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec   `json:"spec"`
	Status ConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionList contains a list of Connection.
type ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connection `json:"items"`
}
