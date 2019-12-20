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

package storage

import (
	"context"
	"reflect"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	gcpv1alpha3 "github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	gcpstorage "github.com/crossplaneio/stack-gcp/pkg/clients/storage"
)

const (
	controllerName = "bucket.storage.gcp.crossplane.io"
	finalizer      = "finalizer." + controllerName

	reconcileTimeout      = 1 * time.Minute
	requeueAfterOnSuccess = 30 * time.Second
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
	resultRequeue    = reconcile.Result{Requeue: true}
	requeueOnSuccess = reconcile.Result{RequeueAfter: requeueAfterOnSuccess}

	log = logging.Logger.WithName("controller." + controllerName)
)

// Reconciler reconciles a GCP storage bucket bucket
type Reconciler struct {
	client.Client
	factory
	resource.ManagedReferenceResolver
}

// BucketController is responsible for adding the Bucket controller and its
// corresponding reconciler to the manager with any runtime configuration.
type BucketController struct{}

// SetupWithManager creates a newSyncDeleter Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func (c *BucketController) SetupWithManager(mgr ctrl.Manager) error {
	r := &Reconciler{
		Client:                   mgr.GetClient(),
		factory:                  &bucketFactory{mgr.GetClient()},
		ManagedReferenceResolver: resource.NewAPIManagedReferenceResolver(mgr.GetClient()),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&v1alpha3.Bucket{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// Reconcile reads that state of the cluster for a Provider bucket and makes changes based on the state read
// and what is in the Provider.Spec
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.V(logging.Debug).Info("reconciling", "kind", v1alpha3.BucketKindAPIVersion, "request", request)

	ctx, cancel := context.WithTimeout(context.Background(), reconcileTimeout)
	defer cancel()

	b := &v1alpha3.Bucket{}
	if err := r.Get(ctx, request.NamespacedName, b); err != nil {
		if kerrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if !resource.IsConditionTrue(b.GetCondition(runtimev1alpha1.TypeReferencesResolved)) {
		if err := r.ResolveReferences(ctx, b); err != nil {
			condition := runtimev1alpha1.ReconcileError(err)
			if resource.IsReferencesAccessError(err) {
				condition = runtimev1alpha1.ReferenceResolutionBlocked(err)
			}

			b.Status.SetConditions(condition)
			return reconcile.Result{RequeueAfter: aLongWait}, errors.Wrap(r.Update(ctx, b), errUpdateManagedStatus)
		}

		// Add ReferenceResolutionSuccess to the conditions
		b.Status.SetConditions(runtimev1alpha1.ReferenceResolutionSuccess())
	}

	bh, err := r.newSyncDeleter(ctx, b)
	if err != nil {
		b.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
		return resultRequeue, r.Status().Update(ctx, b)
	}

	// Check for deletion
	if b.DeletionTimestamp != nil {
		return bh.delete(ctx)
	}

	return bh.sync(ctx)
}

type factory interface {
	newSyncDeleter(context.Context, *v1alpha3.Bucket) (syncdeleter, error)
}

type bucketFactory struct {
	client.Client
}

func (m *bucketFactory) newSyncDeleter(ctx context.Context, b *v1alpha3.Bucket) (syncdeleter, error) {
	p := &gcpv1alpha3.Provider{}
	if err := m.Get(ctx, meta.NamespacedNameOf(b.Spec.ProviderReference), p); err != nil {
		return nil, err
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := m.Get(ctx, n, s); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider's secret %s", n)
	}

	creds, err := google.CredentialsFromJSON(context.Background(), s.Data[p.Spec.CredentialsSecretRef.Key], storage.ScopeFullControl)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot retrieve creds from json")
	}

	sc, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, errors.Wrapf(err, "error creating storage client")
	}

	ops := &bucketHandler{
		Bucket: b,
		gcp:    &gcpstorage.BucketClient{BucketHandle: sc.Bucket(b.GetBucketName())},
		kube:   m.Client,
	}

	return &bucketSyncDeleter{
		operations:    ops,
		createupdater: &bucketCreateUpdater{operations: ops, projectID: creds.ProjectID},
	}, nil

}

type syncdeleter interface {
	delete(context.Context) (reconcile.Result, error)
	sync(context.Context) (reconcile.Result, error)
}

