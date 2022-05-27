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
	"google.golang.org/api/option"
	storagev1 "google.golang.org/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gcp/apis/storage/v1alpha1"
	"github.com/crossplane-contrib/provider-gcp/pkg/clients/bucketpolicy"
)

const (
	bpmMetadataName = "test-bucket-policy-member"
)

type bpmValueModifier func(ring *v1alpha1.BucketPolicyMember)

func bpmWithName(s string) bpmValueModifier {
	return func(i *v1alpha1.BucketPolicyMember) { i.Name = s }
}

func bpmWithExternalNameAnnotation(externalName string) bpmValueModifier {
	return func(i *v1alpha1.BucketPolicyMember) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[keyExternalName] = externalName
	}
}

func bpmWithCondition(condition xpv1.Condition) bpmValueModifier {
	return func(i *v1alpha1.BucketPolicyMember) { i.SetConditions(condition) }
}

func BucketPolicyMember(im ...bpmValueModifier) *v1alpha1.BucketPolicyMember {
	bpm := &v1alpha1.BucketPolicyMember{
		ObjectMeta: metav1.ObjectMeta{
			Name:       bpmMetadataName,
			Finalizers: []string{},
		},
		Spec: v1alpha1.BucketPolicyMemberSpec{
			ForProvider: v1alpha1.BucketPolicyMemberParameters{
				Bucket: &testBucketName,
				Role:   testRole,
				Member: &testMember,
			},
		},
	}

	for _, m := range im {
		m(bpm)
	}

	return bpm
}

func TestBucketPolicyMemberObserve(t *testing.T) {
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
		"NotBucketPolicyMember": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotBucketPolicyMember),
			},
		},
		"FailedToObserve": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
				),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName)),
				observation: managed.ExternalObservation{},
				err:         errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"ObserveSucceededWhileBucketNotFound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
				),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName)),
				observation: managed.ExternalObservation{},
				err:         nil,
			},
		},
		"ObservedPolicyEmpty": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				bpm := &storagev1.Policy{}
				_ = json.NewEncoder(w).Encode(bpm)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
				),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName)),
				observation: managed.ExternalObservation{},
			},
		},
		"ObservedPolicyNeedsUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bpm := &storagev1.Policy{
					Bindings: []*storagev1.PolicyBindings{
						{
							Members: []string{"some-other-member"},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(bpm)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
				),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName)),
				observation: managed.ExternalObservation{},
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
				bpm := &storagev1.Policy{
					Bindings: []*storagev1.PolicyBindings{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(bpm)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
				),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithCondition(xpv1.Available()),
					bpmWithName(bpmMetadataName)),
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
			e := &bucketPolicyMemberExternal{bucketpolicy: buckets}
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

func TestBucketPolicyMemberUpdate(t *testing.T) {
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
		"NotBucketPolicyMember": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotBucketPolicyMember),
			},
		},
		"UpdateSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var bpm *storagev1.Policy
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					bpm = &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{"another-member"},
								Role:    "another-role",
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
					bpm = &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{"another-member"},
								Role:    "another-role",
							},
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					if !bucketpolicy.ArePoliciesSame(bpm, i) {
						t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(bpm, i, cmpopts.IgnoreFields(storagev1.Policy{}, "Version")))
					}
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(bpm)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName),
					bpmWithCondition(xpv1.Available())),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName),
					bpmWithCondition(xpv1.Available())),
			},
		},
		"FailedToGet": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"AlreadyUpToDate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bpm := &storagev1.Policy{
					Bindings: []*storagev1.PolicyBindings{
						{
							Members: []string{testMember},
							Role:    testRole,
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(bpm)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName),
					bpmWithCondition(xpv1.Available())),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName),
					bpmWithCondition(xpv1.Available())),
			},
		},
		"FailedToUpdate": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var bpm *storagev1.Policy
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					bpm = &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{"another-member"},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
				case http.MethodPut:
					bpm = &storagev1.Policy{}
					w.WriteHeader(http.StatusInternalServerError)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}

				_ = json.NewEncoder(w).Encode(bpm)
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName),
					bpmWithCondition(xpv1.Available())),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName),
					bpmWithCondition(xpv1.Available())),
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
			e := &bucketPolicyMemberExternal{bucketpolicy: buckets}
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

func TestBucketPolicyMemberDelete(t *testing.T) {
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
		"NotBucketPolicyMember": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotBucketPolicyMember),
			},
		},
		"DeleteSucceeded": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					i := &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{testMember, "another-member"},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(i)
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
					exp := &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{"another-member"},
								Role:    testRole,
							},
						},
					}
					if !bucketpolicy.ArePoliciesSame(exp, i) {
						t.Errorf("policy in setIamPolicyRequest not equal to expected, diff: %s", cmp.Diff(exp, i, cmpopts.IgnoreFields(storagev1.Policy{}, "Version")))
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(exp)
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
				}

			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
		},
		"DeleteFailedWhileGetting": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errGetPolicy),
			},
		},
		"DeleteFailedWhileSetting": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					p := &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{testMember},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(p)
				case http.MethodPut:
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
				}

			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
				err: errors.Wrap(gError(http.StatusInternalServerError, "{}\n"), errSetPolicy),
			},
		},
		"AlreadyDeletedMemberNotThere": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					p := &storagev1.Policy{
						Bindings: []*storagev1.PolicyBindings{
							{
								Members: []string{"another-member"},
								Role:    testRole,
							},
						},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(p)
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
				}

			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
		},
		"AlreadyDeletedEmptyPolicy": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					p := &storagev1.Policy{}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(p)
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&storagev1.Policy{})
				}

			}),
			args: args{
				ctx: context.Background(),
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
			want: want{
				mg: BucketPolicyMember(
					bpmWithName(bpmMetadataName),
					bpmWithExternalNameAnnotation(bpmMetadataName)),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := storagev1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			buckets := storagev1.NewBucketsService(s)
			e := &bucketPolicyMemberExternal{bucketpolicy: buckets}
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
