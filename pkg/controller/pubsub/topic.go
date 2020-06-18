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

	pubsub2 "cloud.google.com/go/pubsub"
	pubsub "cloud.google.com/go/pubsub/apiv1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	pubsub3 "google.golang.org/genproto/googleapis/pubsub/v1"
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
	"github.com/crossplane/provider-gcp/pkg/clients/topic"
)

const (
	errGetProvider       = "cannot get Provider"
	errProviderSecretRef = "cannot find Secret reference on Provider"
	errGetProviderSecret = "cannot get Provider Secret"
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

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return nil, errors.New("managed resource is not of type Topic")
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
		option.WithScopes(pubsub2.ScopePubSub))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create client")
	}
	return &external{projectID: p.Spec.ProjectID, client: c.client, ps: ps}, nil
}

// A external does nothing.
type external struct {
	projectID string
	client    client.Client
	ps        *pubsub.PublisherClient
}

// Observe does nothing. It returns an empty ExternalObservation and no error.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalObservation{}, errors.New("managed resource is not of type Topic")
	}
	t, err := e.ps.GetTopic(ctx, &pubsub3.GetTopicRequest{Topic: fmt.Sprintf("projects/%s/topics/%s", e.projectID, meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), "cannot get Topic")
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	topic.LateInitialize(&cr.Spec.ForProvider, *t)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := e.client.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "cannot update Topic")
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

// Create does nothing. It returns an empty ExternalCreation and no error.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalCreation{}, errors.New("managed resource is not of type Topic")
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	if _, err := e.ps.CreateTopic(ctx, topic.GenerateTopic(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider)); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create Topic")
	}
	return managed.ExternalCreation{}, nil
}

// Update does nothing. It returns an empty ExternalUpdate and no error.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalUpdate{}, errors.New("managed resource is not of type Topic")
	}

	t, err := e.ps.GetTopic(ctx, &pubsub3.GetTopicRequest{Topic: fmt.Sprintf("projects/%s/topics/%s", e.projectID, meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot get Topic")
	}

	_, err = e.ps.UpdateTopic(ctx, topic.GenerateUpdateRequest(e.projectID, meta.GetExternalName(cr), cr.Spec.ForProvider, *t))
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update Topic")
}

// Delete does nothing. It never returns an error.
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return errors.New("managed resource is not of type Topic")
	}
	err := e.ps.DeleteTopic(ctx, &pubsub3.DeleteTopicRequest{Topic: fmt.Sprintf("projects/%s/topics/%s", e.projectID, meta.GetExternalName(cr))})
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFoundGRPC, err), "cannot delete Topic")
}
