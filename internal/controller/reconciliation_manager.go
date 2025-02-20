package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	iamv1alpha1 "aws-iam-provisioner.operators.infra/api/v1alpha1"
	"aws-iam-provisioner.operators.infra/internal/aws_sdk"
	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ekscontrolplanev1 "sigs.k8s.io/cluster-api-provider-aws/v2/controlplane/eks/api/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const awsIAMProvisionFinalizerName = "awsiamprovision.iam.aws.edenlab.io/finalizer"

type awsIAMResources struct {
	awsIAMProvision *iamv1alpha1.AWSIAMProvision
	eksCP           *ekscontrolplanev1.AWSManagedControlPlane
	eksCPNamespace  types.NamespacedName
}

type oidcProviderTemplateData struct {
	OIDCProviderARN  string
	OIDCProviderName string
}

type ReconciliationManager struct {
	client.Client
	ctx       context.Context
	IAMClient aws_sdk.IAMManager
	logger    logr.Logger
	request   ctrl.Request
	Scheme    *runtime.Scheme
}

func newAWSIAMResources() *awsIAMResources {
	return &awsIAMResources{
		awsIAMProvision: &iamv1alpha1.AWSIAMProvision{},
		eksCP:           &ekscontrolplanev1.AWSManagedControlPlane{},
		eksCPNamespace:  types.NamespacedName{},
	}
}

func (rm *ReconciliationManager) getClusterResources() (*awsIAMResources, error) {
	air := newAWSIAMResources()
	if err := rm.Get(rm.ctx, rm.request.NamespacedName, air.awsIAMProvision); err != nil {
		if k8serrors.IsNotFound(err) {
			rm.logger.Info(fmt.Sprintf("AWSIAMProvision not found: %s", rm.request.NamespacedName))

			return nil, nil
		}

		return nil, err
	}

	air.eksCPNamespace = types.NamespacedName{
		Name:      air.awsIAMProvision.Spec.EksClusterName,
		Namespace: rm.request.NamespacedName.Namespace,
	}
	if err := rm.Get(rm.ctx, air.eksCPNamespace, air.eksCP); err != nil {
		if k8serrors.IsNotFound(err) {
			msg := fmt.Sprintf("AWSManagedControlPlane of %s AWSIAMProvision not found: %s",
				rm.request.NamespacedName, air.eksCPNamespace)
			rm.logger.Info(msg)

			if err := rm.updateCRDStatus(air, "Provisioning", "", msg, nil); err != nil {
				return nil, err
			}

			return nil, nil
		}

		if err := rm.updateCRDStatus(air, "Failed", "", err.Error(), nil); err != nil {
			return nil, err
		}

		return nil, err
	}

	if !air.eksCP.Status.Ready {
		msg := fmt.Sprintf("AWSManagedControlPlane of %s AWSIAMProvision not ready: %s",
			rm.request.NamespacedName, air.eksCPNamespace)
		rm.logger.Info(msg)
		if err := rm.updateCRDStatus(air, "Provisioning", "", msg, nil); err != nil {
			return nil, err
		}

		return nil, nil
	}

	return air, nil
}

