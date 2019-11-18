module github.com/crossplaneio/stack-gcp

go 1.12

replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
)

require (
	cloud.google.com/go v0.43.0
	github.com/crossplaneio/crossplane v0.4.0
	github.com/crossplaneio/crossplane-runtime v0.2.3-0.20191118184433-bc1193ea4781
	github.com/crossplaneio/crossplane-tools v0.0.0-20191023215726-61fa1eff2a2e
	github.com/google/go-cmp v0.3.1
	github.com/googleapis/gax-go v1.0.3
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	golang.org/x/exp v0.0.0-20190731235908-ec7cb31e5a56 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.8.0
	google.golang.org/genproto v0.0.0-20190801165951-fa694d86fc64
	google.golang.org/grpc v1.23.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	honnef.co/go/tools v0.0.1-2019.2.2 // indirect
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf // indirect
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/controller-tools v0.2.2
)
