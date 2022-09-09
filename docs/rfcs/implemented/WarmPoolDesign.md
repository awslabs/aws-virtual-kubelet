# Warmpool support for Virtual Kubelet

## What is the problem?

When deploying Pods using AWS Virtual Kubelet for MacOS containers, the boot time of the mac1.metal EC2 instances that are dependencies are too long to provide good user experience for those utilizing the Virtual Kubelet.
 
## What will be changed?
 
A new optional feature named WarmPool will be implemented. This feature will allocate EC2 ahead of when it is needed for a Kubernetes Pod. This will implement a new provider type for Virtual Kubelet, allowing for previous functionality to exist if no WarmPool is desired for cost considerations.
 
## Why is it changing?
 
The current Virtual Kubelet is unable to pre-provision EC2 and only creates EC2 on an as-needed basis. This change seeks to decouple sourcing EC2 instances and creating Kubernetes Pods.
 
## Success criteria for the change
 
A deployed Virtual Kubelet can maintain one or many pools of unused EC2 for future Kubernetes Pods, in addition to existing functionality. The Virtual Kubelet will maintain an ideal set of unused EC2 based on Administrator configurations of the Virtual Kubelet. Per Chime, the following behaviors are minimum viable:
1. Have multiple subnets per Warmpool, choose one randomly when creating a Warmpool EC2 instance
2. Have support for multiple Warmpool per Virtual Kubelet and pod spec need to specify the pool to use when we support multiple Warmpool.
 
## Out of scope
 
1. Virtual Kubelet is not responsible for managing each of the following, as it is AWS account setup:
 
    * Mac1.Metal Dedicated Host capacity. If there is no capacity available, then the Virtual Kubelet has no access to required resources to maintain its WarmPool
    * Service Quotas for Mac1.Metal Dedicated Hosts or EC2 in general. The Virtual Kubelet will only attempt to create and maintain what it is configuration is set to do, but if it is asked to do more than the AWS account can support from a limits perspective, it is not the responsibility of the Virtual Kubelet to remediate it.
    * IP allocation and management in a given Subnet.
 
## Design
 
WarmPool will define the ability to create and terminate EC2 instances that are not yet running Kubernetes Pods, but are candidates to be targets for payloads in Kubernetes CreatePod requests being fulfilled by Virtual Kubelet.
 
WarmPool will be an optional feature of Virtual Kubelet, allowing consumers to opt-out of the additional costs of maintaining a WarmPool if they are willing to incur the wait time for awaiting an EC2 instance to be created, booted, and configured. WarmPool will maintain a “DesiredCount” of EC2 that are ready to accept Kubernetes CreatePod requests. This will either provision additional EC2 if the amount of running EC2 in the WarmPool is below the DesiredCount, or terminate EC2 instances if the amount of running EC2 is in excess of the DesiredCount. WarmPool will check on a time basis to determine if actions are needed to maintain the DesiredCount.
 
WarmPool EC2 properties will be the default of RunInstances API calls, unless overwritten explicitly by a Pod’s Annotations for properties like “instance-type” or “image-id”. WarmPool will send CreatePod and DeletePod requests to EC2 via gRPC protocol. Pod state will be maintained via Pod annotation.
 
WarmPool depth will be computed based upon Tagging on EC2 Key:Value example “WarmPool”:”[Ready | InUse | Pending* | Unknown] WarmPool State definitions:
Ready: The EC2 instance is in status “Running” and the Bootstrap Agent is reporting “Healthy” status, awaiting a CreatePod request.
InUse: The EC2 instance is in status “Running” and there is an active Pod scheduled to this EC2 instance.
Pending*: Describes an EC2 that’s undergoing a state transition. E.G. EC2 instance is involved with an active CreatePod or DeletePod request that is not yet complete. Causes for Pending state will be postpended, e.g. “PENDING_WARMPOOL_PROVISIONING”
Unhealthy: If any of the above States are not fulfilled, the EC2 state will be assigned to Unhealthy for manual inspection. If an instance stays in an Unhealthy state for enough periods of HealthChecks, then the EC2 will be terminated. The primary method of triggering the Unhealthy state will be instance becomes unreachable or the EC2 API is unreachable during routine healthchecks
 
### Managing WarmPool Status:
 
The Create EC2 functionality will assign a configured Security Group and Port to attempt contact for gRPC based connectivity with a default Tag that applies WarmPool status “PENDING_WARMPOOL_PROVISIONING”. An asynchronous process of Virtual Kubelet will poll the EC2 until the HealthCheck of the EC2 instance returns true, then the EC2 WarmPool Tag will be set to “Ready” and push the EC2 instance information into a queue that holds all “ready” instances.
 
There will be 4 queues which contain a list of instances in various statuses:
 
 
* Provisioning
* Ready
* Allocating Pod
* Unhealthy
 
