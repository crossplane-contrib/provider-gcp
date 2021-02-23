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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

var _ = compute.BackendService{}

// BackendServiceParameters define the desired state of a Google Compute Engine VPC
// BackendService. Most fields map directly to a BackendService:
// https://cloud.google.com/compute/docs/reference/rest/v1/networks
type BackendServiceParameters struct {
	// AffinityCookieTtlSec: If set to 0, the cookie is non-persistent and
	// lasts only until the end of the browser session (or equivalent). The
	// maximum allowed value is one day (86,400).
	AffinityCookieTTLSec int64 `json:"affinityCookieTtlSec,omitempty"`

	// Backends: The list of backends that serve this BackendService.
	Backends []*Backend `json:"backends,omitempty"`

	// CdnPolicy: Cloud CDN configuration for this BackendService.
	CdnPolicy *BackendServiceCdnPolicy `json:"cdnPolicy,omitempty"`

	// CircuitBreakers: Settings controlling the volume of connections to a
	// backend service. If not set, this feature is considered
	// disabled.
	//
	// This field is applicable to either:
	// - A regional backend service with the service_protocol set to HTTP,
	// HTTPS, or HTTP2, and load_balancing_scheme set to INTERNAL_MANAGED.
	//
	// - A global backend service with the load_balancing_scheme set to
	// INTERNAL_SELF_MANAGED.
	CircuitBreakers *CircuitBreakers `json:"circuitBreakers,omitempty"`

	ConnectionDraining *ConnectionDraining `json:"connectionDraining,omitempty"`

	// ConsistentHash: Consistent Hash-based load balancing can be used to
	// provide soft session affinity based on HTTP headers, cookies or other
	// properties. This load balancing policy is applicable only for HTTP
	// connections. The affinity to a particular destination host will be
	// lost when one or more hosts are added/removed from the destination
	// service. This field specifies parameters that control consistent
	// hashing. This field is only applicable when localityLbPolicy is set
	// to MAGLEV or RING_HASH.
	//
	// This field is applicable to either:
	// - A regional backend service with the service_protocol set to HTTP,
	// HTTPS, or HTTP2, and load_balancing_scheme set to INTERNAL_MANAGED.
	//
	// - A global backend service with the load_balancing_scheme set to
	// INTERNAL_SELF_MANAGED.
	ConsistentHash *ConsistentHashLoadBalancerSettings `json:"consistentHash,omitempty"`

	// CustomRequestHeaders: Headers that the HTTP/S load balancer should
	// add to proxied requests.
	CustomRequestHeaders []string `json:"customRequestHeaders,omitempty"`

	// Description: An optional description of this resource. Provide this
	// property when you create the resource.
	Description string `json:"description,omitempty"`

	// EnableCDN: If true, enables Cloud CDN for the backend service. Only
	// applicable if the loadBalancingScheme is EXTERNAL and the protocol is
	// HTTP or HTTPS.
	EnableCDN bool `json:"enableCDN,omitempty"`

	// FailoverPolicy: Applicable only to Failover for Internal TCP/UDP Load
	// Balancing. Requires at least one backend instance group to be defined
	// as a backup (failover) backend.
	FailoverPolicy *BackendServiceFailoverPolicy `json:"failoverPolicy,omitempty"`

	// Fingerprint: Fingerprint of this resource. A hash of the contents
	// stored in this object. This field is used in optimistic locking. This
	// field will be ignored when inserting a BackendService. An up-to-date
	// fingerprint must be provided in order to update the BackendService,
	// otherwise the request will fail with error 412 conditionNotMet.
	//
	// To see the latest fingerprint, make a get() request to retrieve a
	// BackendService.
	Fingerprint string `json:"fingerprint,omitempty"`

	// HealthChecks: The list of URLs to the healthChecks, httpHealthChecks
	// (legacy), or httpsHealthChecks (legacy) resource for health checking
	// this backend service. Not all backend services support legacy health
	// checks. See  Load balancer guide. Currently at most one health check
	// can be specified. Backend services with instance group or zonal NEG
	// backends must have a health check. Backend services with internet NEG
	// backends must not have a health check. A health check must
	HealthChecks []string `json:"healthChecks,omitempty"`

	Iap *BackendServiceIAP `json:"iap,omitempty"`

	// LoadBalancingScheme: Specifies the load balancer type. Choose
	// EXTERNAL for load balancers that receive traffic from external
	// clients. Choose INTERNAL for Internal TCP/UDP Load Balancing. Choose
	// INTERNAL_MANAGED for Internal HTTP(S) Load Balancing. Choose
	// INTERNAL_SELF_MANAGED for Traffic Director. A backend service created
	// for one type of load balancing cannot be used with another. For more
	// information, refer to Choosing a load balancer.
	//
	// Possible values:
	//   "EXTERNAL"
	//   "INTERNAL"
	//   "INTERNAL_MANAGED"
	//   "INTERNAL_SELF_MANAGED"
	//   "INVALID_LOAD_BALANCING_SCHEME"
	LoadBalancingScheme string `json:"loadBalancingScheme,omitempty"`

	// LocalityLbPolicy: The load balancing algorithm used within the scope
	// of the locality. The possible values are:
	// - ROUND_ROBIN: This is a simple policy in which each healthy backend
	// is selected in round robin order. This is the default.
	// - LEAST_REQUEST: An O(1) algorithm which selects two random healthy
	// hosts and picks the host which has fewer active requests.
	// - RING_HASH: The ring/modulo hash load balancer implements consistent
	// hashing to backends. The algorithm has the property that the
	// addition/removal of a host from a set of N hosts only affects 1/N of
	// the requests.
	// - RANDOM: The load balancer selects a random healthy host.
	// - ORIGINAL_DESTINATION: Backend host is selected based on the client
	// connection metadata, i.e., connections are opened to the same address
	// as the destination address of the incoming connection before the
	// connection was redirected to the load balancer.
	// - MAGLEV: used as a drop in replacement for the ring hash load
	// balancer. Maglev is not as stable as ring hash but has faster table
	// lookup build times and host selection times. For more information
	// about Maglev, refer to https://ai.google/research/pubs/pub44824
	//
	//
	// This field is applicable to either:
	// - A regional backend service with the service_protocol set to HTTP,
	// HTTPS, or HTTP2, and load_balancing_scheme set to INTERNAL_MANAGED.
	//
	// - A global backend service with the load_balancing_scheme set to
	// INTERNAL_SELF_MANAGED.
	//
	// If sessionAffinity is not NONE, and this field is not set to >MAGLEV
	// or RING_HASH, session affinity settings will not take effect.
	//
	// Possible values:
	//   "INVALID_LB_POLICY"
	//   "LEAST_REQUEST"
	//   "MAGLEV"
	//   "ORIGINAL_DESTINATION"
	//   "RANDOM"
	//   "RING_HASH"
	//   "ROUND_ROBIN"
	LocalityLbPolicy string `json:"localityLbPolicy,omitempty"`

	// LogConfig: This field denotes the logging options for the load
	// balancer traffic served by this backend service. If logging is
	// enabled, logs will be exported to Stackdriver.
	LogConfig *BackendServiceLogConfig `json:"logConfig,omitempty"`

	// Network: URI of the network to which this router belongs.
	// +optional
	// +immutable
	Network *string `json:"network,omitempty"`

	// NetworkRef references a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkRef *xpv1.Reference `json:"networkRef,omitempty"`

	// NetworkSelector selects a reference to a Network
	// +optional
	// +immutable
	NetworkSelector *xpv1.Selector `json:"networkSelector,omitempty"`

	// OutlierDetection: Settings controlling the eviction of unhealthy
	// hosts from the load balancing pool for the backend service. If not
	// set, this feature is considered disabled.
	//
	// This field is applicable to either:
	// - A regional backend service with the service_protocol set to HTTP,
	// HTTPS, or HTTP2, and load_balancing_scheme set to INTERNAL_MANAGED.
	//
	// - A global backend service with the load_balancing_scheme set to
	// INTERNAL_SELF_MANAGED.
	OutlierDetection *OutlierDetection `json:"outlierDetection,omitempty"`

	// Port: Deprecated in favor of portName. The TCP port to connect on the
	// backend. The default value is 80.
	//
	// This cannot be used if the loadBalancingScheme is INTERNAL (Internal
	// TCP/UDP Load Balancing).
	Port int64 `json:"port,omitempty"`

	// PortName: A named port on a backend instance group representing the
	// port for communication to the backend VMs in that group. Required
	// when the loadBalancingScheme is EXTERNAL, INTERNAL_MANAGED, or
	// INTERNAL_SELF_MANAGED and the backends are instance groups. The named
	// port must be defined on each backend instance group. This parameter
	// has no meaning if the backends are NEGs.
	//
	//
	//
	// Must be omitted when the loadBalancingScheme is INTERNAL (Internal
	// TCP/UDP Load Blaancing).
	PortName string `json:"portName,omitempty"`

	// Protocol: The protocol this BackendService uses to communicate with
	// backends.
	//
	// Possible values are HTTP, HTTPS, HTTP2, TCP, SSL, or UDP. depending
	// on the chosen load balancer or Traffic Director configuration. Refer
	// to the documentation for the load balancer or for Traffic Director
	// for more information.
	//
	// Possible values:
	//   "HTTP"
	//   "HTTP2"
	//   "HTTPS"
	//   "SSL"
	//   "TCP"
	//   "UDP"
	Protocol string `json:"protocol,omitempty"`

	// SessionAffinity: Type of session affinity to use. The default is
	// NONE. Session affinity is not applicable if the --protocol is
	// UDP.
	//
	// When the loadBalancingScheme is EXTERNAL, possible values are NONE,
	// CLIENT_IP, or GENERATED_COOKIE. You can use GENERATED_COOKIE if the
	// protocol is HTTP or HTTPS.
	//
	// When the loadBalancingScheme is INTERNAL, possible values are NONE,
	// CLIENT_IP, CLIENT_IP_PROTO, or CLIENT_IP_PORT_PROTO.
	//
	// When the loadBalancingScheme is INTERNAL_SELF_MANAGED, or
	// INTERNAL_MANAGED, possible values are NONE, CLIENT_IP,
	// GENERATED_COOKIE, HEADER_FIELD, or HTTP_COOKIE.
	//
	// Possible values:
	//   "CLIENT_IP"
	//   "CLIENT_IP_PORT_PROTO"
	//   "CLIENT_IP_PROTO"
	//   "GENERATED_COOKIE"
	//   "HEADER_FIELD"
	//   "HTTP_COOKIE"
	//   "NONE"
	SessionAffinity string `json:"sessionAffinity,omitempty"`

	// TimeoutSec: The backend service timeout has a different meaning
	// depending on the type of load balancer. For more information read,
	// Backend service settings The default is 30 seconds.
	TimeoutSec int64 `json:"timeoutSec,omitempty"`
}

