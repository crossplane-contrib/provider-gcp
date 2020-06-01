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
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Application serving states
const (
	StateUnspecified    = "UNSPECIFIED"
	StateServing        = "SERVING"
	StateUserDisabled   = "USER_DISABLED"
	StateSystemDisabled = "SYSTEM_DISABLED"
)

// ApplicationParameters is the desired state of an AppEngine
// Application.
type ApplicationParameters struct {
	// AuthDomain: Google Apps authentication domain that controls which
	// users can access this application.Defaults to open access for any
	// Google Account.
	// +optional
	AuthDomain *string `json:"authDomain,omitempty"`

	// DatabaseType: The type of the Cloud Firestore or Cloud Datastore
	// database associated with this application.
	//
	// Possible values:
	//   "DATABASE_TYPE_UNSPECIFIED" - Database type is unspecified.
	//   "CLOUD_DATASTORE" - Cloud Datastore
	//   "CLOUD_FIRESTORE" - Cloud Firestore Native
	//   "CLOUD_DATASTORE_COMPATIBILITY" - Cloud Firestore in Datastore Mode
	// +immutable
	// +optional
	DatabaseType *string `json:"databaseType,omitempty"`

	// DefaultCookieExpiration: Cookie expiration policy for this
	// application.
	// +optional
	DefaultCookieExpiration *string `json:"defaultCookieExpiration,omitempty"`

	// DispatchRules: HTTP path dispatch rules for requests to the
	// application that do not explicitly target a service or version. Rules
	// are order-dependent. Up to 20 dispatch rules can be supported.
	// +immutable
	// +optional
	DispatchRules []*URLDispatchRule `json:"dispatchRules,omitempty"`

	// FeatureSettings: The feature specific settings to be used in the
	// application.
	// +immutable
	// +optional
	FeatureSettings *FeatureSettings `json:"featureSettings,omitempty"`

	// GcrDomain: The Google Container Registry domain used for storing
	// managed build docker images for this application.
	// +immutable
	// +optional
	GcrDomain *string `json:"gcrDomain,omitempty"`

	// Iap is TODO
	// +optional
	// Iap *IdentityAwareProxy `json:"iap,omitempty"`

	// LocationID: Location from which this application runs. Application
	// instances run out of the data centers in the specified location,
	// which is also where all of the application's end user content is
	// stored.Defaults to us-central.View the list of supported locations
	// (https://cloud.google.com/appengine/docs/locations).
	// +immutable
	// +optional
	LocationID *string `json:"locationId,omitempty"`
}

// URLDispatchRule is a rules to match an HTTP request and dispatch that
// request to a service.
type URLDispatchRule struct {
	// Domain: Domain name to match against. The wildcard "*" is supported
	// if specified before a period: "*.".Defaults to matching all domains:
	// "*".
	// +optional
	Domain string `json:"domain,omitempty"`

	// Path: Pathname within the host. Must start with a "/". A single "*"
	// can be included at the end of the path.The sum of the lengths of the
	// domain and path may not exceed 100 characters.
	Path string `json:"path,omitempty"`

	// Service: Resource ID of a service in this application that should
	// serve the matched request. The service must already exist. Example:
	// default.
	Service string `json:"service,omitempty"`
}

// FeatureSettings are feature specific settings to be used in the
// application. These define behaviors that are user configurable.
type FeatureSettings struct {
	// SplitHealthChecks: Boolean value indicating if split health checks
	// should be used instead of the legacy health checks. At an app.yaml
	// level, this means defaulting to 'readiness_check' and
	// 'liveness_check' values instead of 'health_check' ones. Once the
	// legacy 'health_check' behavior is deprecated, and this value is
	// always true, this setting can be removed.
	SplitHealthChecks bool `json:"splitHealthChecks,omitempty"`

	// UseContainerOptimizedOs: If true, use Container-Optimized OS
	// (https://cloud.google.com/container-optimized-os/) base image for
	// VMs, rather than a base Debian image.
	UseContainerOptimizedOs bool `json:"useContainerOptimizedOs,omitempty"`
}

// IdentityAwareProxy is an Identity-Aware Proxy
type IdentityAwareProxy struct {
	// Enabled: Whether the serving infrastructure will authenticate and
	// authorize all incoming requests.If true, the oauth2_client_id and
	// oauth2_client_secret fields must be non-empty.
	Enabled bool `json:"enabled,omitempty"`

	// Oauth2ClientId: OAuth2 client ID to use for the authentication flow.
	Oauth2ClientID string `json:"oauth2ClientId,omitempty"`

	// Oauth2ClientSecret: OAuth2 client secret to use for the
	// authentication flow.For security reasons, this value cannot be
	// retrieved via the API. Instead, the SHA-256 hash of the value is
	// returned in the oauth2_client_secret_sha256 field.@InputOnly
	Oauth2ClientSecret string `json:"oauth2ClientSecret,omitempty"`

	// Oauth2ClientSecretSha256: Hex-encoded SHA-256 hash of the client
	// secret.@OutputOnly
	Oauth2ClientSecretSha256 string `json:"oauth2ClientSecretSha256,omitempty"`
}

// ApplicationObservation is the current state of an AppEngine
// Application.
type ApplicationObservation struct {
	// CodeBucket: Google Cloud Storage bucket that can be used for storing
	// files associated with this application. This bucket is associated
	// with the application and can be used by the gcloud deployment
	// commands.
	CodeBucket string `json:"codeBucket,omitempty"`

	// DefaultBucket: Google Cloud Storage bucket that can be used by this
	// application to store content.
	DefaultBucket string `json:"defaultBucket,omitempty"`

	// DefaultHostname: Hostname used to reach this application, as resolved
	// by App Engine.
	DefaultHostname string `json:"defaultHostname,omitempty"`

	// Name: Full path to the Application resource in the API. Example:
	// apps/myapp.
	Name string `json:"name,omitempty"`

	// ServingStatus: Serving status of this application.
	//
	// Possible values:
	//   "UNSPECIFIED" - Serving status is unspecified.
	//   "SERVING" - Application is serving.
	//   "USER_DISABLED" - Application has been disabled by the user.
	//   "SYSTEM_DISABLED" - Application has been disabled by the system.
	ServingStatus string `json:"servingStatus,omitempty"`
}

// A ApplicationSpec defines the desired state of a Application.
type ApplicationSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ApplicationParameters `json:"forProvider"`
}

// A ApplicationStatus represents the observed state of a Application.
type ApplicationStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ApplicationObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Application is a managed resource that represents a Google Kubernetes Engine
// cluster.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application items
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}
