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
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	. "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-gcp/apis"
	. "github.com/crossplane/provider-gcp/apis/compute/v1alpha3"
	"github.com/crossplane/provider-gcp/pkg/clients/fake"
	"github.com/crossplane/provider-gcp/pkg/clients/gke"
)

const (
	namespace    = "default"
	providerName = "test-provider"
	clusterName  = "test-cluster"
)

var (
	key = types.NamespacedName{
		Name: clusterName,
	}
	request = reconcile.Request{
		NamespacedName: key,
	}

	masterAuth = &container.MasterAuth{
		Username:             "test-user",
		Password:             "test-pass",
		ClusterCaCertificate: base64.StdEncoding.EncodeToString([]byte("test-ca")),
		ClientCertificate:    base64.StdEncoding.EncodeToString([]byte("test-cert")),
		ClientKey:            base64.StdEncoding.EncodeToString([]byte("test-key")),
	}
)

func init() {
	_ = apis.AddToScheme(scheme.Scheme)
}

func testCluster() *GKECluster {
	return &GKECluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: GKEClusterSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &runtimev1alpha1.Reference{Name: providerName},
			},
		},
	}
}

// assertResource a helper function to check on cluster and its status
func assertResource(g *GomegaWithT, r *Reconciler, s runtimev1alpha1.ConditionedStatus) *GKECluster {
	rc := &GKECluster{}
	err := r.Get(ctx, key, rc)
	g.Expect(err).To(BeNil())
	g.Expect(cmp.Diff(s, rc.Status.ConditionedStatus, test.EquateConditions())).Should(BeZero())
	return rc
}

type mockReferenceResolver struct{}

func (*mockReferenceResolver) ResolveReferences(ctx context.Context, mg resource.Managed) (err error) {
	return nil
}

func TestSyncErroredCluster(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	r := &Reconciler{
		Client: NewFakeClient(tc),
		log:    logging.NewNopLogger(),
	}

	errorMessage := "Something went wrong on gcloud side."
	cluster := &container.Cluster{
		Status:        ClusterStateError,
		StatusMessage: errorMessage,
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedState := ClusterStateError
	expectedStatus.SetConditions(runtimev1alpha1.Unavailable().
		WithMessage(fmt.Sprintf(erroredClusterErrorMessageFormat, ClusterStateError, errorMessage)))

	rs, err := r._sync(tc, cluster)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(tc.Status.State).To(Equal(expectedState))
	assertResource(g, r, expectedStatus)
}

func TestSyncPublishConnectionDetailsError(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	testError := errors.New("publish-connection-details-error")
	r := &Reconciler{
		Client: NewFakeClient(tc),
		publisher: managed.ConnectionPublisherFns{
			PublishConnectionFn: func(_ context.Context, _ resource.Managed, _ managed.ConnectionDetails) error {
				return testError
			},
		},
		log: logging.NewNopLogger(),
	}

	auth := masterAuth
	endpoint := "test-ep"

	cluster := &container.Cluster{
		Status:     ClusterStateRunning,
		Endpoint:   endpoint,
		MasterAuth: auth,
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))

	rs, err := r._sync(tc, cluster)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	assertResource(g, r, expectedStatus)
}

func TestSyncClusterNotReady(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	r := &Reconciler{
		Client: NewFakeClient(tc),
		log:    logging.NewNopLogger(),
	}

	cluster := &container.Cluster{
		Status: ClusterStateProvisioning,
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}

	rs, err := r._sync(tc, cluster)
	g.Expect(rs).To(Equal(reconcile.Result{RequeueAfter: requeueOnWait}))
	g.Expect(err).NotTo(HaveOccurred())
	assertResource(g, r, expectedStatus)
}

