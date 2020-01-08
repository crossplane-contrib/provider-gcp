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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	storagev1alpha1 "github.com/crossplaneio/crossplane/apis/storage/v1alpha1"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
)

// A BucketClaimSchedulingController reconciles Bucket claims that include a
// class selector but omit their class and resource references by picking a
// random matching GCS BucketClass, if any.
type BucketClaimSchedulingController struct{}

// SetupWithManager sets up the BucketClaimSchedulingController using the
// supplied manager.
func (c *BucketClaimSchedulingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.BucketKind,
		v1alpha3.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimSchedulingReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.BucketClassGroupVersionKind),
		))
}

// A BucketClaimDefaultingController reconciles Bucket claims that omit their
// resource ref, class ref, and class selector by choosing a default GCS
// BucketClass if one exists.
type BucketClaimDefaultingController struct{}

// SetupWithManager sets up the BucketClaimDefaultingController using the
// supplied manager.
func (c *BucketClaimDefaultingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.BucketKind,
		v1alpha3.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimDefaultingReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.BucketClassGroupVersionKind),
		))
}

// A BucketClaimController reconciles Bucket claims with GCS Buckets,
// dynamically provisioning them if needed.
type BucketClaimController struct{}

// SetupWithManager adds a controller that reconciles Bucket resource claims.
func (c *BucketClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.BucketKind,
		v1alpha3.Group))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha3.BucketClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha3.BucketGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha3.BucketGroupVersionKind), mgr.GetScheme()),
	))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKind(v1alpha3.BucketClassGroupVersionKind),
		resource.ManagedKind(v1alpha3.BucketGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureBucket),
			resource.ManagedConfiguratorFn(resource.ConfigureReclaimPolicy),
			resource.ManagedConfiguratorFn(resource.ConfigureNames),
		))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha3.Bucket{}}, &resource.EnqueueRequestForClaim{}).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureBucket configures the supplied resource (presumed
// to be a Bucket) using the supplied resource claim (presumed
// to be a Bucket) and resource class.
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
		ForProvider: rs.SpecTemplate.ForProvider,
	}
	// TODO(muvaf): it should be safe to use managed resource name as the name
	// of the secret instead of claim's UID.
	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(bcm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	bmg.Spec = *spec

	return nil
}
