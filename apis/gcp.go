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

// Generate deepcopy methodsets
//go:generate ${CONTROLLERGEN} object:headerFile=../hack/boilerplate.go.txt paths=./...

// Generate crossplane-runtime methodsets (resource.Managed, etc)
//go:generate ${CROSSPLANETOOLS_ANGRYJET} generate-methodsets --header-file=../hack/boilerplate.go.txt ./...

// Package apis contains Kubernetes API for GCP cloud provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	cachev1beta1 "github.com/crossplaneio/stack-gcp/apis/cache/v1beta1"
	computev1alpha2 "github.com/crossplaneio/stack-gcp/apis/compute/v1alpha2"
	databasev1beta1 "github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
	servicenetworkingv1alpha2 "github.com/crossplaneio/stack-gcp/apis/servicenetworking/v1alpha2"
	storagev1alpha2 "github.com/crossplaneio/stack-gcp/apis/storage/v1alpha2"
	gcpv1alpha2 "github.com/crossplaneio/stack-gcp/apis/v1alpha2"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		gcpv1alpha2.SchemeBuilder.AddToScheme,
		cachev1beta1.SchemeBuilder.AddToScheme,
		computev1alpha2.SchemeBuilder.AddToScheme,
		databasev1beta1.SchemeBuilder.AddToScheme,
		servicenetworkingv1alpha2.SchemeBuilder.AddToScheme,
		storagev1alpha2.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
