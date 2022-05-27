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
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"
	redis "google.golang.org/api/redis/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/errors"

	"github.com/crossplane-contrib/provider-gcp/apis/cache/v1beta1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
)

const (
	instanceNameFormat = "projects/%s/locations/%s/instances/%s"
	parentFormat       = "projects/%s/locations/%s"
)

// Valid states for a CloudMemorystore Instance.
const (
	StateUnspecified = "STATE_UNSPECIFIED"
	StateCreating    = "CREATING"
	StateReady       = "READY"
	StateUpdating    = "UPDATING"
	StateDeleting    = "DELETING"
	StateRepairing   = "REPAIRING"
	StateMaintenance = "MAINTENANCE"
	StateImporting   = "IMPORTING"
	StateFailingOver = "FAILING_OVER"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// GetFullyQualifiedParent builds the fully qualified name of the instance
// parent.
func GetFullyQualifiedParent(project string, p v1beta1.CloudMemorystoreInstanceParameters) string {
	return fmt.Sprintf(parentFormat, project, p.Region)
}

// GetFullyQualifiedName builds the fully qualified name of the instance.
func GetFullyQualifiedName(project string, p v1beta1.CloudMemorystoreInstanceParameters, name string) string {
	return fmt.Sprintf(instanceNameFormat, project, p.Region, name)
}

// GenerateRedisInstance is used to convert Crossplane CloudMemorystoreInstanceParameters
// to GCP's Redis Instance object. Name must be a fully qualified name for the instance.
func GenerateRedisInstance(name string, s v1beta1.CloudMemorystoreInstanceParameters, r *redis.Instance) {
	r.Name = name
	r.Tier = s.Tier
	r.MemorySizeGb = s.MemorySizeGB
	r.Labels = s.Labels
	r.RedisConfigs = s.RedisConfigs
	r.DisplayName = gcp.StringValue(s.DisplayName)
	r.LocationId = gcp.StringValue(s.LocationID)
	r.AlternativeLocationId = gcp.StringValue(s.AlternativeLocationID)
	r.RedisVersion = gcp.StringValue(s.RedisVersion)
	r.ReservedIpRange = gcp.StringValue(s.ReservedIPRange)
	r.AuthorizedNetwork = gcp.StringValue(s.AuthorizedNetwork)
	r.ConnectMode = gcp.StringValue(s.ConnectMode)
	r.AuthEnabled = gcp.BoolValue(s.AuthEnabled)
}

// GenerateObservation is used to produce an observation object from GCP's Redis
// Instance object.
func GenerateObservation(r redis.Instance) v1beta1.CloudMemorystoreInstanceObservation {
	o := v1beta1.CloudMemorystoreInstanceObservation{
		Name:                   r.Name,
		Host:                   r.Host,
		Port:                   r.Port,
		CurrentLocationID:      r.CurrentLocationId,
		State:                  r.State,
		StatusMessage:          r.StatusMessage,
		PersistenceIAMIdentity: r.PersistenceIamIdentity,
	}
	t, err := time.Parse(time.RFC3339, r.CreateTime)
	if err != nil {
		return o
	}
	m := metav1.NewTime(t)
	o.CreateTime = &m
	return o
}

// GenerateAuthStringObservation is used to produce an observation object from GCP's Redis
// Instance AuthString data.
func GenerateAuthStringObservation(r redis.InstanceAuthString) string {
	return r.AuthString
}

// LateInitializeSpec fills empty spec fields with the data retrieved from GCP.
func LateInitializeSpec(spec *v1beta1.CloudMemorystoreInstanceParameters, r redis.Instance) {
	if spec.Tier == "" {
		spec.Tier = r.Tier
	}
	if spec.MemorySizeGB == 0 {
		spec.MemorySizeGB = r.MemorySizeGb
	}
	spec.DisplayName = gcp.LateInitializeString(spec.DisplayName, r.DisplayName)
	spec.Labels = gcp.LateInitializeStringMap(spec.Labels, r.Labels)
	spec.LocationID = gcp.LateInitializeString(spec.LocationID, r.LocationId)
	spec.AlternativeLocationID = gcp.LateInitializeString(spec.AlternativeLocationID, r.AlternativeLocationId)
	spec.RedisVersion = gcp.LateInitializeString(spec.RedisVersion, r.RedisVersion)
	spec.ReservedIPRange = gcp.LateInitializeString(spec.ReservedIPRange, r.ReservedIpRange)
	spec.RedisConfigs = gcp.LateInitializeStringMap(spec.RedisConfigs, r.RedisConfigs)
	spec.AuthorizedNetwork = gcp.LateInitializeString(spec.AuthorizedNetwork, r.AuthorizedNetwork)
	spec.ConnectMode = gcp.LateInitializeString(spec.ConnectMode, r.ConnectMode)
	spec.AuthEnabled = gcp.LateInitializeBool(spec.AuthEnabled, r.AuthEnabled)
}

// IsUpToDate returns true if the supplied Kubernetes resource differs from the
// supplied GCP resource. It considers only fields that can be modified in
// place without deleting and recreating the instance.
func IsUpToDate(name string, in *v1beta1.CloudMemorystoreInstanceParameters, observed *redis.Instance) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*redis.Instance)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateRedisInstance(name, *in, desired)
	if desired.MemorySizeGb != observed.MemorySizeGb {
		return false, nil
	}
	if desired.DisplayName != observed.DisplayName {
		return false, nil
	}
	if !cmp.Equal(desired.RedisConfigs, observed.RedisConfigs) {
		return false, nil
	}
	if !cmp.Equal(desired.Labels, observed.Labels) {
		return false, nil
	}
	return true, nil
}
