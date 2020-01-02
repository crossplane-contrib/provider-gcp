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

package v1beta1

import (
	"context"
	"strings"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// URIPrefix is the prefix that is used in selfLink
	URIPrefix = "https://container.googleapis.com/v1beta1/"
)

// GKEClusterURIReferencer retrieves a GKEClusterURI from a referenced GKECluster object
type GKEClusterURIReferencer struct {
	corev1.LocalObjectReference `json:",inline"`
}

// GetStatus implements GetStatus method of AttributeReferencer interface
func (v *GKEClusterURIReferencer) GetStatus(ctx context.Context, _ resource.CanReference, reader client.Reader) ([]resource.ReferenceStatus, error) {
	cluster := GKECluster{}
	nn := types.NamespacedName{Name: v.Name}
	if err := reader.Get(ctx, nn, &cluster); err != nil {
		if kerrors.IsNotFound(err) {
			return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceNotFound}}, nil
		}

		return nil, err
	}

	if !resource.IsConditionTrue(cluster.GetCondition(runtimev1alpha1.TypeReady)) {
		return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceNotReady}}, nil
	}

	return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceReady}}, nil
}

// Build retrieves and builds the GKEClusterURI
func (v *GKEClusterURIReferencer) Build(ctx context.Context, _ resource.CanReference, reader client.Reader) (string, error) {
	cluster := GKECluster{}
	nn := types.NamespacedName{Name: v.Name}
	if err := reader.Get(ctx, nn, &cluster); err != nil {
		return "", err
	}

	return strings.TrimPrefix(cluster.Status.AtProvider.SelfLink, URIPrefix), nil
}
