module github.com/crossplane/provider-gcp

go 1.13

require (
	cloud.google.com/go v0.45.1
	github.com/crossplane/crossplane v0.10.0-rc.0.20200410142608-84b1c08d1890
	github.com/crossplane/crossplane-runtime v0.7.0
	github.com/crossplane/crossplane-tools v0.0.0-20200303232609-b3831cbb446d
	github.com/google/go-cmp v0.4.0
	github.com/googleapis/gax-go v1.0.3
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af // indirect
	github.com/mitchellh/copystructure v1.0.0
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/tools v0.0.0-20200328031815-3db5fc6bac03 // indirect
	google.golang.org/api v0.20.0
	google.golang.org/genproto v0.0.0-20200312145019-da6875a35672
	google.golang.org/grpc v1.28.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	honnef.co/go/tools v0.0.1-2019.2.2 // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/controller-tools v0.2.4
	sigs.k8s.io/yaml v1.1.0
)