// Backend is a message containing information of one individual backend.
type Backend struct {
	// BalancingMode: Specifies the balancing mode for the backend.
	//
	// When choosing a balancing mode, you need to consider the
	// loadBalancingScheme, and protocol for the backend service, as well as
	// the type of backend (instance group or NEG).
	//
	//
	// - If the load balancing mode is CONNECTION, then the load is spread
	// based on how many concurrent connections the backend can handle.
	// You can use the CONNECTION balancing mode if the protocol for the
	// backend service is SSL, TCP, or UDP.
	//
	// If the loadBalancingScheme for the backend service is EXTERNAL (SSL
	// Proxy and TCP Proxy load balancers), you must also specify exactly
	// one of the following parameters: maxConnections (except for regional
	// managed instance groups), maxConnectionsPerInstance, or
	// maxConnectionsPerEndpoint.
	//
	// If the loadBalancingScheme for the backend service is INTERNAL
	// (internal TCP/UDP load balancers), you cannot specify any additional
	// parameters.
	//
	// - If the load balancing mode is RATE, the load is spread based on the
	// rate of HTTP requests per second (RPS).
	// You can use the RATE balancing mode if the protocol for the backend
	// service is HTTP or HTTPS. You must specify exactly one of the
	// following parameters: maxRate (except for regional managed instance
	// groups), maxRatePerInstance, or maxRatePerEndpoint.
	//
	// - If the load balancing mode is UTILIZATION, the load is spread based
	// on the backend utilization of instances in an instance group.
	// You can use the UTILIZATION balancing mode if the loadBalancingScheme
	// of the backend service is EXTERNAL, INTERNAL_SELF_MANAGED, or
	// INTERNAL_MANAGED and the backends are instance groups. There are no
	// restrictions on the backend service protocol.
	//
	// Possible values:
	//   "CONNECTION"
	//   "RATE"
	//   "UTILIZATION"
	BalancingMode string `json:"balancingMode,omitempty"`

	// CapacityScaler: A multiplier applied to the group's maximum servicing
	// capacity (based on UTILIZATION, RATE or CONNECTION). Default value is
	// 1, which means the group will serve up to 100% of its configured
	// capacity (depending on balancingMode). A setting of 0 means the group
	// is completely drained, offering 0% of its available Capacity. Valid
	// range is [0.0,1.0].
	//
	// This cannot be used for internal load balancing.
	CapacityScaler resource.Quantity `json:"capacityScaler,omitempty"`

	// Description: An optional description of this resource. Provide this
	// property when you create the resource.
	Description string `json:"description,omitempty"`

	// Failover: This field designates whether this is a failover backend.
	// More than one failover backend can be configured for a given
	// BackendService.
	Failover bool `json:"failover,omitempty"`

	// Group: The fully-qualified URL of an instance group or network
	// endpoint group (NEG) resource. The type of backend that a backend
	// service supports depends on the backend service's
	// loadBalancingScheme.
	//
	//
	// - When the loadBalancingScheme for the backend service is EXTERNAL,
	// INTERNAL_SELF_MANAGED, or INTERNAL_MANAGED, the backend can be either
	// an instance group or a NEG. The backends on the backend service must
	// be either all instance groups or all NEGs. You cannot mix instance
	// group and NEG backends on the same backend service.
	//
	//
	// - When the loadBalancingScheme for the backend service is INTERNAL,
	// the backend must be an instance group in the same region as the
	// backend service. NEGs are not supported.
	//
	// You must use the fully-qualified URL (starting with
	// https://www.googleapis.com/) to specify the instance group or NEG.
	// Partial URLs are not supported.
	Group string `json:"group,omitempty"`

	// MaxConnections: Defines a target maximum number of simultaneous
	// connections that the backend can handle. Valid for network endpoint
	// group and instance group backends (except for regional managed
	// instance groups). If the backend's balancingMode is UTILIZATION, this
	// is an optional parameter. If the backend's balancingMode is
	// CONNECTION, and backend is attached to a backend service whose
	// loadBalancingScheme is EXTERNAL, you must specify either this
	// parameter, maxConnectionsPerInstance, or
	// maxConnectionsPerEndpoint.
	//
	// Not available if the backend's balancingMode is RATE. If the
	// loadBalancingScheme is INTERNAL, then maxConnections is not
	// supported, even though the backend requires a balancing mode of
	// CONNECTION.
	MaxConnections int64 `json:"maxConnections,omitempty"`

	// MaxConnectionsPerEndpoint: Defines a target maximum number of
	// simultaneous connections for an endpoint of a NEG. This is multiplied
	// by the number of endpoints in the NEG to implicitly calculate a
	// maximum number of target maximum simultaneous connections for the
	// NEG. If the backend's balancingMode is CONNECTION, and the backend is
	// attached to a backend service whose loadBalancingScheme is EXTERNAL,
	// you must specify either this parameter, maxConnections, or
	// maxConnectionsPerInstance.
	//
	// Not available if the backend's balancingMode is RATE. Internal
	// TCP/UDP load balancing does not support setting
	// maxConnectionsPerEndpoint even though its backends require a
	// balancing mode of CONNECTION.
	MaxConnectionsPerEndpoint int64 `json:"maxConnectionsPerEndpoint,omitempty"`

	// MaxConnectionsPerInstance: Defines a target maximum number of
	// simultaneous connections for a single VM in a backend instance group.
	// This is multiplied by the number of instances in the instance group
	// to implicitly calculate a target maximum number of simultaneous
	// connections for the whole instance group. If the backend's
	// balancingMode is UTILIZATION, this is an optional parameter. If the
	// backend's balancingMode is CONNECTION, and backend is attached to a
	// backend service whose loadBalancingScheme is EXTERNAL, you must
	// specify either this parameter, maxConnections, or
	// maxConnectionsPerEndpoint.
	//
	// Not available if the backend's balancingMode is RATE. Internal
	// TCP/UDP load balancing does not support setting
	// maxConnectionsPerInstance even though its backends require a
	// balancing mode of CONNECTION.
	MaxConnectionsPerInstance int64 `json:"maxConnectionsPerInstance,omitempty"`

	// MaxRate: Defines a maximum number of HTTP requests per second (RPS)
	// that the backend can handle. Valid for network endpoint group and
	// instance group backends (except for regional managed instance
	// groups). Must not be defined if the backend is a managed instance
	// group that uses autoscaling based on load balancing.
	//
	// If the backend's balancingMode is UTILIZATION, this is an optional
	// parameter. If the backend's balancingMode is RATE, you must specify
	// maxRate, maxRatePerInstance, or maxRatePerEndpoint.
	//
	// Not available if the backend's balancingMode is CONNECTION.
	MaxRate int64 `json:"maxRate,omitempty"`

	// MaxRatePerEndpoint: Defines a maximum target for requests per second
	// (RPS) for an endpoint of a NEG. This is multiplied by the number of
	// endpoints in the NEG to implicitly calculate a target maximum rate
	// for the NEG.
	//
	// If the backend's balancingMode is RATE, you must specify either this
	// parameter, maxRate (except for regional managed instance groups), or
	// maxRatePerInstance.
	//
	// Not available if the backend's balancingMode is CONNECTION.
	MaxRatePerEndpoint resource.Quantity `json:"maxRatePerEndpoint,omitempty"`

	// MaxRatePerInstance: Defines a maximum target for requests per second
	// (RPS) for a single VM in a backend instance group. This is multiplied
	// by the number of instances in the instance group to implicitly
	// calculate a target maximum rate for the whole instance group.
	//
	// If the backend's balancingMode is UTILIZATION, this is an optional
	// parameter. If the backend's balancingMode is RATE, you must specify
	// either this parameter, maxRate (except for regional managed instance
	// groups), or maxRatePerEndpoint.
	//
	// Not available if the backend's balancingMode is CONNECTION.
	MaxRatePerInstance resource.Quantity `json:"maxRatePerInstance,omitempty"`

	// MaxUtilization: Defines the maximum average backend utilization of a
	// backend VM in an instance group. The valid range is [0.0, 1.0]. This
	// is an optional parameter if the backend's balancingMode is
	// UTILIZATION.
	//
	// This parameter can be used in conjunction with maxRate,
	// maxRatePerInstance, maxConnections (except for regional managed
	// instance groups), or maxConnectionsPerInstance.
	MaxUtilization resource.Quantity `json:"maxUtilization,omitempty"`
}

