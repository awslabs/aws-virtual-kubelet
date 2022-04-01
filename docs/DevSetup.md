**`TODO`** review and update this doc (also remove duplicate content vs. [README.md](../README.md))

# Dev Setup
This document describes an approach to setting up your local development environment.  There are many possible approaches depending on experience, tool set, preference, and other variables.  The specific configuration described below isn't the _only way_ and can be adapted as-needed.

## Prerequisites

### Go (lang)
Tested with [Go](https://golang.org) v1.12, 1.16, and 1.17. See the [Go documentation](https://golang.org/doc/install) for installation steps.

### Docker
[Docker](https://www.docker.com/) is a container virtualization runtime.

See [Get Started](https://www.docker.com/get-started) in the docker documentation for setup steps.

### Kubernetes
A [Kubernetes](https://kubernetes.io/) (k8s) cluster is required to use this [Virtual Kubelet](https://github.com/virtual-kubelet/virtual-kubelet) implementation.  The version bundled with Docker Desktop has been tested successfully.

### Amazon Web Services
Since this virtual kubelet is designed to run EC2 instances in the AWS Cloud, you'll need to [setup an account](https://aws.amazon.com/free).  We'll configure credentials in a later step.

## Building
1. From the root dir of this repository, run `make build`.
2. Now try running the Virtual Kubelet with `./bin/virtual-kubelet version`.  You should see version information.

**`NOTE`** If you have any network issues with Go modules installing, try setting `GOPROXY=direct` in your environment.

## Setup
From here you'll need a Kubernetes cluster for the virtual kubelet to connect to and extend.  For development, you can run a local kubernetes via [Docker Desktop](https://docs.docker.com/desktop/kubernetes/), [minikube](https://minikube.sigs.k8s.io/docs/start/), etc.  We'll use [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) to interact with the Kubernetes cluster in this setup, which should work with any cluster engine.

### VPN
When running `virtual-kubelet` locally, you'll need to provide a way for it to connect to EC2 instances via their private IP address (for the VKVMAgent / gRPC connection).  You can setup a client VPN using these docs:

- ACM (server and client certs) [Authentication - AWS Client VPN](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/client-authentication.html#mutual)
- [OpenVPN Client Connect For Mac OS | OpenVPN](https://openvpn.net/client-connect-vpn-for-mac-os/) (or Shimo)

### Enable ec2 instance connect via web console
- check `https://ip-ranges.amazonaws.com/ip-ranges.json` for the IP ranges for your region (and `EC2_INSTANCE_CONNECT` service)

---
> © 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.
