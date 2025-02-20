package aws_sdk

import (
	"crypto/sha1"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamType "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
)

const (
	TagKeyEKSClusterName = "aws.edenlab.io/aws-iam-provisioner/cluster"
	TagKeyNamespace      = "aws.edenlab.io/aws-iam-provisioner/namespace"
	TagKeyPolicyDocument = "aws.edenlab.io/aws-iam-provisioner/checksum"
)

func compareTags(tagsA, tagsB []iamType.Tag) bool {
	return cmp.Equal(tagsA, tagsB, cmp.AllowUnexported(iamType.Tag{}))
}

func NewChecksumTag(policyDocument *string) iamType.Tag {
	return iamType.Tag{
		Key:   aws.String(TagKeyPolicyDocument),
		Value: aws.String(fmt.Sprintf("%x", sha1.Sum([]byte(aws.ToString(policyDocument))))),
	}
}

func TagsDefine(clusterName, namespace string, tags ...iamType.Tag) []iamType.Tag {
	return append([]iamType.Tag{
		{
			Key:   aws.String(TagKeyEKSClusterName),
			Value: aws.String(clusterName),
		},
		{
			Key:   aws.String(TagKeyNamespace),
			Value: aws.String(namespace),
		},
	}, tags...)
}

func getSimilarTags(compareTags, resultTags []iamType.Tag) []iamType.Tag {
	var similarTags []iamType.Tag
	for _, tag := range compareTags {
		for _, resultTag := range resultTags {
			if *tag.Key == *resultTag.Key && *tag.Value == *resultTag.Value {
				similarTags = append(similarTags, resultTag)
			}
		}
	}

	return similarTags
}

func (c *IAMClient) getSimilarPolicyTags(compareTags []iamType.Tag, policy iamType.Policy) ([]iamType.Tag, error) {
	resultTags, err := c.IAMClient.ListPolicyTags(c.Ctx,
		&iam.ListPolicyTagsInput{
			MaxItems:  aws.Int32(10),
			PolicyArn: policy.Arn,
		},
	)
	if err != nil {
		return nil, err
	}

	return getSimilarTags(compareTags, resultTags.Tags), nil
}

func (c *IAMClient) getSimilarRoleTags(compareTags []iamType.Tag, role iamType.Role) ([]iamType.Tag, error) {
	resultTags, err := c.IAMClient.ListRoleTags(c.Ctx,
		&iam.ListRoleTagsInput{
			MaxItems: aws.Int32(10),
			RoleName: role.RoleName,
		},
	)
	if err != nil {
		return nil, err
	}

	return getSimilarTags(compareTags, resultTags.Tags), nil
}
