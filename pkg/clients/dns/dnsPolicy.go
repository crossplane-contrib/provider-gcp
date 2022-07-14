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
	dns "google.golang.org/api/dns/v1"

	"github.com/crossplane-contrib/provider-gcp/apis/dns/v1alpha1"
)

// const (
// 	err_CheckUpToDate = "unable to determine if external resource is up to date"
// 	err_UpdateManaged = "cannot update managed resource external name"
// )

// GenerateDNSPolicy generates *dns.Policy instance from PolicyParameters
func GenerateDNSPolicy(name string, spec v1alpha1.PolicyParameters, policy *dns.Policy) {
	policy.Kind = "dns#policy"

	policy.Name = spec.Name

	if spec.EnableInboundForwarding != nil {
		policy.EnableInboundForwarding = *spec.EnableInboundForwarding
	}
	if spec.EnableLogging != nil {
		policy.EnableLogging = *spec.EnableLogging
	}
	if spec.Description != nil {
		policy.Description = *spec.Description
	}
}
