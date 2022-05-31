# Configuration
The provided example `config-map.yaml` is used to populate the ConfigMap read by VK pods running in a Kubernetes cluster.  The `config.json` file is used for running VK outside a cluster for dev/testing (and other non-standard use-cases).  Both files have the same configuration content and parameters, which are detailed below.

## General
<dl>
<dt>ManagementSubnet</dt>
<dd>Subnet in which you expect to deploy the Virtual Kubelet, which generates an AWS ENI for the purposes of creating a unique location for the Kubernetes IP address.</dd>
<dt>ClusterName</dt>
<dd>Included for tagging purposes to manage AWS ENIs associated with Virtual Kubelet.</dd>
<dt>Region</dt>
<dd>Code for AWS Region the Virtual Kubelet will be deployed to. e.g. "us-west-c2" or "us-east-1".</dd>
</dl>

## VMConfig
<dl>
<dt>InitialSecurityGroups</dt>
<dd>AWS SecurityGroups assigned to an EC2 instance at launch time, which can be updated later.</dd>
<dt>DefaultAMI</dt>
<dd>AMI used when there is no other AMI specified in Podspec of a Kubernetes Pod.</dd>
<dt>InitData</dt>
<dd>Base64 encoded JSON to be processed by the Bootstrap Agent.</dd>
</dl>

## BootstrapAgent
<dl>
<dt>S3Bucket</dt>
<dd>Bucket location in S3 where bootstrap agent is located.</dd>
<dt>S3Key</dt>
<dd>Key location in S3 where bootstrap agent is located.</dd>
<dt>GRPCPort</dt>
<dd>Port number for GRPC communication between Virtual Kubelet and the EC2 instances it creates.</dd>
<dt>InitData</dt>
<dd>Base64 encoded JSON to be processed by the Bootstrap Agent.</dd>
</dl>

## WarmPoolConfig [OPTIONAL]
If included, a "warm pool" of pre-launched EC2 instances will be created and used when compute is called for.
<dl>
<dt>DesiredCount</dt>
<dd>Amount of EC2 to be maintained in the WarmPool, above and beyond what is required to run Kubernetes Pods.</dd>
<dt>IamInstanceProfile</dt>
<dd>The IAM instance profile assigned to the EC2 at launch time, which can be changed at Pod assignment time.</dd>
<dt>SecurityGroups</dt>
<dd>The AWS Security Groups assigned to the EC2 at launch time, which can be changed at Pod assignment time.</dd>
<dt>KeyPair</dt>
<dd>The EC2 credentials assigned to allow for SSH/RDP access to the instance. Unchangeable at Pod assignment time.</dd>
<dt>ImageID</dt>
<dd>The AWS AMI to launch the EC2 instances with, Unchangeable at Pod assignment time.</dd>
<dt>InstanceType</dt>
<dd>The AWS EC2 InstanceType, e.g. `mac1.metal`. Unchangeable at Pod assignment time.</dd>
<dt>Subnets</dt>
<dd>The AWS VPC Subnet(s) to deploy the WarmPool EC2 instances into. _Not_ changeable at Pod assignment time.</dd>
</dl>

# Other
See [config.go](../internal/config/config.go) for additional configuration items and their defaults.
