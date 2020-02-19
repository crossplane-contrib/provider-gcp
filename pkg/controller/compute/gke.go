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

package compute

import (
	"context"
	"fmt"
	"time"

	container "google.golang.org/api/container/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	gcpcomputev1alpha3 "github.com/crossplane/stack-gcp/apis/compute/v1alpha3"
	gcpv1alpha3 "github.com/crossplane/stack-gcp/apis/v1alpha3"
	gcp "github.com/crossplane/stack-gcp/pkg/clients"
	"github.com/crossplane/stack-gcp/pkg/clients/gke"
)

const (
	controllerName    = "gke.compute.gcp.crossplane.io"
	finalizer         = "finalizer." + controllerName
	clusterNamePrefix = "gke-"

	requeueOnWait   = 30 * time.Second
	requeueOnSucces = 2 * time.Minute

	updateErrorMessageFormat         = "failed to update cluster object: %s"
	erroredClusterErrorMessageFormat = "gke cluster is in %s state with message: %s"
)

// Amounts of time we wait before requeuing a reconcile.
const (
	aLongWait = 60 * time.Second
)

// Error strings
const (
	errUpdateManagedStatus = "cannot update managed resource status"
)

var (
	ctx           = context.Background()
	result        = reconcile.Result{}
	resultRequeue = reconcile.Result{Requeue: true}
)

// Reconciler reconciles a Provider object
type Reconciler struct {
	client.Client
	publisher managed.ConnectionPublisher
	resolver  managed.ReferenceResolver
	log       logging.Logger

	connect func(*gcpcomputev1alpha3.GKECluster) (gke.Client, error)
	create  func(*gcpcomputev1alpha3.GKECluster, gke.Client) (reconcile.Result, error)
	sync    func(*gcpcomputev1alpha3.GKECluster, gke.Client) (reconcile.Result, error)
	delete  func(*gcpcomputev1alpha3.GKECluster, gke.Client) (reconcile.Result, error)
}

// SetupGKECluster returns a reconciler that reconciles GKECluster
// managed resources.
func SetupGKECluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(gcpcomputev1alpha3.GKEClusterGroupKind)

	r := &Reconciler{
		Client:    mgr.GetClient(),
		publisher: managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme()),
		resolver:  managed.NewAPIReferenceResolver(mgr.GetClient()),
		log:       l.WithValues("controller", name),
	}

	r.connect = r._connect
	r.create = r._create
	r.sync = r._sync
	r.delete = r._delete

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&gcpcomputev1alpha3.GKECluster{}).
		Complete(r)
}

// fail - helper function to set fail condition with reason and message
func (r *Reconciler) fail(instance *gcpcomputev1alpha3.GKECluster, err error) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
	return resultRequeue, r.Update(context.TODO(), instance)
}

// connectionSecret return secret object for cluster instance
func connectionDetails(cluster *container.Cluster) (managed.ConnectionDetails, error) {
	config, err := gke.GenerateClientConfig(cluster)
	if err != nil {
		return nil, err
	}
	rawConfig, err := clientcmd.Write(config)
	if err != nil {
		return nil, err
	}
	cd := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey:   []byte(config.Clusters[cluster.Name].Server),
		runtimev1alpha1.ResourceCredentialsSecretUserKey:       []byte(config.AuthInfos[cluster.Name].Username),
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey:   []byte(config.AuthInfos[cluster.Name].Password),
		runtimev1alpha1.ResourceCredentialsSecretCAKey:         config.Clusters[cluster.Name].CertificateAuthorityData,
		runtimev1alpha1.ResourceCredentialsSecretClientCertKey: config.AuthInfos[cluster.Name].ClientCertificateData,
		runtimev1alpha1.ResourceCredentialsSecretClientKeyKey:  config.AuthInfos[cluster.Name].ClientKeyData,
		runtimev1alpha1.ResourceCredentialsSecretKubeconfigKey: rawConfig,
	}
	return cd, nil
}

