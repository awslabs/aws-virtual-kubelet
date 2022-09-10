# Virtual Kubelet Stack
This project contains Infrastructure as Code (IaC) to deploy the required infrastructure for this project. The IaC technology / language is [AWS Cloud Development Kit (CDK)](https://aws.amazon.com/cdk/) Go.

## Prerequisites
### `kubectl`
You will need the [kubectl](https://kubernetes.io/docs/tasks/tools/) tool or similar k8s client utility.  See instructions in the link for your OS.

### AWS CDK
See [Getting started with the AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html#getting_started_prerequisites) doc for setup instructions.

Once you have a working CDK installation, follow the steps below to deploy this example.

## Steps
### Setup
1. Run `go mod tidy` to download/update any missing dependencies
2. [Optional] Run `go test` to run unit tests and validate your setup

### Deploy
Example deployment to the `us-west-2` (Oregon) region.

3. `AWS_DEFAULT_REGION=us-west-2 cdk deploy`

Review and acknowledge the prompts.  After successful deployment, proceed to the post-CDK steps.

## Post-CDK Steps
### Create `kubeconfig` entry
1. Navigate to the CloudFormation stack outputs and copy the `ClusterConfigCommand`.

This generated command will setup `kubectl` access by modifying your `kubeconfig` with the appropriate roles, etc.

**`NOTE`** You will need a working AWS CLI setup to run this command.  See the CDK setup instructions for guidance.

After running this command you should be able to use `kubectl` commands.  e.g.

```
kubectl get svc
NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   172.20.0.1   <none>        443/TCP   4h8m
```

### Configure ECR Repository in Makefile
Update the variables at the top of the [Makefile](../../Makefile) or export overrides in your shell to match the CDK generated ECR repository.  You can find the info by looking at the resources generated in the CloudFormation stack.  There will be a link to the ECR repo there.

### Build and push Docker image
Set `PLATFORM` in the Dockerfile and Makefile if not targeting `linux/amd64`, then run the following to create and publish a docker image.

1. `make docker`
1. `make push`

### Deploy Configuration via ConfigMap
Copy the [config-map.yaml](../../examples/config-map.yaml) example to the `local` dir in this project's root, then modify the `configMapTemplate.yaml` file using values from CloudFormation stack output and resources:

- set `ManagementSubnet` to PrivateSubnet1's Subnet ID
- set the `DefaultAMI` (**`NOTE`** this AMI _must_ start a VKVMAgent on boot listening on the configured port)
- set remaining values as-needed

Next, apply the updated ConfigMap to your Kubernetes cluster (run this command from the project root):
`kubectl apply -f local/config-map.yaml`

### Deploy Service Account, Cluster Role, and Binding
These permissions allow the virtual kubelet provider to interact with Kubernetes cluster objects:
`kubectl apply -f deploy/vk-clusterrole_binding.yaml`

### Deploy virtual-kubelet Stateful Set
Copy the example [vk-statefulset.yaml](../../examples/vk-statefulset.yaml) to `local` and update the image reference to point to the image pushed earlier.  Next apply the StatefulSet to deploy the Virtual Kubelet ED2 provider pods:

`kubectl apply -f local/vk-statefulset.yaml`

### Deploy example pod
Now copy the example [sample-deployment.yaml](../../examples/pods/sample-deployment.yaml) to `local` and update as-needed.  At this point you can launch the example deployment and you should see EC2 instances being created.  Check the logs for the VK pods if you encounter issues.

`kubectl apply -f local/sample-deployment.yaml`

## Optional

#### Install Prometheus for Metrics
https://docs.aws.amazon.com/eks/latest/userguide/prometheus.html#deploy-prometheus
