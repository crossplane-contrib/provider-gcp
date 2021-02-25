package v1alpha1

import (
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceRecordSetParameters define the desired state of a Google DNS ResourceRecordSet
// Network. Most fields map directly to a Network:
// https://cloud.google.com/dns/docs/reference/v1beta2/resourceRecordSets
type ResourceRecordSetParameters struct {
	// ManagedZone: Identifies the managed zone addressed by this request. Can be the managed zone name or id.
	// +immutable
	ManagedZone string `json:"managedZone"`

	// Name: The name of the recordset. For example, www.example.com.
	// +immutable
	Name string `json:"name"`

	// RRdatas:	As defined in RFC 1035 (section 5) and RFC 1034 (section 3.6.1) -- see examples.
	Rrdatas []string `json:"rrdatas,omitempty"`

	// SignatureRrdatas: As defined in RFC 4034 (section 3.2).
	SignatureRrdatas []string `json:"signatureRrdatas,omitempty"`

	// TTL: Number of seconds that this ResourceRecordSet can be cached by resolvers.
	TTL int `json:"ttl"`

	// Type: The identifier of a supported record type. See the list of Supported DNS record types.
	// https://cloud.google.com/dns/docs/overview#supported_dns_record_types
	//
	// Possible values:
	//   "A"
	//   "AAAA"
	//   "CAA"
	//   "CNAME"
	//   "IPSECKEY"
	//   "MX"
	//   "NAPTR"
	//   "NS"
	//   "PTR"
	//   "SOA"
	//   "SPF"
	//   "SRV"
	//   "SSHFP"
	//   "TLSA"
	//   "TXT"
	// +immutable
	// +kubebuilder:validation:Enum=A;AAAA;CAA;CNAME;IPSECKEY;MX;NAPTR;NS;PTR;SOA;SPF;SRV;SSHFP;TLSA;TXT
	Type string `json:"type"`
}

// A ResourceRecordSetObservation represents the observed state of a Google DNS ResourceRecordSet.
type ResourceRecordSetObservation struct {
	// Name: The name of the recordset. For example, www.example.com.
	Name string `json:"name,omitempty"`

	// RRdatas:	As defined in RFC 1035 (section 5) and RFC 1034 (section 3.6.1) -- see examples.
	Rrdatas []string `json:"rrdatas,omitempty"`

	// SignatureRrdatas: As defined in RFC 4034 (section 3.2).
	SignatureRrdatas []string `json:"signatureRrdatas,omitempty"`

	// TTL: Number of seconds that this ResourceRecordSet can be cached by resolvers.
	TTL int `json:"ttl,omitempty"`

	// Type: The identifier of a supported record type. See the list of Supported DNS record types.
	Type string `json:"type,omitempty"`
}

// A ResourceRecordSetSpec defines the desired state of a ResourceRecordSet.
type ResourceRecordSetSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     ResourceRecordSetParameters `json:"forProvider"`
}

// A ResourceRecordSetStatus represents the observed state of a ResourceRecordSet.
type ResourceRecordSetStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        ResourceRecordSetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ResourceRecordSet is a managed resource that represents a Google Compute Engine DNS ResourceRecordSet.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type ResourceRecordSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceRecordSetSpec   `json:"spec"`
	Status ResourceRecordSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceRecordSetList contains a list of ResourceRecordSet.
type ResourceRecordSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceRecordSet `json:"items"`
}
