package dataproc

import (
	"github.com/crossplane/terrajet/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_dataproc_cluster", func(r *config.Resource) {
		// Note(turkenh): We have to modify schema of
		// "cluster_config.software_config.properties", since it is a map where
		// elements configured as nil, but needs to be String:
		r.TerraformResource.
			Schema["cluster_config"].Elem.(*schema.Resource).
			Schema["software_config"].Elem.(*schema.Resource).
			Schema["properties"].Elem = schema.TypeString
	})
}
