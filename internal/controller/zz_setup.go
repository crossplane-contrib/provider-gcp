/*
Copyright 2021 The Crossplane Authors.

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

	"github.com/crossplane/terrajet/pkg/controller"

	folder "github.com/crossplane/provider-gcp/internal/controller/cloudplatform/folder"
	project "github.com/crossplane/provider-gcp/internal/controller/cloudplatform/project"
	serviceaccount "github.com/crossplane/provider-gcp/internal/controller/cloudplatform/serviceaccount"
	serviceaccountkey "github.com/crossplane/provider-gcp/internal/controller/cloudplatform/serviceaccountkey"
	address "github.com/crossplane/provider-gcp/internal/controller/compute/address"
	firewall "github.com/crossplane/provider-gcp/internal/controller/compute/firewall"
	instance "github.com/crossplane/provider-gcp/internal/controller/compute/instance"
	managedsslcertificate "github.com/crossplane/provider-gcp/internal/controller/compute/managedsslcertificate"
	network "github.com/crossplane/provider-gcp/internal/controller/compute/network"
	router "github.com/crossplane/provider-gcp/internal/controller/compute/router"
	routernat "github.com/crossplane/provider-gcp/internal/controller/compute/routernat"
	subnetwork "github.com/crossplane/provider-gcp/internal/controller/compute/subnetwork"
	cluster "github.com/crossplane/provider-gcp/internal/controller/container/cluster"
	nodepool "github.com/crossplane/provider-gcp/internal/controller/container/nodepool"
	alertpolicy "github.com/crossplane/provider-gcp/internal/controller/monitoring/alertpolicy"
	notificationchannel "github.com/crossplane/provider-gcp/internal/controller/monitoring/notificationchannel"
	uptimecheckconfig "github.com/crossplane/provider-gcp/internal/controller/monitoring/uptimecheckconfig"
	providerconfig "github.com/crossplane/provider-gcp/internal/controller/providerconfig"
	instanceredis "github.com/crossplane/provider-gcp/internal/controller/redis/instance"
	database "github.com/crossplane/provider-gcp/internal/controller/sql/database"
	databaseinstance "github.com/crossplane/provider-gcp/internal/controller/sql/databaseinstance"
	sourcerepresentationinstance "github.com/crossplane/provider-gcp/internal/controller/sql/sourcerepresentationinstance"
	sslcert "github.com/crossplane/provider-gcp/internal/controller/sql/sslcert"
	user "github.com/crossplane/provider-gcp/internal/controller/sql/user"
	bucket "github.com/crossplane/provider-gcp/internal/controller/storage/bucket"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		folder.Setup,
		project.Setup,
		serviceaccount.Setup,
		serviceaccountkey.Setup,
		address.Setup,
		firewall.Setup,
		instance.Setup,
		managedsslcertificate.Setup,
		network.Setup,
		router.Setup,
		routernat.Setup,
		subnetwork.Setup,
		cluster.Setup,
		nodepool.Setup,
		alertpolicy.Setup,
		notificationchannel.Setup,
		uptimecheckconfig.Setup,
		providerconfig.Setup,
		instanceredis.Setup,
		database.Setup,
		databaseinstance.Setup,
		sourcerepresentationinstance.Setup,
		sslcert.Setup,
		user.Setup,
		bucket.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
