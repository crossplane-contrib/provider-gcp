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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane/provider-gcp/pkg/controller/cache"
	"github.com/crossplane/provider-gcp/pkg/controller/compute"
	"github.com/crossplane/provider-gcp/pkg/controller/container"
	"github.com/crossplane/provider-gcp/pkg/controller/database"
	"github.com/crossplane/provider-gcp/pkg/controller/servicenetworking"
	"github.com/crossplane/provider-gcp/pkg/controller/storage"
)

// Setup creates all GCP controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger) error{
		cache.SetupCloudMemorystoreInstanceClaimScheduling,
		cache.SetupCloudMemorystoreInstanceClaimDefaulting,
		cache.SetupCloudMemorystoreInstanceClaimBinding,
		cache.SetupCloudMemorystoreInstance,
		compute.SetupGlobalAddress,
		compute.SetupGKEClusterClaimScheduling,
		compute.SetupGKEClusterClaimDefaulting,
		compute.SetupGKEClusterClaimBinding,
		compute.SetupGKEClusterTarget,
		compute.SetupGKECluster,
		compute.SetupNetwork,
		compute.SetupSubnetwork,
		container.SetupGKEClusterClaimScheduling,
		container.SetupGKEClusterClaimDefaulting,
		container.SetupGKEClusterClaimBinding,
		container.SetupGKEClusterTarget,
		container.SetupGKECluster,
		container.SetupNodePool,
		database.SetupPostgreSQLInstanceClaimScheduling,
		database.SetupPostgreSQLInstanceClaimDefaulting,
		database.SetupPostgreSQLInstanceClaimBinding,
		database.SetupMySQLInstanceClaimScheduling,
		database.SetupMySQLInstanceClaimDefaulting,
		database.SetupMySQLInstanceClaimBinding,
		database.SetupCloudSQLInstance,
		servicenetworking.SetupConnection,
		storage.SetupBucketClaimScheduling,
		storage.SetupBucketClaimDefaulting,
		storage.SetupBucketClaimBinding,
		storage.SetupBucket,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}
	return nil
}
