package accessapproval

import (
	"github.com/crossplane/terrajet/pkg/config"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_folder_access_approval_settings", func(r *config.Resource) {
		r.Kind = "FolderSettings"
	})
	p.AddResourceConfigurator("google_organization_access_approval_settings", func(r *config.Resource) {
		r.Kind = "OrganizationSettings"
	})
	p.AddResourceConfigurator("google_project_access_approval_settings", func(r *config.Resource) {
		r.Kind = "ProjectSettings"
	})
}
