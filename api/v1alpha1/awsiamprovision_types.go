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

package v1alpha1

import (
	iamctrlv1alpha1 "github.com/aws-controllers-k8s/iam-controller/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AWSIAMProvisionRole struct {
	Spec iamctrlv1alpha1.RoleSpec `json:"spec"`
}

// AWSIAMProvisionSpec defines the desired state of AWSIAMProvision.
type AWSIAMProvisionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	EksClusterName string                         `json:"eksClusterName"`
	Role           map[string]AWSIAMProvisionRole `json:"role"`
}

// AWSIAMProvisionStatus defines the observed state of AWSIAMProvision.
type AWSIAMProvisionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Error           string       `json:"error,omitempty"`
	LastUpdatedTime *metav1.Time `json:"lastUpdatedTime,omitempty"`
	Phase           string       `json:"phase,omitempty"`
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
