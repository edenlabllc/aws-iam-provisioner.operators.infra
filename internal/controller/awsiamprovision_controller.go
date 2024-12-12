/*
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
*/

package controller

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	iamv1alpha1 "aws-iam-provisioner.operators.infra/api/v1alpha1"
)

const (
	frequency = time.Second * 10
)

// AWSIAMProvisionReconciler reconciles a AWSIAMProvision object
type AWSIAMProvisionReconciler struct {
	client.Client
	*reconciliationManager
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

	r.context = &ctx
	r.logger = &logger
	r.request = &req

	awsIAMProvision, eksControlPlane, err := r.getClusterResources()
	if err != nil {
		return ctrl.Result{}, err
	}

	if awsIAMProvision == nil || eksControlPlane == nil {
		// Resources not ready, re-queuing
		return ctrl.Result{RequeueAfter: frequency}, nil
	}

	provisioned := false
	for name, item := range awsIAMProvision.Spec.Roles {
		k8sResource, err := r.handleRole(awsIAMProvision, eksControlPlane, name, &item)

		if err != nil {
			return ctrl.Result{}, err
		}

		if k8sResource != nil {
			// If a resource has been returned, there was a change to it
			provisioned = true
		}
	}

	if awsIAMProvision.Status.Phase != "Provisioned" || provisioned {
		// Resources have been provisioned successfully
		r.logger.Info(fmt.Sprintf("AWSIAMProvision provisioned: %s", r.request.NamespacedName))

		if err := r.updateCRDStatus(awsIAMProvision, "Provisioned", ""); err != nil {
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
		Complete(r)
}
