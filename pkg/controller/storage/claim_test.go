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

package storage

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	storagev1alpha1 "github.com/crossplane/crossplane/apis/storage/v1alpha1"

	"github.com/crossplane/stack-gcp/apis/storage/v1alpha3"
)

var _ claimbinding.ManagedConfigurator = claimbinding.ManagedConfiguratorFn(ConfigureBucket)

func TestConfigureBucket(t *testing.T) {
	type args struct {
		ctx context.Context
		cm  resource.Claim
		cs  resource.Class
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	namespace := "ns"
	claimUID := types.UID("definitely-a-uuid")
	providerName := "coolprovider"
	bucketName := "coolbucket"
	bucketPrivate := storagev1alpha1.ACLPrivate

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				cm: &storagev1alpha1.Bucket{
					ObjectMeta: metav1.ObjectMeta{UID: claimUID},
					Spec: storagev1alpha1.BucketSpec{
						Name:          bucketName,
						PredefinedACL: &bucketPrivate,
					},
				},
				cs: &v1alpha3.BucketClass{
					SpecTemplate: v1alpha3.BucketClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: namespace,
							ProviderReference:                 &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:                     runtimev1alpha1.ReclaimDelete,
						},
					},
				},
				mg: &v1alpha3.Bucket{},
			},
			want: want{
				mg: &v1alpha3.Bucket{
					Spec: v1alpha3.BucketSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
								Namespace: namespace,
								Name:      string(claimUID),
							},
							ProviderReference: &corev1.ObjectReference{Name: providerName},
						},
						BucketParameters: v1alpha3.BucketParameters{
							NameFormat: bucketName,
							BucketSpecAttrs: v1alpha3.BucketSpecAttrs{
								BucketUpdatableAttrs: v1alpha3.BucketUpdatableAttrs{
									PredefinedACL: string(bucketPrivate),
								},
							},
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ConfigureBucket(tc.args.ctx, tc.args.cm, tc.args.cs, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("ConfigureBucket(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("ConfigureBucket(...) Managed: -want, +got:\n%s", diff)
			}
		})
	}
}