// BackendServiceCdnPolicy: Message containing Cloud CDN configuration
// for a backend service.
type BackendServiceCdnPolicy struct {
	// CacheKeyPolicy: The CacheKeyPolicy for this CdnPolicy.
	CacheKeyPolicy *CacheKeyPolicy `json:"cacheKeyPolicy,omitempty"`

	// SignedUrlCacheMaxAgeSec: Maximum number of seconds the response to a
	// signed URL request will be considered fresh. After this time period,
	// the response will be revalidated before being served. Defaults to 1hr
	// (3600s). When serving responses to signed URL requests, Cloud CDN
	// will internally behave as though all responses from this backend had
	// a "Cache-Control: public, max-age=[TTL]" header, regardless of any
	// existing Cache-Control header. The actual headers served in responses
	// will not be altered.
	SignedUrlCacheMaxAgeSec int64 `json:"signedUrlCacheMaxAgeSec,omitempty,string"`

	// SignedUrlKeyNames: [Output Only] Names of the keys for signing
	// request URLs.
	SignedUrlKeyNames []string `json:"signedUrlKeyNames,omitempty"`
}

// CacheKeyPolicy is a message containing what to include in the cache key
// for a request for Cloud CDN.
type CacheKeyPolicy struct {
	// IncludeHost: If true, requests to different hosts will be cached
	// separately.
	IncludeHost bool `json:"includeHost,omitempty"`

	// IncludeProtocol: If true, http and https requests will be cached
	// separately.
	IncludeProtocol bool `json:"includeProtocol,omitempty"`

	// IncludeQueryString: If true, include query string parameters in the
	// cache key according to query_string_whitelist and
	// query_string_blacklist. If neither is set, the entire query string
	// will be included. If false, the query string will be excluded from
	// the cache key entirely.
	IncludeQueryString bool `json:"includeQueryString,omitempty"`

	// QueryStringBlacklist: Names of query string parameters to exclude in
	// cache keys. All other parameters will be included. Either specify
	// query_string_whitelist or query_string_blacklist, not both. '&' and
	// '=' will be percent encoded and not treated as delimiters.
	QueryStringBlacklist []string `json:"queryStringBlacklist,omitempty"`

	// QueryStringWhitelist: Names of query string parameters to include in
	// cache keys. All other parameters will be excluded. Either specify
	// query_string_whitelist or query_string_blacklist, not both. '&' and
	// '=' will be percent encoded and not treated as delimiters.
	QueryStringWhitelist []string `json:"queryStringWhitelist,omitempty"`
}