These queues will track information about EC2 instance IDs as they move through WarmPool states in memory on a Virtual Kubelet, and reflect tracking information for Virtual Kubelet as reflected by EC2 instance Tags above.
 
When a CreatePod request is “Selecting EC2 from WarmPool”, the Virtual Kubelet will pop a value from the EC2 Ready queue, validate that the instance is still in “Ready” status and Healthy, and set status to “PENDING_POD_PROVISIONING” as well as initiate sending the PodSpec to the EC2 to become a Pod. The EC2 instance will additionally be tagged Kubernetes Pod Name, Namespace, and UUID to assist with operational identification and reconciliation during Virtual Kubelet restarts. Once the Pod reports healthy, the Tag is then updated to “InUse”. If a Pod does not report healthy, instead move to Delete Pod and therefore terminate instance.
 
### Managing Pods:
 
 
Creating a Pod from a WarmPool enabled Virtual Kubelet sends a `LaunchApplication `message via gRPC. A flow is included below for the control logic. Once a Pod is assigned to an EC2 instance, the instance is no longer part of the WarmPool.
[Image: WarmPool-Create Pod Flow.png]
Deleting a Pod from a WarmPool enabled Virtual Kubelet sends a `TerminateApplication `message via gRPC. This behavior is no different than WarmPool disabled Virtual Kubelet.
 
Periodically, WarmPool will take stock of itself such that it will achieve a healthy amount of EC2 instances ready for use. A sample flow of control logic is given below. Depth reflects the total amount of EC2 that are currently not assigned pods. DesiredCount reflects the configured amount which the WarmPool is specified the ideal ready amount of EC2 instances. The amount reflected by DesiredCount may not be achievable due to outside requirements, such as AWS Quota limits for EC2 capacity, Mac1Metal Dedicated Host capacity within the Availability Zone, exhausting the IP range of a subnet, or similar outside factors.
[Image: WarmPool-CheckPoolDepthFlow-CheckPoolDepthFlow.png]
Updating a Pod is not a valid case for Virtual Kubelet, so this functionality will not be executed in WarmPool either.
 
### Considerations
 
