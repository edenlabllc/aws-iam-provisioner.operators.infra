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
	iamctrlv1alpha1 "github.com/aws-controllers-k8s/iam-controller/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AWSIAMProvisionRole struct {
	Spec iamctrlv1alpha1.RoleSpec `json:"spec"`
}

// AWSIAMProvisionSpec defines the desired state of AWSIAMProvision.
type AWSIAMProvisionSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	EksClusterName string                         `json:"eksClusterName"`
	Roles          map[string]AWSIAMProvisionRole `json:"roles"`
}

// AWSIAMProvisionStatusRole defines the observed state of AWSIAMProvision's roles.
type AWSIAMProvisionStatusRole struct {
	// Important: Run "make" to regenerate code after modifying this file

	Message string                     `json:"message,omitempty"`
	Phase   string                     `json:"phase,omitempty"`
	Status  iamctrlv1alpha1.RoleStatus `json:"status,omitempty"`
}

// AWSIAMProvisionStatus defines the observed state of AWSIAMProvision.
type AWSIAMProvisionStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	Message         string                               `json:"message,omitempty"`
	LastUpdatedTime *metav1.Time                         `json:"lastUpdatedTime,omitempty"`
	Phase           string                               `json:"phase,omitempty"`
	Roles           map[string]AWSIAMProvisionStatusRole `json:"roles,omitempty"`
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

func init() {
	SchemeBuilder.Register(&AWSIAMProvision{}, &AWSIAMProvisionList{})
}
