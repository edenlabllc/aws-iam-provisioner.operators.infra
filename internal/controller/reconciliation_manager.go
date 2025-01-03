package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	iamctrlv1alpha1 "github.com/aws-controllers-k8s/iam-controller/apis/v1alpha1"
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ekscontrolplanev1 "sigs.k8s.io/cluster-api-provider-aws/v2/controlplane/eks/api/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1alpha1 "aws-iam-provisioner.operators.infra/api/v1alpha1"
)

var (
	roleMeta = metav1.TypeMeta{
		APIVersion: "iam.services.k8s.aws",
		Kind:       "Role",
	}
)

type oidcProviderTemplateData struct {
	OIDCProviderARN  string
	OIDCProviderName string
}

type ReconciliationManager struct {
	context *context.Context
	client  client.Client
	logger  *logr.Logger
	request *ctrl.Request
	scheme  *runtime.Scheme
	status  client.SubResourceWriter
}

func (rm *ReconciliationManager) getClusterResources() (*iamv1alpha1.AWSIAMProvision, *ekscontrolplanev1.AWSManagedControlPlane, error) {
	awsIAMProvision := &iamv1alpha1.AWSIAMProvision{}

	if err := rm.client.Get(*rm.context, rm.request.NamespacedName, awsIAMProvision); err != nil {
		if k8serrors.IsNotFound(err) {
			rm.logger.Info(fmt.Sprintf("AWSIAMProvision not found: %s", rm.request.NamespacedName))

			return nil, nil, nil
		}

		return nil, nil, err
	}

	rm.logger.Info(fmt.Sprintf("AWSIAMProvision found: %s", rm.request.NamespacedName))

	eksControlPlane := &ekscontrolplanev1.AWSManagedControlPlane{}
	namespacedName := types.NamespacedName{Name: awsIAMProvision.Spec.EksClusterName, Namespace: rm.request.NamespacedName.Namespace}

	if err := rm.client.Get(*rm.context, namespacedName, eksControlPlane); err != nil {
		if k8serrors.IsNotFound(err) {
			msg := fmt.Sprintf("AWSManagedControlPlane of %s AWSIAMProvision not found: %s", rm.request.NamespacedName, namespacedName)
			rm.logger.Info(msg)

			if err := rm.updateCRDStatus(awsIAMProvision, "Provisioning", msg, nil); err != nil {
				return nil, nil, err
			}

			return nil, nil, nil
		}

		if err := rm.updateCRDStatus(awsIAMProvision, "Failed", err.Error(), nil); err != nil {
			return nil, nil, err
		}

		return nil, nil, err
	}

	rm.logger.Info(fmt.Sprintf("AWSManagedControlPlane of %s AWSIAMProvision found: %s", rm.request.NamespacedName, namespacedName))

	if !eksControlPlane.Status.Ready {
		msg := fmt.Sprintf("AWSManagedControlPlane of %s AWSIAMProvision not ready: %s", rm.request.NamespacedName, namespacedName)
		rm.logger.Info(msg)

		if err := rm.updateCRDStatus(awsIAMProvision, "Provisioning", msg, nil); err != nil {
			return nil, nil, err
		}

		return nil, nil, nil
	}

	return awsIAMProvision, eksControlPlane, nil
}

