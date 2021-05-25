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

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	testCases := map[string]struct {
		args *saKeyTestArgs
		want *saKeyTestWant
	}{
		// assert error if not reconciling on a valid v1alpha1.ServiceAccountKey object
		"NotServiceAccountKey": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: &saKeyTestWant{
				mg:       nil,
				expected: []interface{}{nil, errors.New(errNotServiceAccountKey)},
			},
		},
		// assert error if a valid service account reference is not provided
		"InvalidServiceAccountReference": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  newServiceAccountKey(),
			},
			want: &saKeyTestWant{
				mg:       nil,
				expected: []interface{}{managed.ExternalObservation{}, nil},
			},
		},
		// SA rrn (resource path) is set but external annotation on managed resource does not exist
		"ExternalNameNotSet": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{managed.ExternalObservation{
					// if the service account key had already been provisioned, external name annotation should have been set
					ResourceExists: false,
				}, nil},
			},
		},
		// both SA rrn & external name annotations are set but Google Cloud API returns HTTP 404 (not found) for the specified key
		"GoogleCloudAPIReadErrorNotFound": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{managed.ExternalObservation{
					// external name annotation is set but iam.ServiceAccountKey is not yet provisioned on Google side
					ResourceExists: false,
				}, nil},
			},
		},
		// both SA rrn & external name annotations are set but Google Cloud API returns HTTP 500 while fetching the specified key
		"GoogleCloudAPIReadError500": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{
					nil,
					errors.Wrap(gError(http.StatusInternalServerError, ""), errGetServiceAccountKey)},
			},
		},
		// both SA rrn & external name annotations are set and Google Cloud API successfully returns the specified key
		// but we cannot parse the keyID successfully
		"GoogleCloudAPIReadSuccessKeyIDParseError": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					})),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := json.NewEncoder(w).Encode(iamv1.ServiceAccountKey{
						Name: rrnInvalidServiceAccountKey,
					}); err != nil {
						t.Logf(
							"Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)

						w.WriteHeader(http.StatusInternalServerError)
					}
				}),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{
					nil,
					errors.Wrap(fmt.Errorf(getURLParseErrorString(t, rrnInvalidServiceAccountKey)),
						errGetServiceAccountKey),
				},
			},
		},
		// both SA rrn & external name annotations are set and Google Cloud API successfully returns the specified key
		// and we can parse the keyID
		"GoogleCloudAPIReadSuccessWithPublicKey": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
					setPublicKeyType(valIAMPublicKeyType)),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := json.NewEncoder(w).Encode(getIAMSaKeyGetObjectWithEncodedKeyData(iamSaKeyGetObject)); err != nil {
						t.Logf(
							"Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)

						w.WriteHeader(http.StatusInternalServerError)
					}
				}),
			},
			want: &saKeyTestWant{
				expected: []interface{}{managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: map[string][]byte{
						keyPublicKeyType: []byte(valIAMPublicKeyType),
						keyPublicKeyData: []byte(valIAMPublicKeyData),
					},
				},
					nil,
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

	for n, tc := range testCases {
		args := tc.args
		want := tc.want

		runCase(t, n, func(c *serviceAccountKeyExternalClient, args *saKeyTestArgs) []interface{} {
			o, err := c.Observe(args.ctx, args.mg)

			return []interface{}{o, err}
		}, func(t *testing.T, expected, observed []interface{}) {
			compareValueErrorResult(t, expected, observed, "Observe")
		}, args, want)
	}
}

