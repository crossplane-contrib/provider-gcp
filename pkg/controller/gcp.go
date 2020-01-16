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

	"github.com/crossplaneio/crossplane-runtime/pkg/logging"

	"github.com/crossplaneio/stack-gcp/pkg/controller/cache"
	"github.com/crossplaneio/stack-gcp/pkg/controller/compute"
	"github.com/crossplaneio/stack-gcp/pkg/controller/container"
	"github.com/crossplaneio/stack-gcp/pkg/controller/database"
	"github.com/crossplaneio/stack-gcp/pkg/controller/servicenetworking"
	"github.com/crossplaneio/stack-gcp/pkg/controller/storage"
)

// Setup creates all GCP controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	controllers := []func(ctrl.Manager, logging.Logger) error{
		cache.SetupCloudMemorystoreInstanceClaimSchedulingController,
		cache.SetupCloudMemorystoreInstanceClaimDefaultingController,
		cache.SetupCloudMemorystoreInstanceClaimController,
		cache.SetupCloudMemorystoreInstanceController,
		compute.SetupGlobalAddressController,
		compute.SetupGKEClusterClaimSchedulingController,
		compute.SetupGKEClusterClaimDefaultingController,
		compute.SetupGKEClusterClaimController,
		compute.SetupGKEClusterController,
		compute.SetupNetworkController,
		compute.SetupSubnetworkController,
		container.SetupGKEClusterClaimSchedulingController,
		container.SetupGKEClusterClaimDefaultingController,
		container.SetupGKEClusterClaimController,
		container.SetupGKEClusterController,
		container.SetupNodePoolController,
		database.SetupPostgreSQLInstanceClaimSchedulingController,
		database.SetupPostgreSQLInstanceClaimDefaultingController,
		database.SetupPostgreSQLInstanceClaimController,
		database.SetupMySQLInstanceClaimSchedulingController,
		database.SetupMySQLInstanceClaimDefaultingController,
		database.SetupMySQLInstanceClaimController,
		database.SetupCloudSQLInstanceController,
		servicenetworking.SetupConnectionController,
		storage.SetupBucketClaimSchedulingController,
		storage.SetupBucketClaimDefaultingController,
		storage.SetupBucketClaimController,
		storage.SetupBucketController,
	}
	for _, fn := range controllers {
		if err := fn(mgr, l); err != nil {
			return err
		}
	}
	return nil
}
