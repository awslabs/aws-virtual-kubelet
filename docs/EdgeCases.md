# Edge Cases
This section describes some edge-cases and scenarios where behavior is not ideal or is unknown.

## Scaling Provider replicas down
It has been observed that when pods are running across multiple provider instances, scaling the number of providers down doesn't reschedule those pods onto other remaining provider instances.  For instance if 99 pods are running across 3 provider instances in a StatefulSet (33 pods per provider), and the provider instances are scaled from 3 to 2, the 33 pods it was responsible for appeared to "wait" for the provider instance to return (a sort of "stickiness" of pods to provider instances).  This _may_ have been a configuration-related behavior in the particular cluster it was observed in and the experiment should be repeated.

If the behavior is reproducible, it's likely that scaling providers up won't "re-balance" the pods either.  This case should be tested as well to verify though.

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