// CircuitBreakers is a settings controlling the volume of connections to a
// backend service.
type CircuitBreakers struct {
	// MaxConnections: The maximum number of connections to the backend
	// service. If not specified, there is no limit.
	MaxConnections int64 `json:"maxConnections,omitempty"`

	// MaxPendingRequests: The maximum number of pending requests allowed to
	// the backend service. If not specified, there is no limit.
	MaxPendingRequests int64 `json:"maxPendingRequests,omitempty"`

	// MaxRequests: The maximum number of parallel requests that allowed to
	// the backend service. If not specified, there is no limit.
	MaxRequests int64 `json:"maxRequests,omitempty"`

	// MaxRequestsPerConnection: Maximum requests for a single connection to
	// the backend service. This parameter is respected by both the HTTP/1.1
	// and HTTP/2 implementations. If not specified, there is no limit.
	// Setting this parameter to 1 will effectively disable keep alive.
	MaxRequestsPerConnection int64 `json:"maxRequestsPerConnection,omitempty"`

	// MaxRetries: The maximum number of parallel retries allowed to the
	// backend cluster. If not specified, the default is 1.
	MaxRetries int64 `json:"maxRetries,omitempty"`
}

// ConnectionDraining is a message containing connection draining
// configuration.
type ConnectionDraining struct {
	// DrainingTimeoutSec: The amount of time in seconds to allow existing
	// connections to persist while on unhealthy backend VMs. Only
	// applicable if the protocol is not UDP. The valid range is [0, 3600].
	DrainingTimeoutSec int64 `json:"drainingTimeoutSec,omitempty"`
}

