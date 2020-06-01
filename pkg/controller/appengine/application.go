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

package appengine

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	appengine "google.golang.org/api/appengine/v1"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	aev1alpha1 "github.com/crossplane/provider-gcp/apis/appengine/v1alpha1"
	apisv1alpha3 "github.com/crossplane/provider-gcp/apis/v1alpha3"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	ae "github.com/crossplane/provider-gcp/pkg/clients/appengine"
)

const (
	errNotApplication             = "managed resource is not a ApplicationInstance custom resource"
	errProviderNotRetrieved       = "provider could not be retrieved"
	errProviderSecretNil          = "cannot find Secret reference on Provider"
	errProviderSecretNotRetrieved = "secret referred in provider could not be retrieved"
	errManagedUpdateFailed        = "cannot update Application custom resource"

	errNewClient     = "cannot create new AppEngine Application Service"
	errCreateFailed  = "cannot create new Application instance"
	errUpdateFailed  = "cannot update the Application instance"
	errGetFailed     = "cannot get the Application instance"
	errCheckUpToDate = "cannot determine if Application instance is up to date"
)

// SetupApplicationInstance adds a controller that reconciles
// ApplicationInstance managed resources.
func SetupApplicationInstance(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(aev1alpha1.ApplicationKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(aev1alpha1.ApplicationGroupVersionKind),
		managed.WithExternalConnecter(&applicationConnector{kube: mgr.GetClient(), newServiceFn: appengine.NewService}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&aev1alpha1.Application{}).
		Complete(r)
}

type applicationConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*appengine.APIService, error)
}

func (c *applicationConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*aev1alpha1.Application)
	if !ok {
		return nil, errors.New(errNotApplication)
	}

	provider := &apisv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), provider); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}

	if provider.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	secret := &v1.Secret{}
	n := types.NamespacedName{Namespace: provider.Spec.CredentialsSecretRef.Namespace, Name: provider.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, secret); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	s, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(secret.Data[provider.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(sqladmin.SqlserviceAdminScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &applicationExternal{kube: c.kube, ae: s.Apps, projectID: provider.Spec.ProjectID}, nil
}

type applicationExternal struct {
	kube      client.Client
	ae        *appengine.AppsService
	projectID string
}

func (c *applicationExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*aev1alpha1.Application)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplication)
	}

	app, err := c.ae.Get(meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetFailed)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	ae.LateInitializeSpec(&cr.Spec.ForProvider, *app)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}
	cr.Status.AtProvider = ae.GenerateObservation(*app)
	switch cr.Status.AtProvider.ServingStatus {
	case aev1alpha1.StateServing:
		cr.Status.SetConditions(v1alpha1.Available())
		resource.SetBindable(cr)
	case aev1alpha1.StateUserDisabled, aev1alpha1.StateSystemDisabled, aev1alpha1.StateUnspecified:
		cr.Status.SetConditions(v1alpha1.Unavailable())
	}

	upToDate, err := ae.IsUpToDate(c.projectID, &cr.Spec.ForProvider, app)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *applicationExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*aev1alpha1.Application)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplication)
	}
	cr.SetConditions(v1alpha1.Creating())
	app := &appengine.Application{}
	ae.GenerateApplication(c.projectID, cr.Spec.ForProvider, app)
	_, err := c.ae.Create(app).Context(ctx).Do()
	return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errCreateFailed)
}

func (c *applicationExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*aev1alpha1.Application)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotApplication)
	}
	app := &appengine.Application{}
	ae.GenerateApplication(c.projectID, cr.Spec.ForProvider, app)
	_, err := c.ae.Patch(meta.GetExternalName(cr), app).Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (c *applicationExternal) Delete(ctx context.Context, mg resource.Managed) error {
	_, ok := mg.(*aev1alpha1.Application)
	if !ok {
		return errors.New(errNotApplication)
	}
	return nil
}
