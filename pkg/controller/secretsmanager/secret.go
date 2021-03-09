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

package secretsmanager

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	sm "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
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

	"github.com/crossplane/provider-gcp/apis/secretsmanager/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/secret"
)

const (
	errNotSecret        = "managed resource is not of type Secret"
	errNewClient        = "cannot create client"
	errGetSecret        = "cannot get Secret"
	errUpdateSecret     = "cannot update Secret"
	errKubeUpdateSecret = "cannot update Secret custom resource"
	errCreateSecret     = "cannot create Secret"

	errDeleteSecret = "cannot delete Secret"
)

// SetupSecret adds a controller that reconciles Secrets.
func SetupSecret(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.SecretGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.Secret{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SecretGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client client.Client
}

// Connect returns an ExternalClient with necessary information to talk to GCP API.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	s, err := secretmanager.NewClient(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &external{projectID: projectID, client: c.client, sc: s}, nil
}

type external struct {
	projectID string
	client    client.Client
	sc        secret.Client
}

// Observe makes observation about the external resource.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Secret)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSecret)
	}
	s, err := e.sc.GetSecret(ctx, &sm.GetSecretRequest{Name: fmt.Sprintf("projects/%s/secrets/%s", e.projectID, meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), errGetSecret)
	}

	o := secret.Observation{SecretID: meta.GetExternalName(cr)}
	if o.SecretID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	o.CreateTime = s.CreateTime.String()
	if o.CreateTime == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	secret.LateInitialize(&cr.Spec.ForProvider, *s)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.client.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateSecret)
		}
	}

	secret.UpdateStatus(&cr.Status, o)

	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: secret.IsUpToDate(cr.Spec.ForProvider, *s),
		ConnectionDetails: managed.ConnectionDetails{
			v1alpha1.ConnectionSecretKeyName:        []byte(meta.GetExternalName(cr)),
			v1alpha1.ConnectionSecretKeyProjectName: []byte(e.projectID),
		},
	}, nil
}

// Create initiates creation of external resource.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Secret)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSecret)
	}
	cr.SetConditions(xpv1.Creating())
	_, err := e.sc.CreateSecret(ctx, secret.NewCreateSecretRequest(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateSecret)
}

// Update initiates an update to the external resource.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Secret)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSecret)
	}

	s, err := e.sc.GetSecret(ctx, &sm.GetSecretRequest{Name: fmt.Sprintf("projects/%s/secrets/%s", e.projectID, meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetSecret)
	}

	_, err = e.sc.UpdateSecret(ctx, secret.GenerateUpdateRequest(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider, *s))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSecret)
}

// Delete initiates an deletion of the external resource.
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Secret)
	if !ok {
		return errors.New(errNotSecret)
	}
	err := e.sc.DeleteSecret(ctx, &sm.DeleteSecretRequest{Name: fmt.Sprintf("projects/%s/secrets/%s", e.projectID, meta.GetExternalName(cr))})
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), errDeleteSecret)
}
