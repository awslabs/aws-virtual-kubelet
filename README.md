[![Validation](https://github.com/awslabs/aws-virtual-kubelet/actions/workflows/validation.yaml/badge.svg)](https://github.com/awslabs/aws-virtual-kubelet/actions/workflows/validation.yaml)  ![Coverage Badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/jguice/237e23a1e28940815a9fced2b917012c/raw/master.json)

# AWS Virtual Kubelet
AWS Virtual Kubelet aims to provide an extension to your Kubernetes cluster that can provision and maintain EC2 instances through regular Kubernetes operations. This enables usage of non-standard operating systems for container ecosystems, such as MacOS.

[Virtual Kubelet](https://github.com/virtual-kubelet/virtual-kubelet) can be deployed as a binary and joined to an existing [Kubernetes](https://kubernetes.io/) cluster, however, it is recommended to deploy as a Pod to an existing cluster.

## Architecture
![](docs/img/vk.png)

### Components
<dl>
  <dt>Virtual Kubelet (VK)</dt>
  <dd>Upstream library / framework for implementing custom Kubernetes providers</dd>
  <dt>Virtual Kubelet Provider (VKP)</dt>
  <dd>This EC2-based provider implementation (sometimes referred to as <i>virtual-kubelet</i> or <b>VK</b> also)</dd>
  <dt>Virtual Kubelet Virtual Machine (VKVM)</dt>
  <dd>The Virtual Machine providing compute for this provider implementation (i.e. an Amazon EC2 Instance)</dd>
  <dt>Virtual Kubelet Virtual Machine Agent (VKVMA)</dt>
  <dd>The <a href="https://grpc.io/">gRPC</a> agent that exposes an API to manage workloads on EC2 instances (also VKVMAgent, or just Agent)</dd>
</dl>

### Mapping to Kubernetes components
**kubelet** → Virtual Kubelet library + this custom EC2 provider  
**node** → Elastic Network Interface (managed by VKP)  
**pod** → EC2 Instance + VKVMAgent + Custom Workload  

## Prerequisites

### Go (lang)
Tested with [Go](https://golang.org) v1.12, 1.16, and 1.17.  See the [Go documentation](https://golang.org/doc/install) for installation steps.

### Docker
[Docker](https://www.docker.com/) is a container virtualization runtime.

See [Get Started](https://www.docker.com/get-started) in the docker documentation for setup steps.

## Structure
This project uses this [Go Project Layout](https://github.com/golang-standards/project-layout) pattern.  A top-level `Makefile` provides necessary build and utility functions.  Run `make` by itself (or `make help`) to see a list of common targets.

## External Libraries Used
- [virtual-kubelet](https://github.com/virtual-kubelet/virtual-kubelet)
  - provides the Virtual Kubelet (VK) interface between this custom provider and [Kubernetes](https://kubernetes.io/)
- [node-cli](https://github.com/virtual-kubelet/node-cli)
  - abstracts the VK provider command interface into a separate, reusable project[^1]

[^1]: Previously VK providers were either part of the `virtual-kubelet` repository, or copied cmd code into their own repo

## Setup
For local development and testing setup see [DevSetup.md](docs/DevSetup.md)

To configure a pipeline and cluster in AWS see [PipelineSetup.md](docs/PipelineSetup.md)

## Usage
**`TODO`** These were mostly copied from existing docs and need reviewed, reordered, and updated.  [Cookbook.md](docs/Cookbook.md) contains some steps also that may need updated and/or relocated.

### Deploy a ConfigMap with required Virtual Kubelet configurations
Deploy this first, filling the values based on the Configuration section below.
```bash
kubectl apply -f examples/ConfigMap.yaml
```
### Deploy a Virtual Kubelet pod to a Kubernetes cluster on AWS
First, update `deploy/example_vk_sa/yaml` role_arn with your IAM role.
Second, update `deploy/example_vk_statefulset.yaml` with an updated `image:` value based on image registry location.

## Configuration
Create a configuration file (JSON) with the following keys and appropriate values:
**`TODO`** format the config parameters explanation below to be more readable
**`TODO`** update example JSON config file and link to it from here

ManagementSubnet: Subnet in which you expect to deploy the Virtual Kubelet, which generates an AWS ENI for the purposes of creating a unique location for the Kubenernetes IP address.
ClusterName: Included for tagging purposes to manage AWS ENIs associated with Virtual Kubelet.
Region: Code for AWS Region the Virtual Kubelet will be deployed to. e.g. "us-west-2" or "us-east-1".

VMConfig:
InitialSecurityGroups: AWS SecurityGroups assigned to an EC2 instance at launch time, which can be updated later.
DefaultAMI: AMI used when there is no other AMI specified in Podspec of a Kubernetes Pod.
InitData: Base64 encoded JSON to be processed by the Bootstrap Agent.

BootstrapAgent:
S3Bucket: Bucket location in S3 where bootstrap agent is located.
S3Key: Key location in S3 where bootstrap agent is located.
GRPCPort: Port number for GRPC communication between Virtual Kubelet and the EC2 instances it creates.
InitData: Base64 encoded JSON to be processed by the Bootstrap Agent.

WarmPoolConfig:
DesiredCount: Amount of EC2 to be maintained in the WarmPool, above and beyond what is required to run Kubernetes Pods.
IamInstanceProfile: The IAM instance profile assigned to the EC2 at launch time, which can be changed at Pod assignment time.
SecurityGroups: The AWS Security Groups assigned to the EC2 at launch time, which can be changed at Pod assignment time.
KeyPair: The EC2 credentials assigned to allow for SSH/RDP access to the instance. Unchangeable at Pod assignment time.
ImageID: The AWS AMI to launch the EC2 instances with, Unchangeable at Pod assignment time.
InstanceType: The AWS EC2 InstanceType, e.g. `mac1.metal`. Unchangeable at Pod assignment time.
Subnets: The AWS VPC Subnet(s) to deploy the WarmPool EC2 instances into. Unchangeable at Pod assignment time.

```bash
kubectl apply -f deploy/example_vk_sa.yaml
kubectl apply -f deploy/example_vk_statefulset.yaml
```

## Frequently Asked Questions
**`TODO`** add more FAQ items here as-needed
### Why does this project exist?
This project serves as a translation and mediation layer between Kubernetes and EC2-based pods.  It was created in order to run custom workloads directly on any EC2 instance type/size available via AWS (e.g. [Mac Instances](https://aws.amazon.com/ec2/instance-types/mac/)).

### How can I use it?
See **`TODO`** <insert link to doc here> for steps to customize this project for your particular needs.

## Security
See [CONTRIBUTING](docs/Contributing.md#security-issue-notifications) for more information.

## License
This project is licensed under the [Apache-2.0 License](https://www.apache.org/licenses/LICENSE-2.0).

## Style Guide
### Go
**`TODO`**

## Reference
**`TODO`** Add "article" and external reference links here
- [Some useful article](https://example.com)
