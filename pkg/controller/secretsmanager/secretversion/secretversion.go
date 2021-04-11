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

package secretversion

import (
	"context"
	"fmt"
	"strconv"

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

	"github.com/crossplane/provider-gcp/apis/secretsmanager/secretversion/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/secretsmanager/secretversion"
)

const (
	errNotSecretVersion        = "managed resource is not of type Secret Version"
	errNewClient               = "cannot create client"
	errGetSecretVersion        = "cannot get Secret Version"
	errUpdateSecretVersion     = "cannot update Secret Version"
	errKubeUpdateSecretVersion = "cannot update Secret Version custom resource"
	errCreateSecretVersion     = "cannot create Secret Version"
	errGetSecretPayload        = "cannot get secret payload"

	errDeleteSecretVersion = "cannot delete Secret Version"
)

// SetupSecretVersion adds a controller that reconciles Secret versions.
func SetupSecretVersion(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.SecretVersionGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.SecretVersion{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SecretVersionGroupVersionKind),
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
	sc        secretversion.Client
}

// Observe makes observation about the external resource.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.SecretVersion)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSecretVersion)
	}

	version, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetSecretVersion)
	}
	s, err := e.sc.GetSecretVersion(ctx, &sm.GetSecretVersionRequest{Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%d", e.projectID, cr.Spec.ForProvider.SecretRef, version)})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), errGetSecretVersion)
	}

	data, err := e.sc.AccessSecretVersion(ctx, &sm.AccessSecretVersionRequest{Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%d", e.projectID, cr.Spec.ForProvider.SecretRef, version)})
	if err != nil {
		preconditionErr := resource.Ignore(gcp.IsFailedPreCondition, err)
		if preconditionErr != nil {
			return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsFailedPreCondition, err), errGetSecretPayload)
		}
	}
	o := secretversion.Observation{}

	o.CreateTime = s.CreateTime.String()
	if o.CreateTime == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if s.GetDestroyTime() != nil {
		o.DestroyTime = s.CreateTime.String()
	}

	o.State = v1alpha1.SecretVersionState(s.State)

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if data == nil {
		secretversion.LateInitialize(&cr.Spec.ForProvider, s, nil, cr.Spec.ForProvider.SecretRef)
	} else {
		secretversion.LateInitialize(&cr.Spec.ForProvider, s, data.Payload.Data, cr.Spec.ForProvider.SecretRef)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.client.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateSecretVersion)
		}
	}

	secretversion.UpdateStatus(&cr.Status, o)

	cr.SetConditions(xpv1.Available())

	isSynced := secretversion.IsUpToDate(cr.Spec.ForProvider, s)
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: isSynced,
	}, nil

}

// Create initiates creation of external resource.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SecretVersion)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSecretVersion)
	}
	cr.SetConditions(xpv1.Creating())
	_, err := e.sc.AddSecretVersion(ctx, secretversion.NewAddSecretVersionRequest(e.projectID, cr.Spec.ForProvider))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateSecretVersion)
}

// Update initiates an update to the external resource.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.SecretVersion)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSecretVersion)
	}

	version, err := strconv.Atoi(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetSecretVersion)
	}

	s, err := e.sc.GetSecretVersion(ctx, &sm.GetSecretVersionRequest{Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%d", e.projectID, cr.Spec.ForProvider.SecretRef, version)})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetSecretVersion)
	}

	if val, ok := v1alpha1.SecretVersionStateValue[cr.Spec.ForProvider.DesiredSecretVersionState]; ok {
		if v1alpha1.SecretVersionState(val) == v1alpha1.SecretVersionEnabled {
			if s.GetState() != sm.SecretVersion_State(v1alpha1.SecretVersionEnabled) || s.GetState() != sm.SecretVersion_State(v1alpha1.SecretVersionDestroyed) {
				_, err = e.sc.EnableSecretVersion(ctx, &sm.EnableSecretVersionRequest{Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%d", e.projectID, cr.Spec.ForProvider.SecretRef, version)})
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSecretVersion)
				}
			}
		}
	}

	if val, ok := v1alpha1.SecretVersionStateValue[cr.Spec.ForProvider.DesiredSecretVersionState]; ok {
		if v1alpha1.SecretVersionState(val) == v1alpha1.SecretVersionDisabled {
			if s.GetState() != sm.SecretVersion_State(v1alpha1.SecretVersionDisabled) || s.GetState() != sm.SecretVersion_State(v1alpha1.SecretVersionDestroyed) {
				_, err = e.sc.DisableSecretVersion(ctx, &sm.DisableSecretVersionRequest{Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%d", e.projectID, cr.Spec.ForProvider.SecretRef, version)})
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSecretVersion)
				}
			}
		}
	}

	if val, ok := v1alpha1.SecretVersionStateValue[cr.Spec.ForProvider.DesiredSecretVersionState]; ok {
		if v1alpha1.SecretVersionState(val) == v1alpha1.SecretVersionDestroyed {
			if s.GetState() != sm.SecretVersion_State(v1alpha1.SecretVersionDestroyed) {
				_, err = e.sc.DestroySecretVersion(ctx, &sm.DestroySecretVersionRequest{Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%d", e.projectID, cr.Spec.ForProvider.SecretRef, version)})
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSecretVersion)
				}
			}
		}
	}

	return managed.ExternalUpdate{}, nil

}

// Delete initiates an deletion of the external resource.
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.SecretVersion)
	if !ok {
		return errors.New(errNotSecretVersion)
	}
	_, err := e.sc.DestroySecretVersion(ctx, &sm.DestroySecretVersionRequest{Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", e.projectID, cr.Spec.ForProvider.SecretRef, meta.GetExternalName(cr))})
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), errDeleteSecretVersion)
}
