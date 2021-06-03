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

	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/compute/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/healthcheck"
)

const (
	// Error strings.
	errNotHealthCheck           = "managed resource is not a HealthCheck resource"
	errGetHealthCheck           = "cannot get GCP healthcheck"
	errManagedHealthCheckUpdate = "unable to update HealthCheck managed resource"

	errHealthCheckUpdateFailed  = "update of HealthCheck resource has failed"
	errHealthCheckCreateFailed  = "creation of HealthCheck resource has failed"
	errHealthCheckDeleteFailed  = "deletion of HealthCheck resource has failed"
	errCheckHealthCheckUpToDate = "cannot determine if GCP HealthCheck is up to date"
)

// SetupHealthCheck adds a controller that reconciles HealthCheck managed
// resources.
func SetupHealthCheck(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.HealthCheckGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.HealthCheck{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.HealthCheckGroupVersionKind),
			managed.WithExternalConnecter(&hcConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type hcConnector struct {
	kube client.Client
}

func (c *hcConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := compute.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &hcExternal{Service: s, kube: c.kube, projectID: projectID}, nil
}

type hcExternal struct {
	kube client.Client
	*compute.Service
	projectID string
}

func (c *hcExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.HealthCheck)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotHealthCheck)
	}
	observed, err := c.HealthChecks.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetHealthCheck)
	}
	fmt.Println(observed)

	// currentSpec := cr.Spec.ForProvider.DeepCopy()
	// router.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	// if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
	// 	if err := c.kube.Update(ctx, cr); err != nil {
	// 		return managed.ExternalObservation{}, errors.Wrap(err, errManagedHealthCheckUpdate)
	// 	}
	// }

	// cr.Status.AtProvider = router.GenerateHealthCheckObservation(*observed)

	// cr.Status.SetConditions(xpv1.Available())

	// u, _, err := router.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	// if err != nil {
	// 	return managed.ExternalObservation{}, errors.Wrap(err, errCheckHealthCheckUpToDate)
	// }

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *hcExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.HealthCheck)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotHealthCheck)
	}

	cr.Status.SetConditions(xpv1.Creating())

	hc := &compute.HealthCheck{}
	healthcheck.GenerateHealthCheck(meta.GetExternalName(cr), cr.Spec.ForProvider, hc)
	_, err := c.HealthChecks.Insert(c.projectID, hc).
		Context(ctx).
		Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errHealthCheckCreateFailed)
}

func (c *hcExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, ok := mg.(*v1alpha1.HealthCheck)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotHealthCheck)
	}

	// observed, err := c.HealthChecks.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	// if err != nil {
	// 	return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetHealthCheck)
	// }

	// upToDate, switchToCustom, err := router.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	// if err != nil {
	// 	return managed.ExternalUpdate{}, errors.Wrap(err, errCheckSubrouterUpToDate)
	// }
	// if upToDate {
	// 	return managed.ExternalUpdate{}, nil
	// }
	// if switchToCustom {
	// 	_, err := c.HealthChecks.SwitchToCustomMode(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	// 	return managed.ExternalUpdate{}, errors.Wrap(err, errHealthCheckUpdateFailed)
	// }

	// net := &compute.HealthCheck{}
	// router.GenerateHealthCheck(meta.GetExternalName(cr), cr.Spec.ForProvider, net)

	// // NOTE(muvaf): All parameters except routing config are
	// // immutable.
	// _, err = c.HealthChecks.Patch(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr), net).
	// 	Context(ctx).
	// 	Do()
	return managed.ExternalUpdate{}, errors.Wrap(nil, errHealthCheckUpdateFailed)
}

func (c *hcExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.HealthCheck)
	if !ok {
		return errors.New(errNotHealthCheck)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := c.HealthChecks.Delete(c.projectID, meta.GetExternalName(cr)).
		Context(ctx).
		Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errHealthCheckDeleteFailed)
}
