# Cookbook
This document contains a "cookbook style" reference of useful commands and procedures.  Each entry is a mini [runbook](https://wa.aws.amazon.com/wat.concept.runbook.en.html) of sorts designed to accomplish a specific task.

Entries below are divided into sections by major topic area.

## Build
### Build Virtual Kubelet and push to an ECR registry
This example assumes you have set `$MY_AWS_ACCOUNT_ID` and `$MY_AWS_REGION`.  It will generate a Linux binary for the architecture of the building system (e.g. arm64 on an Apple Silicon MacBook), use `GOARCH` to modify this behavior.

```bash
export REGISTRY_ID=$MY_AWS_ACCOUNT_ID
export REGION=$MY_AWS_REGION
export GOOS=linux
make push
```

### Build example VKVMAgent for a different OS/ARCH
To build for alternate operating systems / architectures, use the `GOOS` and `GOARCH` environment variables.

e.g. Build the example `vkvmagent` for Linux AMD64 (x86_64)
```shell
cd examples/vkvmagent
GOOS=linux GOARCH=amd64 go build ./...
file vkvmagent # check binary file type
# vkvmagent: ELF 64-bit LSB executable, x86-64
```

## Operation
### Scale VK StatefulSet replicas _down_
1. Drain VK nodes starting with the highest (however many will be lost due to the scale-down)
2. Once the pods are relaunched on nodes that will _not_ be terminated due to scale-down, scale the number of StatefulSet replicas down as-desired

### Scale VK StatefulSet replicas _up_
1. Set the updated number of VK replicas and wait for the new pods to launch
2. Restart any deployments to re-balance pods across VK nodes