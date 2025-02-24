package v1alpha1

// Tag A structure that represents user-provided metadata that can be associated
// with an IAM resource. For more information about tagging, see Tagging IAM
// resources (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html)
// in the IAM User Guide.
type Tag struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}
