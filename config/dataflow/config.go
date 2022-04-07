package dataflow

import (
	"github.com/crossplane/terrajet/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_dataflow_job", func(r *config.Resource) {
		// Note(turkenh): We have to modify schema of "labels", since is a map
		// where elements configured as nil configured as nil, but needs to be
		// String:
		r.TerraformResource.Schema["labels"].Elem = schema.TypeString
		r.TerraformResource.Schema["parameters"].Elem = schema.TypeString
		r.TerraformResource.Schema["transform_name_mapping"].Elem = schema.TypeString
	})
}
