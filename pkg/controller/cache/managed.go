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

package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/cache/v1alpha2"
	gcpv1alpha2 "github.com/crossplaneio/stack-gcp/apis/v1alpha2"
	"github.com/crossplaneio/stack-gcp/pkg/clients/cloudmemorystore"
)

// Error strings.
const (
	errGetProvider       = "cannot get Provider"
	errGetProviderSecret = "cannot get Provider Secret"
	errNewClient         = "cannot create new CloudMemorystore client"
	errNotInstance       = "managed resource is not an CloudMemorystore instance"
	errGetInstance       = "cannot get CloudMemorystore instance"
	errCreateInstance    = "cannot create CloudMemorystore instance"
	errUpdateInstance    = "cannot update CloudMemorystore instance"
	errDeleteInstance    = "cannot delete CloudMemorystore instance"
)

// CloudMemorystoreInstanceController is responsible for adding the Cloud Memorystore
// controller and its corresponding reconciler to the manager with any runtime configuration.
type CloudMemorystoreInstanceController struct{}

// SetupWithManager creates a new CloudMemorystoreInstance Controller and adds it to the
// Manager with default RBAC. The Manager will set fields on the Controller and
// start it when the Manager is Started.
func (c *CloudMemorystoreInstanceController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(strings.ToLower(fmt.Sprintf("%s.%s", v1alpha2.CloudMemorystoreInstanceKind, v1alpha2.Group))).
		For(&v1alpha2.CloudMemorystoreInstance{}).
		Complete(resource.NewManagedReconciler(mgr,
			resource.ManagedKind(v1alpha2.CloudMemorystoreInstanceGroupVersionKind),
			resource.WithExternalConnecter(&connecter{client: mgr.GetClient(), newCMS: cloudmemorystore.NewClient})))
}

type connecter struct {
	client client.Client
	newCMS func(ctx context.Context, creds []byte) (cloudmemorystore.Client, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	i, ok := mg.(*v1alpha2.CloudMemorystoreInstance)
	if !ok {
		return nil, errors.New(errNotInstance)
	}

	p := &gcpv1alpha2.Provider{}
	n := meta.NamespacedNameOf(i.Spec.ProviderReference)
	if err := c.client.Get(ctx, n, p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n = types.NamespacedName{Namespace: p.Namespace, Name: p.Spec.Secret.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	cms, err := c.newCMS(ctx, s.Data[p.Spec.Secret.Key])
	return &external{cms: cms, projectID: p.Spec.ProjectID}, errors.Wrap(err, errNewClient)
}

type external struct {
	cms       cloudmemorystore.Client
	projectID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha2.CloudMemorystoreInstance)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotInstance)
	}

	id := cloudmemorystore.NewInstanceID(e.projectID, cr)
	existing, err := e.cms.GetInstance(ctx, cloudmemorystore.NewGetInstanceRequest(id))
	if cloudmemorystore.IsNotFound(err) {
		return resource.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errGetInstance)
	}

	cr.Status.AtProvider = cloudmemorystore.GenerateObservation(*existing)

	switch cr.Status.AtProvider.State {
	case cloudmemorystore.StateReady:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case cloudmemorystore.StateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case cloudmemorystore.StateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	o := resource.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  cloudmemorystore.IsUpToDate(cr, existing),
		ConnectionDetails: resource.ConnectionDetails{},
	}

	if cr.Status.AtProvider.Host != "" {
		o.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(cr.Status.AtProvider.Host)
	}

	return o, nil

}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	i, ok := mg.(*v1alpha2.CloudMemorystoreInstance)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotInstance)
	}

	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	i.Status.SetConditions(runtimev1alpha1.Creating())

	_, err := e.cms.CreateInstance(ctx, cloudmemorystore.NewCreateInstanceRequest(id, i))
	return resource.ExternalCreation{}, errors.Wrap(err, errCreateInstance)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	i, ok := mg.(*v1alpha2.CloudMemorystoreInstance)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotInstance)
	}
	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	_, err := e.cms.UpdateInstance(ctx, cloudmemorystore.NewUpdateInstanceRequest(id, i))
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateInstance)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	i, ok := mg.(*v1alpha2.CloudMemorystoreInstance)
	if !ok {
		return errors.New(errNotInstance)
	}
	i.SetConditions(runtimev1alpha1.Deleting())

	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	_, err := e.cms.DeleteInstance(ctx, cloudmemorystore.NewDeleteInstanceRequest(id))
	return errors.Wrap(resource.Ignore(cloudmemorystore.IsNotFound, err), errDeleteInstance)
}
