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
	"strings"

	"github.com/google/go-cmp/cmp"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/database/v1beta1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
	"github.com/crossplane/provider-gcp/pkg/clients/cloudsql"
	corev1 "k8s.io/api/core/v1"
)

const (
	errNotCloudSQL         = "managed resource is not a CloudSQLInstance custom resource"
	errManagedUpdateFailed = "cannot update CloudSQLInstance custom resource"

	errNewClient        = "cannot create new Sqladmin Service"
	errCreateFailed     = "cannot create new CloudSQL instance"
	errNameInUse        = "cannot create new CloudSQL instance, resource name is unavailable because it is in use or was used recently"
	errDeleteFailed     = "cannot delete the CloudSQL instance"
	errUpdateFailed     = "cannot update the CloudSQL instance"
	errGetFailed        = "cannot get the CloudSQL instance"
	errGeneratePassword = "cannot generate root password"
	errCheckUpToDate    = "cannot determine if CloudSQL instance is up to date"
	errGetSecretFailed  = "failed to get Kubernetes secret for spec.replicaConfiguration.mysqlReplicaConfiguration.secretRef"
)

// SetupCloudSQLInstance adds a controller that reconciles
// CloudSQLInstance managed resources.
func SetupCloudSQLInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.CloudSQLInstanceGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.CloudSQLInstanceGroupVersionKind),
		managed.WithExternalConnecter(&cloudsqlConnector{kube: mgr.GetClient()}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &cloudsqlTagger{kube: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.CloudSQLInstance{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type cloudsqlConnector struct {
	kube client.Client
}

func (c *cloudsqlConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	projectID, opts, err := gcp.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	s, err := sqladmin.NewService(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}
	return &cloudsqlExternal{kube: c.kube, db: s.Instances, projectID: projectID}, nil
}

type cloudsqlExternal struct {
	kube      client.Client
	db        *sqladmin.InstancesService
	projectID string
}

func (c *cloudsqlExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.CloudSQLInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCloudSQL)
	}
	instance, err := c.db.Get(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(gcp.IsErrorNotFound, err), errGetFailed)
	}

	// in case secretRef is defined, fetch the Secret
	var sc *corev1.Secret
	if cr.Spec.ForProvider.ReplicaConfiguration != nil &&
		cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration != nil &&
		cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration.SecretRef != nil {

		sc, err = c.fetchSecret(ctx, cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration.SecretRef)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errGetSecretFailed)
		}
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	cloudsql.LateInitializeSpec(cloudsql.CloudSQLOptions{Instance: instance, Spec: &cr.Spec.ForProvider, Secret: sc})
	// TODO(muvaf): reflection in production code might cause performance bottlenecks. Generating comparison
	// methods would make more sense.
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}
	cr.Status.AtProvider = cloudsql.GenerateObservation(*instance)
	switch cr.Status.AtProvider.State {
	case v1beta1.StateRunnable:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta1.StateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta1.StateCreationFailed, v1beta1.StateSuspended, v1beta1.StateMaintenance, v1beta1.StateUnknownState:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	upToDate, err := cloudsql.IsUpToDate(cloudsql.CloudSQLOptions{Name: meta.GetExternalName(cr), Spec: &cr.Spec.ForProvider, Instance: instance, Secret: sc})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: getConnectionDetails(cr, instance),
	}, nil
}

func (c *cloudsqlExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.CloudSQLInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCloudSQL)
	}
	cr.SetConditions(xpv1.Creating())
	instance := &sqladmin.DatabaseInstance{}

    // in case secretRef is defined, fetch the Secret
	var sc *corev1.Secret
	if cr.Spec.ForProvider.ReplicaConfiguration != nil &&
		cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration != nil &&
		cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration.SecretRef != nil {

		var err error
		sc, err = c.fetchSecret(ctx, cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration.SecretRef)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errGetSecretFailed)
		}
	}

	err := cloudsql.GenerateDatabaseInstance(cloudsql.CloudSQLOptions{Name: meta.GetExternalName(cr), Instance: instance, Spec: &cr.Spec.ForProvider, Secret: sc})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	pw, err := password.Generate()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGeneratePassword)
	}

	instance.RootPassword = pw
	if _, err := c.db.Insert(c.projectID, instance).Context(ctx).Do(); err != nil {
		// We don't want to return (and thus publish) our randomly generated
		// password if we didn't actually successfully create a new instance.
		if gcp.IsErrorAlreadyExists(err) {
			return managed.ExternalCreation{}, errors.Wrap(err, errNameInUse)
		}
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	cd := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretPasswordKey: []byte(pw),
	}
	return managed.ExternalCreation{ConnectionDetails: cd}, nil
}