type bucketSyncDeleter struct {
	operations
	createupdater
}

func newBucketSyncDeleter(ops operations, projectID string) *bucketSyncDeleter {
	return &bucketSyncDeleter{
		operations:    ops,
		createupdater: newBucketCreateUpdater(ops, projectID),
	}
}

func (bh *bucketSyncDeleter) delete(ctx context.Context) (reconcile.Result, error) {
	bh.setStatusConditions(runtimev1alpha1.Deleting())

	if bh.isReclaimDelete() {
		if err := bh.deleteBucket(ctx); err != nil && err != storage.ErrBucketNotExist {
			bh.setStatusConditions(runtimev1alpha1.ReconcileError(err))
			return resultRequeue, bh.updateStatus(ctx)
		}
	}

	// NOTE(negz): We don't update the conditioned status here because assuming
	// no other finalizers need to be cleaned up the object should cease to
	// exist after we update it.
	bh.removeFinalizer()
	return reconcile.Result{}, bh.updateObject(ctx)
}

// sync - synchronizes the state of the bucket resource with the state of the
// bucket Kubernetes bucket
func (bh *bucketSyncDeleter) sync(ctx context.Context) (reconcile.Result, error) {
	if err := bh.updateSecret(ctx); err != nil {
		bh.setStatusConditions(runtimev1alpha1.ReconcileError(err))
		return resultRequeue, bh.updateStatus(ctx)
	}

	attrs, err := bh.getAttributes(ctx)
	if err != nil && err != storage.ErrBucketNotExist {
		return resultRequeue, bh.updateStatus(ctx)
	}

	if attrs == nil {
		return bh.create(ctx)
	}

	return bh.update(ctx, attrs)
}

// createupdater interface defining create and update operations on/for bucket resource
type createupdater interface {
	create(context.Context) (reconcile.Result, error)
	update(context.Context, *storage.BucketAttrs) (reconcile.Result, error)
}

// bucketCreateUpdater implementation of createupdater interface
type bucketCreateUpdater struct {
	operations
	projectID string
}

// newBucketCreateUpdater new instance of bucketCreateUpdater
func newBucketCreateUpdater(ops operations, pID string) *bucketCreateUpdater {
	return &bucketCreateUpdater{
		operations: ops,
		projectID:  pID,
	}
}

// create new bucket resource and save changes back to bucket specs
func (bh *bucketCreateUpdater) create(ctx context.Context) (reconcile.Result, error) {
	bh.setStatusConditions(runtimev1alpha1.Creating())
	bh.addFinalizer()

	if err := bh.createBucket(ctx, bh.projectID); err != nil {
		bh.setStatusConditions(runtimev1alpha1.ReconcileError(err))
		return resultRequeue, bh.updateStatus(ctx)
	}

	attrs, err := bh.getAttributes(ctx)
	if err != nil {
		bh.setStatusConditions(runtimev1alpha1.ReconcileError(err))
		return resultRequeue, bh.updateStatus(ctx)
	}
	bh.setSpecAttrs(attrs)

	if err := bh.updateObject(ctx); err != nil {
		return resultRequeue, err
	}
	bh.setStatusAttrs(attrs)

	bh.setStatusConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	bh.setBindable()

	return requeueOnSuccess, bh.updateStatus(ctx)
}

// update bucket resource if needed
func (bh *bucketCreateUpdater) update(ctx context.Context, attrs *storage.BucketAttrs) (reconcile.Result, error) {
	current := v1alpha3.NewBucketUpdatableAttrs(attrs)
	if reflect.DeepEqual(*current, bh.getSpecAttrs()) {
		return requeueOnSuccess, nil
	}

	attrs, err := bh.updateBucket(ctx, attrs.Labels)
	if err != nil {
		bh.setStatusConditions(runtimev1alpha1.ReconcileError(err))
		return resultRequeue, bh.updateStatus(ctx)
	}

	// Sync attributes back to spec
	bh.setSpecAttrs(attrs)
	if err := bh.updateObject(ctx); err != nil {
		return resultRequeue, err
	}

	bh.setStatusConditions(runtimev1alpha1.ReconcileSuccess())
	return requeueOnSuccess, bh.updateStatus(ctx)
}
