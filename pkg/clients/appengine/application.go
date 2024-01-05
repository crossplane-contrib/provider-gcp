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

package appengine

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	appengine "google.golang.org/api/appengine/v1"

	"github.com/crossplane/provider-gcp/apis/appengine/v1alpha1"
	gcp "github.com/crossplane/provider-gcp/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// GenerateApplication generates an *appengine.Application from ApplicationParameters.
func GenerateApplication(project string, in v1alpha1.ApplicationParameters, out *appengine.Application) { // nolint:gocyclo
	out.Id = project
	out.AuthDomain = gcp.StringValue(in.AuthDomain)
	out.DefaultCookieExpiration = gcp.StringValue(in.DefaultCookieExpiration)
	if len(in.DispatchRules) > 0 {
		out.DispatchRules = make([]*appengine.UrlDispatchRule, len(in.DispatchRules))
		for i, r := range in.DispatchRules {
			out.DispatchRules[i] = &appengine.UrlDispatchRule{
				Domain:  r.Domain,
				Path:    r.Path,
				Service: r.Service,
			}
		}
	}
	if in.FeatureSettings != nil {
		out.FeatureSettings = &appengine.FeatureSettings{
			SplitHealthChecks:       in.FeatureSettings.SplitHealthChecks,
			UseContainerOptimizedOs: in.FeatureSettings.UseContainerOptimizedOs,
		}
	}

	out.GcrDomain = gcp.StringValue(in.GcrDomain)
	out.LocationId = gcp.StringValue(in.LocationID)
}

// GenerateObservation generates an ApplicationObservation from an appengine.Application.
func GenerateObservation(in appengine.Application) v1alpha1.ApplicationObservation {
	return v1alpha1.ApplicationObservation{
		CodeBucket:      in.CodeBucket,
		DefaultBucket:   in.DefaultBucket,
		DefaultHostname: in.DefaultHostname,
		Name:            in.Name,
		ServingStatus:   in.ServingStatus,
	}
}

// LateInitializeSpec fills unassigned fields with the values in appengine.Application.
func LateInitializeSpec(spec *v1alpha1.ApplicationParameters, in appengine.Application) {
	spec.AuthDomain = gcp.LateInitializeString(spec.AuthDomain, in.AuthDomain)
	spec.DefaultCookieExpiration = gcp.LateInitializeString(spec.DefaultCookieExpiration, in.DefaultCookieExpiration)
	if len(spec.DispatchRules) == 0 && len(in.DispatchRules) != 0 {
		spec.DispatchRules = make([]*v1alpha1.URLDispatchRule, len(in.DispatchRules))
		for i, r := range in.DispatchRules {
			spec.DispatchRules[i] = &v1alpha1.URLDispatchRule{
				Domain:  r.Domain,
				Path:    r.Path,
				Service: r.Service,
			}
		}
	}
	if spec.FeatureSettings == nil && in.FeatureSettings != nil {
		spec.FeatureSettings = &v1alpha1.FeatureSettings{
			SplitHealthChecks:       in.FeatureSettings.SplitHealthChecks,
			UseContainerOptimizedOs: in.FeatureSettings.UseContainerOptimizedOs,
		}
	}
	spec.GcrDomain = gcp.LateInitializeString(spec.GcrDomain, in.GcrDomain)
	spec.LocationID = gcp.LateInitializeString(spec.LocationID, in.LocationId)
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(project string, in *v1alpha1.ApplicationParameters, observed *appengine.Application) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*appengine.Application)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateApplication(project, *in, desired)
	return cmp.Equal(desired, observed, cmpopts.EquateEmpty()), nil
}
