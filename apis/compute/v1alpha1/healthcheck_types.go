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

var _ = compute.HealthCheck{}

// HealthCheckParameters define the desired state of a Google Compute Engine VPC
// HealthCheck. Most fields map directly to a HealthCheck:
// https://cloud.google.com/compute/docs/reference/rest/v1/networks
type HealthCheckParameters struct {
	// CheckIntervalSec: How often (in seconds) to send a health check. The
	// default value is 5 seconds.
	// +optional
	CheckIntervalSec *int64 `json:"checkIntervalSec,omitempty"`

	// Description: An optional description of this resource. Provide this
	// property when you create the resource.
	// +optional
	// +immutable
	Description *string `json:"description,omitempty"`

	// HealthyThreshold: A so-far unhealthy instance will be marked healthy
	// after this many consecutive successes. The default value is 2.
	// +optional
	HealthyThreshold *int64 `json:"healthyThreshold,omitempty"`

	// +optional
	Http2HealthCheck *HTTP2HealthCheck `json:"http2HealthCheck,omitempty"`

	// +optional
	HttpHealthCheck *HTTPHealthCheck `json:"httpHealthCheck,omitempty"`

	// +optional
	HttpsHealthCheck *HTTPSHealthCheck `json:"httpsHealthCheck,omitempty"`

	// Kind: Type of the resource.
	// +optional
	Kind *string `json:"kind,omitempty"`

	// +optional
	SslHealthCheck *SSLHealthCheck `json:"sslHealthCheck,omitempty"`

	// +optional
	TcpHealthCheck *TCPHealthCheck `json:"tcpHealthCheck,omitempty"`

	// TimeoutSec: How long (in seconds) to wait before claiming failure.
	// The default value is 5 seconds. It is invalid for timeoutSec to have
	// greater value than checkIntervalSec.
	// +optional
	TimeoutSec *int64 `json:"timeoutSec,omitempty"`

	// Type: Specifies the type of the healthCheck, either TCP, SSL, HTTP,
	// HTTPS or HTTP2. If not specified, the default is TCP. Exactly one of
	// the protocol-specific health check field must be specified, which
	// must match type field.
	//
	// Possible values:
	//   "HTTP"
	//   "HTTP2"
	//   "HTTPS"
	//   "INVALID"
	//   "SSL"
	//   "TCP"
	// +optional
	Type *string `json:"type,omitempty"`

	// UnhealthyThreshold: A so-far healthy instance will be marked
	// unhealthy after this many consecutive failures. The default value is
	// 2.
	// +optional
	UnhealthyThreshold *int64 `json:"unhealthyThreshold,omitempty"`
}

type HTTP2HealthCheck struct {
	// Host: The value of the host header in the HTTP/2 health check
	// request. If left empty (default value), the IP on behalf of which
	// this health check is performed will be used.
	// +optional
	Host *string `json:"host,omitempty"`

	// Port: The TCP port number for the health check request. The default
	// value is 443. Valid values are 1 through 65535.
	// +optional
	Port *int64 `json:"port,omitempty"`

	// PortName: Port name as defined in InstanceGroup#NamedPort#name. If
	// both port and port_name are defined, port takes precedence.
	// +optional
	PortName *string `json:"portName,omitempty"`

	// PortSpecification: Specifies how port is selected for health
	// checking, can be one of following values:
	// USE_FIXED_PORT: The port number in port is used for health
	// checking.
	// USE_NAMED_PORT: The portName is used for health
	// checking.
	// USE_SERVING_PORT: For NetworkEndpointGroup, the port specified for
	// each network endpoint is used for health checking. For other
	// backends, the port or named port specified in the Backend Service is
	// used for health checking.
	//
	//
	// If not specified, HTTP2 health check follows behavior specified in
	// port and portName fields.
	//
	// Possible values:
	//   "USE_FIXED_PORT"
	//   "USE_NAMED_PORT"
	//   "USE_SERVING_PORT"
	// +optional
	PortSpecification *string `json:"portSpecification,omitempty"`

	// ProxyHeader: Specifies the type of proxy header to append before
	// sending data to the backend, either NONE or PROXY_V1. The default is
	// NONE.
	//
	// Possible values:
	//   "NONE"
	//   "PROXY_V1"
	// +optional
	ProxyHeader *string `json:"proxyHeader,omitempty"`

	// RequestPath: The request path of the HTTP/2 health check request. The
	// default value is /.
	// +optional
	RequestPath *string `json:"requestPath,omitempty"`

	// Response: The string to match anywhere in the first 1024 bytes of the
	// response body. If left empty (the default value), the status code
	// determines health. The response data can only be ASCII.
	// +optional
	Response *string `json:"response,omitempty"`
}

