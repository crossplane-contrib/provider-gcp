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
	"testing"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"

	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v1"
)

var (
	description               = "coolDescription"
	region                    = "coolRegion"
	zone                      = "coolZone"
	networkEndpointType       = "coolType"
	network                   = "coolNetwork"
	subnetwork                = "coolSubnet"
	defaultPort         int64 = 3002

	service  string = "myservice"
	tag             = "mytag"
	urlMask         = "urlmask"
	version         = "myversion"
	function        = "myfunction"

	ip1             = "myip1"
	fqdn1           = "myfqdn1"
	port1     int64 = 3003
	instance1       = "myinstance1"

	ip2             = "myip2"
	fqdn2           = "myfqdn2"
	port2     int64 = 3004
	instance2       = "myinstance2"

	cloudRun = v1beta1.CloudRunParameters{
		Service: service,
		Tag:     &tag,
		URLMask: &urlMask,
	}
	appEngine = v1beta1.AppEngineParameters{
		Service: &service,
		Version: &version,
		URLMask: &urlMask,
	}
	cloudFunction = v1beta1.CloudFunctionParameters{
		Function: function,
		URLMask:  &urlMask,
	}

	networkEndpoint1 = v1beta1.NetworkEndpoint{
		IPAddress: &ip1,
		FQDN:      &fqdn1,
		Port:      &port1,
		Instance:  &instance1,
	}

	networkEndpoint2 = v1beta1.NetworkEndpoint{
		IPAddress: &ip2,
		FQDN:      &fqdn2,
		Port:      &port2,
		Instance:  &instance2,
	}

	networkEndpoints = []v1beta1.NetworkEndpoint{
		networkEndpoint1,
		networkEndpoint2,
	}

	ccloudRun = compute.NetworkEndpointGroupCloudRun{
		Service: service,
		Tag:     tag,
		UrlMask: urlMask,
	}
	cappEngine = compute.NetworkEndpointGroupAppEngine{
		Service: service,
		UrlMask: urlMask,
		Version: version,
	}
	ccloudFunction = compute.NetworkEndpointGroupCloudFunction{
		Function: function,
		UrlMask:  urlMask,
	}

	cnetworkEndpoint1 = compute.NetworkEndpoint{
		IpAddress: ip1,
		Fqdn:      fqdn1,
		Port:      port1,
		Instance:  instance1,
	}

	cnetworkEndpoint2 = compute.NetworkEndpoint{
		IpAddress: ip2,
		Fqdn:      fqdn2,
		Port:      port2,
		Instance:  instance2,
	}

	cnetworkEndpoints = []compute.NetworkEndpoint{
		cnetworkEndpoint1,
		cnetworkEndpoint2,
	}

	forwardingRule1     = "myforwardingrule1"
	backendService1     = "backendService1"
	healhCheck1         = "myhealhCheck1"
	healthCheckService1 = "healthCheckService1"
	healthok1           = "myhealthok1"

	forwardingRule2     = "myforwardingrule2"
	backendService2     = "backendService2"
	healhCheck2         = "myhealhCheck2"
	healthCheckService2 = "healthCheckService2"
	healthok2           = "myhealthok2"

	networkEndpointHealth1 = v1beta1.NetworkEndpointHealth{
		ForwardingRule: v1beta1.NetworkEndpointForwardingRule{
			ForwardingRule: forwardingRule1,
		},
		BackendService: v1beta1.NetworkEndpointBackendService{
			BackendService: backendService1,
		},
		HealthCheck: v1beta1.NetworkEndpointHealthCheck{
			HealthCheck: healhCheck1,
		},
		HealthCheckService: v1beta1.NetworkEndpointHealthCheckService{
			HealthCheckService: healthCheckService1,
		},
		HealthState: healthok1,
	}

	networkEndpointHealth2 = v1beta1.NetworkEndpointHealth{
		ForwardingRule: v1beta1.NetworkEndpointForwardingRule{
			ForwardingRule: forwardingRule2,
		},
		BackendService: v1beta1.NetworkEndpointBackendService{
			BackendService: backendService2,
		},
		HealthCheck: v1beta1.NetworkEndpointHealthCheck{
			HealthCheck: healhCheck2,
		},
		HealthCheckService: v1beta1.NetworkEndpointHealthCheckService{
			HealthCheckService: healthCheckService2,
		},
		HealthState: healthok2,
	}

	networkEndpointObservations = []v1beta1.NetworkEndpointObservation{
		{
			NetworkEndpoint: networkEndpoint1,
			NetworkEndpointHealths: []v1beta1.NetworkEndpointHealth{
				networkEndpointHealth1,
			},
		},
		{
			NetworkEndpoint: networkEndpoint1,
			NetworkEndpointHealths: []v1beta1.NetworkEndpointHealth{
				networkEndpointHealth2,
			},
		},
	}

	healthStatusForNetworkEndpoint1 = compute.HealthStatusForNetworkEndpoint{
		BackendService: &compute.BackendServiceReference{
			BackendService: backendService1,
		},
		ForwardingRule: &compute.ForwardingRuleReference{
			ForwardingRule: forwardingRule1,
		},
		HealthCheck: &compute.HealthCheckReference{
			HealthCheck: healhCheck1,
		},
		HealthCheckService: &compute.HealthCheckServiceReference{
			HealthCheckService: healthCheckService1,
		},
		HealthState: healthok1,
	}

	healthStatusForNetworkEndpoint2 = compute.HealthStatusForNetworkEndpoint{
		BackendService: &compute.BackendServiceReference{
			BackendService: backendService2,
		},
		ForwardingRule: &compute.ForwardingRuleReference{
			ForwardingRule: forwardingRule2,
		},
		HealthCheck: &compute.HealthCheckReference{
			HealthCheck: healhCheck2,
		},
		HealthCheckService: &compute.HealthCheckServiceReference{
			HealthCheckService: healthCheckService2,
		},
		HealthState: healthok2,
	}

	networkEnpointsStatus = []*compute.NetworkEndpointWithHealthStatus{
		{
			NetworkEndpoint: &cnetworkEndpoint1,
			Healths: []*compute.HealthStatusForNetworkEndpoint{
				&healthStatusForNetworkEndpoint1,
			},
		},
		{
			NetworkEndpoint: &cnetworkEndpoint2,
			Healths: []*compute.HealthStatusForNetworkEndpoint{
				&healthStatusForNetworkEndpoint2,
			},
		},
	}

	timestamp        = "coolTime"
	link             = "coolLink"
	name             = "coolName"
	id        uint64 = 3001
	size      int64  = 2
)

