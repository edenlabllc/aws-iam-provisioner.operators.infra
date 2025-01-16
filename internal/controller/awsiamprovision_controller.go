/*
Copyright 2025 Edenlab

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	iamctrlv1alpha1 "github.com/aws-controllers-k8s/iam-controller/apis/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	iamv1alpha1 "aws-iam-provisioner.operators.infra/api/v1alpha1"
)

const (
	frequency = time.Second * 10
)

// AWSIAMProvisionReconciler reconciles a AWSIAMProvision object
type AWSIAMProvisionReconciler struct {
	client.Client
	*ReconciliationManager
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=iam.aws.edenlab.io,resources=awsiamprovisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=iam.aws.edenlab.io,resources=awsiamprovisions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=iam.aws.edenlab.io,resources=awsiamprovisions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *AWSIAMProvisionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	r.ReconciliationManager = &ReconciliationManager{&ctx, r.Client, &logger, &req, r.Scheme, r.Status()}

	awsIAMProvision, eksControlPlane, err := r.getClusterResources()
	if err != nil {
		return ctrl.Result{}, err
	}

	if awsIAMProvision == nil || eksControlPlane == nil {
		// Resources not ready, re-queuing
		return ctrl.Result{RequeueAfter: frequency}, nil
	}

	awsIAMProvisionProvisioned := false
	sourceK8sResourceStatuses := make(map[string]*iamctrlv1alpha1.RoleStatus)
	for name, item := range awsIAMProvision.Spec.Roles {
		k8sResource, k8sResourceUpdated, err := r.handleRole(awsIAMProvision, eksControlPlane, name, &item)

		if err != nil {
			return ctrl.Result{}, err
		}

		if k8sResource != nil {
			sourceK8sResourceStatuses[k8sResource.Name] = &k8sResource.Status
		}

		if k8sResourceUpdated {
			awsIAMProvisionProvisioned = true
		}
	}

	targetK8sResourceStatuses := make(map[string]*iamctrlv1alpha1.RoleStatus)
	for name, awsIAMProvisionStatusRole := range awsIAMProvision.Status.Roles {
		targetK8sResourceStatuses[name] = &awsIAMProvisionStatusRole.Status
	}

	k8sResourceStatusesEqual := cmp.Equal(sourceK8sResourceStatuses, targetK8sResourceStatuses)
	if k8sResourceStatusesEqual {
		r.logger.Info(fmt.Sprintf("IAM Role statuses of AWSIAMProvision equal: %s", r.request.NamespacedName))
	} else {
		r.logger.Info(fmt.Sprintf("IAM Role statuses of AWSIAMProvision different: %s", r.request.NamespacedName))
	}

	// Check all conditions indicating the resource or its status are actually updated
	if awsIAMProvision.Status.Phase != "Provisioned" || awsIAMProvisionProvisioned || !k8sResourceStatusesEqual {
		// Resources have been provisioned successfully
		r.logger.Info(fmt.Sprintf("AWSIAMProvision provisioned: %s", r.request.NamespacedName))

		if err := r.updateCRDStatus(awsIAMProvision, "Provisioned", "", sourceK8sResourceStatuses); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: frequency}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSIAMProvisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iamv1alpha1.AWSIAMProvision{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Owns(&iamctrlv1alpha1.Role{}). // trigger the r.Reconcile whenever an Own-ed resource is created/updated/deleted
		Complete(r)
}
