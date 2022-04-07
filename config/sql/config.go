package sql

import (
	"context"
	"fmt"

	"github.com/crossplane/terrajet/pkg/config"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane/provider-gcp/config/common"
)

// CloudSQL connection detail keys
const (
	CloudSQLSecretServerCACertificateCertKey            = "serverCACertificateCert"
	CloudSQLSecretServerCACertificateCommonNameKey      = "serverCACertificateCommonName"
	CloudSQLSecretServerCACertificateCreateTimeKey      = "serverCACertificateCreateTime"
	CloudSQLSecretServerCACertificateExpirationTimeKey  = "serverCACertificateExpirationTime"
	CloudSQLSecretServerCACertificateSha1FingerprintKey = "serverCACertificateSha1Fingerprint"

	CloudSQLSecretConnectionName = "connectionName"

	PrivateIPKey = "privateIP"
	PublicIPKey  = "publicIP"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) { //nolint:gocyclo
	p.AddResourceConfigurator("google_sql_database_instance", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/instances/%s", project, externalName), nil
		}

		// NOTE(@tnthornton) most of the connection details that were exported
		// to the connection details secret are marked as non-sensitive for tf.
		// We need to manually construct the secret details for those items.
		r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]interface{}) (map[string][]byte, error) {
			conn := map[string][]byte{}
			if a, ok := attr["connection_name"].(string); ok {
				conn[CloudSQLSecretConnectionName] = []byte(a)
			}
			if a, ok := attr["private_ip_address"].(string); ok {
				conn[PrivateIPKey] = []byte(a)
			}
			if a, ok := attr["public_ip_address"].(string); ok {
				conn[PublicIPKey] = []byte(a)
			}
			if a, ok := attr["root_password"].(string); ok {
				conn[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(a)
			}
			// map
			if certSlice, ok := attr["server_ca_cert"].([]interface{}); ok {
				if certattrs, ok := certSlice[0].(map[string]interface{}); ok {
					if a, ok := certattrs["cert"].(string); ok {
						conn[CloudSQLSecretServerCACertificateCertKey] = []byte(a)
					}
					if a, ok := certattrs["common_name"].(string); ok {
						conn[CloudSQLSecretServerCACertificateCommonNameKey] = []byte(a)
					}
					if a, ok := certattrs["create_time"].(string); ok {
						conn[CloudSQLSecretServerCACertificateCreateTimeKey] = []byte(a)
					}
					if a, ok := certattrs["expiration_time"].(string); ok {
						conn[CloudSQLSecretServerCACertificateExpirationTimeKey] = []byte(a)
					}
					if a, ok := certattrs["sha1_fingerprint"].(string); ok {
						conn[CloudSQLSecretServerCACertificateSha1FingerprintKey] = []byte(a)
					}
				}
			}
			return conn, nil
		}

		r.UseAsync = true
	})
	p.AddResourceConfigurator("google_sql_database", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			instance, err := common.GetField(parameters, "instance")
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, externalName), nil
		}

		r.References["instance"] = config.Reference{
			Type: "DatabaseInstance",
		}

		r.UseAsync = true
	})
	p.AddResourceConfigurator("google_sql_source_representation_instance", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			instance, err := common.GetField(parameters, "instance")
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, externalName), nil
		}

		r.References["instance"] = config.Reference{
			Type: "DatabaseInstance",
		}

		r.UseAsync = true
	})
	p.AddResourceConfigurator("google_sql_user", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			instance, err := common.GetField(parameters, "instance")
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("%s/%s/%s", project, instance, externalName), nil
		}

		r.References["instance"] = config.Reference{
			Type: "DatabaseInstance",
		}

		r.UseAsync = true
	})
	p.AddResourceConfigurator("google_sql_ssl_cert", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.ExternalName = config.IdentifierFromProvider
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			if externalName == "" {
				return "", nil
			}
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			instance, err := common.GetField(parameters, "instance")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/instances/%s/sslCerts/%s", project, instance, externalName), nil
		}

		r.References["instance"] = config.Reference{
			Type: "DatabaseInstance",
		}

		r.UseAsync = true
	})
}
