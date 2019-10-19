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
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"

	"github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
	apisv1alpha2 "github.com/crossplaneio/stack-gcp/apis/v1alpha2"
	gcp "github.com/crossplaneio/stack-gcp/pkg/clients"
	"github.com/crossplaneio/stack-gcp/pkg/clients/cloudsql"
)

const (
	errNotCloudsql                = "managed resource is not a CloudsqlInstance custom resource"
	errProviderNotRetrieved       = "provider could not be retrieved"
	errProviderSecretNotRetrieved = "secret referred in provider could not be retrieved"
	errManagedUpdateFailed        = "cannot update CloudsqlInstance custom resource"
	errConnectionNotRetrieved     = "cannot get connection details"

	errNewClient    = "cannot create new Sqladmin Service"
	errCreateFailed = "cannot create new Cloudsql instance"
	errDeleteFailed = "cannot delete the Cloudsql instance"
	errUpdateFailed = "cannot update the Cloudsql instance"
	errGetFailed    = "cannot get the Cloudsql instance"
)

// CloudsqlInstanceController is the controller for Cloudsql CRD.
type CloudsqlInstanceController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields
// on the Controller and Start it when the Manager is Started.
func (c *CloudsqlInstanceController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1beta1.CloudsqlInstanceGroupVersionKind),
		resource.WithExternalConnecter(&cloudsqlConnector{kube: mgr.GetClient(), newServiceFn: sqladmin.NewService}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.CloudsqlInstanceKindAPIVersion, v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.CloudsqlInstance{}).
		Complete(r)
}

type cloudsqlConnector struct {
	kube         client.Client
	newServiceFn func(ctx context.Context, opts ...option.ClientOption) (*sqladmin.Service, error)
}

func (c *cloudsqlConnector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.CloudsqlInstance)
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

	s, err := c.newServiceFn(ctx,
		option.WithCredentialsJSON(secret.Data[provider.Spec.Secret.Key]),
		option.WithScopes(sqladmin.SqlserviceAdminScope))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &cloudsqlExternal{kube: c.kube, db: s.Instances, projectID: provider.Spec.ProjectID}, nil
}

type cloudsqlExternal struct {
	kube      client.Client
	db        *sqladmin.InstancesService
	projectID string
}

func (c *cloudsqlExternal) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.CloudsqlInstance)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotCloudsql)
	}
	instance, err := c.db.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetFailed)
	}
	cr.Status.AtProvider = cloudsql.GenerateObservation(*instance)
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	cloudsql.LateInitializeSpec(&cr.Spec.ForProvider, *instance)
	// TODO(muvaf): reflection in production code might cause performance bottlenecks. Generating comparison
	// methods would make more sense.
	upToDate := reflect.DeepEqual(currentSpec, &cr.Spec.ForProvider)
	if !upToDate {
		if err := c.kube.Update(ctx, cr); err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}
	var conn resource.ConnectionDetails
	switch cr.Status.AtProvider.State {
	case v1beta1.StateRunnable:
		cr.Status.SetConditions(v1alpha1.Available())
		resource.SetBindable(cr)
		conn, err = c.getConnectionDetails(ctx, cr, instance)
		if err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errConnectionNotRetrieved)
		}
	case v1beta1.StateCreating:
		cr.Status.SetConditions(v1alpha1.Creating())
	case v1beta1.StateCreationFailed, v1beta1.StateSuspended, v1beta1.StateMaintenance, v1beta1.StateUnknownState:
		cr.Status.SetConditions(v1alpha1.Unavailable())
	}
	return resource.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: conn,
	}, nil
}

func (c *cloudsqlExternal) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.CloudsqlInstance)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotCloudsql)
	}
	instance := cloudsql.GenerateDatabaseInstance(cr.Spec.ForProvider, meta.GetExternalName(cr))
	conn, err := c.getConnectionDetails(ctx, cr, instance)
	if err != nil {
		return resource.ExternalCreation{}, err
	}
	password := string(conn[v1alpha1.ResourceCredentialsSecretPasswordKey])
	if len(password) == 0 {
		password, err = util.GeneratePassword(v1beta1.PasswordLength)
		if err != nil {
			return resource.ExternalCreation{}, err
		}
		conn[v1alpha1.ResourceCredentialsSecretPasswordKey] = []byte(password)
	}
	instance.RootPassword = password
	_, err = c.db.Insert(c.projectID, instance).Context(ctx).Do()
	return resource.ExternalCreation{ConnectionDetails: conn}, errors.Wrap(resource.Ignore(gcp.IsErrorAlreadyExists, err), errCreateFailed)
}

func (c *cloudsqlExternal) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.CloudsqlInstance)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotCloudsql)
	}
	instance := cloudsql.GenerateDatabaseInstance(cr.Spec.ForProvider, meta.GetExternalName(cr))
	_, err := c.db.Patch(c.projectID, meta.GetExternalName(cr), instance).Context(ctx).Do()
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (c *cloudsqlExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.CloudsqlInstance)
	if !ok {
		return errors.New(errNotCloudsql)
	}
	_, err := c.db.Delete(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errDeleteFailed)
}

func (c *cloudsqlExternal) getConnectionDetails(ctx context.Context, cr *v1beta1.CloudsqlInstance, instance *sqladmin.DatabaseInstance) (resource.ConnectionDetails, error) {
	m := map[string][]byte{
		v1alpha1.ResourceCredentialsSecretUserKey: []byte(cloudsql.DatabaseUserName(cr.Spec.ForProvider)),
	}
	s := &v1.Secret{}
	name := types.NamespacedName{Name: cr.Spec.WriteConnectionSecretToReference.Name, Namespace: cr.Namespace}
	if err := c.kube.Get(ctx, name, s); resource.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if len(s.Data[v1alpha1.ResourceCredentialsSecretPasswordKey]) != 0 {
		m[v1alpha1.ResourceCredentialsSecretPasswordKey] = s.Data[v1alpha1.ResourceCredentialsSecretPasswordKey]
	}
	// TODO(muvaf): There might be cases where more than 1 private and/or public IP address has been assigned. We should
	// somehow show all addresses that are possible to use.
	for _, ip := range cr.Status.AtProvider.IPAddresses {
		if ip.Type == v1beta1.PrivateIPType {
			m[v1beta1.PrivateIPKey] = []byte(ip.IPAddress)
			// TODO(muvaf): we explicitly enforce use of private IP if it's available. But this should be configured
			// by resource class or claim.
			m[v1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(ip.IPAddress)
		}
		if ip.Type == v1beta1.PublicIPType {
			m[v1beta1.PublicIPKey] = []byte(ip.IPAddress)
			if len(m[v1alpha1.ResourceCredentialsSecretEndpointKey]) == 0 {
				m[v1alpha1.ResourceCredentialsSecretEndpointKey] = []byte(ip.IPAddress)
			}
		}
	}
	serverCACert := cloudsql.GetServerCACertificate(*instance)
	for k, v := range serverCACert {
		m[k] = v
	}
	return m, nil
}
