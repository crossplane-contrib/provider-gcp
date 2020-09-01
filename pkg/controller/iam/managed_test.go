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

package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
)

const (
	namespace = "some-namespace"
	project   = "someProject"

	connectionSecretName = "some-connection-secret"
	metadataName         = "beautiful-serviceAccount"
	accountEmail         = "beautiful-serviceAccount@someProject.iam.gserviceaccount.com"
	wtfConst             = "crossplane.io/external-name"
)

var (
	err500      = &googleapi.Error{Code: 500, Body: "{}\n"}
	displayName = "Beautiful, maybe perfect"
	description = "A perfect description"
	fqName      = fmt.Sprintf("projects/%s/serviceAccounts/%s", project, accountEmail)
	uniqueID    = fqName
)

type strange struct {
	resource.Managed
}

type valueModifier func(*v1alpha1.ServiceAccount)

func withName(s string) valueModifier {
	return func(i *v1alpha1.ServiceAccount) { i.Status.AtProvider.Name = s }
}

func withProjectID(s string) valueModifier {
	return func(i *v1alpha1.ServiceAccount) { i.Status.AtProvider.ProjectID = s }
}

func withDisplayName(s string) valueModifier {
	return func(i *v1alpha1.ServiceAccount) { i.Spec.ForProvider.DisplayName = &s }
}

func withDescription(s string) valueModifier {
	return func(i *v1alpha1.ServiceAccount) { i.Spec.ForProvider.Description = &s }
}

func withUniqueID(s string) valueModifier {
	return func(i *v1alpha1.ServiceAccount) { i.Status.AtProvider.UniqueID = s }
}

func withEmail(s string) valueModifier {
	return func(i *v1alpha1.ServiceAccount) { i.Status.AtProvider.Email = s }
}

func withDisabled(b bool) valueModifier {
	return func(i *v1alpha1.ServiceAccount) { i.Status.AtProvider.Disabled = b }
}

func withExternalNameAnnotation(externalName string) valueModifier {
	return func(i *v1alpha1.ServiceAccount) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[wtfConst] = externalName
	}
}

func serviceAccount(im ...valueModifier) *v1alpha1.ServiceAccount {
	sa := &v1alpha1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:       metadataName,
			Finalizers: []string{},
		},
		Spec: v1alpha1.ServiceAccountSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1alpha1.ServiceAccountParameters{
				DisplayName: &displayName,
			},
		},
	}

	for _, m := range im {
		m(sa)
	}

	return sa
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connecter{}

func TestRelativeResourceNamer(t *testing.T) {
	type args struct {
		rrn RelativeResourceNamer
		mg  *v1alpha1.ServiceAccount
	}

	type want struct {
		projectName  string
		resourceName string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Empty": {
			args: args{
				rrn: NewRelativeResourceNamer(""),
				mg:  serviceAccount(withExternalNameAnnotation("")),
			},
			want: want{
				projectName:  "projects/",
				resourceName: "projects//serviceAccounts/@.iam.gserviceaccount.com",
			},
		},
		"PerfectProjectName": {
			args: args{
				rrn: NewRelativeResourceNamer("perfect"),
				mg:  serviceAccount(withExternalNameAnnotation("my-sa")),
			},
			want: want{
				projectName:  "projects/perfect",
				resourceName: "projects/perfect/serviceAccounts/my-sa@perfect.iam.gserviceaccount.com",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			projectName := tc.args.rrn.ProjectName()
			resourceName := tc.args.rrn.ResourceName(tc.args.mg)
			if diff := cmp.Diff(tc.want.projectName, projectName, test.EquateConditions()); diff != "" {
				t.Errorf("RelativeResourceNamer.ProjectName(): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.resourceName, resourceName, test.EquateConditions()); diff != "" {
				t.Errorf("RelativeResourceNamer.ResourceName(...): -want, +got:\n%s", diff)
			}
		})
	}

}

