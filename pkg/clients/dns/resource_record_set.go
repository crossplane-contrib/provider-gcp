/*
Copyright 2021 The Crossplane Authors.

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

package dns

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	dns "google.golang.org/api/dns/v1"

	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// LateInitializeSpec fills unassigned fields with the values in dns.ResourceRecordSet object.
func LateInitializeSpec(spec *v1alpha1.ResourceRecordSetParameters, external dns.ResourceRecordSet) { // nolint:gocyclo
	if spec.SignatureRRDatas == nil && len(external.SignatureRrdatas) > 0 {
		spec.SignatureRRDatas = &external.SignatureRrdatas
	}
}

// GenerateObservation produces ResourceRecordSetObservation object from dns.ResourceRecordSet object.
func GenerateObservation(external dns.ResourceRecordSet) v1alpha1.ResourceRecordSetObservation {
	return v1alpha1.ResourceRecordSetObservation{
		Name: external.Name,
		Type: external.Type,
	}
}

// GenerateResourceRecordSet generates *dns.ResourceRecordSet instance from ResourceRecordSetParameters.
func GenerateResourceRecordSet(name string, spec v1alpha1.ResourceRecordSetParameters, rrs *dns.ResourceRecordSet) {
	rrs.Name = name
	rrs.Kind = spec.Kind
	rrs.Rrdatas = spec.RRDatas
	rrs.Ttl = spec.TTL
	rrs.Type = spec.Type
	if spec.SignatureRRDatas != nil {
		rrs.SignatureRrdatas = *spec.SignatureRRDatas
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, spec *v1alpha1.ResourceRecordSetParameters, observed *dns.ResourceRecordSet) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*dns.ResourceRecordSet)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateResourceRecordSet(name, *spec, desired)
	return cmp.Equal(desired, observed, cmpopts.EquateEmpty()), nil
}
