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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	iamv1alpha1 "aws-iam-provisioner.operators.infra/api/v1alpha1"
	"aws-iam-provisioner.operators.infra/internal/aws_sdk"
)

const (
	frequency = time.Second * 10
)

// AWSIAMProvisionReconciler reconciles a AWSIAMProvision object
type AWSIAMProvisionReconciler struct {
	*ReconciliationManager
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
	r.ctx = ctx
	r.logger = log.FromContext(ctx)
	r.request = req

	air, err := r.getClusterResources()
	if err != nil {
		return ctrl.Result{}, err
	}

	if air == nil {
		// Resources not ready, re-queuing
		return ctrl.Result{RequeueAfter: setTimer(air)}, nil
	}

	r.IAMClient, err = aws_sdk.NewIAMClient(air.awsIAMProvision.Spec.Region, r.logger)
	if err != nil {
		if err := r.updateCRDStatus(air, failPhase, "", err.Error(), nil); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if air.awsIAMProvision.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(air.awsIAMProvision, awsIAMProvisionFinalizerName) {
			controllerutil.AddFinalizer(air.awsIAMProvision, awsIAMProvisionFinalizerName)
			if err := r.Update(r.ctx, air.awsIAMProvision); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(air.awsIAMProvision, awsIAMProvisionFinalizerName) {
			if err := r.updateCRDStatus(air, destroyIntermediatePhase, "",
				"Destroying AWS IAM resources.", nil); err != nil {
				return ctrl.Result{}, err
			}

			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteIAMResources(air.awsIAMProvision); err != nil {
				if err := r.updateCRDStatus(air, failPhase, "", err.Error(), nil); err != nil {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			if err := r.updateCRDStatus(air, destroyPhase, "",
				"AWS IAM resources was destroyed.", nil); err != nil {
				return ctrl.Result{}, err
			}

			time.Sleep(time.Second * 3)

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(air.awsIAMProvision, awsIAMProvisionFinalizerName)
			if err := r.Update(r.ctx, air.awsIAMProvision); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if err := r.syncAWSIAMResources(air); err != nil {
		return ctrl.Result{}, err
	}

	for _, role := range air.awsIAMProvision.Spec.Roles {
		if err := r.syncRole(air, &role); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.syncPoliciesByRoleSpec(air, &role); err != nil {
			return ctrl.Result{}, err
		}
	}

	msg := fmt.Sprintf("AWS IAM resources synced with the remote state.")
	r.logger.Info(msg)
	if air.awsIAMProvision.Status.LastUpdatedTime == nil || air.awsIAMProvision.Status.Phase == "Failed" {
		if err := r.updateCRDStatus(air, successPhase, "", msg, nil); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: setTimer(air)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSIAMProvisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iamv1alpha1.AWSIAMProvision{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
