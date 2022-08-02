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

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-gcp/pkg/controller/cache"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/compute"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/config"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/container"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/database"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/dns"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/iam"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/kms"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/pubsub"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/registry"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/servicenetworking"
	"github.com/crossplane-contrib/provider-gcp/pkg/controller/storage"
)

// Setup creates all GCP controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		cache.SetupCloudMemorystoreInstance,
		compute.SetupGlobalAddress,
		compute.SetupAddress,
		compute.SetupNetwork,
		compute.SetupSubnetwork,
		compute.SetupFirewall,
		compute.SetupRouter,
		container.SetupCluster,
		container.SetupNodePool,
		database.SetupCloudSQLInstance,
		dns.SetupPolicy,
		dns.SetupResourceRecordSet,
		iam.SetupServiceAccount,
		iam.SetupServiceAccountKey,
		iam.SetupServiceAccountPolicy,
		kms.SetupKeyRing,
		kms.SetupCryptoKey,
		kms.SetupCryptoKeyPolicy,
		pubsub.SetupSubscription,
		pubsub.SetupTopic,
		servicenetworking.SetupConnection,
		storage.SetupBucket,
		storage.SetupBucketPolicy,
		storage.SetupBucketPolicyMember,
		registry.SetupContainerRegistry,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return config.Setup(mgr, o)
}