// ConsistentHashLoadBalancerSettings is a message defines settings for
// a consistent hash style load balancer.
type ConsistentHashLoadBalancerSettings struct {
	// HttpCookie: Hash is based on HTTP Cookie. This field describes a HTTP
	// cookie that will be used as the hash key for the consistent hash load
	// balancer. If the cookie is not present, it will be generated. This
	// field is applicable if the sessionAffinity is set to HTTP_COOKIE.
	HttpCookie *ConsistentHashLoadBalancerSettingsHttpCookie `json:"httpCookie,omitempty"`

	// HttpHeaderName: The hash based on the value of the specified header
	// field. This field is applicable if the sessionAffinity is set to
	// HEADER_FIELD.
	HttpHeaderName string `json:"httpHeaderName,omitempty"`

	// MinimumRingSize: The minimum number of virtual nodes to use for the
	// hash ring. Defaults to 1024. Larger ring sizes result in more
	// granular load distributions. If the number of hosts in the load
	// balancing pool is larger than the ring size, each host will be
	// assigned a single virtual node.
	MinimumRingSize int64 `json:"minimumRingSize,omitempty,string"`
}

// ConsistentHashLoadBalancerSettingsHttpCookie: The information about
// the HTTP Cookie on which the hash function is based for load
// balancing policies that use a consistent hash.
type ConsistentHashLoadBalancerSettingsHttpCookie struct {
	// Name: Name of the cookie.
	Name string `json:"name,omitempty"`

	// Path: Path to set for the cookie.
	Path string `json:"path,omitempty"`

	// Ttl: Lifetime of the cookie.
	Ttl *Duration `json:"ttl,omitempty"`
}

