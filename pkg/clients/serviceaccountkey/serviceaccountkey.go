/*
Copyright 2021 The Crossplane Authors.

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

package serviceaccountkey

import (
	"net/url"
	"path"

	"google.golang.org/api/iam/v1"

	"github.com/crossplane/provider-gcp/apis/iam/v1alpha1"
)

// Client should be satisfied to conduct ServiceAccountKey operations.
type Client interface {
	Create(name string, createserviceaccountkeyrequest *iam.CreateServiceAccountKeyRequest) *iam.ProjectsServiceAccountsKeysCreateCall
	Delete(name string) *iam.ProjectsServiceAccountsKeysDeleteCall
	Get(name string) *iam.ProjectsServiceAccountsKeysGetCall
}

// ParseKeyIDFromRrn parses key id from Google Cloud API relative resource name (resource path) of
//   a service account key
func ParseKeyIDFromRrn(rrn string) (string, error) {
	resourcePath, err := url.Parse(rrn)

	if err != nil {
		return "", err
	}

	return path.Base(resourcePath.Path), nil
}

// PopulateSaKey populates `v1alpha1.ServiceAccountKeyObservation` status from the specified API response
func PopulateSaKey(cr *v1alpha1.ServiceAccountKey, fromProvider *iam.ServiceAccountKey) error {
	keyID, err := ParseKeyIDFromRrn(fromProvider.Name)

	if err != nil {
		return err
	}

	cr.Status.AtProvider.KeyID = keyID
	cr.Status.AtProvider.Name = fromProvider.Name
	cr.Status.AtProvider.PrivateKeyType = fromProvider.PrivateKeyType
	cr.Status.AtProvider.KeyAlgorithm = fromProvider.KeyAlgorithm
	cr.Status.AtProvider.ValidAfterTime = fromProvider.ValidAfterTime
	cr.Status.AtProvider.ValidBeforeTime = fromProvider.ValidBeforeTime
	cr.Status.AtProvider.KeyOrigin = fromProvider.KeyOrigin
	cr.Status.AtProvider.KeyType = fromProvider.KeyType

	return nil
}
