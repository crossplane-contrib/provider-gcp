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

package gcp

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	cmpv1beta1 "github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	"github.com/crossplane/provider-gcp/apis/v1alpha3"
	"github.com/crossplane/provider-gcp/apis/v1beta1"
)

const scopeCloudPlatform = "https://www.googleapis.com/auth/cloud-platform"

// GetConnectionInfo returns the necessary connection information that is necessary
// to use when the controller connects to GCP API in order to reconcile the managed
// resource.
func GetConnectionInfo(ctx context.Context, c client.Client, mg resource.Managed) (projectID string, opts []option.ClientOption, err error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	case mg.GetProviderReference() != nil:
		return UseProvider(ctx, c, mg)
	default:
		return "", nil, errors.New("neither providerConfigRef nor providerRef is given")
	}
}

// UseProvider to return GCP authentication information.
// Deprecated: Use UseProviderConfig
func UseProvider(ctx context.Context, c client.Client, mg resource.Managed) (projectID string, opts []option.ClientOption, err error) {
	opts = make([]option.ClientOption, 0)

	p := &v1alpha3.Provider{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderReference().Name}, p); err != nil {
		return "", nil, err
	}

	if len(p.Spec.Endpoint) > 0 {
		opts = append(opts, option.WithEndpoint(p.Spec.Endpoint))
	}

	if p.Spec.WithoutAuthentication {
		opts = append(opts, option.WithoutAuthentication())
	}

	ref := p.Spec.CredentialsSecretRef
	s := &v1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, s); err != nil {
		return "", nil, err
	}
	opts = append(opts, option.WithCredentialsJSON(s.Data[ref.Key]))
	return p.Spec.ProjectID, opts, nil
}

// UseProviderConfig to return GCP authentication information.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (projectID string, opts []option.ClientOption, err error) {
	opts = make([]option.ClientOption, 0)

	pc := &v1beta1.ProviderConfig{}
	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return "", nil, err
	}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return "", nil, err
	}

	if len(pc.Spec.Endpoint) > 0 {
		opts = append(opts, option.WithEndpoint(pc.Spec.Endpoint))
	}

	if pc.Spec.WithoutAuthentication {
		opts = append(opts, option.WithoutAuthentication())
	}

	switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
	case xpv1.CredentialsSourceInjectedIdentity:
		ts, err := google.DefaultTokenSource(ctx, scopeCloudPlatform)
		if err != nil {
			return "", nil, errors.Wrap(err, "cannot get application default credentials token")
		}
		opts = append(opts, option.WithTokenSource(ts))
	default:
		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, c, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return "", nil, errors.Wrap(err, "cannot get credentials")
		}
		opts = append(opts, option.WithCredentialsJSON(data))
	}

	return pc.Spec.ProjectID, opts, nil
}

// IsErrorNotFoundGRPC gets a value indicating whether the given error represents
// a "not found" response from the Google API. It works only for the clients
// that use gRPC as protocol.
func IsErrorNotFoundGRPC(err error) bool {
	if err == nil {
		return false
	}
	grpcErr, ok := err.(interface{ GRPCStatus() *status.Status })
	return ok && grpcErr.GRPCStatus().Code() == codes.NotFound
}

// IsErrorNotFound gets a value indicating whether the given error represents a "not found" response from the Google API
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	googleapiErr, ok := err.(*googleapi.Error)
	return ok && googleapiErr.Code == http.StatusNotFound
}

// IsErrorAlreadyExists gets a value indicating whether the given error
// represents a "conflict" response from the Google API
func IsErrorAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	googleapiErr, ok := err.(*googleapi.Error)
	return ok && googleapiErr.Code == http.StatusConflict
}

// IsErrorBadRequest gets a value indicating whether the given error represents
// a "bad request" response from the Google API
func IsErrorBadRequest(err error) bool {
	if err == nil {
		return false
	}
	googleapiErr, ok := err.(*googleapi.Error)
	return ok && googleapiErr.Code == http.StatusBadRequest
}

