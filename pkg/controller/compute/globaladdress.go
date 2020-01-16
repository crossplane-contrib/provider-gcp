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
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1alpha3"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	"github.com/crossplaneio/stack-gcp/pkg/clients/globaladdress"
)

// Error strings.
const (
	errNotGlobalAddress = "managed resource is not a GlobalAddress"
	errGetAddress       = "cannot get external Address resource"
	errCreateAddress    = "cannot create external Address resource"
	errDeleteAddress    = "cannot delete external Address resource"
	errUpdateManaged    = "cannot update managed resource"
)

// GlobalAddressController is the controller for GlobalAddress CRD.
type GlobalAddressController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func (c *GlobalAddressController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(strings.ToLower(fmt.Sprintf("%s.%s", v1alpha3.GlobalAddressKindAPIVersion, v1alpha3.Group))).
		For(&v1alpha3.GlobalAddress{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.GlobalAddressGroupVersionKind),
			managed.WithExternalConnecter(&gaConnector{client: mgr.GetClient(), newCompute: compute.NewService}),
			managed.WithConnectionPublishers()))
}

type gaConnector struct {
	client     client.Client
	newCompute func(ctx context.Context, opts ...option.ClientOption) (*compute.Service, error)
}

func (c *gaConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	ga, ok := mg.(*v1alpha3.GlobalAddress)
	if !ok {
		return nil, errors.New(errNotGlobalAddress)
	}

	p := &gcpv1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(ga.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}
	s := &v1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}, s); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	cmp, err := c.newCompute(ctx,
		option.WithCredentialsJSON(s.Data[p.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(compute.ComputeScope))
	return &gaExternal{client: c.client, compute: cmp, projectID: p.Spec.ProjectID}, errors.Wrap(err, errNewClient)
}

type gaExternal struct {
	client    client.Client
	compute   *compute.Service
	projectID string
}

func (e *gaExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	ga, ok := mg.(*v1alpha3.GlobalAddress)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGlobalAddress)
	}
	observed, err := e.compute.GlobalAddresses.Get(e.projectID, ga.Spec.Name).Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetAddress)
	}

	// Global addresses are always "up to date" because they can't be updated. ¯\_(ツ)_/¯
	eo := managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}

	// NOTE(negz): We must update our parameters before our status to avoid
	// client.Update overwriting our newly updated status with that most
	// recently persisted to the API server.
	globaladdress.UpdateParameters(&ga.Spec.GlobalAddressParameters, observed)
	err = e.client.Update(ctx, ga)

	globaladdress.UpdateStatus(&ga.Status, observed)

	return eo, errors.Wrap(err, errUpdateManaged)
}

func (e *gaExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	ga, ok := mg.(*v1alpha3.GlobalAddress)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGlobalAddress)
	}

	ga.Status.SetConditions(runtimev1alpha1.Creating())
	address := globaladdress.FromParameters(ga.Spec.GlobalAddressParameters)
	_, err := e.compute.GlobalAddresses.Insert(e.projectID, address).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(gcp.IsErrorAlreadyExists, err), errCreateAddress)
}

func (e *gaExternal) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// Global addresses cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (e *gaExternal) Delete(ctx context.Context, mg resource.Managed) error {
	ga, ok := mg.(*v1alpha3.GlobalAddress)
	if !ok {
		return errors.New(errNotGlobalAddress)
	}

	ga.Status.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.compute.GlobalAddresses.Delete(e.projectID, ga.Spec.Name).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteAddress)
}
