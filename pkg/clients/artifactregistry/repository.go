package artifactregistry

import (
	"fmt"
	"strings"

	"github.com/crossplane/provider-gcp/apis/artifactregistry/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"

	"github.com/google/go-cmp/cmp"
	artifactregistry "google.golang.org/api/artifactregistry/v1beta2"
)

const (
	repositoryNameFormat   = "projects/%s/locations/%s/repositories/%s"
	repositoryParentFormat = "projects/%s/locations/%s"
)

// GetFullyQualifiedName builds the fully qualified name of the Repository.
func GetFullyQualifiedName(project, location, name string) string {
	return fmt.Sprintf(repositoryNameFormat, project, location, name)
}

// GetFullyQualifiedParent builds the fully qualified parent name of the Repository.
func GetFullyQualifiedParent(project, location string) string {
	return fmt.Sprintf(repositoryParentFormat, project, location)
}

// GenerateRepository produces a Repository that is configured via given RepositoryParameters.
func GenerateRepository(projectID, name string, p v1alpha1.RepositoryParameters) *artifactregistry.Repository {
	r := &artifactregistry.Repository{
		Name:        name,
		Description: p.Description,
		Format:      p.Format,
		Labels:      p.Labels,
		KmsKeyName:  gcp.StringValue(p.KmsKeyName),
	}
	setMavenConfig(p, r)
	return r
}

// setMavenConfig sets MavenConfig of Repository based on RepositoryParameters.
func setMavenConfig(p v1alpha1.RepositoryParameters, r *artifactregistry.Repository) {
	if p.MavenConfig != nil {
		r.MavenConfig = &artifactregistry.MavenRepositoryConfig{
			AllowSnapshotOverwrites: p.MavenConfig.AllowSnapshotOverwrites,
			VersionPolicy:           *p.MavenConfig.VersionPolicy,
		}

	}
}

// LateInitialize fills the empty fields of RepositoryParameters if the corresponding
// fields are given in Repository.
func LateInitialize(p *v1alpha1.RepositoryParameters, r artifactregistry.Repository) { // nolint:gocyclo

	if p.Description == "" && r.Description != "" {
		p.Description = r.Description
	}

	if p.Format == "" && r.Format != "" {
		p.Format = r.Format
	}

	if len(p.Labels) == 0 && len(r.Labels) != 0 {
		p.Labels = map[string]string{}
		for k, v := range r.Labels {
			p.Labels[k] = v
		}
	}
	if p.KmsKeyName == nil && len(r.KmsKeyName) != 0 {
		p.KmsKeyName = gcp.StringPtr(r.KmsKeyName)
	}
	if p.MavenConfig == nil && r.MavenConfig != nil {
		p.MavenConfig = &v1alpha1.MavenRepositoryConfig{
			AllowSnapshotOverwrites: r.MavenConfig.AllowSnapshotOverwrites,
			VersionPolicy:           &r.MavenConfig.VersionPolicy,
		}
	}
}

// IsUpToDate checks whether Repository is configured with given RepositoryParameters.
func IsUpToDate(projectID string, p v1alpha1.RepositoryParameters, r artifactregistry.Repository) bool {
	observed := &v1alpha1.RepositoryParameters{}
	LateInitialize(observed, r)

	return cmp.Equal(observed, &p)
}

// GenerateUpdateRequest produces a (Repository, updateMask) with the difference
// between RepositoryParameters and Repository.
func GenerateUpdateRequest(name string, p v1alpha1.RepositoryParameters, r artifactregistry.Repository) (*artifactregistry.Repository, string) {
	observed := &v1alpha1.RepositoryParameters{}
	LateInitialize(observed, r)
	ur := &artifactregistry.Repository{
		Name: name,
	}
	mask := []string{}

	if !cmp.Equal(p.Description, observed.Description) {
		mask = append(mask, "description")
		ur.Description = p.Description
	}

	if !cmp.Equal(p.MavenConfig, observed.MavenConfig) {
		mask = append(mask, "mavenConfig")
		if p.MavenConfig != nil {
			ur.MavenConfig = &artifactregistry.MavenRepositoryConfig{
				AllowSnapshotOverwrites: p.MavenConfig.AllowSnapshotOverwrites,
				VersionPolicy:           gcp.StringValue(p.MavenConfig.VersionPolicy),
			}
		}
	}
	if !cmp.Equal(p.Labels, observed.Labels) {
		mask = append(mask, "labels")
		ur.Labels = p.Labels
	}
	updateMask := strings.Join(mask, ",")
	return ur, updateMask
}
