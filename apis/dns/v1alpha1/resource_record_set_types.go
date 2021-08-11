/*
Copyright 2021 The Crossplane Authors.

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

// ResourceRecordSetParameters define the desired state of a ResourceRecordSet
type ResourceRecordSetParameters struct {
	// Managed zone name that this ResourceRecordSet will be created in.
	ManagedZone string `json:"managedZone"`

	// Identifies what kind of resource this is.
	//
	// +kubebuilder:validation:Enum=dns#resourceRecordSet
	Kind string `json:"kind"`

	// The identifier of a supported record type.
	//
	// +kubebuilder:validation:Enum=A;AAAA;CAA;CNAME;DNSKEY;DS;IPSECKEY;MX;NAPTR;NS;PTR;SPF;SRV;SSHFP;TLSA;TXT
	Type string `json:"type"`

	// Number of seconds that this ResourceRecordSet
	// can be cached by resolvers.
	TTL int64 `json:"ttl"`

	// List of ResourceRecord datas as defined in
	// RFC 1035 (section 5) and RFC 1034 (section 3.6.1)
	RRDatas []string `json:"rrdatas"`

	// List fo Signature ResourceRecord datas, as
	// defined in RFC 4034 (section 3.2).
	SignatureRRDatas *[]string `json:"signatureRrdatas,omitempty"`
}

// ResourceRecordSetObservation is used to show the observed state of the ResourceRecordSet
type ResourceRecordSetObservation struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// ResourceRecordSetSpec defines the desired state of a ResourceRecordSet.
type ResourceRecordSetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ResourceRecordSetParameters `json:"forProvider"`
}

// ResourceRecordSetStatus represents the observed state of a ResourceRecordSet.
type ResourceRecordSetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ResourceRecordSetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceRecordSet is a managed resource that represents a Resource Record Set in Cloud DNS
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="DNS NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp},shortName=rrs
type ResourceRecordSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceRecordSetSpec   `json:"spec"`
	Status ResourceRecordSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceRecordSetList contains a list of ResourceRecordSet
type ResourceRecordSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceRecordSet `json:"items"`
}
