/*
Copyright 2022 The Crossplane Authors.

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

package dns

import (
	"context"

	"github.com/google/go-cmp/cmp"
	dns "google.golang.org/api/dns/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gcp/apis/dns/v1alpha1"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	mzclient "github.com/crossplane-contrib/provider-gcp/pkg/clients/managedzone"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
)

const (
	errNotManagedZone      = "managed resource is not of type DNS ManagedZone"
	errGetManagedZone      = "cannot get the DNS ManagedZone"
	errCreateManagedZone   = "cannot create DNS ManagedZone"
	errDeleteManagedZone   = "cannot delete DNS ManagedZone"
	errUpToDateManagedZone = "cannot determine if ManagedZone is up to date"
)

// SetupManagedZone adds a controller that reconciles the
// DNS ManagedZone resources.
func SetupManagedZone(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ManagedZoneGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ManagedZoneGroupVersionKind),
		managed.WithExternalConnecter(&managedZoneConnector{kube: mgr.GetClient()}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.ManagedZone{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type managedZoneConnector struct {
	kube client.Client
}

func (c *managedZoneConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetConnectionInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}

	d, err := dns.NewService(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &managedZoneExternal{
		kube:      c.kube,
		dns:       d.ManagedZones,
		projectID: projectID,
	}, nil
}

type managedZoneExternal struct {
	kube      client.Client
	dns       *dns.ManagedZonesService
	projectID string
}

func (e *managedZoneExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ManagedZone)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotManagedZone)
	}

	observed, err := e.dns.Get(
		e.projectID,
		meta.GetExternalName(cr),
	).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(
			resource.Ignore(gcp.IsErrorNotFound, err),
			errGetManagedZone,
		)
	}

	lateInit := false
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	mzclient.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		lateInit = true
	}

	cr.Status.AtProvider = mzclient.GenerateManagedZoneObservation(*observed)

	cr.SetConditions(xpv1.Available())

	upToDate, err := mzclient.IsUpToDate(
		meta.GetExternalName(cr),
		&cr.Spec.ForProvider,
		observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateManagedZone)
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: lateInit,
	}, nil
}

func (e *managedZoneExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ManagedZone)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotManagedZone)
	}

	args := &dns.ManagedZone{}

	mzclient.GenerateManagedZone(
		meta.GetExternalName(cr),
		cr.Spec.ForProvider,
		args,
	)

	_, err := e.dns.Create(
		e.projectID,
		args,
	).Context(ctx).Do()

	return managed.ExternalCreation{}, errors.Wrap(err, errCreateManagedZone)
}

func (e *managedZoneExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ManagedZone)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotManagedZone)
	}

	args := &dns.ManagedZone{}
	mzclient.GenerateManagedZone(
		meta.GetExternalName(cr),
		cr.Spec.ForProvider,
		args,
	)

	_, err := e.dns.Patch(
		e.projectID,
		meta.GetExternalName(cr),
		args,
	).Context(ctx).Do()

	return managed.ExternalUpdate{}, err
}

func (e *managedZoneExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ManagedZone)
	if !ok {
		return errors.New(errNotManagedZone)
	}

	cr.SetConditions(xpv1.Deleting())
	err := e.dns.Delete(
		e.projectID,
		meta.GetExternalName(cr),
	).Context(ctx).Do()

	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteManagedZone)
}
