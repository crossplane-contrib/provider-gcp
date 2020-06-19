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

package topic

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/crossplane/provider-gcp/apis/pubsub/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

// GenerateTopic produces a Topic that is configured via given TopicParameters.
func GenerateTopic(projectID, name string, s v1alpha1.TopicParameters) *pubsub.Topic {
	t := &pubsub.Topic{
		Name:       fmt.Sprintf("projects/%s/topics/%s", projectID, name),
		Labels:     s.Labels,
		KmsKeyName: gcp.StringValue(s.KmsKeyName),
	}
	if s.MessageStoragePolicy != nil {
		t.MessageStoragePolicy = &pubsub.MessageStoragePolicy{AllowedPersistenceRegions: s.MessageStoragePolicy.AllowedPersistenceRegions}
	}
	return t
}

// LateInitialize fills the empty fields of TopicParameters if the corresponding
// fields are given in Topic.
func LateInitialize(s *v1alpha1.TopicParameters, t pubsub.Topic) {
	if len(s.Labels) == 0 && len(t.Labels) != 0 {
		s.Labels = map[string]string{}
		for k, v := range t.Labels {
			s.Labels[k] = v
		}
	}
	if s.KmsKeyName == nil && len(t.KmsKeyName) == 0 {
		s.KmsKeyName = gcp.StringPtr(t.KmsKeyName)
	}
	if s.MessageStoragePolicy == nil && t.MessageStoragePolicy != nil {
		s.MessageStoragePolicy = &v1alpha1.MessageStoragePolicy{AllowedPersistenceRegions: t.MessageStoragePolicy.AllowedPersistenceRegions}
	}
}

// IsUpToDate checks whether Topic is configured with given TopicParameters.
func IsUpToDate(s v1alpha1.TopicParameters, t pubsub.Topic) bool {
	observed := &v1alpha1.TopicParameters{}
	LateInitialize(observed, t)
	return cmp.Equal(observed, &s)
}

// GenerateUpdateRequest produces an UpdateTopicRequest with the difference
// between TopicParameters and Topic.
func GenerateUpdateRequest(projectID, name string, s v1alpha1.TopicParameters, t pubsub.Topic) *pubsub.UpdateTopicRequest {
	observed := &v1alpha1.TopicParameters{}
	LateInitialize(observed, t)
	ut := &pubsub.UpdateTopicRequest{Topic: &pubsub.Topic{Name: fmt.Sprintf("projects/%s/topics/%s", projectID, name)}, UpdateMask: &field_mask.FieldMask{}}
	if !cmp.Equal(s.MessageStoragePolicy, observed.MessageStoragePolicy) {
		ut.UpdateMask.Paths = append(ut.UpdateMask.Paths, "messageStoragePolicy")
		if s.MessageStoragePolicy != nil {
			ut.Topic.MessageStoragePolicy.AllowedPersistenceRegions = s.MessageStoragePolicy.AllowedPersistenceRegions
		}
	}
	if !cmp.Equal(s.Labels, observed.Labels) {
		ut.UpdateMask.Paths = append(ut.UpdateMask.Paths, "labels")
		ut.Topic.Labels = s.Labels
	}
	return ut
}
