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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// ServiceAccountParameters defines parameters for a desired IAM ServiceAccount
// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts
// The name of the service account (ie the `accountId` parameter of the Create
// call) is determined by the value of the `crossplane.io/external-name`
// annotation. Unless overridden by the user, this annotation is automatically
// populated with the value of the `metadata.name` attribute.
type ServiceAccountParameters struct {
	// DisplayName is an optional user-specified name for the service account.
	// Must be less than or equal to 100 characters.
	// +optional
	DisplayName *string `json:"displayName,omitempty"`

	// Description is an optional user-specified opaque description of the
	// service account. Must be less than or equal to 256 characters.
	// +optional
	Description *string `json:"description,omitempty"`
}

// ServiceAccountObservation is used to show the observed state of the
// ServiceAccount resource on GCP. All fields in this structure should only
// be populated from GCP responses; any changes made to the k8s resource outside
// of the crossplane gcp controller will be ignored and overwritten.
type ServiceAccountObservation struct {
	// Name is the "relative resource name" of the service account in the following format:
	// projects/{PROJECT_ID}/serviceAccounts/{external-name}.
	// part of https://godoc.org/google.golang.org/genproto/googleapis/iam/admin/v1#ServiceAccount
	// not to be confused with CreateServiceAccountRequest.Name aka ServiceAccountParameters.ProjectName
	Name string `json:"name,omitempty"`

	// ProjectID is the id of the project that owns the service account.
	ProjectID string `json:"projectId,omitempty"`

	//The unique and stable id of the service account.
	UniqueID string `json:"uniqueId,omitempty"`

	// Email is the the email address of the service account.
	// This matches the EMAIL field you would see using `gcloud iam service-accounts list`
	Email string `json:"email,omitempty"`

	// OAuth2ClientId is the value GCP will use in conjunction with the OAuth2
	// clientconfig API to make three legged OAuth2 (3LO) flows to access the
	// data of Google users.
	Oauth2ClientID string `json:"oauth2ClientId,omitempty"`

	// Disabled is a bool indicating if the service account is disabled.
	// The field is currently in alpha phase.
	Disabled bool `json:"disabled,omitempty"`
}

// ServiceAccountSpec defines the desired state of a
// ServiceAccount.
type ServiceAccountSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ServiceAccountParameters `json:"forProvider"`
}

// ServiceAccountStatus represents the observed state of a
// ServiceAccount.
type ServiceAccountStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ServiceAccountObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccount is a managed resource that represents a Google IAM Service Account.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="DISPLAYNAME",type="string",JSONPath=".spec.forProvider.displayName"
// +kubebuilder:printcolumn:name="EMAIL",type="string",JSONPath=".status.atProvider.email"
// +kubebuilder:printcolumn:name="DISABLED",type="boolean",JSONPath=".status.atProvider.disabled"
// +kubebuilder:resource:scope=Cluster
type ServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceAccountSpec   `json:"spec"`
	Status ServiceAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceAccountList contains a list of ServiceAccount types
type ServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccount `json:"items"`
}
