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

package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/claimdefaulting"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/claimscheduling"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	databasev1alpha1 "github.com/crossplaneio/crossplane/apis/database/v1alpha1"

	"github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
)

// SetupPostgreSQLInstanceClaimSchedulingController adds a controller that
// reconciles PostgreSQLInstance claims that include a class selector but omit
// their class and resource references by picking a random matching
// CloudSQLInstanceClass, if any.
func SetupPostgreSQLInstanceClaimSchedulingController(mgr ctrl.Manager, _ logging.Logger) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		databasev1alpha1.PostgreSQLInstanceKind,
		v1beta1.CloudSQLInstanceKind,
		v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind),
		))
}

// SetupPostgreSQLInstanceClaimDefaultingController adds a controller that
// reconciles PostgreSQLInstance claims that omit their resource ref, class ref,
// and class selector by choosing a default CloudSQLInstanceClass if one exists.
func SetupPostgreSQLInstanceClaimDefaultingController(mgr ctrl.Manager, _ logging.Logger) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		databasev1alpha1.PostgreSQLInstanceKind,
		v1beta1.CloudSQLInstanceKind,
		v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind),
		))
}

// SetupPostgreSQLInstanceClaimController adds a controller that reconciles
// PostgreSQLInstance claims with CloudSQLInstances, dynamically provisioning
// them if needed.
func SetupPostgreSQLInstanceClaimController(mgr ctrl.Manager, _ logging.Logger) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		databasev1alpha1.PostgreSQLInstanceKind,
		v1beta1.CloudSQLInstanceKind,
		v1beta1.Group))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind), mgr.GetScheme()),
	))

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
		resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind),
		resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigurePostgreSQLCloudSQLInstance),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames),
		))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1beta1.CloudSQLInstance{}}, &resource.EnqueueRequestForClaim{}).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigurePostgreSQLCloudSQLInstance configures the supplied instance (presumed
// to be a CloudSQLInstance) using the supplied instance claim (presumed to be a
// PostgreSQLInstance) and instance class.
func ConfigurePostgreSQLCloudSQLInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	pg, cmok := cm.(*databasev1alpha1.PostgreSQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.PostgreSQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1beta1.CloudSQLInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1beta1.CloudSQLInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1beta1.CloudSQLInstance)
	if !mgok {
		return errors.Errorf("expected managed instance %s to be %s", mg.GetName(), v1beta1.CloudSQLInstanceGroupVersionKind)
	}

	spec := &v1beta1.CloudSQLInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}

	if pg.Spec.EngineVersion != "" {
		spec.ForProvider.DatabaseVersion = translateVersion(pg.Spec.EngineVersion, v1beta1.PostgresqlDBVersionPrefix)
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

// SetupMySQLInstanceClaimSchedulingController adds a controller that reconciles
// MySQLInstance claims that include a class selector but omit their class and
// resource references by picking a random matching CloudSQLInstanceClass, if
// any.
func SetupMySQLInstanceClaimSchedulingController(mgr ctrl.Manager, _ logging.Logger) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1beta1.CloudSQLInstanceKind,
		v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind),
		))
}

// SetupMySQLInstanceClaimDefaultingController adds a controller that reconciles
// MySQLInstance claims that omit their resource ref, class ref, and class
// selector by choosing a default CloudSQLInstanceClass if one exists.
func SetupMySQLInstanceClaimDefaultingController(mgr ctrl.Manager, _ logging.Logger) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1beta1.CloudSQLInstanceKind,
		v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind),
		))
}

// SetupMySQLInstanceClaimController adds a controller that reconciles
// MySQLInstance claims with CloudSQLInstances, dynamically provisioning them if
// needed.
func SetupMySQLInstanceClaimController(mgr ctrl.Manager, _ logging.Logger) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1beta1.CloudSQLInstanceKind,
		v1beta1.Group))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind), mgr.GetScheme()),
	))

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
		resource.ClassKind(v1beta1.CloudSQLInstanceClassGroupVersionKind),
		resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigureMyCloudSQLInstance),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames),
		))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1beta1.CloudSQLInstance{}}, &resource.EnqueueRequestForClaim{}).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureMyCloudSQLInstance configures the supplied instance (presumed to be
// a CloudSQLInstance) using the supplied instance claim (presumed to be a
// MySQLInstance) and instance class.
func ConfigureMyCloudSQLInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	my, cmok := cm.(*databasev1alpha1.MySQLInstance)
	if !cmok {
		return errors.Errorf("expected instance claim %s to be %s", cm.GetName(), databasev1alpha1.MySQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1beta1.CloudSQLInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1beta1.CloudSQLInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1beta1.CloudSQLInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1beta1.CloudSQLInstanceGroupVersionKind)
	}

	spec := &v1beta1.CloudSQLInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}

	if my.Spec.EngineVersion != "" {
		spec.ForProvider.DatabaseVersion = translateVersion(my.Spec.EngineVersion, v1beta1.MysqlDBVersionPrefix)
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

func translateVersion(version, versionPrefix string) *string {
	if version == "" {
		return nil
	}
	r := fmt.Sprintf("%s_%s", versionPrefix, strings.Replace(version, ".", "_", -1))
	return &r
}
