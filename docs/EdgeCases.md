# Edge Cases
This document captures edge-cases that require validation or remediation.

## VK Termination
The `virtual-kubelet` (VK) instances can terminate _at any time_, possibly leaving resources in mid-flight.

### EC2
If VK fails right after `RunInstances` but before a result is returned and stored, the instance can become disassociated with the pod.  `GetCompute` _attempts_ to find existing instances, but currently uses a pod annotation to know what to look for.  In some cases this pod annotation will not be set yet.

- [ ] A possible solution is to set one or more tags _during_ instance launch (via `TagSpecification`) that associate the instance with a pod (or warm pool).  Then `GetCompute` can check instance tags to locate an existing instance and re-associate it with the pod.
---
>© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.