func TestServiceAccountKeyCreate(t *testing.T) {
	testCases := map[string]struct {
		args *saKeyTestArgs
		want *saKeyTestWant
	}{
		// assert error if not reconciling on a valid v1alpha1.ServiceAccountKey object
		"NotServiceAccountKey": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: &saKeyTestWant{
				mg:       nil,
				expected: []interface{}{nil, errors.New(errNotServiceAccountKey)},
			},
		},
		// assert error if a valid service account reference is not provided
		"InvalidServiceAccountReference": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  newServiceAccountKey(),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{nil, errors.Wrap(fmt.Errorf(fmtErrInvalidServiceAccountRef,
					v1alpha1.ServiceAccountReferer{}), errCreateServiceAccountKey)},
			},
		},
		// SA rrn is set but Google Cloud API returns HTTP 500 while fetching the specified key
		"GoogleCloudAPIReadError500": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{
					nil,
					errors.Wrap(gError(http.StatusInternalServerError, ""), errCreateServiceAccountKey)},
			},
		},
		// SA rrn is set and Google Cloud API successfully returns the specified key
		// but we cannot parse the keyID successfully
		"GoogleCloudAPIReadSuccessKeyIDParseError": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := json.NewEncoder(w).Encode(iamv1.ServiceAccountKey{
						Name: rrnInvalidServiceAccountKey,
					}); err != nil {
						t.Logf(
							"Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)

						w.WriteHeader(http.StatusInternalServerError)
					}
				}),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{
					nil,
					errors.Wrap(fmt.Errorf(getURLParseErrorString(t, rrnInvalidServiceAccountKey)),
						errCreateServiceAccountKey),
				},
			},
		},
		// SA rrn is set and Google Cloud API successfully returns the specified key
		// and we can parse the keyID
		"GoogleCloudAPIReadSuccessWithPublicKey": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setPublicKeyType(valIAMPublicKeyType),
					setPrivateKeyType(valIAMPrivateKeyType),
				),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := json.NewEncoder(w).Encode(
						getIAMSaKeyGetObjectWithEncodedKeyData(iamSaKeyCreateObject)); err != nil {
						t.Logf(
							"Google Cloud API response failed. Failed to serialize iam.ServiceAccountKey: %s", err)

						w.WriteHeader(http.StatusInternalServerError)
					}
				}),
			},
			want: &saKeyTestWant{
				expected: []interface{}{managed.ExternalCreation{
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
					nil,
				},
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
					setPublicKeyType(valIAMPublicKeyType),
					setPrivateKeyType(valIAMPrivateKeyType),
					setConditions(v1.Creating()),
				),
			},
		},
	}

	for n, tc := range testCases {
		args := tc.args
		want := tc.want

		runCase(t, n, func(c *serviceAccountKeyExternalClient, args *saKeyTestArgs) []interface{} {
			o, err := c.Create(args.ctx, args.mg)

			return []interface{}{o, err}
		}, func(t *testing.T, expected, observed []interface{}) {
			compareValueErrorResult(t, expected, observed, "Create")
		}, args, want)
	}
}

func TestServiceAccountKeyUpdate(t *testing.T) {
	testCases := map[string]struct {
		args *saKeyTestArgs
		want *saKeyTestWant
	}{
		// assert update is a no-op
		"NoOpUpdate": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  nil,
			},
			want: &saKeyTestWant{
				mg:       nil,
				expected: []interface{}{managed.ExternalUpdate{}, nil},
			},
		},
	}

	for n, tc := range testCases {
		args := tc.args
		want := tc.want

		runCase(t, n, func(c *serviceAccountKeyExternalClient, args *saKeyTestArgs) []interface{} {
			o, err := c.Update(args.ctx, args.mg)

			return []interface{}{o, err}
		}, func(t *testing.T, expected, observed []interface{}) {
			compareValueErrorResult(t, expected, observed, "Update")
		}, args, want)
	}
}

