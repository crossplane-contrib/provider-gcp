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
		Location: spec.Location,
		Name:     name,
	}
	b.StorageClass, b.ForceSendFields, b.NullFields = GenerateString("StorageClass", spec.StorageClass, b.ForceSendFields, b.NullFields)
	b.DefaultEventBasedHold, b.ForceSendFields, b.NullFields = GenerateBool("DefaultEventBasedHold", spec.DefaultEventBasedHold, b.ForceSendFields, b.NullFields)
	b.LocationType, b.ForceSendFields, b.NullFields = GenerateString("LocationType", spec.LocationType, b.ForceSendFields, b.NullFields)
	b.Labels, b.ForceSendFields, b.NullFields = GenerateStringMap("Labels", spec.Labels, b.ForceSendFields, b.NullFields)

	b.Billing, b.ForceSendFields, b.NullFields = GenerateBilling(spec, b.ForceSendFields, b.NullFields)
	b.Cors, b.ForceSendFields, b.NullFields = GenerateCors(spec, b.ForceSendFields, b.NullFields)
	b.Encryption, b.ForceSendFields, b.NullFields = GenerateEncryption(spec, b.ForceSendFields, b.NullFields)
	b.IamConfiguration, b.ForceSendFields, b.NullFields = GenerateIamConfiguration(spec, b.ForceSendFields, b.NullFields)
	b.Lifecycle, b.ForceSendFields, b.NullFields = GenerateLifecycle(spec, b.ForceSendFields, b.NullFields)
	b.Logging, b.ForceSendFields, b.NullFields = GenerateBucketLogging(spec, b.ForceSendFields, b.NullFields)
	b.RetentionPolicy, b.ForceSendFields, b.NullFields = GenerateBucketRetentionPolicy(spec, b.ForceSendFields, b.NullFields)
	b.Versioning, b.ForceSendFields, b.NullFields = GenerateBucketVersioning(spec, b.ForceSendFields, b.NullFields)
	b.Website, b.ForceSendFields, b.NullFields = GenerateBucketWebsite(spec, b.ForceSendFields, b.NullFields)

	return b
}
func GenerateString(fieldName string, specValue *string, f []string, n []string) (string, []string, []string) {
	if specValue == nil {
		return "", f, append(n, fieldName)
	}
	return *specValue, append(f, fieldName), n
}
func GenerateStringSlice(fieldName string, specValue []string, f []string, n []string) ([]string, []string, []string) {
	if specValue == nil {
		return nil, f, append(n, fieldName)
	}
	return specValue, append(f, fieldName), n
}
func GenerateInt64(fieldName string, specValue *int64, f []string, n []string) (int64, []string, []string) {
	if specValue == nil {
		return 0, f, append(n, fieldName)
	}
	return *specValue, append(f, fieldName), n
}

func GenerateBool(fieldName string, specValue *bool, f []string, n []string) (bool, []string, []string) {
	if specValue == nil {
		return false, f, append(n, fieldName)
	}
	return *specValue, append(f, fieldName), n
}

