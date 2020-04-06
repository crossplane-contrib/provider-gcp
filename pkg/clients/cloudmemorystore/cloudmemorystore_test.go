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

package cloudmemorystore

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	redisv1pb "google.golang.org/genproto/googleapis/cloud/redis/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-gcp/apis/cache/v1beta1"
)

const (
	region        = "us-cool1"
	project       = "coolProject"
	instanceName  = "claimnamespace-claimname-342sd"
	fullName      = "projects/coolProject/locations/us-cool1/instances/claimnamespace-claimname-342sd"
	qualifiedName = "projects/" + project + "/locations/" + region + "/instances/" + instanceName
	parent        = "projects/" + project + "/locations/" + region

	memorySizeGB = 1
)

var (
	locationID            = region + "-a"
	alternativeLocationID = region + "-b"
	reservedIPRange       = "172.16.0.0/16"
	authorizedNetwork     = "default"
	redisVersion          = "REDIS_3_2"
	displayName           = "my-precious-memory"

	redisConfigs = map[string]string{"cool": "socool"}
	labels       = map[string]string{"key-to": "heaven"}
	updateMask   = &field_mask.FieldMask{Paths: []string{"memory_size_gb", "redis_configs", "labels", "display_name"}}
)

func TestInstanceID(t *testing.T) {
	cases := []struct {
		name       string
		project    string
		i          *v1beta1.CloudMemorystoreInstance
		want       InstanceID
		wantName   string
		wantParent string
	}{
		{
			name:    "Success",
			project: project,
			i: &v1beta1.CloudMemorystoreInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						meta.AnnotationKeyExternalName: instanceName,
					},
				},
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						Region: region,
					},
				},
			},
			want: InstanceID{
				Project:  project,
				Region:   region,
				Instance: instanceName,
			},
			wantName:   qualifiedName,
			wantParent: parent,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewInstanceID(tc.project, tc.i)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewInstanceID(...): -want, +got:\n%s", diff)
			}

			gotName := got.Name()
			if gotName != tc.wantName {
				t.Errorf("got.Name(): want: %s got: %s", tc.wantName, gotName)
			}

			gotParent := got.Parent()
			if gotParent != tc.wantParent {
				t.Errorf("got.Parent(): want: %s got: %s", tc.wantParent, gotParent)
			}
		})
	}
}

func TestNewCreateInstanceRequest(t *testing.T) {
	cases := []struct {
		name    string
		project string
		i       *v1beta1.CloudMemorystoreInstance
		want    *redisv1pb.CreateInstanceRequest
	}{
		{
			name:    "BasicInstance",
			project: project,
			i: &v1beta1.CloudMemorystoreInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						meta.AnnotationKeyExternalName: instanceName,
					},
				},
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						Region:                region,
						Tier:                  redisv1pb.Instance_BASIC.String(),
						MemorySizeGB:          memorySizeGB,
						DisplayName:           &displayName,
						Labels:                labels,
						LocationID:            &locationID,
						AlternativeLocationID: &alternativeLocationID,
						RedisVersion:          &redisVersion,
						ReservedIPRange:       &reservedIPRange,
						RedisConfigs:          redisConfigs,
						AuthorizedNetwork:     &authorizedNetwork,
					},
				},
			},
			want: &redisv1pb.CreateInstanceRequest{
				Parent:     parent,
				InstanceId: instanceName,
				Instance: &redisv1pb.Instance{
					Name:                  qualifiedName,
					DisplayName:           displayName,
					Labels:                labels,
					LocationId:            locationID,
					AlternativeLocationId: alternativeLocationID,
					RedisVersion:          redisVersion,
					ReservedIpRange:       reservedIPRange,
					RedisConfigs:          redisConfigs,
					Tier:                  redisv1pb.Instance_BASIC,
					MemorySizeGb:          memorySizeGB,
					AuthorizedNetwork:     authorizedNetwork,
				},
			},
		},
		{
			name:    "StandardHAInstance",
			project: project,
			i: &v1beta1.CloudMemorystoreInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						meta.AnnotationKeyExternalName: instanceName,
					},
				},
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						Region:       region,
						Tier:         redisv1pb.Instance_STANDARD_HA.String(),
						MemorySizeGB: memorySizeGB,
					},
				},
			},
			want: &redisv1pb.CreateInstanceRequest{
				Parent:     parent,
				InstanceId: instanceName,
				Instance: &redisv1pb.Instance{
					Name:         qualifiedName,
					Tier:         redisv1pb.Instance_STANDARD_HA,
					MemorySizeGb: memorySizeGB,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := NewInstanceID(tc.project, tc.i)
			got := NewCreateInstanceRequest(id, tc.i)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewCreateInstanceRequest(...): -want, +got:\n%v", diff)
			}
		})
	}
}