func (r *Reconciler) _connect(instance *gcpcomputev1alpha3.GKECluster) (gke.Client, error) {
	// Fetch Provider
	p := &gcpv1alpha3.Provider{}
	if err := r.Get(ctx, meta.NamespacedNameOf(instance.Spec.ProviderReference), p); err != nil {
		return nil, err
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	secret := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := r.Client.Get(ctx, n, secret); err != nil {
		return nil, err
	}
	data, ok := secret.Data[p.Spec.CredentialsSecretRef.Key]
	if !ok {
		return nil, fmt.Errorf("secret data is not found for key [%s]", p.Spec.CredentialsSecretRef.Key)
	}
	creds, err := google.CredentialsFromJSON(context.Background(), data, gke.DefaultScope)
	if err != nil {
		return nil, err
	}
	return gke.NewClusterClient(ctx, creds)
}

func (r *Reconciler) _create(instance *gcpcomputev1alpha3.GKECluster, client gke.Client) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.Creating())
	clusterName := fmt.Sprintf("%s%s", clusterNamePrefix, instance.UID)

	meta.AddFinalizer(instance, finalizer)

	_, err := client.CreateCluster(clusterName, instance.Spec)
	if err != nil && !gcp.IsErrorAlreadyExists(err) {
		if gcp.IsErrorBadRequest(err) {
			instance.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
			// do not requeue on bad requests
			return result, r.Update(ctx, instance)
		}
		return r.fail(instance, err)
	}

	instance.Status.State = gcpcomputev1alpha3.ClusterStateProvisioning
	instance.Status.ClusterName = clusterName
	instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

	return reconcile.Result{}, errors.Wrapf(r.Update(ctx, instance), updateErrorMessageFormat, instance.GetName())
}

func (r *Reconciler) _sync(instance *gcpcomputev1alpha3.GKECluster, client gke.Client) (reconcile.Result, error) {
	cluster, err := client.GetCluster(instance.Spec.Zone, instance.Status.ClusterName)
	if err != nil {
		return r.fail(instance, err)
	}

	if cluster.Status == gcpcomputev1alpha3.ClusterStateError {
		instance.Status.State = gcpcomputev1alpha3.ClusterStateError
		instance.Status.SetConditions(runtimev1alpha1.Unavailable().
			WithMessage(fmt.Sprintf(erroredClusterErrorMessageFormat, cluster.Status, cluster.StatusMessage)))
		return resultRequeue, r.Update(context.TODO(), instance)
	}

	if cluster.Status != gcpcomputev1alpha3.ClusterStateRunning {
		return reconcile.Result{RequeueAfter: requeueOnWait}, nil
	}

	// create and publish connection details
	cd, err := connectionDetails(cluster)
	if err != nil {
		return r.fail(instance, err)
	}

	if err := r.publisher.PublishConnection(ctx, instance, cd); err != nil {
		return r.fail(instance, err)
	}

	// update resource status
	instance.Status.Endpoint = cluster.Endpoint
	instance.Status.State = gcpcomputev1alpha3.ClusterStateRunning
	instance.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	resource.SetBindable(instance)

	return reconcile.Result{RequeueAfter: requeueOnSucces},
		errors.Wrapf(r.Update(ctx, instance), updateErrorMessageFormat, instance.GetName())
}

// _delete check reclaim policy and if needed delete the gke cluster resource
func (r *Reconciler) _delete(instance *gcpcomputev1alpha3.GKECluster, client gke.Client) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.Deleting())
	if instance.Spec.ReclaimPolicy == runtimev1alpha1.ReclaimDelete {
		if err := client.DeleteCluster(instance.Spec.Zone, instance.Status.ClusterName); err != nil {
			return r.fail(instance, err)
		}
	}
	meta.RemoveFinalizer(instance, finalizer)
	instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())
	return result, errors.Wrapf(r.Update(ctx, instance), updateErrorMessageFormat, instance.GetName())
}

// Reconcile reads that state of the cluster for a Provider object and makes changes based on the state read
// and what is in the Provider.Spec
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log.Debug("Reconciling", "request", request)
	// Fetch the Provider instance
	instance := &gcpcomputev1alpha3.GKECluster{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Create GKE Client
	gkeClient, err := r.connect(instance)
	if err != nil {
		return r.fail(instance, err)
	}

	if !resource.IsConditionTrue(instance.GetCondition(runtimev1alpha1.TypeReferencesResolved)) {
		if err := r.resolver.ResolveReferences(ctx, instance); err != nil {
			condition := runtimev1alpha1.ReconcileError(err)
			if managed.IsReferencesAccessError(err) {
				condition = runtimev1alpha1.ReferenceResolutionBlocked(err)
			}

			instance.Status.SetConditions(condition)
			return reconcile.Result{RequeueAfter: aLongWait}, errors.Wrap(r.Update(ctx, instance), errUpdateManagedStatus)
		}

		// Add ReferenceResolutionSuccess to the conditions
		instance.Status.SetConditions(runtimev1alpha1.ReferenceResolutionSuccess())
	}

	// Check for deletion
	if instance.DeletionTimestamp != nil {
		return r.delete(instance, gkeClient)
	}

	// Create cluster instance
	if instance.Status.ClusterName == "" {
		return r.create(instance, gkeClient)
	}

	// Sync cluster instance status with cluster status
	return r.sync(instance, gkeClient)
}
