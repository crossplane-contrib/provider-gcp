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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// netowork endpoint types statuses.
const (
	GceVMIPPort         = "GCE_VM_IP_PORT"
	NonGgpPrivateIPPort = "NON_GCP_PRIVATE_IP_PORT"
	InternetFQDNPort    = "INTERNET_FQDN_PORT"
	InternetIPPort      = "INTERNET_IP_PORT"
	Serverless          = "SERVERLESS"
)

// NetworkEndpointGroupParameters define the desired state of a Google Compute Engine
// Global Address. Most fields map directly to an Address:
// https://cloud.google.com/compute/docs/reference/rest/v1/networkEndpointGroups
type NetworkEndpointGroupParameters struct {

	// Description: An optional description of this resource.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// Region : The URL of the region where the network endpoint group is located.
	Region *string `json:"region,omitempty"`

	// Zone : The URL of the zone where the network endpoint group is located.
	Zone *string `json:"zone,omitempty"`

	// NetworkEndpointType: Type of network endpoints in this network endpoint group. Can be one of GCE_VM_IP_PORT, NON_GCP_PRIVATE_IP_PORT, INTERNET_FQDN_PORT, INTERNET_IP_PORT, or SERVERLESS.
	//
	// Possible values:
	//   "GCE_VM_IP_PORT"
	//   "NON_GCP_PRIVATE_IP_PORT"
	//   "INTERNET_FQDN_PORT"
	//   "INTERNET_IP_PORT"
	//   "SERVERLESS"
	// +immutable
	// +kubebuilder:validation:Enum=GCE_VM_IP_PORT;NON_GCP_PRIVATE_IP_PORT;INTERNET_FQDN_PORT;INTERNET_IP_PORT;SERVERLESS
	NetworkEndpointType string `json:"networkEndpointType"`

	// Network: The URL of the network to which all network endpoints in the NEG belong. Uses "default" project network if unspecified.
	// +optional
	// +immutable
	Network *string `json:"network,omitempty"`

	// Subnetwork: Optional URL of the subnetwork to which all network endpoints in the NEG belong..
	// +optional
	// +immutable
	Subnetwork *string `json:"subnetwork,omitempty"`

	// DefaultPort: The default port used if the port number is not specified in the network endpoint.
	// +immutable
	DefaultPort int64 `json:"defaultPort"`

	// CloudRun : Only valid when networkEndpointType is "SERVERLESS". Only one of cloudRun, appEngine or cloudFunction may be set.
	// +optional
	// +immutable
	CloudRun *CloudRunParameters `json:"cloudRun,omitempty"`

	// AppEngine : Only valid when networkEndpointType is "SERVERLESS". Only one of cloudRun, appEngine or cloudFunction may be set.
	// +optional
	// +immutable
	AppEngine *AppEngineParameters `json:"appEngine,omitempty"`

	// CloudFunction : Only valid when networkEndpointType is "SERVERLESS". Only one of cloudRun, appEngine or cloudFunction may be set.
	// +optional
	// +immutable
	CloudFunction *CloudFunctionParameters `json:"cloudFunction,omitempty"`

	// NetworkEndpoints : list of network endpoints for this network endpoint group. List elements are immutable, but they can be added/removed.
	// +optional
	NetworkEndpoints []NetworkEndpoint `json:"networkEndpoints,omitempty"`
}

// NetworkEndpoint : represents a single network endpoint
type NetworkEndpoint struct {
	// IPAddress : Optional IPv4 address of network endpoint. The IP address must belong to a VM in Compute Engine (either the primary IP or as part of an aliased IP range). If the IP address is not specified, then the primary IP address for the VM instance in the network that the network endpoint group belongs to will be used.
	// +optional
	// +immutable
	IPAddress *string `json:"ipAddress,omitempty"`

	// FQDN : Optional fully qualified domain name of network endpoint. This can only be specified when NetworkEndpointGroup.network_endpoint_type is NON_GCP_FQDN_PORT.
	// +optional
	// +immutable
	FQDN *string `json:"fqdn,omitempty"`

	// Port : Optional port number of network endpoint. If not specified, the defaultPort for the network endpoint group will be used.
	// +optional
	// +immutable
	Port *int64 `json:"port,omitempty"`

	// Instance : The name for a specific VM instance that the IP address belongs to. This is required for network endpoints of type GCE_VM_IP_PORT. The instance must be in the same zone of network endpoint group.
	// The name must be 1-63 characters long, and comply with RFC1035.
	// Authorization requires the following IAM permission on the specified resource instance:
	// compute.instances.use
	// +optional
	// +immutable
	Instance *string `json:"instance,omitempty"`
}

// NetworkEndpointHealth : health status of a network endpoint
type NetworkEndpointHealth struct {
	// ForwardingRule : URL of the forwarding rule associated with the health state of the network endpoint.
	ForwardingRule NetworkEndpointForwardingRule `json:"forwardingRule,omitempty"`

	// BackendService : URL of the backend service associated with the health state of the network endpoint.
	BackendService NetworkEndpointBackendService `json:"backendService,omitempty"`

	// HealthCheck : URL of the health check associated with the health state of the network endpoint.
	HealthCheck NetworkEndpointHealthCheck `json:"healthCheck,omitempty"`

	// HealthCheckService : URL of the health check service associated with the health state of the network endpoint.
	HealthCheckService NetworkEndpointHealthCheckService `json:"healthCheckService,omitempty"`

	// HealthState : Health state of the network endpoint determined based on the health checks configured.
	HealthState string `json:"healthState,omitempty"`
}

