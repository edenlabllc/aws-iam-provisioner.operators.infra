package controller

import (
	"fmt"
	"time"

	iamv1alpha1 "aws-iam-provisioner.operators.infra/api/v1alpha1"
	iamType "github.com/aws/aws-sdk-go-v2/service/iam/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	attachPhase              = "Attached"
	createPhase              = "Created"
	deletePhase              = "Deleted"
	destroyIntermediatePhase = "Destroying"
	destroyPhase             = "Destroyed"
	detachPhase              = "Detached"
	failPhase                = "Failed"
	intermediatePhase        = "Provisioning"
	successPhase             = "Provisioned"
	updatePhase              = "Updated"
)

func (rm *ReconciliationManager) updateCRDStatus(air *awsIAMResources, crdPhase, phase, message string, result interface{}) error {
	var (
		ownerAccountID iamv1alpha1.AWSAccountID
		region         iamv1alpha1.AWSRegion
	)

	if result != nil {
		ownerAccountID = iamv1alpha1.AWSAccountID(rm.IAMClient.GetIAMClientMetadata().AccountID)
		region = iamv1alpha1.AWSRegion(rm.IAMClient.GetIAMClientMetadata().Region)
	}

	switch r := result.(type) {
	case *iamType.Role:
		role, exists, err := rm.IAMClient.GetRoleByName(r.RoleName)
		if err != nil {
			return err
		}

		if exists {
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
		} else {
			nums := make(map[bool]int)
			for num, roleStatus := range air.awsIAMProvision.Status.Roles {
				if *roleStatus.Name == *r.RoleName {
					nums[true] = num
					break
				}
			}

			if num, ok := nums[true]; ok {
				air.awsIAMProvision.Status.Roles = append(air.awsIAMProvision.Status.Roles[:num],
					air.awsIAMProvision.Status.Roles[num+1:]...)
			}
		}
	case *iamType.Policy:
		policy, exists, err := rm.IAMClient.GetPolicyByName(r.PolicyName)
		if err != nil {
			return err
		}

		if exists {
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
		} else {
			nums := make(map[bool]int)
			for num, policyStatus := range air.awsIAMProvision.Status.Policies {
				if *policyStatus.Name == *r.PolicyName {
					nums[true] = num
					break
				}
			}

			if num, ok := nums[true]; ok {
				air.awsIAMProvision.Status.Policies = append(air.awsIAMProvision.Status.Policies[:num],
					air.awsIAMProvision.Status.Policies[num+1:]...)
			}
		}
	}

	air.awsIAMProvision.Status.LastUpdatedTime = &metav1.Time{Time: time.Now()}
	air.awsIAMProvision.Status.Phase = crdPhase
	air.awsIAMProvision.Status.Message = message

	if err := rm.Status().Update(rm.ctx, air.awsIAMProvision); err != nil {
		return fmt.Errorf("unable to update status for CRD: %s, error: %s", air.awsIAMProvision.Name, err)
	}

	return nil
}