func GenerateStringMap(fieldName string, specValue map[string]string, f []string, n []string) (map[string]string, []string, []string) {
	if specValue == nil {
		return nil, f, append(n, fieldName)
	}
	return specValue, append(f, fieldName), n
}
func GenerateBilling(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketBilling, []string, []string) {
	if spec.Billing == nil {
		return nil, f, append(n, "Billing")
	}
	return &storage.BucketBilling{
		RequesterPays:   spec.Billing.RequesterPays,
		ForceSendFields: []string{"RequesterPays"},
	}, f, n
}
func GenerateCors(spec v1alpha3.BucketParameters, f []string, n []string) ([]*storage.BucketCors, []string, []string) {
	if spec.Cors == nil {
		return nil, f, append(n, "Cors")
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
	return cors, append(f, "Cors"), n
}
func GenerateEncryption(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketEncryption, []string, []string) {
	if spec.Encryption == nil {
		return nil, f, append(n, "Encryption")
	}
	return &storage.BucketEncryption{
		DefaultKmsKeyName: spec.Encryption.DefaultKmsKeyName,
		ForceSendFields:   []string{"DefaultKmsKeyName"},
	}, f, n
}
func GenerateIamConfiguration(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketIamConfiguration, []string, []string) {
	if spec.IamConfiguration == nil {
		return nil, f, append(n, "IamConfiguration")
	}
	iam := &storage.BucketIamConfiguration{}
	if spec.IamConfiguration.BucketPolicyOnly != nil {
		bpl := &storage.BucketIamConfigurationBucketPolicyOnly{
			Enabled:         spec.IamConfiguration.BucketPolicyOnly.Enabled,
			ForceSendFields: []string{"Enabled"},
		}
		bpl.LockedTime, bpl.ForceSendFields, bpl.NullFields = GenerateString("LockedTime", spec.IamConfiguration.BucketPolicyOnly.LockedTime, bpl.ForceSendFields, bpl.NullFields)
		iam.BucketPolicyOnly = bpl
	} else {
		iam.NullFields = append(iam.NullFields, "BucketPolicyOnly")
	}
	if spec.IamConfiguration.UniformBucketLevelAccess != nil {
		ubla := &storage.BucketIamConfigurationUniformBucketLevelAccess{
			Enabled:         spec.IamConfiguration.UniformBucketLevelAccess.Enabled,
			ForceSendFields: []string{"Enabled"},
		}
		ubla.LockedTime, ubla.ForceSendFields, ubla.NullFields = GenerateString("LockedTime", spec.IamConfiguration.BucketPolicyOnly.LockedTime, ubla.ForceSendFields, ubla.NullFields)
		iam.UniformBucketLevelAccess = ubla
	} else {
		iam.NullFields = append(iam.NullFields, "UniformBucketLevelAccess")
	}
	return iam, f, n
}
func GenerateLifecycle(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketLifecycle, []string, []string) {
	if spec.Lifecycle == nil {
		return nil, f, append(n, "Lifecycle")
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
			r.Action.StorageClass, r.Action.ForceSendFields, r.Action.NullFields = GenerateString("StorageClass", val.Action.StorageClass, r.Action.ForceSendFields, r.Action.NullFields)
		}
		if val.Condition != nil {
			cond := &storage.BucketLifecycleRuleCondition{}
			cond.Age, cond.ForceSendFields, cond.NullFields = GenerateInt64("Age", val.Condition.Age, cond.ForceSendFields, cond.NullFields)
			cond.CreatedBefore, cond.ForceSendFields, cond.NullFields = GenerateString("CreatedBefore", val.Condition.CreatedBefore, cond.ForceSendFields, cond.NullFields)
			cond.MatchesPattern, cond.ForceSendFields, cond.NullFields = GenerateString("MatchesPattern", val.Condition.MatchesPattern, cond.ForceSendFields, cond.NullFields)
			cond.NumNewerVersions, cond.ForceSendFields, cond.NullFields = GenerateInt64("NumNewerVersions", val.Condition.NumNewerVersions, cond.ForceSendFields, cond.NullFields)
			cond.MatchesStorageClass, cond.ForceSendFields, cond.NullFields = GenerateStringSlice("MatchesStorageClass", val.Condition.MatchesStorageClass, cond.ForceSendFields, cond.NullFields)
			var isLive bool
			isLive, cond.ForceSendFields, cond.NullFields = GenerateBool("IsLive", val.Condition.IsLive, cond.ForceSendFields, cond.NullFields)
			cond.IsLive = &isLive
			r.Condition = cond
		}
		l.Rule[i] = r
	}
	return l, f, n
}
func GenerateBucketLogging(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketLogging, []string, []string) {
	if spec.Logging == nil {
		return nil, f, append(n, "Logging")
	}
	l := &storage.BucketLogging{}
	l.LogBucket, l.ForceSendFields, l.NullFields = GenerateString("LogBucket", spec.Logging.LogBucket, l.ForceSendFields, l.NullFields)
	l.LogObjectPrefix, l.ForceSendFields, l.NullFields = GenerateString("LogObjectPrefix", spec.Logging.LogObjectPrefix, l.ForceSendFields, l.NullFields)
	return l, f, n
}

func GenerateBucketRetentionPolicy(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketRetentionPolicy, []string, []string) {
	if spec.RetentionPolicy == nil {
		return nil, f, append(n, "RetentionPolicy")
	}
	r := &storage.BucketRetentionPolicy{}
	r.EffectiveTime, r.ForceSendFields, r.NullFields = GenerateString("EffectiveTime", spec.RetentionPolicy.EffectiveTime, r.ForceSendFields, r.NullFields)
	r.IsLocked, r.ForceSendFields, r.NullFields = GenerateBool("IsLocked", spec.RetentionPolicy.IsLocked, r.ForceSendFields, r.NullFields)
	r.RetentionPeriod, r.ForceSendFields, r.NullFields = GenerateInt64("RetentionPeriod", spec.RetentionPolicy.RetentionPeriod, r.ForceSendFields, r.NullFields)
	return r, f, n
}
func GenerateBucketVersioning(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketVersioning, []string, []string) {
	if spec.Versioning == nil {
		return nil, f, append(n, "Versioning")
	}
	return &storage.BucketVersioning{
		Enabled:         spec.Versioning.Enabled,
		ForceSendFields: []string{"Enabled"},
	}, f, n
}
func GenerateBucketWebsite(spec v1alpha3.BucketParameters, f []string, n []string) (*storage.BucketWebsite, []string, []string) {
	if spec.Website == nil {
		return nil, f, append(n, "Website")
	}
	w := &storage.BucketWebsite{}
	w.MainPageSuffix, w.ForceSendFields, w.NullFields = GenerateString("MainPageSuffix", spec.Website.MainPageSuffix, w.ForceSendFields, w.NullFields)
	w.NotFoundPage, w.ForceSendFields, w.NullFields = GenerateString("NotFoundPage", spec.Website.NotFoundPage, w.ForceSendFields, w.NullFields)
	return w, f, n
}
