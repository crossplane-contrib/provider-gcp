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

package providerconfig

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-gcp/apis/v1beta1"
)

const (
	finalizer      = "in-use." + v1beta1.Group
	shortWait      = 30 * time.Second
	timeout        = 2 * time.Minute
	maxConcurrency = 5

	errGetPC        = "cannot get ProviderConfig"
	errListPCUs     = "cannot list ProviderConfigUsages"
	errDeletePCU    = "cannot delete ProviderConfigUsage"
	errUpdate       = "cannot update ProviderConfig"
	errUpdateStatus = "cannot update ProviderConfig status"
)

// Event reasons.
const (
	reasonAccount event.Reason = "UsageAccounting"
)

// Setup adds a controller that reconciles a ProviderConfig by accounting for
// the managed resources that are using it, and ensuring it cannot be deleted
// until it is no longer in use.
func Setup(mgr ctrl.Manager, log logging.Logger) error {
	name := "provider/" + strings.ToLower(v1beta1.ProviderConfigGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.ProviderConfig{}).
		Watches(&source.Kind{Type: &v1beta1.ProviderConfigUsage{}}, &EnqueueRequestForProviderConfig{}).
		WithOptions(kcontroller.Options{MaxConcurrentReconciles: maxConcurrency}).
		Complete(NewReconciler(mgr,
			WithLogger(log.WithValues("controller", name)),
			WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.log = log
	}
}

// WithRecorder specifies how the Reconciler should record Kubernetes events.
func WithRecorder(er event.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = er
	}
}

// NewReconciler returns a Reconciler of ProviderConfigs.
func NewReconciler(mgr manager.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: mgr.GetClient(),
		log:    logging.NewNopLogger(),
		record: event.NewNopRecorder(),
	}

	for _, f := range opts {
		f(r)
	}
	return r
}

// A Reconciler reconciles ProviderConfigs.
type Reconciler struct {
	client client.Client

	log    logging.Logger
	record event.Recorder
}

// Reconcile a ProviderConfig by accounting for the managed resources that are
// using it, and ensuring it cannot be deleted until it is no longer in use.
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pc := &v1beta1.ProviderConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, pc); err != nil {
		// In case object is not found, most likely the object was deleted and
		// then disappeared while the event was in the processing queue. We
		// don't need to take any action in that case.
		log.Debug(errGetPC, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetPC)
	}

	log = log.WithValues(
		"uid", pc.GetUID(),
		"version", pc.GetResourceVersion(),
		"name", pc.GetName(),
	)

	l := &v1beta1.ProviderConfigUsageList{}
	if err := r.client.List(ctx, l, client.MatchingLabels{v1beta1.LabelKeyProviderName: pc.GetName()}); err != nil {
		log.Debug(errListPCUs, "error", err)
		r.record.Event(pc, event.Warning(reasonAccount, errors.Wrap(err, errListPCUs)))
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	usages := int64(len(l.Items))
	for _, pcu := range l.Items {
		pcu := pcu // Pin range variable so we can take its address.

		if metav1.GetControllerOf(&pcu) == nil {
			// Usages should always have a controller reference. If this one has
			// none it's probably been stripped off (e.g. by a Velero restore).
			// We can safely delete it - it's either stale, or will be recreated
			// next time the relevant managed resource connects.
			if err := r.client.Delete(ctx, &pcu); resource.IgnoreNotFound(err) != nil {
				log.Debug(errDeletePCU, "error", err)
				r.record.Event(pc, event.Warning(reasonAccount, errors.Wrap(err, errDeletePCU)))
				continue
			}
			usages--
		}
	}

	log = log.WithValues("usages", usages)

	if meta.WasDeleted(pc) {
		if usages > 0 {
			log.Debug("Blocking deletion while usages still exist")
			r.record.Event(pc, event.Warning(reasonAccount, errors.New("Blocking deletion while usages still exist")))

			// We're watching our usages, so we'll be requeued when they go.
			return reconcile.Result{Requeue: false}, nil
		}

		meta.RemoveFinalizer(pc, finalizer)
		if err := r.client.Update(ctx, pc); err != nil {
			r.log.Debug(errUpdate, "error", err)
			return reconcile.Result{RequeueAfter: shortWait}, nil
		}

		// We've been deleted - there's no more work to do.
		return reconcile.Result{Requeue: false}, nil
	}

	meta.AddFinalizer(pc, finalizer)
	if err := r.client.Update(ctx, pc); err != nil {
		r.log.Debug(errUpdate, "error", err)
		return reconcile.Result{RequeueAfter: shortWait}, nil
	}

	// There's no need to requeue explicitly - we're watching all PCs.
	pc.Status.Usages = &usages
	return reconcile.Result{Requeue: false}, errors.Wrap(r.client.Status().Update(ctx, pc), errUpdateStatus)
}
