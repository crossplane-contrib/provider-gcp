package cloudiot

import (
	"github.com/crossplane/terrajet/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_cloudiot_device", func(r *config.Resource) {
		// Note(turkenh): We have to modify schema of
		// "last_error_status.details", since it is a map where elements
		// configured as nil, but needs to be String:
		r.TerraformResource.
			Schema["last_error_status"].Elem.(*schema.Resource).
			Schema["details"].Elem = schema.TypeString
	})

	p.AddResourceConfigurator("google_cloudiot_registry", func(r *config.Resource) {
		// Note(turkenh): We have to modify schema of
		// "credentials.public_key_certificate", since it is a map where elements
		// configured as nil, but needs to be String:
		r.TerraformResource.
			Schema["credentials"].Elem.(*schema.Resource).
			Schema["public_key_certificate"].Elem = schema.TypeString
		r.TerraformResource.Schema["http_config"].Elem = schema.TypeString
		r.TerraformResource.Schema["mqtt_config"].Elem = schema.TypeString
		r.TerraformResource.Schema["state_notification_config"].Elem = schema.TypeString
	})
}
