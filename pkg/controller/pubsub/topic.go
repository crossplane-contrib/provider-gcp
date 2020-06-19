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
	"fmt"

	pubsubgo "cloud.google.com/go/pubsub"
	pubsub "cloud.google.com/go/pubsub/apiv1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/pubsub/v1alpha1"
	gcpv1alpha3 "github.com/crossplane/provider-gcp/apis/v1alpha3"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/topic"
)

const (
	errGetProvider       = "cannot get Provider"
	errProviderSecretRef = "cannot find Secret reference on Provider"
	errGetProviderSecret = "cannot get Provider Secret"

	errNotTopic        = "managed resource is not of type Topic"
	errClientCreate    = "cannot create client"
	errGetTopic        = "cannot get Topic"
	errUpdateTopic     = "cannot update Topic"
	errKubeUpdateTopic = "cannot update Topic custom resource"
	errCreateTopic     = "cannot create Topic"
	errDeleteTopic     = "cannot delete Topic"
)

// SetupTopic adds a controller that reconciles Topics.
func SetupTopic(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.TopicGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Topic{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.TopicGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient(), newPubSubClient: pubsub.NewPublisherClient}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client          client.Client
	newPubSubClient func(ctx context.Context, opts ...option.ClientOption) (*pubsub.PublisherClient, error)
}

// Connect returns an ExternalClient with necessary information to talk to GCP API.
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return nil, errors.New(errNotTopic)
	}
	p := &gcpv1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretRef)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	ps, err := c.newPubSubClient(ctx,
		option.WithCredentialsJSON(s.Data[p.Spec.CredentialsSecretRef.Key]),
		option.WithScopes(pubsubgo.ScopePubSub))
	if err != nil {
		return nil, errors.Wrap(err, errClientCreate)
	}
	return &external{projectID: p.Spec.ProjectID, client: c.client, ps: ps}, nil
}

type external struct {
	projectID string
	client    client.Client
	ps        *pubsub.PublisherClient
}

// Observe makes observation about the external resource.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotTopic)
	}
	t, err := e.ps.GetTopic(ctx, &pubsubpb.GetTopicRequest{Topic: fmt.Sprintf("projects/%s/topics/%s", e.projectID, meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), errGetTopic)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	topic.LateInitialize(&cr.Spec.ForProvider, *t)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.client.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateTopic)
		}
	}
	cr.SetConditions(runtimev1alpha1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: topic.IsUpToDate(cr.Spec.ForProvider, *t),
		ConnectionDetails: managed.ConnectionDetails{
			"topic": []byte(meta.GetExternalName(cr)),
		},
	}, nil
}

// Create initiates creation of external resource.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotTopic)
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	if _, err := e.ps.CreateTopic(ctx, topic.GenerateTopic(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider)); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateTopic)
	}
	return managed.ExternalCreation{}, nil
}

// Update initiates an update to the external resource.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotTopic)
	}

	t, err := e.ps.GetTopic(ctx, &pubsubpb.GetTopicRequest{Topic: fmt.Sprintf("projects/%s/topics/%s", e.projectID, meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetTopic)
	}

	_, err = e.ps.UpdateTopic(ctx, topic.GenerateUpdateRequest(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider, *t))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateTopic)
}

// Delete initiates an deletion of the external resource.
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return errors.New(errNotTopic)
	}
	err := e.ps.DeleteTopic(ctx, &pubsubpb.DeleteTopicRequest{Topic: fmt.Sprintf("projects/%s/topics/%s", e.projectID, meta.GetExternalName(cr))})
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), errDeleteTopic)
}
