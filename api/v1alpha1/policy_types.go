package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicySpec defines the desired state of Policy.
//
// Contains information about a managed policy.
//
// This data type is used as a response element in the CreatePolicy, GetPolicy,
// and ListPolicies operations.
//
// For more information about managed policies, refer to Managed policies and
// inline policies (https://docs.aws.amazon.com/IAM/latest/UserGuide/policies-managed-vs-inline.html)
// in the IAM User Guide.
type PolicySpec struct {
	// The friendly name of the policy.
	//
	// IAM user, group, role, and policy names must be unique within the account.
	// Names are not distinguished by case. For example, you cannot create resources
	// named both "MyResource" and "myresource".
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// The JSON policy document that you want to use as the content for the new
	// policy.
	//
	// You must provide policies in JSON format in IAM. However, for CloudFormation
	// templates formatted in YAML, you can provide the policy in JSON or YAML format.
	// CloudFormation always converts a YAML policy to JSON format before submitting
	// it to IAM.
	//
	// The maximum length of the policy document that you can pass in this operation,
	// including whitespace, is listed below. To view the maximum character counts
	// of a managed policy with no whitespaces, see IAM and STS character quotas
	// (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-quotas.html#reference_iam-quotas-entity-length).
	//
	// To learn more about JSON policy grammar, see Grammar of the IAM JSON policy
	// language (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_grammar.html)
	// in the IAM User Guide.
	//
	// The regex pattern (http://wikipedia.org/wiki/regex) used to validate this
	// parameter is a string of characters consisting of the following:
	//
	//   - Any printable ASCII character ranging from the space character (\u0020)
	//     through the end of the ASCII character range
	//
	//   - The printable characters in the Basic Latin and Latin-1 Supplement character
	//     set (through \u00FF)
	//
	//   - The special characters tab (\u0009), line feed (\u000A), and carriage
	//     return (\u000D)
	//
	// +kubebuilder:validation:Required
	PolicyDocument *string `json:"policyDocument"`
	// A list of tags that you want to attach to the new IAM customer managed policy.
	// Each tag consists of a key name and an associated value. For more information
	// about tagging, see Tagging IAM resources (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html)
	// in the IAM User Guide.
	//
	// If any one of the tags is invalid or if you exceed the allowed maximum number
	// of tags, then the entire request fails and the resource is not created.
	Tags []*Tag `json:"tags,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	AWSIAMResourceMetadata *AWSIAMResourceMetadata `json:"awsIAMResourceMetadata"`
	// The number of entities (users, groups, and roles) that the policy is attached
	// to.
	// +kubebuilder:validation:Optional
	AttachmentCount *int32 `json:"attachmentCount,omitempty"`
	// The date and time, in ISO 8601 date-time format (http://www.iso.org/iso/iso8601),
	// when the policy was created.
	// +kubebuilder:validation:Optional
	CreateDate *metav1.Time `json:"createDate,omitempty"`
	// The identifier for the version of the policy that is set as the default version.
	// +kubebuilder:validation:Optional
	DefaultVersionID *string `json:"defaultVersionID,omitempty"`
	// The stable and unique string identifying the policy.
	//
	// For more information about IDs, see IAM identifiers (https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html)
	// in the IAM User Guide.
	// +kubebuilder:validation:Optional
	PolicyID *string `json:"policyID,omitempty"`
}

type AWSIAMProvisionPolicy struct {
	Spec PolicySpec `json:"spec"`
}

// AWSIAMProvisionStatusPolicy defines the observed state of AWSIAMProvision's policies.
type AWSIAMProvisionStatusPolicy struct {
	Name    *string      `json:"name,omitempty"`
	Message string       `json:"message,omitempty"`
	Phase   string       `json:"phase,omitempty"`
	Status  PolicyStatus `json:"status,omitempty"`
}
