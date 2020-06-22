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
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/crossplane/provider-gcp/apis/pubsub/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const (
	projectID = "fooproject"
	name      = "barname"
)

func params() *v1alpha1.TopicParameters {
	return &v1alpha1.TopicParameters{
		Labels: map[string]string{
			"foo": "bar",
		},
		MessageStoragePolicy: &v1alpha1.MessageStoragePolicy{
			AllowedPersistenceRegions: []string{"bar", "foo"},
		},
		KmsKeyName: gcp.StringPtr("mykms"),
	}
}

func topic() *pubsub.Topic {
	return &pubsub.Topic{
		Name: fmt.Sprintf("projects/%s/topics/%s", projectID, name),
		Labels: map[string]string{
			"foo": "bar",
		},
		MessageStoragePolicy: &pubsub.MessageStoragePolicy{
			AllowedPersistenceRegions: []string{"bar", "foo"},
		},
		KmsKeyName: "mykms",
	}
}

func TestGenerateTopic(t *testing.T) {
	type args struct {
		projectID string
		name      string
		s         v1alpha1.TopicParameters
	}
	cases := map[string]struct {
		args
		out *pubsub.Topic
	}{
		"Full": {
			args: args{
				projectID: projectID,
				name:      name,
				s:         *params(),
			},
			out: topic(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateTopic(tc.projectID, tc.name, tc.s)
			if diff := cmp.Diff(tc.out, got); diff != "" {
				t.Errorf("GenerateTopic(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		obs   pubsub.Topic
		param *v1alpha1.TopicParameters
	}
	cases := map[string]struct {
		args
		out *v1alpha1.TopicParameters
	}{
		"Full": {
			args: args{
				obs: *topic(),
				param: &v1alpha1.TopicParameters{
					KmsKeyName: params().KmsKeyName,
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
		obs   pubsub.Topic
		param v1alpha1.TopicParameters
	}
	cases := map[string]struct {
		args
		result bool
	}{
		"NotUpToDate": {
			args: args{
				obs: *topic(),
				param: v1alpha1.TopicParameters{
					KmsKeyName: params().KmsKeyName,
				},
			},
			result: false,
		},
		"UpToDate": {
			args: args{
				obs:   *topic(),
				param: *params(),
			},
			result: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsUpToDate(tc.args.param, tc.args.obs)
			if diff := cmp.Diff(tc.result, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateRequest(t *testing.T) {
	withoutKMS := topic()
	withoutKMS.KmsKeyName = ""
	type args struct {
		projectID string
		name      string
		obs       pubsub.Topic
		param     v1alpha1.TopicParameters
	}
	cases := map[string]struct {
		args
		result *pubsub.UpdateTopicRequest
	}{
		"Full": {
			args: args{
				projectID: projectID,
				name:      name,
				obs:       pubsub.Topic{},
				param:     *params(),
			},
			result: &pubsub.UpdateTopicRequest{
				Topic: withoutKMS,
				UpdateMask: &field_mask.FieldMask{Paths: []string{
					"messageStoragePolicy",
					"labels",
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateRequest(tc.args.projectID, tc.args.name, tc.args.param, tc.args.obs)
			if diff := cmp.Diff(tc.result, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
