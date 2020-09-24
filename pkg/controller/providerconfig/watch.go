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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/provider-gcp/apis/v1beta1"
)

type adder interface {
	Add(item interface{})
}

// EnqueueRequestForProviderConfig enqueues a request for the ProviderConfig
// that a ProviderConfigUsage represents.
type EnqueueRequestForProviderConfig struct{}

// Create adds a NamespacedName for the supplied CreateEvent if its Object is a
// ProviderConfigUsage.
func (e *EnqueueRequestForProviderConfig) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	addPC(evt.Object, q)
}

// Update adds a NamespacedName for the supplied UpdateEvent if of of its
// Objects is a ProviderConfigUsage.
func (e *EnqueueRequestForProviderConfig) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	addPC(evt.ObjectOld, q)
	addPC(evt.ObjectNew, q)
}

// Delete adds a NamespacedName for the supplied DeleteEvent if its Object is a
// ProviderConfigUsage.
func (e *EnqueueRequestForProviderConfig) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	addPC(evt.Object, q)
}

// Generic adds a NamespacedName for the supplied GenericEvent if its Object is a
// ProviderConfigUsage.
func (e *EnqueueRequestForProviderConfig) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	addPC(evt.Object, q)
}

func addPC(obj runtime.Object, queue adder) {
	pcu, ok := obj.(*v1beta1.ProviderConfigUsage)
	if !ok {
		return
	}
	queue.Add(reconcile.Request{NamespacedName: types.NamespacedName{Name: pcu.ProviderRef.Name}})
}
