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
	"github.com/google/go-cmp/cmp/cmpopts"
	kmsv1 "google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	iamv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane-contrib/provider-gcp/apis/kms/v1alpha1"
	"github.com/crossplane-contrib/provider-gcp/pkg/clients/cryptokeypolicy"
)

const (
	ckpMetadataName = "test-CryptoKeyPolicy"
)

var (
	testCryptoKeyRRN = "projects/my-projects/locations/global/keyRings/hello-from-crossplane/cryptoKeys/crossplane-test-key"

	testMember = "serviceAccount:perfect-test-sa@my-project.iam.gserviceaccount.com"
	testRole   = "roles/crossplane.unitTester"
)

type ckpValueModifier func(ring *v1alpha1.CryptoKeyPolicy)

func ckpWithName(s string) ckpValueModifier {
	return func(i *v1alpha1.CryptoKeyPolicy) { i.Name = s }
}

func ckpWithExternalNameAnnotation(externalName string) ckpValueModifier {
	return func(i *v1alpha1.CryptoKeyPolicy) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[keyExternalName] = externalName
	}
}

func ckpWithCondition(condition xpv1.Condition) ckpValueModifier {
	return func(i *v1alpha1.CryptoKeyPolicy) { i.SetConditions(condition) }
}

func ckpWithBinding(binding *iamv1alpha1.Binding) ckpValueModifier {
	return func(i *v1alpha1.CryptoKeyPolicy) {
		i.Spec.ForProvider.Policy.Bindings = append(i.Spec.ForProvider.Policy.Bindings, binding)
	}
}

func CryptoKeyPolicy(im ...ckpValueModifier) *v1alpha1.CryptoKeyPolicy {
	ckp := &v1alpha1.CryptoKeyPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       ckpMetadataName,
			Finalizers: []string{},
		},
		Spec: v1alpha1.CryptoKeyPolicySpec{
			ForProvider: v1alpha1.CryptoKeyPolicyParameters{
				CryptoKey: &testCryptoKeyRRN,
				Policy: iamv1alpha1.Policy{
					Bindings: []*iamv1alpha1.Binding{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				},
			},
		},
	}

	for _, m := range im {
		m(ckp)
	}

	return ckp
}

func TestCryptoKeyPolicyObserve(t *testing.T) {
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
		"NotCryptoKeyPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotCryptoKeyPolicy),
			},
		},
		"FailedToObserve": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&kmsv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
				),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName)),
				observation: managed.ExternalObservation{},
				err:         errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"ObservedPolicyEmpty": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				ckp := &kmsv1.Policy{}
				_ = json.NewEncoder(w).Encode(ckp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
				),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName)),
				observation: managed.ExternalObservation{},
			},
		},
		"ObservedPolicyNeedsUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ckp := &kmsv1.Policy{
					Bindings: []*kmsv1.Binding{
						{
							Members: []string{"some-other-member"},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(ckp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
				),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName)),
				observation: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"ObservedPolicyUpToDate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys/getIamPolicy
				expectedEp := fmt.Sprintf("/v1/%s:getIamPolicy", testCryptoKeyRRN)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				ckp := &kmsv1.Policy{
					Bindings: []*kmsv1.Binding{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(ckp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
				),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithCondition(xpv1.Available()),
					ckpWithName(ckpMetadataName)),
				observation: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			cryptokeys := kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)
			e := &cryptoKeyPolicyExternal{cryptokeyspolicy: cryptokeys}
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

func TestCryptoKeyPolicyCreate(t *testing.T) {
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
		"NotCryptoKeyPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotCryptoKeyPolicy),
			},
		},
		"CreateSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys/setIamPolicy
				expectedEp := fmt.Sprintf("/v1/%s:setIamPolicy", testCryptoKeyRRN)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				i := &kmsv1.SetIamPolicyRequest{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				exp := &kmsv1.Policy{
					Bindings: []*kmsv1.Binding{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				if !cryptokeypolicy.ArePoliciesSame(exp, i.Policy) {
					t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(exp, i.Policy, cmpopts.IgnoreFields(kmsv1.Policy{}, "Version")))
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(exp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Creating())),
			},
		},
		"CreateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&kmsv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			cryptokeys := kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)
			e := &cryptoKeyPolicyExternal{cryptokeyspolicy: cryptokeys}
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

