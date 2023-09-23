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

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/google/go-cmp/cmp"
	dns "google.golang.org/api/dns/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gcp/apis/dns/v1alpha1"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	rrsclient "github.com/crossplane-contrib/provider-gcp/pkg/clients/dns"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
)

const (
	errNewClient            = "cannot create new DNS Service"
	errNotResourceRecordSet = "managed resource is not a ResourceRecordSet custom resource"
	errCreateCluster        = "cannot create new ResourceRecordSet"
	errCannotDelete         = "cannot delete new ResourceRecordSet"
	errGetFailed            = "cannot get the ResourceRecordSet"
	errManagedUpdateFailed  = "cannot update ResourceRecordSet custom resource"
	errCheckUpToDate        = "cannot determine if ResourceRecordSet is up to date"
)

// SetupResourceRecordSet adds a controller that reconciles
// ResourceRecordSet managed resources.
func SetupResourceRecordSet(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ResourceRecordSetGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind, connection.WithTLSConfig(o.ESSOptions.TLSConfig)))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ResourceRecordSetGroupVersionKind),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
		managed.WithInitializers(rrsclient.NewCustomNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.ResourceRecordSet{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetConnectionInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	d, err := dns.NewService(ctx, opts...)
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
	rrsclient.LateInitializeSpec(&cr.Spec.ForProvider, *rrs)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
		lateInit = true
	}
	cr.SetConditions(xpv1.Available())

	upToDate, err := rrsclient.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, rrs)
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
	rrsclient.GenerateResourceRecordSet(
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
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateCluster)
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
	rrsclient.GenerateResourceRecordSet(meta.GetExternalName(cr), cr.Spec.ForProvider, args)

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
