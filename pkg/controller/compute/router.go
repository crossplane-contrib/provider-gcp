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
	"github.com/crossplane/provider-gcp/pkg/clients/router"
)

const (
	// Error strings.
	errNotRouter           = "managed resource is not a Router resource"
	errGetRouter           = "cannot get GCP router"
	errManagedRouterUpdate = "unable to update Router managed resource"

	errRouterUpdateFailed  = "update of Router resource has failed"
	errRouterCreateFailed  = "creation of Router resource has failed"
	errRouterDeleteFailed  = "deletion of Router resource has failed"
	errCheckRouterUpToDate = "cannot determine if GCP Router is up to date"
)

// SetupRouter adds a controller that reconciles Router managed
// resources.
func SetupRouter(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.RouterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Router{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.RouterGroupVersionKind),
			managed.WithExternalConnecter(&routerConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type routerConnector struct {
	kube client.Client
}

func (c *routerConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := compute.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &routerExternal{Service: s, kube: c.kube, projectID: projectID}, nil
}

type routerExternal struct {
	kube client.Client
	*compute.Service
	projectID string
}

func (c *routerExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Router)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRouter)
	}
	observed, err := c.Routers.Get(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetRouter)
	}
	fmt.Println(observed)

	// currentSpec := cr.Spec.ForProvider.DeepCopy()
	// router.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	// if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
	// 	if err := c.kube.Update(ctx, cr); err != nil {
	// 		return managed.ExternalObservation{}, errors.Wrap(err, errManagedRouterUpdate)
	// 	}
	// }

	// cr.Status.AtProvider = router.GenerateRouterObservation(*observed)

	// cr.Status.SetConditions(xpv1.Available())

	// u, _, err := router.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	// if err != nil {
	// 	return managed.ExternalObservation{}, errors.Wrap(err, errCheckRouterUpToDate)
	// }

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *routerExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Router)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRouter)
	}

	cr.Status.SetConditions(xpv1.Creating())

	net := &compute.Router{}
	router.GenerateRouter(meta.GetExternalName(cr), cr.Spec.ForProvider, net)
	_, err := c.Routers.Insert(c.projectID, cr.Spec.ForProvider.Region, net).
		Context(ctx).
		Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errRouterCreateFailed)
}

func (c *routerExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, ok := mg.(*v1alpha1.Router)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRouter)
	}

	// observed, err := c.Routers.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	// if err != nil {
	// 	return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetRouter)
	// }

	// upToDate, switchToCustom, err := router.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, observed)
	// if err != nil {
	// 	return managed.ExternalUpdate{}, errors.Wrap(err, errCheckSubrouterUpToDate)
	// }
	// if upToDate {
	// 	return managed.ExternalUpdate{}, nil
	// }
	// if switchToCustom {
	// 	_, err := c.Routers.SwitchToCustomMode(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	// 	return managed.ExternalUpdate{}, errors.Wrap(err, errRouterUpdateFailed)
	// }

	// net := &compute.Router{}
	// router.GenerateRouter(meta.GetExternalName(cr), cr.Spec.ForProvider, net)

	// // NOTE(muvaf): All parameters except routing config are
	// // immutable.
	// _, err = c.Routers.Patch(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr), net).
	// 	Context(ctx).
	// 	Do()
	return managed.ExternalUpdate{}, errors.Wrap(nil, errRouterUpdateFailed)
}

func (c *routerExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Router)
	if !ok {
		return errors.New(errNotRouter)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := c.Routers.Delete(c.projectID, cr.Spec.ForProvider.Region, meta.GetExternalName(cr)).
		Context(ctx).
		Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errRouterDeleteFailed)
}
