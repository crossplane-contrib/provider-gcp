/*
Copyright 2022 The Crossplane Authors.

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

package managedzone

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	dns "google.golang.org/api/dns/v1"

	"github.com/crossplane/crossplane-runtime/pkg/errors"

	"github.com/crossplane-contrib/provider-gcp/apis/dns/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
)

const (
	errorCheckUpToDate = "unable to determine if external resource is up to date"
)

// GenerateManagedZone generates *dns.ManagedZone instance from ManagedZoneParameters
func GenerateManagedZone(name string, spec v1alpha1.ManagedZoneParameters, mz *dns.ManagedZone) {
	mz.Kind = "dns#managedZone" // This is the only valid value for this field
	mz.Name = name
	mz.DnsName = spec.DNSName
	mz.Labels = spec.Labels

	// Value in describe parameter is required
	if spec.Description != nil {
		mz.Description = *spec.Description
	} else {
		mz.Description = "Managed by Crossplane"
	}

	if spec.Visibility != nil {
		mz.Visibility = *spec.Visibility
	} else {
		mz.Visibility = "public"
	}

	if spec.PrivateVisibilityConfig != nil {
		mz.PrivateVisibilityConfig = &dns.ManagedZonePrivateVisibilityConfig{
			Networks: make([]*dns.ManagedZonePrivateVisibilityConfigNetwork, len(spec.PrivateVisibilityConfig.Networks)),
		}

		for i, v := range spec.PrivateVisibilityConfig.Networks {
			mz.PrivateVisibilityConfig.Networks[i] = &dns.ManagedZonePrivateVisibilityConfigNetwork{
				NetworkUrl: *v.NetworkURL,
			}
		}
	}

}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied ManagedZoneParameters that are set (i.e. non-zero) on the supplied
// ManagedZone.
func LateInitializeSpec(spec *v1alpha1.ManagedZoneParameters, observed dns.ManagedZone) {
	spec.Labels = gcp.LateInitializeStringMap(spec.Labels, observed.Labels)
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, spec *v1alpha1.ManagedZoneParameters, observed *dns.ManagedZone) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errorCheckUpToDate)
	}
	desired, ok := generated.(*dns.ManagedZone)
	if !ok {
		return true, errors.New(errorCheckUpToDate)
	}
	GenerateManagedZone(name, *spec, desired)
	return cmp.Equal(desired, observed, cmpopts.EquateEmpty()), nil
}

// GenerateManagedZoneObservation takes a dns.ManagedZone and returns
// *ManagedZoneObservation.
func GenerateManagedZoneObservation(observed *dns.ManagedZone) v1alpha1.ManagedZoneObservation {
	return v1alpha1.ManagedZoneObservation{
		CreationTime: observed.CreationTime,
		ID:           observed.Id,
		NameServers:  observed.NameServers,
	}
}
