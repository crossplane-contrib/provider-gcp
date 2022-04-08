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

	scv1alpha1 "github.com/crossplane/provider-gcp/apis/classic/v1alpha1"

	v1beta12 "github.com/crossplane/provider-gcp/apis/classic/compute/v1beta1"

	gcp "github.com/crossplane/provider-gcp/internal/classic/clients"
	"github.com/crossplane/provider-gcp/internal/classic/clients/globaladdress"
	"github.com/crossplane/provider-gcp/internal/features"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"
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
)

// Error strings.
const (
	errNotGlobalAddress           = "managed resource is not a GlobalAddress"
	errGetGlobalAddress           = "cannot get external Address resource"
	errCreateGlobalAddress        = "cannot create external Address resource"
	errDeleteGlobalAddress        = "cannot delete external Address resource"
	errManagedGlobalAddressUpdate = "cannot update managed GlobalAddress resource"
)

// SetupGlobalAddress adds a controller that reconciles
// GlobalAddress managed resources.
func SetupGlobalAddress(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta12.GlobalAddressGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta12.GlobalAddressGroupVersionKind),
		managed.WithExternalConnecter(&gaConnector{kube: mgr.GetClient()}),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta12.GlobalAddress{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type gaConnector struct {
	kube client.Client
}

func (c *gaConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := compute.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &gaExternal{kube: c.kube, Service: s, projectID: projectID}, errors.Wrap(err, errNewClient)
}

type gaExternal struct {
	kube      client.Client
	projectID string
	*compute.Service
}

func (e *gaExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta12.GlobalAddress)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGlobalAddress)
	}
	observed, err := e.GlobalAddresses.Get(e.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetGlobalAddress)
	}

	// Global addresses are always "up to date" because they can't be updated. ¯\_(ツ)_/¯
	eo := managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	globaladdress.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return eo, errors.Wrap(err, errManagedGlobalAddressUpdate)
		}
	}

	cr.Status.AtProvider = globaladdress.GenerateGlobalAddressObservation(*observed)

	switch cr.Status.AtProvider.Status {
	case v1beta12.StatusReserving:
		cr.SetConditions(xpv1.Creating())
	case v1beta12.StatusInUse, v1beta12.StatusReserved:
		cr.SetConditions(xpv1.Available())
	}

	return eo, errors.Wrap(err, errManagedGlobalAddressUpdate)
}

func (e *gaExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta12.GlobalAddress)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGlobalAddress)
	}

	cr.Status.SetConditions(xpv1.Creating())
	address := &compute.Address{}
	globaladdress.GenerateGlobalAddress(meta.GetExternalName(cr), cr.Spec.ForProvider, address)
	_, err := e.GlobalAddresses.Insert(e.projectID, address).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateGlobalAddress)
}

func (e *gaExternal) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// Global addresses cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (e *gaExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta12.GlobalAddress)
	if !ok {
		return errors.New(errNotGlobalAddress)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.GlobalAddresses.Delete(e.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteGlobalAddress)
}
