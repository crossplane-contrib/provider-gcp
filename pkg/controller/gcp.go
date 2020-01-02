/*
Copyright 2020 The Crossplane Authors.

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

	"github.com/crossplaneio/stack-gcp/pkg/controller/cache"
	"github.com/crossplaneio/stack-gcp/pkg/controller/compute"
	"github.com/crossplaneio/stack-gcp/pkg/controller/container"
	"github.com/crossplaneio/stack-gcp/pkg/controller/database"
	"github.com/crossplaneio/stack-gcp/pkg/controller/servicenetworking"
	"github.com/crossplaneio/stack-gcp/pkg/controller/storage"
)

// Controllers passes down config and adds individual controllers to the manager.
type Controllers struct{}

// SetupWithManager adds all GCP controllers to the manager.
func (c *Controllers) SetupWithManager(mgr ctrl.Manager) error {
	// TODO(muvaf): Move this interface and logic to controller-runtime as it's common to all.
	controllers := []interface {
		SetupWithManager(ctrl.Manager) error
	}{
		&cache.CloudMemorystoreInstanceClaimSchedulingController{},
		&cache.CloudMemorystoreInstanceClaimDefaultingController{},
		&cache.CloudMemorystoreInstanceClaimController{},
		&cache.CloudMemorystoreInstanceController{},
		&compute.GlobalAddressController{},
		&compute.GKEClusterClaimSchedulingController{},
		&compute.GKEClusterClaimDefaultingController{},
		&compute.GKEClusterClaimController{},
		&compute.GKEClusterController{},
		&compute.NetworkController{},
		&compute.SubnetworkController{},
		&container.GKEClusterClaimSchedulingController{},
		&container.GKEClusterClaimDefaultingController{},
		&container.GKEClusterClaimController{},
		&container.GKEClusterController{},
		&container.NodePoolController{},
		&database.PostgreSQLInstanceClaimSchedulingController{},
		&database.PostgreSQLInstanceClaimDefaultingController{},
		&database.PostgreSQLInstanceClaimController{},
		&database.MySQLInstanceClaimSchedulingController{},
		&database.MySQLInstanceClaimDefaultingController{},
		&database.MySQLInstanceClaimController{},
		&database.CloudSQLInstanceController{},
		&servicenetworking.ConnectionController{},
		&storage.BucketClaimSchedulingController{},
		&storage.BucketClaimDefaultingController{},
		&storage.BucketClaimController{},
		&storage.BucketController{},
	}
	for _, c := range controllers {
		if err := c.SetupWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}
