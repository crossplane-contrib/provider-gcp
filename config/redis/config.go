package redis

import (
	"context"
	"fmt"

	"github.com/crossplane/terrajet/pkg/config"

	"github.com/crossplane/provider-gcp/config/common"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_redis_instance", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(ctx context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			region, err := common.GetField(parameters, "region")
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, externalName), nil
		}
	})
}