// Duration: A Duration represents a fixed-length span of time
// represented as a count of seconds and fractions of seconds at
// nanosecond resolution. It is independent of any calendar and concepts
// like "day" or "month". Range is approximately 10,000 years.
type Duration struct {
	// Nanos: Span of time that's a fraction of a second at nanosecond
	// resolution. Durations less than one second are represented with a 0
	// `seconds` field and a positive `nanos` field. Must be from 0 to
	// 999,999,999 inclusive.
	Nanos int64 `json:"nanos,omitempty"`

	// Seconds: Span of time at a resolution of a second. Must be from 0 to
	// 315,576,000,000 inclusive. Note: these bounds are computed from: 60
	// sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years
	Seconds int64 `json:"seconds,omitempty,string"`
}

// BackendServiceFailoverPolicy: Applicable only to Failover for
// Internal TCP/UDP Load Balancing. On failover or failback, this field
// indicates whether connection draining will be honored. GCP has a
// fixed connection draining timeout of 10 minutes. A setting of true
// terminates existing TCP connections to the active pool during
// failover and failback, immediately draining traffic. A setting of
// false allows existing TCP connections to persist, even on VMs no
// longer in the active pool, for up to the duration of the connection
// draining timeout (10 minutes).
type BackendServiceFailoverPolicy struct {
	// DisableConnectionDrainOnFailover: This can be set to true only if the
	// protocol is TCP.
	//
	// The default is false.
	DisableConnectionDrainOnFailover bool `json:"disableConnectionDrainOnFailover,omitempty"`

	// DropTrafficIfUnhealthy: Applicable only to Failover for Internal
	// TCP/UDP Load Balancing. If set to true, connections to the load
	// balancer are dropped when all primary and all backup backend VMs are
	// unhealthy. If set to false, connections are distributed among all
	// primary VMs when all primary and all backup backend VMs are
	// unhealthy.
	//
	// The default is false.
	DropTrafficIfUnhealthy bool `json:"dropTrafficIfUnhealthy,omitempty"`

	// FailoverRatio: Applicable only to Failover for Internal TCP/UDP Load
	// Balancing. The value of the field must be in the range [0, 1]. If the
	// value is 0, the load balancer performs a failover when the number of
	// healthy primary VMs equals zero. For all other values, the load
	// balancer performs a failover when the total number of healthy primary
	// VMs is less than this ratio.
	FailoverRatio resource.Quantity `json:"failoverRatio,omitempty"`
}

