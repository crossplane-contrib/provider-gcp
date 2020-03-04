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

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	apiv1alpha3 "github.com/crossplane/provider-gcp/apis/v1alpha3"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/network"
)

const (
	// Error strings.
	errNewClient                  = "cannot create new Compute Service"
	errNotNetwork                 = "managed resource is not a Network resource"
	errGetNetwork                 = "cannot get GCP network"
	errProviderNotRetrieved       = "provider could not be retrieved"
	errProviderSecretNotRetrieved = "secret referred in provider could not be retrieved"
	errManagedNetworkUpdate       = "unable to update Network managed resource"

	errNetworkUpdateFailed  = "update of Network resource has failed"
	errNetworkCreateFailed  = "creation of Network resource has failed"
	errNetworkDeleteFailed  = "deletion of Network resource has failed"
	errCheckNetworkUpToDate = "cannot determine if GCP Network is up to date"
)

// SetupNetwork adds a controller that reconciles Network managed
// resources.
func SetupNetwork(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.NetworkGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.Network{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.NetworkGroupVersionKind),
			managed.WithExternalConnecter(&networkConnector{kube: mgr.GetClient()}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type networkConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*compute.Service, error)
}

func (c *networkConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Network)
	if !ok {
		return nil, errors.New(errNotNetwork)
	}

	provider := &apiv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), provider); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}

	if provider.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	secret := &v1.Secret{}
	n := types.NamespacedName{Namespace: provider.Spec.CredentialsSecretRef.Namespace, Name: provider.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, secret); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	if c.newServiceFn == nil {
		c.newServiceFn = compute.NewService
	}
	s, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(secret.Data[provider.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(compute.ComputeScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &networkExternal{Service: s, kube: c.kube, projectID: provider.Spec.ProjectID}, nil
}

type networkExternal struct {
	kube client.Client
	*compute.Service
	projectID string
}

func (c *networkExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Network)
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

	cr.Status.SetConditions(runtimev1alpha1.Available())

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
	cr, ok := mg.(*v1beta1.Network)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNetwork)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	net := &compute.Network{}
	network.GenerateNetwork(meta.GetExternalName(cr), cr.Spec.ForProvider, net)
	_, err := c.Networks.Insert(c.projectID, net).
		Context(ctx).
		Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errNetworkCreateFailed)
}

func (c *networkExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Network)
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
	cr, ok := mg.(*v1beta1.Network)
	if !ok {
		return errors.New(errNotNetwork)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())
	_, err := c.Networks.Delete(c.projectID, meta.GetExternalName(cr)).
		Context(ctx).
		Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errNetworkDeleteFailed)
}
