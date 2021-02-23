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
	"strconv"

	redisv1 "cloud.google.com/go/redis/apiv1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/cache/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/cloudmemorystore"
)

// Error strings.
const (
	errNewClient      = "cannot create new CloudMemorystore client"
	errNotInstance    = "managed resource is not an CloudMemorystore instance"
	errUpdateCR       = "cannot update CloudMemorystore custom resource"
	errGetInstance    = "cannot get CloudMemorystore instance"
	errCreateInstance = "cannot create CloudMemorystore instance"
	errUpdateInstance = "cannot update CloudMemorystore instance"
	errDeleteInstance = "cannot delete CloudMemorystore instance"
	errCheckUpToDate  = "cannot determine if CloudMemorystore instance is up to date"
)

// SetupCloudMemorystoreInstance adds a controller that reconciles
// CloudMemorystoreInstances.
func SetupCloudMemorystoreInstance(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1beta1.CloudMemorystoreInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1beta1.CloudMemorystoreInstance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.CloudMemorystoreInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := redisv1.NewCloudRedisClient(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &external{cms: s, projectID: projectID, kube: c.client}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube      client.Client
	cms       cloudmemorystore.Client
	projectID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotInstance)
	}

	id := cloudmemorystore.NewInstanceID(e.projectID, cr)
	existing, err := e.cms.GetInstance(ctx, cloudmemorystore.NewGetInstanceRequest(id))
	if cloudmemorystore.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetInstance)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	cloudmemorystore.LateInitializeSpec(&cr.Spec.ForProvider, *existing)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
		}
	}
	cr.Status.AtProvider = cloudmemorystore.GenerateObservation(*existing)
	conn := managed.ConnectionDetails{}
	switch cr.Status.AtProvider.State {
	case cloudmemorystore.StateReady:
		cr.Status.SetConditions(xpv1.Available())
		conn[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(cr.Status.AtProvider.Host)
		conn[xpv1.ResourceCredentialsSecretPortKey] = []byte(strconv.Itoa(int(cr.Status.AtProvider.Port)))
	case cloudmemorystore.StateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case cloudmemorystore.StateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	u, err := cloudmemorystore.IsUpToDate(id, &cr.Spec.ForProvider, existing)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  u,
		ConnectionDetails: conn,
	}

	return o, nil

}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotInstance)
	}

	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	i.Status.SetConditions(xpv1.Creating())

	_, err := e.cms.CreateInstance(ctx, cloudmemorystore.NewCreateInstanceRequest(id, i))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateInstance)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotInstance)
	}
	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	_, err := e.cms.UpdateInstance(ctx, cloudmemorystore.NewUpdateInstanceRequest(id, i))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateInstance)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return errors.New(errNotInstance)
	}
	i.SetConditions(xpv1.Deleting())

	id := cloudmemorystore.NewInstanceID(e.projectID, i)
	_, err := e.cms.DeleteInstance(ctx, cloudmemorystore.NewDeleteInstanceRequest(id))
	return errors.Wrap(resource.Ignore(cloudmemorystore.IsNotFound, err), errDeleteInstance)
}
