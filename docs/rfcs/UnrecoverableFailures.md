## Unrecoverable failures
If a fatal error happens during handling of a PodLifecycle function call, the provider can return an error to Kubernetes (via the Virtual Kubelet library) which will trigger Kubernetes standard retry/failure behavior.  If, however, an unresolvable exception happens during normal operation then the provider can only update Pod (or Node) status and trigger the PodNotifier[^3].

In the case where a pod cannot be launched successfully, the provider will destroy affected resources and attempt to retry indefinitely.  This may not be desirable though, in which case a maximum number of attempts could be added to the configuration (and behavior when that max is reached given some definition).

Another interesting situation is one in which an EC2 instance is created but some problem keeps the application from starting successfully, then when the provider tries to delete the EC2 instance to retry pod creation it fails at the deletion step (i.e. if the EC2 credentials just expired).  This is another case in which the provider will retry indefinitely.

There are most likely some failure scenarios where it's possible for a resource to become untracked (orphaned) and fall out of management.  A tag-based periodic sweep, using Kubernetes finalizers, and other options have surfaced in discussions.  These ideas need definition and detail to be actionable (and ideally test-cases that demonstrate the failure mode and can be used to verify solutions).

[^3]: PodNotifier is a callback provided by the Virtual Kubelet library that notifies Kubernetes of pod status changes without waiting for Kubernetes to poll for status.