// NetworkEndpointForwardingRule : forwarding rule that indirectly references this endpoint
type NetworkEndpointForwardingRule struct {
	// ForwardingRule : URL of the forwarding rule associated with the health state of the network endpoint.
	ForwardingRule string `json:"forwardingRule,omitempty"`
}

// NetworkEndpointBackendService : forwarding rule that indirectly references this endpoint
type NetworkEndpointBackendService struct {
	// BackendService : URL of the backend service associated with the health state of the network endpoint.
	BackendService string `json:"backendService,omitempty"`
}

// NetworkEndpointHealthCheck : forwarding rule that references this endpoint
type NetworkEndpointHealthCheck struct {
	// HealthCheck : URL of the health check associated with the health state of the network endpoint.
	HealthCheck string `json:"healthCheck,omitempty"`
}

// NetworkEndpointHealthCheckService : forwarding rule that indirectly references this endpoint
type NetworkEndpointHealthCheckService struct {
	// HealthCheckService : URL of the health check service associated with the health state of the network endpoint.
	HealthCheckService string `json:"healthCheckService,omitempty"`
}

// NetworkEndpointObservation : an observation of a network endpoint
type NetworkEndpointObservation struct {
	NetworkEndpoint        `json:",inline"`
	Annotations            map[string]string       `json:"annotations,omitempty"`
	NetworkEndpointHealths []NetworkEndpointHealth `json:"networkEndpointHealths,omitempty"`
}

// CloudRunParameters : fill these if you load balance to a cloud run function
type CloudRunParameters struct {
	// Service : Cloud Run service is the main resource of Cloud Run.
	// The service must be 1-63 characters long, and comply with RFC1035.
	// Example value: "run-service".
	Service string `json:"service"`

	// Tag : Optional Cloud Run tag represents the "named-revision" to provide additional fine-grained traffic routing information.
	// The tag must be 1-63 characters long, and comply with RFC1035.
	// Example value: "revision-0010".
	// +optional
	// +immutable
	Tag *string `json:"tag,omitempty"`

	// URLMask : A template to parse service and tag fields from a request URL. URL mask allows for routing to multiple Run services without having to create multiple network endpoint groups and backend services.
	// For example, request URLs "foo1.domain.com/bar1" and "foo1.domain.com/bar2" can be backed by the same Serverless Network Endpoint Group (NEG) with URL mask ".domain.com/". The URL mask will parse them to { service="bar1", tag="foo1" } and { service="bar2", tag="foo2" } respectively.
	// +optional
	// +immutable
	URLMask *string `json:"urlMask,omitempty"`
}

// AppEngineParameters : fill these if you load balance to a app engine function
type AppEngineParameters struct {
	// Service : Optional serving service.
	// The service name is case-sensitive and must be 1-63 characters long.
	// Example value: "default", "my-service".
	// +optional
	// +immutable
	Service *string `json:"service,omitempty"`

	// Version : Optional serving version.
	// The version name is case-sensitive and must be 1-100 characters long.
	// Example value: "v1", "v2".
	// +optional
	// +immutable
	Version *string `json:"version,omitempty"`

	// URLMask : A template to parse service and version fields from a request URL. URL mask allows for routing to multiple App Engine services without having to create multiple Network Endpoint Groups and backend services.
	// For example, the request URLs "foo1-dot-appname.appspot.com/v1" and "foo1-dot-appname.appspot.com/v2" can be backed by the same Serverless NEG with URL mask "-dot-appname.appspot.com/". The URL mask will parse them to { service = "foo1", version = "v1" } and { service = "foo1", version = "v2" } respectively.
	// +optional
	// +immutable
	URLMask *string `json:"urlMask,omitempty"`
}

// CloudFunctionParameters : fill these if you load balance to a cloud function
type CloudFunctionParameters struct {
	// Function : A user-defined name of the Cloud Function.
	// The function name is case-sensitive and must be 1-63 characters long.
	// Example value: "func1".
	// +immutable
	Function string `json:"function"`

	// URLMask : A template to parse function field from a request URL. URL mask allows for routing to multiple Cloud Functions without having to create multiple Network Endpoint Groups and backend services.
	// For example, request URLs "mydomain.com/function1" and "mydomain.com/function2" can be backed by the same Serverless NEG with URL mask "/". The URL mask will parse them to { function = "function1" } and { function = "function2" } respectively.
	// +optional
	// +immutable
	URLMask *string `json:"urlMask,omitempty"`
}

// A NetworkEndpointGroupObservation reflects the observed state of a GlobalAddress on GCP.
type NetworkEndpointGroupObservation struct {
	// CreationTimestamp in RFC3339 text format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// ID for the resource. This identifier is defined by the server.
	ID uint64 `json:"id,omitempty"`

	// SelfLink: Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`

	// Size: Number of network endpoints in the network endpoint group.
	Size int64 `json:"size,omitempty"`

	NetworkEndpointObservations []NetworkEndpointObservation `json:"networkEndpointObservations,omitempty"`
}

// A NetworkEndpointGroupSpec defines the desired state of a NetworkEndpointGroup.
type NetworkEndpointGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       NetworkEndpointGroupParameters `json:"forProvider"`
}

// A NetworkEndpointGroupStatus represents the observed state of a NetworkEndpointGroup.
type NetworkEndpointGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          NetworkEndpointGroupObservation `json:"atProvider,omitempty"`
}

// A NetworkEndpointGroup is a managed resource that represents a Google Compute Engine
// Network Endpoint Group (https://cloud.google.com/compute/docs/reference/rest/v1/networkEndpointGroups).
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type NetworkEndpointGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkEndpointGroupSpec   `json:"spec"`
	Status NetworkEndpointGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkEndpointGroupList contains a list of NetworkEndpointGroups.
type NetworkEndpointGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkEndpointGroup `json:"items"`
}
