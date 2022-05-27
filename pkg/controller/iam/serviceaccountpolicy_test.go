package iam

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
	"google.golang.org/api/googleapi"
	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gcp/apis/iam/v1alpha1"
	iamv1alpha1 "github.com/crossplane-contrib/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane-contrib/provider-gcp/pkg/clients/serviceaccountpolicy"
)

const (
	sapMetadataName = "test-service-account-policy"
	keyExternalName = "crossplane.io/external-name"
)

var (
	testServiceAccountRRN = "projects/wesaas-playground/serviceAccounts/perfect-test-sa@wesaas-playground.iam.gserviceaccount.com"

	testMember = "serviceAccount:perfect-test-sa@my-project.iam.gserviceaccount.com"
	testRole   = "roles/crossplane.unitTester"
)

type sapValueModifier func(ring *v1alpha1.ServiceAccountPolicy)

func sapWithName(s string) sapValueModifier {
	return func(i *v1alpha1.ServiceAccountPolicy) { i.Name = s }
}

func sapWithExternalNameAnnotation(externalName string) sapValueModifier {
	return func(i *v1alpha1.ServiceAccountPolicy) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[keyExternalName] = externalName
	}
}

func sapWithCondition(condition xpv1.Condition) sapValueModifier {
	return func(i *v1alpha1.ServiceAccountPolicy) { i.SetConditions(condition) }
}

func sapWithBinding(binding *iamv1alpha1.Binding) sapValueModifier {
	return func(i *v1alpha1.ServiceAccountPolicy) {
		i.Spec.ForProvider.Policy.Bindings = append(i.Spec.ForProvider.Policy.Bindings, binding)
	}
}

func gError(code int, message string) error {
	return googleapi.CheckResponse(&http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(strings.NewReader(message)),
	})
}

func ServiceAccountPolicy(im ...sapValueModifier) *v1alpha1.ServiceAccountPolicy {
	sap := &v1alpha1.ServiceAccountPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       sapMetadataName,
			Finalizers: []string{},
		},
		Spec: v1alpha1.ServiceAccountPolicySpec{
			ForProvider: v1alpha1.ServiceAccountPolicyParameters{
				ServiceAccountReferer: v1alpha1.ServiceAccountReferer{
					ServiceAccount: &testServiceAccountRRN,
				},
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
		m(sap)
	}

	return sap
}

func TestServiceAccountPolicyObserve(t *testing.T) {
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
		"NotServiceAccountPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccountPolicy),
			},
		},
		"FailedToObserve": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&iamv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
				),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName)),
				observation: managed.ExternalObservation{},
				err:         errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"ObservedPolicyEmpty": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				sap := &iamv1.Policy{}
				_ = json.NewEncoder(w).Encode(sap)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
				),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName)),
				observation: managed.ExternalObservation{},
			},
		},
		"ObservedPolicyNeedsUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sap := &iamv1.Policy{
					Bindings: []*iamv1.Binding{
						{
							Members: []string{"some-other-member"},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(sap)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
				),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName)),
				observation: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"ObservedPolicyUpToDate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/getIamPolicy
				expectedEp := fmt.Sprintf("/v1/%s:getIamPolicy", testServiceAccountRRN)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				sap := &iamv1.Policy{
					Bindings: []*iamv1.Binding{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(sap)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
				),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithCondition(xpv1.Available()),
					sapWithName(sapMetadataName)),
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
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceaccounts := iamv1.NewProjectsServiceAccountsService(s)
			e := &serviceAccountPolicyExternal{serviceaccountspolicy: serviceaccounts}
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

func TestServiceAccountPolicyCreate(t *testing.T) {
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
		"NotServiceAccountPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccountPolicy),
			},
		},
		"CreateSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/setIamPolicy
				expectedEp := fmt.Sprintf("/v1/%s:setIamPolicy", testServiceAccountRRN)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				i := &iamv1.SetIamPolicyRequest{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				exp := &iamv1.Policy{
					Bindings: []*iamv1.Binding{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				if !serviceaccountpolicy.ArePoliciesSame(exp, i.Policy) {
					t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(exp, i.Policy, cmpopts.IgnoreFields(iamv1.Policy{}, "Version")))
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(exp)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Creating())),
			},
		},
		"CreateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&iamv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceaccounts := iamv1.NewProjectsServiceAccountsService(s)
			e := &serviceAccountPolicyExternal{serviceaccountspolicy: serviceaccounts}
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

func TestServiceAccountPolicyUpdate(t *testing.T) {
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
		"NotServiceAccountPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccountPolicy),
			},
		},
		"UpdateSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var sap *iamv1.Policy
				defer r.Body.Close()
				switch r.URL.Path {
				case fmt.Sprintf("/v1/%s:getIamPolicy", testServiceAccountRRN):
					sap = &iamv1.Policy{
						Bindings: []*iamv1.Binding{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
				case fmt.Sprintf("/v1/%s:setIamPolicy", testServiceAccountRRN):
					i := &iamv1.SetIamPolicyRequest{}
					b, err := ioutil.ReadAll(r.Body)
					if diff := cmp.Diff(err, nil); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					err = json.Unmarshal(b, i)
					if diff := cmp.Diff(err, nil); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					sap = &iamv1.Policy{
						Bindings: []*iamv1.Binding{
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
					if !serviceaccountpolicy.ArePoliciesSame(sap, i.Policy) {
						t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(sap, i.Policy, cmpopts.IgnoreFields(iamv1.Policy{}, "Version")))
					}
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(sap)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Available()),
					sapWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Available()),
					sapWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
		},
		"FailedToGet": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&iamv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"AlreadyUpToDate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sap := &iamv1.Policy{
					Bindings: []*iamv1.Binding{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(sap)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Available())),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Available())),
			},
		},
		"FailedToUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var sap *iamv1.Policy
				defer r.Body.Close()
				switch r.URL.Path {
				case fmt.Sprintf("/v1/%s:getIamPolicy", testServiceAccountRRN):
					sap = &iamv1.Policy{
						Bindings: []*iamv1.Binding{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
				case fmt.Sprintf("/v1/%s:setIamPolicy", testServiceAccountRRN):
					sap = &iamv1.Policy{}
					w.WriteHeader(http.StatusInternalServerError)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(sap)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Available()),
					sapWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName),
					sapWithCondition(xpv1.Available()),
					sapWithBinding(&iamv1alpha1.Binding{
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
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceaccounts := iamv1.NewProjectsServiceAccountsService(s)
			e := &serviceAccountPolicyExternal{serviceaccountspolicy: serviceaccounts}
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

func TestServiceAccountPolicyDelete(t *testing.T) {
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
		"NotServiceAccountPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccountPolicy),
			},
		},
		"DeleteSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/setIamPolicy
				expectedEp := fmt.Sprintf("/v1/%s:setIamPolicy", testServiceAccountRRN)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				i := &iamv1.SetIamPolicyRequest{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				exp := &iamv1.Policy{}
				if !serviceaccountpolicy.ArePoliciesSame(exp, i.Policy) {
					t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(exp, i.Policy, cmpopts.IgnoreFields(iamv1.Policy{}, "Version")))
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(exp)
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
			},
		},
		"CreateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&iamv1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
			},
			want: want{
				mg: ServiceAccountPolicy(
					sapWithName(sapMetadataName),
					sapWithExternalNameAnnotation(sapMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceaccounts := iamv1.NewProjectsServiceAccountsService(s)
			e := &serviceAccountPolicyExternal{serviceaccountspolicy: serviceaccounts}
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
