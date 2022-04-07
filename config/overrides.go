package config

import (
	"regexp"
	"strings"

	tjconfig "github.com/crossplane/terrajet/pkg/config"
	"github.com/crossplane/terrajet/pkg/types/name"
	"github.com/pkg/errors"
)

// GroupKindCalculator returns the correct group and kind name for given TF
// resource.
type GroupKindCalculator func(resource string) (string, string)

func externalNameConfig() tjconfig.ResourceOption {
	return func(r *tjconfig.Resource) {
		r.ExternalName = tjconfig.IdentifierFromProvider
	}
}

func groupOverrides() tjconfig.ResourceOption {
	return func(r *tjconfig.Resource) {
		for k, v := range groupMap {
			ok, err := regexp.MatchString(k, r.Name)
			if err != nil {
				panic(errors.Wrap(err, "cannot match regular expression"))
			}
			if ok {
				r.ShortGroup, r.Kind = v(r.Name)
				return
			}
		}
	}
}

var groupMap = map[string]GroupKindCalculator{
	// Note(turkenh): The following resources are listed under "Cloud Platform"
	// section in Terraform Documentation.
	"google_billing_subaccount$":                   ReplaceGroupWords("cloudplatform", 0),
	"google_folder$":                               ReplaceGroupWords("cloudplatform", 0),
	"google_folder_iam.*":                          ReplaceGroupWords("cloudplatform", 0),
	"google_folder_organization.*":                 ReplaceGroupWords("cloudplatform", 0),
	"google_organization_iam.*":                    ReplaceGroupWords("cloudplatform", 0),
	"google_organization_policy.*":                 ReplaceGroupWords("cloudplatform", 0),
	"google_project$":                              ReplaceGroupWords("cloudplatform", 0),
	"google_project_iam.*":                         ReplaceGroupWords("cloudplatform", 0),
	"google_project_service.*":                     ReplaceGroupWords("cloudplatform", 0),
	"google_project_default_service_accounts$":     ReplaceGroupWords("cloudplatform", 0),
	"google_project_organization_policy$":          ReplaceGroupWords("cloudplatform", 0),
	"google_project_usage_export_bucket$":          ReplaceGroupWords("cloudplatform", 0),
	"google_service_account.*":                     ReplaceGroupWords("cloudplatform", 0),
	"google_service_networking_peered_dns_domain$": ReplaceGroupWords("cloudplatform", 0),

	// Resources in "Access Approval" group.
	// Note(turkenh): The following resources are listed under "Access Approval"
	// section in Terraform Documentation.
	"google_.+_approval_settings$": ReplaceGroupWords("accessapproval", 0),

	"google_access_context_manager.+": ReplaceGroupWords("", 3),
	"google_data_loss_prevention.+":   ReplaceGroupWords("", 3),

	"google_service_networking_connection$": ReplaceGroupWords("", 2),
	"google_active_directory.+":             ReplaceGroupWords("", 2),
	"google_app_engine.+":                   ReplaceGroupWords("", 2),
	"google_assured_workloads.+":            ReplaceGroupWords("", 2),
	"google_binary_authorization.+":         ReplaceGroupWords("", 2),
	"google_container_analysis.+":           ReplaceGroupWords("", 2),
	"google_deployment_manager.+":           ReplaceGroupWords("", 2),
	"google_dialogflow_cx.+":                ReplaceGroupWords("", 2),
	"google_essential_contacts.+":           ReplaceGroupWords("", 2),
	"google_game_services.+":                ReplaceGroupWords("", 2),
	"google_gke_hub.+":                      ReplaceGroupWords("", 2),
	"google_identity_platform.+":            ReplaceGroupWords("", 2),
	"google_ml_engine.+":                    ReplaceGroupWords("", 2),
	"google_network_management.+":           ReplaceGroupWords("", 2),
	"google_network_services.+":             ReplaceGroupWords("", 2),
	"google_resource_manager.+":             ReplaceGroupWords("", 2),
	"google_secret_manager.+":               ReplaceGroupWords("", 2),
	"google_storage_transfer.+":             ReplaceGroupWords("", 2),
	"google_org_policy.+":                   ReplaceGroupWords("", 2),
	"google_vertex_ai.+":                    ReplaceGroupWords("", 2),
	"google_vpc_access.+":                   ReplaceGroupWords("", 2),

	"google_cloud_asset.+":     ReplaceGroupWords("", 2),
	"google_cloud_build.+":     ReplaceGroupWords("", 2),
	"google_cloud_functions.+": ReplaceGroupWords("", 2),
	"google_cloud_identity.+":  ReplaceGroupWords("", 2),
	"google_cloud_iot.+":       ReplaceGroupWords("", 2),
	"google_cloud_tasks.+":     ReplaceGroupWords("", 2),
	"google_cloud_scheduler.+": ReplaceGroupWords("", 2),
	"google_cloud_run.+":       ReplaceGroupWords("", 2),

	"google_data_catalog.+": ReplaceGroupWords("", 2),
	"google_data_flow.+":    ReplaceGroupWords("", 2),
	"google_data_fusion.+":  ReplaceGroupWords("", 2),

	"google_os_config.+": ReplaceGroupWords("", 2),
	"google_os_login.+":  ReplaceGroupWords("", 2),
}

// ReplaceGroupWords uses given group as the group of the resource and removes
// a number of words in resource name before calculating the kind of the resource.
func ReplaceGroupWords(group string, count int) GroupKindCalculator {
	return func(resource string) (string, string) {
		// "google_cloud_run_domain_mapping": "cloudrun" -> (cloudrun, DomainMapping)
		words := strings.Split(strings.TrimPrefix(resource, "google_"), "_")
		if group == "" {
			group = strings.Join(words[:count], "")
		}
		snakeKind := strings.Join(words[count:], "_")
		return group, name.NewFromSnake(snakeKind).Camel
	}
}