func TestObserve(t *testing.T) {
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
		"ObservedAccountGot": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				sa := &iamv1.ServiceAccount{
					Name:        fqName,
					UniqueId:    uniqueID,
					Email:       accountEmail,
					DisplayName: displayName,
				}
				_ = json.NewEncoder(w).Encode(sa)
			}),
			args: args{
				ctx: context.Background(),
				mg: serviceAccount(
					withName(fqName),
					withExternalNameAnnotation(fqName),
				),
			},
			want: want{
				mg: serviceAccount(
					withName(fqName),
					withUniqueID(uniqueID),
					withEmail(accountEmail),
					withDisplayName(displayName),
					withExternalNameAnnotation(fqName),
					withDisabled(false)),
				observation: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"ObservedServiceAccountDoesNotExist": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg:  serviceAccount(),
			},
			want: want{
				mg:          serviceAccount(),
				observation: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"NotServiceAccount": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccount),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceAccounts := iamv1.NewProjectsService(s).ServiceAccounts
			rrn := NewRelativeResourceNamer("perfect-project")
			e := &external{serviceAccounts: serviceAccounts, rrn: rrn}
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

	type createSA struct {
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	}
	type createRequest struct {
		AccountID      string   `json:"accountId"`
		ServiceAccount createSA `json:"serviceAccount"`
	}

	cases := map[string]struct {
		handler http.Handler
		args    args
		want    want
	}{
		"CreatedAccount": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				ur := &createRequest{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, ur)
				if err != nil {
					t.Errorf("unexpected json body: %s", b)
				}
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}

				expected := &createRequest{
					AccountID: metadataName,
					ServiceAccount: createSA{
						DisplayName: displayName,
						Description: description,
					},
				}
				if diff := cmp.Diff(ur, expected); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				sa := &iamv1.ServiceAccount{
					Name:        fqName,
					Email:       accountEmail,
					DisplayName: displayName,
					Description: description,
					ProjectId:   project,
					UniqueId:    uniqueID,
					Disabled:    false,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(sa)
			}),
			args: args{
				ctx: context.Background(),
				mg: serviceAccount(
					withProjectID(project),
					withExternalNameAnnotation(metadataName),
					withDisplayName(displayName), withDescription(description)),
			},
			want: want{
				mg: serviceAccount(
					withProjectID(project), withName(fqName),
					withExternalNameAnnotation(metadataName),
					withDisplayName(displayName), withDescription(description),
					withEmail(accountEmail), withUniqueID(uniqueID)),
			},
		},
		"NotServiceAccount": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccount),
			},
		},
		"FailedToCreateAccount": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&iamv1.Empty{})
			}),
			args: args{
				ctx: context.Background(),
				mg: serviceAccount(
					withName(metadataName), withProjectID(project),
					withDisplayName(displayName), withDescription(description)),
			},
			want: want{
				mg: serviceAccount(
					withName(metadataName), withProjectID(project),
					withDisplayName(displayName), withDescription(description)),
				err: errors.Wrap(err500, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceAccounts := iamv1.NewProjectsService(s).ServiceAccounts
			rrn := NewRelativeResourceNamer("perfect-project")
			e := &external{serviceAccounts: serviceAccounts, rrn: rrn}
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
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}
	type updatedSA struct {
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	}
	type updateRequest struct {
		ServiceAccount updatedSA `json:"serviceAccount"`
	}

	var respondWith = func(w http.ResponseWriter, status int, sa *iamv1.ServiceAccount) {
	}

	updatedDisplayName := fmt.Sprintf("updated: %s", displayName)
	cases := map[string]struct {
		handler http.Handler
		args    args
		want    want
	}{
		"UpdatedInstance": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					sa := &iamv1.ServiceAccount{
						Name:        fqName,
						ProjectId:   project,
						UniqueId:    metadataName,
						Email:       accountEmail,
						DisplayName: displayName,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(sa)
				case http.MethodPatch:
					req := &updateRequest{}
					b, err := ioutil.ReadAll(r.Body)
					if diff := cmp.Diff(err, nil); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					err = json.Unmarshal(b, req)
					if err != nil {
						t.Errorf("unexpected json body: %s", b)
						respondWith(w, http.StatusInternalServerError, &iamv1.ServiceAccount{})
					}
					if req.ServiceAccount.DisplayName != updatedDisplayName {
						t.Errorf("unexpected displayName, got=%s want=%s", req.ServiceAccount.DisplayName, updatedDisplayName)
						respondWith(w, http.StatusInternalServerError, &iamv1.ServiceAccount{})
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&iamv1.Empty{})
				}
			}),
			args: args{
				ctx: context.Background(),
				mg: serviceAccount(
					withName(fqName),
					withProjectID(project),
					withUniqueID(metadataName),
					withDisplayName(updatedDisplayName),
					withDescription(description),
					withEmail(accountEmail),
				),
			},
			want: want{
				mg: serviceAccount(
					withName(fqName),
					withProjectID(project),
					withUniqueID(metadataName),
					withDisplayName(updatedDisplayName),
					withDescription(description),
					withEmail(accountEmail),
				),
			},
		},
		"NotServiceAccount": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccount),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceAccounts := iamv1.NewProjectsService(s).ServiceAccounts
			rrn := NewRelativeResourceNamer("perfect-project")
			e := &external{serviceAccounts: serviceAccounts, rrn: rrn}
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

func TestDelete(t *testing.T) {
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
		"DeletedServiceAccount": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&iamv1.Empty{})
			}),
			args: args{
				ctx: context.Background(),
				mg:  serviceAccount(),
			},
			want: want{
				mg: serviceAccount(),
			},
		},
		"NotServiceAccount": {
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotServiceAccount),
			},
		},
		"DeleteServiceAccountNotFound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()
				w.WriteHeader(http.StatusNotFound)
			}),
			args: args{
				ctx: context.Background(),
				mg:  serviceAccount(),
			},
			want: want{
				mg: serviceAccount(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := iamv1.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			serviceAccounts := iamv1.NewProjectsService(s).ServiceAccounts
			rrn := NewRelativeResourceNamer("perfect-project")
			e := &external{serviceAccounts: serviceAccounts, rrn: rrn}
			err := e.Delete(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Delete(...): want error != got error:\n%s", diff)
			}
		})
	}
}
