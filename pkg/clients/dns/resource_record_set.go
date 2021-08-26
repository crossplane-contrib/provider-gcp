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
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	dns "google.golang.org/api/dns/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
)

const (
	errCheckUpToDate = "unable to determine if external resource is up to date"
	errUpdateManaged = "cannot update managed resource external name"
)

// LateInitializeSpec fills unassigned fields with the values in dns.ResourceRecordSet object.
func LateInitializeSpec(spec *v1alpha1.ResourceRecordSetParameters, external dns.ResourceRecordSet) {
	if len(spec.SignatureRRDatas) == 0 && len(external.SignatureRrdatas) > 0 {
		spec.SignatureRRDatas = external.SignatureRrdatas
	}
}

// GenerateResourceRecordSet generates *dns.ResourceRecordSet instance from ResourceRecordSetParameters.
func GenerateResourceRecordSet(name string, spec v1alpha1.ResourceRecordSetParameters, rrs *dns.ResourceRecordSet) {
	rrs.Kind = "dns#resourceRecordSet" // This is the only valid value for this field
	rrs.Name = name
	rrs.Rrdatas = spec.RRDatas
	rrs.Ttl = spec.TTL
	rrs.Type = spec.Type
	if spec.SignatureRRDatas != nil {
		rrs.SignatureRrdatas = spec.SignatureRRDatas
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

// CustomNameAsExternalName writes the name of the managed resource to
// the external name annotation field in order to be used as name of
// the external resource in provider.
// This external name will have a . appended at the end of the name
type CustomNameAsExternalName struct{ client client.Client }

// NewCustomNameAsExternalName returns a new CustomNameAsExternalName.
func NewCustomNameAsExternalName(c client.Client) *CustomNameAsExternalName {
	return &CustomNameAsExternalName{client: c}
}

// Initialize the given managed resource.
func (a *CustomNameAsExternalName) Initialize(ctx context.Context, mg resource.Managed) error {
	if meta.GetExternalName(mg) != "" {
		return nil
	}
	meta.SetExternalName(mg, mg.GetName()+".")
	return errors.Wrap(a.client.Update(ctx, mg), errUpdateManaged)
}
