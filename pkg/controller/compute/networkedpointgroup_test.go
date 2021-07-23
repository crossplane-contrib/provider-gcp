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

package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
)

const (
	testNetworkEndpointGroupName = "test-networkendpointgroup"
)

var _ managed.ExternalConnecter = &negConnector{}
var _ managed.ExternalClient = &negExternal{}

type networkEndpointGroupModifier func(*v1beta1.NetworkEndpointGroup)

func networkEndpointGroupWithConditions(c ...xpv1.Condition) networkEndpointGroupModifier {
	return func(i *v1beta1.NetworkEndpointGroup) { i.Status.SetConditions(c...) }
}

func networkEndpointGroupWithDescription(d string) networkEndpointGroupModifier {
	return func(i *v1beta1.NetworkEndpointGroup) { i.Spec.ForProvider.Description = &d }
}

// func subnetworkWithPrivateAccess(p bool) subnetworkModifier {
// 	return func(i *v1beta1.Subnetwork) { i.Spec.ForProvider.PrivateIPGoogleAccess = &p }
// }

func networkEndpointGroupObj(im ...networkEndpointGroupModifier) *v1beta1.NetworkEndpointGroup {
	i := &v1beta1.NetworkEndpointGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testNetworkEndpointGroupName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: testNetworkEndpointGroupName,
			},
		},
		Spec: v1beta1.NetworkEndpointGroupSpec{
			ForProvider: v1beta1.NetworkEndpointGroupParameters{},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

func TestNetworkEndpointGroupObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"NotNetworkEndpointGroup": {
			handler: nil,
			args: args{
				mg: &v1beta1.NetworkEndpointGroup{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotNetworkEndpointGroup),
			},
		},
		"NotFound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(&compute.NetworkEndpointGroup{})
			}),
			args: args{
				mg: networkEndpointGroupObj(),
			},
			want: want{
				mg:  networkEndpointGroupObj(),
				err: nil,
			},
		},
		"GetFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&compute.NetworkEndpointGroup{})
			}),
			args: args{
				mg: networkEndpointGroupObj(),
			},
			want: want{
				mg:  networkEndpointGroupObj(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetNetworkEndpointGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}
