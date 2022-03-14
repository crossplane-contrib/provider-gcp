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

package compute

import (
	"context"

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

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	scv1alpha1 "github.com/crossplane/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/address"
	"github.com/crossplane/provider-gcp/pkg/features"
)

// Error strings.
const (
	errNotAddress           = "managed resource is not an Address"
	errGetAddress           = "cannot get external Address resource"
	errCreateAddress        = "cannot create external Address resource"
	errDeleteAddress        = "cannot delete external Address resource"
	errManagedAddressUpdate = "cannot update managed Address resource"
)

// SetupAddress adds a controller that reconciles Address managed resources.
func SetupAddress(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.AddressGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.AddressGroupVersionKind),
		managed.WithExternalConnecter(&addressConnector{kube: mgr.GetClient()}),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.Address{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type addressConnector struct {
	kube client.Client
}

func (c *addressConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := compute.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &addressExternal{kube: c.kube, Service: s, projectID: projectID}, errors.Wrap(err, errNewClient)
}

type addressExternal struct {
	kube      client.Client
	projectID string
	*compute.Service
}

func (e *addressExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Address)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAddress)
	}

	observed, err := e.Addresses.Get(e.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr)).Context(ctx).Do()

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetAddress)
	}

	//  addresses are always "up to date" because they can't be updated. ¯\_(ツ)_/¯
	eo := managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	address.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	eo.ResourceLateInitialized = !cmp.Equal(currentSpec, &cr.Spec.ForProvider)

	cr.Status.AtProvider = address.GenerateAddressObservation(*observed)

	switch cr.Status.AtProvider.Status {
	case v1beta1.StatusReserving:
		cr.SetConditions(xpv1.Creating())
	case v1beta1.StatusInUse, v1beta1.StatusReserved:
		cr.SetConditions(xpv1.Available())
	}

	return eo, errors.Wrap(err, errManagedAddressUpdate)
}

func (e *addressExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Address)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAddress)
	}

	addr := &compute.Address{}
	address.GenerateAddress(meta.GetExternalName(cr), cr.Spec.ForProvider, addr)
	_, err := e.Addresses.Insert(e.projectID, addr.Region, addr).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateAddress)
}

func (e *addressExternal) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	//  addresses cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (e *addressExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Address)
	if !ok {
		return errors.New(errNotAddress)
	}

	_, err := e.Addresses.Delete(e.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr)).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteAddress)
}
