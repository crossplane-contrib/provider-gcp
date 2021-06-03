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

// Package apis contains Kubernetes API for GCP cloud provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	cachev1beta1 "github.com/crossplane/provider-gcp/apis/cache/v1beta1"
	computev1beta1 "github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	containerv1alpha1 "github.com/crossplane/provider-gcp/apis/container/v1alpha1"
	containerv1beta1 "github.com/crossplane/provider-gcp/apis/container/v1beta1"
	databasev1beta1 "github.com/crossplane/provider-gcp/apis/database/v1beta1"
	iam "github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	kms "github.com/crossplane/provider-gcp/apis/kms/v1alpha1"
	pubsub "github.com/crossplane/provider-gcp/apis/pubsub/v1alpha1"
	secretsmanager "github.com/crossplane/provider-gcp/apis/secretsmanager/secret/v1alpha1"
	secretversion "github.com/crossplane/provider-gcp/apis/secretsmanager/secretversion/v1alpha1"
	servicenetworkingv1beta1 "github.com/crossplane/provider-gcp/apis/servicenetworking/v1beta1"
	storagev1alpha1 "github.com/crossplane/provider-gcp/apis/storage/v1alpha1"
	storagev1alpha3 "github.com/crossplane/provider-gcp/apis/storage/v1alpha3"
	gcpv1alpha3 "github.com/crossplane/provider-gcp/apis/v1alpha3"
	gcpv1beta1 "github.com/crossplane/provider-gcp/apis/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		gcpv1alpha3.SchemeBuilder.AddToScheme,
		gcpv1beta1.SchemeBuilder.AddToScheme,
		cachev1beta1.SchemeBuilder.AddToScheme,
		computev1beta1.SchemeBuilder.AddToScheme,
		containerv1beta1.SchemeBuilder.AddToScheme,
		containerv1alpha1.SchemeBuilder.AddToScheme,
		databasev1beta1.SchemeBuilder.AddToScheme,
		iam.SchemeBuilder.AddToScheme,
		kms.SchemeBuilder.AddToScheme,
		pubsub.SchemeBuilder.AddToScheme,
		secretsmanager.SchemeBuilder.AddToScheme,
		secretversion.SchemeBuilder.AddToScheme,
		servicenetworkingv1beta1.SchemeBuilder.AddToScheme,
		storagev1alpha1.SchemeBuilder.AddToScheme,
		storagev1alpha3.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