// BackendServiceIAP: Identity-Aware Proxy
type BackendServiceIAP struct {
	Enabled bool `json:"enabled,omitempty"`

	Oauth2ClientId string `json:"oauth2ClientId,omitempty"`

	Oauth2ClientSecret string `json:"oauth2ClientSecret,omitempty"`

	// Oauth2ClientSecretSha256: [Output Only] SHA256 hash value for the
	// field oauth2_client_secret above.
	Oauth2ClientSecretSha256 string `json:"oauth2ClientSecretSha256,omitempty"`
}

// BackendServiceLogConfig is the available logging options for the load
// balancer traffic served by this backend service.
type BackendServiceLogConfig struct {
	// Enable: This field denotes whether to enable logging for the load
	// balancer traffic served by this backend service.
	Enable bool `json:"enable,omitempty"`

	// SampleRate: This field can only be specified if logging is enabled
	// for this backend service. The value of the field must be in [0, 1].
	// This configures the sampling rate of requests to the load balancer
	// where 1.0 means all logged requests are reported and 0.0 means no
	// logged requests are reported. The default value is 1.0.
	SampleRate resource.Quantity `json:"sampleRate,omitempty"`
}

// OutlierDetection: Settings controlling the eviction of unhealthy
// hosts from the load balancing pool for the backend service.
type OutlierDetection struct {
	// BaseEjectionTime: The base time that a host is ejected for. The real
	// ejection time is equal to the base ejection time multiplied by the
	// number of times the host has been ejected. Defaults to 30000ms or
	// 30s.
	BaseEjectionTime *Duration `json:"baseEjectionTime,omitempty"`

	// ConsecutiveErrors: Number of errors before a host is ejected from the
	// connection pool. When the backend host is accessed over HTTP, a 5xx
	// return code qualifies as an error. Defaults to 5.
	ConsecutiveErrors int64 `json:"consecutiveErrors,omitempty"`

	// ConsecutiveGatewayFailure: The number of consecutive gateway failures
	// (502, 503, 504 status or connection errors that are mapped to one of
	// those status codes) before a consecutive gateway failure ejection
	// occurs. Defaults to 3.
	ConsecutiveGatewayFailure int64 `json:"consecutiveGatewayFailure,omitempty"`

	// EnforcingConsecutiveErrors: The percentage chance that a host will be
	// actually ejected when an outlier status is detected through
	// consecutive 5xx. This setting can be used to disable ejection or to
	// ramp it up slowly. Defaults to 0.
	EnforcingConsecutiveErrors int64 `json:"enforcingConsecutiveErrors,omitempty"`

	// EnforcingConsecutiveGatewayFailure: The percentage chance that a host
	// will be actually ejected when an outlier status is detected through
	// consecutive gateway failures. This setting can be used to disable
	// ejection or to ramp it up slowly. Defaults to 100.
	EnforcingConsecutiveGatewayFailure int64 `json:"enforcingConsecutiveGatewayFailure,omitempty"`

	// EnforcingSuccessRate: The percentage chance that a host will be
	// actually ejected when an outlier status is detected through success
	// rate statistics. This setting can be used to disable ejection or to
	// ramp it up slowly. Defaults to 100.
	EnforcingSuccessRate int64 `json:"enforcingSuccessRate,omitempty"`

	// Interval: Time interval between ejection analysis sweeps. This can
	// result in both new ejections as well as hosts being returned to
	// service. Defaults to 1 second.
	Interval *Duration `json:"interval,omitempty"`

	// MaxEjectionPercent: Maximum percentage of hosts in the load balancing
	// pool for the backend service that can be ejected. Defaults to 50%.
	MaxEjectionPercent int64 `json:"maxEjectionPercent,omitempty"`

	// SuccessRateMinimumHosts: The number of hosts in a cluster that must
	// have enough request volume to detect success rate outliers. If the
	// number of hosts is less than this setting, outlier detection via
	// success rate statistics is not performed for any host in the cluster.
	// Defaults to 5.
	SuccessRateMinimumHosts int64 `json:"successRateMinimumHosts,omitempty"`

	// SuccessRateRequestVolume: The minimum number of total requests that
	// must be collected in one interval (as defined by the interval
	// duration above) to include this host in success rate based outlier
	// detection. If the volume is lower than this setting, outlier
	// detection via success rate statistics is not performed for that host.
	// Defaults to 100.
	SuccessRateRequestVolume int64 `json:"successRateRequestVolume,omitempty"`

	// SuccessRateStdevFactor: This factor is used to determine the ejection
	// threshold for success rate outlier ejection. The ejection threshold
	// is the difference between the mean success rate, and the product of
	// this factor and the standard deviation of the mean success rate: mean
	// - (stdev * success_rate_stdev_factor). This factor is divided by a
	// thousand to get a double. That is, if the desired factor is 1.9, the
	// runtime value should be 1900. Defaults to 1900.
	SuccessRateStdevFactor int64 `json:"successRateStdevFactor,omitempty"`
}

