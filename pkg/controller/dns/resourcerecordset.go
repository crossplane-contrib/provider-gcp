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

package dns

import (
	"context"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/database/v1beta1"
	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	rrs "github.com/crossplane/provider-gcp/pkg/clients/resourcerecordset"
)

// Error strings.
const (
	errNewClient    = "cannot create new GCP DNS API client"
	errNotRecordSet = "managed resource is not a DNS ResourceRecordSet"
	errGet          = "cannot get GCP ResourceRecordSet object via DNS API"
	errCreate       = "cannot create GCP ResourceRecordSet object via DNS API"
	errUpdate       = "cannot update GCP ResourceRecordSet object via DNS API"
	errDelete       = "cannot delete GCP ResourceRecordSet object via DNS API"
)

// SetupResourceRecordSet adds a controller that reconciles
// ResourceRecordSet managed resources.
func SetupResourceRecordSet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.CloudSQLInstanceGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ResourceRecordSet{}).
		Complete(r)
}

type connecter struct {
	client client.Client
}

// Connect sets up iam client using credentials from the provider
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}

	cl, err := rrs.NewClient(ctx, projectID, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &external{client: cl}, errors.Wrap(err, errNewClient)
}

type external struct {
	client rrs.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRecordSet)
	}

	fromProvider, err := e.client.Get(ctx, cr)
	if gcp.IsErrorNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGet)
	}
	obs := rrs.GenerateObservation(fromProvider)
	cr.Status.AtProvider = obs

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  rrs.IsUpToDate(&cr.Spec.ForProvider, fromProvider),
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

// https://cloud.google.com/dns/docs/reference/v1beta2/resourceRecordSets/create
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRecordSet)
	}

	_, err := e.client.Create(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

// https://cloud.google.com/dns/docs/reference/v1beta2/resourceRecordSets/patch
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRecordSet)
	}

	_, err := e.client.Update(ctx, cr)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

// https://cloud.google.com/dns/docs/reference/v1beta2/resourceRecordSets/delete
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return errors.New(errNotRecordSet)
	}

	_, err := e.client.Delete(ctx, cr)
	if gcp.IsErrorNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errDelete)
}