func (rm *ReconciliationManager) handleRole(awsIAMProvision *iamv1alpha1.AWSIAMProvision, eksControlPlane *ekscontrolplanev1.AWSManagedControlPlane, name string, item *iamv1alpha1.AWSIAMProvisionRole) (*iamctrlv1alpha1.Role, bool, error) {
	k8sResource := &iamctrlv1alpha1.Role{}
	namespacedName := types.NamespacedName{Name: name, Namespace: rm.request.NamespacedName.Namespace}

	if err := rm.client.Get(*rm.context, namespacedName, k8sResource); err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, false, err
		}

		// Create new role
		if err := rm.setDefaultValues(awsIAMProvision, eksControlPlane, item); err != nil {
			return nil, false, err
		}

		if err := rm.validateRolePolicyRefs(awsIAMProvision, item); err != nil {
			return nil, false, err
		}

		k8sResource.TypeMeta = roleMeta
		k8sResource.ObjectMeta = metav1.ObjectMeta{
			Name:      name,
			Namespace: rm.request.NamespacedName.Namespace,
		}
		k8sResource.Spec = item.Spec

		// Set ownerReferences to ensure that the created resource will be deleted when the custom resource object is removed
		if err := ctrl.SetControllerReference(awsIAMProvision, k8sResource, rm.scheme); err != nil {
			if err := rm.updateCRDStatus(awsIAMProvision, "Failed", err.Error(), nil); err != nil {
				return nil, false, err
			}

			return nil, false, err
		}

		if err = rm.client.Create(*rm.context, k8sResource); err != nil {
			if err := rm.updateCRDStatus(awsIAMProvision, "Failed", err.Error(), nil); err != nil {
				return nil, false, err
			}

			return nil, false, err
		}

		rm.logger.Info(fmt.Sprintf("IAM Role of %s AWSIAMProvision created: %s", rm.request.NamespacedName, namespacedName))

		return k8sResource, true, nil
	}

	if err := rm.setDefaultValues(awsIAMProvision, eksControlPlane, item); err != nil {
		return nil, false, err
	}

	if cmp.Equal(item.Spec, k8sResource.Spec) {
		// No diff with existing resource, exiting without error
		rm.logger.Info(fmt.Sprintf("IAM Role of %s AWSIAMProvision equal: %s", rm.request.NamespacedName, namespacedName))

		return k8sResource, false, nil
	}

	rm.logger.Info(fmt.Sprintf("IAM Role of %s AWSIAMProvision different: %s", rm.request.NamespacedName, namespacedName))

	if err := rm.validateRolePolicyRefs(awsIAMProvision, item); err != nil {
		return nil, false, err
	}

	// Update role with new values
	k8sResource.Spec = item.Spec

	if err := rm.client.Update(*rm.context, k8sResource); err != nil {
		return nil, false, err
	}

	rm.logger.Info(fmt.Sprintf("IAM Role of %s AWSIAMProvision updated: %s", rm.request.NamespacedName, namespacedName))

	return k8sResource, true, nil
}

func (rm *ReconciliationManager) updateCRDStatus(awsIAMProvision *iamv1alpha1.AWSIAMProvision, phase, message string, roleStatuses map[string]*iamctrlv1alpha1.RoleStatus) error {
	awsIAMProvision.Status.LastUpdatedTime = &metav1.Time{Time: time.Now()}
	awsIAMProvision.Status.Phase = phase
	awsIAMProvision.Status.Message = message

	if roleStatuses != nil {
		statusRoles := make(map[string]iamv1alpha1.AWSIAMProvisionStatusRole)

		for roleName, roleStatus := range roleStatuses {
			statusRolePhase := "Provisioned"
			statusRoleMessage := ""

			for _, condition := range roleStatus.Conditions {
				// if any of roleStatus.condition.Status is not True, then the overall roleStatus.phase is considered Failed
				if condition.Status != v1.ConditionTrue {
					statusRolePhase = "Failed"
					statusRoleMessage = *condition.Message

					break
				}
			}

			statusRoles[roleName] = iamv1alpha1.AWSIAMProvisionStatusRole{
				Phase:   statusRolePhase,
				Message: statusRoleMessage,
				Status:  *roleStatus,
			}
		}

		awsIAMProvision.Status.Roles = statusRoles
	}

	if err := rm.status.Update(*rm.context, awsIAMProvision); err != nil {
		return errors.New(fmt.Sprintf("Unable to update status for CRD: %s, error: %s", awsIAMProvision.Name, err))
	}

	return nil
}

func (rm *ReconciliationManager) validateRolePolicyRefs(awsIAMProvision *iamv1alpha1.AWSIAMProvision, item *iamv1alpha1.AWSIAMProvisionRole) error {
	for _, policyRef := range item.Spec.PolicyRefs {
		// Check IAM policy exists
		_, err := rm.getPolicy(awsIAMProvision, item, policyRef)

		if err != nil {
			if err := rm.updateCRDStatus(awsIAMProvision, "Failed", err.Error(), nil); err != nil {
				return err
			}

			return err
		}
	}

	return nil
}