func (rm *ReconciliationManager) deleteIAMResources(awsIAMProvision *iamv1alpha1.AWSIAMProvision) error {
	for _, role := range awsIAMProvision.Spec.Roles {
		_, exists, err := rm.IAMClient.GetRoleByName(role.Spec.Name)
		if err != nil {
			return err
		}

		if exists {
			policies, err := rm.IAMClient.ListAttachedRolePolicies(role.Spec.Name)
			if err != nil {
				return err
			}

			if err := rm.IAMClient.BatchAttachDetachRolePolicies(aws_sdk.ButchDetachProc, policies, role.Spec.Name); err != nil {
				return err
			}

			if err := rm.IAMClient.BatchDeletePolicies(policies); err != nil {
				return err
			}

			if err := rm.IAMClient.DeleteRole(role.Spec.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rm *ReconciliationManager) syncAWSIAMResources(air *awsIAMResources) error {
	tags := aws_sdk.TagsDefine(air.awsIAMProvision.Spec.EksClusterName, air.awsIAMProvision.Namespace)

	iamRoles, err := rm.IAMClient.ListRolesByTags(tags)
	if err != nil {
		return err
	}

	iamPolicies, err := rm.IAMClient.ListPoliciesByTags(tags)
	if err != nil {
		return err
	}

	deleteRoles := make(map[string]struct{})
	for _, iamRole := range iamRoles {
		deleteRoles[*iamRole.RoleName] = struct{}{}
	}

	for _, role := range air.awsIAMProvision.Spec.Roles {
		if _, ok := deleteRoles[*role.Spec.Name]; ok {
			delete(deleteRoles, *role.Spec.Name)
		}

		detachRolePolicies := make(map[string]struct{})
		for _, iamPolicy := range iamPolicies {
			entities, err := rm.IAMClient.ListEntitiesForPolicy(&iamPolicy)
			if err != nil {
				return err
			}

			for _, entity := range entities {
				if *role.Spec.Name == *entity.RoleName {
					detachRolePolicies[*iamPolicy.PolicyName] = struct{}{}
				}
			}
		}

		for _, rolePolicy := range role.Spec.Policies {
			if _, ok := detachRolePolicies[*rolePolicy]; ok {
				delete(detachRolePolicies, *rolePolicy)
			}
		}

		for detachRolePolicy := range detachRolePolicies {
			if err := rm.IAMClient.DetachRolePolicy(&detachRolePolicy, role.Spec.Name); err != nil {
				return err
			}
		}
	}

	for role := range deleteRoles {
		_, exists, err := rm.IAMClient.GetRoleByName(&role)
		if err != nil {
			return err
		}

		if exists {
			policies, err := rm.IAMClient.ListAttachedRolePolicies(&role)
			if err != nil {
				return err
			}

			if err := rm.IAMClient.BatchAttachDetachRolePolicies(aws_sdk.ButchDetachProc, policies, &role); err != nil {
				return err
			}

			if err := rm.IAMClient.DeleteRole(&role); err != nil {
				return err
			}
		}
	}

	deletePolicies := make(map[string]struct{})
	for _, iamPolicy := range iamPolicies {
		deletePolicies[*iamPolicy.PolicyName] = struct{}{}
	}

	for _, policy := range air.awsIAMProvision.Spec.Policies {
		if _, ok := deletePolicies[*policy.Spec.Name]; ok {
			delete(deletePolicies, *policy.Spec.Name)
		}
	}

	for policyName := range deletePolicies {
		policy, exists, err := rm.IAMClient.GetPolicyByName(&policyName)
		if err != nil {
			return err
		}

		if exists {
			entities, err := rm.IAMClient.ListEntitiesForPolicy(policy)
			if err != nil {
				return err
			}

			if len(entities) == 0 {
				if err := rm.IAMClient.DeletePolicy(policy.PolicyName); err != nil {
					return err
				}
			}

			for _, role := range entities {
				if len(*role.RoleName) > 0 {
					if err := rm.IAMClient.DetachRolePolicy(policy.PolicyName, role.RoleName); err != nil {
						return err
					}

					if err := rm.IAMClient.DeletePolicy(policy.PolicyName); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (rm *ReconciliationManager) syncPoliciesByRoleSpec(air *awsIAMResources, role *iamv1alpha1.AWSIAMProvisionRole) error {
	for _, rolePolicy := range role.Spec.Policies {
		for _, policy := range air.awsIAMProvision.Spec.Policies {
			// Coordination of the list `spec.role.spec.policies` with list `spec.policies`.
			if *rolePolicy == *policy.Spec.Name {
				iamPolicy, exists, err := rm.IAMClient.GetPolicyByName(policy.Spec.Name)
				if err != nil {
					return err
				}

				checkSumTag := aws_sdk.NewChecksumTag(policy.Spec.PolicyDocument)
				tags := aws_sdk.TagsDefine(
					air.awsIAMProvision.Spec.EksClusterName,
					air.awsIAMProvision.Namespace,
					checkSumTag,
				)
				description := fmt.Sprintf("%s%s. %s",
					aws_sdk.PolicyDescriptionPrefix, air.awsIAMProvision.Spec.EksClusterName, aws_sdk.IAMDescription)
				// Creating and attaching policy if not created early.
				if !exists {
					result, err := rm.IAMClient.CreatePolicy(policy.Spec.Name, policy.Spec.PolicyDocument, &description, tags)
					if err != nil {
						return err
					}

					if err := rm.IAMClient.AttachRolePolicy(policy.Spec.Name, role.Spec.Name); err != nil {
						return err
					}

					if err := rm.updateCRDStatus(air, "Provisioned", "Created and Attached",
						fmt.Sprintf("Policy %s was created and attached with role %s.",
							*policy.Spec.Name, *role.Spec.Name), aws_sdk.IAMPolicy(result)); err != nil {
						return err
					}
				} else {
					// Updating policy document if was changed
					for _, tag := range iamPolicy.Tags {
						if *tag.Key == aws_sdk.TagKeyPolicyDocument && *tag.Value != *checkSumTag.Value {
							if err := rm.IAMClient.DetachRolePolicy(iamPolicy.PolicyName, role.Spec.Name); err != nil {
								return err
							}

							if err := rm.IAMClient.DeletePolicy(iamPolicy.PolicyName); err != nil {
								return err
							}

							_, err := rm.IAMClient.CreatePolicy(policy.Spec.Name, policy.Spec.PolicyDocument, &description, tags)
							if err != nil {
								return err
							}

							if err := rm.IAMClient.AttachRolePolicy(policy.Spec.Name, role.Spec.Name); err != nil {
								return err
							}

							if err := rm.updateCRDStatus(air, "Provisioned", "CreatedAttached",
								fmt.Sprintf("Policy document for policy %s was updated.",
									*policy.Spec.Name), aws_sdk.IAMPolicy(iamPolicy)); err != nil {
								return err
							}
						}
					}

					// Sync attachment the policies by list of `spec.role.spec.policies`.
					roleIAMPolicies, err := rm.IAMClient.ListAttachedRolePolicies(role.Spec.Name)
					if err != nil {
						return err
					}

					isAttachedToRole := make(map[string]struct{})
					for _, roleIAMPolicy := range roleIAMPolicies {
						isAttachedToRole[*roleIAMPolicy.PolicyName] = struct{}{}
					}

					if _, ok := isAttachedToRole[*rolePolicy]; !ok {
						if err := rm.IAMClient.AttachRolePolicy(policy.Spec.Name, role.Spec.Name); err != nil {
							return err
						}

						if err := rm.updateCRDStatus(air, "Provisioned", "Attached",
							fmt.Sprintf("Policy %s was attached to role %s.",
								*policy.Spec.Name, *role.Spec.Name), aws_sdk.IAMPolicy(iamPolicy)); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (rm *ReconciliationManager) syncRole(air *awsIAMResources, role *iamv1alpha1.AWSIAMProvisionRole) error {
	tags := aws_sdk.TagsDefine(air.awsIAMProvision.Spec.EksClusterName, air.awsIAMProvision.Namespace)

	if err := rm.setAssumeRolePolicyDocument(air, role); err != nil {
		return err
	}

	iamRole, exists, err := rm.IAMClient.GetRoleByName(role.Spec.Name)
	if err != nil {
		return err
	}

	if !exists {
		description := fmt.Sprintf("%s%s. %s",
			aws_sdk.RoleDescriptionPrefix, air.awsIAMProvision.Spec.EksClusterName, aws_sdk.IAMDescription)
		result, err := rm.IAMClient.CreateRole(role.Spec.Name, role.Spec.AssumeRolePolicyDocument, &description, tags)
		if err != nil {
			return err
		}

		if err := rm.updateCRDStatus(air, "Provisioned", "Created",
			fmt.Sprintf("Role %s was created.", *role.Spec.Name), aws_sdk.IAMRole(result)); err != nil {
			return err
		}

		if err := rm.syncPoliciesByRoleSpec(air, role); err != nil {
			return err
		}
	} else {
		diff, err := rm.IAMClient.DiffRoleByParams(iamRole.AssumeRolePolicyDocument,
			role.Spec.AssumeRolePolicyDocument, iamRole.Tags, tags)
		if err != nil {
			return err
		}

		if diff {
			if err := rm.IAMClient.UpdateRole(role.Spec.Name, role.Spec.AssumeRolePolicyDocument); err != nil {
				return err
			}

			if err := rm.updateCRDStatus(air, "Provisioned", "Updated",
				fmt.Sprintf("The trust relationship policy document for role %s was updated.",
					*role.Spec.Name), aws_sdk.IAMRole(iamRole)); err != nil {
				return err
			}
		}

		if err := rm.syncPoliciesByRoleSpec(air, role); err != nil {
			return err
		}
	}

	return nil
}

func (rm *ReconciliationManager) updateCRDStatus(air *awsIAMResources, crdPhase, phase, message string, result interface{}) error {
	var (
		ownerAccountID = iamv1alpha1.AWSAccountID(rm.IAMClient.GetIAMClientMetadata().AccountID)
		region         = iamv1alpha1.AWSRegion(rm.IAMClient.GetIAMClientMetadata().Region)
	)

	switch r := result.(type) {
	case aws_sdk.IAMRole:
		role, _, err := rm.IAMClient.GetRoleByName(r.RoleName)
		if err != nil {
			return err
		}

		arn := iamv1alpha1.AWSResourceName(*r.Arn)
		awsIAMProvisionStatusRole := iamv1alpha1.AWSIAMProvisionStatusRole{
			Name:    r.RoleName,
			Message: message,
			Phase:   phase,
			Status: iamv1alpha1.RoleStatus{
				AWSIAMResourceMetadata: &iamv1alpha1.AWSIAMResourceMetadata{
					ARN:            &arn,
					OwnerAccountID: &ownerAccountID,
					Region:         &region,
				},
				CreateDate: &metav1.Time{Time: *r.CreateDate},
				RoleID:     r.RoleId,
			},
		}

		nums := make(map[bool]int)
		for num, roleStatus := range air.awsIAMProvision.Status.Roles {
			if *roleStatus.Name == *role.RoleName {
				nums[true] = num
				break
			}
		}

		if num, ok := nums[true]; ok {
			air.awsIAMProvision.Status.Roles[num] = awsIAMProvisionStatusRole
		} else {
			air.awsIAMProvision.Status.Roles = append(air.awsIAMProvision.Status.Roles, awsIAMProvisionStatusRole)
		}
	case aws_sdk.IAMPolicy:
		policy, _, err := rm.IAMClient.GetPolicyByName(r.PolicyName)
		if err != nil {
			return err
		}

		arn := iamv1alpha1.AWSResourceName(*policy.Arn)
		awsIAMProvisionStatusPolicy := iamv1alpha1.AWSIAMProvisionStatusPolicy{
			Name:    policy.PolicyName,
			Message: message,
			Phase:   phase,
			Status: iamv1alpha1.PolicyStatus{
				AWSIAMResourceMetadata: &iamv1alpha1.AWSIAMResourceMetadata{
					ARN:            &arn,
					OwnerAccountID: &ownerAccountID,
					Region:         &region,
				},
				AttachmentCount:  policy.AttachmentCount,
				DefaultVersionID: policy.DefaultVersionId,
				CreateDate:       &metav1.Time{Time: *policy.CreateDate},
				PolicyID:         policy.PolicyId,
			},
		}

		nums := make(map[bool]int)
		for num, policyStatus := range air.awsIAMProvision.Status.Policies {
			if *policyStatus.Name == *policy.PolicyName {
				nums[true] = num
				break
			}
		}

		if num, ok := nums[true]; ok {
			air.awsIAMProvision.Status.Policies[num] = awsIAMProvisionStatusPolicy
		} else {
			air.awsIAMProvision.Status.Policies = append(air.awsIAMProvision.Status.Policies, awsIAMProvisionStatusPolicy)
		}
	}

	air.awsIAMProvision.Status.LastUpdatedTime = &metav1.Time{Time: time.Now()}
	air.awsIAMProvision.Status.Phase = crdPhase
	air.awsIAMProvision.Status.Message = message

	if err := rm.Status().Update(rm.ctx, air.awsIAMProvision); err != nil {
		return errors.New(fmt.Sprintf("Unable to update status for CRD: %s, error: %s", air.awsIAMProvision.Name, err))
	}

	return nil
}

func (rm *ReconciliationManager) setAssumeRolePolicyDocument(air *awsIAMResources, role *iamv1alpha1.AWSIAMProvisionRole) error {
	var oidcProviderARNFound bool

	oidcPr := &oidcProviderTemplateData{OIDCProviderARN: air.eksCP.Status.OIDCProvider.ARN}
	_, oidcPr.OIDCProviderName, oidcProviderARNFound = strings.Cut(oidcPr.OIDCProviderARN, "/")

	if !oidcProviderARNFound {
		err := errors.New(
			fmt.Sprintf("OIDC ARN of %s AWSManagedControlPlane of %s AWSIAMProvision malformed: %s",
				air.eksCPNamespace, rm.request.NamespacedName, oidcPr.OIDCProviderARN))
		if err := rm.updateCRDStatus(air, "Failed", "", err.Error(), nil); err != nil {
			return err
		}

		return err
	}

	if assumeRolePolicyDocument, err := rm.renderOIDCProviderTemplate(*role.Spec.AssumeRolePolicyDocument, oidcPr); err != nil {
		if err := rm.updateCRDStatus(air, "Failed", "", err.Error(), nil); err != nil {
			return err
		}

		return err
	} else {
		role.Spec.AssumeRolePolicyDocument = &assumeRolePolicyDocument

		return nil
	}
}

func (rm *ReconciliationManager) renderOIDCProviderTemplate(assumeRolePolicyDocument string, oidcPr *oidcProviderTemplateData) (string, error) {
	tmpl, err := template.New("").Parse(assumeRolePolicyDocument)
	if err != nil {
		return "", err
	}

	var tmplString bytes.Buffer
	if err := tmpl.Execute(&tmplString, oidcPr); err != nil {
		return "", err
	}

	return tmplString.String(), nil
}
