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
	"io/ioutil"
	"os"
	"testing"

	"github.com/crossplaneio/crossplane-runtime/pkg/test/integration"
	crossplaneapis "github.com/crossplaneio/crossplane/apis"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/crossplaneio/stack-gcp/apis"
	cachev1beta1 "github.com/crossplaneio/stack-gcp/apis/cache/v1beta1"
	computev1alpha3 "github.com/crossplaneio/stack-gcp/apis/compute/v1alpha3"
	containerv1beta1 "github.com/crossplaneio/stack-gcp/apis/container/v1beta1"
	databasev1beta1 "github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
	storagev1alpha3 "github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	"github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/crossplaneio/stack-gcp/pkg/controller"
)

func TestCreateAllClasses(t *testing.T) {
	cases := map[string]struct {
		reason string
		test   func(c client.Client) error
	}{
		"CreateProvider": {
			reason: "A GCP Provider should be created without error.",
			test: func(c client.Client) error {
				dat, err := ioutil.ReadFile("../../examples/gcp-provider.yaml")
				if err != nil {
					return err
				}
				p := &v1alpha3.Provider{}
				if err := yaml.Unmarshal(dat, p); err != nil {
					return err
				}

				defer func() {
					if err := c.Delete(context.Background(), p); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), p)
			},
		},
		"CreateBucketClass": {
			reason: "A GCP BucketClass should be created without error.",
			test: func(c client.Client) error {
				dat, err := ioutil.ReadFile("../../examples/storage/bucket/resource-class.yaml")
				if err != nil {
					return err
				}
				b := &storagev1alpha3.BucketClass{}
				if err := yaml.Unmarshal(dat, b); err != nil {
					return err
				}

				defer func() {
					if err := c.Delete(context.Background(), b); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), b)
			},
		},
		"CreatePostgreSQLCloudSQLClass": {
			reason: "A GCP PostgreSQL CloudSQLClass should be created without error.",
			test: func(c client.Client) error {
				dat, err := ioutil.ReadFile("../../examples/database/postgresqlinstance/resource-class.yaml")
				if err != nil {
					return err
				}
				s := &databasev1beta1.CloudSQLInstanceClass{}
				if err := yaml.Unmarshal(dat, s); err != nil {
					return err
				}

				defer func() {
					if err := c.Delete(context.Background(), s); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), s)
			},
		},
		"CreateMySQLCloudSQLClass": {
			reason: "A GCP MySQL CloudSQLClass should be created without error.",
			test: func(c client.Client) error {
				dat, err := ioutil.ReadFile("../../examples/database/mysqlinstance/resource-class.yaml")
				if err != nil {
					return err
				}
				s := &databasev1beta1.CloudSQLInstanceClass{}
				if err := yaml.Unmarshal(dat, s); err != nil {
					return err
				}

				defer func() {
					if err := c.Delete(context.Background(), s); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), s)
			},
		},
		"CreateV1Beta1GKEClusterClass": {
			reason: "A v1beta1 GCP GKEClusterClass should be created without error.",
			test: func(c client.Client) error {
				dat, err := ioutil.ReadFile("../../examples/container/kubernetescluster/resource-class.yaml")
				if err != nil {
					return err
				}
				s := &containerv1beta1.GKEClusterClass{}
				if err := yaml.Unmarshal(dat, s); err != nil {
					return err
				}

				defer func() {
					if err := c.Delete(context.Background(), s); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), s)
			},
		},
		"CreateV1Alpha1GKEClusterClass": {
			reason: "A v1alpha3 GCP GKEClusterClass should be created without error.",
			test: func(c client.Client) error {
				dat, err := ioutil.ReadFile("../../examples/compute/kubernetescluster/resource-class.yaml")
				if err != nil {
					return err
				}
				s := &computev1alpha3.GKEClusterClass{}
				if err := yaml.Unmarshal(dat, s); err != nil {
					return err
				}

				defer func() {
					if err := c.Delete(context.Background(), s); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), s)
			},
		},
		"CreateCloudMemorystoreInstanceClass": {
			reason: "A GCP CloudMemorystoreInstanceClass should be created without error.",
			test: func(c client.Client) error {
				dat, err := ioutil.ReadFile("../../examples/cache/rediscluster/resource-class.yaml")
				if err != nil {
					return err
				}
				s := &cachev1beta1.CloudMemorystoreInstanceClass{}
				if err := yaml.Unmarshal(dat, s); err != nil {
					return err
				}

				defer func() {
					if err := c.Delete(context.Background(), s); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), s)
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

	if err := (&controller.Controllers{}).SetupWithManager(i); err != nil {
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
