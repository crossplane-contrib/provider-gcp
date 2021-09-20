package kms

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	kmsv1 "google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/googleapi"
	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/kms/v1alpha1"
)

const (
	ckMetadataName = "test-cryptoKey"
)

var (
	parentKeyRing = "projects/test-project/locations/test-location/keyRings/test-keyring"
	keyRingRRN    = fmt.Sprintf("%s/cryptoKeys/%s", parentKeyRing, ckMetadataName)
)

type ckValueModifier func(ring *v1alpha1.CryptoKey)

func ckWithName(s string) ckValueModifier {
	return func(i *v1alpha1.CryptoKey) { i.Name = s }
}

func ckWithKeyRing(s string) ckValueModifier {
	return func(i *v1alpha1.CryptoKey) { i.Spec.ForProvider.KeyRing = &s }
}

func ckWithRotationPeriod(s string) ckValueModifier {
	return func(i *v1alpha1.CryptoKey) { i.Spec.ForProvider.RotationPeriod = &s }
}

func ckWithAtProviderName(s string) ckValueModifier {
	return func(i *v1alpha1.CryptoKey) { i.Status.AtProvider.Name = s }
}

func ckWithExternalNameAnnotation(externalName string) ckValueModifier {
	return func(i *v1alpha1.CryptoKey) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[keyExternalName] = externalName
	}
}

func ckWithCondition(condition xpv1.Condition) ckValueModifier {
	return func(i *v1alpha1.CryptoKey) { i.SetConditions(condition) }
}

func ckWithDeletionTimestamp(ts metav1.Time) ckValueModifier {
	return func(i *v1alpha1.CryptoKey) { i.SetDeletionTimestamp(&ts) }
}

func cryptoKey(im ...ckValueModifier) *v1alpha1.CryptoKey {
	ck := &v1alpha1.CryptoKey{
		ObjectMeta: metav1.ObjectMeta{
			Name:       ckMetadataName,
			Finalizers: []string{},
		},
		Spec: v1alpha1.CryptoKeySpec{
			ResourceSpec: xpv1.ResourceSpec{
				WriteConnectionSecretToReference: &xpv1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1alpha1.CryptoKeyParameters{
				KeyRing: &parentKeyRing,
				Purpose: "ENCRYPT_DECRYPT",
			},
		},
	}

	for _, m := range im {
		m(ck)
	}

	return ck
}

func gError(code int, message string) error {
	return googleapi.CheckResponse(&http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(strings.NewReader(message)),
	})
}

