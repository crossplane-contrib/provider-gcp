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

package keyring

import (
	"google.golang.org/api/cloudkms/v1"

	"github.com/crossplane/provider-gcp/apis/classic/kms/v1alpha1"
)

// Client should be satisfied to conduct SA operations.
type Client interface {
	Create(parent string, keyring *cloudkms.KeyRing) *cloudkms.ProjectsLocationsKeyRingsCreateCall
	Get(name string) *cloudkms.ProjectsLocationsKeyRingsGetCall
}

// GenerateObservation produces KeyRingObservation object from cloudkms.KeyRing object.
func GenerateObservation(in cloudkms.KeyRing) v1alpha1.KeyRingObservation { // nolint:gocyclo
	return v1alpha1.KeyRingObservation{
		Name:       in.Name,
		CreateTime: in.CreateTime,
	}
}
