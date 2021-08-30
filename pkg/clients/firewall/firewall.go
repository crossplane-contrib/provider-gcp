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

package firewall

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"

	"github.com/crossplane/provider-gcp/apis/compute/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// GenerateFirewall takes a *FirewallParameters and returns *compute.Firewall.
// It assigns only the fields that are writable, i.e. not labelled as [Output Only]
// in Google's reference.
func GenerateFirewall(name string, in v1alpha1.FirewallParameters, firewall *compute.Firewall) {
	firewall.Name = name
	firewall.Description = gcp.StringValue(in.Description)
	firewall.Network = gcp.StringValue(in.Network)
	firewall.Priority = gcp.Int64Value(in.Priority)
	firewall.SourceRanges = in.SourceRanges
	firewall.DestinationRanges = in.DestinationRanges
	firewall.SourceTags = in.SourceTags
	firewall.TargetTags = in.TargetTags
	firewall.SourceServiceAccounts = in.SourceServiceAccounts
	firewall.TargetServiceAccounts = in.TargetServiceAccounts
	firewall.Direction = gcp.StringValue(in.Direction)
	firewall.Disabled = gcp.BoolValue(in.Disabled)
	if in.Allowed != nil {
		firewall.Allowed = make([]*compute.FirewallAllowed, len(in.Allowed))
		for idx, rule := range in.Allowed {
			firewall.Allowed[idx] = &compute.FirewallAllowed{
				IPProtocol: rule.IPProtocol,
				Ports:      rule.Ports,
			}
		}
	}

	if in.Denied != nil {
		firewall.Denied = make([]*compute.FirewallDenied, len(in.Denied))
		for idx, rule := range in.Denied {
			firewall.Denied[idx] = &compute.FirewallDenied{
				IPProtocol: rule.IPProtocol,
				Ports:      rule.Ports,
			}
		}
	}

	if in.LogConfig != nil {
		firewall.LogConfig = &compute.FirewallLogConfig{
			Enable: in.LogConfig.Enable,
		}
	}
}

// GenerateFirewallObservation takes a compute.Firewall and returns *FirewallObservation.
func GenerateFirewallObservation(in compute.Firewall) v1alpha1.FirewallObservation {
	fw := v1alpha1.FirewallObservation{
		CreationTimestamp: in.CreationTimestamp,
		ID:                in.Id,
		SelfLink:          in.SelfLink,
	}
	return fw
}

// LateInitializeSpec fills unassigned fields with the values in compute.Firewall object.
func LateInitializeSpec(spec *v1alpha1.FirewallParameters, in compute.Firewall) {
	spec.Disabled = gcp.LateInitializeBool(spec.Disabled, in.Disabled)
	spec.Network = gcp.LateInitializeString(spec.Network, in.Network)
	spec.Priority = gcp.LateInitializeInt64(spec.Priority, in.Priority)
	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)
	spec.Direction = gcp.LateInitializeString(spec.Direction, in.Direction)
	spec.SourceRanges = gcp.LateInitializeStringSlice(spec.SourceRanges, in.SourceRanges)
	spec.DestinationRanges = gcp.LateInitializeStringSlice(spec.DestinationRanges, in.DestinationRanges)
	spec.SourceServiceAccounts = gcp.LateInitializeStringSlice(spec.SourceServiceAccounts, in.SourceServiceAccounts)
	spec.TargetServiceAccounts = gcp.LateInitializeStringSlice(spec.TargetServiceAccounts, in.TargetServiceAccounts)
	spec.SourceTags = gcp.LateInitializeStringSlice(spec.SourceTags, in.SourceTags)
	spec.TargetTags = gcp.LateInitializeStringSlice(spec.TargetTags, in.TargetTags)

	if in.LogConfig != nil && spec.LogConfig == nil {
		spec.LogConfig = &v1alpha1.FirewallLogConfig{
			Enable: in.LogConfig.Enable,
		}
	}

	if len(in.Allowed) != 0 && len(spec.Allowed) == 0 {
		spec.Allowed = make([]*v1alpha1.FirewallAllowed, len(in.Allowed))
		for idx, rule := range in.Allowed {
			spec.Allowed[idx] = &v1alpha1.FirewallAllowed{
				IPProtocol: rule.IPProtocol,
				Ports:      rule.Ports,
			}
		}
	}

	if len(in.Denied) != 0 && len(spec.Denied) == 0 {
		spec.Denied = make([]*v1alpha1.FirewallDenied, len(in.Allowed))
		for idx, rule := range in.Denied {
			spec.Denied[idx] = &v1alpha1.FirewallDenied{
				IPProtocol: rule.IPProtocol,
				Ports:      rule.Ports,
			}
		}
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, in *v1alpha1.FirewallParameters, observed *compute.Firewall) (upTodate bool, err error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*compute.Firewall)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateFirewall(name, *in, desired)
	return cmp.Equal(desired, observed, cmpopts.EquateEmpty(), gcp.EquateComputeURLs(), cmpopts.IgnoreFields(compute.Firewall{}, "ForceSendFields")), nil
}
