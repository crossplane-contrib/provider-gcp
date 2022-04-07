package container

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/crossplane/terrajet/pkg/config"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/crossplane/provider-gcp/config/common"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_container_cluster", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.Kind = "Cluster"
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"cluster_ipv4_cidr", "ip_allocation_policy"},
		}
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			location, err := common.GetField(parameters, "location")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, externalName), nil
		}
		r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]interface{}) (map[string][]byte, error) {
			name, err := common.GetField(attr, "name")
			if err != nil {
				return nil, err
			}
			server, err := common.GetField(attr, "endpoint")
			if err != nil {
				return nil, err
			}
			caData, err := common.GetField(attr, "master_auth[0].cluster_ca_certificate")
			if err != nil {
				return nil, err
			}
			caDataBytes, err := base64.StdEncoding.DecodeString(caData)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot serialize cluster ca data")
			}
			kc := clientcmdapi.Config{
				Kind:       "Config",
				APIVersion: "v1",
				Clusters: map[string]*clientcmdapi.Cluster{
					name: {
						Server:                   server,
						CertificateAuthorityData: caDataBytes,
					},
				},
				AuthInfos: map[string]*clientcmdapi.AuthInfo{
					name: {},
				},
				Contexts: map[string]*clientcmdapi.Context{
					name: {
						Cluster:  name,
						AuthInfo: name,
					},
				},
				CurrentContext: name,
			}
			kcb, err := clientcmd.Write(kc)
			if err != nil {
				return nil, errors.Wrap(err, "cannot serialize kubeconfig")
			}
			return map[string][]byte{
				"kubeconfig": kcb,
			}, nil
		}
		r.UseAsync = true
	})

	p.AddResourceConfigurator("google_container_node_pool", func(r *config.Resource) {
		r.Version = common.VersionV1alpha2
		r.Kind = "NodePool"
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		r.ExternalName.GetIDFn = func(_ context.Context, externalName string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			project, err := common.GetField(providerConfig, common.KeyProject)
			if err != nil {
				return "", err
			}
			clusterID, err := common.GetField(parameters, "cluster")
			if err != nil {
				return "", err
			}
			location := strings.Split(clusterID, "/")[3]
			cluster := strings.Split(clusterID, "/")[5]
			return fmt.Sprintf("%s/%s/%s/%s", project, location, cluster, externalName), nil
		}
		r.References["cluster"] = config.Reference{
			Type:      "Cluster",
			Extractor: common.ExtractResourceIDFuncPath,
		}
		r.UseAsync = true
	})
}
