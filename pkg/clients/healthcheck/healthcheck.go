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

package healthcheck

import (
	compute "google.golang.org/api/compute/v1"

	"github.com/crossplane/provider-gcp/apis/compute/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// GenerateHealthCheck takes a *HealthCheckParameters and returns *compute.HealthCheck.
// It assigns only the fields that are writable, i.e. not labelled as [Output Only]
// in Google's reference.
func GenerateHealthCheck(name string, in v1alpha1.HealthCheckParameters, hc *compute.HealthCheck) {
	hc.Name = name
	hc.Description = gcp.StringValue(in.Description)
	hc.CheckIntervalSec = gcp.Int64Value(in.CheckIntervalSec)
	hc.HealthyThreshold = gcp.Int64Value(in.HealthyThreshold)
	hc.Kind = gcp.StringValue(in.Kind)
	hc.TimeoutSec = gcp.Int64Value(in.TimeoutSec)
	hc.Type = gcp.StringValue(in.Type)
	hc.UnhealthyThreshold = gcp.Int64Value(in.UnhealthyThreshold)
	if in.Http2HealthCheck != nil {
		hc.Http2HealthCheck = &compute.HTTP2HealthCheck{
			Host:              gcp.StringValue(in.Http2HealthCheck.Host),
			Port:              gcp.Int64Value(in.Http2HealthCheck.Port),
			PortName:          gcp.StringValue(in.Http2HealthCheck.PortName),
			PortSpecification: gcp.StringValue(in.Http2HealthCheck.PortSpecification),
			ProxyHeader:       gcp.StringValue(in.Http2HealthCheck.ProxyHeader),
			RequestPath:       gcp.StringValue(in.Http2HealthCheck.RequestPath),
			Response:          gcp.StringValue(in.Http2HealthCheck.Response),
		}
	}
	if in.HttpHealthCheck != nil {
		hc.HttpHealthCheck = &compute.HTTPHealthCheck{
			Host:              gcp.StringValue(in.HttpHealthCheck.Host),
			Port:              gcp.Int64Value(in.HttpHealthCheck.Port),
			PortName:          gcp.StringValue(in.HttpHealthCheck.PortName),
			PortSpecification: gcp.StringValue(in.HttpHealthCheck.PortSpecification),
			ProxyHeader:       gcp.StringValue(in.HttpHealthCheck.ProxyHeader),
			RequestPath:       gcp.StringValue(in.HttpHealthCheck.RequestPath),
			Response:          gcp.StringValue(in.HttpHealthCheck.Response),
		}
	}
	if in.HttpsHealthCheck != nil {
		hc.HttpsHealthCheck = &compute.HTTPSHealthCheck{
			Host:              gcp.StringValue(in.HttpsHealthCheck.Host),
			Port:              gcp.Int64Value(in.HttpsHealthCheck.Port),
			PortName:          gcp.StringValue(in.HttpsHealthCheck.PortName),
			PortSpecification: gcp.StringValue(in.HttpsHealthCheck.PortSpecification),
			ProxyHeader:       gcp.StringValue(in.HttpsHealthCheck.ProxyHeader),
			RequestPath:       gcp.StringValue(in.HttpsHealthCheck.RequestPath),
			Response:          gcp.StringValue(in.HttpsHealthCheck.Response),
		}
	}
	if in.SslHealthCheck != nil {
		hc.SslHealthCheck = &compute.SSLHealthCheck{
			Port:              gcp.Int64Value(in.SslHealthCheck.Port),
			PortName:          gcp.StringValue(in.SslHealthCheck.PortName),
			PortSpecification: gcp.StringValue(in.SslHealthCheck.PortSpecification),
			ProxyHeader:       gcp.StringValue(in.SslHealthCheck.ProxyHeader),
			Response:          gcp.StringValue(in.SslHealthCheck.Response),
		}
	}
	if in.TcpHealthCheck != nil {
		hc.TcpHealthCheck = &compute.TCPHealthCheck{
			Port:              gcp.Int64Value(in.TcpHealthCheck.Port),
			PortName:          gcp.StringValue(in.TcpHealthCheck.PortName),
			PortSpecification: gcp.StringValue(in.TcpHealthCheck.PortSpecification),
			ProxyHeader:       gcp.StringValue(in.TcpHealthCheck.ProxyHeader),
			Response:          gcp.StringValue(in.TcpHealthCheck.Response),
		}
	}
}

// // GenerateHealthCheckObservation takes a compute.HealthCheck and returns *HealthCheckObservation.
// func GenerateHealthCheckObservation(in compute.HealthCheck) v1alpha1.HealthCheckObservation {
// 	ro := v1alpha1.HealthCheckObservation{
// 		CreationTimestamp: in.CreationTimestamp,
// 		GatewayIPv4:       in.GatewayIPv4,
// 		ID:                in.Id,
// 		SelfLink:          in.SelfLink,
// 		SubHealthChecks:        in.SubHealthChecks,
// 	}
// 	return ro
// }

// // LateInitializeSpec fills unassigned fields with the values in compute.HealthCheck object.
// func LateInitializeSpec(spec *v1beta1.HealthCheckParameters, in compute.HealthCheck) {
// 	spec.AutoCreateSubHealthChecks = gcp.LateInitializeBool(spec.AutoCreateSubHealthChecks, in.AutoCreateSubHealthChecks)
// 	if in.RoutingConfig != nil && spec.RoutingConfig == nil {
// 		spec.RoutingConfig = &v1beta1.HealthCheckRoutingConfig{
// 			RoutingMode: in.RoutingConfig.RoutingMode,
// 		}
// 	}

// 	spec.Description = gcp.LateInitializeString(spec.Description, in.Description)
// }

// // IsUpToDate checks whether current state is up-to-date compared to the given
// // set of parameters.
// func IsUpToDate(name string, in *v1beta1.HealthCheckParameters, observed *compute.HealthCheck) (upTodate bool, switchToCustom bool, err error) {
// 	generated, err := copystructure.Copy(observed)
// 	if err != nil {
// 		return true, false, errors.Wrap(err, errCheckUpToDate)
// 	}
// 	desired, ok := generated.(*compute.HealthCheck)
// 	if !ok {
// 		return true, false, errors.New(errCheckUpToDate)
// 	}
// 	GenerateHealthCheck(name, *in, desired)
// 	if !desired.AutoCreateSubHealthChecks && observed.AutoCreateSubHealthChecks {
// 		return false, true, nil
// 	}
// 	return cmp.Equal(desired, observed, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(compute.HealthCheck{}, "ForceSendFields")), false, nil
// }
