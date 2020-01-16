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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1alpha3"
	gcpapis "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	clients "github.com/crossplaneio/stack-gcp/pkg/clients"
	"github.com/crossplaneio/stack-gcp/pkg/clients/subnetwork"
)

const (
	// Error strings.
	errNotSubnetwork              = "managed resource is not a Subnetwork resource"
	errInsufficientSubnetworkSpec = "name or region for network external resource is not provided"

	errUpdateSubnetworkFailed = "update of Subnetwork resource has failed"
	errCreateSubnetworkFailed = "creation of Subnetwork resource has failed"
	errDeleteSubnetworkFailed = "deletion of Subnetwork resource has failed"
)

// SetupSubnetworkController adds a controller that reconciles Subnetwork
// managed resources.
func SetupSubnetworkController(mgr ctrl.Manager, l logging.Logger) error {
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha3.SubnetworkGroupVersionKind),
		managed.WithExternalConnecter(&subnetworkConnector{kube: mgr.GetClient()}),
		managed.WithConnectionPublishers(),
		managed.WithLogger(l))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha3.SubnetworkKindAPIVersion, v1alpha3.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.Subnetwork{}).
		Complete(r)
}

type subnetworkConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*googlecompute.Service, error)
}

func (c *subnetworkConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha3.Subnetwork)
	if !ok {
		return nil, errors.New(errNotSubnetwork)
	}
	// TODO(muvaf): we do not yet have a way for configure the Spec with defaults for statically provisioned resources
	// such as this. Setting it directly here does not work since managed reconciler issues updates only to
	// `status` subresource. We require name to be given until we have a pre-process hook like configurator in Claim
	// reconciler
	if cr.Spec.Name == "" || cr.Spec.Region == "" {
		return nil, errors.New(errInsufficientSubnetworkSpec)
	}

	provider := &gcpapis.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), provider); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}
	secret := &v1.Secret{}
	n := types.NamespacedName{Namespace: provider.Spec.CredentialsSecretRef.Namespace, Name: provider.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, secret); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	if c.newServiceFn == nil {
		c.newServiceFn = googlecompute.NewService
	}
	s, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(secret.Data[provider.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(googlecompute.ComputeScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &subnetworkExternal{Service: s, projectID: provider.Spec.ProjectID}, nil
}

type subnetworkExternal struct {
	*googlecompute.Service
	projectID string
}

func (c *subnetworkExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha3.Subnetwork)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubnetwork)
	}
	observed, err := c.Subnetworks.Get(c.projectID, cr.Spec.Region, cr.Spec.Name).Context(ctx).Do()
	if clients.IsErrorNotFound(err) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.Status.GCPSubnetworkStatus = subnetwork.GenerateGCPSubnetworkStatus(observed)
	cr.Status.SetConditions(runtimev1alpha1.Available())
	return managed.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (c *subnetworkExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha3.Subnetwork)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubnetwork)
	}
	_, err := c.Subnetworks.Insert(c.projectID, cr.Spec.Region, subnetwork.GenerateSubnetwork(cr.Spec.SubnetworkParameters)).
		Context(ctx).
		Do()
	if clients.IsErrorAlreadyExists(err) {
		return managed.ExternalCreation{}, nil
	}
	cr.Status.SetConditions(runtimev1alpha1.Creating())
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateSubnetworkFailed)
}

func (c *subnetworkExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha3.Subnetwork)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubnetwork)
	}
	if cr.Spec.IsSameAs(cr.Status.GCPSubnetworkStatus) {
		return managed.ExternalUpdate{}, nil
	}
	subnetworkBody := subnetwork.GenerateSubnetwork(cr.Spec.SubnetworkParameters)
	// Fingerprint from the last GET is required for updates.
	subnetworkBody.Fingerprint = cr.Status.Fingerprint
	// The API rejects region and network to be updated, in fact, it rejects the update when this field is even included. Calm down.
	subnetworkBody.Region = ""
	subnetworkBody.Network = ""
	_, err := c.Subnetworks.Patch(c.projectID, cr.Spec.Region, cr.Spec.Name, subnetworkBody).Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSubnetworkFailed)
}

func (c *subnetworkExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha3.Subnetwork)
	if !ok {
		return errors.New(errNotSubnetwork)
	}
	_, err := c.Subnetworks.Delete(c.projectID, cr.Spec.Region, cr.Spec.Name).Context(ctx).Do()
	if clients.IsErrorNotFound(err) {
		return nil
	}
	cr.Status.SetConditions(runtimev1alpha1.Deleting())
	return errors.Wrap(err, errDeleteSubnetworkFailed)
}
