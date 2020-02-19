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
	computev1alpha1 "github.com/crossplane/crossplane/apis/compute/v1alpha1"

	"github.com/crossplane/stack-gcp/apis/compute/v1alpha3"
)

// SetupGKEClusterClaimScheduling adds a controller that reconciles
// KubernetesCluster claims that include a class selector but omit their class
// and resource references by picking a random matching GKEClusterClass, if any.
func SetupGKEClusterClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(computev1alpha1.KubernetesClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&computev1alpha1.KubernetesCluster{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(computev1alpha1.KubernetesClusterGroupVersionKind),
			resource.ClassKind(v1alpha3.GKEClusterClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupGKEClusterClaimDefaulting adds a controller that reconciles
// KubernetesCluster claims that omit their resource ref, class ref, and class
// selector by choosing a default GKEClusterClass if one exists.
func SetupGKEClusterClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(computev1alpha1.KubernetesClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&computev1alpha1.KubernetesCluster{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(computev1alpha1.KubernetesClusterGroupVersionKind),
			resource.ClassKind(v1alpha3.GKEClusterClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupGKEClusterClaimBinding adds a controller that reconciles
// KubernetesCluster claims with GKEClusters, dynamically provisioning them if
// needed.
func SetupGKEClusterClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(computev1alpha1.KubernetesClusterGroupKind)

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha3.GKEClusterClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha3.GKEClusterGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha3.GKEClusterGroupVersionKind), mgr.GetScheme()),
	))

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(computev1alpha1.KubernetesClusterGroupVersionKind),
		resource.ClassKind(v1alpha3.GKEClusterClassGroupVersionKind),
		resource.ManagedKind(v1alpha3.GKEClusterGroupVersionKind),
		claimbinding.WithBinder(claimbinding.NewAPIBinder(mgr.GetClient(), mgr.GetScheme())),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigureGKECluster),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames),
		),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha3.GKECluster{}}, &resource.EnqueueRequestForClaim{}).
		For(&computev1alpha1.KubernetesCluster{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureGKECluster configures the supplied resource (presumed to be a
// GKECluster) using the supplied resource claim (presumed to be a
// KubernetesCluster) and resource class.
func ConfigureGKECluster(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	if _, cmok := cm.(*computev1alpha1.KubernetesCluster); !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), computev1alpha1.KubernetesClusterGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha3.GKEClusterClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha3.GKEClusterClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha3.GKECluster)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha3.GKEClusterGroupVersionKind)
	}

	spec := &v1alpha3.GKEClusterSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: v1alpha3.DefaultReclaimPolicy,
		},
		GKEClusterParameters: rs.SpecTemplate.GKEClusterParameters,
	}

	// NOTE(hasheddan): consider moving defaulting to either CRD or managed reconciler level
	if spec.Labels == nil {
		spec.Labels = map[string]string{}
	}
	if spec.NumNodes == 0 {
		spec.NumNodes = v1alpha3.DefaultNumberOfNodes
	}
	if spec.Scopes == nil {
		spec.Scopes = []string{}
	}

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	i.Spec = *spec

	return nil
}
