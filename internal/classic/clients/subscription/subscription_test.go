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
	"testing"

	"github.com/google/go-cmp/cmp"
	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/crossplane/provider-gcp/apis/classic/pubsub/v1alpha1"
)

const (
	projectID         = "my-project"
	name              = "my-subscription"
	topicName         = "my-topic"
	topicNameExternal = "projects/my-project/topics/my-topic"
)

func params() *v1alpha1.SubscriptionParameters {
	return &v1alpha1.SubscriptionParameters{
		AckDeadlineSeconds: 15,
		DeadLetterPolicy: &v1alpha1.DeadLetterPolicy{
			DeadLetterTopic:     topicName,
			MaxDeliveryAttempts: 5,
		},
		Detached:                 true,
		EnableMessageOrdering:    true,
		ExpirationPolicy:         &v1alpha1.ExpirationPolicy{TTL: "1296000s"},
		Filter:                   "foo",
		Labels:                   map[string]string{"example": "true"},
		MessageRetentionDuration: "864000s",
		PushConfig: &v1alpha1.PushConfig{
			Attributes: map[string]string{"attribute": "my-attribute"},
			OidcToken: &v1alpha1.OidcToken{
				Audience:            "my-audience",
				ServiceAccountEmail: "example@gmail.coom",
			},
			PushEndpoint: "example.com",
		},
		RetryPolicy: &v1alpha1.RetryPolicy{
			MaximumBackoff: "100s",
			MinimumBackoff: "15s",
		},
		RetainAckedMessages: true,
		Topic:               topicName,
	}
}

func subscription() *pubsub.Subscription {
	return &pubsub.Subscription{
		Name:               name,
		AckDeadlineSeconds: 15,
		DeadLetterPolicy: &pubsub.DeadLetterPolicy{
			DeadLetterTopic:     topicNameExternal,
			MaxDeliveryAttempts: 5,
		},
		Detached:              true,
		EnableMessageOrdering: true,
		ExpirationPolicy:      &pubsub.ExpirationPolicy{Ttl: "1296000s"},
		Filter:                "foo",
		Labels: map[string]string{
			"example": "true",
		},
		MessageRetentionDuration: "864000s",
		PushConfig: &pubsub.PushConfig{
			Attributes: map[string]string{"attribute": "my-attribute"},
			OidcToken: &pubsub.OidcToken{
				Audience:            "my-audience",
				ServiceAccountEmail: "example@gmail.coom",
			},
			PushEndpoint: "example.com",
		},
		RetryPolicy: &pubsub.RetryPolicy{
			MaximumBackoff: "100s",
			MinimumBackoff: "15s",
		},
		RetainAckedMessages: true,
		Topic:               topicNameExternal,
	}
}

func TestGenerateSubscription(t *testing.T) {
	type args struct {
		projectID string
		name      string
		s         v1alpha1.SubscriptionParameters
	}
	cases := map[string]struct {
		args
		out *pubsub.Subscription
	}{
		"Full": {
			args: args{
				projectID: projectID,
				name:      name,
				s:         *params(),
			},
			out: subscription(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSubscription(tc.projectID, tc.name, tc.s)
			if diff := cmp.Diff(tc.out, got); diff != "" {
				t.Errorf("GenerateSubscription(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		obs   pubsub.Subscription
		param *v1alpha1.SubscriptionParameters
	}
	cases := map[string]struct {
		args
		out *v1alpha1.SubscriptionParameters
	}{
		"Full": {
			args: args{
				obs: *subscription(),
				param: &v1alpha1.SubscriptionParameters{
					AckDeadlineSeconds: 15,
					DeadLetterPolicy: &v1alpha1.DeadLetterPolicy{
						DeadLetterTopic:     topicName,
						MaxDeliveryAttempts: 5,
					},
					Detached:                 true,
					EnableMessageOrdering:    true,
					ExpirationPolicy:         &v1alpha1.ExpirationPolicy{TTL: "1296000s"},
					Filter:                   "foo",
					Labels:                   map[string]string{"example": "true"},
					MessageRetentionDuration: "864000s",
					PushConfig: &v1alpha1.PushConfig{
						Attributes: map[string]string{"attribute": "my-attribute"},
						OidcToken: &v1alpha1.OidcToken{
							Audience:            "my-audience",
							ServiceAccountEmail: "example@gmail.coom",
						},
						PushEndpoint: "example.com",
					},
					RetryPolicy: &v1alpha1.RetryPolicy{
						MaximumBackoff: "100s",
						MinimumBackoff: "15s",
					},
					RetainAckedMessages: true,
					Topic:               topicName,
				},
			},
			out: params(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.args.param, tc.args.obs)
			if diff := cmp.Diff(tc.args.param, tc.out); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		obs   pubsub.Subscription
		param v1alpha1.SubscriptionParameters
	}
	cases := map[string]struct {
		args
		result bool
	}{
		"NotUpToDate": {
			args: args{
				obs: *subscription(),
				param: v1alpha1.SubscriptionParameters{
					RetryPolicy: nil,
				},
			},
			result: false,
		},
		"UpToDate": {
			args: args{
				obs:   *subscription(),
				param: *params(),
			},
			result: true,
		},
	}

	IsUpToDate(projectID, *params(), *subscription())
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsUpToDate(projectID, tc.args.param, tc.args.obs)
			if diff := cmp.Diff(tc.result, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateRequest(t *testing.T) {
	mutableSubscription := subscription()
	mutableSubscription.Topic = ""
	mutableSubscription.EnableMessageOrdering = false
	mutableSubscription.DeadLetterPolicy = nil

	type args struct {
		projectID string
		name      string
		obs       pubsub.Subscription
		param     v1alpha1.SubscriptionParameters
	}

	cases := map[string]struct {
		args
		result *pubsub.UpdateSubscriptionRequest
	}{
		"Full": {
			args: args{
				projectID: projectID,
				name:      name,
				obs:       pubsub.Subscription{},
				param:     *params(),
			},
			result: &pubsub.UpdateSubscriptionRequest{
				Subscription: mutableSubscription,
				UpdateMask:   "ackDeadlineSeconds,detached,filter,labels,messageRetentionDuration,retainAckedMessages,expirationPolicy,pushConfig,retryPolicy",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateRequest(tc.args.name, tc.args.param, tc.args.obs)
			if diff := cmp.Diff(tc.result, got); diff != "" {
				t.Errorf("GenerateUpdateRequest(...): -want, +got:\n%s", diff)
			}
		})
	}
}