// IsErrorForbidden gets a value indicating whether the given error represents a
// "forbidden" response from the Google API
func IsErrorForbidden(err error) bool {
	if err == nil {
		return false
	}
	googleapiErr, ok := err.(*googleapi.Error)
	return ok && googleapiErr.Code == http.StatusForbidden
}

// StringValue converts the supplied string pointer to a string, returning the
// empty string if the pointer is nil.
func StringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// Int64Value converts the supplied int64 pointer to an int, returning zero if
// the pointer is nil.
func Int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// BoolValue converts the supplied bool pointer to an bool, returning false if
// the pointer is nil.
func BoolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// StringPtr converts the supplied string to a pointer to that string.
func StringPtr(p string) *string { return &p }

// Int64Ptr converts the supplied int64 to a pointer to that int64.
func Int64Ptr(p int64) *int64 { return &p }

// BoolPtr converts the supplied bool to a pointer to that bool
func BoolPtr(p bool) *bool { return &p }

// LateInitialize functions initialize s(first argument), presumed to be an
// optional field of a Kubernetes API object's spec per Kubernetes
// "late initialization" semantics. s is returned unchanged if it is non-nil
// or from(second argument) is the empty string, otherwise a pointer to from
// is returned.
// https://github.com/kubernetes/community/blob/db7f270f/contributors/devel/sig-architecture/api-conventions.md#optional-vs-required
// https://github.com/kubernetes/community/blob/db7f270f/contributors/devel/sig-architecture/api-conventions.md#late-initialization
// TODO(muvaf): These functions will probably be needed by other providers.
// Consider moving them to crossplane-runtime.

// LateInitializeString implements late initialization for string type.
func LateInitializeString(s *string, from string) *string {
	if s != nil || from == "" {
		return s
	}
	return &from
}

// LateInitializeInt64 implements late initialization for int64 type.
func LateInitializeInt64(i *int64, from int64) *int64 {
	if i != nil || from == 0 {
		return i
	}
	return &from
}

// LateInitializeBool implements late initialization for bool type.
func LateInitializeBool(b *bool, from bool) *bool {
	if b != nil || !from {
		return b
	}
	return &from
}

// LateInitializeStringSlice implements late initialization for
// string slice type.
func LateInitializeStringSlice(s []string, from []string) []string {
	if len(s) != 0 || len(from) == 0 {
		return s
	}
	return from
}

// LateInitializeStringMap implements late initialization for
// string map type.
func LateInitializeStringMap(s map[string]string, from map[string]string) map[string]string {
	if len(s) != 0 || len(from) == 0 {
		return s
	}
	return from
}

// EquateComputeURLs considers compute APIs to be equal whether they are fully
// qualified, partially qualified, or unqualified. The compute API will accept
// unqualified or partially qualified URLs for certain fields, but return fully
// qualified URLs. For example it may accept 'us-central1' but return
// 'https://www.googleapis.com/compute/v1/projects/example/regions/us-central1'.
// 'projects/example/global/networks/eg' is also valid, but the API may return
// 'https://www.googleapis.com/compute/v1/projects/example/global/networks/eg'.
func EquateComputeURLs() cmp.Option {
	return cmp.Comparer(func(a, b string) bool {
		if a == b {
			return true
		}

		if !strings.HasPrefix(a, cmpv1beta1.ComputeURIPrefix) && !strings.HasPrefix(b, cmpv1beta1.ComputeURIPrefix) {
			return a == b
		}

		ta := strings.TrimPrefix(a, cmpv1beta1.ComputeURIPrefix)
		tb := strings.TrimPrefix(b, cmpv1beta1.ComputeURIPrefix)

		// Partially qualified URLs are considered equal to their corresponding
		// fully qualified URLs.
		if ta == tb {
			return true
		}

		// Completely unqualified names should be considered equal to their
		// partial or fully qualified equivalents.
		return path.Base(ta) == path.Base(tb)
	})
}
