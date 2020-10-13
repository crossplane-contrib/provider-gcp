module github.com/crossplane/provider-gcp

go 1.13

replace github.com/crossplane/crossplane-runtime => github.com/negz/crossplane-runtime v0.0.0-20201013014754-ee8aafd46cb5

replace (
	google.golang.org/api => google.golang.org/api v0.21.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200331122359-1ee6d9798940
	google.golang.org/grpc => google.golang.org/grpc v1.28.0
)

require (
	cloud.google.com/go v0.57.0
	cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/storage v1.6.0
	github.com/crossplane/crossplane v0.13.0
	github.com/crossplane/crossplane-runtime v0.10.0
	github.com/crossplane/crossplane-tools v0.0.0-20201007233256-88b291e145bb
	github.com/google/go-cmp v0.5.0
	github.com/googleapis/gax-go v1.0.3
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/imdario/mergo v0.3.9
	github.com/mitchellh/copystructure v1.0.0
	github.com/pkg/errors v0.9.1
	google.golang.org/api v0.22.0
	google.golang.org/genproto v0.0.0-20200527145253-8367513e4ece
	google.golang.org/grpc v1.29.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.2.4
)