func TestCryptoKeyRRN(t *testing.T) {
	type args struct {
		mg *v1alpha1.CryptoKey
	}

	type want struct {
		keyRingRRN string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NoKeyRingSet": {
			args: args{
				mg: cryptoKey(ckWithKeyRing("")),
			},
			want: want{
				keyRingRRN: "/cryptoKeys/",
			},
		},
		"ValidCR": {
			args: args{
				mg: cryptoKey(ckWithExternalNameAnnotation(ckMetadataName)),
			},
			want: want{
				keyRingRRN: keyRingRRN,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			kr := cryptoKeyRRN(tc.args.mg)
			if diff := cmp.Diff(tc.want.keyRingRRN, kr, test.EquateConditions()); diff != "" {
				t.Errorf("cryptoKeyRRN(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCryptoKeyObserve(t *testing.T) {
	now := metav1.Now()

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg          resource.Managed
		observation managed.ExternalObservation
		err         error
	}
	cases := map[string]struct {
		handler http.Handler
		args    args
		want    want
	}{
		"ObservedCryptoKeyGot": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys/get
				if !strings.HasSuffix(r.URL.Path, keyRingRRN) {
					t.Errorf("requested URL.Path to get keyRing should end with: %s, got %s instead",
						keyRingRRN, r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				ck := &kmsv1.CryptoKey{
					Name:    keyRingRRN,
					Purpose: "ENCRYPT_DECRYPT",
				}
				_ = json.NewEncoder(w).Encode(ck)
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
				),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithAtProviderName(keyRingRRN),
					ckWithCondition(xpv1.Available())),
				observation: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"ObservedCryptoKeyGotButCRDeleted": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				sa := &kmsv1.CryptoKey{
					Name: fqName,
				}
				_ = json.NewEncoder(w).Encode(sa)
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithDeletionTimestamp(now),
				),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithDeletionTimestamp(now)),
				observation: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
			},
		},
		"ObservedCryptoKeyDoesNotExist": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg:  cryptoKey(),
			},
			want: want{
				mg:          cryptoKey(),
				observation: managed.ExternalObservation{},
			},
		},
		"NotCryptoKey": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotCryptoKey),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			cryptokeys := kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)
			e := &cryptoKeyExternal{cryptokeys: cryptokeys}
			obs, err := e.Observe(context.Background(), tc.args.mg)

			if err != nil {
				if tc.want.err != nil {
					// we expected a different error than we got
					if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
						t.Errorf("Observe(...): want error string != got error string:\n%s", diff)
					}
				} else {
					t.Errorf("Observe(...): unexpected error %s", err)
				}
			} else {
				if tc.want.err != nil {
					t.Errorf("Observe(...) want error %s got nil", tc.want.err)
				}
			}

			if diff := cmp.Diff(tc.want.observation, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCryptoKeyCreate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		args    args
		want    want
	}{
		"CreatedCryptoKey": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				id := r.URL.Query()["cryptoKeyId"]
				if diff := cmp.Diff(id[0], ckMetadataName); diff != "" {
					t.Errorf("cryptoKeyId: -want, +got:\n%s", diff)
				}
				kr := &kmsv1.CryptoKey{
					Name: keyRingRRN,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(kr)
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName)),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithCondition(xpv1.Creating())),
			},
		},
		"NotCryptoKey": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotCryptoKey),
			},
		},
		"FailedToCreateCryptoKey": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&iamv1.Empty{})
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName)),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithCondition(xpv1.Creating())),
				err: errors.Wrap(err500, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			cryptokeys := kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)
			e := &cryptoKeyExternal{cryptokeys: cryptokeys}
			_, err := e.Create(context.Background(), tc.args.mg)

			if err != nil {
				if tc.want.err != nil {
					// we expected a different error than we got
					if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
						t.Errorf("Create(...): want error string != got error string:\n%s", diff)
					}
				} else {
					t.Errorf("Create(...): unexpected error %s", err)
				}
			} else {
				if tc.want.err != nil {
					t.Errorf("Create(...) want error %s got nil", tc.want.err)
				}
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCryptoKeyUpdate(t *testing.T) {
	rotationPeriod := "2592000s"

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		args    args
		want    want
	}{
		"UpdatedCryptoKey_UpdatesRotationPeriod": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					kr := &kmsv1.CryptoKey{
						Name:    keyRingRRN,
						Purpose: "ENCRYPT_DECRYPT",
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(kr)
				case http.MethodPatch:
					// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys/patch
					if !strings.HasSuffix(r.URL.Path, keyRingRRN) {
						t.Errorf("requested URL.Path to get keyRing should end with: %s, got %s instead",
							keyRingRRN, r.URL.Path)
					}
					id := r.URL.Query()["updateMask"]
					if diff := cmp.Diff(id[0], "rotationPeriod"); diff != "" {
						t.Errorf("updateMask: -want, +got:\n%s", diff)
					}
					kr := &kmsv1.CryptoKey{
						Name:           keyRingRRN,
						Purpose:        "ENCRYPT_DECRYPT",
						RotationPeriod: rotationPeriod,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(kr)
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&kmsv1.CryptoKey{})
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithRotationPeriod(rotationPeriod),
					ckWithExternalNameAnnotation(ckMetadataName)),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithRotationPeriod(rotationPeriod),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithCondition(xpv1.Creating())),
			},
		},
		"AlreadyUpToDate_NoPatchCalled": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					kr := &kmsv1.CryptoKey{
						Name:    keyRingRRN,
						Purpose: "ENCRYPT_DECRYPT",
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(kr)
				case http.MethodPatch:
					t.Errorf("should not call patch when already up to date")
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&kmsv1.CryptoKey{})
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName)),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithCondition(xpv1.Creating())),
			},
		},
		"FailedToCheckDifference": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&kmsv1.CryptoKey{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&kmsv1.CryptoKey{})
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName)),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithCondition(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusBadRequest, "{}\n"), errGet),
			},
		},
		"FailedToPatch_UpdatesRotationPeriod": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					kr := &kmsv1.CryptoKey{
						Name:    keyRingRRN,
						Purpose: "ENCRYPT_DECRYPT",
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(kr)
				case http.MethodPatch:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&kmsv1.CryptoKey{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&kmsv1.CryptoKey{})
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithRotationPeriod(rotationPeriod),
					ckWithExternalNameAnnotation(ckMetadataName)),
			},
			want: want{
				mg: cryptoKey(
					ckWithName(ckMetadataName),
					ckWithRotationPeriod(rotationPeriod),
					ckWithExternalNameAnnotation(ckMetadataName),
					ckWithCondition(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusBadRequest, "{}\n"), errUpdate),
			},
		},
		"NotCryptoKey": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotCryptoKey),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			cryptokeys := kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)
			e := &cryptoKeyExternal{cryptokeys: cryptokeys}
			_, err := e.Update(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Update(...): want error != got error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Observe(...): want error != got error:\n%s", diff)
				}
			}
		})
	}
}

func TestCryptoKeyDelete(t *testing.T) {
	type want struct {
		err error
	}

	cases := map[string]struct {
		want want
	}{
		"ReturnsNoErr": {
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &cryptoKeyExternal{}
			err := e.Delete(context.Background(), keyRing())
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Update(...): want error != got error:\n%s", diff)
			}
		})
	}
}
