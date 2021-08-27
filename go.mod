module github.com/crossplane/provider-gcp

go 1.16

require (
	cloud.google.com/go/storage v1.15.0
	github.com/crossplane/crossplane-runtime v0.13.1-0.20210531122928-ded177829557
	github.com/crossplane/crossplane-tools v0.0.0-20210320162312-1baca298c527
	github.com/google/go-cmp v0.5.6
	github.com/imdario/mergo v0.3.10
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/copystructure v1.0.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/scylladb/go-set v1.0.2
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/api v0.52.0
	google.golang.org/grpc v1.39.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/controller-tools v0.3.0
)
