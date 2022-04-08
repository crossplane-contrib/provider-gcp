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

	"github.com/crossplane/provider-gcp/internal/classic/controller/cache"
	compute2 "github.com/crossplane/provider-gcp/internal/classic/controller/compute"
	"github.com/crossplane/provider-gcp/internal/classic/controller/config"
	container2 "github.com/crossplane/provider-gcp/internal/classic/controller/container"
	"github.com/crossplane/provider-gcp/internal/classic/controller/database"
	"github.com/crossplane/provider-gcp/internal/classic/controller/dns"
	iam2 "github.com/crossplane/provider-gcp/internal/classic/controller/iam"
	kms2 "github.com/crossplane/provider-gcp/internal/classic/controller/kms"
	pubsub2 "github.com/crossplane/provider-gcp/internal/classic/controller/pubsub"
	"github.com/crossplane/provider-gcp/internal/classic/controller/registry"
	"github.com/crossplane/provider-gcp/internal/classic/controller/servicenetworking"
	storage2 "github.com/crossplane/provider-gcp/internal/classic/controller/storage"
)

// Setup creates all GCP controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		cache.SetupCloudMemorystoreInstance,
		compute2.SetupGlobalAddress,
		compute2.SetupAddress,
		compute2.SetupNetwork,
		compute2.SetupSubnetwork,
		compute2.SetupFirewall,
		compute2.SetupRouter,
		container2.SetupCluster,
		container2.SetupNodePool,
		database.SetupCloudSQLInstance,
		dns.SetupResourceRecordSet,
		iam2.SetupServiceAccount,
		iam2.SetupServiceAccountKey,
		iam2.SetupServiceAccountPolicy,
		kms2.SetupKeyRing,
		kms2.SetupCryptoKey,
		kms2.SetupCryptoKeyPolicy,
		pubsub2.SetupSubscription,
		pubsub2.SetupTopic,
		servicenetworking.SetupConnection,
		storage2.SetupBucket,
		storage2.SetupBucketPolicy,
		storage2.SetupBucketPolicyMember,
		registry.SetupContainerRegistry,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return config.Setup(mgr, o)
}