type HTTPHealthCheck struct {
	// Host: The value of the host header in the HTTP health check request.
	// If left empty (default value), the IP on behalf of which this health
	// check is performed will be used.
	// +optional
	Host *string `json:"host,omitempty"`

	// Port: The TCP port number for the health check request. The default
	// value is 80. Valid values are 1 through 65535.
	// +optional
	Port *int64 `json:"port,omitempty"`

	// PortName: Port name as defined in InstanceGroup#NamedPort#name. If
	// both port and port_name are defined, port takes precedence.
	// +optional
	PortName *string `json:"portName,omitempty"`

	// PortSpecification: Specifies how port is selected for health
	// checking, can be one of following values:
	// USE_FIXED_PORT: The port number in port is used for health
	// checking.
	// USE_NAMED_PORT: The portName is used for health
	// checking.
	// USE_SERVING_PORT: For NetworkEndpointGroup, the port specified for
	// each network endpoint is used for health checking. For other
	// backends, the port or named port specified in the Backend Service is
	// used for health checking.
	//
	//
	// If not specified, HTTP health check follows behavior specified in
	// port and portName fields.
	//
	// Possible values:
	//   "USE_FIXED_PORT"
	//   "USE_NAMED_PORT"
	//   "USE_SERVING_PORT"
	// +optional
	PortSpecification *string `json:"portSpecification,omitempty"`

	// ProxyHeader: Specifies the type of proxy header to append before
	// sending data to the backend, either NONE or PROXY_V1. The default is
	// NONE.
	//
	// Possible values:
	//   "NONE"
	//   "PROXY_V1"
	// +optional
	ProxyHeader *string `json:"proxyHeader,omitempty"`

	// RequestPath: The request path of the HTTP health check request. The
	// default value is /.
	// +optional
	RequestPath *string `json:"requestPath,omitempty"`

	// Response: The string to match anywhere in the first 1024 bytes of the
	// response body. If left empty (the default value), the status code
	// determines health. The response data can only be ASCII.
	// +optional
	Response *string `json:"response,omitempty"`
}

type HTTPSHealthCheck struct {
	// Host: The value of the host header in the HTTPS health check request.
	// If left empty (default value), the IP on behalf of which this health
	// check is performed will be used.
	// +optional
	Host *string `json:"host,omitempty"`

	// Port: The TCP port number for the health check request. The default
	// value is 443. Valid values are 1 through 65535.
	// +optional
	Port *int64 `json:"port,omitempty"`

	// PortName: Port name as defined in InstanceGroup#NamedPort#name. If
	// both port and port_name are defined, port takes precedence.
	// +optional
	PortName *string `json:"portName,omitempty"`

	// PortSpecification: Specifies how port is selected for health
	// checking, can be one of following values:
	// USE_FIXED_PORT: The port number in port is used for health
	// checking.
	// USE_NAMED_PORT: The portName is used for health
	// checking.
	// USE_SERVING_PORT: For NetworkEndpointGroup, the port specified for
	// each network endpoint is used for health checking. For other
	// backends, the port or named port specified in the Backend Service is
	// used for health checking.
	//
	//
	// If not specified, HTTPS health check follows behavior specified in
	// port and portName fields.
	//
	// Possible values:
	//   "USE_FIXED_PORT"
	//   "USE_NAMED_PORT"
	//   "USE_SERVING_PORT"
	// +optional
	PortSpecification *string `json:"portSpecification,omitempty"`

	// ProxyHeader: Specifies the type of proxy header to append before
	// sending data to the backend, either NONE or PROXY_V1. The default is
	// NONE.
	//
	// Possible values:
	//   "NONE"
	//   "PROXY_V1"
	// +optional
	ProxyHeader *string `json:"proxyHeader,omitempty"`

	// RequestPath: The request path of the HTTPS health check request. The
	// default value is /.
	// +optional
	RequestPath *string `json:"requestPath,omitempty"`

	// Response: The string to match anywhere in the first 1024 bytes of the
	// response body. If left empty (the default value), the status code
	// determines health. The response data can only be ASCII.
	// +optional
	Response *string `json:"response,omitempty"`
}

