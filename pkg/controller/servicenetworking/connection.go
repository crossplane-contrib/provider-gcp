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

package servicenetworking

import (
	"context"
	"fmt"
	"path"

	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"

	compute "google.golang.org/api/compute/v1"
	servicenetworking "google.golang.org/api/servicenetworking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	xpconnection "github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gcp/apis/servicenetworking/v1beta1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	"github.com/crossplane-contrib/provider-gcp/pkg/clients/connection"
)

// Error strings.
const (
	errNewClient        = "cannot create new Compute Service"
	errNotConnection    = "managed resource is not a Connection"
	errListConnections  = "cannot list external Connection resources"
	errGetNetwork       = "cannot get VPC Network"
	errCreateConnection = "cannot create external Connection resource"
	errUpdateConnection = "cannot update external Connection resource"
	errDeleteConnection = "cannot delete external Connection resource"
)

// NOTE(negz): There is no 'Get' method for connections, only 'List', and the
// behaviour of the API is not well documented. I am assuming based on the docs
// and my observations of the API, Console, and Terraform implementation of this
// resource that:
//
// * You can only create connections for service
//   'services/servicenetworking.googleapis.com' via the API.
// * You cannot create multiple connections for service
//   'services/servicenetworking.googleapis.com' via the API.
// * Connections created via the API for service
//   'services/servicenetworking.googleapis.com' always produce a peering named
//   'servicenetworking-googleapis-com'.
//
// I note that when I create a MySQL instance with a private IP via the console
// I am prompted to create a new connection if one does not exist. This creates
// a connection for service 'services/servicenetworking.googleapis.com' with a
// peering (to a different VPC network) named 'cloudsql-mysql-googleapis-com'. I
// presume this is dark Google magic that is not exposed to API callers.
// https://cloud.google.com/service-infrastructure/docs/service-networking/reference/rest/v1/services.connections/list

// SetupConnection adds a controller that reconciles Connection
// managed resources.
func SetupConnection(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.ConnectionGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, xpconnection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind, xpconnection.WithTLSConfig(o.ESSOptions.TLSConfig)))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.ConnectionGroupVersionKind),
		managed.WithExternalConnecter(&connector{client: mgr.GetClient()}),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Connection{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	client client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetConnectionInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}

	cmp, err := compute.NewService(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	sn, err := servicenetworking.NewService(ctx, opts...)
	return &external{sn: sn, compute: cmp, projectID: projectID}, errors.Wrap(err, errNewClient)
}

type external struct {
	compute   *compute.Service
	sn        *servicenetworking.APIService
	projectID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cn, ok := mg.(*v1beta1.Connection)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotConnection)
	}
	r, err := e.sn.Services.Connections.List(cn.Spec.ForProvider.Parent).Network(gcp.StringValue(cn.Spec.ForProvider.Network)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errListConnections)
	}

	o := connection.Observation{Connection: findConnection(r.Connections)}
	if o.Connection == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if o.Network, err = e.compute.Networks.Get(e.projectID, path.Base(o.Connection.Network)).Context(ctx).Do(); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetNetwork)
	}

	eo := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: connection.IsUpToDate(cn.Spec.ForProvider, o.Connection),
	}

	connection.UpdateStatus(&cn.Status, o)

	return eo, nil
}

func findConnection(conns []*servicenetworking.Connection) *servicenetworking.Connection {
	for _, c := range conns {
		if c.Peering == connection.PeeringName {
			return c
		}
	}
	return nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cn, ok := mg.(*v1beta1.Connection)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotConnection)
	}

	cn.Status.SetConditions(xpv1.Creating())
	conn := connection.FromParameters(cn.Spec.ForProvider)
	// We use Patch instead of Create, because the resource is not created
	// if we're creating a connection in a VPC whose name had been used
	// before. It doesn't return error either, so, we just use this hack
	// found in https://github.com/terraform-providers/terraform-provider-google-beta/blob/67b258a/google-beta/resource_service_networking_connection.go#L86
	_, err := e.sn.Services.Connections.Patch(cn.Spec.ForProvider.Parent+"/connections/-", conn).UpdateMask("reservedPeeringRanges").Force(true).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(gcp.IsErrorAlreadyExists, err), errCreateConnection)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cn, ok := mg.(*v1beta1.Connection)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotConnection)
	}

	name := fmt.Sprintf("%s/connections/%s", cn.Spec.ForProvider.Parent, connection.PeeringName)
	conn := connection.FromParameters(cn.Spec.ForProvider)
	_, err := e.sn.Services.Connections.Patch(name, conn).Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateConnection)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cn, ok := mg.(*v1beta1.Connection)
	if !ok {
		return errors.New(errNotConnection)
	}

	cn.Status.SetConditions(xpv1.Deleting())
	rm := &compute.NetworksRemovePeeringRequest{Name: cn.Status.AtProvider.Peering}
	_, err := e.compute.Networks.RemovePeering(e.projectID, path.Base(gcp.StringValue(cn.Spec.ForProvider.Network)), rm).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteConnection)
}
