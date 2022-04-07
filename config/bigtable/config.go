package bigtable

import "github.com/crossplane/terrajet/pkg/config"

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_bigtable_gc_policy", func(r *config.Resource) {
		r.Kind = "GarbageCollectionPolicy"
	})
}
