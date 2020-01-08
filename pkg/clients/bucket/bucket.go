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
	"google.golang.org/api/storage/v1"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
)

func GenerateBucket(spec v1alpha3.BucketParameters, name string) *storage.Bucket {
	b := &storage.Bucket{
		Location:        spec.Location,
		Name:            name,
		ForceSendFields: []string{"Location", "Name"},
	}
	b.StorageClass, b.ForceSendFields = GenerateString("StorageClass", spec.StorageClass, b.ForceSendFields)
	b.DefaultEventBasedHold, b.ForceSendFields = GenerateBool("DefaultEventBasedHold", spec.DefaultEventBasedHold, b.ForceSendFields)
	b.LocationType, b.ForceSendFields = GenerateString("LocationType", spec.LocationType, b.ForceSendFields)
	b.Labels, b.ForceSendFields = GenerateStringMap("Labels", spec.Labels, b.ForceSendFields)

	b.Billing, b.ForceSendFields = GenerateBilling(spec, b.ForceSendFields)
	b.Cors, b.ForceSendFields = GenerateCors(spec, b.ForceSendFields)
	b.Encryption, b.ForceSendFields = GenerateEncryption(spec, b.ForceSendFields)
	b.IamConfiguration, b.ForceSendFields = GenerateIamConfiguration(spec, b.ForceSendFields)
	b.Lifecycle, b.ForceSendFields = GenerateLifecycle(spec, b.ForceSendFields)
	b.Logging, b.ForceSendFields = GenerateBucketLogging(spec, b.ForceSendFields)
	b.RetentionPolicy, b.ForceSendFields = GenerateBucketRetentionPolicy(spec, b.ForceSendFields)
	b.Versioning, b.ForceSendFields = GenerateBucketVersioning(spec, b.ForceSendFields)
	b.Website, b.ForceSendFields = GenerateBucketWebsite(spec, b.ForceSendFields)
	return b
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
func GenerateString(fieldName string, specValue *string, f []string) (string, []string) {
	if specValue == nil {
		return "", f
	}
	return *specValue, append(f, fieldName)
}
func GenerateStringSlice(fieldName string, specValue []string, f []string) ([]string, []string) {
	if specValue == nil {
		return nil, f
	}
	return specValue, append(f, fieldName)
}
func GenerateInt64(fieldName string, specValue *int64, f []string) (int64, []string) {
	if specValue == nil {
		return 0, f
	}
	return *specValue, append(f, fieldName)
}

func GenerateBool(fieldName string, specValue *bool, f []string) (bool, []string) {
	if specValue == nil {
		return false, f
	}
	return *specValue, append(f, fieldName)
}

func GenerateStringMap(fieldName string, specValue map[string]string, f []string) (map[string]string, []string) {
	if specValue == nil {
		return nil, f
	}
	return specValue, append(f, fieldName)
}
func GenerateBilling(spec v1alpha3.BucketParameters, f []string) (*storage.BucketBilling, []string) {
	if spec.Billing == nil {
		return nil, f
	}
	return &storage.BucketBilling{
		RequesterPays:   spec.Billing.RequesterPays,
		ForceSendFields: []string{"RequesterPays"},
	}, f
}
func GenerateCors(spec v1alpha3.BucketParameters, f []string) ([]*storage.BucketCors, []string) {
	if spec.Cors == nil {
		return nil, f
	}
	cors := make([]*storage.BucketCors, len(spec.Cors))
	for i, val := range spec.Cors {
		cors[i] = &storage.BucketCors{
			MaxAgeSeconds:   val.MaxAgeSeconds,
			Method:          val.Method,
			Origin:          val.Origin,
			ResponseHeader:  val.ResponseHeader,
			ForceSendFields: []string{"MaxAgeSeconds", "Method", "Origin", "ResponseHeader"},
		}
	}
	return cors, append(f, "Cors")
}
func GenerateEncryption(spec v1alpha3.BucketParameters, f []string) (*storage.BucketEncryption, []string) {
	if spec.Encryption == nil {
		return nil, f
	}
	return &storage.BucketEncryption{
		DefaultKmsKeyName: spec.Encryption.DefaultKmsKeyName,
		ForceSendFields:   []string{"DefaultKmsKeyName"},
	}, f
}
func GenerateIamConfiguration(spec v1alpha3.BucketParameters, f []string) (*storage.BucketIamConfiguration, []string) {
	if spec.IamConfiguration == nil {
		return nil, f
	}
	iam := &storage.BucketIamConfiguration{}
	if spec.IamConfiguration.BucketPolicyOnly != nil {
		bpl := &storage.BucketIamConfigurationBucketPolicyOnly{
			Enabled:         spec.IamConfiguration.BucketPolicyOnly.Enabled,
			ForceSendFields: []string{"Enabled"},
		}
		bpl.LockedTime, bpl.ForceSendFields = GenerateString("LockedTime", spec.IamConfiguration.BucketPolicyOnly.LockedTime, bpl.ForceSendFields)
		iam.BucketPolicyOnly = bpl
	}
	if spec.IamConfiguration.UniformBucketLevelAccess != nil {
		ubla := &storage.BucketIamConfigurationUniformBucketLevelAccess{
			Enabled:         spec.IamConfiguration.UniformBucketLevelAccess.Enabled,
			ForceSendFields: []string{"Enabled"},
		}
		ubla.LockedTime, ubla.ForceSendFields = GenerateString("LockedTime", spec.IamConfiguration.BucketPolicyOnly.LockedTime, ubla.ForceSendFields)
		iam.UniformBucketLevelAccess = ubla
	}
	return iam, f
}
func GenerateLifecycle(spec v1alpha3.BucketParameters, f []string) (*storage.BucketLifecycle, []string) {
	if spec.Lifecycle == nil {
		return nil, f
	}
	l := &storage.BucketLifecycle{}
	l.Rule = make([]*storage.BucketLifecycleRule, len(spec.Lifecycle.Rule))
	for i, val := range spec.Lifecycle.Rule {
		r := &storage.BucketLifecycleRule{}
		if val.Action != nil {
			r.Action = &storage.BucketLifecycleRuleAction{
				Type:            val.Action.Type,
				ForceSendFields: []string{"Type"},
			}
			r.Action.StorageClass, r.Action.ForceSendFields = GenerateString("StorageClass", val.Action.StorageClass, r.Action.ForceSendFields)
		}
		if val.Condition != nil {
			cond := &storage.BucketLifecycleRuleCondition{}
			cond.Age, cond.ForceSendFields = GenerateInt64("Age", val.Condition.Age, cond.ForceSendFields)
			cond.CreatedBefore, cond.ForceSendFields = GenerateString("CreatedBefore", val.Condition.CreatedBefore, cond.ForceSendFields)
			cond.MatchesPattern, cond.ForceSendFields = GenerateString("MatchesPattern", val.Condition.MatchesPattern, cond.ForceSendFields)
			cond.NumNewerVersions, cond.ForceSendFields = GenerateInt64("NumNewerVersions", val.Condition.NumNewerVersions, cond.ForceSendFields)
			cond.MatchesStorageClass, cond.ForceSendFields = GenerateStringSlice("MatchesStorageClass", val.Condition.MatchesStorageClass, cond.ForceSendFields)
			var isLive bool
			isLive, cond.ForceSendFields = GenerateBool("IsLive", val.Condition.IsLive, cond.ForceSendFields)
			cond.IsLive = &isLive
			r.Condition = cond
		}
		l.Rule[i] = r
	}
	return l, f
}
func GenerateBucketLogging(spec v1alpha3.BucketParameters, f []string) (*storage.BucketLogging, []string) {
	if spec.Logging == nil {
		return nil, f
	}
	l := &storage.BucketLogging{}
	l.LogBucket, l.ForceSendFields = GenerateString("LogBucket", spec.Logging.LogBucket, l.ForceSendFields)
	l.LogObjectPrefix, l.ForceSendFields = GenerateString("LogObjectPrefix", spec.Logging.LogObjectPrefix, l.ForceSendFields)
	return l, f
}

func GenerateBucketRetentionPolicy(spec v1alpha3.BucketParameters, f []string) (*storage.BucketRetentionPolicy, []string) {
	if spec.RetentionPolicy == nil {
		return nil, f
	}
	r := &storage.BucketRetentionPolicy{}
	r.EffectiveTime, r.ForceSendFields = GenerateString("EffectiveTime", spec.RetentionPolicy.EffectiveTime, r.ForceSendFields)
	r.IsLocked, r.ForceSendFields = GenerateBool("IsLocked", spec.RetentionPolicy.IsLocked, r.ForceSendFields)
	r.RetentionPeriod, r.ForceSendFields = GenerateInt64("RetentionPeriod", spec.RetentionPolicy.RetentionPeriod, r.ForceSendFields)
	return r, f
}
func GenerateBucketVersioning(spec v1alpha3.BucketParameters, f []string) (*storage.BucketVersioning, []string) {
	if spec.Versioning == nil {
		return nil, f
	}
	return &storage.BucketVersioning{
		Enabled:         spec.Versioning.Enabled,
		ForceSendFields: []string{"Enabled"},
	}, f
}
func GenerateBucketWebsite(spec v1alpha3.BucketParameters, f []string) (*storage.BucketWebsite, []string) {
	if spec.Website == nil {
		return nil, f
	}
	w := &storage.BucketWebsite{}
	w.MainPageSuffix, w.ForceSendFields = GenerateString("MainPageSuffix", spec.Website.MainPageSuffix, w.ForceSendFields)
	w.NotFoundPage, w.ForceSendFields = GenerateString("NotFoundPage", spec.Website.NotFoundPage, w.ForceSendFields)
	return w, f
}
