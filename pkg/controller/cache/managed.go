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
	"strings"

	"github.com/google/go-cmp/cmp"
	redis "google.golang.org/api/redis/v1"
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

	"github.com/crossplane-contrib/provider-gcp/apis/cache/v1beta1"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	"github.com/crossplane-contrib/provider-gcp/pkg/clients/cloudmemorystore"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
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
	errAuthString     = "cannot retrieve AuthString for instance"
)

// SetupCloudMemorystoreInstance adds a controller that reconciles
// CloudMemorystoreInstances.
func SetupCloudMemorystoreInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.CloudMemorystoreInstanceGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind, connection.WithTLSConfig(o.ESSOptions.TLSConfig)))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.CloudMemorystoreInstanceGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.CloudMemorystoreInstance{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetConnectionInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := redis.NewService(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &external{cms: s, projectID: projectID, kube: c.client}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube      client.Client
	cms       *redis.Service
	projectID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotInstance)
	}

	existing, err := e.cms.Projects.Locations.Instances.Get(cloudmemorystore.GetFullyQualifiedName(e.projectID, cr.Spec.ForProvider, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetInstance)
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
		if cr.Spec.ForProvider.AuthEnabled != nil {
			if *cr.Spec.ForProvider.AuthEnabled {
				existingAuthString, err := e.cms.Projects.Locations.Instances.GetAuthString(existing.Name).Context(ctx).Do()
				if err != nil {
					return managed.ExternalObservation{}, errors.Wrap(err, errAuthString)
				}
				conn[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(cloudmemorystore.GenerateAuthStringObservation(*existingAuthString))
			}
		}
	case cloudmemorystore.StateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case cloudmemorystore.StateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	u, err := cloudmemorystore.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, existing)
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

	i.Status.SetConditions(xpv1.Creating())

	// Generate Redis instance from resource spec.
	instance := &redis.Instance{}
	cloudmemorystore.GenerateRedisInstance(cloudmemorystore.GetFullyQualifiedName(e.projectID, i.Spec.ForProvider, meta.GetExternalName(i)), i.Spec.ForProvider, instance)

	_, err := e.cms.Projects.Locations.Instances.Create(cloudmemorystore.GetFullyQualifiedParent(e.projectID, i.Spec.ForProvider), instance).InstanceId(meta.GetExternalName(i)).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateInstance)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotInstance)
	}
	// Generate Redis instance from resource spec.
	instance := &redis.Instance{}
	fqn := cloudmemorystore.GetFullyQualifiedName(e.projectID, i.Spec.ForProvider, meta.GetExternalName(i))
	cloudmemorystore.GenerateRedisInstance(fqn, i.Spec.ForProvider, instance)
	updateMask := strings.Join([]string{"display_name", "labels", "memory_size_gb", "redis_configs"}, ",")
	_, err := e.cms.Projects.Locations.Instances.Patch(fqn, instance).UpdateMask(updateMask).Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateInstance)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	i, ok := mg.(*v1beta1.CloudMemorystoreInstance)
	if !ok {
		return errors.New(errNotInstance)
	}
	i.SetConditions(xpv1.Deleting())

	_, err := e.cms.Projects.Locations.Instances.Delete(cloudmemorystore.GetFullyQualifiedName(e.projectID, i.Spec.ForProvider, meta.GetExternalName(i))).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteInstance)
}
