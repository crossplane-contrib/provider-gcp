/*
Copyright 2021 The Crossplane Authors.

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

package dns

import (
	"context"
	"time"

	"github.com/google/go-cmp/cmp"
	dns "google.golang.org/api/dns/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	rrsClient "github.com/crossplane/provider-gcp/pkg/clients/dns"
)

const (
	errNewClient            = "cannot create new DNS Service"
	errNotResourceRecordSet = "managed resource is not a ResourceRecordSet custom resource"
	errCannotCreate         = "cannot create new ResourceRecordSet"
	errCannotDelete         = "cannot delete new ResourceRecordSet"
	errGetFailed            = "cannot get the ResourceRecordSet"
	errManagedUpdateFailed  = "cannot update ResourceRecordSet custom resource"
	errCheckUpToDate        = "cannot determine if ResourceRecordSet is up to date"
)

// SetupResourceRecordSet adds a controller that reconciles
// ResourceRecordSet managed resources.
func SetupResourceRecordSet(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.ResourceRecordSetGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ResourceRecordSetGroupVersionKind),
		managed.WithExternalConnecter(
			&connector{
				kube: mgr.GetClient(),
			},
		),
		managed.WithInitializers(
			rrsClient.NewCustomNameAsExternalName(mgr.GetClient()),
		),
		managed.WithReferenceResolver(
			managed.NewAPISimpleReferenceResolver(mgr.GetClient()),
		),
		managed.WithPollInterval(poll),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(
			mgr.GetEventRecorderFor(name)),
		),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.ResourceRecordSet{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	d, err := dns.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &external{
		kube:      c.kube,
		dns:       d.ResourceRecordSets,
		projectID: projectID,
	}, nil
}

type external struct {
	kube      client.Client
	dns       *dns.ResourceRecordSetsService
	projectID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotResourceRecordSet)
	}

	rrs, err := e.dns.Get(
		e.projectID,
		cr.Spec.ForProvider.ManagedZone,
		meta.GetExternalName(cr),
		cr.Spec.ForProvider.Type,
	).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(
			resource.Ignore(gcp.IsErrorNotFound, err),
			errGetFailed,
		)
	}

	lateInit := false
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	rrsClient.LateInitializeSpec(&cr.Spec.ForProvider, *rrs)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
		lateInit = true
	}
	cr.SetConditions(xpv1.Available())

	upToDate, err := rrsClient.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, rrs)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: lateInit,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotResourceRecordSet)
	}

	args := &dns.ResourceRecordSet{}
	rrsClient.GenerateResourceRecordSet(
		meta.GetExternalName(cr),
		cr.Spec.ForProvider,
		args,
	)

	_, err := e.dns.Create(
		e.projectID,
		cr.Spec.ForProvider.ManagedZone,
		args,
	).Context(ctx).Do()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCannotCreate)
	}
	cr.SetConditions(xpv1.Creating())
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotResourceRecordSet)
	}

	args := &dns.ResourceRecordSet{}
	rrsClient.GenerateResourceRecordSet(meta.GetExternalName(cr), cr.Spec.ForProvider, args)

	_, err := e.dns.Patch(
		e.projectID,
		cr.Spec.ForProvider.ManagedZone,
		meta.GetExternalName(cr),
		cr.Spec.ForProvider.Type,
		args,
	).Context(ctx).Do()
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return errors.New(errNotResourceRecordSet)
	}

	_, err := e.dns.Delete(
		e.projectID,
		cr.Spec.ForProvider.ManagedZone,
		meta.GetExternalName(cr),
		cr.Spec.ForProvider.Type,
	).Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errCannotDelete)
}
