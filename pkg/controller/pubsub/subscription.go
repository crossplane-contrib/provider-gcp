/*
Copyright 2020 The Crossplane Authors.

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

package pubsub

import (
	"context"
	"time"

	"github.com/google/go-cmp/cmp"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/pubsub/v1alpha1"
	"github.com/crossplane/provider-gcp/pkg/clients/subscription"
)

const (
	errNotSubscription        = "managed resource is not of type Subscription"
	errGetSubscription        = "cannot get Subscription"
	errUpdateSubscription     = "cannot update Subscription"
	errKubeUpdateSubscription = "cannot update Subscription custom resource"
	errCreateSubscription     = "cannot create Subscription"
	errDeleteSubscription     = "cannot delete Subscription"
)

// SetupSubscription adds a controller that reconciles Subscriptions.
func SetupSubscription(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.SubscriptionGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.Subscription{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SubscriptionGroupVersionKind),
			managed.WithExternalConnecter(&subscriptionConnector{client: mgr.GetClient()}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type subscriptionConnector struct {
	client client.Client
}

// Connect returns an ExternalClient with necessary information to talk to GCP API.
func (c *subscriptionConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}

	s, err := pubsub.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &subscriptionExternal{projectID: projectID, client: c.client, ps: s}, nil
}

type subscriptionExternal struct {
	projectID string
	client    client.Client
	ps        *pubsub.Service
}

// Observe makes observation about the external resource.
func (e *subscriptionExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Subscription)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubscription)
	}

	s, err := e.ps.Projects.Subscriptions.Get(subscription.GetFullyQualifiedName(e.projectID, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetSubscription)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	subscription.LateInitialize(e.projectID, &cr.Spec.ForProvider, *s)

	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.client.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateSubscription)
		}
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: subscription.IsUpToDate(e.projectID, cr.Spec.ForProvider, *s),
	}, nil
}

// Create initiates creation of external resource.
func (e *subscriptionExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Subscription)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubscription)
	}

	cr.SetConditions(xpv1.Creating())

	_, err := e.ps.Projects.Subscriptions.Create(subscription.GetFullyQualifiedName(e.projectID, meta.GetExternalName(cr)),
		subscription.GenerateSubscription(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider)).Context(ctx).Do()

	return managed.ExternalCreation{}, errors.Wrap(err, errCreateSubscription)
}

// Update initiates an update to the external resource.
func (e *subscriptionExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Subscription)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubscription)
	}

	s, err := e.ps.Projects.Subscriptions.Get(subscription.GetFullyQualifiedName(e.projectID, meta.GetExternalName(cr))).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetSubscription)
	}

	_, err = e.ps.Projects.Subscriptions.Patch(subscription.GetFullyQualifiedName(e.projectID, meta.GetExternalName(cr)),
		subscription.GenerateUpdateRequest(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider, *s)).Context(ctx).Do()

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSubscription)
}

// Delete initiates an deletion of the external resource.
func (e *subscriptionExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Subscription)
	if !ok {
		return errors.New(errNotSubscription)
	}

	_, err := e.ps.Projects.Subscriptions.Delete(subscription.GetFullyQualifiedName(e.projectID,
		meta.GetExternalName(cr))).Context(ctx).Do()

	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteSubscription)
}
