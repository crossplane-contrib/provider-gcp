package monitoring

import (
	"github.com/crossplane/terrajet/pkg/config"

	"github.com/crossplane/provider-gcp/config/common"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_monitoring_alert_policy", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
	})

	p.AddResourceConfigurator("google_monitoring_notification_channel", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
	})

	p.AddResourceConfigurator("google_monitoring_uptime_check_config", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
	})
}
