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

package subscription

import (
	"fmt"
	"strings"

	"github.com/crossplane/provider-gcp/apis/pubsub/v1alpha1"
	"github.com/crossplane/provider-gcp/pkg/clients/topic"

	"github.com/google/go-cmp/cmp"
	pubsub "google.golang.org/api/pubsub/v1"
)

const (
	subscriptionNameFormat = "projects/%s/subscriptions/%s"
)

// GetFullyQualifiedName builds the fully qualified name of the subscription.
func GetFullyQualifiedName(project string, name string) string {
	return fmt.Sprintf(subscriptionNameFormat, project, name)
}

// GenerateSubscription produces a Subscription that is configured via given SubscriptionParameters.
func GenerateSubscription(projectID, name string, p v1alpha1.SubscriptionParameters) *pubsub.Subscription {
	s := &pubsub.Subscription{
		AckDeadlineSeconds:       p.AckDeadlineSeconds,
		Detached:                 p.Detached,
		EnableMessageOrdering:    p.EnableMessageOrdering,
		Filter:                   p.Filter,
		Labels:                   p.Labels,
		MessageRetentionDuration: p.MessageRetentionDuration,
		Name:                     name,
		RetainAckedMessages:      p.RetainAckedMessages,
		Topic:                    topic.GetFullyQualifiedName(projectID, p.Topic),
	}

	setDeadLetterPolicy(p, s)
	setExpirationPolicy(p, s)
	setPushConfig(p, s)
	setRetryPolicy(p, s)

	return s
}

// setRetryPolicy sets RetryPolicy of subscription based on SubscriptionParameters.
func setRetryPolicy(p v1alpha1.SubscriptionParameters, s *pubsub.Subscription) {
	if p.RetryPolicy != nil {
		s.RetryPolicy = &pubsub.RetryPolicy{
			MaximumBackoff: p.RetryPolicy.MaximumBackoff,
			MinimumBackoff: p.RetryPolicy.MinimumBackoff,
		}
	}
}

// setPushConfig sets PushConfig of subscription based on SubscriptionParameters.
func setPushConfig(p v1alpha1.SubscriptionParameters, s *pubsub.Subscription) {
	if p.PushConfig != nil {
		s.PushConfig = &pubsub.PushConfig{
			Attributes:   p.PushConfig.Attributes,
			PushEndpoint: p.PushConfig.PushEndpoint,
		}

		if p.PushConfig.OidcToken != nil {
			s.PushConfig.OidcToken = &pubsub.OidcToken{
				Audience:            p.PushConfig.OidcToken.Audience,
				ServiceAccountEmail: p.PushConfig.OidcToken.ServiceAccountEmail,
			}
		}
	}
}

// setExpirationPolicy sets ExpirationPolicy of subscription based on SubscriptionParameters.
func setExpirationPolicy(p v1alpha1.SubscriptionParameters, s *pubsub.Subscription) {
	if p.ExpirationPolicy != nil {
		s.ExpirationPolicy = &pubsub.ExpirationPolicy{
			Ttl: p.ExpirationPolicy.TTL,
		}
	}
}

// setDeadLetterPolicy sets DeadLetterPolicy of subscription based on SubscriptionParameters.
func setDeadLetterPolicy(p v1alpha1.SubscriptionParameters, s *pubsub.Subscription) {
	if p.DeadLetterPolicy != nil {
		s.DeadLetterPolicy = &pubsub.DeadLetterPolicy{
			DeadLetterTopic:     p.DeadLetterPolicy.DeadLetterTopic,
			MaxDeliveryAttempts: p.DeadLetterPolicy.MaxDeliveryAttempts,
		}
	}
}

