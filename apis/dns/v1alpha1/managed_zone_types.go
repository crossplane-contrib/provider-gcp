/*
Copyright 2022 The Crossplane Authors.

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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ManagedZoneParameters define the desired state of a ManagedZone
type ManagedZoneParameters struct {

	// Description: A mutable string of at most 1024 characters associated
	// with this resource for the user's convenience. Has no effect on the
	// managed zone's function. Defaults to 'Managed by Crossplane'
	// +optional
	Description *string `json:"description,omitempty"`

	// DNSName: The DNS name of this managed zone, for instance "example.com.".
	// +immutable
	DNSName string `json:"dnsName"`

	// Labels: User labels.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// PrivateVisibilityConfig: For privately visible zones, the set of
	// Virtual Private Cloud resources that the zone is visible from.
	// +optional
	PrivateVisibilityConfig *ManagedZonePrivateVisibilityConfig `json:"privateVisibilityConfig,omitempty"`

	// Visibility: The zone's visibility: public zones are exposed to the
	// Internet, while private zones are visible only to Virtual Private
	// Cloud resources. Defaults to 'public`
	//
	// Possible values:
	//   "public"
	//   "private"
	// +optional
	// +immutable
	// +kubebuilder:validation:Enum=public;private
	Visibility *string `json:"visibility,omitempty"`

	// TODO(danielinclouds): support CloudLoggingConfig parameters
	// TODO(danielinclouds): support DnssecConfig parameters
	// TODO(danielinclouds): support ForwardingConfig parameters
	// TODO(danielinclouds): support NameServerSet parameters
	// TODO(danielinclouds): support PeeringConfig parameters
	// TODO(danielinclouds): support ReverseLookupConfig parameters
	// TODO(danielinclouds): support ServiceDirectoryConfig parameters
}

// ManagedZonePrivateVisibilityConfig the set of Virtual Private Cloud resources
// that the zone is visible from
type ManagedZonePrivateVisibilityConfig struct {

	// Networks: The list of VPC networks that can see this zone.
	Networks []*ManagedZonePrivateVisibilityConfigNetwork `json:"networks"`
}

// ManagedZonePrivateVisibilityConfigNetwork is a list of VPC networks
type ManagedZonePrivateVisibilityConfigNetwork struct {

	// NetworkUrl: The fully qualified URL of the VPC network to bind to.
	// Format this URL like
	// https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}
	// +optional
	// +immutable
	NetworkURL *string `json:"networkUrl,omitempty"`

	// NetworkRef references to a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkRef *xpv1.Reference `json:"networkRef,omitempty"`

	// NetworkSelector selects a reference to a Network and retrieves its URI
	// +optional
	// +immutable
	NetworkSelector *xpv1.Selector `json:"networkSelector,omitempty"`
}

// ManagedZoneObservation is used to show the observed state of the ManagedZone
type ManagedZoneObservation struct {

	// CreationTime: The time that this resource was created on the server.
	// This is in RFC3339 text format. Output only.
	CreationTime string `json:"creationTime,omitempty"`

	// Id: Unique identifier for the resource; defined by the server (output only)
	ID uint64 `json:"id,omitempty"`

	// NameServers: Delegate your managed_zone to these virtual name
	// servers; defined by the server (output only)
	NameServers []string `json:"nameServers,omitempty"`
}

// ManagedZoneSpec defines the desired state of a ManagedZone.
type ManagedZoneSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ManagedZoneParameters `json:"forProvider"`
}

// ManagedZoneStatus represents the observed state of a ManagedZone.
type ManagedZoneStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ManagedZoneObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ManagedZone is a managed resource that represents a Managed Zone in Cloud DNS
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="DNS NAME",type="string",JSONPath=".spec.forProvider.dnsName"
// +kubebuilder:printcolumn:name="VISIBILITY",type="string",JSONPath=".spec.forProvider.visibility"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type ManagedZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedZoneSpec   `json:"spec"`
	Status ManagedZoneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ManagedZoneList contains a list of ManagedZones
type ManagedZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedZone `json:"items"`
}
