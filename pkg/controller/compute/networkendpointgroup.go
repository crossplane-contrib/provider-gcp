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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/networkendpointgroup"
)

// Error strings.
const (
	errNotNetworkEndpointGroup            = "managed resource is not a NetworkEndpointGroup"
	errGetNetworkEndpointGroup            = "cannot get external NetworkEndpointGroup resource"
	errGetNetworkEndpointWithHealthStatus = "cannot get external NetworkEndpoint with health status for NetworkEndpointGroup resource"
	errCreateNetworkEndpointGroup         = "cannot create external NetworkEndpointGroup resource"
	errDeleteNetworkEndpointGroup         = "cannot delete external NetworkEndpointGroup resource"
	errManagedNetworkEndpointGroupUpdate  = "cannot update managed NetworkEndpointGroup resource"
	errCheckNetworkEndpointGroupUpToDate  = "cannot determine if GCP NetworkEndpointGroup is up to date"
	errUpdateNetworkEndpointGroupFailed   = "update of GCP NetworkEndpointGroup has failed"
)

// SetupNetworkEndpointGroup adds a controller that reconciles
// NetworkEndpointGroup managed resources.
func SetupNetworkEndpointGroup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.NetworkEndpointGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1beta1.NetworkEndpointGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.NetworkEndpointGroupGroupVersionKind),
			managed.WithExternalConnecter(&gaConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type negConnector struct {
	kube client.Client
}

func (c *negConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := compute.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &negExternal{kube: c.kube, Service: s, projectID: projectID}, errors.Wrap(err, errNewClient)
}

type negExternal struct {
	kube      client.Client
	projectID string
	*compute.Service
}

func (e *negExternal) getNetworkEndpointGroup(ctx context.Context, neg *v1beta1.NetworkEndpointGroup) (observed *compute.NetworkEndpointGroup, observedEndPoints []*compute.NetworkEndpointWithHealthStatus, err error) {

	if neg.Spec.ForProvider.Zone != nil {
		observed, err = e.NetworkEndpointGroups.Get(e.projectID, *neg.Spec.ForProvider.Zone, meta.GetExternalName(neg)).Context(ctx).Do()
		if err != nil {
			return nil, nil, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNetworkEndpointGroup)
		}
		err = e.NetworkEndpointGroups.ListNetworkEndpoints(e.projectID, *neg.Spec.ForProvider.Zone, meta.GetExternalName(neg), &compute.NetworkEndpointGroupsListEndpointsRequest{
			HealthStatus: "SHOW",
		}).Pages(ctx, func(neglne *compute.NetworkEndpointGroupsListNetworkEndpoints) error {
			observedEndPoints = append(observedEndPoints, neglne.Items...)
			return nil
		})
		if err != nil {
			return nil, nil, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNetworkEndpointWithHealthStatus)
		}
	}
	if neg.Spec.ForProvider.Region != nil {
		observed, err = e.RegionNetworkEndpointGroups.Get(e.projectID, *neg.Spec.ForProvider.Region, meta.GetExternalName(neg)).Context(ctx).Do()
		if err != nil {
			return nil, nil, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNetworkEndpointGroup)
		}
	}
	if neg.Spec.ForProvider.Zone == nil && neg.Spec.ForProvider.Region == nil {
		observed, err = e.GlobalNetworkEndpointGroups.Get(e.projectID, meta.GetExternalName(neg)).Context(ctx).Do()
		if err != nil {
			return nil, nil, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNetworkEndpointGroup)
		}
		err = e.GlobalNetworkEndpointGroups.ListNetworkEndpoints(e.projectID, meta.GetExternalName(neg)).Pages(ctx, func(neglne *compute.NetworkEndpointGroupsListNetworkEndpoints) error {
			observedEndPoints = append(observedEndPoints, neglne.Items...)
			return nil
		})
		if err != nil {
			return nil, nil, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetNetworkEndpointWithHealthStatus)
		}
	}
	return observed, observedEndPoints, nil
}

func (e *negExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	neg, ok := mg.(*v1beta1.NetworkEndpointGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotNetworkEndpointGroup)
	}

	observed, observedEndPoints, err := e.getNetworkEndpointGroup(ctx, neg)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	currentSpec := neg.Spec.ForProvider.DeepCopy()
	networkendpointgroup.LateInitializeSpec(&neg.Spec.ForProvider, *observed, observedEndPoints)

	if !cmp.Equal(currentSpec, &neg.Spec.ForProvider) {
		if err := e.kube.Update(ctx, neg); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedNetworkEndpointGroupUpdate)
		}
	}

	neg.Status.AtProvider = networkendpointgroup.GenerateNetworkEndpointGroupObservation(*observed, observedEndPoints)

	neg.SetConditions(xpv1.Available())

	u, _, _, _, err := networkendpointgroup.IsUpToDate(meta.GetExternalName(neg), &neg.Spec.ForProvider, observed, observedEndPoints)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckNetworkEndpointGroupUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        u,
		ResourceLateInitialized: !cmp.Equal(currentSpec, &neg.Spec.ForProvider),
	}, nil
}

