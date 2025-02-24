package aws_sdk

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamType "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/go-logr/logr"
	"golang.org/x/net/context"
)

const (
	IAMDescription = `Do not change the tag values, as this may affect work of the operator. If you need to add tags, do so through the AWSIAMProvision custom resource.`
	pathPrefix     = "/aws-iam-provisioner/"
)

type IAMManager interface {
	AttachRolePolicy(policyName, roleName *string) error
	BatchAttachDetachRolePolicies(proc string, policies []iamType.Policy, roleName *string) error
	BatchDeletePolicies(policies []iamType.Policy) error
	CreatePolicy(policyName, policyData, description *string, tags []iamType.Tag) (*iamType.Policy, error)
	CreateRole(roleName, assumeRolePolicyDocument, description *string, tags []iamType.Tag) (*iamType.Role, error)
	DeletePolicy(policyName *string) error
	DeleteRole(roleName *string) error
	DetachRolePolicy(policyName, roleName *string) error
	DiffRoleByPolicyDocument(rolePolicyDocumentA, rolePolicyDocumentB *string) (bool, error)
	GetIAMClientMetadata() *IAMClientMetadata
	GetPolicyByName(policyName *string) (*iamType.Policy, bool, error)
	GetRoleByName(roleName *string) (*iamType.Role, bool, error)
	ListAttachedRolePolicies(roleName *string) ([]iamType.Policy, error)
	ListEntitiesForPolicy(policy *iamType.Policy) ([]iamType.PolicyRole, error)
	ListPoliciesByTags(tags []iamType.Tag) ([]iamType.Policy, error)
	ListRolesByTags(tags []iamType.Tag) ([]iamType.Role, error)
	UpdateRole(roleName, assumeRolePolicyDocument *string) error
}

type IAMClient struct {
	Ctx       context.Context
	IAMClient *iam.Client
	*IAMClientMetadata
	Logger logr.Logger
}

type IAMClientMetadata struct {
	AccountID string
	Region    string
}

func NewIAMClient(region string, logger logr.Logger) (*IAMClient, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	identity, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	return &IAMClient{
		Ctx:               ctx,
		IAMClient:         iam.NewFromConfig(cfg),
		IAMClientMetadata: &IAMClientMetadata{AccountID: aws.ToString(identity.Account), Region: region},
		Logger:            logger,
	}, nil
}

func (c *IAMClient) GetIAMClientMetadata() *IAMClientMetadata {
	return c.IAMClientMetadata
}
