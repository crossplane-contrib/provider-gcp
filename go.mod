module github.com/crossplane/provider-gcp

go 1.13

replace (
	google.golang.org/api => google.golang.org/api v0.35.0
	google.golang.org/grpc => google.golang.org/grpc v1.28.0
)

require (
	cloud.google.com/go v0.65.0
	cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/storage v1.10.0
	github.com/crossplane/crossplane-runtime v0.12.1-0.20210219155338-30a941c3c3c6
	github.com/crossplane/crossplane-tools v0.0.0-20201201125637-9ddc70edfd0d
	github.com/google/go-cmp v0.5.2
	github.com/googleapis/gax-go v1.0.3
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/imdario/mergo v0.3.10
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/copystructure v1.0.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	google.golang.org/api v0.30.0
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a
	google.golang.org/grpc v1.31.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/controller-tools v0.3.0
)
