package storage

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
	"google.golang.org/api/option"
	storagev1 "google.golang.org/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	iamv1alpha1 "github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
	"github.com/crossplane/provider-gcp/apis/storage/v1alpha1"
	"github.com/crossplane/provider-gcp/pkg/clients/bucketpolicy"
)

const (
	bpMetadataName  = "test-BucketPolicy"
	keyExternalName = "crossplane.io/external-name"
)

var (
	testBucketName = "my-bucket"

	testMember = "serviceAccount:perfect-test-sa@my-project.iam.gserviceaccount.com"
	testRole   = "roles/crossplane.unitTester"
)

type strange struct {
	resource.Managed
}

func gError(code int, message string) error {
	return googleapi.CheckResponse(&http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(strings.NewReader(message)),
	})
}

type bpValueModifier func(ring *v1alpha1.BucketPolicy)

func bpWithName(s string) bpValueModifier {
	return func(i *v1alpha1.BucketPolicy) { i.Name = s }
}

func bpWithExternalNameAnnotation(externalName string) bpValueModifier {
	return func(i *v1alpha1.BucketPolicy) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[keyExternalName] = externalName
	}
}

func bpWithCondition(condition xpv1.Condition) bpValueModifier {
	return func(i *v1alpha1.BucketPolicy) { i.SetConditions(condition) }
}

func bpWithBinding(binding *iamv1alpha1.Binding) bpValueModifier {
	return func(i *v1alpha1.BucketPolicy) {
		i.Spec.ForProvider.Policy.Bindings = append(i.Spec.ForProvider.Policy.Bindings, binding)
	}
}

func BucketPolicy(im ...bpValueModifier) *v1alpha1.BucketPolicy {
	bp := &v1alpha1.BucketPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       bpMetadataName,
			Finalizers: []string{},
		},
		Spec: v1alpha1.BucketPolicySpec{
			ForProvider: v1alpha1.BucketPolicyParameters{
				Bucket: &testBucketName,
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
		m(bp)
	}

	return bp
}

func TestBucketPolicyObserve(t *testing.T) {
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
		"NotBucketPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotBucketPolicy),
			},
		},
		"FailedToObserve": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
				),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName)),
				observation: managed.ExternalObservation{},
				err:         errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"ObservedPolicyEmpty": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				bp := &storagev1.Policy{}
				_ = json.NewEncoder(w).Encode(bp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
				),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName)),
				observation: managed.ExternalObservation{},
			},
		},
		"ObservedPolicyNeedsUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bp := &storagev1.Policy{
					Bindings: []*storagev1.PolicyBindings{
						{
							Members: []string{"some-other-member"},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(bp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
				),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName)),
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
				// https://cloud.google.com/storage/docs/json_api/v1/buckets/getIamPolicy
				expectedEp := fmt.Sprintf("/b/%s/iam", testBucketName)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				bp := &storagev1.Policy{
					Bindings: []*storagev1.PolicyBindings{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(bp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
				),
			},
			want: want{
				mg: BucketPolicy(
					bpWithCondition(xpv1.Available()),
					bpWithName(bpMetadataName)),
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
			s, _ := storagev1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			buckets := storagev1.NewBucketsService(s)
			e := &bucketPolicyExternal{bucketpolicy: buckets}
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

func TestBucketPolicyCreate(t *testing.T) {
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
		"NotBucketPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotBucketPolicy),
			},
		},
		"CreateSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/storage/docs/json_api/v1/buckets/setIamPolicy
				expectedEp := fmt.Sprintf("/b/%s/iam", testBucketName)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				i := &storagev1.Policy{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				exp := &storagev1.Policy{
					Bindings: []*storagev1.PolicyBindings{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				if !bucketpolicy.ArePoliciesSame(exp, i) {
					t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(exp, i, cmpopts.IgnoreFields(storagev1.Policy{}, "Version")))
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(exp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Creating())),
			},
		},
		"CreateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Creating())),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := storagev1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			buckets := storagev1.NewBucketsService(s)
			e := &bucketPolicyExternal{bucketpolicy: buckets}

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

func TestBucketPolicyUpdate(t *testing.T) {
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
		"NotBucketPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotBucketPolicy),
			},
		},
		"UpdateSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var bp *storagev1.Policy
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					bp = &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
				case http.MethodPut:
					i := &storagev1.Policy{}
					b, err := ioutil.ReadAll(r.Body)
					if diff := cmp.Diff(err, nil); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					err = json.Unmarshal(b, i)
					if diff := cmp.Diff(err, nil); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					bp = &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
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
					if !bucketpolicy.ArePoliciesSame(bp, i) {
						t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(bp, i, cmpopts.IgnoreFields(storagev1.Policy{}, "Version")))
					}
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(bp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Available()),
					bpWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Available()),
					bpWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
		},
		"FailedToGet": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"AlreadyUpToDate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bp := &storagev1.Policy{
					Bindings: []*storagev1.PolicyBindings{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(bp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Available())),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Available())),
			},
		},
		"FailedToUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var bp *storagev1.Policy
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					bp = &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
				case http.MethodPut:
					bp = &storagev1.Policy{}
					w.WriteHeader(http.StatusInternalServerError)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(bp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Available()),
					bpWithBinding(&iamv1alpha1.Binding{
						Members: []string{"another-member"},
						Role:    "another-role",
					})),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName),
					bpWithCondition(xpv1.Available()),
					bpWithBinding(&iamv1alpha1.Binding{
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
			s, _ := storagev1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			buckets := storagev1.NewBucketsService(s)
			e := &bucketPolicyExternal{bucketpolicy: buckets}
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

func TestBucketPolicyDelete(t *testing.T) {
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
		"NotBucketPolicy": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotBucketPolicy),
			},
		},
		"DeleteSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				// https://cloud.google.com/storage/docs/json_api/v1/buckets/setIamPolicy
				expectedEp := fmt.Sprintf("/b/%s/iam", testBucketName)
				if !strings.EqualFold(r.URL.Path, expectedEp) {
					t.Errorf("requested URL.Path to get policy should end with: %s, got %s instead",
						expectedEp, r.URL.Path)
				}
				i := &storagev1.Policy{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				exp := &storagev1.Policy{}
				if !bucketpolicy.ArePoliciesSame(exp, i) {
					t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(exp, i, cmpopts.IgnoreFields(storagev1.Policy{}, "Version")))
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(exp)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
			},
		},
		"CreateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
			},
			want: want{
				mg: BucketPolicy(
					bpWithName(bpMetadataName),
					bpWithExternalNameAnnotation(bpMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := storagev1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			buckets := storagev1.NewBucketsService(s)
			e := &bucketPolicyExternal{bucketpolicy: buckets}
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
