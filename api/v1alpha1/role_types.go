package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleSpec defines the desired state of Role.
//
// Contains information about an IAM role. This structure is returned as a response
// element in several API operations that interact with roles.
type RoleSpec struct {
	// The trust relationship policy document that grants an entity permission to
	// assume the role.
	//
	// In IAM, you must provide a JSON policy that has been converted to a string.
	// However, for CloudFormation templates formatted in YAML, you can provide
	// the policy in JSON or YAML format. CloudFormation always converts a YAML
	// policy to JSON format before submitting it to IAM.
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
	// Upon success, the response includes the same trust policy in JSON format.
	// +kubebuilder:validation:Required
	AssumeRolePolicyDocument *string `json:"assumeRolePolicyDocument"`
	// The name of the role to create.
	//
	// IAM user, group, role, and policy names must be unique within the account.
	// Names are not distinguished by case. For example, you cannot create resources
	// named both "MyResource" and "myresource".
	//
	// This parameter allows (through its regex pattern (http://wikipedia.org/wiki/regex))
	// a string of characters consisting of upper and lowercase alphanumeric characters
	// with no spaces. You can also include any of the following characters: _+=,.@-
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// A list of policies that you want to attach to the new role.
	Policies []*string `json:"policies,omitempty"`
	// A list of tags that you want to attach to the new role. Each tag consists
	// of a key name and an associated value. For more information about tagging,
	// see Tagging IAM resources (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html)
	// in the IAM User Guide.
	//
	// If any one of the tags is invalid or if you exceed the allowed maximum number
	// of tags, then the entire request fails and the resource is not created.
	Tags []*Tag `json:"tags,omitempty"`
}

// RoleStatus defines the observed state of Role
type RoleStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	AWSIAMResourceMetadata *AWSIAMResourceMetadata `json:"awsIAMResourceMetadata"`
	// The date and time, in ISO 8601 date-time format (http://www.iso.org/iso/iso8601),
	// when the role was created.
	// +kubebuilder:validation:Optional
	CreateDate *metav1.Time `json:"createDate,omitempty"`
	// The stable and unique string identifying the role. For more information about
	// IDs, see IAM identifiers (https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html)
	// in the IAM User Guide.
	// +kubebuilder:validation:Optional
	RoleID *string `json:"roleID,omitempty"`
}

type AWSIAMProvisionRole struct {
	Spec RoleSpec `json:"spec"`
}

// AWSIAMProvisionStatusRole defines the observed state of AWSIAMProvision's roles.
type AWSIAMProvisionStatusRole struct {
	Name    *string    `json:"name,omitempty"`
	Message string     `json:"message,omitempty"`
	Phase   string     `json:"phase,omitempty"`
	Status  RoleStatus `json:"status,omitempty"`
}