func (e *negExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	neg, ok := mg.(*v1beta1.NetworkEndpointGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNetworkEndpointGroup)
	}

	neg.Status.SetConditions(xpv1.Creating())

	networkEndpointGroup, networkendpoints := networkendpointgroup.GenerateNetworkEndpointGroup(meta.GetExternalName(neg), neg.Spec.ForProvider)

	if neg.Spec.ForProvider.Zone != nil {
		_, err := e.NetworkEndpointGroups.Insert(e.projectID, *neg.Spec.ForProvider.Zone, networkEndpointGroup).Context(ctx).Do()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreateNetworkEndpointGroup)
		}
		_, err = e.NetworkEndpointGroups.AttachNetworkEndpoints(e.projectID, *neg.Spec.ForProvider.Zone, meta.GetExternalName(neg), &compute.NetworkEndpointGroupsAttachEndpointsRequest{
			NetworkEndpoints: networkendpoints,
		}).Context(ctx).Do()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreateNetworkEndpointGroup)
		}
	}
	if neg.Spec.ForProvider.Region != nil {
		_, err := e.RegionNetworkEndpointGroups.Insert(e.projectID, *neg.Spec.ForProvider.Region, networkEndpointGroup).Context(ctx).Do()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreateNetworkEndpointGroup)
		}
		// no endpoints allowed for regional network endpoint group
	}
	if neg.Spec.ForProvider.Zone == nil && neg.Spec.ForProvider.Region == nil {
		_, err := e.GlobalNetworkEndpointGroups.Insert(e.projectID, networkEndpointGroup).Context(ctx).Do()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreateNetworkEndpointGroup)
		}
		_, err = e.GlobalNetworkEndpointGroups.AttachNetworkEndpoints(e.projectID, meta.GetExternalName(neg), &compute.GlobalNetworkEndpointGroupsAttachEndpointsRequest{
			NetworkEndpoints: networkendpoints,
		}).Context(ctx).Do()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreateNetworkEndpointGroup)
		}
	}
	return managed.ExternalCreation{}, nil
}

func (e *negExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	neg, ok := mg.(*v1beta1.NetworkEndpointGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotNetworkEndpointGroup)
	}

	observed, observedEndPoints, err := e.getNetworkEndpointGroup(ctx, neg)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	// only network endpoints can be updated
	u, _, toBeAdded, toBeRemoved, err := networkendpointgroup.IsUpToDate(meta.GetExternalName(neg), &neg.Spec.ForProvider, observed, observedEndPoints)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCheckNetworkEndpointGroupUpToDate)
	}
	if u {
		if neg.Spec.ForProvider.Zone != nil {
			_, err := e.NetworkEndpointGroups.DetachNetworkEndpoints(e.projectID, *neg.Spec.ForProvider.Zone, meta.GetExternalName(neg), &compute.NetworkEndpointGroupsDetachEndpointsRequest{
				NetworkEndpoints: toBeRemoved,
			}).Context(ctx).Do()
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateNetworkEndpointGroupFailed)
			}
			_, err = e.NetworkEndpointGroups.AttachNetworkEndpoints(e.projectID, *neg.Spec.ForProvider.Zone, meta.GetExternalName(neg), &compute.NetworkEndpointGroupsAttachEndpointsRequest{
				NetworkEndpoints: toBeAdded,
			}).Context(ctx).Do()
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateNetworkEndpointGroupFailed)
			}
		}

		if neg.Spec.ForProvider.Region != nil {
			// no endpoints for regional networkendpoint groups, so we can return
			return managed.ExternalUpdate{}, nil
		}
		if neg.Spec.ForProvider.Zone == nil && neg.Spec.ForProvider.Region == nil {
			_, err := e.GlobalNetworkEndpointGroups.DetachNetworkEndpoints(e.projectID, meta.GetExternalName(neg), &compute.GlobalNetworkEndpointGroupsDetachEndpointsRequest{
				NetworkEndpoints: toBeRemoved,
			}).Context(ctx).Do()
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateNetworkEndpointGroupFailed)
			}
			_, err = e.GlobalNetworkEndpointGroups.AttachNetworkEndpoints(e.projectID, meta.GetExternalName(neg), &compute.GlobalNetworkEndpointGroupsAttachEndpointsRequest{
				NetworkEndpoints: toBeAdded,
			}).Context(ctx).Do()
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateNetworkEndpointGroupFailed)
			}
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *negExternal) Delete(ctx context.Context, mg resource.Managed) error {
	neg, ok := mg.(*v1beta1.NetworkEndpointGroup)
	if !ok {
		return errors.New(errNotNetworkEndpointGroup)
	}

	neg.Status.SetConditions(xpv1.Deleting())
	if neg.Spec.ForProvider.Zone != nil {
		_, err := e.NetworkEndpointGroups.Delete(e.projectID, *neg.Spec.ForProvider.Zone, meta.GetExternalName(neg)).Context(ctx).Do()
		if err != nil {
			return errors.Wrap(err, errDeleteNetworkEndpointGroup)
		}
	}
	if neg.Spec.ForProvider.Region != nil {
		_, err := e.RegionNetworkEndpointGroups.Delete(e.projectID, *neg.Spec.ForProvider.Region, meta.GetExternalName(neg)).Context(ctx).Do()
		if err != nil {
			return errors.Wrap(err, errDeleteNetworkEndpointGroup)
		}
	}
	if neg.Spec.ForProvider.Zone == nil && neg.Spec.ForProvider.Region == nil {
		_, err := e.GlobalNetworkEndpointGroups.Delete(e.projectID, meta.GetExternalName(neg)).Context(ctx).Do()
		if err != nil {
			return errors.Wrap(err, errDeleteNetworkEndpointGroup)
		}
	}
	return nil
}
