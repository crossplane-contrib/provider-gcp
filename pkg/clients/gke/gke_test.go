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

package gke

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"google.golang.org/api/container/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/oauth2/google"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/crossplaneio/crossplane-runtime/pkg/test"
)

func TestNewClusterClient(t *testing.T) {
	type want struct {
		err error
		res *ClusterClient
	}
	tests := []struct {
		name string
		args *google.Credentials
		want want
	}{
		{name: "Test", args: &google.Credentials{}, want: want{res: &ClusterClient{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClusterClient(context.Background(), tt.args)
			if diff := cmp.Diff(err, tt.want.err, test.EquateErrors()); diff != "" {
				t.Errorf("NewClusterClient() error = %v, want.err %v\n%s", err, tt.want.err, diff)
				return
			}

			// TODO(negz): Do we really want to ignore unexported fields? I did
			// so to match the previous deep.Equal semantics here, but
			// ClusterClient _only_ has unexported fields so we're only testing
			// that NewClusterClient returns the expected type here.
			if diff := cmp.Diff(got, tt.want.res, cmpopts.IgnoreUnexported(ClusterClient{})); diff != "" {
				t.Errorf("NewClusterClient() = %v, want %v\n%s", got, tt.want.res, diff)
			}
		})
	}
}

func TestGenerateClientConfig(t *testing.T) {
	name := "gke-cluster"
	endpoint := "endpoint"
	username := "username"
	password := "password"
	clusterCA, _ := base64.StdEncoding.DecodeString("clusterCA")
	clientCert, _ := base64.StdEncoding.DecodeString("clientCert")
	clientKey, _ := base64.StdEncoding.DecodeString("clientKey")

	cases := map[string]struct {
		in  *container.Cluster
		out clientcmdapi.Config
		err error
	}{
		"Full": {
			in: &container.Cluster{
				Name:     name,
				Endpoint: endpoint,
				MasterAuth: &container.MasterAuth{
					Username:             username,
					Password:             password,
					ClusterCaCertificate: base64.StdEncoding.EncodeToString(clusterCA),
					ClientCertificate:    base64.StdEncoding.EncodeToString(clientCert),
					ClientKey:            base64.StdEncoding.EncodeToString(clientKey),
				},
			},
			out: clientcmdapi.Config{
				Clusters: map[string]*clientcmdapi.Cluster{
					name: {
						Server:                   fmt.Sprintf("https://%s", endpoint),
						CertificateAuthorityData: clusterCA,
					},
				},
				Contexts: map[string]*clientcmdapi.Context{
					name: {
						Cluster:  name,
						AuthInfo: name,
					},
				},
				AuthInfos: map[string]*clientcmdapi.AuthInfo{
					name: {
						Username:              username,
						Password:              password,
						ClientKeyData:         clientKey,
						ClientCertificateData: clientCert,
					},
				},
				CurrentContext: name,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := GenerateClientConfig(tc.in)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("GenerateClientConfig(...): -want error, +got error:\n%s", diff)
				return
			}
			if diff := cmp.Diff(tc.out, got); diff != "" {
				t.Errorf("GenerateClientConfig(...): -want error, +got error:\n%s", diff)
			}
		})
	}

}
