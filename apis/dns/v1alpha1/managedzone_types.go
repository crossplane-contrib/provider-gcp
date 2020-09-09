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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// ManagedZoneParameters defines parameters for a ManagedZone.
type ManagedZoneParameters struct {
	// Name: User assigned name for this resource. Must be unique within the
	// project. The name must be 1-63 characters long, must begin with a
	// letter, end with a letter or digit, and only contain lowercase
	// letters, digits or dashes.
	Name string `json:"name,omitempty"`

	// Labels are used as additional metadata on ManagedZone.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Description: A mutable string of at most 1024 characters associated
	// with this resource for the user's convenience. Has no effect on the
	// managed zone's function.
	// +optional
	Description string `json:"description,omitempty"`

	// Network: the VPC network to bind to.
	Network string `json:"network,omitempty"`

	// DNSName: The DNS name of this managed zone, for instance
	// "example.com.".
	DNSName string `json:"dnsName,omitempty"`

	// Visibility: The zone's visibility: public zones are exposed to the
	// Internet, while private zones are visible only to Virtual Private
	// Cloud resources.
	//
	// Possible values:
	//   "private"
	//   "public"
	// +optional
	Visibility string `json:"visibility,omitempty"`
}

// ManagedZoneObservation is used to show the observed state of the
// ManagedZone resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type ManagedZoneObservation struct {
}

// ManagedZoneSpec defines the desired state of a
// ManagedZone.
type ManagedZoneSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ManagedZoneParameters `json:"forProvider"`
}

// ManagedZoneStatus represents the observed state of a
// ManagedZone.
type ManagedZoneStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ManagedZoneObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ManagedZone is a managed resource that represents a Google IAM Service Account.
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type ManagedZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedZoneSpec   `json:"spec"`
	Status ManagedZoneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ManagedZoneList contains a list of ManagedZone types
type ManagedZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedZone `json:"items"`
}
