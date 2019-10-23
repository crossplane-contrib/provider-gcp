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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
)

// A ConnectionSpec defines the desired state of a Connection.
type ConnectionSpec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ConnectionParameters  `json:",inline"`
}

// ConnectionParameters define the desired state of a Google Cloud Service
// Networking Connection. Most fields map direct to a Connection:
// https://cloud.google.com/service-infrastructure/docs/service-networking/reference/rest/v1/services.connections#Connection
type ConnectionParameters struct {
	// Parent: The service that is managing peering connectivity for a service
	// producer's organization. For Google services that support this
	// functionality, this value is services/servicenetworking.googleapis.com.
	Parent string `json:"parent"`

	// Network: The name of service consumer's VPC network that's connected
	// with service producer network, in the following format:
	// `projects/{project}/global/networks/{network}`.
	// `{project}` is a project number, such as in `12345` that includes
	// the VPC service consumer's VPC network. `{network}` is the name of
	// the service consumer's VPC network.
	Network string `json:"network"`

	// ReservedPeeringRanges: The name of one or more allocated IP address
	// ranges for this service producer of type `PEERING`.
	// +optional
	ReservedPeeringRanges []string `json:"reservedPeeringRanges"`
}

// A ConnectionStatus represents the observed state of a Connection.
type ConnectionStatus struct {
	v1alpha1.ResourceStatus `json:",inline"`

	// Peering: The name of the VPC Network Peering connection that was created
	// by the service producer.
	Peering string `json:"peering,omitempty"`

	// Service: The name of the peering service that's associated with this
	// connection, in the following format: `services/{service name}`.
	Service string `json:"service,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// A Connection is a managed resource that represents a Google Cloud Service
// Networking Connection.
type Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec   `json:"spec,omitempty"`
	Status ConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionList contains a list of Connection.
type ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connection `json:"items"`
}