func (c *cloudsqlExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.CloudSQLInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCloudSQL)
	}
	if cr.Status.AtProvider.State == v1beta1.StateCreating {
		return managed.ExternalUpdate{}, nil
	}

    // in case secretRef is defined, fetch the Secret
	var sc *corev1.Secret
	if cr.Spec.ForProvider.ReplicaConfiguration != nil &&
		cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration != nil &&
		cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration.SecretRef != nil {

		var err error
		sc, err = c.fetchSecret(ctx, cr.Spec.ForProvider.ReplicaConfiguration.MysqlReplicaConfiguration.SecretRef)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errGetSecretFailed)
		}
	}
	instance := &sqladmin.DatabaseInstance{}
	err := cloudsql.GenerateDatabaseInstance(cloudsql.CloudSQLOptions{Name: meta.GetExternalName(cr), Spec: &cr.Spec.ForProvider, Instance: instance, Secret: sc})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}
	// TODO(muvaf): the returned operation handle could help us not to send Patch
	// request aggressively.
	_, err = c.db.Patch(c.projectID, meta.GetExternalName(cr), instance).Context(ctx).Do()
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (c *cloudsqlExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.CloudSQLInstance)
	if !ok {
		return errors.New(errNotCloudSQL)
	}
	cr.SetConditions(xpv1.Deleting())
	_, err := c.db.Delete(c.projectID, meta.GetExternalName(cr)).Context(ctx).Do()
	if gcp.IsErrorNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errDeleteFailed)
}

func getConnectionDetails(cr *v1beta1.CloudSQLInstance, instance *sqladmin.DatabaseInstance) managed.ConnectionDetails {
	m := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretUserKey: []byte(cloudsql.DatabaseUserName(cr.Spec.ForProvider)),
		v1beta1.CloudSQLSecretConnectionName:  []byte(instance.ConnectionName),
	}

	// TODO(muvaf): There might be cases where more than 1 private and/or public IP address has been assigned. We should
	// somehow show all addresses that are possible to use.
	for _, ip := range cr.Status.AtProvider.IPAddresses {
		if ip.Type == v1beta1.PrivateIPType {
			m[v1beta1.PrivateIPKey] = []byte(ip.IPAddress)
			// TODO(muvaf): we explicitly enforce use of private IP if it's available. But this should be configured
			// by resource class or claim.
			m[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(ip.IPAddress)
		}
		if ip.Type == v1beta1.PublicIPType {
			m[v1beta1.PublicIPKey] = []byte(ip.IPAddress)
			if len(m[xpv1.ResourceCredentialsSecretEndpointKey]) == 0 {
				m[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(ip.IPAddress)
			}
		}
	}
	serverCACert := cloudsql.GetServerCACertificate(*instance)
	for k, v := range serverCACert {
		m[k] = v
	}

	return m
}

type cloudsqlTagger struct {
	kube client.Client
}

// Initialize adds the external tags to spec.forProvider.settings.userLabels
func (t *cloudsqlTagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.CloudSQLInstance)
	if !ok {
		return errors.New(errNotCloudSQL)
	}
	if cr.Spec.ForProvider.Settings.UserLabels == nil {
		cr.Spec.ForProvider.Settings.UserLabels = map[string]string{}
	}
	for k, v := range resource.GetExternalTags(cr) {
		// NOTE(muvaf): See label constraints here https://cloud.google.com/compute/docs/labeling-resources
		cr.Spec.ForProvider.Settings.UserLabels[k] = strings.ToLower(strings.ReplaceAll(v, ".", "_"))
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errManagedUpdateFailed)
}

// fetchSecret get Secret from SecretReference
func (c *cloudsqlExternal) fetchSecret(ctx context.Context, ref *xpv1.SecretReference) (*corev1.Secret, error) {
	nn := types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
	sc := &corev1.Secret{}
	err := c.kube.Get(ctx, nn, sc)

	return sc, err
}
