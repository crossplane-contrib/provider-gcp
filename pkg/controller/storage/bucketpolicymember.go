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

package storage

import (
	"context"

	"google.golang.org/api/storage/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	iamv1alpha1 "github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane/provider-gcp/apis/storage/v1alpha1"
	scv1alpha1 "github.com/crossplane/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/bucketpolicy"
	"github.com/crossplane/provider-gcp/pkg/features"
)

const (
	errNotBucketPolicyMember = "managed resource is not a GCP BucketPolicyMember"
)

// SetupBucketPolicyMember adds a controller that reconciles BucketPolicyMembers.
func SetupBucketPolicyMember(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.BucketPolicyMemberGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), scv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.BucketPolicyMemberGroupVersionKind),
		managed.WithExternalConnecter(&bucketPolicyMemberConnecter{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.BucketPolicyMember{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type bucketPolicyMemberConnecter struct {
	client client.Client
}

// Connect sets up iam client using credentials from the provider
func (c *bucketPolicyMemberConnecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := storage.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &bucketPolicyMemberExternal{kube: c.client, bucketpolicy: storage.NewBucketsService(s)}, nil
}

type bucketPolicyMemberExternal struct {
	kube         client.Client
	bucketpolicy bucketpolicy.Client
}

func (e *bucketPolicyMemberExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.BucketPolicyMember)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBucketPolicyMember)
	}

	instance, err := e.bucketpolicy.GetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.Bucket)).OptionsRequestedPolicyVersion(iamv1alpha1.PolicyVersion).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetPolicy)
	}

	changed := bucketpolicy.BindRoleToMember(cr.Spec.ForProvider, instance)
	if !changed {
		cr.Status.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	return managed.ExternalObservation{}, nil
}

func (e *bucketPolicyMemberExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.BucketPolicyMember)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBucketPolicyMember)
	}
	instance, err := e.bucketpolicy.GetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.Bucket)).OptionsRequestedPolicyVersion(iamv1alpha1.PolicyVersion).Context(ctx).Do()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGetPolicy)
	}

	changed := bucketpolicy.BindRoleToMember(cr.Spec.ForProvider, instance)
	if !changed {
		return managed.ExternalCreation{}, nil
	}

	if _, err := e.bucketpolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.Bucket), instance).
		Context(ctx).Do(); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSetPolicy)
	}

	return managed.ExternalCreation{}, nil
}

func (e *bucketPolicyMemberExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, err := e.Create(ctx, mg)
	return managed.ExternalUpdate{}, err
}

func (e *bucketPolicyMemberExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.BucketPolicyMember)
	if !ok {
		return errors.New(errNotBucketPolicyMember)
	}
	instance, err := e.bucketpolicy.GetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.Bucket)).OptionsRequestedPolicyVersion(iamv1alpha1.PolicyVersion).Context(ctx).Do()
	if err != nil {
		return errors.Wrap(err, errGetPolicy)
	}

	changed := bucketpolicy.UnbindRoleFromMember(cr.Spec.ForProvider, instance)
	if !changed {
		return nil
	}
	if _, err := e.bucketpolicy.SetIamPolicy(gcp.StringValue(cr.Spec.ForProvider.Bucket), instance).
		Context(ctx).Do(); err != nil {
		return errors.Wrap(err, errSetPolicy)
	}

	return nil
}
