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

	v1beta12 "github.com/crossplane/provider-gcp/apis/classic/compute/v1beta1"

	gcp "github.com/crossplane/provider-gcp/internal/classic/clients"
	"github.com/crossplane/provider-gcp/internal/classic/clients/network"
	"github.com/crossplane/provider-gcp/internal/features"

	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v1"
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

	scv1alpha1 "github.com/crossplane/provider-gcp/apis/v1alpha1"
)

const (
	// Error strings.
	errNewClient            = "cannot create new Compute Service"
	errNotNetwork           = "managed resource is not a Network resource"
	errGetNetwork           = "cannot get GCP network"
	errManagedNetworkUpdate = "unable to update Network managed resource"

	errNetworkUpdateFailed  = "update of Network resource has failed"
	errNetworkCreateFailed  = "creation of Network resource has failed"
	errNetworkDeleteFailed  = "deletion of Network resource has failed"
	errCheckNetworkUpToDate = "cannot determine if GCP Network is up to date"
)

// SetupNetwork adds a controller that reconciles Network managed
// resources.
func SetupNetwork(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta12.NetworkGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta12.NetworkGroupVersionKind),
		managed.WithExternalConnecter(&networkConnector{kube: mgr.GetClient()}),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta12.Network{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type networkConnector struct {
	kube client.Client
}

func (c *networkConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := compute.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &networkExternal{Service: s, kube: c.kube, projectID: projectID}, nil
}

type networkExternal struct {
	kube client.Client
	*compute.Service
	projectID string
}

func (c *networkExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta12.Network)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotNetwork)
	}
	observed, err := c.Networks.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNetwork)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	network.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedNetworkUpdate)
		}
	}

	cr.Status.AtProvider = network.GenerateNetworkObservation(*observed)

	cr.Status.SetConditions(xpv1.Available())

	u, _, err := network.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckNetworkUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: u,
	}, nil
}

func (c *networkExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta12.Network)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNetwork)
	}

	cr.Status.SetConditions(xpv1.Creating())

	net := &compute.Network{}
	network.GenerateNetwork(meta.GetExternalName(cr), cr.Spec.ForProvider, net)
	_, err := c.Networks.Insert(c.projectID, net).
		Context(ctx).
		Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errNetworkCreateFailed)
}

func (c *networkExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta12.Network)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotNetwork)
	}

	observed, err := c.Networks.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNetwork)
	}

	upToDate, switchToCustom, err := network.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckSubnetworkUpToDate)
	}
	if upToDate {
		return managed.ExternalUpdate{}, nil
	}
	if switchToCustom {
		_, err := c.Networks.SwitchToCustomMode(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
		return managed.ExternalUpdate{}, errors.Wrap(err, errNetworkUpdateFailed)
	}

	net := &compute.Network{}
	network.GenerateNetwork(meta.GetExternalName(cr), cr.Spec.ForProvider, net)

	// NOTE(muvaf): All parameters except routing config are
	// immutable.
	_, err = c.Networks.Patch(c.projectID, meta.GetExternalName(cr), net).
		Context(ctx).
		Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errNetworkUpdateFailed)
}

func (c *networkExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta12.Network)
	if !ok {
		return errors.New(errNotNetwork)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := c.Networks.Delete(c.projectID, meta.GetExternalName(cr)).
		Context(ctx).
		Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errNetworkDeleteFailed)
}
