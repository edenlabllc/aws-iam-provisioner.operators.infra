package controller

import (
	iamv1alpha1 "aws-iam-provisioner.operators.infra/api/v1alpha1"
	"bytes"
	"context"
	"errors"
	"fmt"
	iamctrlv1alpha1 "github.com/aws-controllers-k8s/iam-controller/apis/v1alpha1"
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ekscontrolplanev1 "sigs.k8s.io/cluster-api-provider-aws/v2/controlplane/eks/api/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"text/template"
	"time"
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
	r      *AWSIAMProvisionReconciler
	ctx    *context.Context
	req    *ctrl.Request
	logger *logr.Logger
}

func NewReconciliationManager(r *AWSIAMProvisionReconciler, ctx *context.Context, req *ctrl.Request) *ReconciliationManager {
	logger := log.FromContext(*ctx)

	return &ReconciliationManager{
		r,
		ctx,
		req,
		&logger,
	}
}

func (rm *ReconciliationManager) GetClusterResources() (*iamv1alpha1.AWSIAMProvision, *ekscontrolplanev1.AWSManagedControlPlane, error) {
	awsIAMProvision := &iamv1alpha1.AWSIAMProvision{}

	if err := rm.r.Client.Get(*rm.ctx, rm.req.NamespacedName, awsIAMProvision); err != nil {
		if k8serrors.IsNotFound(err) {
			err = errors.New(fmt.Sprintf("Cannot find AWSIAMProvision: %s", rm.req.NamespacedName))
		}

		return nil, nil, err
	}

	rm.logger.Info(fmt.Sprintf("Found AWSIAMProvision: %s", rm.req.NamespacedName))

	eksControlPlane := &ekscontrolplanev1.AWSManagedControlPlane{}
	eksControlPlaneNamespacedName := types.NamespacedName{Name: awsIAMProvision.Spec.EksClusterName, Namespace: rm.req.NamespacedName.Namespace}

	if err := rm.r.Client.Get(*rm.ctx, eksControlPlaneNamespacedName, eksControlPlane); err != nil {
		if k8serrors.IsNotFound(err) {
			err = errors.New(fmt.Sprintf("Cannot get AWSManagedControlPlane: %s", eksControlPlaneNamespacedName))
		}

		if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
			return nil, nil, err
		}

		return nil, nil, err
	}

	rm.logger.Info(fmt.Sprintf("Found AWSManagedControlPlane: %s", eksControlPlaneNamespacedName))

	if !eksControlPlane.Status.Ready {
		err := errors.New(fmt.Sprintf("AWSManagedControlPlane not ready: %s", eksControlPlaneNamespacedName))

		if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
			return nil, nil, err
		}

		return nil, nil, err
	}

	return awsIAMProvision, eksControlPlane, nil
}

