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

package resourcerecordset

import (
	"context"
	"fmt"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	dns "google.golang.org/api/dns/v1beta2"
	"google.golang.org/api/option"

	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
)

const (
	resourceRecordSetKind = "dns#resourceRecordSet"
)

type resourceRecordSetService interface {
	Create(project string, managedZone string, resourcerecordset *dns.ResourceRecordSet) *dns.ProjectsManagedZonesRrsetsCreateCall
	Get(project string, managedZone string, name string, type_ string) *dns.ProjectsManagedZonesRrsetsGetCall
	Patch(project string, managedZone string, name string, type_ string, resourcerecordset *dns.ResourceRecordSet) *dns.ProjectsManagedZonesRrsetsPatchCall
	Delete(project string, managedZone string, name string, type_ string) *dns.ProjectsManagedZonesRrsetsDeleteCall
}

type Client interface {
	Get(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSet, error)
	Create(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSet, error)
	Update(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSet, error)
	Delete(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSetsDeleteResponse, error)
}

type recordClient struct {
	rrs       resourceRecordSetService
	projectID string
}

func NewClient(ctx context.Context, projectID string, opts option.ClientOption) (Client, error) {
	s, err := dns.NewService(ctx, opts)
	if err != nil {
		return nil, err
	}
	return recordClient{rrs: dns.NewProjectsManagedZonesRrsetsService(s), projectID: projectID}, nil
}

func (c recordClient) Get(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSet, error) {
	req := c.rrs.Get(c.projectID, cr.Spec.ForProvider.ManagedZone, cr.Spec.ForProvider.Name, cr.Spec.ForProvider.Type)
	return req.Context(ctx).Do()
}

func (c recordClient) Create(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSet, error) {
	r := &dns.ResourceRecordSet{
		Kind:             resourceRecordSetKind,
		Name:             getFQDNFromExternalName(cr),
		Rrdatas:          cr.Spec.ForProvider.Rrdatas,
		SignatureRrdatas: cr.Spec.ForProvider.SignatureRrdatas,
		Ttl:              int64(cr.Spec.ForProvider.TTL),
		Type:             cr.Spec.ForProvider.Type,
	}

	// The first parameter to the Create method is the resource name of the GCP project
	// where the service account should be created
	req := c.rrs.Create(c.projectID, cr.Spec.ForProvider.ManagedZone, r)
	return req.Context(ctx).Do()
}

func (c recordClient) Update(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSet, error) {
	r := &dns.ResourceRecordSet{}
	populateProviderFromCR(r, cr)
	req := c.rrs.Patch(c.projectID, cr.Spec.ForProvider.ManagedZone, getFQDNFromExternalName(cr), cr.Spec.ForProvider.Type, r)
	return req.Context(ctx).Do()
}

func (c recordClient) Delete(ctx context.Context, cr *v1alpha1.ResourceRecordSet) (*dns.ResourceRecordSetsDeleteResponse, error) {
	req := c.rrs.Delete(c.projectID, cr.Spec.ForProvider.ManagedZone, getFQDNFromExternalName(cr), cr.Spec.ForProvider.Type)
	return req.Context(ctx).Do()
}

func getFQDNFromExternalName(cr *v1alpha1.ResourceRecordSet) string {
	name := meta.GetExternalName(cr)
	if !strings.HasSuffix(name, ".") {
		return fmt.Sprintf("%s.", name)
	}
	return name
}

// IsUpToDate returns true if the supplied Kubernetes resource does not differ
//  from the supplied GCP resource. It considers only fields that can be
//  modified in place without deleting and recreating the ResourceRecordSet.
func IsUpToDate(in *v1alpha1.ResourceRecordSetParameters, observed *dns.ResourceRecordSet) bool {
	if in.TTL != int(observed.Ttl) {
		return false
	}

	if !cmp.Equal(in.Rrdatas, observed.Rrdatas, cmpopts.SortSlices(func(i, j string) bool { return i < j })) {
		return false
	}

	if !cmp.Equal(in.SignatureRrdatas, observed.SignatureRrdatas, cmpopts.SortSlices(func(i, j string) bool { return i < j })) {
		return false
	}

	return true
}

// GenerateObservation marshall API response into ResourceRecordSetObservation
func GenerateObservation(fromProvider *dns.ResourceRecordSet) v1alpha1.ResourceRecordSetObservation {
	ro := v1alpha1.ResourceRecordSetObservation{}
	ro.Name = fromProvider.Name
	ro.Rrdatas = fromProvider.Rrdatas
	ro.SignatureRrdatas = fromProvider.SignatureRrdatas
	ro.Type = fromProvider.Type
	ro.TTL = int(fromProvider.Ttl)
	return ro
}

func populateProviderFromCR(forProvider *dns.ResourceRecordSet, cr *v1alpha1.ResourceRecordSet) {
	forProvider.Kind = resourceRecordSetKind
	forProvider.Ttl = int64(cr.Spec.ForProvider.TTL)
	forProvider.Rrdatas = cr.Spec.ForProvider.Rrdatas
	forProvider.SignatureRrdatas = cr.Spec.ForProvider.SignatureRrdatas
}
