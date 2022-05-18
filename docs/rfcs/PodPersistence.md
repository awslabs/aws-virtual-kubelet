# Pod Persistence
Some of the [Edge Cases](../EdgeCases.md) that aren't handled well are a result of relying on Kubernetes for workload pod _persistence)_ when a `virtual-kubelet` pod restarts.

This document describes the current state and proposes a solution that does not require orchestration of additional AWS services.

## Current State
In the current design, the list of workload pods is present in 2 places:  The Kubernetes control plane and the VK provider pod's in-memory Pod Cache.  When a VK pod is (re)started, it queries the k8s API _directly_ [^1] during startup to populate its list (cache) of workload pods.  This happens _before_ the VK provider is fully instantiated (at which point k8s asks the provider what pods it is managing).

[This section of code](https://github.com/awslabs/aws-virtual-kubelet/blob/706a7b10d050484c969dcda233c84753557d793b/cmd/virtual-kubelet/main.go#L106-L113) in `main.go` is responsible for populating that initial cache before k8s has a chance to ask for the data right back).

## Issues
As a virtual kubelet provider, this EC2-backed implementation is expected to keep track of pods it is responsible for _independently_ of k8s own tracking.

By relying on k8s instead of an independent data store, we risk losing track of pods (and their associated resources) entirely.  This happens when we launch a "bare pod" [^2] and the VK provider is down long enough for k8s to evict/forget the pod (currently 5 minutes).

When the VK provider comes back up it asks k8s for the pods that belong to the VK provider's node.  k8s replies there are none and the VK provider loses track of any associated EC2 instances at this point.

## Solution Options
This section describes solution options and their pros/cons.

### Ignore the problem
In a large deployment (or over time), losing track of resources could cause significant waste by paying for unused EC2 instances.

This is not a viable option.

#### Pros
- Easy
#### Cons
- Wasteful
- Potentially Costly

### Use DynamoDB or similar AWS service
This option would externalize the cache entirely onto a currently unused AWS service.  While the purpose-built DynamoDB database is an appropriate choice in the general sense, it adds significant complexity to the current design and resource management requirements.

This option is viable, but not desirable.

#### Pros
- Highly scalable (e.g. [Global Tables](https://aws.amazon.com/dynamodb/global-tables/))
- Highly performant (_"single-digit-millisecond latency"_)
- AWS-native

#### Cons
- Requires creation and coordination of a new resource type
- Design and implementation needed to create / manage DynamoDB-backed cache
- Complicates the architecture

### Use EC2 Tags
This option would utilize additional key/value pairs in the form of EC2 instance tags to track the instance ↔ pod relationship.  On startup the VK provider would query instances using a [server-side (API) filter](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Filtering.html) to search for a specific key/value tag pair.

This option is currently the most likely solution and as such, a detailed implementation proposal can be found below.

#### Pros
- Uses existing AWS resources already managed by the VK provider
- Key/value pair is a simple yet adequate data structure to meet the requirements
- Code to query and set EC2 tags is already present in the codebase
- Having pod ↔ instance metadata in Tags is useful for other things (e.g. resource reporting and tracking, unused resource cleanup, etc.)

#### Cons
- EC2 tag API calls may be [throttled](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/throttling.html) [^3]
- The EC2 API is generally [eventually consistent](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/query-api-troubleshooting.html#eventual-consistency) which may give rise to unexpected behavior or require additional consistency management in the implementation to mitigate

#### Implementation Proposal
As part of managing the EC2 instance resource lifecycle, the VK provider should also manage a Tag `PodUid` on each instance.

Pod names aren't guaranteed to be unique over time or across namespaces.  To prevent the case where an EC2 instance is detected as belonging to a current pod because it shares the name of a previous pod, the pod [UID](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids) should be used as the value for this tag.

In the case where an instance has been launched, but not associated with a pod (e.g. [Warm Pool](../WarmPoolDesign.md)), the Tag key _should not be present_.  A server-side EC2 API filter `NotTagged` can be used to query instances that do not have the `PodUid` tag key (i.e. are not associated with a launched pod).  Combined with other tag queries that describe the instance state with respect to Warm Pool, such a query can uniquely categorize every instance's state.

When an instance is either created or associated with a pod, the tag key `PodUid` should be set to the value of the actual pod UID.

When a pod is terminated, the instance should be terminated without altering any tags.  Should the termination fail then the list of known active pod UIDs can be compared with those found on instance tags to identify EC2 instances that the VK provider is no longer aware of (and take appropriate action).

[^1]: This is one of the places we currently violate the [virtual kubelet library supported provider requirement](https://github.com/virtual-kubelet/virtual-kubelet/#providers) to _"not have access to the Kubernetes API Server"_.
[^2]: A single pod instance with no deployment, replica set, or other supervisory k8s construct managing it.
[^3]: DynamoDB tables may _also_ be [throttled](https://aws.amazon.com/premiumsupport/knowledge-center/dynamodb-table-throttled/) but there are more options to alleviate this than with the EC2 API