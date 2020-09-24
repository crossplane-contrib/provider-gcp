module github.com/crossplane/provider-gcp

go 1.13

require (
	cloud.google.com/go v0.56.0
	cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/storage v1.6.0
	github.com/crossplane/crossplane v0.13.0-rc.0.20200923012106-3a18fb78178f
	github.com/crossplane/crossplane-runtime v0.9.1-0.20200922233740-7ad38b523afe
	github.com/crossplane/crossplane-tools v0.0.0-20200923030414-95b434323cd4
	github.com/google/go-cmp v0.4.0
	github.com/googleapis/gax-go v1.0.3
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/imdario/mergo v0.3.9
	github.com/mitchellh/copystructure v1.0.0
	github.com/pkg/errors v0.9.1
	google.golang.org/api v0.21.0
	google.golang.org/genproto v0.0.0-20200331122359-1ee6d9798940
	google.golang.org/grpc v1.28.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.2.4
)
