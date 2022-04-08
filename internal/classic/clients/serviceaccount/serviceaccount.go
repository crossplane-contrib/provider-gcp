/*
Copyright 2020 The Crossplane Authors.

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

package serviceaccount

import "google.golang.org/api/iam/v1"

// Client should be satisfied to conduct SA operations.
type Client interface {
	Create(name string, createserviceaccountrequest *iam.CreateServiceAccountRequest) *iam.ProjectsServiceAccountsCreateCall
	Get(name string) *iam.ProjectsServiceAccountsGetCall
	Patch(name string, patchserviceaccountrequest *iam.PatchServiceAccountRequest) *iam.ProjectsServiceAccountsPatchCall
	Delete(name string) *iam.ProjectsServiceAccountsDeleteCall
}
