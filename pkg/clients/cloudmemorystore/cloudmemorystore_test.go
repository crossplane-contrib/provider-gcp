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

	redis "google.golang.org/api/redis/v1"

	"github.com/crossplane-contrib/provider-gcp/apis/cache/v1beta1"
)

const (
	fullName = "projects/coolProject/locations/us-cool1/instances/claimnamespace-claimname-342sd"

	memorySizeGB = 1
)

var (
	authorizedNetwork = "default"
	redisVersion      = "REDIS_6_X"
	redisConfigs      = map[string]string{"cool": "socool"}
	tlsMode           = "SERVER_AUTHENTICATION"
)

func TestIsUpToDate(t *testing.T) {
	randString := "wat"
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := []struct {
		name string
		id   string
		kube *v1beta1.CloudMemorystoreInstance
		gcp  *redis.Instance
		want want
	}{
		{
			name: "NeedsLessMemory",
			id:   fullName,
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
						MemorySizeGB: memorySizeGB,
					},
				},
			},
			gcp: &redis.Instance{
				Name:         fullName,
				MemorySizeGb: memorySizeGB + 1,
			},
			want: want{upToDate: false, isErr: false},
		},
		{
			name: "NeedsNewRedisConfigs",
			id:   fullName,
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
					},
				},
			},
			gcp: &redis.Instance{
				Name:         fullName,
				RedisConfigs: map[string]string{"super": "cool"},
			},
			want: want{upToDate: false, isErr: false},
		},
		{
			name: "NeedsNoUpdate",
			id:   fullName,
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
						MemorySizeGB: memorySizeGB,
					},
				},
			},
			gcp: &redis.Instance{
				Name:         fullName,
				RedisConfigs: redisConfigs,
				MemorySizeGb: memorySizeGB,
			},
			want: want{upToDate: true, isErr: false},
		},
		{
			name: "NeedsNoUpdateWithOutputFields",
			id:   fullName,
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisConfigs: redisConfigs,
						MemorySizeGB: memorySizeGB,
					},
				},
			},
			gcp: &redis.Instance{
				Name:          fullName,
				RedisConfigs:  redisConfigs,
				MemorySizeGb:  memorySizeGB,
				StatusMessage: "definitely not in spec",
			},
			want: want{upToDate: true, isErr: false},
		},
		{
			name: "CannotUpdateField",
			id:   fullName,
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
			gcp: &redis.Instance{
				Name:              fullName,
				MemorySizeGb:      memorySizeGB,
				AuthorizedNetwork: authorizedNetwork,
			},
			want: want{upToDate: true, isErr: false},
		},
		{
			name: "TlsEnabled",
			id:   fullName,
			kube: &v1beta1.CloudMemorystoreInstance{
				Spec: v1beta1.CloudMemorystoreInstanceSpec{
					ForProvider: v1beta1.CloudMemorystoreInstanceParameters{
						RedisVersion:          &redisVersion,
						MemorySizeGB:          memorySizeGB,
						TransitEncryptionMode: &tlsMode,
					},
				},
			},
			gcp: &redis.Instance{
				Name:                  fullName,
				RedisVersion:          redisVersion,
				MemorySizeGb:          memorySizeGB,
				AuthorizedNetwork:     authorizedNetwork,
				TransitEncryptionMode: tlsMode,
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
