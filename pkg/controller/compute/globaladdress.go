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

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/event"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1beta1"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	"github.com/crossplaneio/stack-gcp/pkg/clients/globaladdress"
)

// Error strings.
const (
	errNotGlobalAddress     = "managed resource is not a GlobalAddress"
	errProviderSecretNil    = "cannot find Secret reference on Provider"
	errGetAddress           = "cannot get external Address resource"
	errCreateAddress        = "cannot create external Address resource"
	errDeleteAddress        = "cannot delete external Address resource"
	errManagedAddressUpdate = "cannot update managed GlobalAddress resource"
)

// SetupGlobalAddress adds a controller that reconciles
// GlobalAddress managed resources.
func SetupGlobalAddress(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.GlobalAddressGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.GlobalAddress{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.GlobalAddressGroupVersionKind),
			managed.WithExternalConnecter(&gaConnector{kube: mgr.GetClient(), newServiceFn: compute.NewService}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type gaConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*compute.Service, error)
}

func (c *gaConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.GlobalAddress)
	if !ok {
		return nil, errors.New(errNotGlobalAddress)
	}

	p := &gcpv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &v1.Secret{}
	if err := c.kube.Get(ctx, types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}, s); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	svc, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(s.Data[p.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(compute.ComputeScope))
	return &gaExternal{kube: c.kube, Service: svc, projectID: p.Spec.ProjectID}, errors.Wrap(err, errNewClient)
}

type gaExternal struct {
	kube      client.Client
	projectID string
	*compute.Service
}

func (e *gaExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.GlobalAddress)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGlobalAddress)
	}
	observed, err := e.GlobalAddresses.Get(e.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetAddress)
	}

	// Global addresses are always "up to date" because they can't be updated. ¯\_(ツ)_/¯
	eo := managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	globaladdress.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return eo, errors.Wrap(err, errManagedAddressUpdate)
		}
	}

	cr.Status.AtProvider = globaladdress.GenerateGlobalAddressObservation(*observed)

	switch cr.Status.AtProvider.Status {
	case v1beta1.StatusReserving:
		cr.SetConditions(runtimev1alpha1.Creating())
	case v1beta1.StatusInUse, v1beta1.StatusReserved:
		cr.SetConditions(runtimev1alpha1.Available())
	}

	return eo, errors.Wrap(err, errManagedAddressUpdate)
}

func (e *gaExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.GlobalAddress)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGlobalAddress)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())
	address := &compute.Address{}
	globaladdress.GenerateGlobalAddress(meta.GetExternalName(cr), cr.Spec.ForProvider, address)
	_, err := e.GlobalAddresses.Insert(e.projectID, address).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateAddress)
}

func (e *gaExternal) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// Global addresses cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (e *gaExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.GlobalAddress)
	if !ok {
		return errors.New(errNotGlobalAddress)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.GlobalAddresses.Delete(e.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteAddress)
}