func TestNewUpdateInstanceRequest(t *testing.T) {
	cases := []struct {
		name    string
		project string
		i       *v1beta1.CloudMemorystoreInstance
		want    *redisv1pb.UpdateInstanceRequest
	}{
		{
			name:    "UpdatableFieldsOnly",
			project: project,
			i: &v1beta1.CloudMemorystoreInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						meta.AnnotationKeyExternalName: instanceName,
					},
				},
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						Region:            region,
						RedisConfigs:      redisConfigs,
						MemorySizeGB:      memorySizeGB,
						AuthorizedNetwork: &authorizedNetwork,
					},
				},
			},
			want: &redisv1pb.UpdateInstanceRequest{
				UpdateMask: updateMask,
				Instance: &redisv1pb.Instance{
					Name:              qualifiedName,
					RedisConfigs:      redisConfigs,
					MemorySizeGb:      memorySizeGB,
					AuthorizedNetwork: authorizedNetwork,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := NewInstanceID(tc.project, tc.i)
			got := NewUpdateInstanceRequest(id, tc.i)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewUpdateInstanceRequest(...): -want, +got:\n%v", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	randString := "wat"
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := []struct {
		name string
		id   InstanceID
		kube *v1beta1.CloudMemorystoreInstance
		gcp  *redisv1pb.Instance
		want want
	}{
		{
			name: "NeedsLessMemory",
			id:   InstanceID{Project: project, Region: region, Instance: instanceName},
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
						MemorySizeGB: memorySizeGB,
					},
				},
			},
			gcp: &redisv1pb.Instance{
				Name:         fullName,
				MemorySizeGb: memorySizeGB + 1,
			},
			want: want{upToDate: false, isErr: false},
		},
		{
			name: "NeedsNewRedisConfigs",
			id:   InstanceID{Project: project, Region: region, Instance: instanceName},
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
					},
				},
			},
			gcp: &redisv1pb.Instance{
				Name:         fullName,
				RedisConfigs: map[string]string{"super": "cool"},
			},
			want: want{upToDate: false, isErr: false},
		},
		{
			name: "NeedsNoUpdate",
			id:   InstanceID{Project: project, Region: region, Instance: instanceName},
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
						MemorySizeGB: memorySizeGB,
					},
				},
			},
			gcp: &redisv1pb.Instance{
				Name:         fullName,
				RedisConfigs: redisConfigs,
				MemorySizeGb: memorySizeGB,
			},
			want: want{upToDate: true, isErr: false},
		},
		{
			name: "NeedsNoUpdateWithOutputFields",
			id:   InstanceID{Project: project, Region: region, Instance: instanceName},
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
						MemorySizeGB: memorySizeGB,
					},
				},
			},
			gcp: &redisv1pb.Instance{
				Name:          fullName,
				RedisConfigs:  redisConfigs,
				MemorySizeGb:  memorySizeGB,
				StatusMessage: "definitely not in spec",
			},
			want: want{upToDate: true, isErr: false},
		},
		{
			name: "CannotUpdateField",
			id:   InstanceID{Project: project, Region: region, Instance: instanceName},
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						MemorySizeGB: memorySizeGB,

						// We can't change this field without destroying and recreating
						// the instance so it does not register as needing an update.
						AuthorizedNetwork: &randString,
					},
				},
			},
			gcp: &redisv1pb.Instance{
				Name:              fullName,
				MemorySizeGb:      memorySizeGB,
				AuthorizedNetwork: authorizedNetwork,
			},
			want: want{upToDate: true, isErr: false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := IsUpToDate(tc.id, &tc.kube.Spec.ForProvider, tc.gcp)
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if got != tc.want.upToDate {
				t.Errorf("IsUpToDate(...): want: %t got: %t", tc.want, got)
			}
		})
	}
}

func TestNewDeleteInstanceRequest(t *testing.T) {
	cases := []struct {
		name    string
		project string
		id      InstanceID
		want    *redisv1pb.DeleteInstanceRequest
	}{
		{
			name:    "DeleteInstance",
			project: project,
			id:      InstanceID{Project: project, Region: region, Instance: instanceName},
			want:    &redisv1pb.DeleteInstanceRequest{Name: qualifiedName},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDeleteInstanceRequest(tc.id)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewDeleteInstanceRequest(...): -want, +got:\n%v", diff)
			}
		})
	}
}

func TestNewGetInstanceRequest(t *testing.T) {
	cases := []struct {
		name    string
		project string
		id      InstanceID
		want    *redisv1pb.GetInstanceRequest
	}{
		{
			name:    "GetInstance",
			project: project,
			id:      InstanceID{Project: project, Region: region, Instance: instanceName},
			want:    &redisv1pb.GetInstanceRequest{Name: qualifiedName},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewGetInstanceRequest(tc.id)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewGetInstanceRequest(...): -want, +got:\n%v", diff)
			}
		})
	}
}