type SSLHealthCheck struct {
	// Port: The TCP port number for the health check request. The default
	// value is 443. Valid values are 1 through 65535.
	// +optional
	Port *int64 `json:"port,omitempty"`

	// PortName: Port name as defined in InstanceGroup#NamedPort#name. If
	// both port and port_name are defined, port takes precedence.
	// +optional
	PortName *string `json:"portName,omitempty"`

	// PortSpecification: Specifies how port is selected for health
	// checking, can be one of following values:
	// USE_FIXED_PORT: The port number in port is used for health
	// checking.
	// USE_NAMED_PORT: The portName is used for health
	// checking.
	// USE_SERVING_PORT: For NetworkEndpointGroup, the port specified for
	// each network endpoint is used for health checking. For other
	// backends, the port or named port specified in the Backend Service is
	// used for health checking.
	//
	//
	// If not specified, SSL health check follows behavior specified in port
	// and portName fields.
	//
	// Possible values:
	//   "USE_FIXED_PORT"
	//   "USE_NAMED_PORT"
	//   "USE_SERVING_PORT"
	// +optional
	PortSpecification *string `json:"portSpecification,omitempty"`

	// ProxyHeader: Specifies the type of proxy header to append before
	// sending data to the backend, either NONE or PROXY_V1. The default is
	// NONE.
	//
	// Possible values:
	//   "NONE"
	//   "PROXY_V1"
	// +optional
	ProxyHeader *string `json:"proxyHeader,omitempty"`

	// Request: The application data to send once the SSL connection has
	// been established (default value is empty). If both request and
	// response are empty, the connection establishment alone will indicate
	// health. The request data can only be ASCII.
	// +optional
	Request *string `json:"request,omitempty"`

	// Response: The bytes to match against the beginning of the response
	// data. If left empty (the default value), any response will indicate
	// health. The response data can only be ASCII.
	// +optional
	Response *string `json:"response,omitempty"`
}

type TCPHealthCheck struct {
	// Port: The TCP port number for the health check request. The default
	// value is 80. Valid values are 1 through 65535.
	// +optional
	Port *int64 `json:"port,omitempty"`

	// PortName: Port name as defined in InstanceGroup#NamedPort#name. If
	// both port and port_name are defined, port takes precedence.
	// +optional
	PortName *string `json:"portName,omitempty"`

	// PortSpecification: Specifies how port is selected for health
	// checking, can be one of following values:
	// USE_FIXED_PORT: The port number in port is used for health
	// checking.
	// USE_NAMED_PORT: The portName is used for health
	// checking.
	// USE_SERVING_PORT: For NetworkEndpointGroup, the port specified for
	// each network endpoint is used for health checking. For other
	// backends, the port or named port specified in the Backend Service is
	// used for health checking.
	//
	//
	// If not specified, TCP health check follows behavior specified in port
	// and portName fields.
	//
	// Possible values:
	//   "USE_FIXED_PORT"
	//   "USE_NAMED_PORT"
	//   "USE_SERVING_PORT"
	// +optional
	PortSpecification *string `json:"portSpecification,omitempty"`

	// ProxyHeader: Specifies the type of proxy header to append before
	// sending data to the backend, either NONE or PROXY_V1. The default is
	// NONE.
	//
	// Possible values:
	//   "NONE"
	//   "PROXY_V1"
	// +optional
	ProxyHeader *string `json:"proxyHeader,omitempty"`

	// Request: The application data to send once the TCP connection has
	// been established (default value is empty). If both request and
	// response are empty, the connection establishment alone will indicate
	// health. The request data can only be ASCII.
	// +optional
	Request *string `json:"request,omitempty"`

	// Response: The bytes to match against the beginning of the response
	// data. If left empty (the default value), any response will indicate
	// health. The response data can only be ASCII.
	// +optional
	Response *string `json:"response,omitempty"`
}

// A HealthCheckObservation represents the observed state of a Google Compute Engine
// VPC HealthCheck.
type HealthCheckObservation struct {
	// CreationTimestamp: [Output Only] Creation timestamp in 3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// ID: [Output Only] The unique identifier for the resource. This
	// identifier is defined by the server.
	ID int64 `json:"id,omitempty"`

	// Region: [Output Only] Region where the health check resides. Not
	// applicable to global health checks.
	Region string `json:"region,omitempty"`

	// SelfLink: [Output Only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`
}

// A HealthCheckSpec defines the desired state of a HealthCheck.
type HealthCheckSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       HealthCheckParameters `json:"forProvider"`
}

// A HealthCheckStatus represents the observed state of a HealthCheck.
type HealthCheckStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          HealthCheckObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A HealthCheck is a managed resource that represents a Google Compute Engine VPC
// HealthCheck.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type HealthCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HealthCheckSpec   `json:"spec"`
	Status HealthCheckStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HealthCheckList contains a list of HealthCheck.
type HealthCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HealthCheck `json:"items"`
}
