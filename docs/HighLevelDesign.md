# High-Level Design
This document describes the overall design of the system, including the different components and purpose(s) they serve.  Scenarios describing various execution flows and how the system behaves are also covered.

# Layers
This section describes the layers of the system and their general responsibilities.

## Kubernetes
The top-most layer, and the one that users interact with most frequently, is Kubernetes itself.  When running the Virtual Kubelet EC2 provider, or workloads that utilizes the provider's functionality, Kubernetes provides that API and interface layer.

## Virtual Kubelet
The Virtual Kubelet framework is an upstream library that translates between the Kubernetes API and the PodLifecycleHandler (and other interfaces) implemented by this provider.  The main interfaces to implement are dictated by the Virtual Kubelet framework itself. This library subscribes to relevant events for providers it manages via Informers.  When it receives such an event, it calls the appropriate provider function via the agreed interface.

## Virtual Kubelet EC2 Provider
This provider implements the actual handling of requests relayed by Virtual Kubelet for things like creating pods, getting pod status, and terminating pods.  In general, the provider code does not interact with Kubernetes directly [^1].  The provider is responsible not only for determining what constitutes a "pod", but also for managing its lifecycle and tracking (caching) pods it is managing.

The provider also owns creation of a "node" which represents a VM, fargate task, or other similar resource pool.  In this implementation a node is simply an Elastic Network Interface which provides a required name component (private IP address) and enables EKS to "see" the node as a usable measure of compute.  The provider manages a tag on the ENI so each instance of the provider can re-associate with the ENI on restarts.

This provider itself runs in the traditional container-based portion of Kubernetes.  It is launched as a [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) to provide consistent pod names, which are used in the aforementioned ENI tags.

# General Behavior
This section covers behavior of the system in different scenarios, as well as to-be-determined cases. The EC2 provider receives pod creation, status, termination, etc. requests from the Virtual Kubelet and takes appropriate action based on the request.  Some common scenarios are described below.

## Startup
When a virtual kubelet provider instance starts, it first attempts to find an ENI node with an expected tag.  Failing that it will create and tag an ENI.  In both cases, the ENI private IP becomes part of the node name presented to Kubernetes.

Since the provider currently relies on Kubernetes to cache pods when provider instances are restarted, it also asks Kubernetes if it's aware of any pods that belong on the provider's node.  If the list of pods is greater than zero, the provider will repopulate its cache and restart PodMonitors on the pods returned.

This happens _before_ the provider is fully instantiated, because as soon as the provider is "ready" (as determined by the Virtual Kubelet library's startup process), Kubernetes will ask the provider for the list of pods it is managing.

This behavior can be problematic for "bare" pods (those without a Deployment, ReplicaSet, etc. abstraction).  Pods without a level of management above will be "cleaned" from Kubernetes after some time if they don't respond to a request for status.  This means that if the provider instances are shut down for very long, then on restart when the provder asks Kubernetes for the list of pods, Kubernetes may reply that there aren't any and resources utilized by the pods become orphaned (e.g. EC2 instances).[^2]

## CreatePod
When a pod creation request is received, the provider will obtain an appropriate EC2 instance, then ask the VKVMAgent to launch its application.  If Warm Pool is configured, the instance may already be running and will just get reconfigured to participate in the pod.  If not, or if no appropriate instance exists, then one will be launched.  After that point the behavior for both cases is the same (even through termination of the pod).

The steps leading up to (and including) application launch are configured with retries and timeouts.  An attempt has been made to keep the startup behavior consistent with later behavior when connections are lost, degraded, or resources become unhealthy.  There are likely some gaps here still though and tests should be developed to exercise these scenarios.

The final step after launching the application is to start a PodMonitor to both monitor the health of pod resources and to report status back to Kubernetes from the VKVMAgent.  A PodMonitor is a collection of monitors that cover the resources that make up a pod.  Some of those monitors may be polling (check) type, while others may be streaming (watch) types that block until health/status messages are received from the VKVMAgent.  Both of these monitor types run in a separate goroutine to avoid impacting the main program flow (or each other).

## DeletePod
When a pod deletion request is received, the steps in the create flow are generally performed in reverse.  PodMonitoring is stopped, the application is terminated, the EC2 instance is terminated (regardless of whether it was a Warm Pool instance or freshly created), and finally Kubernetes is notified that all containers and the pod itself are stopped/terminated.

## GetPod, UpdatePod, etc.
Other PodLifecycle interface functions handle returning pod status and making updates to pods as-requested.

# Edge Cases
This section describes some edge-cases and scenarios where behavior is not ideal or is unknown.

## Scaling Provider replicas down
It has been observed that when pods are running across multiple provider instances, scaling the number of providers down doesn't reschedule those pods onto other remaining provider instances.  For instance if 99 pods are running across 3 provider instances in a StatefulSet (33 pods per provider), and the provider instances are scaled from 3 to 2, the 33 pods it was responsible for appeared to "wait" for the provider instance to return (a sort of "stickiness" of pods to provider instances).  This _may_ have been a configuration-related behavior in the particular cluster it was observed in and the experiment should be repeated.

If the behavior is reproducible, it's likely that scaling providers up won't "re-balance" the pods either.  This case should be tested as well to verify though.

## Unrecoverable failures
If a fatal error happens during handling of a PodLifecycle function call, the provider can return an error to Kubernetes (via the Virtual Kubelet library) which will trigger Kubernetes standard retry/failure behavior.  If, however, an unresolvable exception happens during normal operation then the provider can only update Pod (or Node) status and trigger the PodNotifier[^3].

In the case where a pod cannot be launched successfully, the provider will destroy affected resources and attempt to retry indefinitely.  This may not be desirable though, in which case a maximum number of attempts could be added to the configuration (and behavior when that max is reached given some definition).

Another interesting situation is one in which an EC2 instance is created but some problem keeps the application from starting successfully, then when the provider tries to delete the EC2 instance to retry pod creation it fails at the deletion step (i.e. if the EC2 credentials just expired).  This is another case in which the provider will retry indefinitely.

There are most likely some failure scenarios where it's possible for a resource to become untracked (orphaned) and fall out of management.  A tag-based periodic sweep, using Kubernetes finalizers, and other options have surfaced in discussions.  These ideas need definition and detail to be actionable (and ideally test-cases that demonstrate the failure mode and can be used to verify solutions).


[^1]: There are a few exceptions due to lack of functionality in the Virtual Kubelet library.
[^2]: This is similar to the situation where a user tries the provider for a bit, then terminates the cluster without first deleting pods.  A cleanup script that searches for tags has been proposed as a way to handle this.
[^3]: PodNotifier is a callback provided by the Virtual Kubelet library that notifies Kubernetes of pod status changes without waiting for Kubernetes to poll for status.