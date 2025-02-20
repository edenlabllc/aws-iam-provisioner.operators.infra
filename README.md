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

For Kubernetes-based resource
provisioning, [AWS Controllers for Kubernetes](https://aws-controllers-k8s.github.io/community/)
can be used to provision IAM [policies](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/policy/)
and [roles](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/role/).
Custom resources (CRs) for the controller might look like the following:

```yaml
apiVersion: iam.services.k8s.aws/v1alpha1
kind: Policy
metadata:
  name: deps-develop-ebs-csi-controller-core
  namespace: capa-system
  # truncated
spec:
  name: deps-develop-ebs-csi-controller-core
  # truncated
```

```yaml
apiVersion: iam.services.k8s.aws/v1alpha1
kind: Role
metadata:
  name: deps-develop-ebs-csi-controller
  namespace: capa-system
  # truncated
spec:
  name: deps-develop-ebs-csi-controller
  # truncated
```

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
        maxSessionDuration: 3600
        name: deps-develop-ebs-csi-controller
        path: /
        policyRefs:
          - from:
              name: deps-develop-ebs-csi-controller-core
              namespace: capa-system
  # truncated
```

The `assumeRolePolicyDocument` field of the `AWSIAMProvision` CR supports the following Golang template's placeholders:

- `{{ .OIDCProviderARN }}`: rendered to something
  like `arn:aws:iam::012345678901:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/AAAAABBBBB0000011111222223333344`
- `{{ .OIDCProviderName }}`: rendered to something
  like `oidc.eks.us-east-1.amazonaws.com/id/AAAAABBBBB0000011111222223333344`

> In this example, the `kube-system:ebs-csi-controller` part means, that the `ebs-csi-controller` K8S service account is
> in the `kube-system` namespace.

The rest of the `spec.roles.*.spec` fields are identical to the original AWS
IAM [Role](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/role/).

The [AWSManagedControlPlane](https://cluster-api-aws.sigs.k8s.io/crd/#controlplane.cluster.x-k8s.io/v1beta2.AWSManagedControlPlane)
Cluster API CR is watched the AWS IAM provisioner operator. As the result, a role CR will be created on the fly by the
operator
upon a EKS cluster provisioning.

> `policyRefs` should reference existing AWS IAM `Policy` CRs created by AWS IAM Controller.

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

Example of a result AWS IAM [Role](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/role/) created
by the operator:

```yaml
apiVersion: iam.services.k8s.aws/v1alpha1
kind: Role
metadata:
  name: deps-develop-ebs-csi-controller
  namespace: capa-system
  ownerReferences:
    - apiVersion: iam.aws.edenlab.io/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: AWSIAMProvision
      name: deps-develop
      uid: 77b58794-73cc-4a36-bbd9-572165ff6664
  # truncated
spec:
  assumeRolePolicyDocument: |
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Sid": "",
          "Effect": "Allow",
          "Principal": {
            "Federated": "arn:aws:iam::012345678901:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/AAAAABBBBB0000011111222223333344"
          },
          "Action": "sts:AssumeRoleWithWebIdentity",
          "Condition": {
            "StringEquals": {
              "oidc.eks.us-east-1.amazonaws.com/id/AAAAABBBBB0000011111222223333344:sub": "system:serviceaccount:kube-system:ebs-csi-controller"
            }
          }
        }
      ]
    }
  maxSessionDuration: 3600
  name: deps-develop-ebs-csi-controller
  path: /
  policyRefs:
    - from:
        name: deps-develop-ebs-csi-controller-core
        namespace: capa-system
  # truncated
```

> The `ownerReferences` field is set to define a parent-child relationship between AWSIAMProvision and the managed Role.

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
