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

package database

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"

	"github.com/crossplaneio/stack-gcp/apis/database/v1alpha2"
	apisv1alpha2 "github.com/crossplaneio/stack-gcp/apis/v1alpha2"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	"github.com/crossplaneio/stack-gcp/pkg/clients/cloudsql"
)

const (
	errNotCloudsql                = "managed resource is not a Cloudsql resource"
	errProviderNotRetrieved       = "provider could not be retrieved"
	errProviderSecretNotRetrieved = "secret referred in provider could not be retrieved"
	errNewClient                  = "cannot create new Sqladmin Service"
	errInsertFailed               = "cannot insert new Cloudsql instance"
	errDeleteFailed               = "cannot delete the Cloudsql instance"
	errUpdateFailed               = "cannot update the Cloudsql instance"
	errGetFailed                  = "cannot get the Cloudsql instance"
)

// CloudsqlInstanceController is the controller for Cloudsql CRD.
type CloudsqlInstanceController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields
// on the Controller and Start it when the Manager is Started.
func (c *CloudsqlInstanceController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1alpha2.CloudsqlInstanceGroupVersionKind),
		resource.WithExternalConnecter(&cloudsqlConnector{kube: mgr.GetClient()}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha2.CloudsqlInstanceKindAPIVersion, v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha2.CloudsqlInstance{}).
		Complete(r)
}

type cloudsqlConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*sqladmin.Service, error)
}

func (c *cloudsqlConnector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	cr, ok := mg.(*v1alpha2.CloudsqlInstance)
	if !ok {
		return nil, errors.New(errNotCloudsql)
	}

	provider := &apisv1alpha2.Provider{}
	n := meta.NamespacedNameOf(cr.Spec.ProviderReference)
	if err := c.kube.Get(ctx, n, provider); err != nil {
		return nil, errors.Wrap(err, errProviderNotRetrieved)
	}
	secret := &v1.Secret{}
	name := meta.NamespacedNameOf(&v1.ObjectReference{
		Name:      provider.Spec.Secret.Name,
		Namespace: provider.Namespace,
	})
	if err := c.kube.Get(ctx, name, secret); err != nil {
		return nil, errors.Wrap(err, errProviderSecretNotRetrieved)
	}

	if c.newServiceFn == nil {
		c.newServiceFn = sqladmin.NewService
	}
	s, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(secret.Data[provider.Spec.Secret.Key]),
		option.WithScopes(sqladmin.SqlserviceAdminScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &cloudsqlExternal{kube: c.kube, db: s.Instances, user: s.Users, projectID: provider.Spec.ProjectID}, nil
}

type cloudsqlExternal struct {
	kube      client.Client
	db        *sqladmin.InstancesService
	user      *sqladmin.UsersService
	projectID string
}

func (c *cloudsqlExternal) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha2.CloudsqlInstance)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotCloudsql)
	}
	if err := c.EnsureExternalNameAnnotation(ctx, cr); err != nil {
		return resource.ExternalObservation{}, err
	}
	instance, err := c.db.Get(c.projectID, cr.Annotations[v1alpha1.ExternalNameAnnotationKey]).Context(ctx).Do()
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetFailed)
	}
	cr.Status.AtProvider = cloudsql.GenerateObservation(*instance)
	if cloudsql.FillSpecWithDefaults(&cr.Spec.ForProvider, *instance) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return resource.ExternalObservation{ResourceExists: true}, errors.Wrap(err, "cannot update CloudsqlInstance CR")
		}
	}
	upToDate := cloudsql.IsUpToDate(cr.Spec.ForProvider, *instance)
	var conn resource.ConnectionDetails
	switch cr.Status.AtProvider.State {
	case v1alpha2.StateRunnable:
		cr.Status.SetConditions(v1alpha1.Available())
		if cr.Status.Phase != v1alpha1.BindingPhaseBound {
			resource.SetBindable(cr)
		}
		conn, err = c.getConnectionDetails(ctx, cr)
		if err != nil {
			return resource.ExternalObservation{ResourceExists: true, ResourceUpToDate: upToDate},
			errors.Wrap(err, "cannot get connection details")
		}
		if err := c.updateRootCredentials(ctx, cr, string(conn[v1alpha1.ResourceCredentialsSecretPasswordKey])); err != nil {
			return resource.ExternalObservation{ResourceExists: true, ResourceUpToDate: upToDate},
			errors.Wrap(err, "cannot update root user credentials")
		}
	case v1alpha2.StateCreating:
		cr.Status.SetConditions(v1alpha1.Creating())
	case v1alpha2.StateCreationFailed, v1alpha2.StateSuspended, v1alpha2.StateMaintenance, v1alpha2.StateUnknownState:
		cr.Status.SetConditions(v1alpha1.Unavailable())
	}
	return resource.ExternalObservation{ResourceExists: true, ResourceUpToDate: upToDate, ConnectionDetails: conn}, nil
}

