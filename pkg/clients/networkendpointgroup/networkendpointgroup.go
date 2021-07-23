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

package networkendpointgroup

import (
	"fmt"

	"github.com/scylladb/go-set/strset"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"

	compute "google.golang.org/api/compute/v1"

	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied NetworkEndpointGroupParameters that are set (i.e. non-zero) on the supplied
// NetworkEndpointGroup.
func LateInitializeSpec(p *v1beta1.NetworkEndpointGroupParameters, observed compute.NetworkEndpointGroup, observedEndpoints []*compute.NetworkEndpointWithHealthStatus) {
	p.Network = gcp.LateInitializeString(p.Network, observed.Network)
	p.Subnetwork = gcp.LateInitializeString(p.Subnetwork, observed.Subnetwork)
	p.Region = gcp.LateInitializeString(p.Region, observed.Region)
	p.Zone = gcp.LateInitializeString(p.Zone, observed.Zone)

	p.AppEngine = gcp.LateInitializeStruct(p.AppEngine, observed.AppEngine).(*v1beta1.AppEngineParameters)
	p.CloudFunction = gcp.LateInitializeStruct(p.CloudFunction, observed.CloudFunction).(*v1beta1.CloudFunctionParameters)
	p.CloudRun = gcp.LateInitializeStruct(p.CloudRun, observed.CloudRun).(*v1beta1.CloudRunParameters)

	networkEndpoints := []v1beta1.NetworkEndpoint{}
	for _, observedEndpoint := range observedEndpoints {
		networkEndpoints = append(networkEndpoints, v1beta1.NetworkEndpoint{
			IPAddress: &observedEndpoint.NetworkEndpoint.IpAddress,
			FQDN:      &observedEndpoint.NetworkEndpoint.Fqdn,
			Port:      &observedEndpoint.NetworkEndpoint.Port,
			Instance:  &observedEndpoint.NetworkEndpoint.Instance,
		})
	}
	p.NetworkEndpoints = networkEndpoints
}

// GenerateNetworkEndpointGroupObservation takes a compute.Address and returns
// *NetworkEndpointGroupObservation.
func GenerateNetworkEndpointGroupObservation(observed compute.NetworkEndpointGroup, observedEndpoints []*compute.NetworkEndpointWithHealthStatus) v1beta1.NetworkEndpointGroupObservation {
	networkEndpointObservations := []v1beta1.NetworkEndpointObservation{}
	for _, observedEndpoint := range observedEndpoints {
		networkEndpointHealths := []v1beta1.NetworkEndpointHealth{}
		for _, networkEndpointHealth := range observedEndpoint.Healths {
			networkEndpointHealths = append(networkEndpointHealths, v1beta1.NetworkEndpointHealth{
				ForwardingRule: v1beta1.NetworkEndpointForwardingRule{
					ForwardingRule: networkEndpointHealth.ForwardingRule.ForwardingRule,
				},
				BackendService: v1beta1.NetworkEndpointBackendService{
					BackendService: networkEndpointHealth.BackendService.BackendService,
				},
				HealthCheck: v1beta1.NetworkEndpointHealthCheck{
					HealthCheck: networkEndpointHealth.HealthCheck.HealthCheck,
				},
				HealthCheckService: v1beta1.NetworkEndpointHealthCheckService{
					HealthCheckService: networkEndpointHealth.HealthCheckService.HealthCheckService,
				},
				HealthState: networkEndpointHealth.HealthState,
			})
		}
		networkEndpointObservations = append(networkEndpointObservations, v1beta1.NetworkEndpointObservation{
			NetworkEndpoint: v1beta1.NetworkEndpoint{
				IPAddress: &observedEndpoint.NetworkEndpoint.IpAddress,
				FQDN:      &observedEndpoint.NetworkEndpoint.Fqdn,
				Port:      &observedEndpoint.NetworkEndpoint.Port,
				Instance:  &observedEndpoint.NetworkEndpoint.Instance,
			},
			NetworkEndpointHealths: networkEndpointHealths,
		})
	}
	return v1beta1.NetworkEndpointGroupObservation{
		CreationTimestamp:           observed.CreationTimestamp,
		ID:                          observed.Id,
		SelfLink:                    observed.SelfLink,
		Size:                        observed.Size,
		NetworkEndpointObservations: networkEndpointObservations,
	}
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(name string, p *v1beta1.NetworkEndpointGroupParameters, observed *compute.NetworkEndpointGroup, observedEndpoints []*compute.NetworkEndpointWithHealthStatus) (upTodate bool, switchToCustom bool, toBeAdded []*compute.NetworkEndpoint, toBeRemoved []*compute.NetworkEndpoint, err error) {
	// The onlu thing that can be updated in a network endpoint group are the networkendpoints.
	// To find out of we need to update any network endpoints, we turn them into string set and we calculate the left and right difference.
	desiredEndpointMap := map[string]*compute.NetworkEndpoint{}
	actualEndpointMap := map[string]*compute.NetworkEndpoint{}
	desiredEndpointKeyList := []string{}
	actualEndpointKeyList := []string{}
	for _, observedNetworkEndpoint := range observedEndpoints {
		actualEndpointKeyList = append(actualEndpointKeyList, getKeyFromProviderNetworkEndpoint(observedNetworkEndpoint.NetworkEndpoint))
		actualEndpointMap[getKeyFromProviderNetworkEndpoint(observedNetworkEndpoint.NetworkEndpoint)] = observedNetworkEndpoint.NetworkEndpoint
	}
	for _, desiredNetworkEndpoint := range p.NetworkEndpoints {
		key := getKeyFromNetworkEndpoint(desiredNetworkEndpoint.DeepCopy())
		desiredEndpointKeyList = append(desiredEndpointKeyList, key)
		desiredEndpointMap[key] = &compute.NetworkEndpoint{
			Fqdn:      gcp.StringValue(desiredNetworkEndpoint.FQDN),
			Instance:  gcp.StringValue(desiredNetworkEndpoint.Instance),
			IpAddress: gcp.StringValue(desiredNetworkEndpoint.IPAddress),
			Port:      gcp.Int64Value(desiredNetworkEndpoint.Port),
		}
	}
	desiredEndpointSet := strset.New(desiredEndpointKeyList...)
	actualEndpointSet := strset.New(actualEndpointKeyList...)
	toBeAdded = []*compute.NetworkEndpoint{}
	toBeRemoved = []*compute.NetworkEndpoint{}
	leftDiff := strset.Difference(desiredEndpointSet, actualEndpointSet)
	rightDiff := strset.Difference(actualEndpointSet, desiredEndpointSet)
	for _, endpointKey := range leftDiff.List() {
		toBeAdded = append(toBeAdded, desiredEndpointMap[endpointKey])
	}
	for _, endpointkey := range rightDiff.List() {
		toBeRemoved = append(toBeRemoved, actualEndpointMap[endpointkey])
	}
	return leftDiff.Size() == 0 && rightDiff.Size() == 0, false, toBeAdded, toBeRemoved, nil
}

func getKeyFromNetworkEndpoint(networkEndpoint *v1beta1.NetworkEndpoint) string {
	return *networkEndpoint.FQDN + "_" + *networkEndpoint.IPAddress + "_" + fmt.Sprint(*networkEndpoint.Port) + "_" + *networkEndpoint.Instance
}

func getKeyFromProviderNetworkEndpoint(networkEndpoint *compute.NetworkEndpoint) string {
	return networkEndpoint.Fqdn + "_" + networkEndpoint.IpAddress + "_" + fmt.Sprint(networkEndpoint.Port) + "_" + networkEndpoint.Instance
}

// GenerateNetworkEndpointGroup converts the supplied NetworkEndpointGroupParameters into an
// NetworkEndpointGroup suitable for use with the Google Compute API.
func GenerateNetworkEndpointGroup(name string, in v1beta1.NetworkEndpointGroupParameters) (*compute.NetworkEndpointGroup, []*compute.NetworkEndpoint) {
	// Kubernetes API conventions dictate that optional, unspecified fields must
	// be nil. GCP API clients omit any field set to its zero value, using
	// NullFields and ForceSendFields to handle edge cases around unsetting
	// previously set values, or forcing zero values to be set. The Address API
	// does not support updates, so we can safely convert any nil pointer to
	// string or int64 to their zero values.
	networkEndpointGroup := &compute.NetworkEndpointGroup{}
	networkEndpointGroup.Description = gcp.StringValue(in.Description)
	if in.AppEngine != nil {
		networkEndpointGroup.AppEngine = &compute.NetworkEndpointGroupAppEngine{
			Service: gcp.StringValue(in.AppEngine.Service),
			UrlMask: gcp.StringValue(in.AppEngine.URLMask),
			Version: gcp.StringValue(in.AppEngine.Version),
		}
	}
	if in.CloudFunction != nil {
		networkEndpointGroup.CloudFunction = &compute.NetworkEndpointGroupCloudFunction{
			Function: gcp.StringValue(&in.CloudFunction.Function),
			UrlMask:  gcp.StringValue(in.CloudFunction.URLMask),
		}
	}
	if in.CloudRun != nil {
		networkEndpointGroup.CloudRun = &compute.NetworkEndpointGroupCloudRun{
			Service: gcp.StringValue(&in.CloudRun.Service),
			UrlMask: gcp.StringValue(in.CloudRun.URLMask),
			Tag:     gcp.StringValue(in.CloudRun.Tag),
		}
	}

	networkEndpointGroup.DefaultPort = in.DefaultPort
	networkEndpointGroup.Subnetwork = gcp.StringValue(in.Subnetwork)
	networkEndpointGroup.Network = gcp.StringValue(in.Network)
	networkEndpointGroup.NetworkEndpointType = gcp.StringValue(&in.NetworkEndpointType)

	networkendpoints := []*compute.NetworkEndpoint{}

	for _, networkendpoint := range in.NetworkEndpoints {
		networkendpoints = append(networkendpoints, &compute.NetworkEndpoint{
			Fqdn:      gcp.StringValue(networkendpoint.FQDN),
			Instance:  gcp.StringValue(networkendpoint.Instance),
			IpAddress: gcp.StringValue(networkendpoint.IPAddress),
			Port:      gcp.Int64Value(networkendpoint.Port),
		})
	}

	return networkEndpointGroup, networkendpoints
}
