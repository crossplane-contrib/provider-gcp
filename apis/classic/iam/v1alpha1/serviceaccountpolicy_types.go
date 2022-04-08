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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ServiceAccountPolicyParameters defines parameters for a desired IAM ServiceAccountPolicy
type ServiceAccountPolicyParameters struct {
	// ServiceAccountRef is a reference to a ServiceAccount which this policy is associated with
	ServiceAccountReferer `json:",inline"`

	// Policy: An Identity and Access Management (IAM) policy, which
	// specifies access controls for Google Cloud resources.
	Policy Policy `json:"policy"`
}

// ServiceAccountPolicySpec defines the desired state of a
// ServiceAccountPolicy.
type ServiceAccountPolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ServiceAccountPolicyParameters `json:"forProvider"`
}

// ServiceAccountPolicyStatus represents the observed state of a
// ServiceAccountPolicy.
type ServiceAccountPolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// ServiceAccountPolicy is a managed resource that represents a Google IAM ServiceAccount.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type ServiceAccountPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceAccountPolicySpec   `json:"spec"`
	Status ServiceAccountPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountPolicyList contains a list of ServiceAccountPolicy types
type ServiceAccountPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccountPolicy `json:"items"`
}
