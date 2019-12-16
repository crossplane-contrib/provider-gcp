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
	"testing"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"google.golang.org/api/container/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	. "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-gcp/apis"
	. "github.com/crossplaneio/stack-gcp/apis/compute/v1alpha3"
	"github.com/crossplaneio/stack-gcp/pkg/clients/fake"
	"github.com/crossplaneio/stack-gcp/pkg/clients/gke"
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
				ProviderReference: &corev1.ObjectReference{Name: providerName},
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

func (*mockReferenceResolver) ResolveReferences(ctx context.Context, res resource.CanReference) (err error) {
	return nil
}

func TestSyncClusterGetError(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	r := &Reconciler{
		Client:   NewFakeClient(tc),
		resolver: &mockReferenceResolver{},
	}

	called := false
	testError := errors.New("test-cluster-retriever-error")

	cl := fake.NewGKEClient()
	cl.MockGetCluster = func(string, string) (*container.Cluster, error) {
		called = true
		return nil, testError
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))

	rs, err := r._sync(tc, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)
}

func TestSyncErroredCluster(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	r := &Reconciler{
		Client: NewFakeClient(tc),
	}

	errorMessage := "Something went wrong on gcloud side."

	called := false

	cl := fake.NewGKEClient()
	cl.MockGetCluster = func(string, string) (*container.Cluster, error) {
		called = true
		return &container.Cluster{
			Status:        ClusterStateError,
			StatusMessage: errorMessage,
		}, nil
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedState := ClusterStateError
	expectedStatus.SetConditions(runtimev1alpha1.Unavailable().
		WithMessage(fmt.Sprintf(erroredClusterErrorMessageFormat, ClusterStateError, errorMessage)))

	rs, err := r._sync(tc, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	g.Expect(tc.Status.State).To(Equal(expectedState))
	assertResource(g, r, expectedStatus)
}

func TestSyncPublishConnectionDetailsError(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	testError := errors.New("publish-connection-details-error")
	r := &Reconciler{
		Client: NewFakeClient(tc),
		publisher: resource.ManagedConnectionPublisherFns{
			PublishConnectionFn: func(_ context.Context, _ resource.Managed, _ resource.ConnectionDetails) error {
				return testError
			},
		},
	}

	called := false

	auth := masterAuth
	endpoint := "test-ep"

	cl := fake.NewGKEClient()
	cl.MockGetCluster = func(string, string) (*container.Cluster, error) {
		called = true
		return &container.Cluster{
			Status:     ClusterStateRunning,
			Endpoint:   endpoint,
			MasterAuth: auth,
		}, nil
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.ReconcileError(testError))

	rs, err := r._sync(tc, cl)
	g.Expect(rs).To(Equal(resultRequeue))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)
}

func TestSyncClusterNotReady(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	r := &Reconciler{
		Client: NewFakeClient(tc),
	}

	called := false

	cl := fake.NewGKEClient()
	cl.MockGetCluster = func(string, string) (*container.Cluster, error) {
		called = true
		return &container.Cluster{
			Status: ClusterStateProvisioning,
		}, nil
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}

	rs, err := r._sync(tc, cl)
	g.Expect(rs).To(Equal(reconcile.Result{RequeueAfter: requeueOnWait}))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
	assertResource(g, r, expectedStatus)
}

func TestSync(t *testing.T) {
	g := NewGomegaWithT(t)

	tc := testCluster()

	r := &Reconciler{
		Client:    NewFakeClient(tc),
		publisher: resource.PublisherChain{},
	}

	called := false

	auth := masterAuth
	endpoint := "test-ep"

	cl := fake.NewGKEClient()
	cl.MockGetCluster = func(string, string) (*container.Cluster, error) {
		called = true
		return &container.Cluster{
			Status:     ClusterStateRunning,
			Endpoint:   endpoint,
			MasterAuth: auth,
		}, nil
	}

	expectedStatus := runtimev1alpha1.ConditionedStatus{}
	expectedStatus.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())

	rs, err := r._sync(tc, cl)
	g.Expect(rs).To(Equal(reconcile.Result{RequeueAfter: requeueOnSucces}))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(called).To(BeTrue())
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
	}
	rs, err := r.Reconcile(request)
	g.Expect(rs).To(Equal(result))
	g.Expect(err).To(BeNil())
}

func TestReconcileClientError(t *testing.T) {
	g := NewGomegaWithT(t)

	testError := errors.New("test-client-error")

	called := false

	r := &Reconciler{
		Client: NewFakeClient(testCluster()),
		connect: func(*GKECluster) (gke.Client, error) {
			called = true
			return nil, testError
		},
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

	r := &Reconciler{
		Client: NewFakeClient(tc),
		connect: func(*GKECluster) (gke.Client, error) {
			return nil, nil
		},
		delete: func(*GKECluster, gke.Client) (reconcile.Result, error) {
			called = true
			return result, nil
		},
		resolver: &mockReferenceResolver{},
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

	r := &Reconciler{
		Client: NewFakeClient(testCluster()),
		connect: func(*GKECluster) (gke.Client, error) {
			return nil, nil
		},
		create: func(*GKECluster, gke.Client) (reconcile.Result, error) {
			called = true
			return resultRequeue, nil
		},
		resolver: &mockReferenceResolver{},
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
	tc.Status.ClusterName = "test-status- cluster-name"
	tc.Finalizers = []string{finalizer}

	r := &Reconciler{
		Client: NewFakeClient(tc),
		connect: func(*GKECluster) (gke.Client, error) {
			return nil, nil
		},
		sync: func(*GKECluster, gke.Client) (reconcile.Result, error) {
			called = true
			return resultRequeue, nil
		},
		resolver: &mockReferenceResolver{},
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

	want := resource.ConnectionDetails{
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
