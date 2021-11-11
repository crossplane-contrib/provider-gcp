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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ContainerRegistryParameters define the desired state of a ContainerRegistry
type ContainerRegistryParameters struct {
	// The location of the registry.
	// Possible Values: ASIA, EU, US
	// +optional
	// +immutable
	Location string `json:"location,omitempty"`
}

// ContainerRegistryObservation is used to show the observed state of the ContainerRegistry
type ContainerRegistryObservation struct {
	// The name of the bucket.
	ID string `json:"id,omitempty"`

	// The URI of the bucket.
	BucketLink string `json:"bucketLink,omitempty"`
}

// ContainerRegistrySpec defines the desired state of ContainerRegistry
type ContainerRegistrySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ContainerRegistryParameters `json:"forProvider"`
}

// ContainerRegistryStatus defines the observed state of ContainerRegistry
type ContainerRegistryStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ContainerRegistryObservation `json:"atProvider,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}

// ContainerRegistry is the Schema for the containerregistries API
type ContainerRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerRegistrySpec   `json:"spec,omitempty"`
	Status ContainerRegistryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContainerRegistryList contains a list of ContainerRegistry
type ContainerRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContainerRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContainerRegistry{}, &ContainerRegistryList{})
}
