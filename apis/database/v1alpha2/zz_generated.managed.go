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

// Code generated by angryjet. DO NOT EDIT.

package v1alpha2

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// GetBindingPhase of this CloudsqlInstance.
func (mg *CloudsqlInstance) GetBindingPhase() runtimev1alpha1.BindingPhase {
	return mg.Status.GetBindingPhase()
}

// GetClaimReference of this CloudsqlInstance.
func (mg *CloudsqlInstance) GetClaimReference() *corev1.ObjectReference {
	return mg.Spec.ClaimReference
}

// GetNonPortableClassReference of this CloudsqlInstance.
func (mg *CloudsqlInstance) GetNonPortableClassReference() *corev1.ObjectReference {
	return mg.Spec.NonPortableClassReference
}

// GetReclaimPolicy of this CloudsqlInstance.
func (mg *CloudsqlInstance) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return mg.Spec.ReclaimPolicy
}

// GetWriteConnectionSecretToReference of this CloudsqlInstance.
func (mg *CloudsqlInstance) GetWriteConnectionSecretToReference() corev1.LocalObjectReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetBindingPhase of this CloudsqlInstance.
func (mg *CloudsqlInstance) SetBindingPhase(p runtimev1alpha1.BindingPhase) {
	mg.Status.SetBindingPhase(p)
}

// SetClaimReference of this CloudsqlInstance.
func (mg *CloudsqlInstance) SetClaimReference(r *corev1.ObjectReference) {
	mg.Spec.ClaimReference = r
}

// SetConditions of this CloudsqlInstance.
func (mg *CloudsqlInstance) SetConditions(c ...runtimev1alpha1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetNonPortableClassReference of this CloudsqlInstance.
func (mg *CloudsqlInstance) SetNonPortableClassReference(r *corev1.ObjectReference) {
	mg.Spec.NonPortableClassReference = r
}

// SetReclaimPolicy of this CloudsqlInstance.
func (mg *CloudsqlInstance) SetReclaimPolicy(r runtimev1alpha1.ReclaimPolicy) {
	mg.Spec.ReclaimPolicy = r
}

// SetWriteConnectionSecretToReference of this CloudsqlInstance.
func (mg *CloudsqlInstance) SetWriteConnectionSecretToReference(r corev1.LocalObjectReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}