func TestServiceAccountKeyDelete(t *testing.T) {
	testCases := map[string]struct {
		args *saKeyTestArgs
		want *saKeyTestWant
	}{
		// assert error if not reconciling on a valid v1alpha1.ServiceAccountKey object
		"NotServiceAccountKey": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: &saKeyTestWant{
				mg:       nil,
				expected: []interface{}{errors.New(errNotServiceAccountKey)},
			},
		},
		// assert error if a valid service account reference is not provided
		"InvalidServiceAccountReference": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  newServiceAccountKey(),
			},
			want: &saKeyTestWant{
				mg: nil,
				expected: []interface{}{errors.Wrap(fmt.Errorf(fmtErrInvalidServiceAccountRef,
					v1alpha1.ServiceAccountReferer{}), errDeleteServiceAccountKey)},
			},
		},
		// no action is expected if external name annotation is missing
		"MissingExternalNameAnnotation": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg:  newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
			},
			want: &saKeyTestWant{
				mg:       newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
				expected: []interface{}{errors.Wrap(errors.New(errNoExternalName), errDeleteServiceAccountKey)},
			},
		},
		// report no errors if Google Cloud API returns HTTP 404 for the resource being deleted (already deleted)
		"GoogleCloudAPINotFound": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}),
			},
			want: &saKeyTestWant{
				expected: []interface{}{nil},
			},
		},
		"GoogleCloudAPIDeleteError500": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}),
			},
			want: &saKeyTestWant{
				expected: []interface{}{
					errors.Wrap(gError(http.StatusInternalServerError, ""), errDeleteServiceAccountKey)},
			},
		},
		"GoogleCloudAPIDeleteSuccess": {
			args: &saKeyTestArgs{
				ctx: context.Background(),
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := json.NewEncoder(w).Encode(iamv1.Empty{}); err != nil {
						t.Logf(
							"Google Cloud API response failed. Failed to serialize iam.Empty: %s", err)

						w.WriteHeader(http.StatusInternalServerError)
					}
				}),
			},
			want: &saKeyTestWant{
				expected: []interface{}{nil},
				mg: newServiceAccountKey(
					setServiceAccount(rrnTestServiceAccount),
					setAnnotations(map[string]string{
						meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
					}),
				),
			},
		},
	}

	for n, tc := range testCases {
		args := tc.args
		want := tc.want

		runCase(t, n, func(c *serviceAccountKeyExternalClient, args *saKeyTestArgs) []interface{} {
			return []interface{}{c.Delete(args.ctx, args.mg)}
		}, func(t *testing.T, expected, observed []interface{}) {
			compareErrorResult(t, expected, observed, 0, "Delete")
		}, args, want)
	}
}

func TestNewSaKeyRelativeResourceNamer(t *testing.T) {
	testCases := map[string]struct {
		saKey                *v1alpha1.ServiceAccountKey
		expectedSAPath       string
		expectedSAKeyPath    string
		expectedSAKeyPathErr error
		expectedSAPathErr    error
	}{
		"NoServiceAccountReference": {
			saKey:                newServiceAccountKey(),
			expectedSAPathErr:    fmt.Errorf(fmtErrInvalidServiceAccountRef, v1alpha1.ServiceAccountReferer{}),
			expectedSAKeyPathErr: fmt.Errorf(fmtErrInvalidServiceAccountRef, v1alpha1.ServiceAccountReferer{}),
		},
		"ValidServiceAccountReference": {
			saKey:                newServiceAccountKey(setServiceAccount(rrnTestServiceAccount)),
			expectedSAPath:       rrnTestServiceAccount,
			expectedSAKeyPathErr: errors.New(errNoExternalName),
		},
		"ValidServiceAccountReferenceWithExtName": {
			saKey: newServiceAccountKey(
				setServiceAccount(rrnTestServiceAccount),
				setAnnotations(map[string]string{
					meta.AnnotationKeyExternalName: nameExternalServiceAccountKey,
				}),
			),
			expectedSAPath:    rrnTestServiceAccount,
			expectedSAKeyPath: rrnTestServiceAccountKey,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			n, err := referencedServiceAccountPath(tc.saKey)

			if compareErrors(t, err, tc.expectedSAPathErr, "referencedServiceAccountPath.expectedSAPathErr") {
				return
			}

			compareValues(t, n, tc.expectedSAPath, "referencedServiceAccountPath.expectedSAPath")
			// check resourcePath
			saKeyRRN, err := resourcePath(tc.saKey)

			if compareErrors(t, err, tc.expectedSAKeyPathErr, "resourcePath.expectedSAKeyPathErr") {
				return
			}

			compareValues(t, saKeyRRN, tc.expectedSAKeyPath, "resourcePath.saKeyRRN")
		})
	}
}

