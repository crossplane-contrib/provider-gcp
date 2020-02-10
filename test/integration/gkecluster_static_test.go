// +build integration

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

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/test/integration"
	crossplaneapis "github.com/crossplaneio/crossplane/apis"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/crossplaneio/stack-gcp/apis"
	containerv1beta1 "github.com/crossplaneio/stack-gcp/apis/container/v1beta1"
	"github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/crossplaneio/stack-gcp/pkg/controller"
)

func TestGKEClusterStatic(t *testing.T) {
	cases := map[string]struct {
		reason string
		test   func(c client.Client) error
	}{
		"CreateCluster": {
			reason: "A GKECluster should be statically provisioned successfully.",
			test: func(c client.Client) error {
				ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
				defer cancel()

				g := &containerv1beta1.GKECluster{}
				if err := unmarshalFromFile("../testdata/gkecluster_static/gkecluster.yaml", g); err != nil {
					return err
				}

				p := &v1alpha3.Provider{}
				if err := unmarshalFromFile("../testdata/gkecluster_static/provider.yaml", p); err != nil {
					return err
				}

				defer func() {
					if err := waitFor(context.Background(), 10*time.Second, func() (bool, error) {
						err := c.Delete(context.Background(), g)

						if kerrors.IsNotFound(err) {
							return true, nil
						}

						if err != nil {
							return true, err
						}

						return false, nil
					}); err != nil {
						t.Error(err)
					}

					if err := c.Delete(context.Background(), p); err != nil {
						t.Error(err)
					}
				}()

				if err := c.Create(ctx, p); err != nil {
					return err
				}

				if err := c.Create(ctx, g); err != nil {
					return err
				}

				return waitFor(ctx, 10*time.Second, func() (bool, error) {
					to := &containerv1beta1.GKECluster{}
					if err := c.Get(ctx, types.NamespacedName{Name: g.Name}, to); err != nil {
						return true, err
					}

					if to.Status.AtProvider.Status == containerv1beta1.ClusterStateRunning {
						return true, nil
					}

					return false, nil
				})
			},
		},
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", "../../kubeconfig.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "../../sa.json"); err != nil {
		t.Fatal(err)
	}

	i, err := integration.New(cfg,
		integration.WithCRDPaths("../../config/crd"),
		integration.WithCleaners(
			integration.NewCRDCleaner(),
			integration.NewCRDDirCleaner()),
	)

	if err != nil {
		t.Fatal(err)
	}

	if err := apis.AddToScheme(i.GetScheme()); err != nil {
		t.Fatal(err)
	}

	if err := crossplaneapis.AddToScheme(i.GetScheme()); err != nil {
		t.Fatal(err)
	}

	zl := zap.New(zap.UseDevMode(true))
	log := logging.NewLogrLogger(zl.WithName("stack-gcp-gkecluster_static_test"))
	if err := controller.Setup(i, log); err != nil {
		t.Fatal(err)
	}

	i.Run()

	defer func() {
		if err := i.Cleanup(); err != nil {
			t.Fatal(err)
		}
	}()

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.test(i.GetClient())
			if err != nil {
				t.Error(err)
			}
		})
	}
}