// LateInitialize fills the empty fields of SubscriptionParameters if the corresponding
// fields are given in Subscription.
func LateInitialize(projectID string, p *v1alpha1.SubscriptionParameters, s pubsub.Subscription) {
	if (p.AckDeadlineSeconds == 10 || p.AckDeadlineSeconds == 0) && s.AckDeadlineSeconds != 0 {
		p.AckDeadlineSeconds = s.AckDeadlineSeconds
	}

	if !p.Detached && s.Detached {
		p.Detached = s.Detached
	}

	if !p.EnableMessageOrdering && s.EnableMessageOrdering {
		p.EnableMessageOrdering = s.EnableMessageOrdering
	}

	if p.Filter == "" && s.Filter != "" {
		p.Filter = s.Filter
	}

	if len(p.Labels) == 0 && len(s.Labels) != 0 {
		p.Labels = map[string]string{}
		for k, v := range s.Labels {
			p.Labels[k] = v
		}
	}

	if p.MessageRetentionDuration == "" && s.MessageRetentionDuration != "" {
		p.MessageRetentionDuration = s.MessageRetentionDuration
	}

	if !p.RetainAckedMessages && s.RetainAckedMessages {
		p.RetainAckedMessages = s.RetainAckedMessages
	}

	if p.Topic == "" && s.Topic != "" {
		p.Topic = s.Topic
	}

	if p.DeadLetterPolicy == nil && s.DeadLetterPolicy != nil {
		p.DeadLetterPolicy = &v1alpha1.DeadLetterPolicy{
			DeadLetterTopic:     s.DeadLetterPolicy.DeadLetterTopic,
			MaxDeliveryAttempts: s.DeadLetterPolicy.MaxDeliveryAttempts,
		}
	}

	if p.ExpirationPolicy == nil && s.ExpirationPolicy != nil {
		p.ExpirationPolicy = &v1alpha1.ExpirationPolicy{
			TTL: s.ExpirationPolicy.Ttl,
		}
	}

	if p.PushConfig == nil && s.PushConfig != nil {
		p.PushConfig = &v1alpha1.PushConfig{
			Attributes:   s.PushConfig.Attributes,
			PushEndpoint: s.PushConfig.PushEndpoint,
		}

		if p.PushConfig.OidcToken == nil && s.PushConfig.OidcToken != nil {
			p.PushConfig.OidcToken = &v1alpha1.OidcToken{
				Audience:            s.PushConfig.OidcToken.Audience,
				ServiceAccountEmail: s.PushConfig.OidcToken.ServiceAccountEmail,
			}
		}
	}

	if p.RetryPolicy == nil && s.RetryPolicy != nil {
		p.RetryPolicy = &v1alpha1.RetryPolicy{
			MaximumBackoff: s.RetryPolicy.MaximumBackoff,
			MinimumBackoff: s.RetryPolicy.MinimumBackoff,
		}
	}
}

// IsUpToDate checks whether Subscription is configured with given SubscriptionParameters.
func IsUpToDate(projectID string, p v1alpha1.SubscriptionParameters, s pubsub.Subscription) bool {
	observed := &v1alpha1.SubscriptionParameters{}
	LateInitialize(projectID, observed, s)
	if p.Topic != "" {
		p.Topic = topic.GetFullyQualifiedName(projectID, p.Topic)
	}

	return cmp.Equal(observed, &p)
}

// GenerateUpdateRequest produces an UpdateSubscriptionRequest with the difference
// between SubscriptionParameters and Subscription.
// enableMessageOrdering, deadLetterPolicy, topic are not mutable
func GenerateUpdateRequest(projectID, name string, p v1alpha1.SubscriptionParameters, s pubsub.Subscription) *pubsub.UpdateSubscriptionRequest {
	observed := &v1alpha1.SubscriptionParameters{}
	LateInitialize(projectID, observed, s)

	us := &pubsub.UpdateSubscriptionRequest{
		Subscription: &pubsub.Subscription{Name: name},
	}

	mask := []string{}

	if !cmp.Equal(p.AckDeadlineSeconds, observed.AckDeadlineSeconds) {
		mask = append(mask, "ackDeadlineSeconds")
		us.Subscription.AckDeadlineSeconds = p.AckDeadlineSeconds
	}

	if !cmp.Equal(p.Detached, observed.Detached) {
		mask = append(mask, "detached")
		us.Subscription.Detached = p.Detached
	}

	if !cmp.Equal(p.Filter, observed.Filter) {
		mask = append(mask, "filter")
		us.Subscription.Filter = p.Filter
	}

	if !cmp.Equal(p.Labels, observed.Labels) {
		mask = append(mask, "labels")
		us.Subscription.Labels = p.Labels
	}

	if !cmp.Equal(p.MessageRetentionDuration, observed.MessageRetentionDuration) {
		mask = append(mask, "messageRetentionDuration")
		us.Subscription.MessageRetentionDuration = p.MessageRetentionDuration
	}

	if !cmp.Equal(p.RetainAckedMessages, observed.RetainAckedMessages) {
		mask = append(mask, "retainAckedMessages")
		us.Subscription.RetainAckedMessages = p.RetainAckedMessages
	}

	if !cmp.Equal(p.ExpirationPolicy, observed.ExpirationPolicy) {
		mask = append(mask, "expirationPolicy")
		setExpirationPolicy(p, us.Subscription)
	}

	if !cmp.Equal(p.PushConfig, observed.PushConfig) {
		mask = append(mask, "pushConfig")
		setPushConfig(p, us.Subscription)
	}

	if !cmp.Equal(p.RetryPolicy, observed.RetryPolicy) {
		mask = append(mask, "retryPolicy")
		setRetryPolicy(p, us.Subscription)
	}

	us.UpdateMask = strings.Join(mask, ",")

	return us
}
