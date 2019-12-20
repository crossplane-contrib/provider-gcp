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
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	container "google.golang.org/api/container/v1beta1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/container/v1alpha1"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	np "github.com/crossplaneio/stack-gcp/pkg/clients/nodepool"
)

// Error strings.
const (
	errManagedNodePoolUpdateFailed = "cannot update NodePool custom resource"
	errNotNodePool                 = "managed resource is not a NodePool"
	errGetNodePool                 = "cannot get GKE node pool"
	errCreateNodePool              = "cannot create GKE node pool"
	errUpdateNodePool              = "cannot update GKE node pool"
	errDeleteNodePool              = "cannot delete GKE node pool"
)

// NodePoolController is responsible for adding the NodePool
// controller and its corresponding reconciler to the manager with any runtime configuration.
type NodePoolController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func (c *NodePoolController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1alpha1.NodePoolGroupVersionKind),
		resource.WithExternalConnecter(&nodePoolConnector{kube: mgr.GetClient(), newServiceFn: container.NewService}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha1.NodePoolKindAPIVersion, v1alpha1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.NodePool{}).
		Complete(r)
}

type nodePoolConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*container.Service, error)
}

func (c *nodePoolConnector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	i, ok := mg.(*v1alpha1.NodePool)
	if !ok {
		return nil, errors.New(errNotNodePool)
	}

	p := &gcpv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(i.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.Secret.Namespace, Name: p.Spec.Secret.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	client, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(s.Data[p.Spec.Secret.Key]),
		option.WithScopes(container.CloudPlatformScope))
	return &nodePoolExternal{container: client, projectID: p.Spec.ProjectID, kube: c.kube}, errors.Wrap(err, errNewClient)
}

type nodePoolExternal struct {
	kube      client.Client
	container *container.Service
	projectID string
}

func (e *nodePoolExternal) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.NodePool)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotNodePool)
	}

	existing, err := e.container.Projects.Locations.Clusters.NodePools.Get(np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNodePool)
	}

	cr.Status.AtProvider = np.GenerateObservation(*existing)
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	np.LateInitializeSpec(&cr.Spec.ForProvider, *existing)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errManagedNodePoolUpdateFailed)
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

	u, _ := np.IsUpToDate(&cr.Spec.ForProvider, *existing)

	return resource.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: u,
	}, nil
}

func (e *nodePoolExternal) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.NodePool)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotNodePool)
	}
	cr.SetConditions(runtimev1alpha1.Creating())

	// Wait until creation is complete if already provisioning.
	if cr.Status.AtProvider.Status == v1alpha1.NodePoolStateProvisioning {
		return resource.ExternalCreation{}, nil
	}

	// Generate GKE node pool from resource spec.
	pool := np.GenerateNodePool(cr.Spec.ForProvider, meta.GetExternalName(cr))

	create := &container.CreateNodePoolRequest{
		NodePool: pool,
	}

	if _, err := e.container.Projects.Locations.Clusters.NodePools.Create(cr.Spec.ForProvider.Cluster, create).Context(ctx).Do(); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errCreateNodePool)
	}

	return resource.ExternalCreation{}, nil
}

func (e *nodePoolExternal) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.NodePool)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotNodePool)
	}
	// Do not issue another update until the node pool finishes the previous
	// one.
	if cr.Status.AtProvider.Status == v1alpha1.NodePoolStateReconciling || cr.Status.AtProvider.Status == v1alpha1.NodePoolStateProvisioning {
		return resource.ExternalUpdate{}, nil
	}

	// We have to get the node pool again here to determine how to update.
	existing, err := e.container.Projects.Locations.Clusters.NodePools.Get(np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return resource.ExternalUpdate{}, errors.Wrap(err, errGetNodePool)
	}

	u, fn := np.IsUpToDate(&cr.Spec.ForProvider, *existing)
	if u {
		return resource.ExternalUpdate{}, nil
	}

	// GKE uses different update methods depending on the field that is being
	// changed. np.IsUpToDate returns the appropriate update operation based on
	// the difference in the desired and existing spec. If it is a specialized
	// update, only one can be performed at a time. If it is not, then updates
	// can be mass applied.
	_, err = fn(ctx, e.container, np.GetFullyQualifiedName(cr.Spec.ForProvider, meta.GetExternalName(cr)))
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateNodePool)
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
