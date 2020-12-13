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

// Keys used in connection secret.
const (
	ConnectionSecretKeyTopic       = "topic"
	ConnectionSecretKeyProjectName = "projectName"
)

// TopicParameters defines parameters for a desired PubSub Topic.
type TopicParameters struct {
	// Labels are used as additional metadata on Topic.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// MessageStoragePolicy is the policy constraining the set of Google Cloud
	// Platform regions where messages published to the topic may be stored. If
	// not present, then no constraints are in effect.
	// +optional
	MessageStoragePolicy *MessageStoragePolicy `json:"messageStoragePolicy,omitempty"`

	// TODO(muvaf): Add referencer & selector when we have KMS as managed resource.

	// KmsKeyName is the resource name of the Cloud KMS CryptoKey to be used to
	// protect access to messages published on this topic.
	//
	// The expected format is `projects/*/locations/*/keyRings/*/cryptoKeys/*`.
	// +optional
	// +immutable
	KmsKeyName *string `json:"kmsKeyName,omitempty"`
}

// MessageStoragePolicy contains configuration for message storage policy.
type MessageStoragePolicy struct {
	// AllowedPersistenceRegions is the list of IDs of GCP regions where messages
	// that are published to the topic may be persisted in storage. Messages
	// published by publishers running in non-allowed GCP regions (or running
	// outside of GCP altogether) will be routed for storage in one of the
	// allowed regions. An empty list means that no regions are allowed, and is
	// not a valid configuration.
	AllowedPersistenceRegions []string `json:"allowedPersistenceRegions,omitempty"`
}

// TopicSpec defines the desired state of a
// Topic.
type TopicSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       TopicParameters `json:"forProvider"`
}

// TopicStatus represents the observed state of a
// Topic.
type TopicStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// Topic is a managed resource that represents a Google PubSub Topic.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec"`
	Status TopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicList contains a list of Topic types
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}
