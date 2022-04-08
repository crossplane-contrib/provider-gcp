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

package container

import (
	"context"

	v1beta12 "github.com/crossplane/provider-gcp/apis/classic/container/v1beta1"

	gcp "github.com/crossplane/provider-gcp/internal/classic/clients"
	np "github.com/crossplane/provider-gcp/internal/classic/clients/nodepool"
	"github.com/crossplane/provider-gcp/internal/features"

	"github.com/google/go-cmp/cmp"
	container "google.golang.org/api/container/v1"
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

// Error strings.
const (
	errManagedNodePoolUpdateFailed = "cannot update NodePool custom resource"
	errNotNodePool                 = "managed resource is not a NodePool"
	errGetNodePool                 = "cannot get GKE node pool"
	errCreateNodePool              = "cannot create GKE node pool"
	errUpdateNodePool              = "cannot update GKE node pool"
	errDeleteNodePool              = "cannot delete GKE node pool"
	errCheckNodePoolUpToDate       = "cannot determine if GKE node pool is up to date"
)

// SetupNodePool adds a controller that reconciles NodePool managed
// resources.
func SetupNodePool(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta12.NodePoolGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta12.NodePoolGroupVersionKind),
		managed.WithExternalConnecter(&nodePoolConnector{kube: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta12.NodePool{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type nodePoolConnector struct {
	kube client.Client
}

func (c *nodePoolConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := container.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &nodePoolExternal{container: s, projectID: projectID, kube: c.kube}, errors.Wrap(err, errNewClient)
}

type nodePoolExternal struct {
	kube      client.Client
	container *container.Service
	projectID string
}

func (e *nodePoolExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta12.NodePool)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotNodePool)
	}

	existing, err := e.container.Projects.Locations.Clusters.NodePools.Get(np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNodePool)
	}

	cr.Status.AtProvider = np.GenerateObservation(*existing)
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	np.LateInitializeSpec(&cr.Spec.ForProvider, *existing)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedNodePoolUpdateFailed)
		}
	}

	switch cr.Status.AtProvider.Status {
	case v1beta12.NodePoolStateRunning, v1beta12.NodePoolStateReconciling:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta12.NodePoolStateProvisioning:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta12.NodePoolStateUnspecified, v1beta12.NodePoolStateRunningError, v1beta12.NodePoolStateError:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	u, _, err := np.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, existing)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckNodePoolUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: u,
	}, nil
}

func (e *nodePoolExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta12.NodePool)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNodePool)
	}
	cr.SetConditions(xpv1.Creating())

	// Wait until creation is complete if already provisioning.
	if cr.Status.AtProvider.Status == v1beta12.NodePoolStateProvisioning {
		return managed.ExternalCreation{}, nil
	}

	// Generate GKE node pool from resource spec.
	pool := &container.NodePool{}
	np.GenerateNodePool(meta.GetExternalName(cr), cr.Spec.ForProvider, pool)

	create := &container.CreateNodePoolRequest{
		NodePool: pool,
	}

	_, err := e.container.Projects.Locations.Clusters.NodePools.Create(cr.Spec.ForProvider.Cluster, create).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateNodePool)
}

func (e *nodePoolExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta12.NodePool)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotNodePool)
	}
	// Do not issue another update until the node pool finishes the previous
	// one.
	if cr.Status.AtProvider.Status == v1beta12.NodePoolStateReconciling || cr.Status.AtProvider.Status == v1beta12.NodePoolStateProvisioning {
		return managed.ExternalUpdate{}, nil
	}

	// We have to get the node pool again here to determine how to update.
	existing, err := e.container.Projects.Locations.Clusters.NodePools.Get(np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetNodePool)
	}

	u, fn, err := np.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, existing)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckNodePoolUpToDate)
	}
	if u {
		return managed.ExternalUpdate{}, nil
	}

	// GKE uses different update methods depending on the field that is being
	// changed. np.IsUpToDate returns the appropriate update operation based on
	// the difference in the desired and existing spec. If it is a specialized
	// update, only one can be performed at a time. If it is not, then updates
	// can be mass applied.
	_, err = fn(ctx, e.container, np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr)))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateNodePool)
}

func (e *nodePoolExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta12.NodePool)
	if !ok {
		return errors.New(errNotNodePool)
	}
	cr.SetConditions(xpv1.Deleting())
	// Wait until deletion is complete if already stopping.
	if cr.Status.AtProvider.Status == v1beta12.NodePoolStateStopping {
		return nil
	}

	_, err := e.container.Projects.Locations.Clusters.NodePools.Delete(np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteNodePool)
}
