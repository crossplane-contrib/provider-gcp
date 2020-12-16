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

package cryptokey

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	"google.golang.org/api/cloudkms/v1"

	"github.com/crossplane/provider-gcp/apis/kms/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// Client should be satisfied to conduct SA operations.
type Client interface {
	Create(parent string, cryptokey *cloudkms.CryptoKey) *cloudkms.ProjectsLocationsKeyRingsCryptoKeysCreateCall
	Get(name string) *cloudkms.ProjectsLocationsKeyRingsCryptoKeysGetCall
	Patch(name string, cryptokey *cloudkms.CryptoKey) *cloudkms.ProjectsLocationsKeyRingsCryptoKeysPatchCall
}

// GenerateCryptoKeyInstance generates *kmsv1.CryptoKey instance from CryptoKeyParameters.
func GenerateCryptoKeyInstance(name string, in v1alpha1.CryptoKeyParameters, ck *cloudkms.CryptoKey) {
	ck.Labels = in.Labels
	ck.Purpose = in.Purpose
	ck.RotationPeriod = gcp.StringValue(in.RotationPeriod)
	ck.NextRotationTime = gcp.StringValue(in.NextRotationTime)
	if in.VersionTemplate != nil {
		if ck.VersionTemplate == nil {
			ck.VersionTemplate = &cloudkms.CryptoKeyVersionTemplate{}
		}
		ck.VersionTemplate.Algorithm = gcp.StringValue(in.VersionTemplate.Algorithm)
		ck.VersionTemplate.ProtectionLevel = gcp.StringValue(in.VersionTemplate.ProtectionLevel)
	}
}

// GenerateObservation produces CryptoKeyObservation object from cloudkms.CryptoKey object.
func GenerateObservation(in cloudkms.CryptoKey) v1alpha1.CryptoKeyObservation { // nolint:gocyclo
	o := v1alpha1.CryptoKeyObservation{
		CreateTime:       in.CreateTime,
		Name:             in.Name,
		NextRotationTime: in.NextRotationTime,
	}

	if in.Primary != nil {
		o.Primary = &v1alpha1.CryptoKeyVersion{
			Algorithm:           in.Primary.Algorithm,
			CreateTime:          in.Primary.CreateTime,
			DestroyEventTime:    in.Primary.DestroyEventTime,
			DestroyTime:         in.Primary.DestroyTime,
			GenerateTime:        in.Primary.GenerateTime,
			ImportFailureReason: in.Primary.ImportFailureReason,
			ImportJob:           in.Primary.ImportJob,
			ImportTime:          in.Primary.ImportTime,
			Name:                in.Primary.Name,
			ProtectionLevel:     in.Primary.ProtectionLevel,
			State:               in.Primary.State,
		}
		if in.Primary.Attestation != nil {
			o.Primary.Attestation = &v1alpha1.KeyOperationAttestation{
				Content: in.Primary.Attestation.Content,
				Format:  in.Primary.Attestation.Format,
			}
		}
		if in.Primary.ExternalProtectionLevelOptions != nil {
			o.Primary.ExternalProtectionLevelOptions = &v1alpha1.ExternalProtectionLevelOptions{
				ExternalKeyUri: in.Primary.ExternalProtectionLevelOptions.ExternalKeyUri,
			}
		}
	}

	return o
}

// LateInitializeSpec fills unassigned fields with the values in cloudkms.CryptoKey object.
func LateInitializeSpec(spec *v1alpha1.CryptoKeyParameters, in cloudkms.CryptoKey) {
	spec.Labels = in.Labels
	spec.RotationPeriod = gcp.LateInitializeString(spec.RotationPeriod, in.RotationPeriod)
	spec.NextRotationTime = gcp.LateInitializeString(spec.NextRotationTime, in.NextRotationTime)
	if in.VersionTemplate != nil {
		if spec.VersionTemplate == nil {
			spec.VersionTemplate = &v1alpha1.CryptoKeyVersionTemplate{}
		}
		spec.VersionTemplate.ProtectionLevel = gcp.LateInitializeString(
			spec.VersionTemplate.ProtectionLevel, in.VersionTemplate.ProtectionLevel)
		spec.VersionTemplate.Algorithm = gcp.LateInitializeString(
			spec.VersionTemplate.Algorithm, in.VersionTemplate.Algorithm)
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, in *v1alpha1.CryptoKeyParameters, observed *cloudkms.CryptoKey) (bool, string, error) { // nolint:gocyclo
	um := make([]string, 0, 6)
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, "", errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*cloudkms.CryptoKey)
	if !ok {
		return true, "", errors.New(errCheckUpToDate)
	}
	GenerateCryptoKeyInstance(name, *in, desired)

	if !cmp.Equal(desired.Labels, observed.Labels, cmpopts.EquateEmpty()) {
		um = append(um, "labels")
	}
	if !cmp.Equal(desired.Purpose, observed.Purpose, cmpopts.EquateEmpty()) {
		um = append(um, "purpose")
	}
	if !cmp.Equal(desired.RotationPeriod, observed.RotationPeriod, cmpopts.EquateEmpty()) {
		um = append(um, "rotationPeriod")
	}
	if !cmp.Equal(desired.NextRotationTime, observed.NextRotationTime, cmpopts.EquateEmpty()) {
		um = append(um, "nextRotationTime")
	}

	if !cmp.Equal(desired.VersionTemplate, observed.VersionTemplate, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(cloudkms.CryptoKeyVersionTemplate{}, "ForceSendFields"),
		cmpopts.IgnoreFields(cloudkms.CryptoKeyVersionTemplate{}, "NullFields"),
	) {
		if !cmp.Equal(desired.VersionTemplate.Algorithm, observed.VersionTemplate.Algorithm, cmpopts.EquateEmpty()) {
			um = append(um, "versionTemplate.algorithm")
		}
		if !cmp.Equal(desired.VersionTemplate.ProtectionLevel, observed.VersionTemplate.ProtectionLevel, cmpopts.EquateEmpty()) {
			um = append(um, "versionTemplate.protectionLevel")
		}

	}

	if len(um) > 0 {
		return false, strings.Join(um, ","), nil
	}
	return true, "", nil
}
