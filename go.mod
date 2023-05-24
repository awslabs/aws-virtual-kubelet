module github.com/aws/aws-virtual-kubelet

go 1.20

require (
	github.com/aws/aws-sdk-go-v2 v1.18.0
	github.com/aws/aws-sdk-go-v2/config v1.1.7
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.16.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.11.0
	github.com/creasty/defaults v1.7.0
	github.com/gogo/googleapis v1.4.1
	github.com/golang/mock v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.15.1
	github.com/prometheus/client_model v0.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.0
	github.com/virtual-kubelet/node-cli v0.7.0
	github.com/virtual-kubelet/virtual-kubelet v1.6.0
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.30.0
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.7.1
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.1.7 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.0.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.1.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.3.1 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.0+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/spdystream v0.0.0-20170912183627-bc6354cbbc29 // indirect
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20191227052852-215e87163ea7 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.opencensus.io v0.22.2 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/oauth2 v0.5.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.19.10 // indirect
	k8s.io/component-base v0.19.10 // indirect
	k8s.io/kube-openapi v0.0.0-20200805222855-6aeccd4b50c6 // indirect
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.15 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	//	github.com/aws/smithy-go v1.7.0 => github.com/aws/smithy-go v1.7.0
	//	github.com/virtual-kubelet/virtual-kubelet => github.com/virtual-kubelet/virtual-kubelet v1.6.0
	//	go.opencensus.io => go.opencensus.io v0.19.3
	//	k8s.io/api => k8s.io/api v0.19.10
	//	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.2
	//	k8s.io/apimachinery => k8s.io/apimachinery v0.19.10
	//	k8s.io/apiserver => k8s.io/apiserver v0.19.10
	//	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.10

	// Match virtual-kubelet/node-cli version of dependencies to resolve various errors
	// https://github.com/virtual-kubelet/node-cli/blob/bfe728730f54651b279d6e19e92599ede4e6fa1c/go.mod
	k8s.io/api => k8s.io/api v0.19.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.10
	k8s.io/client-go => k8s.io/client-go v0.19.10

//	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.2
//	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.2
//	k8s.io/code-generator => k8s.io/code-generator v0.22.2
//	k8s.io/component-base => k8s.io/component-base v0.22.2
//	k8s.io/component-helpers => k8s.io/component-helpers v0.22.2
//	k8s.io/controller-manager => k8s.io/controller-manager v0.22.2
//	k8s.io/cri-api => k8s.io/cri-api v0.22.2
//	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.2
//	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.2
//	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.2
//	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.2
//	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.2
//	k8s.io/kubectl => k8s.io/kubectl v0.22.2
//	k8s.io/kubelet => k8s.io/kubelet v0.22.2
//	k8s.io/kubernetes => k8s.io/kubernetes v1.22.2
//	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.2
//	k8s.io/metrics => k8s.io/metrics v0.22.2
//	k8s.io/mount-utils => k8s.io/mount-utils v0.22.2
//	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.2
//	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.2
)
