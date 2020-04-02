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

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimdefaulting"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimscheduling"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	storagev1alpha1 "github.com/crossplane/crossplane/apis/storage/v1alpha1"

	"github.com/crossplane/provider-gcp/apis/storage/v1alpha3"
)

// SetupBucketClaimScheduling adds a controller that reconciles Bucket claims
// that include a class selector but omit their class and resource references by
// picking a random matching GCS BucketClass, if any.
func SetupBucketClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(storagev1alpha1.BucketGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.BucketClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupBucketClaimDefaulting adds a controller that reconciles Bucket claims
// that omit their resource ref, class ref, and class selector by choosing a
// default GCS BucketClass if one exists.
func SetupBucketClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(storagev1alpha1.BucketGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.BucketClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupBucketClaimBinding adds a controller that reconciles Bucket claims with
// GCS Buckets, dynamically provisioning them if needed.
func SetupBucketClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(storagev1alpha1.BucketGroupKind)

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha3.BucketClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha3.BucketGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha3.BucketGroupVersionKind), mgr.GetScheme()),
	))

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKind(v1alpha3.BucketClassGroupVersionKind),
		resource.ManagedKind(v1alpha3.BucketGroupVersionKind),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigureBucket),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames)),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha3.Bucket{}}, &resource.EnqueueRequestForClaim{}).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureBucket configures the supplied resource (presumed to be a Bucket)
// using the supplied resource claim (presumed to be a Bucket) and resource
// class.
func ConfigureBucket(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	bcm, cmok := cm.(*storagev1alpha1.Bucket)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), storagev1alpha1.BucketGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha3.BucketClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha3.BucketClassGroupVersionKind)
	}

	bmg, mgok := mg.(*v1alpha3.Bucket)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha3.BucketGroupVersionKind)
	}

	spec := &v1alpha3.BucketSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		BucketParameters: rs.SpecTemplate.BucketParameters,
	}

	// Set PredefinedACL from bucketClaim claim only iff: claim has this value and
	// it is not defined in the resource class (i.e. not already in the spec)
	if bcm.Spec.PredefinedACL != nil && spec.PredefinedACL == "" {
		spec.PredefinedACL = string(*bcm.Spec.PredefinedACL)
	}

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	bmg.Spec = *spec

	return nil
}
