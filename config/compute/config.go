package compute

import (
	"context"
	"fmt"

	"github.com/crossplane/terrajet/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/crossplane/provider-gcp/config/common"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) { //nolint: gocyclo
	// Note(turkenh): We ignore gocyclo in this function since it configures
	//  all resources separately and no complex logic here.

	p.AddResourceConfigurator("google_compute_managed_ssl_certificate", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.Kind = "ManagedSSLCertificate"
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/global/sslCertificates/%s", project, externalName), nil
		}
		r.UseAsync = true
	})

	p.AddResourceConfigurator("google_compute_subnetwork", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			region, err := common.GetField(parameters, "region")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", project, region, externalName), nil
		}
		r.References["network"] = config.Reference{
			Type: "Network",
		}
		r.UseAsync = true
	})

	p.AddResourceConfigurator("google_compute_address", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			region, err := common.GetField(parameters, "region")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/regions/%s/addresses/%s", project, region, externalName), nil
		}
		r.References["network"] = config.Reference{
			Type: "Network",
		}
		r.References["subnetwork"] = config.Reference{
			Type: "Subnetwork",
		}
	})

	p.AddResourceConfigurator("google_compute_firewall", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/global/firewalls/%s", project, externalName), nil
		}
		r.References["network"] = config.Reference{
			Type:      "Network",
			Extractor: common.PathSelfLinkExtractor,
		}
	})

	p.AddResourceConfigurator("google_compute_router", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			region, err := common.GetField(parameters, "region")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/regions/%s/routers/%s", project, region, externalName), nil
		}
		r.References["network"] = config.Reference{
			Type:      "Network",
			Extractor: common.PathSelfLinkExtractor,
		}
	})

	p.AddResourceConfigurator("google_compute_router_nat", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			region, err := common.GetField(parameters, "region")
			if err != nil {
				return "", err
			}
			router, err := common.GetField(parameters, "router")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/regions/%s/routers/%s/%s", project, region, router, externalName), nil
		}
		r.References["router"] = config.Reference{
			Type: "Router",
		}
		r.References["subnetwork.name"] = config.Reference{
			Type: "Subnetwork",
		}
		r.UseAsync = true
	})

	p.AddResourceConfigurator("google_compute_instance_template", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		// Note(turkenh): We have to modify schema of
		// "boot_disk.initialize_params.labels", since it is a map where
		// elements configured as nil, defaulting to map[string]string:
		r.TerraformResource.Schema["metadata"].Elem = schema.TypeString

		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/global/instanceTemplates/%s", project, externalName), nil
		}
	})

	p.AddResourceConfigurator("google_compute_instance", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		// Note(turkenh): We have to modify schema of
		// "boot_disk.initialize_params", since "labels" key here is a map where
		// elements configured as nil, defaulting to map[string]string:
		// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_instance#nested_initialize_params
		r.TerraformResource.
			Schema["boot_disk"].Elem.(*schema.Resource).
			Schema["initialize_params"].Elem.(*schema.Resource).
			Schema["labels"].Elem = schema.TypeString

		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			zone, err := common.GetField(parameters, "zone")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/zones/%s/instances/%s", project, zone, externalName), nil
		}

		r.References["network_interface.network"] = config.Reference{
			Type: "Network",
		}
		r.References["network_interface.subnetwork"] = config.Reference{
			Type: "Subnetwork",
		}

		r.UseAsync = true
	})

	p.AddResourceConfigurator("google_compute_instance_from_template", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		// Note(turkenh): We have to modify schema of
		// "boot_disk.initialize_params.labels", since it is a map where
		// elements configured as nil, defaulting to map[string]string:
		// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_instance#nested_initialize_params
		r.TerraformResource.
			Schema["boot_disk"].Elem.(*schema.Resource).
			Schema["initialize_params"].Elem.(*schema.Resource).
			Schema["labels"].Elem = schema.TypeString
		r.TerraformResource.Schema["metadata"].Elem = schema.TypeString

		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			zone, err := common.GetField(parameters, "zone")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/zones/%s/instances/%s", project, zone, externalName), nil
		}

		r.References["network_interface.network"] = config.Reference{
			Type: "Network",
		}
		r.References["network_interface.subnetwork"] = config.Reference{
			Type: "Subnetwork",
		}

		r.UseAsync = true
	})

	p.AddResourceConfigurator("google_compute_network", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		// projects/{{project}}/global/networks/{{name}}
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			return fmt.Sprintf("projects/%s/global/networks/%s", project, externalName), err
		}
	})
}
