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
	"time"

	iamv1 "google.golang.org/api/iam/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/serviceaccountkey"
)

const (
	// error messages
	errNotServiceAccountKey    = "managed resource is not a GCP ServiceAccountKey"
	errGetServiceAccountKey    = "cannot get GCP ServiceAccountKey object via IAM API"
	errCreateServiceAccountKey = "cannot create GCP ServiceAccountKey object via IAM API"
	errDeleteServiceAccountKey = "cannot delete GCP ServiceAccountKey object via IAM API"
	errNoExternalName          = "empty external name"
	// format string for the relative resource names of ServiceAccountKeys built upon rrns of ServiceAccounts
	fmtKeyRelativeResourceName = "%s/keys/%s"
	// format string for invalid service account reference errors
	fmtErrInvalidServiceAccountRef = "invalid service account reference: %v"
	// connection detail keys
	keyPrivateKeyType = "privateKeyType"
	keyPrivateKeyData = "privateKey"
	keyPublicKeyType  = "publicKeyType"
	keyPublicKeyData  = "publicKey"
)

// SetupServiceAccountKey adds a controller that reconciles ServiceAccountKeys.
func SetupServiceAccountKey(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.ServiceAccountKeyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.ServiceAccountKey{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ServiceAccountKeyGroupVersionKind),
			managed.WithInitializers(),
			managed.WithExternalConnecter(&serviceAccountKeyServiceConnector{client: mgr.GetClient()}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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

func (s *serviceAccountKeyExternalClient) Observe(ctx context.Context,
	mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccountKey)

	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotServiceAccountKey)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	rrn, err := resourcePath(cr)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetServiceAccountKey)
	}

	getCall := s.serviceAccountKeyClient.Get(rrn)

	if cr.Spec.ForProvider.PublicKeyType != nil && *cr.Spec.ForProvider.PublicKeyType != "" {
		getCall = getCall.PublicKeyType(*cr.Spec.ForProvider.PublicKeyType)
	}

	fromProvider, err := getCall.Context(ctx).Do()

	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetServiceAccountKey)
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
func (s *serviceAccountKeyExternalClient) Create(ctx context.Context,
	mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ServiceAccountKey)

	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotServiceAccountKey)
	}

	cr.SetConditions(xpv1.Creating())
	// The first parameter to the Create method is the resource name of the GCP service account
	// which the service account key belongs to. Retry resolution because reference could have been altered
	// between invocations of Observe and Create
	saPath, err := referencedServiceAccountPath(cr)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateServiceAccountKey)
	}

	req := s.serviceAccountKeyClient.Create(saPath, &iamv1.CreateServiceAccountKeyRequest{
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

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
		ConnectionDetails:    connDetails,
	}, nil
}

func (s *serviceAccountKeyExternalClient) Update(_ context.Context,
	_ resource.Managed) (managed.ExternalUpdate, error) {
	// ServiceAccountKeys are immutable, i.e.,GCP IAM Rest API does not provide an update method:
	// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts.keys
	return managed.ExternalUpdate{}, nil
}

func (s *serviceAccountKeyExternalClient) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ServiceAccountKey)

	if !ok {
		return errors.New(errNotServiceAccountKey)
	}

	rrn, err := resourcePath(cr)

	if err != nil {
		return errors.Wrap(err, errDeleteServiceAccountKey)
	}

	_, err = s.serviceAccountKeyClient.Delete(rrn).Context(ctx).Do()

	return errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errDeleteServiceAccountKey)
}

// referencedServiceAccountPath returns the external name of the service account which saKey belongs to
func referencedServiceAccountPath(saKey *v1alpha1.ServiceAccountKey) (string, error) {
	// we assume Google Cloud API relative resource name (aka resource path) of the ServiceAccount is stored in spec.ForProvider.ServiceAccount
	if gcp.StringValue(saKey.Spec.ForProvider.ServiceAccount) == "" {
		return "", fmt.Errorf(fmtErrInvalidServiceAccountRef, saKey.Spec.ForProvider.ServiceAccountReferer)
	}

	return *saKey.Spec.ForProvider.ServiceAccount, nil
}

// resourcePath yields the Google Cloud API relative resource name for the ServiceAccountKey resource
//   returns <the relative resource name>, <whether the external name annotation is non-empty>, <error encountered during resolution>
func resourcePath(saKey *v1alpha1.ServiceAccountKey) (string, error) {
	if saPath, err := referencedServiceAccountPath(saKey); err != nil {
		return "", err
	} else if extName := meta.GetExternalName(saKey); extName == "" {
		return "", errors.New(errNoExternalName)
	} else {
		return fmt.Sprintf(fmtKeyRelativeResourceName, saPath, extName), nil
	}
}

func getConnectionDetails(publicKeyType *string, fromProvider *iamv1.ServiceAccountKey) (result managed.ConnectionDetails, err error) {
	result = make(map[string][]byte, 4)

	if fromProvider.PublicKeyData != "" {
		if result[keyPublicKeyData], err = base64.StdEncoding.DecodeString(fromProvider.PublicKeyData); err != nil {
			return
		}
		// only provided optionally in keys.get responses
		if publicKeyType != nil {
			result[keyPublicKeyType] = []byte(*publicKeyType)
		}
	}

	// only provided in keys.create responses
	if fromProvider.PrivateKeyData != "" {
		if result[keyPrivateKeyData], err = base64.StdEncoding.DecodeString(fromProvider.PrivateKeyData); err != nil {
			return
		}
		// only provided in keys.create responses
		result[keyPrivateKeyType] = []byte(fromProvider.PrivateKeyType)
	}

	return
}