func TestSync(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	r := &Reconciler{
		Client:    NewFakeClient(tc),
		publisher: managed.PublisherChain{},
		log:       logging.NewNopLogger(),
	}
	auth := masterAuth
	endpoint := "test-ep"

	cluster := &container.Cluster{
		Status:     ClusterStateRunning,
		Endpoint:   endpoint,
		MasterAuth: auth,
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())

	rs, err := r._sync(tc, cluster)
	g.Expect(rs).To(Equal(reconcile.Result{RequeueAfter: requeueOnSucces}))
	g.Expect(err).NotTo(HaveOccurred())
	assertResource(g, r, expectedStatus)
}

func TestDeleteReclaimDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()
	tc.Finalizers = []string{finalizer}
	tc.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimDelete

	r := &Reconciler{
		Client:   NewFakeClient(tc),
		resolver: &mockReferenceResolver{},
		log:      logging.NewNopLogger(),
	}

	called := false
	cl := fake.NewGKEClient()
	cl.MockDeleteCluster = func(string, string) error {
		called = true
		return nil
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileSuccess())

	rs, err := r._delete(tc, cl)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)
}

func TestDeleteReclaimRetain(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()
	tc.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimRetain
	tc.Finalizers = []string{finalizer}

	r := &Reconciler{
		Client:   NewFakeClient(tc),
		resolver: &mockReferenceResolver{},
		log:      logging.NewNopLogger(),
	}

	called := false
	cl := fake.NewGKEClient()
	cl.MockDeleteCluster = func(string, string) error {
		called = true
		return nil
	}

	rs, err := r._delete(tc, cl)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).To(BeNil())
	// there should be no delete calls on gke client since policy is set to Retain
	g.Expect(called).To(BeFalse())

	// expected to have all conditions set to inactive
	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileSuccess())

	assertResource(g, r, expectedStatus)
}

func TestDeleteFailed(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()
	tc.Spec.ReclaimPolicy = runtimev1alpha1.ReclaimDelete
	tc.Finalizers = []string{finalizer}

	r := &Reconciler{
		Client: NewFakeClient(tc),
		log:    logging.NewNopLogger(),
	}

	testError := errors.New("test-delete-error")

	called := false
	cl := fake.NewGKEClient()
	cl.MockDeleteCluster = func(string, string) error {
		called = true
		return testError
	}

	rs, err := r._delete(tc, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).To(BeNil())
	// there should be no delete calls on gke client since policy is set to Retain
	g.Expect(called).To(BeTrue())

	// expected status
	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileError(testError))

	assertResource(g, r, expectedStatus)
}

func TestReconcileObjectNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &Reconciler{
		Client: NewFakeClient(),
		log:    logging.NewNopLogger(),
	}
	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).To(BeNil())
}

func TestReconcileClientError(t *testing.T) {
	g := NewGomegaWithT(t)

	testError := errors.New("test-client-error")

	called := false
	kube := NewFakeClient(testCluster())
	r := &Reconciler{
		Client: kube,
		connect: func(*GKECluster) (gke.Client, error) {
			called = true
			return nil, testError
		},
		log:         logging.NewNopLogger(),
		initializer: managed.NewNameAsExternalName(kube),
	}

	// expected to have a failed condition
	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())

	assertResource(g, r, expectedStatus)
}

func TestReconcileDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	// test objects
	tc := testCluster()
	dt := metav1.Now()
	tc.DeletionTimestamp = &dt

	called := false
	kube := NewFakeClient(tc)
	r := &Reconciler{
		Client: kube,
		connect: func(*GKECluster) (gke.Client, error) {
			return nil, nil
		},
		delete: func(*GKECluster, gke.Client) (reconcile.Result, error) {
			called = true
			return result, nil
		},
		resolver:    &mockReferenceResolver{},
		log:         logging.NewNopLogger(),
		initializer: managed.NewNameAsExternalName(kube),
	}

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, runtimev1alpha1.ConditionedStatus{})
}

func TestReconcileCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	called := false
	kube := NewFakeClient(testCluster())
	r := &Reconciler{
		Client: kube,
		connect: func(*GKECluster) (gke.Client, error) {
			return &fake.GKEClient{MockGetCluster: func(_ string, _ string) (*container.Cluster, error) {
				return nil, &googleapi.Error{Code: http.StatusNotFound}
			}}, nil
		},
		create: func(*GKECluster, gke.Client) (reconcile.Result, error) {
			called = true
			return resultRequeue, nil
		},
		resolver:    &mockReferenceResolver{},
		log:         logging.NewNopLogger(),
		initializer: managed.NewNameAsExternalName(kube),
	}

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())
}

func TestReconcileSync(t *testing.T) {
	g := NewGomegaWithT(t)

	called := false

	tc := testCluster()
	tc.Finalizers = []string{finalizer}
	kube := NewFakeClient(tc)
	r := &Reconciler{
		Client: kube,
		connect: func(*GKECluster) (gke.Client, error) {
			return &fake.GKEClient{MockGetCluster: func(_ string, _ string) (*container.Cluster, error) {
				return &container.Cluster{}, nil
			}}, nil
		},
		sync: func(*GKECluster, *container.Cluster) (reconcile.Result, error) {
			called = true
			return resultRequeue, nil
		},
		resolver:    &mockReferenceResolver{},
		log:         logging.NewNopLogger(),
		initializer: managed.NewNameAsExternalName(kube),
	}

	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())

	rc := assertResource(g, r, runtimev1alpha1.ConditionedStatus{})
	g.Expect(rc.Finalizers).To(HaveLen(1))
	g.Expect(rc.Finalizers).To(ContainElement(finalizer))
}

func TestConnectionDetails(t *testing.T) {
	g := NewGomegaWithT(t)

	endpoint := "endpoint"
	username := "username"
	password := "password"
	clusterCA := "clusterCA"
	clientCert := "clientCert"
	clientKey := "clientKey"

	cluster := &container.Cluster{
		Endpoint: endpoint,
		MasterAuth: &container.MasterAuth{
			Username:             username,
			Password:             password,
			ClusterCaCertificate: base64.StdEncoding.EncodeToString([]byte(clusterCA)),
			ClientCertificate:    base64.StdEncoding.EncodeToString([]byte(clientCert)),
			ClientKey:            base64.StdEncoding.EncodeToString([]byte(clientKey)),
		},
	}
	config, _ := gke.GenerateClientConfig(cluster)
	kubeconfig, _ := clientcmd.Write(config)

	want := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey:   []byte(fmt.Sprintf("https://%s", endpoint)),
		runtimev1alpha1.ResourceCredentialsSecretUserKey:       []byte(username),
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey:   []byte(password),
		runtimev1alpha1.ResourceCredentialsSecretCAKey:         []byte(clusterCA),
		runtimev1alpha1.ResourceCredentialsSecretClientCertKey: []byte(clientCert),
		runtimev1alpha1.ResourceCredentialsSecretClientKeyKey:  []byte(clientKey),
		runtimev1alpha1.ResourceCredentialsSecretKubeconfigKey: kubeconfig,
	}
	got, err := connectionDetails(cluster)
	g.Expect(got).To(Equal(want))
	g.Expect(err).To(BeNil())
}

func TestConnectionDetailsError(t *testing.T) {
	g := NewGomegaWithT(t)

	endpoint := "endpoint"
	username := "username"
	password := "password"
	clusterCA := "clusterCA"
	clientCert := "clientCert"

	got, err := connectionDetails(&container.Cluster{
		Endpoint: endpoint,
		MasterAuth: &container.MasterAuth{
			Username:             username,
			Password:             password,
			ClusterCaCertificate: base64.StdEncoding.EncodeToString([]byte(clusterCA)),
			ClientCertificate:    base64.StdEncoding.EncodeToString([]byte(clientCert)),

			// Just testing the one key since it's the same error case for all
			// of them.
			ClientKey: "probably-not-base64",
		},
	})
	g.Expect(got).To(BeNil())
	g.Expect(err).To(HaveOccurred())
}
