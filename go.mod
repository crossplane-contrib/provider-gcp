module github.com/crossplaneio/stack-gcp

go 1.13

require (
	cloud.google.com/go v0.45.1
	github.com/crossplaneio/crossplane v0.6.0-rc.0.20191226165033-a452562456e0
	github.com/crossplaneio/crossplane-runtime v0.3.1-0.20200115232149-cd8c52b483c3
	github.com/crossplaneio/crossplane-tools v0.0.0-20191220202319-9033bd8a02ce
	github.com/google/go-cmp v0.3.1
	github.com/googleapis/gax-go v1.0.3
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	golang.org/x/exp v0.0.0-20190731235908-ec7cb31e5a56 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.9.0
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.23.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	honnef.co/go/tools v0.0.1-2019.2.2 // indirect
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/controller-tools v0.2.4
	sigs.k8s.io/yaml v1.1.0
)
