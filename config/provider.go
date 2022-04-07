package config

import (
	// Note(ezgidemirel): we are importing this to embed provider schema document
	_ "embed"

	tjconfig "github.com/crossplane/terrajet/pkg/config"
	"github.com/crossplane/terrajet/pkg/types/name"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/crossplane/provider-gcp/config/accessapproval"
	"github.com/crossplane/provider-gcp/config/bigtable"
	"github.com/crossplane/provider-gcp/config/cloudfunctions"
	"github.com/crossplane/provider-gcp/config/cloudiot"
	"github.com/crossplane/provider-gcp/config/cloudplatform"
	"github.com/crossplane/provider-gcp/config/compute"
	"github.com/crossplane/provider-gcp/config/container"
	"github.com/crossplane/provider-gcp/config/dataflow"
	"github.com/crossplane/provider-gcp/config/dataproc"
	"github.com/crossplane/provider-gcp/config/monitoring"
	"github.com/crossplane/provider-gcp/config/project"
	"github.com/crossplane/provider-gcp/config/redis"
	"github.com/crossplane/provider-gcp/config/sql"
	"github.com/crossplane/provider-gcp/config/storage"
)

const (
	resourcePrefix = "gcp"
	modulePath     = "github.com/crossplane/provider-gcp"
)

//go:embed schema.json
var providerSchema string

var skipList = []string{
	// Note(turkenh): Following two resources conflicts their singular versions
	// "google_access_context_manager_access_level" and
	// "google_access_context_manager_service_perimeter". Skipping for now.
	"google_access_context_manager_access_levels$",
	"google_access_context_manager_service_perimeters$",
}

var includeList = []string{
	// Storage
	"google_storage_bucket$",

	// Compute
	"google_compute_network$",
	"google_compute_subnetwork$",
	"google_compute_address$",
	"google_compute_firewall$",
	"google_compute_instance$",
	"google_compute_managed_ssl_certificate$",
	"google_compute_router$",
	"google_compute_router_nat$",

	// Container
	"google_container_cluster",
	"google_container_node_pool",

	// Monitoring
	"google_monitoring_alert_policy",
	"google_monitoring_notification_channel",
	"google_monitoring_uptime_check_config",

	// Memory Store
	"google_redis_instance",

	// CloudPlatform
	"google_folder$",
	"google_project$",
	"google_service_account$",
	"google_service_account_key$",

	// Sql
	"google_sql_.+",
}

// GetProvider returns provider configuration
func GetProvider() *tjconfig.Provider {
	pc := tjconfig.NewProviderWithSchema([]byte(providerSchema), resourcePrefix, modulePath,
		tjconfig.WithDefaultResourceFn(DefaultResource(
			groupOverrides(),
			externalNameConfig(),
		)),
		tjconfig.WithRootGroup("gcp.jet.crossplane.io"),
		tjconfig.WithShortName("gcpjet"),
		// Comment out the following line to generate all resources.
		tjconfig.WithIncludeList(includeList),
		tjconfig.WithSkipList(skipList))

	for _, configure := range []func(provider *tjconfig.Provider){
		accessapproval.Configure,
		bigtable.Configure,
		cloudfunctions.Configure,
		cloudiot.Configure,
		cloudplatform.Configure,
		container.Configure,
		compute.Configure,
		dataflow.Configure,
		dataproc.Configure,
		redis.Configure,
		monitoring.Configure,
		project.Configure,
		storage.Configure,
		sql.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}

// DefaultResource returns a DefaultResourceFn that makes sure the original
// DefaultResource call is made with given options here.
func DefaultResource(opts ...tjconfig.ResourceOption) tjconfig.DefaultResourceFn {
	return func(name string, terraformResource *schema.Resource, orgOpts ...tjconfig.ResourceOption) *tjconfig.Resource {
		return tjconfig.DefaultResource(name, terraformResource, append(orgOpts, opts...)...)
	}
}

func init() {
	// GCP specific acronyms

	// Todo(turkenh): move to Terrajet?
	name.AddAcronym("idp", "IdP")
	name.AddAcronym("oauth", "OAuth")
}
