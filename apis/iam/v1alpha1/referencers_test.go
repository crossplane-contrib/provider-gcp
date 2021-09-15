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

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

const (
	testEmail             = "test-service-account@key-test-project.iam.gserviceaccount.com"
	nameTestSa            = "test-sa"
	labelTest             = "test-key"
	valueTest             = "test-value"
	rrnTestServiceAccount = "projects/key-test-project/serviceAccounts/" + testEmail
)

func TestServiceAccountMemberName(t *testing.T) {
	testCases := map[string]struct {
		mg   resource.Managed
		want string
	}{
		"NotServiceAccount": {
			mg:   &ServiceAccountKey{},
			want: "",
		},
		"EmptyEmail": {
			mg: &ServiceAccount{
				Status: ServiceAccountStatus{
					AtProvider: ServiceAccountObservation{
						Email: "",
					},
				},
			},
			want: "",
		},
		"NonEmptyEmail": {
			mg: &ServiceAccount{
				Status: ServiceAccountStatus{
					AtProvider: ServiceAccountObservation{
						Email: testEmail,
					},
				},
			},
			want: "serviceAccount:" + testEmail,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, ServiceAccountMemberName()(tc.mg)); diff != "" {
				t.Fatalf("Expected %s rrn differs from actual rrn, -expected +got: %s",
					name, diff)
			}
		})
	}
}

func TestServiceAccountKey_ResolveReferences(t *testing.T) {
	type args struct {
		saKey *ServiceAccountKey
		ctx   context.Context
		c     client.Reader
	}

	testSaRrn := rrnTestServiceAccount
	testClient, err := getFakeClient()

	if err != nil {
		t.Fatalf("Failed to initialize fake client: %s", err)
	}

	testCases := map[string]struct {
		args        args
		expectedErr error
		expectedSA  *string
	}{
		"NoOpServiceAccountReference": {
			args: args{
				saKey: &ServiceAccountKey{
					Spec: ServiceAccountKeySpec{
						ForProvider: ServiceAccountKeyParameters{
							ServiceAccountReferer: ServiceAccountReferer{
								ServiceAccount: &testSaRrn,
							},
						},
					},
				},
			},
			expectedSA: &testSaRrn,
		},
		"ResolveServiceAccountByName": {
			args: args{
				c: testClient,
				saKey: &ServiceAccountKey{
					Spec: ServiceAccountKeySpec{
						ForProvider: ServiceAccountKeyParameters{
							ServiceAccountReferer: ServiceAccountReferer{
								ServiceAccountRef: &xpv1.Reference{
									Name: nameTestSa,
								},
							},
						},
					},
				},
			},
			expectedSA: &testSaRrn,
		},
		"ResolveServiceAccountBySelector": {
			args: args{
				c: testClient,
				saKey: &ServiceAccountKey{
					Spec: ServiceAccountKeySpec{
						ForProvider: ServiceAccountKeyParameters{
							ServiceAccountReferer: ServiceAccountReferer{
								ServiceAccountSelector: &xpv1.Selector{
									MatchLabels: map[string]string{
										labelTest: valueTest,
									},
								},
							},
						},
					},
				},
			},
			expectedSA: &testSaRrn,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.saKey.ResolveReferences(tc.args.ctx, tc.args.c)

			if diff := cmp.Diff(tc.expectedErr, err); diff != "" {
				t.Fatalf("Expected %s error differs from actual error, -expected +got: %s",
					name, diff)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(tc.expectedSA, tc.args.saKey.Spec.ForProvider.ServiceAccount); diff != "" {
				t.Fatalf("Expected %s ServiceAccount reference differs from actual reference, -expected +got: %s",
					name, diff)
			}
		})
	}
}

func TestServiceAccountRRN(t *testing.T) {
	testCases := map[string]struct {
		mg   resource.Managed
		want string
	}{
		"NotServiceAccount": {
			mg:   &ServiceAccountKey{},
			want: "",
		},
		"ServiceAccountWithStatusName": {
			mg: &ServiceAccount{
				Status: ServiceAccountStatus{
					AtProvider: ServiceAccountObservation{
						Name: rrnTestServiceAccount,
					},
				},
			},
			want: rrnTestServiceAccount,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, ServiceAccountRRN()(tc.mg)); diff != "" {
				t.Fatalf("Expected %s rrn differs from actual rrn, -expected +got: %s",
					name, diff)
			}
		})
	}
}

func getFakeClient() (client.Client, error) {
	testSa := &ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name: nameTestSa,
			Labels: map[string]string{
				labelTest: valueTest,
			},
		},
		Status: ServiceAccountStatus{
			AtProvider: ServiceAccountObservation{
				Name: rrnTestServiceAccount,
			},
		},
	}

	scheme, err := SchemeBuilder.Build()

	if err != nil {
		return nil, err
	}

	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(testSa).Build(), nil
}