func params(m ...func(*v1beta1.NetworkEndpointGroupParameters)) *v1beta1.NetworkEndpointGroupParameters {
	o := &v1beta1.NetworkEndpointGroupParameters{
		Region:              &region,
		Zone:                &zone,
		NetworkEndpointType: networkEndpointType,
		Description:         &description,
		Network:             &network,
		Subnetwork:          &subnetwork,
		DefaultPort:         defaultPort,
		CloudRun:            &cloudRun,
		AppEngine:           &appEngine,
		CloudFunction:       &cloudFunction,
		NetworkEndpoints:    networkEndpoints,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func networkEnpointGroup(m ...func(*compute.NetworkEndpointGroup)) *compute.NetworkEndpointGroup {
	o := &compute.NetworkEndpointGroup{
		Name:                name,
		Region:              region,
		Zone:                zone,
		NetworkEndpointType: networkEndpointType,
		Description:         description,
		Network:             network,
		Subnetwork:          subnetwork,
		DefaultPort:         defaultPort,
		CloudRun:            &ccloudRun,
		AppEngine:           &cappEngine,
		CloudFunction:       &ccloudFunction,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func networkEnpoints(m ...func([]compute.NetworkEndpoint)) []compute.NetworkEndpoint {
	o := cnetworkEndpoints

	for _, f := range m {
		f(o)
	}

	return o
}

func networkEnpointsWithHealthStatus(m ...func([]*compute.NetworkEndpointWithHealthStatus)) []*compute.NetworkEndpointWithHealthStatus {
	o := networkEnpointsStatus

	for _, f := range m {
		f(o)
	}

	return o
}

func addOutputFields(n *compute.NetworkEndpointGroup) {
	n.CreationTimestamp = timestamp
	n.Id = id
	n.SelfLink = link
	n.Size = size
}

func observation(m ...func(*v1beta1.NetworkEndpointGroupObservation)) *v1beta1.NetworkEndpointGroupObservation {
	o := &v1beta1.NetworkEndpointGroupObservation{
		CreationTimestamp:           timestamp,
		ID:                          id,
		SelfLink:                    link,
		Size:                        size,
		NetworkEndpointObservations: networkEndpointObservations,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestGenerateNetworkEndpointGroup(t *testing.T) {
	type args struct {
		name string
		in   v1beta1.NetworkEndpointGroupParameters
	}
	cases := map[string]struct {
		args    args
		want    *compute.NetworkEndpointGroup
		wantnes []compute.NetworkEndpoint
	}{
		"AllFilled": {
			args: args{
				name: name,
				in:   *params(),
			},
			want:    networkEnpointGroup(),
			wantnes: networkEnpoints(),
		},
		"PartialFilled": {
			args: args{
				name: name,
				in: *params(func(p *v1beta1.NetworkEndpointGroupParameters) {
					p.AppEngine = nil
					p.CloudRun = nil
					p.CloudFunction = nil
					p.Zone = nil
					p.Region = nil
					p.NetworkEndpoints = nil
				}),
			},
			want: networkEnpointGroup(func(a *compute.NetworkEndpointGroup) {
				a.AppEngine = nil
				a.CloudRun = nil
				a.CloudFunction = nil
				a.Zone = ""
				a.Region = ""
			}),
			wantnes: []compute.NetworkEndpoint{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r, rs := GenerateNetworkEndpointGroup(tc.args.name, tc.args.in)
			if diff := cmp.Diff(r, tc.want); diff != "" {
				t.Errorf("GenerateGlobalAddress(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(rs, tc.wantnes); diff != "" {
				t.Errorf("GenerateGlobalAddress(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateNetworkEndpointGroupObservation(t *testing.T) {
	cases := map[string]struct {
		in     compute.NetworkEndpointGroup
		innegs []*compute.NetworkEndpointWithHealthStatus
		out    v1beta1.NetworkEndpointGroupObservation
	}{
		"AllFilled": {
			in:     *networkEnpointGroup(addOutputFields),
			innegs: networkEnpointsWithHealthStatus(),
			out:    *observation(),
		},
		"PartialFilled": {
			in: *networkEnpointGroup(addOutputFields, func(a *compute.NetworkEndpointGroup) {
				a.CreationTimestamp = ""
			}),
			innegs: []*compute.NetworkEndpointWithHealthStatus{},
			out: *observation(func(o *v1beta1.NetworkEndpointGroupObservation) {
				o.CreationTimestamp = ""
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateNetworkEndpointGroupObservation(tc.in, tc.innegs)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateGlobalAddressObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpec(t *testing.T) {
	type args struct {
		spec   *v1beta1.NetworkEndpointGroupParameters
		in     *compute.NetworkEndpointGroup
		innegs []*compute.NetworkEndpointWithHealthStatus
	}
	cases := map[string]struct {
		args args
		want *v1beta1.NetworkEndpointGroupParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec:   params(),
				in:     networkEnpointGroup(),
				innegs: networkEnpointsWithHealthStatus(),
			},
			want: params(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: params(),
				in: networkEnpointGroup(func(a *compute.NetworkEndpointGroup) {
					a.Description = "some other description"
				}),
				innegs: networkEnpointsWithHealthStatus(),
			},

			want: params(),
		},
		"PartialFilled": {
			args: args{
				spec: params(func(p *v1beta1.NetworkEndpointGroupParameters) {
					p.Region = nil
					p.Zone = nil
				}),
				in:     networkEnpointGroup(),
				innegs: networkEnpointsWithHealthStatus(),
			},
			want: params(func(p *v1beta1.NetworkEndpointGroupParameters) {
				p.Region = &region
				p.Zone = &zone
			}),
		},
		"NoEndpoints": {
			args: args{
				spec:   params(),
				in:     networkEnpointGroup(),
				innegs: []*compute.NetworkEndpointWithHealthStatus{},
			},
			want: params(func(p *v1beta1.NetworkEndpointGroupParameters) {
				p.NetworkEndpoints = []v1beta1.NetworkEndpoint{}
			}),
		},
		"LessEndpoints": {
			args: args{
				spec:   params(),
				in:     networkEnpointGroup(),
				innegs: networkEnpointsWithHealthStatus()[0:1],
			},
			want: params(func(p *v1beta1.NetworkEndpointGroupParameters) {
				p.NetworkEndpoints = p.NetworkEndpoints[0:1]
			}),
		},
		"MoreEndpoints": {
			args: args{
				spec: params(),
				in:   networkEnpointGroup(),
				innegs: append(networkEnpointsWithHealthStatus(), &compute.NetworkEndpointWithHealthStatus{
					NetworkEndpoint: &cnetworkEndpoint2,
					Healths: []*compute.HealthStatusForNetworkEndpoint{
						&healthStatusForNetworkEndpoint2,
					},
				}),
			},
			want: params(func(p *v1beta1.NetworkEndpointGroupParameters) {
				p.NetworkEndpoints = append(p.NetworkEndpoints, networkEndpoint2)
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.spec, *tc.args.in, tc.args.innegs)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}
