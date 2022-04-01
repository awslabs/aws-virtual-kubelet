# Virtual Kubelet Virtual Machine Agent (VKVMA) Example

This is an example VKVMA that provides a starting point for an actual agent implementation.  The VKVMA manages the actual workloads running on EC2 instances and implements an agreed API that allows the virtual kubelet to interact in a consistent way.

## Security

When implementing a functional agent based on (or inspired by) this example, please note that the security and management of workloads must be considered.

In a typical `kubelet` environment, Docker or another container management system is responsible for workload isolation and providing security controls.  Since the goal of this project is to allow workloads that may not fit in a container environment to be managed with Kubernetes, it is the responsibility of the agent implementation to handle these concerns.

Here are some principles and available EC2 components that may aid in this effort:

#### Principles

1. Isolate the application from network access
2. Isolate the application from the access to the instance's file system
3. Regulate the EC2 host machine resources that available to the application 

#### Components
- [IAM roles for Amazon EC2](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)
- [Instance MetaData Service (IMDSv2)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html)

## Shared Responsibility

When extending this project and implementing the new VKVMAgent for production purposes, it is your responsibility to
also ensure the security of the operating system and applications running on the worker nodes.

Configuring and operating the kubernetes cluster securely also requires understanding and following kubernetes security
practices.

### Additional Resources
- [Securing a Cluster](https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/)
- [Amazon EKS Best Practices Guide for Security](https://aws.github.io/aws-eks-best-practices/security/docs/)

---
>© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.