func (c *cloudsqlExternal) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha2.CloudsqlInstance)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotCloudsql)
	}
	// This pre-reconcile hook logic will be moved to crossplane-runtime
	if err := c.EnsureExternalNameAnnotation(ctx, cr); err != nil {
		return resource.ExternalCreation{}, err
	}
	instance := cloudsql.GenerateDatabaseInstance(cr.Spec.ForProvider, cr.Annotations[v1alpha1.ExternalNameAnnotationKey])
	_, err := c.db.Insert(c.projectID, instance).Context(ctx).Do()
	if err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errInsertFailed)
	}
	return resource.ExternalCreation{}, errors.Wrap(err, errInsertFailed)
}

func (c *cloudsqlExternal) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha2.CloudsqlInstance)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotCloudsql)
	}
	instance := cloudsql.GenerateDatabaseInstance(cr.Spec.ForProvider, cr.Annotations[v1alpha1.ExternalNameAnnotationKey])
	_, err := c.db.Patch(c.projectID, cr.Annotations[v1alpha1.ExternalNameAnnotationKey], instance).Context(ctx).Do()
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (c *cloudsqlExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha2.CloudsqlInstance)
	if !ok {
		return errors.New(errNotCloudsql)
	}
	if err := c.EnsureExternalNameAnnotation(ctx, cr); err != nil {
		return err
	}
	_, err := c.db.Delete(c.projectID, cr.Annotations[v1alpha1.ExternalNameAnnotationKey]).Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errDeleteFailed)
}

// TODO(muvaf): consider implementing a set of configurators for naming.
func (c *cloudsqlExternal) EnsureExternalNameAnnotation(ctx context.Context, cr *v1alpha2.CloudsqlInstance) error {
	if cr.Annotations[v1alpha1.ExternalNameAnnotationKey] == "" {
		if cr.Annotations == nil {
			cr.Annotations = make(map[string]string)
		}
		cr.Annotations[v1alpha1.ExternalNameAnnotationKey] = cr.Name
		if err := c.kube.Update(ctx, cr); err != nil {
			return err
		}
	}
	return nil
}

func (c *cloudsqlExternal) getConnectionDetails(ctx context.Context, cr *v1alpha2.CloudsqlInstance) (resource.ConnectionDetails, error) {
	password, err := util.GeneratePassword(v1alpha2.PasswordLength)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate password")
	}
	s := &v1.Secret{}
	name := types.NamespacedName{Name: cr.Spec.WriteConnectionSecretToReference.Name, Namespace: cr.Namespace}
	if err := c.kube.Get(ctx, name, s); !k8serrors.IsNotFound(err) && err != nil {
		return nil, errors.Wrap(err, "connection secret could not be retrieved")
	}
	if s.Data != nil && len(s.Data[v1alpha1.ResourceCredentialsSecretPasswordKey]) != 0 {
		password = string(s.Data[v1alpha1.ResourceCredentialsSecretPasswordKey])
	}
	m := map[string][]byte{
		v1alpha1.ResourceCredentialsSecretUserKey:     []byte(cr.DatabaseUserName()),
		v1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password),
	}
	endpoint := ""
	// TODO(muvaf): There might be cases where more than 1 private and/or public IP address has been assigned. We should
	// somehow show all addresses that are possible to use.
	for _, ip := range cr.Status.AtProvider.IPAddresses {
		if ip.Type == v1alpha2.PrivateIPType {
			m[v1alpha2.PrivateIPKey] = []byte(ip.IPAddress)
			// TODO(muvaf): we explicitly enforce use of private IP if it's available. But this should be configured
			// by resource class or claim.
			endpoint = ip.IPAddress
		}
		if ip.Type == v1alpha2.PublicIPType {
			m[v1alpha2.PublicIPKey] = []byte(ip.IPAddress)
			if endpoint == "" {
				endpoint = ip.IPAddress
			}
		}
	}
	m[v1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(endpoint)
	return m, nil
}

func (c *cloudsqlExternal) updateRootCredentials(ctx context.Context, cr *v1alpha2.CloudsqlInstance, password string) error {
	users, err := c.user.List(c.projectID, cr.Annotations[v1alpha1.ExternalNameAnnotationKey]).Context(ctx).Do()
	if err != nil {
		return err
	}
	var rootUser *sqladmin.User
	for _, val := range users.Items {
		if val.Name == cr.DatabaseUserName() {
			rootUser = val
			break
		}
	}
	if rootUser == nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("user: %s is not found", cr.DatabaseUserName()),
		}
	}
	rootUser.Password = password
	_, err = c.user.Update(c.projectID, cr.Annotations[v1alpha1.ExternalNameAnnotationKey], rootUser.Name, rootUser).
		Host(rootUser.Host).
		Context(ctx).
		Do()
	return errors.Wrap(err, "cannot update root user credentials")
}
