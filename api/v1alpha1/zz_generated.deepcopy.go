//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvision) DeepCopyInto(out *AWSIAMProvision) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvision.
func (in *AWSIAMProvision) DeepCopy() *AWSIAMProvision {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvision)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSIAMProvision) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvisionList) DeepCopyInto(out *AWSIAMProvisionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSIAMProvision, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvisionList.
func (in *AWSIAMProvisionList) DeepCopy() *AWSIAMProvisionList {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvisionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSIAMProvisionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvisionPolicy) DeepCopyInto(out *AWSIAMProvisionPolicy) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvisionPolicy.
func (in *AWSIAMProvisionPolicy) DeepCopy() *AWSIAMProvisionPolicy {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvisionPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvisionRole) DeepCopyInto(out *AWSIAMProvisionRole) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvisionRole.
func (in *AWSIAMProvisionRole) DeepCopy() *AWSIAMProvisionRole {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvisionRole)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvisionSpec) DeepCopyInto(out *AWSIAMProvisionSpec) {
	*out = *in
	if in.Frequency != nil {
		in, out := &in.Frequency, &out.Frequency
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Policies != nil {
		in, out := &in.Policies, &out.Policies
		*out = make(map[string]AWSIAMProvisionPolicy, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Roles != nil {
		in, out := &in.Roles, &out.Roles
		*out = make(map[string]AWSIAMProvisionRole, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvisionSpec.
func (in *AWSIAMProvisionSpec) DeepCopy() *AWSIAMProvisionSpec {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvisionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvisionStatus) DeepCopyInto(out *AWSIAMProvisionStatus) {
	*out = *in
	if in.LastUpdatedTime != nil {
		in, out := &in.LastUpdatedTime, &out.LastUpdatedTime
		*out = (*in).DeepCopy()
	}
	if in.Policies != nil {
		in, out := &in.Policies, &out.Policies
		*out = make([]AWSIAMProvisionStatusPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Roles != nil {
		in, out := &in.Roles, &out.Roles
		*out = make([]AWSIAMProvisionStatusRole, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvisionStatus.
func (in *AWSIAMProvisionStatus) DeepCopy() *AWSIAMProvisionStatus {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvisionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvisionStatusPolicy) DeepCopyInto(out *AWSIAMProvisionStatusPolicy) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvisionStatusPolicy.
func (in *AWSIAMProvisionStatusPolicy) DeepCopy() *AWSIAMProvisionStatusPolicy {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvisionStatusPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMProvisionStatusRole) DeepCopyInto(out *AWSIAMProvisionStatusRole) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMProvisionStatusRole.
func (in *AWSIAMProvisionStatusRole) DeepCopy() *AWSIAMProvisionStatusRole {
	if in == nil {
		return nil
	}
	out := new(AWSIAMProvisionStatusRole)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIAMResourceMetadata) DeepCopyInto(out *AWSIAMResourceMetadata) {
	*out = *in
	if in.ARN != nil {
		in, out := &in.ARN, &out.ARN
		*out = new(AWSResourceName)
		**out = **in
	}
	if in.OwnerAccountID != nil {
		in, out := &in.OwnerAccountID, &out.OwnerAccountID
		*out = new(AWSAccountID)
		**out = **in
	}
	if in.Region != nil {
		in, out := &in.Region, &out.Region
		*out = new(AWSRegion)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIAMResourceMetadata.
func (in *AWSIAMResourceMetadata) DeepCopy() *AWSIAMResourceMetadata {
	if in == nil {
		return nil
	}
	out := new(AWSIAMResourceMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicySpec) DeepCopyInto(out *PolicySpec) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.PolicyDocument != nil {
		in, out := &in.PolicyDocument, &out.PolicyDocument
		*out = new(string)
		**out = **in
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]*Tag, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Tag)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicySpec.
func (in *PolicySpec) DeepCopy() *PolicySpec {
	if in == nil {
		return nil
	}
	out := new(PolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyStatus) DeepCopyInto(out *PolicyStatus) {
	*out = *in
	if in.AWSIAMResourceMetadata != nil {
		in, out := &in.AWSIAMResourceMetadata, &out.AWSIAMResourceMetadata
		*out = new(AWSIAMResourceMetadata)
		(*in).DeepCopyInto(*out)
	}
	if in.AttachmentCount != nil {
		in, out := &in.AttachmentCount, &out.AttachmentCount
		*out = new(int32)
		**out = **in
	}
	if in.CreateDate != nil {
		in, out := &in.CreateDate, &out.CreateDate
		*out = (*in).DeepCopy()
	}
	if in.DefaultVersionID != nil {
		in, out := &in.DefaultVersionID, &out.DefaultVersionID
		*out = new(string)
		**out = **in
	}
	if in.PolicyID != nil {
		in, out := &in.PolicyID, &out.PolicyID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyStatus.
func (in *PolicyStatus) DeepCopy() *PolicyStatus {
	if in == nil {
		return nil
	}
	out := new(PolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoleSpec) DeepCopyInto(out *RoleSpec) {
	*out = *in
	if in.AssumeRolePolicyDocument != nil {
		in, out := &in.AssumeRolePolicyDocument, &out.AssumeRolePolicyDocument
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.Policies != nil {
		in, out := &in.Policies, &out.Policies
		*out = make([]*string, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(string)
				**out = **in
			}
		}
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]*Tag, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Tag)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoleSpec.
func (in *RoleSpec) DeepCopy() *RoleSpec {
	if in == nil {
		return nil
	}
	out := new(RoleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoleStatus) DeepCopyInto(out *RoleStatus) {
	*out = *in
	if in.AWSIAMResourceMetadata != nil {
		in, out := &in.AWSIAMResourceMetadata, &out.AWSIAMResourceMetadata
		*out = new(AWSIAMResourceMetadata)
		(*in).DeepCopyInto(*out)
	}
	if in.CreateDate != nil {
		in, out := &in.CreateDate, &out.CreateDate
		*out = (*in).DeepCopy()
	}
	if in.RoleID != nil {
		in, out := &in.RoleID, &out.RoleID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoleStatus.
func (in *RoleStatus) DeepCopy() *RoleStatus {
	if in == nil {
		return nil
	}
	out := new(RoleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tag) DeepCopyInto(out *Tag) {
	*out = *in
	if in.Key != nil {
		in, out := &in.Key, &out.Key
		*out = new(string)
		**out = **in
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tag.
func (in *Tag) DeepCopy() *Tag {
	if in == nil {
		return nil
	}
	out := new(Tag)
	in.DeepCopyInto(out)
	return out
}
