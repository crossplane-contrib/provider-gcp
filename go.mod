module github.com/crossplane/provider-gcp

go 1.13

require (
	cloud.google.com/go v0.56.0
	cloud.google.com/go/storage v1.6.0
	github.com/crossplane/crossplane v0.11.0
	github.com/crossplane/crossplane-runtime v0.9.0
	github.com/crossplane/crossplane-tools v0.0.0-20200412230150-efd0edd4565b
	github.com/google/go-cmp v0.4.0
	github.com/googleapis/gax-go v1.0.3
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af // indirect
	github.com/mitchellh/copystructure v1.0.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.8.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.21.0
	google.golang.org/genproto v0.0.0-20200331122359-1ee6d9798940
	google.golang.org/grpc v1.28.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/controller-tools v0.2.4
	sigs.k8s.io/yaml v1.2.0
)