* How does it impact customers who already have a Virtual Kubelet deployed? This feature will require a deployment of the updated version of Virtual Kubelet to an existing Kubernetes Cluster. It will also require a Config update to enable WarmPool. 
* Does this change introduce scale or availability inversion? (Making a high scale or availability component depend on a low scale or availability component). No. This will provide early detection of lack of EC2 capacity.
* Does your change require storing long-lived state (i.e. state that doesn't go away after the request is over). How will you ensure that state is not leaked? State will be managed in-memory for what EC2 are Ready to be utilized for a CreatePod request. If this state is lost, Virtual Kubelet will attempt to scan existing EC2 instances for “Ready” state EC2 when creating its WarmPool during Virtual Kubelet pod startup for reuse.
* How can you deliver the feature incrementally? We can begin with the Toggle of WarmPool based on Config Map information, then test the control flow. Then we can test Creating and Maintaining the WarmPool. Then we can test taking an EC2 from WarmPool to be the compute source for Create Pod workflow instead of creating a net-new EC2 instance if WarmPool is enabled.
* What happens if a Virtual Kubelet Pod terminates by accident or on purpose? Depending on Pod configuration, Virtual Kubelet Pod may or may not attempt to restart. This will be determined by the `Kind` of deployment of the Pod, as well as the `NodeSelector `value being placed. If a restart is not attempted, then the Pods will be lost to the Kubernetes system and eventually collected by the Pod Garbage Collector. The underlying EC2 will continue to run with no further management by Virtual Kubelet, and will need an outside process to either restart Virtual Kubelet appropriately or terminate the EC2. If a restart is attempted, the Virtual Kubelet with WarmPool enabled will attempt to scan existing EC2 with “`Ready`” tag and re-use them if they fit a WarmPool configuration upon Virtual Kubelet startup. If the EC2 is not compatible with a WarmPool, it is left orphaned. Currently, there is no graceful termination behavior which would re-assign or terminate EC2 during a Virtual Kubelet shutdown.
 
## Customer Experience
 
### **ConfigMap Changes**
 
This change will introduce a new top-level key to the ConfigMap, which I propose as WarmPoolConfig. The WarmPoolConfig will contain a list, which each entry has the following fields:
**Default -** This indicates that if a Pod has insufficient annotations to fulfill an EC2 RunInstances API call, to use this WarmPoolConfig instead. Exactly 1 must be specified in the list of WarmPoolConfigs. If not specified, default to `false
`**DesiredCount** - Describes how many instances to keep ready for assignment to a Pod
**KeyPair** - Describes the KeyPair associated with the EC2 instances at creation.
**ImageID** - The AMI ID of the Amazon Machine Image to be used for Pod Creation.
**InstanceType** - The EC2 InstanceType to create the WarmPool from. If not specified, default to `mac1.metal` instance type.
**Subnets** - A list of VPC subnets to create a WarmPool from. If not specified, default to `ManagementSubnet `Virtual Kubelet configuration.

**`TODO`** This sample Warm Pool configuration does not work (required values `Subnets`, `IamInstanceProfile`[undocumented] are missing)...also the combination of configurations below causes WP to constantly create and terminate instances (but ultimately results in a new positive number of instances)

### Sample Configuration Block
This block goes inside the top-level `{}`'s in the `config.json` or equivalent file.
```json
  "WarmPoolConfig": [
    {
      "DesiredCount": "2",
      "Default": true,
      "KeyPair": "MyPreprovisionedKeypair",
      "ImageID": "ami-1287634276",
      "InstanceType": "mac1.metal",
      "Subnets": [
        "subnet-1234asdf",
        "subnet-asdf1234",
        "subnet-qwerty"
      ]
    },
    {
      "DesiredCount": "5",
      "Default": false,
      "KeyPair": "MyOtherKeypair",
      "ImageID": "ami-9876789679",
      "InstanceType": "m5.large"
    },
    {
      "DesiredCount": "3",
      "KeyPair": "MyOtherKeypair",
      "ImageID": "ami-9876789679",
      "Subnets": [
        "subnet-1234asdf",
        "subnet-asdf1234",
        "subnet-qwerty"
      ]
    },
    {
      "DesiredCount": "1",
      "KeyPair": "MyPreprovisionedKeypair",
      "ImageID": "ami-1287634276"
    }
  ]
``` 

 
 
### **Latency impact**
 
WarmPool will reduce the time between a CreatePod request and a container successfully running by approximately 5-8 minutes.
 
### **Financial impact**
 
WarmPool runs more EC2 than is needed to support the workloads that are currently provisioned to it, and that cost is flexible based upon Configurations set by the Virtual Kubelet administrator. See EC2 Pricing to estimate costs for the configuration that is proposed.
 
## Notable failure modes
 
For example, what happens when ...
 
* an ec2 instance fails? The instance will no longer be reporting as “Ready” and removed from the pool as a result. If this reduces the count of EC2 to below WarmPool configuration requirements, this will automatically trigger CreateEC2 workflow to add to the pool. 
* Insufficient Dedicated Hosts are provisioned for mac1.metal instances? The Virtual Kubelet does no management of Dedicated Instances and therefore will continue to error on EC2 instance creation until Dedicated Host capacity becomes available, either by deleting other pods and releasing EC2 instances from consumed Dedicated Host capacity or some other process creating Dedicated Hosts.
* an AZ goes down? No changes occur, and if EC2 instances are lost during the AZ failure then the Virtual Kubelet will detect that the WarmPool does not meet the DesiredCount and therefore attempts to rebuild a healthy pool.
* an invalid configuration is provided? The Virtual Kubelet will terminate while loading configurations.
 
## Testing
 
* What specific areas of this change are susceptible to failure?
    * Control logic between a “WarmPool enabled” and “WarmPool disabled” option sets.
    * Control logic for Pod specific overrides over WarmPool configurations.
* How will this change be tested?
    * Create a Virtual Kubelet Pod with needed configurations to enable WarmPool
        * Track in the AWS EC2 Console that EC2 are being created for use according to DesiredCount configurations
        * Submit a valid Create Pod request to the WarmPool, track that an EC2 is being repurposed and tags updated. Subsequently, Terminate the pod
        * Submit an invalid Create Pod request to the WarmPool, track that the EC2 is attempted to be repurposed but eventually Terminated. Subsequently, Terminate the EC2.
    * Create a Virtual Kubelet Pod without configurations for WarmPool
        * Track that no additional EC2 are created outside of in direct response to CreatePod calls to Virtual Kubelet.
* Will manual testing be required? If so, why?
    * Yes, the decision to add CI/CD has been actively deferred on this project. We have no mechanism to do any kind of integration testing in an automated way.
 
## Logging
 
Logging is to include
 
* Non-transient Errors, particularly in aquiring new EC2 instances or polling their health.
* Retries
* System progress:
    * Creating new EC2 for WarmPool
    * Assigning EC2 from WarmPool to a Pod
    * Deleting EC2 from WarmPool
    * Polling State of EC2 in WarmPool
 
## Deployment
 
* How do deploy as safely as possible? (Can we deploy incrementally?) A new version of Virtual Kubelet Pods can be deployed from a new Docker container image and require no-operation changes for customers who do not want to use WarmPool.
* Will this be safe when the fleet is in a heterogeneous state where some hosts are on the new revision and some are on the old? Yes, as old behavior will continue to be supported from ConfigMap and only new behaviors would occur.
* Is this change backwards compatible? The WarmPool feature is not backwards compatible, but existing functionalities would be.
* How do we rollback? How can we make rollback as quick as possible? Rollback would occur by deploying a previous revision of Virtual Kubelet. This can be done within 5 minutes of the operation being ordered.

---
>© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.