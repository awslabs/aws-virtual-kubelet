# Backlog Fodder
This file captures ideas and other potential backlog items.  These items are captured here rather than GitHub issues to ensure that the issue backlog represents only committed (or to-be-triaged) work.  Each entry is a **Title** followed by a **Description* (these map directly to GitHub Issue fields for easy import).

#### Remove need for ManagementSubnet
Instead of creating dummy ENI to support virtual nodes in AWS CloudProvider, Kubernetes node controller should have alternative approach to check the existence of Virtual node when it stops posting health status.

#### Add custom pod readiness gates
Use [Pod readiness](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-readiness-gate) gates (custom conditions) to track and report EC2 state to Kubernetes. Could also report VKVMA connectivity or other relevant status outside that which VKVMA manages here.

#### Ensure cleanup of resources in failure cases
Use `defer()` calls and/or other mechanisms (like [Finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)) to ensure cleanup of resources in failure cases.

#### Add tracing capability
- Add tracing capability.  See https://github.com/virtual-kubelet/azure-aci/blob/master/provider/aci.go#L1155-L1156 for an example.

#### Remove direct Kubernetes communication
The [virtual-kubelet](https://github.com/virtual-kubelet/virtual-kubelet) library indicates we should not be talking directly to k8s[^1].  We will likely need add functionality to the upstream library to support this requirement.  Once this is complete, we can remove `KubeConfigPath` and related handling, along with the `k8sclient` altogether.

#### Use contexts more effectively
Can we share a context with all entities involved in a "pod" then use cancellation of that context to trigger cleanup of all associated resources (e.g. delete app, then drop ec2, etc.)?

#### Use full EC2 Instance objects as arguments and return values
There are places in the code like this:
```go
return *resp.Instances[0].InstanceId, err
```

and

```go
createCompute...return instanceID, privateIP, nil
```
Instead of extracting a subset of the Instance object, just return the entire thing (i.e. `return *resp.Instances[0]`).  This will allow receivers of the arguments to determine what portion they need to access.  Since this object is not passed over the network, object/struct size is generally less of a concern (in this case it's a single instance also).

#### Periodically log overall VK status
Create goroutine loop to periodically log status such as vk version, os, arch, which node it runs on, how many pods it's managing, warm pool status, etc.

#### Migrate pod notifications to use utils package version
Update references to the provider's pod notifier member to use the one from `utils` instead.  That notifier is set automatically in the provider's `NotifyPods` function (by virtual kubelet) and should be used from then on.

Remove the following line to quickly find other code that needs to be updated:
https://github.com/aws/aws-virtual-kubelet/blob/ed2bc780457f4a32c92bc8815c8678fc2b02ef41/internal/ec2provider/ec2provider.go#L282

#### Use actual Service Quota values for resource capacity
Obtain default values for CPU, Memory, and Storage from some account attribute or quota value possibly.  Decrement resources on consumption and generally adhere to Service Quotas reported capacity for actual VK capacity metrics.  See https://docs.aws.amazon.com/servicequotas/2019-06-24/apireference/API_GetServiceQuota.html for details.

#### Use custom errors where appropriate
The following code was unused, so it was removed, but having custom errors could be useful and help minimize duplication:
```go
package ec2provider

type PodErrorType int

const (
	GeneralError PodErrorType = iota + 1
	Ec2Error
	GrpcError
)

type PodError struct {
	errorType PodErrorType
	message   string
}
```

#### Implement "flapping" detection in monitoring
Detect "flapping" between healthy/unhealthy states by tracking transitions.

#### Add / detect debug flag and additional goroutine labeling for easier debugging
```go
	// TODO: could use this version conditionally based on presence of debug flag/log-level
	// NOTE example that applies labels to the goroutine for easier debugging
	//labels := pprof.Labels("pod", pm.pod.Name, "namespace", pm.pod.Namespace)
	//go pprof.Do(context.Background(), labels, func(ctx context.Context) { pm.monitorLoop(ctx, done) })
```

#### Handle case where health check can be UNKNOWN forever
```go
	// TODO: add Unknowns count and additional threshold to HealthConfig, then increment failure when Unknowns
	//  threshold is reached (and reset Unknowns count)?  Otherwise state could be unknown forever and never be
	//  marked unhealthy
```

#### Simplify health check type handling
```go
	// TODO: just have a type in monitor and have it be monitor.CheckerType or monitor.WatcherType
	m.isWatcher = true
```

## Tests
#### Add parameters for `launch_pods.go` number of pods to generate and template path

## VKVMAgent
Items here relate to the agent example(s) provided.

#### Add an endpoint to allow state change simulation?
This would need to be added to the `.proto` file for proper client generation.  The endpoint would allow remote manipulation of state to simulate actual agent behavior.  This functionality would be useful as an example to users and also in automated testing.

#### Add random pod status change function
A "happy pod" function exists to generate and return a healthy pod status.  This item is to add functions capable of producing randomized pod status elements.  Here's an example for a random pod phase:
```go

// TODO(guicejg): implement randomConditions, randomContainerStatuses, ... "variety is the spice of life..."
func randomPodPhase() corev1.PodPhase {
	phases := []corev1.PodPhase{
		corev1.PodPending,
		corev1.PodRunning,
		corev1.PodSucceeded,
		corev1.PodFailed,
		corev1.PodUnknown,
	}

	return phases[rand.Intn(len(phases))]
}
```

A `randomConditions`, `randomContainerStatuses`, etc. could be added to generate interesting output for testing or other uses.

#### Add option to choose listen interfaces
By default, the example VKVMAgent listens on all network interfaces.  Provide a flag to limit this to a particular interface.

## CDK
These items relate to the CDK example infrastructure code.

#### Set ECR repository retention
It is currently using some default.  Set this value explicitly instead and document the setting.

#### Add pipeline related resources and CDK Context flag to control activation
The following code was removed from the original pipeline-oriented version of the CDK IaC:

```go
pipelineActionsUser := createPipelineIAMAccount(stack)

// grant bucket read/write to the github IAM user
s3Bucket.GrantReadWrite(pipelineActionsUser, "*")  // may want to limit to a specific S3 key also

func createPipelineIAMAccount(stack awscdk.Stack) awsiam.User {
	// NOTE This capability is currently unused, but left here as an example
	// add user for pipeline actions activity
	pipelineActionsUser := awsiam.NewUser(stack, jsii.String("PipelineActionsUser"), &awsiam.UserProps{
		Path:     jsii.String("/managed/"),
		UserName: jsii.String("pipeline-user"),
	})

	// create API keys for github user
	accessKey := awsiam.NewAccessKey(stack, jsii.String("AccessKey"), &awsiam.AccessKeyProps{
		User:   pipelineActionsUser,
		Serial: jsii.Number(1), // increment this to rotate credentials
	})

	// create a JSON key/value structure for the Secrets Manager value
	secretValue := map[string]*string{
		"AWS_ACCESS_KEY_ID":     accessKey.AccessKeyId(),
		"AWS_SECRET_ACCESS_KEY": accessKey.SecretAccessKey().ToString(),
	}
	secretValueJson, err := json.Marshal(secretValue)
	if err != nil {
		panic(err) // TODO add error message
	}

	// store the access/secret key in Secrets Manager
	awssecretsmanager.NewSecret(stack, jsii.String("S3UserCredentials"), &awssecretsmanager.SecretProps{
		Description:       jsii.String("User for Pipeline Actions workflows"),
		SecretName:        jsii.String("/managed/pipeline-user"),
		SecretStringBeta1: awssecretsmanager.SecretStringValueBeta1_FromToken(jsii.String(string(secretValueJson))),
	})
	return pipelineActionsUser
```

This code creates an IAM user for a CI/CD system (e.g. GitHub) and exposes the access keys via AWS Secrets Manager.  The activation of this code should be controlled by a [CDK Context](https://docs.aws.amazon.com/cdk/v2/guide/context.html) flag that allows additional CI/CD pipeline related resources and actions to be enabled.

[^1]: See https://github.com/virtual-kubelet/virtual-kubelet/#providers point #3
---
>© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.
