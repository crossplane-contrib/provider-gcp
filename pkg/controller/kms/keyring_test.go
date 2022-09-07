package kms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

	"github.com/crossplane-contrib/provider-gcp/apis/kms/v1alpha1"
)

const (
	namespace = "some-namespace"
	project   = "someProject"
	location  = "test-location"

	connectionSecretName = "some-connection-secret"
	metadataName         = "test-keyring"
	keyExternalName      = "crossplane.io/external-name"
)

var (
	err500 = &googleapi.Error{Code: 500, Body: "{}\n"}
	fqName = fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, metadataName)
)

type strange struct {
	resource.Managed
}

type valueModifier func(ring *v1alpha1.KeyRing)

func withName(s string) valueModifier {
	return func(i *v1alpha1.KeyRing) { i.Name = s }
}

func withAtProviderName(s string) valueModifier {
	return func(i *v1alpha1.KeyRing) { i.Status.AtProvider.Name = s }
}

func withLocation(s string) valueModifier {
	return func(i *v1alpha1.KeyRing) { i.Spec.ForProvider.Location = s }
}

func withExternalNameAnnotation(externalName string) valueModifier {
	return func(i *v1alpha1.KeyRing) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[keyExternalName] = externalName
	}
}

func withCondition(condition xpv1.Condition) valueModifier {
	return func(i *v1alpha1.KeyRing) { i.SetConditions(condition) }
}

func withDeletionTimestamp(ts metav1.Time) valueModifier {
	return func(i *v1alpha1.KeyRing) { i.SetDeletionTimestamp(&ts) }
}

func keyRing(im ...valueModifier) *v1alpha1.KeyRing {
	kr := &v1alpha1.KeyRing{
		ObjectMeta: metav1.ObjectMeta{
			Name:       metadataName,
			Finalizers: []string{},
		},
		Spec: v1alpha1.KeyRingSpec{
			ResourceSpec: xpv1.ResourceSpec{
				WriteConnectionSecretToReference: &xpv1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1alpha1.KeyRingParameters{
				Location: location,
			},
		},
	}

	for _, m := range im {
		m(kr)
	}

	return kr
}

func TestRelativeResourceNamer(t *testing.T) {
	type args struct {
		rrn RelativeResourceNamerKeyRing
		mg  *v1alpha1.KeyRing
	}

	type want struct {
		projectName  string
		location     string
		resourceName string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Empty": {
			args: args{
				rrn: NewRelativeResourceNamerKeyRing("", ""),
				mg:  keyRing(withExternalNameAnnotation("")),
			},
			want: want{
				projectName:  "projects/",
				location:     "projects//locations/",
				resourceName: "projects//locations//keyRings/",
			},
		},
		"PerfectProjectName": {
			args: args{
				rrn: NewRelativeResourceNamerKeyRing("perfect", "home"),
				mg:  keyRing(withExternalNameAnnotation("my-keyring")),
			},
			want: want{
				projectName:  "projects/perfect",
				location:     "projects/perfect/locations/home",
				resourceName: "projects/perfect/locations/home/keyRings/my-keyring",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			projectName := tc.args.rrn.ProjectRRN()
			location := tc.args.rrn.LocationRRN()
			resourceName := tc.args.rrn.ResourceName(tc.args.mg)
			if diff := cmp.Diff(tc.want.projectName, projectName, test.EquateConditions()); diff != "" {
				t.Errorf("RelativeResourceNamerKeyRing.ProjectName(): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.location, location, test.EquateConditions()); diff != "" {
				t.Errorf("RelativeResourceNamerKeyRing.Location(): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.resourceName, resourceName, test.EquateConditions()); diff != "" {
				t.Errorf("RelativeResourceNamerKeyRing.ResourceName(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
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
		"ObservedKeyRingGot": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				sa := &kmsv1.KeyRing{
					Name: fqName,
				}
				if err := json.NewEncoder(w).Encode(sa); err != nil {
					t.Error(err)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: keyRing(
					withName(metadataName),
					withExternalNameAnnotation(metadataName),
				),
			},
			want: want{
				mg: keyRing(
					withName(metadataName),
					withLocation(location),
					withExternalNameAnnotation(metadataName),
					withAtProviderName(fqName),
					withCondition(xpv1.Available())),
				observation: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"ObservedKeyRingGotButCRDeleted": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				sa := &kmsv1.KeyRing{
					Name: fqName,
				}
				if err := json.NewEncoder(w).Encode(sa); err != nil {
					t.Error(err)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: keyRing(
					withName(metadataName),
					withExternalNameAnnotation(metadataName),
					withDeletionTimestamp(now),
				),
			},
			want: want{
				mg: keyRing(
					withName(metadataName),
					withLocation(location),
					withExternalNameAnnotation(metadataName),
					withDeletionTimestamp(now)),
			},
		},
		"ObservedKeyRingDoesNotExist": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg:  keyRing(),
			},
			want: want{
				mg:          keyRing(),
				observation: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NotKeyRing": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotKeyRing),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			keyrings := kmsv1.NewProjectsLocationsKeyRingsService(s)
			rrn := NewRelativeResourceNamerKeyRing(project, location)
			e := &keyRingExternal{keyrings: keyrings, rrn: rrn}
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

func TestCreate(t *testing.T) {
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
		"CreatedKeyRing": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				id := r.URL.Query()["keyRingId"]
				if diff := cmp.Diff(id[0], metadataName); diff != "" {
					t.Errorf("keyRingId: -want, +got:\n%s", diff)
				}
				kr := &kmsv1.KeyRing{
					Name: fqName,
				}
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(kr); err != nil {
					t.Error(err)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: keyRing(
					withName(metadataName),
					withLocation(location),
					withExternalNameAnnotation(metadataName)),
			},
			want: want{
				mg: keyRing(
					withName(metadataName),
					withLocation(location),
					withCondition(xpv1.Creating()),
					withExternalNameAnnotation(metadataName)),
			},
		},
		"NotKeyRing": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotKeyRing),
			},
		},
		"FailedToCreateKeyRing": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(&iamv1.Empty{}); err != nil {
					t.Error(err)
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: keyRing(
					withName(metadataName),
					withLocation(location)),
			},
			want: want{
				mg: keyRing(
					withName(metadataName),
					withLocation(location),
					withCondition(xpv1.Creating())),
				err: errors.Wrap(err500, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			keyrings := kmsv1.NewProjectsLocationsKeyRingsService(s)
			rrn := NewRelativeResourceNamerKeyRing(project, location)
			e := &keyRingExternal{keyrings: keyrings, rrn: rrn}
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

func TestUpdate(t *testing.T) {
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
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(""), option.WithoutAuthentication())
			keyrings := kmsv1.NewProjectsLocationsKeyRingsService(s)
			rrn := NewRelativeResourceNamerKeyRing(project, location)
			e := &keyRingExternal{keyrings: keyrings, rrn: rrn}
			_, err := e.Update(context.Background(), keyRing())
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Update(...): want error != got error:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
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
			s, _ := kmsv1.NewService(context.Background(), option.WithEndpoint(""), option.WithoutAuthentication())
			keyrings := kmsv1.NewProjectsLocationsKeyRingsService(s)
			rrn := NewRelativeResourceNamerKeyRing(project, location)
			e := &keyRingExternal{keyrings: keyrings, rrn: rrn}
			err := e.Delete(context.Background(), keyRing())
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Update(...): want error != got error:\n%s", diff)
			}
		})
	}
}
