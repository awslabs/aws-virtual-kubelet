[![Validation](https://github.com/awslabs/aws-virtual-kubelet/actions/workflows/validation.yaml/badge.svg)](https://github.com/awslabs/aws-virtual-kubelet/actions/workflows/validation.yaml)  ![Coverage Badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jguice/237e23a1e28940815a9fced2b917012c/raw/main.json)

# AWS Virtual Kubelet
AWS Virtual Kubelet provides an extension to your [Kubernetes](https://kubernetes.io/) cluster that can provision and maintain [EC2](https://aws.amazon.com/ec2/) based [Pods](https://kubernetes.io/docs/concepts/workloads/pods/).  These EC2 pods can run arbitrary applications which might not otherwise fit into containers.

This expands the management capabilities of Kubernetes, enabling use-cases such as macOS native application lifecycle control via standard Kubernetes tooling.

## Architecture
A typical [EKS](https://aws.amazon.com/eks/) Kubernetes (k8s) cluster is shown in the diagram below.  It consists of a k8s API layer, a number of _nodes_ which each run a `kubelet` process, and pods (one or more containerized apps) managed by those `kubelet` processes.

Using the [Virtual Kubelet library](https://github.com/virtual-kubelet/virtual-kubelet), this EC2 provider implements a _virtual kubelet_ which looks like a typical `kubelet` to k8s.  API requests to create workload pods, etc. are received by the _virtual kubelet_ and passed to our custom EC2 provider.

This provider implements pod-handling endpoints using EC2 instances and an agent that runs on them.  The agent is responsible for launching and terminating "containers" (applications) and reporting status.  The provider ↔ agent API contract is defined using the [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3) spec and implemented via [gRPC](https://grpc.io/).  This enables agents to be written in any support language and run on a variety of operating systems and architectures [^1].

Nodes are represented by [ENIs](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html) that maintain a predictable IP address used for naming and consistent association of workload pods with _virtual kubelet_ instances. 

![](docs/img/vk.png)

See [Software Architecture](docs/SoftwareArchitecture.md) for an overview of the code organization and general behavior.  For detailed coverage of specific aspects of system/code behavior, see [implemented RFCs](docs/rfcs/implemented).

### Components
<dl>
  <dt>Virtual Kubelet (VK)</dt>
  <dd>Upstream library / framework for implementing custom Kubernetes providers</dd>
  <dt>Virtual Kubelet Provider (VKP)</dt>
  <dd>This EC2-based provider implementation (sometimes referred to as <i>virtual-kubelet</i> or <b>VK</b> also)</dd>
  <dt>Virtual Kubelet Virtual Machine (VKVM)</dt>
  <dd>The Virtual Machine providing compute for this provider implementation (i.e. an Amazon EC2 Instance)</dd>
  <dt>Virtual Kubelet Virtual Machine Agent (VKVMA)</dt>
  <dd>The gRPC agent that exposes an API to manage workloads on EC2 instances (also VKVMAgent, or just Agent)</dd>
</dl>

### Mapping to Kubernetes components
**kubelet** → Virtual Kubelet library + this custom EC2 provider  
**node** → Elastic Network Interface (managed by VKP)  
**pod** → EC2 Instance + VKVMAgent + Custom Workload  

## Prerequisites
The following are required to build and deploy this project.  Additional tools may be needed to utilize examples or set up a development environment.

### Go (lang)
Tested with [Go](https://golang.org) v1.12, 1.16, and 1.17.  See the [Go documentation](https://golang.org/doc/install) for installation steps.

### Docker
[Docker](https://www.docker.com/) is a container virtualization runtime.

See [Get Started](https://www.docker.com/get-started) in the docker documentation for setup steps.

### AWS account
The provider interacts directly with AWS APIs and launches EC2 instances so an AWS account is needed.  Click **Create an AWS Account** at https://aws.amazon.com/ to get started.

### AWS command line interface
Some commands utilize the AWS CLI.  See the [AWS CLI](https://aws.amazon.com/cli/) page for installation and configuration instructions.

### Kubernetes cluster
EKS is _strongly_ recommended, though any k8s cluster with sufficient access to make AWS API calls and communicate over the network with gRPC agents _could_ work.

## Infrastructure QuickStart
To get the needed infrastructure up and running quickly, see the deploy [README](deploy/vkstack/README.md) which details using the [AWS CDK](https://aws.amazon.com/cdk/) Infrastructure-as-Code framework to automatically provision the required resources.

## Build
Once the required infrastructure is in place, follow the steps in this section to build the VK provider.

### Makefile
This project comes with a [Makefile](https://www.gnu.org/software/make/manual/make.html#Introduction) to simplify build-related tasks.

Run `make` in this directory to get a list of subcommands and their description.

Some commands (such as `make push`) require appropriately set Environment Variables to function correctly.  Review variables at the top of the Makefile with `?=` and set in your shell/environment before running these commands.

1. Run `make build` to build the project.  This will also generate protobuf files and other generated files if needed.
2. Next run `make docker` to create a docker image with the `virtual-kubelet` binary.
3. Run `make push` to deploy the docker image to your [Elastic Container Registry](https://aws.amazon.com/ecr/).

## Deploy
Now we're ready to deploy the VK provider using the steps outlined in this section.

Some commands below utilize the [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl) tool to manage and configure k8s.  Other tools such as [Lens](https://k8slens.dev/index.html) may be used if desired (adapt instructions accordingly).

Example files that require updating placeholders with actual (environment-specific) data are copied to `./local` before modification.  The `local` directory's contents are ignored, which prevents accidental commits and _leaking_ account numbers, etc. into the GitHub repo.

### Cluster Role and Binding
The [ClusterRole](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole) and [Binding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) give VK pods the necessary permissions to manage k8s workloads.

1. Run `kubectl apply -f deploy/vk-clusterrole_binding.yaml` to deploy the cluster role and binding.

### ConfigMap
The [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/) provides global and default VK/VKP configuration elements.  Some of these settings may be overridden on a per-pod basis.

1. Copy the provided [examples/config-map.yaml](examples/config-map.yaml) to the `./local` dir and modify as-needed.  See [Config](docs/Config.md) for a detailed explanation of the various configuration options.

2. Next, run `kubectl apply -f local/config-map.yaml` to deploy the config map.

### StatefulSet
This configuration will deploy a set of VK providers using the docker image built and pushed earlier.

1. Copy the provided [examples/vk-statefulset.yaml](examples/vk-statefulset.yaml) file to `./local`.
2. Replace these placeholders in the `image:` reference with the values from your account/environment
   1. `AWS_ACCOUNT_ID`
   2. `AWS_REGION`
   3. `DOCKER_TAG`
3. Run `kubectl apply -f local/vk-statefulset.yaml` to deploy the VK provider pods.

## Usage
At this point you should have at least one running VK provider pod running successfully.  This section describes how to launch EC2-backed pods using the provider.

[examples/pods](examples/pods) contains both a single (unmanaged) pod example and a pod [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) example.

**`NOTE`** It is _strongly_ recommended that workload pods run via a supervisory management construct such as a Deployment (even for single-instance pods).  This will help minimize unexpected loss of pod resources and allow Kubernetes to efficiently use resources.

1. Copy the desired pod example(s) to `./local`
    2. Run `kubectl apply -f <filename>` (replacing `<filename>` with the actual file name).

See the [Cookbook](docs/Cookbook.md) for more usage examples.

## Frequently Asked Questions
### Why does this project exist?
This project serves as a translation and mediation layer between Kubernetes and EC2-based pods.  It was created in order to run custom workloads directly on any EC2 instance type/size available via AWS (e.g. [Mac Instances](https://aws.amazon.com/ec2/instance-types/mac/)).

### How can I use it?
1. Follow the steps in this README to get all the infrastructure and requirements in place and working with the example agent.
2. Using the example agent as a guide, implement your own gRPC agent to support the desired workloads.

### How can I help?
Take a look at the [`good first issue` Issues](https://github.com/awslabs/aws-virtual-kubelet/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).  Read the [CONTRIBUTING](docs/Contributing.md) guidelines and submit a Pull Request! 🚀

### Are there any known issues and/or planned features or improvements?
Yes. See [RFCs](docs/rfcs/README.md) for improvement proposals and [EdgeCases](docs/EdgeCases.md) for known issues / workarounds.

[BacklogFodder](docs/BacklogFodder.md) contains additional items that may become roadmap elements.

### Are there metrics I can use to monitor system state / behavior?
Yes.  See [Metrics](docs/Metrics.md) for details.

## Security
See [CONTRIBUTING](docs/Contributing.md#security-issue-notifications) for more information.

## License
This project is licensed under the [Apache-2.0 License](https://www.apache.org/licenses/LICENSE-2.0).

## Style Guide
### Go
[gofmt](https://pkg.go.dev/cmd/gofmt) formatting is enforced via GitHub Actions workflow.

[^1]: A Golang sample agent is included.