func (rm *ReconciliationManager) getPolicy(awsIAMProvision *iamv1alpha1.AWSIAMProvision, item *iamv1alpha1.AWSIAMProvisionRole, policyRef *ackv1alpha1.AWSResourceReferenceWrapper) (*iamctrlv1alpha1.Policy, error) {
	k8sResource := &iamctrlv1alpha1.Policy{}
	namespacedName := types.NamespacedName{Name: *policyRef.From.Name, Namespace: *policyRef.From.Namespace}

	if err := rm.client.Get(*rm.context, namespacedName, k8sResource); err != nil {
		if k8serrors.IsNotFound(err) {
			err = errors.New(fmt.Sprintf("IAM Policy of %s IAM Role of %s AWSIAMProvision not found: %s", *item.Spec.Name, rm.request.NamespacedName, namespacedName))
		}

		if err := rm.updateCRDStatus(awsIAMProvision, "Failed", err.Error(), nil); err != nil {
			return nil, err
		}

		return nil, err
	}

	return k8sResource, nil
}

func (rm *ReconciliationManager) setDefaultValues(awsIAMProvision *iamv1alpha1.AWSIAMProvision, eksControlPlane *ekscontrolplanev1.AWSManagedControlPlane, item *iamv1alpha1.AWSIAMProvisionRole) error {
	// Set default values to prevent unwanted diffs (the logic is similar to aws-iam-controller)
	if item.Spec.MaxSessionDuration == nil {
		defaultMaxSessionDuration := int64(3600)
		item.Spec.MaxSessionDuration = &defaultMaxSessionDuration
	}

	if item.Spec.Path == nil {
		defaultPath := "/"
		item.Spec.Path = &defaultPath
	}

	// Set rendered template to detect the diff correctly
	if err := rm.setAssumeRolePolicyDocument(awsIAMProvision, eksControlPlane, item); err != nil {
		return err
	}

	return nil
}

func (rm *ReconciliationManager) setAssumeRolePolicyDocument(awsIAMProvision *iamv1alpha1.AWSIAMProvision, eksControlPlane *ekscontrolplanev1.AWSManagedControlPlane, item *iamv1alpha1.AWSIAMProvisionRole) error {
	oidcProviderARN := eksControlPlane.Status.OIDCProvider.ARN
	_, oidcProviderName, oidcProviderARNFound := strings.Cut(oidcProviderARN, "/")

	if !oidcProviderARNFound {
		namespacedName := types.NamespacedName{Name: awsIAMProvision.Spec.EksClusterName, Namespace: rm.request.NamespacedName.Namespace}
		err := errors.New(fmt.Sprintf("OIDC ARN of %s AWSManagedControlPlane of %s AWSIAMProvision malformed: %s", namespacedName, rm.request.NamespacedName, oidcProviderARN))

		if err := rm.updateCRDStatus(awsIAMProvision, "Failed", err.Error(), nil); err != nil {
			return err
		}

		return err
	}

	if assumeRolePolicyDocument, err := rm.renderOIDCProviderTemplate(*item.Spec.AssumeRolePolicyDocument, oidcProviderARN, oidcProviderName); err != nil {
		if err := rm.updateCRDStatus(awsIAMProvision, "Failed", err.Error(), nil); err != nil {
			return err
		}

		return err
	} else {
		item.Spec.AssumeRolePolicyDocument = &assumeRolePolicyDocument

		return nil
	}
}

func (rm *ReconciliationManager) renderOIDCProviderTemplate(oidcProviderTemplate, oidcProviderARN, oidcProviderName string) (string, error) {
	oidcProviderTemplateData := oidcProviderTemplateData{oidcProviderARN, oidcProviderName}

	tmpl, err := template.New("").Parse(oidcProviderTemplate)
	if err != nil {
		return "", err
	}

	var tmplString bytes.Buffer
	if err := tmpl.Execute(&tmplString, oidcProviderTemplateData); err != nil {
		return "", err
	}

	return tmplString.String(), nil
}

func (rm *ReconciliationManager) transformPoliciesToRefs(policies []*iamctrlv1alpha1.Policy) []*ackv1alpha1.AWSResourceReferenceWrapper {
	var policyRefs []*ackv1alpha1.AWSResourceReferenceWrapper

	for _, policy := range policies {
		policyRefs = append(policyRefs, &ackv1alpha1.AWSResourceReferenceWrapper{
			From: &ackv1alpha1.AWSResourceReference{
				Name:      &policy.Name,
				Namespace: &policy.Namespace,
			},
		})
	}

	return policyRefs
}