func (rm *ReconciliationManager) HandleRole(awsIAMProvision *iamv1alpha1.AWSIAMProvision, eksControlPlane *ekscontrolplanev1.AWSManagedControlPlane, name string, item *iamv1alpha1.AWSIAMProvisionRole) (*iamctrlv1alpha1.Role, error) {
	k8sResource := &iamctrlv1alpha1.Role{}
	k8sResourceNamespacedName := types.NamespacedName{Name: name, Namespace: rm.req.NamespacedName.Namespace}

	oidcProviderArn := eksControlPlane.Status.OIDCProvider.ARN
	_, oidcProviderName, oidcProviderArnFound := strings.Cut(oidcProviderArn, "/")

	if !oidcProviderArnFound {
		err := errors.New(fmt.Sprintf("OIDC ARN malformed: %s", oidcProviderArn))

		if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := rm.r.Client.Get(*rm.ctx, k8sResourceNamespacedName, k8sResource); err != nil {
		if k8serrors.IsNotFound(err) {
			// Create new role
			k8sResource.TypeMeta = roleMeta
			k8sResource.ObjectMeta = metav1.ObjectMeta{
				Name:      name,
				Namespace: rm.req.NamespacedName.Namespace,
			}
			k8sResource.Spec = item.Spec

			assumeRolePolicyDocument, err := rm.renderOIDCProviderTemplate(*item.Spec.AssumeRolePolicyDocument, oidcProviderArn, oidcProviderName)
			if err != nil {
				if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
					return nil, err
				}

				return nil, err
			}

			k8sResource.Spec.AssumeRolePolicyDocument = &assumeRolePolicyDocument

			if err := rm.handleRolePolicyRefs(awsIAMProvision, item, k8sResource); err != nil {
				return nil, err
			}

			// Used to ensure that the created resource will be deleted when the custom resource object is removed
			if err := ctrl.SetControllerReference(awsIAMProvision, k8sResource, rm.r.Scheme); err != nil {
				if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
					return nil, err
				}

				return nil, err
			}

			if err = rm.r.Client.Create(*rm.ctx, k8sResource); err != nil {
				if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
					return nil, err
				}

				return nil, err
			}

			rm.logger.Info(fmt.Sprintf("Created IAM Role: %s", k8sResourceNamespacedName))
		} else {
			return nil, err
		}
	} else {
		// Update role with new values
		k8sResource.Spec = item.Spec

		assumeRolePolicyDocument, err := rm.renderOIDCProviderTemplate(*item.Spec.AssumeRolePolicyDocument, oidcProviderArn, oidcProviderName)
		if err != nil {
			if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
				return nil, err
			}

			return nil, err
		}

		k8sResource.Spec.AssumeRolePolicyDocument = &assumeRolePolicyDocument

		if err := rm.handleRolePolicyRefs(awsIAMProvision, item, k8sResource); err != nil {
			return nil, err
		}

		if !cmp.Equal(item.Spec, k8sResource.Spec) {
			if err := rm.r.Client.Update(*rm.ctx, k8sResource); err != nil {
				return nil, err
			}

			rm.logger.Info(fmt.Sprintf("Updated IAM Role: %s", k8sResourceNamespacedName))
		}
	}

	return k8sResource, nil
}

func (rm *ReconciliationManager) handleRolePolicyRefs(awsIAMProvision *iamv1alpha1.AWSIAMProvision, item *iamv1alpha1.AWSIAMProvisionRole, k8sResource *iamctrlv1alpha1.Role) error {
	for _, policyRef := range item.Spec.PolicyRefs {
		// Check IAM policy exists
		_, err := rm.getPolicy(awsIAMProvision, policyRef)

		if err != nil {
			if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
				return err
			}

			return err
		}
	}

	return nil
}

func (rm *ReconciliationManager) getPolicy(awsIAMProvision *iamv1alpha1.AWSIAMProvision, policyRef *ackv1alpha1.AWSResourceReferenceWrapper) (*iamctrlv1alpha1.Policy, error) {
	k8sResource := &iamctrlv1alpha1.Policy{}
	k8sResourceNamespacedName := types.NamespacedName{Name: *policyRef.From.Name, Namespace: *policyRef.From.Namespace}

	if err := rm.r.Client.Get(*rm.ctx, k8sResourceNamespacedName, k8sResource); err != nil {
		if k8serrors.IsNotFound(err) {
			err = errors.New(fmt.Sprintf("Cannot get IAM Policy: %s", k8sResourceNamespacedName))
		}

		if err := rm.UpdateCRDStatus(awsIAMProvision, "Error", err.Error()); err != nil {
			return nil, err
		}

		return nil, err
	}

	return k8sResource, nil
}

func (rm *ReconciliationManager) renderOIDCProviderTemplate(oidcProviderTemplate, oidcProviderArn, oidcProviderName string) (string, error) {
	oidcProviderTemplateData := oidcProviderTemplateData{oidcProviderArn, oidcProviderName}

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

func (rm *ReconciliationManager) UpdateCRDStatus(awsIAMProvision *iamv1alpha1.AWSIAMProvision, phase, errText string) error {
	awsIAMProvision.Status.LastUpdatedTime = &metav1.Time{Time: time.Now()}
	awsIAMProvision.Status.Phase = phase
	awsIAMProvision.Status.Error = errText

	if err := rm.r.Status().Update(*rm.ctx, awsIAMProvision); err != nil {
		return errors.New(fmt.Sprintf("Unable to update status for CRD: %s", awsIAMProvision.Name))
	}

	return nil
}
