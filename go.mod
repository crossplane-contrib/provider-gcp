module github.com/crossplane/provider-gcp

go 1.16

require (
	cloud.google.com/go/iam v0.1.1 // indirect
	cloud.google.com/go/storage v1.15.0
	github.com/crossplane/crossplane-runtime v0.15.1-0.20211202230900-d43d510ec578
	github.com/crossplane/crossplane-tools v0.0.0-20210916125540-071de511ae8e
	github.com/google/go-cmp v0.5.6
	github.com/google/go-containerregistry v0.6.0
	github.com/imdario/mergo v0.3.12
	github.com/mitchellh/copystructure v1.0.0
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/api v0.64.0
	google.golang.org/genproto v0.0.0-20220126215142-9970aeb2e350 // indirect
	google.golang.org/grpc v1.40.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/controller-tools v0.6.2
)
