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
	"context"
	"fmt"

	redisv1 "cloud.google.com/go/redis/apiv1"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	redisv1pb "google.golang.org/genproto/googleapis/cloud/redis/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-gcp/apis/cache/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// Cloud Memorystore instance states. Only the subset that is used
// is listed.
var (
	StateCreating = redisv1pb.Instance_CREATING.String()
	StateReady    = redisv1pb.Instance_READY.String()
	StateDeleting = redisv1pb.Instance_DELETING.String()
)

// A Client handles CRUD operations for Cloud Memorystore instances. This
// interface is compatible with the upstream CloudRedisClient.
type Client interface {
	CreateInstance(ctx context.Context, req *redisv1pb.CreateInstanceRequest, opts ...gax.CallOption) (*redisv1.CreateInstanceOperation, error)
	UpdateInstance(ctx context.Context, req *redisv1pb.UpdateInstanceRequest, opts ...gax.CallOption) (*redisv1.UpdateInstanceOperation, error)
	DeleteInstance(ctx context.Context, req *redisv1pb.DeleteInstanceRequest, opts ...gax.CallOption) (*redisv1.DeleteInstanceOperation, error)
	GetInstance(ctx context.Context, req *redisv1pb.GetInstanceRequest, opts ...gax.CallOption) (*redisv1pb.Instance, error)
}

// An InstanceID represents a CloudMemorystore instance in the GCP API.
type InstanceID struct {
	// Project in which this instance exists.
	Project string

	// Region in which this instance exists. The API calls this a 'location',
	// which is an overloaded term considering instances also have a 'location
	// id' (and 'alternative location id'), which represent zones.
	Region string

	// Instance name, or ID. The GCP API appears to call the unqualified name
	// (e.g. 'coolinstance') an ID, and the qualified name (e.g.
	// 'projects/coolproject/locations/us-west2/instances/coolinstance') a name.
	Instance string
}

// NewInstanceID returns an identifier used to represent CloudMemorystore
// instances in the GCP API. Instances may have names of up to 40 characters.
// https://godoc.org/google.golang.org/genproto/googleapis/cloud/redis/v1#CreateInstanceRequest
func NewInstanceID(project string, i *v1beta1.CloudMemorystoreInstance) InstanceID {
	return InstanceID{
		Project:  project,
		Region:   i.Spec.ForProvider.Region,
		Instance: meta.GetExternalName(i),
	}
}

// Parent returns the instance's parent, suitable for the create API call.
func (id InstanceID) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s", id.Project, id.Region)
}

// Name returns the instance's name, suitable for get and delete API calls.
func (id InstanceID) Name() string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", id.Project, id.Region, id.Instance)
}

// GenerateRedisInstance is used to convert Crossplane CloudMemorystoreInstanceParameters
// to GCP's Redis Instance object.
func GenerateRedisInstance(id InstanceID, s v1beta1.CloudMemorystoreInstanceParameters, r *redisv1pb.Instance) {
	r.Name = id.Name()
	r.Tier = redisv1pb.Instance_Tier(redisv1pb.Instance_Tier_value[s.Tier])
	r.MemorySizeGb = s.MemorySizeGB
	r.Labels = s.Labels
	r.RedisConfigs = s.RedisConfigs
	r.DisplayName = gcp.StringValue(s.DisplayName)
	r.LocationId = gcp.StringValue(s.LocationID)
	r.AlternativeLocationId = gcp.StringValue(s.AlternativeLocationID)
	r.RedisVersion = gcp.StringValue(s.RedisVersion)
	r.ReservedIpRange = gcp.StringValue(s.ReservedIPRange)
	r.AuthorizedNetwork = gcp.StringValue(s.AuthorizedNetwork)
}

// GenerateObservation is used to produce an observation object from GCP's Redis
// Instance object.
func GenerateObservation(r *redisv1pb.Instance) v1beta1.CloudMemorystoreInstanceObservation {
	o := v1beta1.CloudMemorystoreInstanceObservation{
		Name:                   r.Name,
		Host:                   r.Host,
		Port:                   r.Port,
		CurrentLocationID:      r.CurrentLocationId,
		State:                  r.State.String(),
		StatusMessage:          r.StatusMessage,
		PersistenceIAMIdentity: r.PersistenceIamIdentity,
	}
	if r.CreateTime != nil {
		t := metav1.Unix(r.CreateTime.Seconds, int64(r.CreateTime.Nanos))
		o.CreateTime = &t
	}
	return o
}

// LateInitializeSpec fills empty spec fields with the data retrieved from GCP.
func LateInitializeSpec(spec *v1beta1.CloudMemorystoreInstanceParameters, r *redisv1pb.Instance) {
	if spec.Tier == "" {
		spec.Tier = r.Tier.String()
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
}

// NewCreateInstanceRequest creates a request to create an instance suitable for
// use with the GCP API.
func NewCreateInstanceRequest(id InstanceID, i *v1beta1.CloudMemorystoreInstance) *redisv1pb.CreateInstanceRequest {
	r := &redisv1pb.Instance{}
	GenerateRedisInstance(id, i.Spec.ForProvider, r)
	return &redisv1pb.CreateInstanceRequest{
		Parent:     id.Parent(),
		InstanceId: id.Instance,
		Instance:   r,
	}
}

// NewUpdateInstanceRequest creates a request to update an instance suitable for
// use with the GCP API.
func NewUpdateInstanceRequest(id InstanceID, i *v1beta1.CloudMemorystoreInstance) *redisv1pb.UpdateInstanceRequest {
	r := &redisv1pb.Instance{}
	GenerateRedisInstance(id, i.Spec.ForProvider, r)
	return &redisv1pb.UpdateInstanceRequest{
		// These are the only fields we're concerned with that can be updated.
		// The documentation is incorrect regarding field masks - they must be
		// specified as snake case rather than camel case.
		// https://godoc.org/google.golang.org/genproto/googleapis/cloud/redis/v1#UpdateInstanceRequest
		UpdateMask: &field_mask.FieldMask{Paths: []string{"memory_size_gb", "redis_configs", "labels", "display_name"}},
		Instance:   r,
	}
}

// IsUpToDate returns true if the supplied Kubernetes resource differs from the
// supplied GCP resource. It considers only fields that can be modified in
// place without deleting and recreating the instance.
func IsUpToDate(id InstanceID, in *v1beta1.CloudMemorystoreInstanceParameters, observed *redisv1pb.Instance) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*redisv1pb.Instance)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateRedisInstance(id, *in, desired)
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

// NewDeleteInstanceRequest creates a request to delete an instance suitable for
// use with the GCP API.
func NewDeleteInstanceRequest(id InstanceID) *redisv1pb.DeleteInstanceRequest {
	return &redisv1pb.DeleteInstanceRequest{Name: id.Name()}
}

// NewGetInstanceRequest creates a request to get an instance from the GCP API.
func NewGetInstanceRequest(id InstanceID) *redisv1pb.GetInstanceRequest {
	return &redisv1pb.GetInstanceRequest{Name: id.Name()}
}

// IsNotFound returns true if the supplied error indicates a CloudMemorystore
// instance was not found.
func IsNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}
