/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gke

import (
	"encoding/base64"
	"fmt"

	"google.golang.org/api/container/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	computev1alpha3 "github.com/crossplane/provider-gcp/apis/compute/v1alpha3"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const (
	// TODO(negz): Is this username special? I can't see any ClusterRoleBindings
	// that bind it to a role.
	adminUser = "admin"
)

// Client interface to perform cluster operations
type Client interface {
	CreateCluster(string, computev1alpha3.GKEClusterSpec) (*container.Cluster, error)
	GetCluster(zone, name string) (*container.Cluster, error)
	DeleteCluster(zone, name string) error
}

// ClusterClient implementation
type ClusterClient struct {
	projectID string
	client    *container.Service
}

// NewClusterClient return new instance of the Client based on credentials
func NewClusterClient(projectID string, service *container.Service) *ClusterClient {
	return &ClusterClient{
		projectID: projectID,
		client:    service,
	}
}

// CreateCluster creates a new GKE cluster.
func (c *ClusterClient) CreateCluster(name string, spec computev1alpha3.GKEClusterSpec) (*container.Cluster, error) {
	cr := &container.CreateClusterRequest{
		Cluster: &container.Cluster{
			Name:                  name,
			InitialClusterVersion: spec.ClusterVersion,
			InitialNodeCount:      spec.NumNodes,
			IpAllocationPolicy: &container.IPAllocationPolicy{
				UseIpAliases:               spec.EnableIPAlias,
				CreateSubnetwork:           spec.CreateSubnetwork,
				NodeIpv4CidrBlock:          spec.NodeIPV4CIDR,
				ClusterIpv4CidrBlock:       spec.ClusterIPV4CIDR,
				ServicesIpv4CidrBlock:      spec.ServiceIPV4CIDR,
				ServicesSecondaryRangeName: spec.ServicesSecondaryRangeName,
				ClusterSecondaryRangeName:  spec.ClusterSecondaryRangeName,
			},
			NodeConfig: &container.NodeConfig{
				MachineType: spec.MachineType,
				OauthScopes: spec.Scopes,
			},
			ResourceLabels: spec.Labels,
			Zone:           spec.Zone,
			Network:        spec.Network,
			Subnetwork:     spec.Subnetwork,

			// As of Kubernetes 1.12 GKE must be asked to generate a client cert
			// that will be available via the GKE MasterAuth API. The certificate is
			// generated with CN=client - a user with no RBAC permissions. Instead
			// we user basic auth, which is still granted full admin privileges.
			MasterAuth: &container.MasterAuth{
				Username: adminUser,
				ClientCertificateConfig: &container.ClientCertificateConfig{
					IssueClientCertificate: false,
				},
			},
		},
		ProjectId: c.projectID,
		Zone:      spec.Zone,
	}

	if _, err := c.client.Projects.Zones.Clusters.Create(cr.ProjectId, spec.Zone, cr).Do(); err != nil {
		return nil, err
	}

	return c.GetCluster(spec.Zone, name)
}

// GetCluster retrieve GKE Cluster based on provided zone and name
func (c *ClusterClient) GetCluster(zone, name string) (*container.Cluster, error) {
	return c.client.Projects.Zones.Clusters.Get(c.projectID, zone, name).Do()
}

// DeleteCluster in the given zone with the given name
func (c *ClusterClient) DeleteCluster(zone, name string) error {
	_, err := c.client.Projects.Zones.Clusters.Delete(c.projectID, zone, name).Do()
	if err != nil {
		if gcp.IsErrorNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

// DefaultKubernetesVersion is the default Kubernetes Cluster version supported by GKE for given project/zone
func (c *ClusterClient) DefaultKubernetesVersion(zone string) (string, error) {
	sc, err := c.client.Projects.Zones.GetServerconfig(c.projectID, zone).Fields("validMasterVersions").Do()
	if err != nil {
		return "", err
	}

	return sc.DefaultClusterVersion, nil
}

// GenerateClientConfig generates a clientcmdapi.Config that can be used by any
// kubernetes client.
func GenerateClientConfig(cluster *container.Cluster) (clientcmdapi.Config, error) {
	c := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			cluster.Name: {
				Server: fmt.Sprintf("https://%s", cluster.Endpoint),
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			cluster.Name: {
				Cluster:  cluster.Name,
				AuthInfo: cluster.Name,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			cluster.Name: {
				Username: cluster.MasterAuth.Username,
				Password: cluster.MasterAuth.Password,
			},
		},
		CurrentContext: cluster.Name,
	}

	val, err := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.Clusters[cluster.Name].CertificateAuthorityData = val

	val, err = base64.StdEncoding.DecodeString(cluster.MasterAuth.ClientCertificate)
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.AuthInfos[cluster.Name].ClientCertificateData = val

	val, err = base64.StdEncoding.DecodeString(cluster.MasterAuth.ClientKey)
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.AuthInfos[cluster.Name].ClientKeyData = val

	return c, nil
}
