# AWS IAM provisioner operator

[![Release](https://img.shields.io/github/v/release/edenlabllc/aws-iam-provisioner.operators.infra.svg?style=for-the-badge)](https://github.com/edenlabllc/aws-iam-provisioner.operators.infra/releases/latest)
[![Software License](https://img.shields.io/github/license/edenlabllc/aws-iam-provisioner.operators.infra.svg?style=for-the-badge)](LICENSE)
[![Powered By: Edenlab](https://img.shields.io/badge/powered%20by-edenlab-8A2BE2.svg?style=for-the-badge)](https://edenlab.io)

The AWS IAM provisioner operator provisions IAM roles and policies on the fly for the Kubernetes clusters
managed using [Kubernetes Cluster API Provider AWS](https://cluster-api-aws.sigs.k8s.io/getting-started).

## Description

After a managed [AWS EKS](https://aws.amazon.com/eks/) cluster is provisioned using
[Kubernetes Cluster API Provider AWS](https://cluster-api-aws.sigs.k8s.io/getting-started), it might be required
to provision IAM roles and policies for the installed services,
e.g. [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/),
[AWS Elastic Block Store CSI driver](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/tree/master).

The AWS IAM Provisioner Operator implements the creation and management of AWS IAM resources, including `roles` and `policies`, 
based on information obtained after provisioning the target cluster via the [Kubernetes Cluster API Provider AWS.](https://cluster-api-aws.sigs.k8s.io/getting-started)

While a CR of an IAM policy is a static definition, which can be defined in advance, a CR of an IAM role might contain
dynamic parts such as OIDC ARN/name of a created EKS cluster. It means, that while a role can reference existing
policies,
the dynamic parts should be provisioned on the fly, upon a managed EKS cluster is provisioned by Cluster API and is
ready.

For that purposes, the `AWSIAMProvision` CR can be used for defining a Golang template for such on-the-fly IAM role
provisioning:

```yaml
apiVersion: iam.aws.edenlab.io/v1alpha1
kind: AWSIAMProvision
metadata:
  name: deps-develop
  namespace: capa-system
  # truncated
spec:
  eksClusterName: deps-develop
  region: us-east-1
  roles:
    deps-develop-ebs-csi-controller:
      spec:
        assumeRolePolicyDocument: |
          {
            "Version": "2012-10-17",
            "Statement": [
              {
                "Sid": "",
                "Effect": "Allow",
                "Principal": {
                  "Federated": "{{ .OIDCProviderARN }}"
                },
                "Action": "sts:AssumeRoleWithWebIdentity",
                "Condition": {
                  "StringEquals": {
                    "{{ .OIDCProviderName }}:sub": "system:serviceaccount:kube-system:ebs-csi-controller"
                  }
                }
              }
            ]
          }
        policies:
          - deps-develop-ebs-csi-controller-core
        tags:
          - key: Foo
            value: Bar
  policies:
    deps-develop-ebs-csi-controller-core:
    spec:
      name: deps-develop-ebs-csi-controller-core
      tags:
        - key: Foo
          value: Bar
      policyDocument: |
        {
            "Statement": [
                {
                    "Action": [
                        "ec2:CreateSnapshot",
                        "ec2:AttachVolume",
                        "ec2:DetachVolume",
                        "ec2:ModifyVolume",
                        "ec2:DescribeAvailabilityZones",
                        "ec2:DescribeInstances",
                        "ec2:DescribeSnapshots",
  # truncated
```

The `assumeRolePolicyDocument` field of the `AWSIAMProvision` CR supports the following Golang template's placeholders:

- `{{ .OIDCProviderARN }}`: rendered to something
  like `arn:aws:iam::012345678901:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/AAAAABBBBB0000011111222223333344`
- `{{ .OIDCProviderName }}`: rendered to something
  like `oidc.eks.us-east-1.amazonaws.com/id/AAAAABBBBB0000011111222223333344`

> In this example, the `kube-system:ebs-csi-controller` part means, that the `ebs-csi-controller` K8S service account is
> in the `kube-system` namespace.

The rest of the `spec.roles.*.spec` fields are identical to the original AWS IAM role.

The [AWSManagedControlPlane](https://cluster-api-aws.sigs.k8s.io/crd/#controlplane.cluster.x-k8s.io/v1beta2.AWSManagedControlPlane)
Cluster API CR is watched the AWS IAM provisioner operator. As the result, a role CR will be created on the fly by the
operator upon a EKS cluster provisioning.

Example of an original `AWSManagedControlPlane` resource:

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta2
kind: AWSManagedControlPlane
metadata:
  name: deps-develop
  namespace: capa-system
  # truncated
spec:
  controlPlaneEndpoint:
    host: https://AAAAABBBBB0000011111222223333344.gr7.us-east-1.eks.amazonaws.com
    port: 443
  eksClusterName: deps-develop
  region: us-east-1
  roleName: deps-develop-iam-service-role
  version: v1.29.8
  # truncated
```

> `spec.roles.*.policies` should be attached to an existing AWS IAM `Policy` created by the AWS IAM Provisioner Operator.

> `spec.*.*.tags` field is used to define additional custom tags. Tags can only be specified at the time of `policy` or `role`
> creation and cannot be updated after a resource has been created.

### AWS IAM Provisioner Operator behavior

The AWS IAM Provisioner Operator follows idempotent behavior and a declarative configuration approach.
It continuously synchronizes `policy` or `role` resources based on the current state of AWS resources and 
the CR configuration.

Supported operations on AWS IAM Resources (`policy` or `role`):
- Creating or deleting resources based on the CR configuration.
- Updating the trust relationship policy document for a `role` or the `policy` document for a policy.
- Attaching or detaching policies from roles according to the CR configuration.

> [Full Example of CR Configuration](config/samples/iam_v1alpha1_awsiamprovision.yaml)

## Getting Started

### Prerequisites

- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/aws-iam-provisioner:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/aws-iam-provisioner:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
> privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

> **NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/aws-iam-provisioner:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/aws-iam-provisioner/<tag or branch>/dist/install.yaml
```
