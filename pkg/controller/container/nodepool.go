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

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/container/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	np "github.com/crossplane/provider-gcp/pkg/clients/nodepool"
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
func SetupNodePool(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.NodePoolGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.NodePool{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.NodePoolGroupVersionKind),
			managed.WithExternalConnecter(&nodePoolConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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
	cr, ok := mg.(*v1alpha1.NodePool)
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
	case v1alpha1.NodePoolStateRunning, v1alpha1.NodePoolStateReconciling:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1alpha1.NodePoolStateProvisioning:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.NodePoolStateUnspecified, v1alpha1.NodePoolStateRunningError, v1alpha1.NodePoolStateError:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
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
	cr, ok := mg.(*v1alpha1.NodePool)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNodePool)
	}
	cr.SetConditions(runtimev1alpha1.Creating())

	// Wait until creation is complete if already provisioning.
	if cr.Status.AtProvider.Status == v1alpha1.NodePoolStateProvisioning {
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
	cr, ok := mg.(*v1alpha1.NodePool)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotNodePool)
	}
	// Do not issue another update until the node pool finishes the previous
	// one.
	if cr.Status.AtProvider.Status == v1alpha1.NodePoolStateReconciling || cr.Status.AtProvider.Status == v1alpha1.NodePoolStateProvisioning {
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
	cr, ok := mg.(*v1alpha1.NodePool)
	if !ok {
		return errors.New(errNotNodePool)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	// Wait until deletion is complete if already stopping.
	if cr.Status.AtProvider.Status == v1alpha1.NodePoolStateStopping {
		return nil
	}

	_, err := e.container.Projects.Locations.Clusters.NodePools.Delete(np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteNodePool)
}
