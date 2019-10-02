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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	googlecompute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1alpha2"
	apisv1alpha2 "github.com/crossplaneio/stack-gcp/apis/v1alpha2"
	clients "github.com/crossplaneio/stack-gcp/pkg/clients"
	"github.com/crossplaneio/stack-gcp/pkg/clients/network"
)

const (
	// Error strings.
	errNewClient                  = "cannot create new Compute Service"
	errNotNetwork                 = "managed resource is not a Network resource"
	errInsufficientNetworkSpec    = "name for network external resource is not provided"
	errProviderNotRetrieved       = "provider could not be retrieved"
	errProviderSecretNotRetrieved = "secret referred in provider could not be retrieved"

	errNetworkUpdateFailed = "update of Network resource has failed"
	errNetworkCreateFailed = "creation of Network resource has failed"
	errNetworkDeleteFailed = "deletion of Network resource has failed"

	// TEMPORARY. This should go to crossplane core repo.
	externalResourceNameAnnotationKey = "crossplane.io/external-name"
	errExternalName = "external name for the resource could not be decided"
)

// NetworkController is the controller for Network CRD.
type NetworkController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func (c *NetworkController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1alpha2.NetworkGroupVersionKind),
		resource.WithExternalConnecter(&networkConnector{kube: mgr.GetClient()}),
		resource.WithManagedConnectionPublishers())

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha2.NetworkKindAPIVersion, v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha2.Network{}).
		Complete(r)
}

type networkConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*googlecompute.Service, error)
}

func (c *networkConnector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	cr, ok := mg.(*v1alpha2.Network)
	if !ok {
		return nil, errors.New(errNotNetwork)
	}
	// TODO(muvaf): we do not yet have a way for configure the Spec with defaults for statically provisioned resources
	// such as this. Setting it directly here does not work since managed reconciler issues updates only to
	// `status` subresource. We require name to be given until we have a pre-process hook like configurator in Claim
	// reconciler
	if cr.Spec.Name == "" {
		return nil, errors.New(errInsufficientNetworkSpec)
	}

	provider := &apisv1alpha2.Provider{}
	n := meta.NamespacedNameOf(cr.Spec.ProviderReference)
	if err := c.kube.Get(ctx, n, provider); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}
	secret := &v1.Secret{}
	name := meta.NamespacedNameOf(&v1.ObjectReference{
		Name:      provider.Spec.Secret.Name,
		Namespace: provider.Namespace,
	})
	if err := c.kube.Get(ctx, name, secret); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	if c.newServiceFn == nil {
		c.newServiceFn = googlecompute.NewService
	}
	s, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(secret.Data[provider.Spec.Secret.Key]),
		option.WithScopes(googlecompute.ComputeScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &networkExternal{Service: s, projectID: provider.Spec.ProjectID}, nil
}

type networkExternal struct {
	*googlecompute.Service
	projectID string
}

func (c *networkExternal) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha2.Network)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotNetwork)
	}
	observed, err := c.Networks.Get(c.projectID, cr.Spec.Name).Context(ctx).Do()
	if clients.IsErrorNotFound(err) {
		return resource.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	if err != nil {
		return resource.ExternalObservation{}, err
	}
	cr.Status.GCPNetworkStatus = network.GenerateGCPNetworkStatus(*observed)
	// If the Network resource is retrieved, it is ready to be used
	cr.Status.SetConditions(runtimev1alpha1.Available())
	return resource.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (c *networkExternal) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha2.Network)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotNetwork)
	}
	if cr.Annotations[externalResourceNameAnnotationKey] == "" {
		if network.ValidateName(cr.Name) {
			cr.Annotations[externalResourceNameAnnotationKey] = cr.Name
		} else {
			cr.Annotations[externalResourceNameAnnotationKey] = network.GenerateName(cr.ObjectMeta)
		}
	}
	_, err := c.Networks.Insert(c.projectID, network.GenerateNetwork(*cr)).
		Context(ctx).
		Do()
	if clients.IsErrorAlreadyExists(err) {
		return resource.ExternalCreation{}, nil
	}
	cr.Status.SetConditions(runtimev1alpha1.Creating())
	return resource.ExternalCreation{}, errors.Wrap(err, errNetworkCreateFailed)
}

func (c *networkExternal) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha2.Network)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotNetwork)
	}
	if cr.Spec.IsSameAs(cr.Status.GCPNetworkStatus) {
		return resource.ExternalUpdate{}, nil
	}
	_, err := c.Networks.Patch(
		c.projectID,
		cr.Spec.Name,
		network.GenerateNetwork(cr.Spec.NetworkParameters)).
		Context(ctx).
		Do()
	return resource.ExternalUpdate{}, errors.Wrap(err, errNetworkUpdateFailed)
}

func (c *networkExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha2.Network)
	if !ok {
		return errors.New(errNotNetwork)
	}
	_, err := c.Networks.Delete(c.projectID, cr.Spec.Name).
		Context(ctx).
		Do()
	if clients.IsErrorNotFound(err) {
		return nil
	}
	cr.Status.SetConditions(runtimev1alpha1.Deleting())
	return errors.Wrap(err, errNetworkDeleteFailed)
}
