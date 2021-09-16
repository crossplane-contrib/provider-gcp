/*
Copyright 2021 The Crossplane Authors.

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

package iam

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
)

const (
	nameKeyTestProject            = "key-test-project"
	nameTestServiceAccountKey     = "test-service-account-key"
	nameExternalServiceAccount    = "test-service-account"
	nameExternalServiceAccountKey = "eba77d0929159603f16f3b6dccef14f2924548f1"
	rrnTestServiceAccount         = "projects/" + nameKeyTestProject +
		"/serviceAccounts/" + nameExternalServiceAccount + "@" + nameKeyTestProject + ".iam.gserviceaccount.com"
	rrnTestServiceAccountKey    = rrnTestServiceAccount + "/keys/" + nameExternalServiceAccountKey
	rrnInvalidServiceAccountKey = ":invalid-rrn:"
	// Google Cloud API iam.ServiceAccountKey response consts
	valIAMPrivateKeyType  = "iam.PrivateKeyType"
	valIAMKeyAlgorithm    = "iam.KeyAlgorithm"
	valIAMValidAfterTime  = "iam.ValidAfterTime"
	valIAMValidBeforeTime = "iam.ValidBeforeTime"
	valIAMKeyOrigin       = "iam.KeyOrigin"
	valIAMKeyType         = "iam.KeyType"
	valIAMPrivateKeyData  = "iam.PrivateKeyData"
	valIAMPublicKeyData   = "iam.PublicKeyData"
	valIAMPublicKeyType   = "iam.PublicKeyType"
)

var (
	iamSaKeyGetObject = iamv1.ServiceAccountKey{
		KeyAlgorithm:    valIAMKeyAlgorithm,
		KeyOrigin:       valIAMKeyOrigin,
		KeyType:         valIAMKeyType,
		Name:            rrnTestServiceAccountKey,
		PublicKeyData:   valIAMPublicKeyData,
		ValidAfterTime:  valIAMValidAfterTime,
		ValidBeforeTime: valIAMValidBeforeTime,
	}

	iamSaKeyCreateObject = iamv1.ServiceAccountKey{
		KeyAlgorithm:    valIAMKeyAlgorithm,
		KeyOrigin:       valIAMKeyOrigin,
		KeyType:         valIAMKeyType,
		Name:            rrnTestServiceAccountKey,
		PublicKeyData:   valIAMPublicKeyData,
		ValidAfterTime:  valIAMValidAfterTime,
		ValidBeforeTime: valIAMValidBeforeTime,
		// private key is available only in iam.ServiceAccountKey.create responses
		PrivateKeyType: valIAMPrivateKeyType,
		PrivateKeyData: valIAMPrivateKeyData,
	}
)

func TestServiceAccountKeyObserve(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		mg  resource.Managed
		err error
	}

	testCases := map[string]struct {
		reason  string
		handler http.Handler
		args    args
		want    want
	}{
		"NotServiceAccountKey": {
			reason: "assert error if not reconciling on a valid v1alpha1.ServiceAccountKey object",
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccountKey),
			},
		},
		"ExternalNameNotSet": {
			reason: "SA rrn (resource path) is set but external annotation on managed resource does not exist",
			args: args{
				ctx: context.Background(),
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
			},
			want: want{
				// if the service account key had already been provisioned, external name annotation should have been set
				o:  managed.ExternalObservation{ResourceExists: false},
				mg: newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
			},
		},
		"GoogleCloudAPIReadErrorNotFound": {
			reason: "both SA rrn & external name annotations are set but Google Cloud API returns HTTP 404 (not found) for the specified key",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
			},
			want: want{
				// external name annotation is set but iam.ServiceAccountKey is not yet provisioned on Google side
				o: managed.ExternalObservation{ResourceExists: false},
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
			},
		},
		"GoogleCloudAPIReadError500": {
			reason: "both SA rrn & external name annotations are set but Google Cloud API returns HTTP 500 while fetching the specified key",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
			},
			want: want{
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
				err: errors.Wrap(gError(http.StatusInternalServerError, ""), errGetServiceAccountKey),
			},
		},
		"GoogleCloudAPIReadSuccessKeyIDParseError": {
			reason: "both SA rrn & external name annotations are set and Google Cloud API successfully returns the specified key but we cannot parse the keyID successfully",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(iamv1.ServiceAccountKey{Name: rrnInvalidServiceAccountKey}); err != nil {
					t.Logf(
						"Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)

					w.WriteHeader(http.StatusInternalServerError)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
			},
			want: want{
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
				err: errors.Wrap(fmt.Errorf(getURLParseErrorString(t, rrnInvalidServiceAccountKey)), errGetServiceAccountKey),
			},
		},
		"GoogleCloudAPIReadSuccessWithPublicKey": {
			reason: "both SA rrn & external name annotations are set and Google Cloud API successfully returns the specified key and we can parse the keyID",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(getIAMSaKeyGetObjectWithEncodedKeyData(iamSaKeyGetObject)); err != nil {
					t.Logf(
						"Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)

					w.WriteHeader(http.StatusInternalServerError)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
					setPublicKeyType(valIAMPublicKeyType)),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: map[string][]byte{
						keyPublicKeyType: []byte(valIAMPublicKeyType),
						keyPublicKeyData: []byte(valIAMPublicKeyData),
					},
				},
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
					setPublicKeyType(valIAMPublicKeyType),
					setObservedIAMServiceAccountKey(&iamSaKeyGetObject, nameExternalServiceAccountKey),
					setConditions(v1.Available()),
				),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			s, err := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			if err != nil {
				t.Fatalf("iam.NewService failed while running test case %q: %s", name, err)
			}

			c := &serviceAccountKeyExternalClient{serviceAccountKeyClient: iamv1.NewProjectsServiceAccountsKeysService(s)}
			got, err := c.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Observe(...): -want error, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("%s\nc.Observe(...): -want observation, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("%s\nc.Observe(...): -want managed resource, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestServiceAccountKeyCreate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		c   managed.ExternalCreation
		mg  resource.Managed
		err error
	}

	testCases := map[string]struct {
		reason  string
		handler http.Handler
		args    args
		want    want
	}{
		"NotServiceAccountKey": {
			reason: "assert error if not reconciling on a valid v1alpha1.ServiceAccountKey object",
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccountKey),
			},
		},
		"GoogleCloudAPIReadError500": {
			reason: "SA rrn is set but Google Cloud API returns HTTP 500 while fetching the specified key",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			args: args{
				ctx: context.Background(),
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
			},
			want: want{
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
				err: errors.Wrap(gError(http.StatusInternalServerError, ""), errCreateServiceAccountKey),
			},
		},
		"GoogleCloudAPIReadSuccessKeyIDParseError": {
			reason: "SA rrn is set and Google Cloud API successfully returns the specified key but we cannot parse the keyID successfully",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(iamv1.ServiceAccountKey{Name: rrnInvalidServiceAccountKey}); err != nil {
					t.Logf("Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
			},
			want: want{
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
				err: errors.Wrap(fmt.Errorf(getURLParseErrorString(t, rrnInvalidServiceAccountKey)), errCreateServiceAccountKey),
			},
		},
		"GoogleCloudAPIReadSuccessWithPublicKey": {
			reason: "SA rrn is set and Google Cloud API successfully returns the specified key and we can parse the keyID",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(
					getIAMSaKeyGetObjectWithEncodedKeyData(iamSaKeyCreateObject)); err != nil {
					t.Logf(
						"Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)

					w.WriteHeader(http.StatusInternalServerError)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setPublicKeyType(valIAMPublicKeyType),
					setPrivateKeyType(valIAMPrivateKeyType),
				),
			},
			want: want{
				c: managed.ExternalCreation{
					ExternalNameAssigned: true,
					ConnectionDetails: map[string][]byte{
						keyPublicKeyType: []byte(valIAMPublicKeyType),
						keyPublicKeyData: []byte(valIAMPublicKeyData),
						// private key data is available in iam.ServiceAccountKey.create response, and hence
						// is expected to be available in connection details
						keyPrivateKeyType: []byte(valIAMPrivateKeyType),
						keyPrivateKeyData: []byte(valIAMPrivateKeyData),
					},
				},
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
					setPublicKeyType(valIAMPublicKeyType),
					setPrivateKeyType(valIAMPrivateKeyType),
				),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			s, err := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			if err != nil {
				t.Fatalf("iam.NewService failed while running test case %q: %s", name, err)
			}

			c := &serviceAccountKeyExternalClient{serviceAccountKeyClient: iamv1.NewProjectsServiceAccountsKeysService(s)}
			got, err := c.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Create(...): -want error, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.c, got); diff != "" {
				t.Errorf("%s\nc.Create(...): -want creation, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("%s\nc.Create(...): -want managed resource, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestServiceAccountKeyUpdate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		u   managed.ExternalUpdate
		mg  resource.Managed
		err error
	}

	testCases := map[string]struct {
		reason  string
		handler http.Handler
		args    args
		want    want
	}{
		"NoOpUpdate": {
			reason: "assert update is a no-op",
			args: args{
				ctx: context.Background(),
				mg:  nil,
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			s, err := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			if err != nil {
				t.Fatalf("iam.NewService failed while running test case %q: %s", name, err)
			}

			c := &serviceAccountKeyExternalClient{serviceAccountKeyClient: iamv1.NewProjectsServiceAccountsKeysService(s)}
			got, err := c.Update(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Update(...): -want error, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.u, got); diff != "" {
				t.Errorf("%s\nc.Update(...): -want update, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("%s\nc.Update(...): -want managed resource, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestServiceAccountKeyDelete(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	testCases := map[string]struct {
		reason  string
		handler http.Handler
		args    args
		want    want
	}{
		"NotServiceAccountKey": {
			reason: "assert error if not reconciling on a valid v1alpha1.ServiceAccountKey object",
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccountKey),
			},
		},
		"GoogleCloudAPINotFound": {
			reason: "report no errors if Google Cloud API returns HTTP 404 for the resource being deleted (already deleted)",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
			},
			want: want{
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
				err: nil,
			},
		},
		"GoogleCloudAPIDeleteError500": {
			reason: "Unexpected errors returned by the GCP API should be wrapped and returned.",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
			},
			want: want{
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
				err: errors.Wrap(gError(http.StatusInternalServerError, ""), errDeleteServiceAccountKey),
			},
		},
		"GoogleCloudAPIDeleteSuccess": {
			reason: "Successful deletion should not return an error.",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(iamv1.Empty{}); err != nil {
					t.Logf("Google Cloud API response failed. Failed to serialize iam.Empty: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
			},
			want: want{
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
				err: nil,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			s, err := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			if err != nil {
				t.Fatalf("iam.NewService failed while running test case %q: %s", name, err)
			}

			c := &serviceAccountKeyExternalClient{serviceAccountKeyClient: iamv1.NewProjectsServiceAccountsKeysService(s)}
			err = c.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Delete(...): -want error, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("%s\nc.Delete(...): -want managed resource, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

type serviceAccountKeyModifier func(key *v1alpha1.ServiceAccountKey)

func newServiceAccountKey(modifiers ...serviceAccountKeyModifier) *v1alpha1.ServiceAccountKey {
	pubKeyType := valIAMPublicKeyType

	saKey := &v1alpha1.ServiceAccountKey{
		ObjectMeta: metav1.ObjectMeta{
			Name: nameTestServiceAccountKey,
		},
		Spec: v1alpha1.ServiceAccountKeySpec{
			ForProvider: v1alpha1.ServiceAccountKeyParameters{
				PublicKeyType: &pubKeyType,
			},
		},
	}

	for _, m := range modifiers {
		m(saKey)
	}

	return saKey
}

func setServiceAccount(saPath string) serviceAccountKeyModifier {
	return func(saKey *v1alpha1.ServiceAccountKey) {
		saKey.Spec.ForProvider.ServiceAccount = &saPath
	}
}

func setAnnotations(annotations map[string]string) serviceAccountKeyModifier {
	return func(saKey *v1alpha1.ServiceAccountKey) {
		saKey.ObjectMeta.Annotations = annotations
	}
}

func setPublicKeyType(publicKeyType string) serviceAccountKeyModifier {
	return func(saKey *v1alpha1.ServiceAccountKey) {
		saKey.Spec.ForProvider.PublicKeyType = &publicKeyType
	}
}

func setPrivateKeyType(privateKeyType string) serviceAccountKeyModifier {
	return func(saKey *v1alpha1.ServiceAccountKey) {
		saKey.Spec.ForProvider.PrivateKeyType = &privateKeyType
	}
}

func setObservedIAMServiceAccountKey(provider *iamv1.ServiceAccountKey, keyID string) serviceAccountKeyModifier {
	return func(saKey *v1alpha1.ServiceAccountKey) {
		saKey.Status.AtProvider.KeyID = keyID
		saKey.Status.AtProvider.KeyOrigin = provider.KeyOrigin
		saKey.Status.AtProvider.KeyAlgorithm = provider.KeyAlgorithm
		saKey.Status.AtProvider.KeyType = provider.KeyType
		saKey.Status.AtProvider.Name = provider.Name
		saKey.Status.AtProvider.PrivateKeyType = provider.PrivateKeyType
		saKey.Status.AtProvider.ValidAfterTime = provider.ValidAfterTime
		saKey.Status.AtProvider.ValidBeforeTime = provider.ValidBeforeTime
	}
}

func setConditions(conditions ...v1.Condition) serviceAccountKeyModifier {
	return func(saKey *v1alpha1.ServiceAccountKey) {
		for _, c := range conditions {
			saKey.Status.SetConditions(c)
		}
	}
}

func getURLParseErrorString(t *testing.T, invalidURL string) string {
	_, err := url.Parse(invalidURL)

	if err == nil {
		t.Fatalf("Expecting %q to be an invalid URL", invalidURL)
	}

	return err.Error()
}

func getIAMSaKeyGetObjectWithEncodedKeyData(srcSaKey iamv1.ServiceAccountKey) *iamv1.ServiceAccountKey {
	result := &iamv1.ServiceAccountKey{}

	*result = srcSaKey

	if result.PublicKeyData != "" {
		result.PublicKeyData = base64.StdEncoding.EncodeToString([]byte(result.PublicKeyData))
	}

	if result.PrivateKeyData != "" {
		result.PrivateKeyData = base64.StdEncoding.EncodeToString([]byte(result.PrivateKeyData))
	}

	return result
}
