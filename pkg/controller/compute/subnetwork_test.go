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

package compute

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis/compute/v1beta1"
	gcpv1alpha3 "github.com/crossplane/provider-gcp/apis/v1alpha3"
	"github.com/crossplane/provider-gcp/pkg/clients/subnetwork"
)

const (
	testSubnetworkName = "test-subnetwork"
)

var _ managed.ExternalConnecter = &subnetworkConnector{}
var _ managed.ExternalClient = &subnetworkExternal{}

type subnetworkModifier func(*v1beta1.Subnetwork)

func subnetworkWithConditions(c ...runtimev1alpha1.Condition) subnetworkModifier {
	return func(i *v1beta1.Subnetwork) { i.Status.SetConditions(c...) }
}

func subnetworkWithDescription(d string) subnetworkModifier {
	return func(i *v1beta1.Subnetwork) { i.Spec.ForProvider.Description = &d }
}

func subnetworkWithPrivateAccess(p bool) subnetworkModifier {
	return func(i *v1beta1.Subnetwork) { i.Spec.ForProvider.PrivateIPGoogleAccess = &p }
}

func subnetworkObj(im ...subnetworkModifier) *v1beta1.Subnetwork {
	i := &v1beta1.Subnetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testSubnetworkName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.ExternalNameAnnotationKey: testSubnetworkName,
			},
		},
		Spec: v1beta1.SubnetworkSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
			ForProvider: v1beta1.SubnetworkParameters{},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

func TestSubnetworkConnect(t *testing.T) {
	provider := gcpv1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: gcpv1alpha3.ProviderSpec{
			ProjectID: projectID,
			ProviderSpec: runtimev1alpha1.ProviderSpec{
				CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
					SecretReference: runtimev1alpha1.SecretReference{
						Namespace: namespace,
						Name:      providerSecretName,
					},
					Key: providerSecretKey,
				},
			},
		},
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte("super-secret")},
	}

	type args struct {
		mg resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		conn managed.ExternalConnecter
		args args
		want want
	}{
		"NotSubnetwork": {
			conn: &subnetworkConnector{},
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				err: errors.New(errNotSubnetwork),
			},
		},
		"Connected": {
			conn: &subnetworkConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newServiceFn: func(ctx context.Context, opts ...option.ClientOption) (*compute.Service, error) {
					return &compute.Service{}, nil
				},
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				err: nil,
			},
		},
		"FailedToGetProvider": {
			conn: &subnetworkConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					return errBoom
				}},
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				err: errors.Wrap(errBoom, errProviderNotRetrieved),
			},
		},
		"ProviderSecretNil": {
			conn: &subnetworkConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						nilSecretProvider := provider
						nilSecretProvider.SetCredentialsSecretReference(nil)
						*obj.(*gcpv1alpha3.Provider) = nilSecretProvider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return errBoom
					}
					return nil
				}},
			},
			args: args{mg: subnetworkObj()},
			want: want{err: errors.New(errProviderSecretNil)},
		},
		"FailedToGetProviderSecret": {
			conn: &subnetworkConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return errBoom
					}
					return nil
				}},
			},
			args: args{mg: subnetworkObj()},
			want: want{err: errors.Wrap(errBoom, errProviderSecretNotRetrieved)},
		},
		"FailedToCreateComputeClient": {
			conn: &subnetworkConnector{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*gcpv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newServiceFn: func(_ context.Context, _ ...option.ClientOption) (*compute.Service, error) { return nil, errBoom },
			},
			args: args{mg: subnetworkObj()},
			want: want{err: errors.Wrap(errBoom, errNewClient)},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.conn.Connect(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}
		})
	}
}

func TestSubnetworkObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"NotFound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(),
				err: nil,
			},
		},
		"GetFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetSubnetwork),
			},
		},
		"NotUpToDateSpecUpdateFailed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				c := subnetworkObj()
				gn := &compute.Subnetwork{}
				subnetwork.GenerateSubnetwork(testSubnetworkName, c.Spec.ForProvider, gn)
				gn.Description = "a very interesting description"
				_ = json.NewEncoder(w).Encode(gn)
			}),
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(errBoom),
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithDescription("a very interesting description")),
				err: errors.Wrap(errBoom, errManagedSubnetworkUpdate),
			},
		},
		"RunnableUnbound": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				c := &compute.Subnetwork{}
				subnetwork.GenerateSubnetwork(testSubnetworkName, subnetworkObj().Spec.ForProvider, c)
				_ = json.NewEncoder(w).Encode(c)
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: subnetworkObj(subnetworkWithConditions(runtimev1alpha1.Available())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSubnetworkCreate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg  resource.Managed
		cre managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				i := &compute.Subnetwork{}
				b, err := ioutil.ReadAll(r.Body)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				err = json.Unmarshal(b, i)
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = r.Body.Close()
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(runtimev1alpha1.Creating())),
				cre: managed.ExternalCreation{},
				err: nil,
			},
		},
		"AlreadyExists": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(gError(http.StatusConflict, ""), errCreateSubnetworkFailed),
			},
		},
		"Failed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errCreateSubnetworkFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			_, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSubnetworkDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(runtimev1alpha1.Deleting())),
				err: nil,
			},
		},
		"AlreadyGone": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(runtimev1alpha1.Deleting())),
				err: nil,
			},
		},
		"Failed": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errDeleteSubnetworkFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			err := e.Delete(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSubnetworkUpdate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		upd managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		handler http.Handler
		kube    client.Client
		args    args
		want    want
	}{
		"NotSubnetwork": {
			handler: nil,
			args: args{
				mg: &v1beta1.Network{},
			},
			want: want{
				mg:  &v1beta1.Network{},
				err: errors.New(errNotSubnetwork),
			},
		},
		"GetExternalFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(),
			},
			want: want{
				mg:  subnetworkObj(),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errGetSubnetwork),
			},
		},
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{
						Description: "not the one I want",
					})
				case http.MethodPatch:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(subnetworkWithDescription("a new description")),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithDescription("a new description")),
				err: nil,
			},
		},
		"SuccessfulPrivateAccess": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{
						PrivateIpGoogleAccess: false,
					})
				case http.MethodPost:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(subnetworkWithPrivateAccess(true)),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithPrivateAccess(true)),
				err: nil,
			},
		},
		"UpdateGeneralFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{})
				case http.MethodPatch:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				// Must include field that causes update.
				mg: subnetworkObj(subnetworkWithDescription("a new description")),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithDescription("a new description")),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateSubnetworkFailed),
			},
		},
		"UpdatePrivateAccessFails": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.Body.Close()
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&compute.Subnetwork{
						PrivateIpGoogleAccess: false,
					})
				case http.MethodPost:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				default:
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(&compute.Operation{})
				}
			}),
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(nil),
			},
			args: args{
				mg: subnetworkObj(subnetworkWithPrivateAccess(true)),
			},
			want: want{
				mg:  subnetworkObj(subnetworkWithPrivateAccess(true)),
				err: errors.Wrap(gError(http.StatusBadRequest, ""), errUpdateSubnetworkPAFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			s, _ := compute.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
			e := subnetworkExternal{
				kube:      tc.kube,
				projectID: projectID,
				Service:   s,
			}
			upd, err := e.Update(context.Background(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}

		})
	}
}
