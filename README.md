# aws-iam-provisioner

The AWS IAM provisioner operator provisions IAM roles on the fly for the Kubernetes clusters 
provisioned using [Cluster API](https://cluster-api-aws.sigs.k8s.io/getting-started).

## Description

After a managed [AWS EKS](https://aws.amazon.com/eks/) cluster is provisioned using 
[Kubernetes Cluster API Provider AWS](https://cluster-api-aws.sigs.k8s.io/getting-started), it might be required 
to provision IAM roles and policies for installed services, 
e.g. [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/), 
[AWS Elastic Block Store CSI driver](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/tree/master).

For Kubernetes-based resource provisioning, [AWS Controllers for Kubernetes](https://aws-controllers-k8s.github.io/community/)
can be used to provision IAM [policies](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/policy/) 
and [roles](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/role/). 
Custom resources (CRs) for the controller might look like the following:

```yaml
apiVersion: iam.services.k8s.aws/v1alpha1
kind: Policy
metadata:
  name: deps-ffs-1-ebs-csi-controller-core
  namespace: capa-system
  # truncated
spec:
  name: deps-ffs-1-ebs-csi-controller-core
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
                    "ec2:DescribeTags",
                    "ec2:DescribeVolumes",
                    "ec2:DescribeVolumesModifications"
                ],
                "Effect": "Allow",
                "Resource": "*"
            }
        ],
        "Version": "2012-10-17"
    }
  # truncated
```

```yaml
apiVersion: iam.services.k8s.aws/v1alpha1
kind: Role
metadata:
  name: deps-ffs-1-ebs-csi-controller
  namespace: capa-system
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
            "Federated": "arn:aws:iam::288509344804:oidc-provider/oidc.eks.eu-north-1.amazonaws.com/id/A71AAF56A08649E2055C1343D2FE70C8"
          },
          "Action": "sts:AssumeRoleWithWebIdentity",
          "Condition": {
            "StringEquals": {
              "oidc.eks.eu-north-1.amazonaws.com/id/A71AAF56A08649E2055C1343D2FE70C8:sub": "system:serviceaccount:kube-system:ebs-csi-controller"
            }
          }
        }
      ]
    }
  maxSessionDuration: 3600
  name: deps-ffs-1-ebs-csi-controller
  path: /
  policyRefs:
    - from:
        name: deps-ffs-1-ebs-csi-controller-core
        namespace: capa-system
  # truncated
```

While a CR of an IAM policy is a static definition, which can be defined in advance, a CR of IAM role might contain 
dynamic parts such as OIDC ARN/name of a created cluster. It means that while a role can reference existing policies,
the dynamic parts should be provisioned on the fly upon a managed EKS cluster is provisioned by Cluster API and ready.
Therefor, the `assumeRolePolicyDocument` field might contain the following Golang template's placeholders:
- `{{ .OIDCProviderARN }}`: will be rendered to something like `arn:aws:iam::288509344804:oidc-provider/oidc.eks.eu-north-1.amazonaws.com/id/A71AAF56A08649E2055C1343D2FE70C8`
- `{{ .OIDCProviderName }}`: will be rendered to something like `oidc.eks.eu-north-1.amazonaws.com/id/A71AAF56A08649E2055C1343D2FE70C8`

Example of the templated `assumeRolePolicyDocument`:

```json
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
```

Example of the `AWSIAMProvision` CR:

```yaml
apiVersion: iam.aws.edenlab.io/v1alpha1
kind: AWSIAMProvision
metadata:
  name: deps-ffs-1
  namespace: capa-system
  # truncated
spec:
  eksClusterName: deps-ffs-1
  roles:
    deps-ffs-1-ebs-csi-controller:
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
        name: deps-ffs-1-ebs-csi-controller
        path: /
        policyRefs:
          - from:
              name: deps-ffs-1-ebs-csi-controller-core
              namespace: capa-system
  # truncated
```

As the result, an AWS IAM [Role](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/role/) CR is created on the fly.

> `policyRefs` should reference existing AWS IAM `Policy` resources created by AWS IAM Controller.

Example of an [AWSManagedControlPlane](https://cluster-api-aws.sigs.k8s.io/crd/#controlplane.cluster.x-k8s.io/v1beta2.AWSManagedControlPlane):

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta2
kind: AWSManagedControlPlane
metadata:
  name: deps-ffs-1
  namespace: capa-system
  # truncated
spec:
  addons: []
  associateOIDCProvider: true
  bastion:
    allowedCIDRBlocks:
    - 0.0.0.0/0
    enabled: false
  controlPlaneEndpoint:
    host: https://A71AAF56A08649E2055C1343D2FE70C8.gr7.eu-north-1.eks.amazonaws.com
    port: 443
  eksClusterName: deps-ffs-1
  endpointAccess:
    private: false
    public: true
    publicCIDRs:
    - 0.0.0.0/0
  iamAuthenticatorConfig: {}
  identityRef:
    kind: AWSClusterStaticIdentity
    name: aws-cluster-identity
  kubeProxy:
    disable: false
  logging:
    apiServer: false
    audit: false
    authenticator: false
    controllerManager: false
    scheduler: false
  network:
    cni: {}
    subnets:
      # truncated
    vpc:
      availabilityZoneSelection: Ordered
      availabilityZoneUsageLimit: 3
      cidrBlock: 10.0.0.0/16
      emptyRoutesDefaultVPCSecurityGroup: true
      id: vpc-06a4940fc7a2ee655
      internetGatewayId: igw-0720f332efe5f4af5
      privateDnsHostnameTypeOnLaunch: ip-name
      subnetSchema: PreferPrivate
      tags:
        Name: deps-ffs-1-vpc
        sigs.k8s.io/cluster-api-provider-aws/cluster/deps-ffs-1: owned
        sigs.k8s.io/cluster-api-provider-aws/role: common
  partition: aws
  region: eu-north-1
  restrictPrivateSubnets: true
  roleName: deps-ffs-1-iam-service-role
  sshKeyName: deps-ffs-1
  tokenMethod: iam-authenticator
  version: v1.29.8
  vpcCni:
    disable: false
  # truncated
```

Example of a created AWS IAM [Role](https://aws-controllers-k8s.github.io/community/reference/iam/v1alpha1/role/):

```yaml
apiVersion: iam.services.k8s.aws/v1alpha1
kind: Role
metadata:
  name: deps-ffs-1-ebs-csi-controller
  namespace: capa-system
  ownerReferences:
  - apiVersion: iam.aws.edenlab.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AWSIAMProvision
    name: deps-ffs-1
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
            "Federated": "arn:aws:iam::288509344804:oidc-provider/oidc.eks.eu-north-1.amazonaws.com/id/A71AAF56A08649E2055C1343D2FE70C8"
          },
          "Action": "sts:AssumeRoleWithWebIdentity",
          "Condition": {
            "StringEquals": {
              "oidc.eks.eu-north-1.amazonaws.com/id/A71AAF56A08649E2055C1343D2FE70C8:sub": "system:serviceaccount:kube-system:ebs-csi-controller"
            }
          }
        }
      ]
    }
  maxSessionDuration: 3600
  name: deps-ffs-1-ebs-csi-controller
  path: /
  policyRefs:
  - from:
      name: deps-ffs-1-ebs-csi-controller-core
      namespace: capa-system
  # truncated
```

> `ownerReferences` is set to define a parent-child relationship between AWSIAMProvision and the created Role.

This operator contains one CRD which directs the operator to manage IAM roles upon an EKS cluster creation:

```yaml
spec:
  eksClusterName: deps-ffs-1
  roles:
    deps-ffs-1-ebs-csi-controller:
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
        name: deps-ffs-1-ebs-csi-controller
        path: /
        policyRefs:
          - from:
              name: deps-ffs-1-ebs-csi-controller-core
              namespace: capa-system
```

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

[//]: # (todo cluster api requirements?)

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
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

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

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024 anovikov-el.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

