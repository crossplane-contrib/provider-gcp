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

package iam

import (
	"context"
	"encoding/base64"
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

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	scv1alpha1 "github.com/crossplane/provider-gcp/apis/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/serviceaccountkey"
)

// Error messages
const (
	errNotServiceAccountKey    = "managed resource is not a GCP ServiceAccountKey"
	errGetServiceAccountKey    = "cannot get GCP ServiceAccountKey object via IAM API"
	errCreateServiceAccountKey = "cannot create GCP ServiceAccountKey object via IAM API"
	errDeleteServiceAccountKey = "cannot delete GCP ServiceAccountKey object via IAM API"
	errDecodePrivateKey        = "cannot decode private key"
	errDecodePublicKey         = "cannot decode public key"
)

const (
	// Format string for the relative resource names of ServiceAccountKeys
	// built upon relative resource names of ServiceAccounts. For example
	// projects/<project-name>/serviceAccounts/<service-account-email>/keys/<key-id>
	fmtKeyRelativeResourceName = "%s/keys/%s"

	// connection detail keys
	keyPrivateKeyType = "privateKeyType"
	keyPrivateKeyData = "privateKey"
	keyPublicKeyType  = "publicKeyType"
	keyPublicKeyData  = "publicKey"
)

// SetupServiceAccountKey adds a controller that reconciles ServiceAccountKeys.
func SetupServiceAccountKey(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ServiceAccountKeyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	enableFeature := true
	if enableFeature {
		cps = append(cps, connection.NewManager(mgr.GetClient(), mgr.GetScheme(), connection.StoreConfigKind(scv1alpha1.StoreConfigGroupVersionKind)))
	}
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ServiceAccountKeyGroupVersionKind),
		managed.WithInitializers(),
		managed.WithExternalConnecter(&serviceAccountKeyServiceConnector{client: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.ServiceAccountKey{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type serviceAccountKeyServiceConnector struct {
	client client.Client
}

// Connect sets up SA key external client using credentials from the provider
func (c *serviceAccountKeyServiceConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, opts, err := gcp.GetAuthInfo(ctx, c.client, mg)

	if err != nil {
		return nil, err
	}

	s, err := iamv1.NewService(ctx, opts)

	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &serviceAccountKeyExternalClient{
			serviceAccountKeyClient: s.Projects.ServiceAccounts.Keys,
		},
		errors.Wrap(err, errNewClient)
}

type serviceAccountKeyExternalClient struct {
	serviceAccountKeyClient serviceaccountkey.Client
}

func (s *serviceAccountKeyExternalClient) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccountKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotServiceAccountKey)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	getCall := s.serviceAccountKeyClient.Get(resourcePath(cr))
	if cr.Spec.ForProvider.PublicKeyType != nil && *cr.Spec.ForProvider.PublicKeyType != "" {
		getCall = getCall.PublicKeyType(*cr.Spec.ForProvider.PublicKeyType)
	}

	fromProvider, err := getCall.Context(ctx).Do()
	if err != nil {
		// This API appears to return an HTTP 403 forbidden for some
		// period of time immediately after a service account has been
		// deleted. It should be safe to ignore this error and assume
		// that the key doesn't exist. Presumably if the error was real
		// (i.e. the key does exist but we don't have permission to read
		// it) we'd get the same error at Create time.
		return managed.ExternalObservation{}, errors.Wrap(resource.IgnoreAny(err, gcp.IsErrorNotFound, gcp.IsErrorForbidden), errGetServiceAccountKey)
	}

	if err := serviceaccountkey.PopulateSaKey(cr, fromProvider); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetServiceAccountKey)
	}

	cr.Status.SetConditions(xpv1.Available())

	connDetails, err := getConnectionDetails(cr.Spec.ForProvider.PublicKeyType, fromProvider)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetServiceAccountKey)
	}

	return managed.ExternalObservation{
		ResourceExists: true,
		// all service account key parameters are immutable, no update method exists in Google Cloud API for SA keys
		ResourceUpToDate:  true,
		ConnectionDetails: connDetails,
	}, nil
}

// Create https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts.keys/create
func (s *serviceAccountKeyExternalClient) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccountKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotServiceAccountKey)
	}

	// Technically ServiceAccount can be nil, but reference resolution
	// should always make sure a value is set before we get to this point.
	req := s.serviceAccountKeyClient.Create(gcp.StringValue(cr.Spec.ForProvider.ServiceAccount), &iamv1.CreateServiceAccountKeyRequest{
		KeyAlgorithm:   gcp.StringValue(cr.Spec.ForProvider.KeyAlgorithm),
		PrivateKeyType: gcp.StringValue(cr.Spec.ForProvider.PrivateKeyType),
	})

	fromProvider, err := req.Context(ctx).Do()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateServiceAccountKey)
	}
	connDetails, err := getConnectionDetails(cr.Spec.ForProvider.PublicKeyType, fromProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateServiceAccountKey)
	}
	keyID, err := serviceaccountkey.ParseKeyIDFromRrn(fromProvider.Name)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateServiceAccountKey)
	}

	meta.SetExternalName(cr, keyID) // set external name to key id parsing it from Google Cloud API relative resource name

	return managed.ExternalCreation{ExternalNameAssigned: true, ConnectionDetails: connDetails}, nil
}

func (s *serviceAccountKeyExternalClient) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// ServiceAccountKeys are immutable, i.e.,GCP IAM Rest API does not provide an update method:
	// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts.keys
	return managed.ExternalUpdate{}, nil
}

func (s *serviceAccountKeyExternalClient) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ServiceAccountKey)
	if !ok {
		return errors.New(errNotServiceAccountKey)
	}

	_, err := s.serviceAccountKeyClient.Delete(resourcePath(cr)).Context(ctx).Do()
	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteServiceAccountKey)
}

// resourcePath yields the Google Cloud API relative resource name for the ServiceAccountKey resource
func resourcePath(saKey *v1alpha1.ServiceAccountKey) string {
	// Technically ServiceAccount can be nil, but reference resolution
	// should always make sure a value is set before we get to this point.
	// Similarly, we always make sure the external name is set before this
	// function is called.
	return fmt.Sprintf(fmtKeyRelativeResourceName, gcp.StringValue(saKey.Spec.ForProvider.ServiceAccount), meta.GetExternalName(saKey))
}

func getConnectionDetails(publicKeyType *string, fromProvider *iamv1.ServiceAccountKey) (managed.ConnectionDetails, error) {
	result := make(map[string][]byte, 4)

	if fromProvider.PublicKeyData != "" {
		d, err := base64.StdEncoding.DecodeString(fromProvider.PublicKeyData)
		if err != nil {
			return nil, errors.Wrap(err, errDecodePublicKey)
		}
		result[keyPublicKeyData] = d

		// only provided optionally in keys.get responses
		if publicKeyType != nil {
			result[keyPublicKeyType] = []byte(*publicKeyType)
		}
	}

	// only provided in keys.create responses
	if fromProvider.PrivateKeyData != "" {
		d, err := base64.StdEncoding.DecodeString(fromProvider.PrivateKeyData)
		if err != nil {
			return nil, errors.Wrap(err, errDecodePrivateKey)
		}
		result[keyPrivateKeyData] = d

		// only provided in keys.create responses
		result[keyPrivateKeyType] = []byte(fromProvider.PrivateKeyType)
	}

	return result, nil
}
