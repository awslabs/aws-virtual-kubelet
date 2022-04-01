module github.com/aws/aws-virtual-kubelet

go 1.14

require (
	github.com/aws/aws-sdk-go-v2 v1.9.0
	github.com/aws/aws-sdk-go-v2/config v1.1.7
	github.com/aws/aws-sdk-go-v2/internal/ini v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.16.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.11.0
	github.com/creasty/defaults v1.5.2
	github.com/gogo/googleapis v1.4.1
	github.com/golang/mock v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/virtual-kubelet/node-cli v0.7.0
	github.com/virtual-kubelet/virtual-kubelet v1.6.0
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.7.1
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
