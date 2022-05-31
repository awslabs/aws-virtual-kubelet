# Edge Cases
This section describes some edge-cases and scenarios where behavior is not ideal or is unknown.

## Scaling Provider replicas down
It has been observed that when pods are running across multiple VK instances/nodes, scaling the number of VK instances down doesn't reschedule those pods onto other remaining provider instances.  For instance if 99 pods are running across 3 provider instances in a StatefulSet (33 pods per provider), and the provider instances are scaled from 3 to 2, the 33 pods it was responsible for will "wait" for the provider instance to return (a sort of "stickiness" of pods to provider instances).  If the provider instance doesn't return after a time, the pods will be evicted and any EC2 instances associated with the pods will become "orphaned".

To mitigate this undesirable behavior, operators should [drain](https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/) the VK nodes that will be terminated during scale-down _first_.  This will cause Kubernetes to evict the workload pods from the VK instances and re-launch them on other active nodes.  Once the pods are no longer running on the drained nodes, the scale-down can be initiated which will terminate the now empty nodes/instances. 

See the [Operation section of the Cookbook](Cookbook.md#operation) for a summary of the procedure described above.

## VK Terminate during EC2 launch
If VK fails right after `RunInstances` but before a result is returned and stored, the instance can become disassociated with the pod.  `GetCompute` _attempts_ to find existing instances, but currently uses a pod annotation to know what to look for.  In some cases this pod annotation will not be set yet.

The [Pod Persistence](rfcs/PodPersistence.md) RFC proposes a mitigation.