// A BackendServiceObservation represents the observed state of a Google Compute Engine
// VPC BackendService.
type BackendServiceObservation struct {
	// CreationTimestamp: [Output Only] Creation timestamp in RFC3339 text
	// format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// Id: [Output Only] The unique identifier for the resource. This
	// identifier is defined by the server.
	ID int64 `json:"id,omitempty"`

	// Kind: [Output Only] Type of resource. Always compute#backendService
	// for backend services.
	Kind string `json:"kind,omitempty"`

	// Region: [Output Only] URL of the region where the regional backend
	// service resides. This field is not applicable to global backend
	// services. You must specify this field as part of the HTTP request
	// URL. It is not settable as a field in the request body.
	Region string `json:"region,omitempty"`

	// SecurityPolicy: [Output Only] The resource URL for the security
	// policy associated with this backend service.
	SecurityPolicy string `json:"securityPolicy,omitempty"`

	// SelfLink: [Output Only] Server-defined URL for the resource.
	SelfLink string `json:"selfLink,omitempty"`
}

// A BackendServiceSpec defines the desired state of a BackendService.
type BackendServiceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       BackendServiceParameters `json:"forProvider"`
}

// A BackendServiceStatus represents the observed state of a BackendService.
type BackendServiceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          BackendServiceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A BackendService is a managed resource that represents a Google Compute Engine VPC
// BackendService.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type BackendService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackendServiceSpec   `json:"spec"`
	Status BackendServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackendServiceList contains a list of BackendService.
type BackendServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendService `json:"items"`
}
