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
	googlecompute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/subnetwork"
)

const (
	// Error strings.
	errNotSubnetwork           = "managed resource is not a Subnetwork resource"
	errManagedSubnetworkUpdate = "unable to update Subnetwork managed resource"

	errGetSubnetwork            = "unable to get GCP Subnetwork"
	errUpdateSubnetworkFailed   = "update of GCP Subnetwork has failed"
	errUpdateSubnetworkPAFailed = "unable to update GCP Subnetwork Private IP Google Access"
	errCreateSubnetworkFailed   = "creation of GCP Subnetwork resource has failed"
	errDeleteSubnetworkFailed   = "deletion of GCP Subnetwork resource has failed"
	errCheckSubnetworkUpToDate  = "cannot determine if GCP Subnetwork is up to date"
)

// SetupSubnetwork adds a controller that reconciles Subnetwork
// managed resources.
func SetupSubnetwork(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.SubnetworkGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.Subnetwork{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.SubnetworkGroupVersionKind),
			managed.WithExternalConnecter(&subnetworkConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type subnetworkConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*googlecompute.Service, error)
}

func (c *subnetworkConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := c.newServiceFn(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &subnetworkExternal{Service: s, kube: c.kube, projectID: projectID}, nil
}

type subnetworkExternal struct {
	kube client.Client
	*googlecompute.Service
	projectID string
}

func (c *subnetworkExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Subnetwork)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubnetwork)
	}
	observed, err := c.Subnetworks.Get(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetSubnetwork)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	subnetwork.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedSubnetworkUpdate)
		}
	}

	cr.Status.AtProvider = subnetwork.GenerateSubnetworkObservation(*observed)

	u, _, err := subnetwork.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckSubnetworkUpToDate)
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: u,
	}, nil
}

func (c *subnetworkExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Subnetwork)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubnetwork)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	subnet := &googlecompute.Subnetwork{}
	subnetwork.GenerateSubnetwork(meta.GetExternalName(cr), cr.Spec.ForProvider, subnet)
	_, err := c.Subnetworks.Insert(c.projectID, cr.Spec.ForProvider.Region, subnet).
		Context(ctx).
		Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateSubnetworkFailed)
}

func (c *subnetworkExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Subnetwork)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubnetwork)
	}

	observed, err := c.Subnetworks.Get(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetSubnetwork)
	}

	upToDate, privateAccess, err := subnetwork.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckSubnetworkUpToDate)
	}
	if upToDate {
		return managed.ExternalUpdate{}, nil
	}
	if privateAccess {
		update := &googlecompute.SubnetworksSetPrivateIpGoogleAccessRequest{PrivateIpGoogleAccess: *cr.Spec.ForProvider.PrivateIPGoogleAccess}
		_, err = c.Subnetworks.SetPrivateIpGoogleAccess(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr), update).Context(ctx).Do()
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSubnetworkPAFailed)
	}

	subnetUpdate := subnetwork.GenerateSubnetworkForUpdate(*cr, meta.GetExternalName(cr))
	_, err = c.Subnetworks.Patch(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr), subnetUpdate).Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSubnetworkFailed)
}

func (c *subnetworkExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Subnetwork)
	if !ok {
		return errors.New(errNotSubnetwork)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())
	_, err := c.Subnetworks.Delete(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr)).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteSubnetworkFailed)
}
