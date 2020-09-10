package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// ResourceRecordSetParameters define the desired state of an gcp dns Resource Record.
type ResourceRecordSetParameters struct {
	// Rrdatas: As defined in RFC 1035 (section 5) and RFC 1034 (section
	// 3.6.1)
	// +optional
	Rrdatas []string `json:"aliasTarget,omitempty"`

	// Ttl: Number of seconds that this ResourceRecordSet can be cached by
	// resolvers.
	// +optional
	TTL *int64 `json:"ttl,omitempty"`

	// Type: The identifier of a supported record type. See the list of
	// Supported DNS record types.
	Type *string `json:"type"`

	// ZoneID is the ID of the hosted zone that contains the resource record sets
	// that you want to change.
	ZoneID *string `json:"zoneId,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceRecordSet is a managed resource that represents an gcp dns Resource Record.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type ResourceRecordSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceRecordSetSpec   `json:"spec"`
	Status ResourceRecordSetStatus `json:"status,omitempty"`
}

// ResourceRecordSetSpec defines the desired state of an gcp dns Resource Record.
type ResourceRecordSetSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ResourceRecordSetParameters `json:"forProvider"`
}

// ResourceRecordSetStatus represents the observed state of a ResourceRecordSet.
type ResourceRecordSetStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// ResourceRecordSetList contains a list of ResourceRecordSet.
type ResourceRecordSetList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResourceRecordSet `json:"items"`
}