func TestCryptoKeyPolicyUpdate(t *testing.T) {
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
		"NotCryptoKeyPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotCryptoKeyPolicy),
			},
		},
		"UpdateSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var ckp *kmsv1.Policy
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					ckp = &kmsv1.Policy{
						Bindings: []*kmsv1.Binding{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
				case http.MethodPost:
					i := &kmsv1.SetIamPolicyRequest{}
					b, err := ioutil.ReadAll(r.Body)
					if diff := cmp.Diff(err, nil); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					err = json.Unmarshal(b, i)
					if diff := cmp.Diff(err, nil); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					ckp = &kmsv1.Policy{
						Bindings: []*kmsv1.Binding{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
							{
								Members: []string{"another-member"},
								Role:    "another-role",
							},
						},
					}
					if !cryptokeypolicy.ArePoliciesSame(ckp, i.Policy) {
						t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(ckp, i.Policy, cmpopts.IgnoreFields(kmsv1.Policy{}, "Version")))
					}
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(ckp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Available()),
					ckpWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Available()),
					ckpWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
		},
		"FailedToGet": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&kmsv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"AlreadyUpToDate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ckp := &kmsv1.Policy{
					Bindings: []*kmsv1.Binding{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(ckp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Available())),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Available())),
			},
		},
		"FailedToUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var ckp *kmsv1.Policy
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					ckp = &kmsv1.Policy{
						Bindings: []*kmsv1.Binding{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
				case http.MethodPost:
					ckp = &kmsv1.Policy{}
					w.WriteHeader(http.StatusInternalServerError)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(ckp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Available()),
					ckpWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName),
					ckpWithCondition(xpv1.Available()),
					ckpWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			cryptokeys := kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)
			e := &cryptoKeyPolicyExternal{cryptokeyspolicy: cryptokeys}
			_, err := e.Update(context.Background(), tc.args.mg)
			if err != nil {
				if tc.want.err != nil {
					// we expected a different error than we got
					if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
						t.Errorf("Update(...): want error string != got error string:\n%s", diff)
					}
				} else {
					t.Errorf("Update(...): unexpected error %s", err)
				}
			} else {
				if tc.want.err != nil {
					t.Errorf("Update(...) want error %s got nil", tc.want.err)
				}
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCryptoKeyPolicyDelete(t *testing.T) {
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
		"NotCryptoKeyPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotCryptoKeyPolicy),
			},
		},
		"DeleteSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys/setIamPolicy
				expectedEp := fmt.Sprintf("/v1/%s:setIamPolicy", testCryptoKeyRRN)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				i := &kmsv1.SetIamPolicyRequest{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				exp := &kmsv1.Policy{}
				if !cryptokeypolicy.ArePoliciesSame(exp, i.Policy) {
					t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(exp, i.Policy, cmpopts.IgnoreFields(kmsv1.Policy{}, "Version")))
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(exp)
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
			},
		},
		"CreateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&kmsv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
			},
			want: want{
				mg: CryptoKeyPolicy(
					ckpWithName(ckpMetadataName),
					ckpWithExternalNameAnnotation(ckpMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			cryptokeys := kmsv1.NewProjectsLocationsKeyRingsCryptoKeysService(s)
			e := &cryptoKeyPolicyExternal{cryptokeyspolicy: cryptokeys}
			err := e.Delete(context.Background(), tc.args.mg)
			if err != nil {
				if tc.want.err != nil {
					// we expected a different error than we got
					if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
						t.Errorf("Update(...): want error string != got error string:\n%s", diff)
					}
				} else {
					t.Errorf("Update(...): unexpected error %s", err)
				}
			} else {
				if tc.want.err != nil {
					t.Errorf("Update(...) want error %s got nil", tc.want.err)
				}
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
		})
	}
}
