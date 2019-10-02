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

package network

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1alpha2"
)

var (
	testIPv4Range         = "10.0.0.0/256"
	testDescription       = "some desc"
	testName              = "some-name"
	testRoutingMode       = "GLOBAL"
	testCreationTimestamp = "10/10/2023"
	testGatewayIPv4       = "10.0.0.0"
	testSelfLink          = "/my/self/link"

	testPeeringName                = "some-peering-name"
	testPeeringNetwork             = "name"
	testPeeringState               = "ACTIVE"
	testPeeringStateDetails        = "more-detailed than ACTIVE"
	testID                  uint64 = 2029819203
	trueVal                        = true
	falseVal                       = false
)

func TestNetworkParameters_GenerateNetwork(t *testing.T) {

	cases := map[string]struct {
		in   v1alpha2.NetworkParameters
		out  *compute.Network
		fail bool
	}{
		"AutoCreateSubnetworksNil": {
			in: v1alpha2.NetworkParameters{
				IPv4Range:   &testIPv4Range,
				Description: &testDescription,
				RoutingConfig: &v1alpha2.GCPNetworkRoutingConfig{
					RoutingMode: &testRoutingMode,
				},
			},
			out: &compute.Network{
				IPv4Range:             testIPv4Range,
				Description:           testDescription,
				AutoCreateSubnetworks: false,
				Name:                  testName,
				RoutingConfig: &compute.NetworkRoutingConfig{
					RoutingMode: testRoutingMode,
				},
			},
		},
		"AutoCreateSubnetworksFalse": {
			in: v1alpha2.NetworkParameters{
				IPv4Range:             &testIPv4Range,
				Description:           &testDescription,
				AutoCreateSubnetworks: &falseVal,
				RoutingConfig: &v1alpha2.GCPNetworkRoutingConfig{
					RoutingMode: &testRoutingMode,
				},
			},
			out: &compute.Network{
				IPv4Range:             testIPv4Range,
				Description:           testDescription,
				AutoCreateSubnetworks: false,
				Name:                  testName,
				RoutingConfig: &compute.NetworkRoutingConfig{
					RoutingMode: testRoutingMode,
				},
				ForceSendFields: []string{"AutoCreateSubnetworks"},
			},
		},
		"AutoCreateSubnetworksTrue": {
			in: v1alpha2.NetworkParameters{
				IPv4Range:             &testIPv4Range,
				Description:           &testDescription,
				AutoCreateSubnetworks: &trueVal,
				RoutingConfig: &v1alpha2.GCPNetworkRoutingConfig{
					RoutingMode: &testRoutingMode,
				},
			},
			out: &compute.Network{
				IPv4Range:             testIPv4Range,
				Description:           testDescription,
				AutoCreateSubnetworks: true,
				Name:                  testName,
				RoutingConfig: &compute.NetworkRoutingConfig{
					RoutingMode: testRoutingMode,
				},
			},
		},
		"AutoCreateSubnetworksTrueFail": {
			in: v1alpha2.NetworkParameters{
				IPv4Range:             &testIPv4Range,
				Description:           &testDescription,
				AutoCreateSubnetworks: &trueVal,
				RoutingConfig: &v1alpha2.GCPNetworkRoutingConfig{
					RoutingMode: &testRoutingMode,
				},
			},
			out: &compute.Network{
				IPv4Range:             testIPv4Range,
				Description:           testDescription,
				AutoCreateSubnetworks: false,
				Name:                  testName,
				RoutingConfig: &compute.NetworkRoutingConfig{
					RoutingMode: testRoutingMode,
				},
			},
			fail: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			n := v1alpha2.Network{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						externalResourceNameAnnotationKey: testName,
					},
				},
				Spec: v1alpha2.NetworkSpec{
					ForProvider: tc.in,
				},
			}
			r := GenerateNetwork(n)
			if diff := cmp.Diff(r, tc.out); diff != "" && !tc.fail {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateGCPNetworkStatus(t *testing.T) {
	cases := map[string]struct {
		in   compute.Network
		out  v1alpha2.GCPNetworkStatus
		fail bool
	}{
		"AllFilled": {
			in: compute.Network{
				IPv4Range:         testIPv4Range,
				Description:       testDescription,
				CreationTimestamp: testCreationTimestamp,
				GatewayIPv4:       testGatewayIPv4,
				Id:                testID,
				SelfLink:          testSelfLink,
				Peerings: []*compute.NetworkPeering{
					{
						AutoCreateRoutes:     true,
						ExchangeSubnetRoutes: true,
						Name:                 testPeeringName,
						Network:              testPeeringNetwork,
						State:                testPeeringState,
						StateDetails:         testPeeringStateDetails,
					},
				},
				Subnetworks: []string{
					"my-subnetwork",
				},
				Name: testName,
				RoutingConfig: &compute.NetworkRoutingConfig{
					RoutingMode: testRoutingMode,
				},
			},
			out: v1alpha2.GCPNetworkStatus{
				IPv4Range:             &testIPv4Range,
				Description:           &testDescription,
				CreationTimestamp:     &testCreationTimestamp,
				GatewayIPv4:           &testGatewayIPv4,
				ID:                    &testID,
				AutoCreateSubnetworks: &falseVal,
				SelfLink:              &testSelfLink,
				Peerings: []*v1alpha2.GCPNetworkPeering{
					{
						AutoCreateRoutes:     &trueVal,
						ExchangeSubnetRoutes: &trueVal,
						Name:                 &testPeeringName,
						Network:              &testPeeringNetwork,
						State:                &testPeeringState,
						StateDetails:         &testPeeringStateDetails,
					},
				},
				Subnetworks: []string{
					"my-subnetwork",
				},
				RoutingConfig: &v1alpha2.GCPNetworkRoutingConfig{
					RoutingMode: &testRoutingMode,
				},
			},
		},
		"AllFilledFail": {
			in: compute.Network{
				IPv4Range:             testIPv4Range,
				Description:           testDescription,
				CreationTimestamp:     testCreationTimestamp,
				GatewayIPv4:           testGatewayIPv4,
				Id:                    testID,
				SelfLink:              testSelfLink,
				AutoCreateSubnetworks: true,
				Peerings: []*compute.NetworkPeering{
					{
						AutoCreateRoutes:     true,
						ExchangeSubnetRoutes: true,
						Name:                 testPeeringName,
						Network:              testPeeringNetwork,
						State:                testPeeringState,
						StateDetails:         testPeeringStateDetails,
					},
				},
				Subnetworks: []string{
					"my-subnetwork",
				},
				Name: testName,
				RoutingConfig: &compute.NetworkRoutingConfig{
					RoutingMode: testRoutingMode,
				},
			},
			out: v1alpha2.GCPNetworkStatus{
				IPv4Range:             &testIPv4Range,
				Description:           &testDescription,
				CreationTimestamp:     &testCreationTimestamp,
				GatewayIPv4:           &testGatewayIPv4,
				AutoCreateSubnetworks: &trueVal,
				SelfLink:              &testSelfLink,
				ID:                    &testID,
				Peerings: []*v1alpha2.GCPNetworkPeering{
					{
						AutoCreateRoutes:     &trueVal,
						ExchangeSubnetRoutes: &trueVal,
						Name:                 &testPeeringName,
						Network:              &testPeeringNetwork,
						State:                &testPeeringState,
						StateDetails:         &testPeeringStateDetails,
					},
				},
				Subnetworks: []string{
					"my-subnetwork",
				},
				RoutingConfig: &v1alpha2.GCPNetworkRoutingConfig{
					RoutingMode: &testRoutingMode,
				},
			},
			fail: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateGCPNetworkStatus(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" && !tc.fail {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
