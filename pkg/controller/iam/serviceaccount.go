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

package iam

import (
	"context"
	"fmt"

	iamv1 "google.golang.org/api/iam/v1"
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

	"github.com/crossplane-contrib/provider-gcp/apis/iam/v1alpha1"
	scv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane-contrib/provider-gcp/pkg/clients"
	"github.com/crossplane-contrib/provider-gcp/pkg/clients/serviceaccount"
	"github.com/crossplane-contrib/provider-gcp/pkg/features"
)

// Error strings.
const (
	errNewClient         = "cannot create new GCP IAM API client"
	errNotServiceAccount = "managed resource is not a GCP ServiceAccount"
	errGet               = "cannot get GCP ServiceAccount object via IAM API"
	errCreate            = "cannot create GCP ServiceAccount object via IAM API"
	errUpdate            = "cannot update GCP ServiceAccount object via IAM API"
	errDelete            = "cannot delete GCP ServiceAccount object via IAM API"
)

// SetupServiceAccount adds a controller that reconciles ServiceAccounts.
func SetupServiceAccount(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ServiceAccountGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind, connection.WithTLSConfig(o.ESSOptions.TLSConfig)))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ServiceAccountGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.ServiceAccount{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connecter struct {
	client client.Client
}

// Connect sets up iam client using credentials from the provider
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetConnectionInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := iamv1.NewService(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	rrn := NewRelativeResourceNamer(projectID)
	return &external{serviceAccounts: s.Projects.ServiceAccounts, rrn: rrn}, errors.Wrap(err, errNewClient)
}

type external struct {
	serviceAccounts serviceaccount.Client
	rrn             RelativeResourceNamer
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotServiceAccount)
	}

	req := e.serviceAccounts.Get(e.rrn.ResourceName(cr))
	fromProvider, err := req.Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGet)
	}

	populateCRFromProvider(cr, fromProvider)
	if fromProvider.Email != "" {
		cr.Status.SetConditions(xpv1.Available())
	}
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  isUpToDate(&cr.Spec.ForProvider, fromProvider),
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/create
// Note that the metadata.Name from the Kubernetes custom resource is used as the AccountID parameter
// All other API methods use the external-name annotation
// (set via the RelativeResourceNameAsExternalName Initializer)
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotServiceAccount)
	}

	csar := &iamv1.CreateServiceAccountRequest{
		AccountId: meta.GetExternalName(cr),
		ServiceAccount: &iamv1.ServiceAccount{
			DisplayName: gcp.StringValue(cr.Spec.ForProvider.DisplayName),
			Description: gcp.StringValue(cr.Spec.ForProvider.Description),
		},
	}

	// The first parameter to the Create method is the resource name of the GCP project
	// where the service account should be created
	req := e.serviceAccounts.Create(e.rrn.ProjectName(), csar)
	fromProvider, err := req.Context(ctx).Do()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}
	populateCRFromProvider(cr, fromProvider)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/patch
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotServiceAccount)
	}

	sa := &iamv1.ServiceAccount{}
	populateProviderFromCR(sa, cr)
	psar := &iamv1.PatchServiceAccountRequest{
		ServiceAccount: sa,
		UpdateMask:     "description,displayName",
	}
	req := e.serviceAccounts.Patch(e.rrn.ResourceName(cr), psar)
	// we don't pay attention to the result of the patch request because it is only guaranteed to contain
	// `description` and `displayName` ie the fields we are trying to change
	_, err := req.Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/delete
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ServiceAccount)
	if !ok {
		return errors.New(errNotServiceAccount)
	}

	req := e.serviceAccounts.Delete(e.rrn.ResourceName(cr))
	_, err := req.Context(ctx).Do()

	if gcp.IsErrorNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errDelete)
}

// isUpToDate returns true if the supplied Kubernetes resource does not differ
//
//	from the supplied GCP resource. It considers only fields that can be
//	modified in place without deleting and recreating the Service Account.
func isUpToDate(in *v1alpha1.ServiceAccountParameters, observed *iamv1.ServiceAccount) bool {
	// see comment in serviceaccount_types.go
	if in.DisplayName != nil && *in.DisplayName != observed.DisplayName {
		return false
	}
	if in.Description != nil && *in.Description != observed.Description {
		return false
	}
	return true
}

func populateCRFromProvider(cr *v1alpha1.ServiceAccount, fromProvider *iamv1.ServiceAccount) {
	cr.Status.AtProvider.UniqueID = fromProvider.UniqueId
	cr.Status.AtProvider.Email = fromProvider.Email
	cr.Status.AtProvider.Oauth2ClientID = fromProvider.Oauth2ClientId
	cr.Status.AtProvider.Disabled = fromProvider.Disabled
	cr.Status.AtProvider.Name = fromProvider.Name
}

func populateProviderFromCR(forProvider *iamv1.ServiceAccount, cr *v1alpha1.ServiceAccount) {
	forProvider.DisplayName = gcp.StringValue(cr.Spec.ForProvider.DisplayName)
	forProvider.Description = gcp.StringValue(cr.Spec.ForProvider.Description)
}

// NewRelativeResourceNamer makes an instance of the RelativeResourceNamer
// which is the only type that is allowed to know how to construct GCP resource names
// for the IAM type.
func NewRelativeResourceNamer(projectName string) RelativeResourceNamer {
	return RelativeResourceNamer{projectName: projectName}
}

// RelativeResourceNamer allows the controller to generate the "relative resource name"
// for the service account and GCP project based on the external-name annotation.
// https://cloud.google.com/apis/design/resource_names#relative_resource_name
// The relative resource name for service accounts has the following format:
// projects/{project_id}/serviceAccounts/{account name}
type RelativeResourceNamer struct {
	projectName string
}

// ProjectName yields the relative resource name for a GCP project
func (rrn RelativeResourceNamer) ProjectName() string {
	return fmt.Sprintf("projects/%s", rrn.projectName)
}

// ResourceName yields the relative resource name for the Service Account resource
func (rrn RelativeResourceNamer) ResourceName(sa *v1alpha1.ServiceAccount) string {
	return fmt.Sprintf("projects/%s/serviceAccounts/%s@%s.iam.gserviceaccount.com",
		rrn.projectName, meta.GetExternalName(sa), rrn.projectName)
}