type saKeyTestArgs struct {
	ctx     context.Context
	mg      resource.Managed
	handler http.Handler
}

type compareTestResultFunc func(t *testing.T, expected, observed []interface{})

type saKeyTestWant struct {
	mg       resource.Managed
	expected []interface{}
}

type testMethod func(c *serviceAccountKeyExternalClient, args *saKeyTestArgs) []interface{}

func runCase(t *testing.T, name string, test testMethod, compareResult compareTestResultFunc,
	args *saKeyTestArgs, want *saKeyTestWant) bool {
	return t.Run(name, func(t *testing.T) {
		server := httptest.NewServer(args.handler)

		defer server.Close()

		s, err := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL),
			option.WithoutAuthentication())

		if err != nil {
			t.Fatalf("iam.NewService failed while running test case %q: %s", name, err)
		}

		extClient := &serviceAccountKeyExternalClient{
			serviceAccountKeyClient: iamv1.NewProjectsServiceAccountsKeysService(s),
		}

		// assert expected return values of the tested method
		compareResult(t, want.expected, test(extClient, args))
		// assert expected reconciled managed resource if expected managed resource is specified
		if want.mg != nil {
			compareValues(t, args.mg, want.mg, name)
		}
	})
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

// returns true if the test function returned an error so that value comparison should not be done
func compareErrors(t *testing.T, err, errExpected error, context string) bool {
	if err == nil && errExpected == nil {
		return false
	}

	if err != nil && errExpected == nil {
		t.Fatalf("Unexpected %s error: %s", context, err)
	}

	if err == nil && errExpected != nil {
		t.Fatalf("Expected %s error but got nil: %s", context, errExpected)
	}

	if //goland:noinspection GoNilness
	diff := cmp.Diff(errExpected.Error(), err.Error()); diff != "" {
		t.Fatalf("Expected %s error string differs from actual error, -expected +got: %s", context, diff)
	}

	return true
}

func compareValues(t *testing.T, o, oExpected interface{}, context string) {
	if diff := cmp.Diff(oExpected, o, cmp.Comparer(compareStatusConditionsIgnoreTimestamp)); diff != "" {
		t.Fatalf("Expected %s return values differ: -expected, +got:\n%s", context, diff)
	}
}

// {expected,observed}[1] are errors
func compareErrorResult(t *testing.T, expected, observed []interface{}, index int, context string) bool {
	var err, errExpected error

	if observed[index] != nil {
		err = observed[index].(error)
	}

	if expected[index] != nil {
		errExpected = expected[index].(error)
	}

	return compareErrors(t, err, errExpected, context)
}

// {expected,observed}[0] are values, {expected,observed}[1] are errors
func compareValueErrorResult(t *testing.T, expected, observed []interface{}, context string) {
	if compareErrorResult(t, expected, observed, 1, context) {
		return // if the tested method errored, do not try to compare expected & actual value
	}

	compareValues(t, observed[0], expected[0], "Observe")
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

// compare v1.Conditions relaxing on equality of last transition times
func compareStatusConditionsIgnoreTimestamp(c1, c2 v1.Condition) bool {
	if (c1.LastTransitionTime.IsZero() && !c2.LastTransitionTime.IsZero()) ||
		(!c1.LastTransitionTime.IsZero() && c2.LastTransitionTime.IsZero()) {
		return false
	}

	// as long as both are zero or non-zero, ignore last transition times because they will probably different
	c1.LastTransitionTime.Time = c2.LastTransitionTime.Time.In(c2.LastTransitionTime.Time.Location())

	return cmp.Equal(c1, c2)
}
