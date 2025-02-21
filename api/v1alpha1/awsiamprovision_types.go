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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSIAMProvisionSpec defines the desired state of AWSIAMProvision.
type AWSIAMProvisionSpec struct {
	// EKSClusterName - target cluster name provisioned Cluster API.
	EKSClusterName string `json:"eksClusterName"`
	// Frequency - AWS IAM resources synchronization frequency.
	// It is not recommended to set values below 30s to avoid being blocked by the AWS API.
	Frequency *metav1.Duration `json:"frequency,omitempty"`
	// Region for AWS config authentication.
	Region string `json:"region"`
	// Policies - map of policies with specifications.
	Policies map[string]AWSIAMProvisionPolicy `json:"policies,omitempty"`
	// Roles - map of roles with specifications.
	Roles map[string]AWSIAMProvisionRole `json:"roles,omitempty"`
}

// AWSIAMProvisionStatus defines the observed state of AWSIAMProvision.
type AWSIAMProvisionStatus struct {
	Message         string                        `json:"message,omitempty"`
	LastUpdatedTime *metav1.Time                  `json:"lastUpdatedTime,omitempty"`
	Phase           string                        `json:"phase,omitempty"`
	Policies        []AWSIAMProvisionStatusPolicy `json:"policies,omitempty"`
	Roles           []AWSIAMProvisionStatusRole   `json:"roles,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PHASE",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="LAST-UPDATED-TIME",type=string,JSONPath=".status.lastUpdatedTime"

// AWSIAMProvision is the Schema for the awsiamprovisions API.
type AWSIAMProvision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSIAMProvisionSpec   `json:"spec,omitempty"`
	Status AWSIAMProvisionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSIAMProvisionList contains a list of AWSIAMProvision.
type AWSIAMProvisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSIAMProvision `json:"items"`
}

// AWSIAMResourceMetadata is common to all custom resources (CRs) managed by an ACK
// service controller. It is contained in the CR's `Status` member field and
// comprises various status and identifier fields useful to ACK for tracking
// state changes between Kubernetes and the backend AWS service API.
type AWSIAMResourceMetadata struct {
	// ARN is the Amazon Resource Name for the resource. This is a
	// globally-unique identifier and is set only by the ACK service controller
	// once the controller has orchestrated the creation of the resource OR
	// when it has verified that an "adopted" resource (a resource where the
	// ARN annotation was set by the Kubernetes user on the CR) exists and
	// matches the supplied CR's Spec field values.
	ARN *AWSResourceName `json:"arn,omitempty"`
	// OwnerAccountID is the AWS Account ID of the account that owns the
	// backend AWS service API resource.
	OwnerAccountID *AWSAccountID `json:"ownerAccountID"`
	// Region is the AWS region in which the resource exists or will exist.
	Region *AWSRegion `json:"region"`
}

// AWSRegion represents an AWS regional identifier
type AWSRegion string

// AWSAccountID represents an AWS account identifier
type AWSAccountID string

// AWSResourceName represents an AWS Resource Name (ARN)
type AWSResourceName string

func init() {
	SchemeBuilder.Register(&AWSIAMProvision{}, &AWSIAMProvisionList{})
}
