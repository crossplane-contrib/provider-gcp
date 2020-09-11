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

package dns

import (
	"fmt"

	"google.golang.org/api/dns/v1"

	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
)

// GenerateManagedZone produces a ManagedZone that is configured via given ManagedZoneParameters.
func GenerateManagedZone(name string, projectID string, s v1alpha1.ManagedZoneParameters) *dns.ManagedZone {
	network := &dns.ManagedZonePrivateVisibilityConfigNetwork{
		Kind:       "dns#managedZonePrivateVisibilityConfigNetwork",
		NetworkUrl: fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", projectID, s.Network),
	}

	visibility := "private"
	if s.Visibility != "" {
		visibility = s.Visibility
	}

	zone := &dns.ManagedZone{
		Name:        name,
		DnsName:     s.DNSName,
		Description: s.Description,
		Visibility:  visibility,
		PrivateVisibilityConfig: &dns.ManagedZonePrivateVisibilityConfig{
			Kind:     "dns#managedZonePrivateVisibilityConfig",
			Networks: []*dns.ManagedZonePrivateVisibilityConfigNetwork{network},
		},
	}

	if s.Labels != nil && len(s.Labels) != 0 {
		zone.Labels = s.Labels
	}

	return zone
}

// LateInitialize fills the empty fields in *v1alpha1.ManagedZoneParameters with
// the values seen in ManagedZone.
func LateInitialize(spec *v1alpha1.ManagedZoneParameters, obs *dns.ManagedZone) {
	if obs == nil {
		return
	}

	if len(spec.Labels) == 0 && len(obs.Labels) != 0 {
		for k, v := range obs.Labels {
			spec.Labels[k] = v
		}
	}
}

// IsUpToDate check whether the dns in Spec and Response are same or not
func IsUpToDate(spec v1alpha1.ManagedZoneParameters, obs dns.ManagedZone) bool {
	return spec.DNSName == obs.DnsName
}
