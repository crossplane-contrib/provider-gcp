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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SubscriptionParameters defines parameters for a desired Subscription.
type SubscriptionParameters struct {
	// AckDeadlineSeconds is the approximate amount of time Pub/Sub waits for
	// the subscriber to acknowledge receipt before resending the message.
	// The minimum custom deadline you can specify is 10 seconds. The maximum
	// custom deadline you can specify is 600 seconds (10 minutes). If this
	// parameter is 0, a default value of 10 seconds is used.
	// +optional
	AckDeadlineSeconds int64 `json:"ackDeadlineSeconds,omitempty"`

	// DeadLetterPolicy is the policy that specifies the conditions for dead
	// lettering messages in this subscription. If dead_letter_policy is not
	// set, dead lettering is disabled.
	// +optional
	DeadLetterPolicy *DeadLetterPolicy `json:"deadLetterPolicy,omitempty"`

	// Detached is the flag which indicates whether the subscription is detached from its
	// topic. Detached subscriptions don't receive messages from their topic
	// and don't retain any backlog.
	// +optional
	Detached bool `json:"detached,omitempty"`

	// EnableMessageOrdering is the flag which controls message delivery order
	// to subscribers. When it is true, messages published with the same
	// `ordering_key` in `PubsubMessage` will be delivered to the subscribers
	// in the order in which they are received by the Pub/Sub system.
	// Otherwise, they may be delivered in any order.
	// +optional
	EnableMessageOrdering bool `json:"enableMessageOrdering,omitempty"`

	// ExpirationPolicy is the policy that specifies the conditions for this
	// subscription's expiration. If `expiration_policy` is not set, a
	// *default policy* with `ttl` of 31 days will be used. The minimum allowed value
	// for `expiration_policy.ttl` is 1 day.
	// +optional
	ExpirationPolicy *ExpirationPolicy `json:"expirationPolicy,omitempty"`

	// Filter is an expression written in the Pub/Sub filter language
	// (https://cloud.google.com/pubsub/docs/filtering). If non-empty, then
	// only `PubsubMessage`s whose `attributes` field matches the filter are
	// delivered on this subscription. If empty, then no messages are
	// filtered out.
	// +optional
	Filter string `json:"filter,omitempty"`

	// Labels are used as additional metadata on Subscription.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// MessageRetentionDuration is a parameter which defines how long to retain
	// unacknowledged messages in the subscription's backlog, from the moment
	// a message is published. If `retain_acked_messages` is true, then this also
	// configures the retention of acknowledged messages, and thus
	// configures how far back in time a `Seek` can be done. Defaults to 7
	// days. Cannot be more than 7 days or less than 10 minutes.
	// +optional
	MessageRetentionDuration string `json:"messageRetentionDuration,omitempty"`

	// PushConfig is a parameter which configures push delivery. An empty
	// `pushConfig` signifies that the subscriber will pull and ack messages
	// using API methods.
	// +optional
	PushConfig *PushConfig `json:"pushConfig,omitempty"`

	// RetainAckedMessages is a message which indicates whether to retain acknowledged
	// messages. If true, then messages are not expunged from the
	// subscription's backlog, even if they are acknowledged, until they
	// fall out of the `message_retention_duration` window.
	// +optional
	RetainAckedMessages bool `json:"retainAckedMessages,omitempty"`

	// RetryPolicy is the policy that specifies how Pub/Sub retries message
	// delivery for this subscription. If not set, the default retry policy
	// is applied. This generally implies that messages will be retried as
	// soon as possible for healthy subscribers.
	// +optional
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`

	// TODO: Add referencer & selector for Topic resource.
	// Topic is the name of the topic from which this subscription
	// is receiving messages. Format is `projects/{project}/topics/{topic}`.
	Topic string `json:"topic,omitempty"`
}

// DeadLetterPolicy contains configuration for dead letter policy.
type DeadLetterPolicy struct {
	// DeadLetterTopic is the name of the topic to which dead letter messages
	// should be published. Format is `projects/{project}/topics/{topic}`.
	DeadLetterTopic string `json:"deadLetterTopic,omitempty"`

	// MaxDeliveryAttempts is the maximum number of delivery attempts for any
	// message. The value must be between 5 and 100.
	// +optional
	MaxDeliveryAttempts int64 `json:"maxDeliveryAttempts,omitempty"`
}

// ExpirationPolicy contains configuration for resource expiration.
type ExpirationPolicy struct {
	// Ttl is the duration of "time-to-live" for an associated resource.
	// The resource expires if it is not active for a period of `ttl`.
	Ttl string `json:"ttl,omitempty"`
}

// PushConfig contains configuration for a push delivery endpoint.
type PushConfig struct {
	// Attributes is the map of endpoint configuration attributes that can be used to
	// control different aspects of the message delivery.
	Attributes map[string]string `json:"attributes,omitempty"`

	// OidcToken is a set of parameters to attach OIDC JWT
	// token as an `Authorization` header in the HTTP request for every
	// pushed message.
	OidcToken *OidcToken `json:"oidcToken,omitempty"`

	// PushEndpoint is a URL locating the endpoint to which messages should be
	// pushed.
	PushEndpoint string `json:"pushEndpoint,omitempty"`
}

// OidcToken contains information needed for generating an OpenID Connect token
type OidcToken struct {
	// Audience is the "audience" to be used when generating OIDC token.
	Audience string `json:"audience,omitempty"`

	// ServiceAccountEmail is the email to be used for generating the OIDC token
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty"`
}

// RetryPolicy is the policy that specifies how Cloud Pub/Sub retries
// message delivery. Retry delay will be exponential based on provided
// minimum and maximum backoffs.
type RetryPolicy struct {
	// MaximumBackoff is the maximum delay between consecutive deliveries of a
	// given message. Value should be between 0 and 600 seconds. Defaults to
	// 600 seconds.
	MaximumBackoff string `json:"maximumBackoff,omitempty"`

	// MinimumBackoff is the minimum delay between consecutive deliveries of a
	// given message. Value should be between 0 and 600 seconds. Defaults to
	// 10 seconds.
	MinimumBackoff string `json:"minimumBackoff,omitempty"`
}

// SubscriptionSpec defines the desired state of a Subscription.
type SubscriptionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SubscriptionParameters `json:"forProvider"`
}

// SubscriptionStatus represents the observed state of a Subscription.
type SubscriptionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// Subscription is a managed resource that represents a Google PubSub Subscription.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Cluster
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionSpec   `json:"spec"`
	Status SubscriptionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubscriptionList contains a list of Subscription types
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}
