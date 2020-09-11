package dns

import (
	"fmt"
	"strings"

	"google.golang.org/api/dns/v1"

	"github.com/crossplane/provider-gcp/apis/dns/v1alpha1"
)

// GenerateResourceRecordSet returns ResourceRecordSet.
func GenerateResourceRecordSet(name string, s v1alpha1.ResourceRecordSetParameters) *dns.ResourceRecordSet {
	rset := &dns.ResourceRecordSet{
		Name:    AppendDot(name),
		Rrdatas: s.Rrdatas,
	}

	if s.TTL != nil {
		rset.Ttl = *s.TTL
	}

	if s.Type != nil {
		rset.Type = *s.Type
	}

	return rset
}

// LateInitializeResourceRecordSet fills the empty fields in *v1alpha1.ResourceRecordSetParameters with
// the values seen in ManagedZone.
func LateInitializeResourceRecordSet(spec *v1alpha1.ResourceRecordSetParameters, obs *dns.ResourceRecordSet) {
	if obs == nil {
		return
	}

	if spec.TTL == nil {
		spec.TTL = &obs.Ttl
	}

	if spec.Type == nil {
		spec.Type = &obs.Type
	}
}

// IsUpToDateResourceRecordSet check whether the type in ResourceRecordSet and Response are same or not
func IsUpToDateResourceRecordSet(spec v1alpha1.ResourceRecordSetParameters, obs dns.ResourceRecordSet) bool {
	if obs.Type == "" {
		return true
	}

	if spec.Type == nil {
		return false
	}

	return *spec.Type == obs.Type
}

// AppendDot append dot for record.
func AppendDot(s string) string {
	if !strings.HasSuffix(s, ".") {
		return fmt.Sprintf("%s.", s)
	}
	return s
}
