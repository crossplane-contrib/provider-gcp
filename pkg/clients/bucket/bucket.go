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

package bucket

import (
	jsonpatch "github.com/muvaf/json-patch"
	"google.golang.org/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
)

func GenerateBucket(spec v1alpha3.BucketParameters, name string) (*storage.Bucket, error) {
	desired, err := json.Marshal(&spec)
	if err != nil {
		return nil, err
	}
	b := &storage.Bucket{}
	if err := json.Unmarshal(desired, b); err != nil {
		return nil, err
	}
	return b, err
}

func GenerateObservation(b storage.Bucket) v1alpha3.BucketObservation {
	o := v1alpha3.BucketObservation{
		SelfLink:       b.SelfLink,
		TimeCreated:    b.TimeCreated,
		Updated:        b.Updated,
		ProjectNumber:  int64(b.ProjectNumber),
		Metageneration: b.Metageneration,
	}
	if b.Owner != nil {
		o.Owner = &v1alpha3.BucketOwner{
			Entity:   b.Owner.Entity,
			EntityId: b.Owner.EntityId,
		}
	}
	return o
}
func LateInitialize(spec *v1alpha3.BucketParameters, b storage.Bucket) error {
	desired, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	actual, err := json.Marshal(&b)
	if err != nil {
		return err
	}
	patch, err := jsonpatch.CreateLateInitPatch(desired, actual)
	if err != nil {
		return err
	}
	final, err := jsonpatch.MergePatch(desired, patch)
	if err := json.Unmarshal(final, spec); err != nil {
		return err
	}
	return nil
}
