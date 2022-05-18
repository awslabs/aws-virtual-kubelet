# Virtual Kubelet Stack
This a CDK Infrastructure as Code (IaC) project that creates environments (dev, automated pipelines, etc.).

- [ ] insert CDK boilerlate
 
## Steps
`go mod tidy` to download/update any missing dependencies
`go test` to run unit tests

### Deploy infrastructure to `us-west-2`
AWS_DEFAULT_REGION=us-west-2 cdk deploy

### Post-CDK Steps
#### Create `kubeconfig` entry
Navigate to the CloudFormation stack outputs and copy the `ClusterConfigCommand`.  This generated command will setup `kubectl` access by modifying your `kubeconfig` with the appropriate roles, etc.

**`NOTE`** You will need a working AWS CLI setup to run this command.

After running this command you should be able to use `kubectl` commands.  e.g.

```
kubectl get svc
NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   172.20.0.1   <none>        443/TCP   4h8m
```

#### Configure ECR Repository in Makefile
Update the variables at the top to match the CDK generated ECR repository.  You can find the info by looking at the resources generated in the CloudFormation stack.  There will be a link to the ECR repo there.

#### Build and push Docker image
Set `PLATFORM` in the Dockerfile and Makefile if not targeting `linux/amd64`
`make docker`
`make push`

#### Deploy Configuration via ConfigMap
Modify `configMapTemplate.yaml` using values from CloudFormation stack output
- set `ManagementSubnet` to PrivateSubnet1's Subnet ID

`kubectl apply -f examples/configMap.yaml`

#### Deploy Service Account, Cluster Role, and Binding
These permissions allow the virtual kubelet provider to interact with Kubernetes cluster objects
`kubectl apply -f deploy/my_vk_clusterrole_binding.yaml`

#### Deploy virtual-kubelet Stateful Set
`kubectl apply -f deploy/sample_vk_statefulset.yaml`

- [ ] determine unique values needed per-deployment and add as outputs in CDK to simplify setup
  - private subnets
  - public subnets
  - default security group

#### Deploy example pod
kubectl apply -f examples/pods/sample-pod.yaml

### Optional

#### Install Prometheus for Metrics
https://docs.aws.amazon.com/eks/latest/userguide/prometheus.html#deploy-prometheus
