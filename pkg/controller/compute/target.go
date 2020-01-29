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
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplaneio/crossplane-runtime/pkg/event"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/target"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane/apis/workload/v1alpha1"

	"github.com/crossplaneio/stack-gcp/apis/compute/v1alpha3"
)

// SetupGKEClusterTarget adds a controller that propagates GKECluster
// connection secrets to the connection secrets of their targets.
func SetupGKEClusterTarget(mgr ctrl.Manager, l logging.Logger) error {
	name := target.ControllerName(v1alpha3.GKEClusterKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.KubernetesTarget{}).
		WithEventFilter(resource.NewPredicates(resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha3.GKEClusterGroupVersionKind)))).
		Complete(target.NewReconciler(mgr,
			resource.TargetKind(v1alpha1.KubernetesTargetGroupVersionKind),
			resource.ManagedKind(v1alpha3.GKEClusterGroupVersionKind),
			target.WithLogger(l.WithValues("controller", name)),
			target.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}
