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
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/cache/v1beta1"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/crossplaneio/stack-gcp/pkg/clients/cloudmemorystore"
)

// Error strings.
const (
	errGetProvider       = "cannot get Provider"
	errGetProviderSecret = "cannot get Provider Secret"
	errNewClient         = "cannot create new CloudMemorystore client"
	errNotInstance       = "managed resource is not an CloudMemorystore instance"
	errUpdateCR          = "cannot update CloudMemorystore custom resource"
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
		Named(strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.CloudMemorystoreInstanceKind, v1beta1.Group))).
		For(&v1beta1.CloudMemorystoreInstance{}).
		Complete(resource.NewManagedReconciler(mgr,
			resource.ManagedKind(v1beta1.CloudMemorystoreInstanceGroupVersionKind),
			resource.WithExternalConnecter(&connecter{client: mgr.GetClient(), newCMS: cloudmemorystore.NewClient})))
}

type connecter struct {
	client client.Client
	newCMS func(ctx context.Context, creds []byte) (cloudmemorystore.Client, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return nil, errors.New(errNotInstance)
	}

	p := &gcpv1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(i.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	cms, err := c.newCMS(ctx, s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{cms: cms, projectID: p.Spec.ProjectID, kube: c.client}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube      client.Client
	cms       cloudmemorystore.Client
	projectID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.CloudMemorystoreInstance)
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

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	cloudmemorystore.LateInitializeSpec(&cr.Spec.ForProvider, *existing)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
		}
	}

	conn := resource.ConnectionDetails{}
	switch cr.Status.AtProvider.State {
	case cloudmemorystore.StateReady:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		conn[runtimev1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(cr.Status.AtProvider.Host)
		conn[runtimev1alpha1.ResourceCredentialsSecretPortKey] = []byte(strconv.Itoa(int(cr.Status.AtProvider.Port)))
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
		ConnectionDetails: conn,
	}

	return o, nil

}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotInstance)
	}

	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	i.Status.SetConditions(runtimev1alpha1.Creating())

	_, err := e.cms.CreateInstance(ctx, cloudmemorystore.NewCreateInstanceRequest(id, i))
	return resource.ExternalCreation{}, errors.Wrap(err, errCreateInstance)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotInstance)
	}
	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	_, err := e.cms.UpdateInstance(ctx, cloudmemorystore.NewUpdateInstanceRequest(id, i))
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateInstance)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return errors.New(errNotInstance)
	}
	i.SetConditions(runtimev1alpha1.Deleting())

	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	_, err := e.cms.DeleteInstance(ctx, cloudmemorystore.NewDeleteInstanceRequest(id))
	return errors.Wrap(resource.Ignore(cloudmemorystore.IsNotFound, err), errDeleteInstance)
}
