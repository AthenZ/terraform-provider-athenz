package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/ardielle/ardielle-go/rdl"
)

const (
	StateCreateIfNecessary = 0x01
	StateAlwaysDelete      = 0x02

	ErrCodeRateLimit = 429
)

var retryDelays = []time.Duration{
	3 * time.Second,
	5 * time.Second,
	7 * time.Second,
}

type ZmsClient interface {
	GetRole(domain string, roleName string) (*zms.Role, error)
	DeleteRole(domain string, roleName string, auditRef string) error
	PutRole(domain string, roleName string, auditRef string, role *zms.Role) error
	PutMembership(domain string, roleName string, memberName zms.MemberName, auditRef string, membership *zms.Membership) error
	DeleteMembership(domain string, roleMember string, member zms.MemberName, auditRef string) error
	PutPolicy(domain string, policyName string, auditRef string, policy *zms.Policy) error
	GetPolicy(domain string, policy string) (*zms.Policy, error)
	DeletePolicy(domain string, policyName string, auditRef string) error
	GetGroup(domain string, groupName string) (*zms.Group, error)
	DeleteGroup(domain string, groupName string, auditRef string) error
	PutGroup(domain string, groupName string, auditRef string, group *zms.Group) error
	DeleteGroupMembership(domain string, groupName string, member zms.GroupMemberName, auditRef string) error
	PutGroupMembership(domain string, groupName string, memberName zms.GroupMemberName, auditRef string, membership *zms.GroupMembership) error
	GetServiceIdentity(domain string, serviceName string) (*zms.ServiceIdentity, error)
	PutServiceIdentity(domain string, serviceName string, auditRef string, detail *zms.ServiceIdentity) error
	PutServiceIdentitySystemMeta(domain string, serviceName string, attribute string, auditRef string, detail *zms.ServiceIdentitySystemMeta) error
	DeleteServiceIdentity(domain string, serviceName string, auditRef string) error
	GetDomain(domainName string) (*zms.Domain, error)
	PostUserDomain(domainName string, auditRef string, detail *zms.UserDomain) (*zms.Domain, error)
	DeleteUserDomain(domainName string, auditRef string) error
	PostSubDomain(parentDomain string, auditRef string, detail *zms.SubDomain) (*zms.Domain, error)
	DeleteSubDomain(parentDomain string, subDomainName string, auditRef string) error
	PostTopLevelDomain(auditRef string, detail *zms.TopLevelDomain) (*zms.Domain, error)
	DeleteTopLevelDomain(name string, auditRef string) error
	PutDomainMeta(name string, auditRef string, detail *zms.DomainMeta) error
	GetRoleList(domainName string, limit *int32, skip string) (*zms.RoleList, error)
	GetPolicyList(domainName string, limit *int32, skip string) (*zms.PolicyList, error)
	GetServiceIdentityList(domainName string, limit *int32, skip string) (*zms.ServiceIdentityList, error)
	GetGroups(domainName string, members *bool) (*zms.Groups, error)
	GetRoles(domainName string, members *bool, tagKey string, tagValue string) (*zms.Roles, error)
	PutPolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string) error
	PutAssertionPolicyVersion(domainName string, policyName string, version string, auditRef string, assertion *zms.Assertion) (*zms.Assertion, error)
	GetPolicyVersion(domainName string, policyName string, version string) (*zms.Policy, error)
	SetActivePolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string) error
	GetPolicyVersionList(domainName string, policyName string) (*zms.PolicyList, error)
	DeletePolicyVersion(domainName string, policyName string, version string, auditRef string) error
	DeleteAssertionPolicyVersion(domainName string, policyName string, version string, assertionId int64, auditRef string) error
	PutAssertionConditions(domainName string, policyName string, assertionId int64, auditRef string, assertionConditions *zms.AssertionConditions) (*zms.AssertionConditions, error)
	GetPolicies(domainName string, assertions bool, includeNonActive bool) (*zms.Policies, error)
	PutGroupMeta(domain string, groupName string, auditRef string, group *zms.GroupMeta) error
	PutRoleMeta(domain string, roleName string, auditRef string, group *zms.RoleMeta) error
	GetRoleMetaResourceState(roleMetaResourceState, requestedState int) bool
	GetGroupMetaResourceState(groupMetaResourceState, requestedState int) bool
}

type Client struct {
	Url                    string
	Transport              *http.Transport
	ResourceOwner          string
	RoleMetaResourceState  int
	GroupMetaResourceState int
}

type ZmsConfig struct {
	Url                    string
	Cert                   string
	Key                    string
	CaCert                 string
	ResourceOwner          string
	RoleMetaResourceState  int
	GroupMetaResourceState int
}

func (c Client) GetPolicies(domainName string, assertions bool, includeNonActive bool) (*zms.Policies, error) {
	var (
		policies *zms.Policies
		err      error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		policies, err = zmsClient.GetPolicies(zms.DomainName(domainName), &assertions, &includeNonActive, "", "")
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return policies, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeletePolicyVersion(domainName string, policyName string, version string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeletePolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) SetActivePolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.SetActivePolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), policyOptions, auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutPolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string) error {
	var err error
	retObject := false
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_, err = zmsClient.PutPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), policyOptions, auditRef, &retObject, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetPolicyVersion(domainName string, policyName string, version string) (*zms.Policy, error) {
	var (
		policy *zms.Policy
		err    error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		policy, err = zmsClient.GetPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version))
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return policy, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetPolicyVersionList(domainName string, policyName string) (*zms.PolicyList, error) {
	var (
		policyList *zms.PolicyList
		err        error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		policyList, err = zmsClient.GetPolicyVersionList(zms.DomainName(domainName), zms.EntityName(policyName))
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return policyList, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteAssertionPolicyVersion(domainName string, policyName string, version string, assertionId int64, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteAssertionPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), assertionId, auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutAssertionPolicyVersion(domainName string, policyName string, version string, auditRef string, assertion *zms.Assertion) (*zms.Assertion, error) {
	var (
		retAssertion *zms.Assertion
		err          error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		retAssertion, err = zmsClient.PutAssertionPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), auditRef, c.ResourceOwner, assertion)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return retAssertion, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetGroups(domainName string, members *bool) (*zms.Groups, error) {
	var (
		groups *zms.Groups
		err    error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		groups, err = zmsClient.GetGroups(zms.DomainName(domainName), members, "", "")
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return groups, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetServiceIdentityList(domainName string, limit *int32, skip string) (*zms.ServiceIdentityList, error) {
	var (
		serviceIdentityList *zms.ServiceIdentityList
		err                 error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		serviceIdentityList, err = zmsClient.GetServiceIdentityList(zms.DomainName(domainName), limit, skip)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return serviceIdentityList, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetPolicyList(domainName string, limit *int32, skip string) (*zms.PolicyList, error) {
	var (
		policyList *zms.PolicyList
		err        error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		policyList, err = zmsClient.GetPolicyList(zms.DomainName(domainName), limit, skip)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return policyList, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetRoles(domainName string, members *bool, tagKey string, tagValue string) (*zms.Roles, error) {
	var (
		roles *zms.Roles
		err   error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		roles, err = zmsClient.GetRoles(zms.DomainName(domainName), members, zms.TagKey(tagKey), zms.TagCompoundValue(tagValue))
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return roles, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetRoleList(domainName string, limit *int32, skip string) (*zms.RoleList, error) {
	var (
		roleList *zms.RoleList
		err      error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		roleList, err = zmsClient.GetRoleList(zms.DomainName(domainName), limit, skip)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return roleList, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutDomainMeta(name string, auditRef string, detail *zms.DomainMeta) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.PutDomainMeta(zms.DomainName(name), auditRef, c.ResourceOwner, detail)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PostTopLevelDomain(auditRef string, detail *zms.TopLevelDomain) (*zms.Domain, error) {
	var (
		domain *zms.Domain
		err    error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		domain, err = zmsClient.PostTopLevelDomain(auditRef, c.ResourceOwner, detail)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return domain, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteTopLevelDomain(name string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteTopLevelDomain(zms.SimpleName(name), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteSubDomain(parentDomain string, subDomainName string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteSubDomain(zms.DomainName(parentDomain), zms.SimpleName(subDomainName), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PostSubDomain(parentDomain string, auditRef string, detail *zms.SubDomain) (*zms.Domain, error) {
	var (
		domain *zms.Domain
		err    error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		domain, err = zmsClient.PostSubDomain(zms.DomainName(parentDomain), auditRef, c.ResourceOwner, detail)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return domain, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteUserDomain(domainName string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteUserDomain(zms.SimpleName(domainName), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PostUserDomain(domainName string, auditRef string, detail *zms.UserDomain) (*zms.Domain, error) {
	var (
		domain *zms.Domain
		err    error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		domain, err = zmsClient.PostUserDomain(zms.SimpleName(domainName), auditRef, c.ResourceOwner, detail)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return domain, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetDomain(domainName string) (*zms.Domain, error) {
	var (
		domain *zms.Domain
		err    error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		domain, err = zmsClient.GetDomain(zms.DomainName(domainName))
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return domain, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutServiceIdentity(domain string, serviceName string, auditRef string, detail *zms.ServiceIdentity) error {
	var err error
	retObject := false
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_, err = zmsClient.PutServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName), auditRef, &retObject, c.ResourceOwner, detail)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutServiceIdentitySystemMeta(domain string, serviceName string, attribute string, auditRef string, detail *zms.ServiceIdentitySystemMeta) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	err := zmsClient.PutServiceIdentitySystemMeta(zms.DomainName(domain), zms.SimpleName(serviceName), zms.SimpleName(attribute), auditRef, detail)
	return err
}

func (c Client) DeleteServiceIdentity(domain string, serviceName string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetServiceIdentity(domain string, serviceName string) (*zms.ServiceIdentity, error) {
	var (
		serviceIdentity *zms.ServiceIdentity
		err             error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		serviceIdentity, err = zmsClient.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName))
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return serviceIdentity, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutGroupMembership(domain string, groupName string, memberName zms.GroupMemberName, auditRef string, membership *zms.GroupMembership) error {
	var err error
	retObject := false
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_, err = zmsClient.PutGroupMembership(zms.DomainName(domain), zms.EntityName(groupName), memberName, auditRef, &retObject, c.ResourceOwner, membership)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteGroupMembership(domain string, groupName string, member zms.GroupMemberName, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteGroupMembership(zms.DomainName(domain), zms.EntityName(groupName), member, auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutGroup(domain string, groupName string, auditRef string, group *zms.Group) error {
	var err error
	retObject := false
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_, err = zmsClient.PutGroup(zms.DomainName(domain), zms.EntityName(groupName), auditRef, &retObject, c.ResourceOwner, group)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteGroup(domain string, groupName string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteGroup(zms.DomainName(domain), zms.EntityName(groupName), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetGroup(domain string, groupName string) (*zms.Group, error) {
	var (
		group *zms.Group
		err   error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		group, err = zmsClient.GetGroup(zms.DomainName(domain), zms.EntityName(groupName), nil, nil)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return group, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetPolicy(domain string, policy string) (*zms.Policy, error) {
	var (
		retPolicy *zms.Policy
		err       error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		retPolicy, err = zmsClient.GetPolicy(zms.DomainName(domain), zms.EntityName(policy))
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return retPolicy, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutPolicy(domain string, policyName string, auditRef string, policy *zms.Policy) error {
	var err error
	retObject := false
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_, err = zmsClient.PutPolicy(zms.DomainName(domain), zms.EntityName(policyName), auditRef, &retObject, c.ResourceOwner, policy)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeletePolicy(domain string, policyName string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeletePolicy(zms.DomainName(domain), zms.EntityName(policyName), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutAssertionConditions(domainName string, policyName string, assertionId int64, auditRef string, assertionConditions *zms.AssertionConditions) (*zms.AssertionConditions, error) {
	var (
		retAssertionConditions *zms.AssertionConditions
		err                    error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		retAssertionConditions, err = zmsClient.PutAssertionConditions(zms.DomainName(domainName), zms.EntityName(policyName), assertionId, auditRef, c.ResourceOwner, assertionConditions)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return retAssertionConditions, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetRole(domain string, roleName string) (*zms.Role, error) {
	var (
		role *zms.Role
		err  error
	)
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		role, err = zmsClient.GetRole(zms.DomainName(domain), zms.EntityName(roleName), nil, nil, nil)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return role, err
	}
	return nil, fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutRole(domain string, roleName string, auditRef string, role *zms.Role) error {
	var err error
	retObject := false
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_, err = zmsClient.PutRole(zms.DomainName(domain), zms.EntityName(roleName), auditRef, &retObject, c.ResourceOwner, role)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteRole(domain string, roleName string, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteRole(zms.DomainName(domain), zms.EntityName(roleName), auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutMembership(domain string, roleName string, memberName zms.MemberName, auditRef string, membership *zms.Membership) error {
	var err error
	retObject := false
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_, err = zmsClient.PutMembership(zms.DomainName(domain), zms.EntityName(roleName), memberName, auditRef, &retObject, c.ResourceOwner, membership)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) DeleteMembership(domain string, roleMember string, member zms.MemberName, auditRef string) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.DeleteMembership(zms.DomainName(domain), zms.EntityName(roleMember), member, auditRef, c.ResourceOwner)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutGroupMeta(domain string, groupName string, auditRef string, groupMeta *zms.GroupMeta) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.PutGroupMeta(zms.DomainName(domain), zms.EntityName(groupName), auditRef, c.ResourceOwner, groupMeta)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) PutRoleMeta(domain string, roleName string, auditRef string, roleMeta *zms.RoleMeta) error {
	var err error
	zmsClient := zms.NewClient(c.Url, c.Transport)
	for _, delay := range append([]time.Duration{0}, retryDelays...) {
		if delay > 0 {
			time.Sleep(delay)
		}
		err = zmsClient.PutRoleMeta(zms.DomainName(domain), zms.EntityName(roleName), auditRef, c.ResourceOwner, roleMeta)
		if errObj, ok := err.(rdl.ResourceError); ok && errObj.Code == ErrCodeRateLimit {
			continue
		}
		return err
	}
	return fmt.Errorf("too many requests, retried 3 times but still failed: %w", err)
}

func (c Client) GetRoleMetaResourceState(roleMetaResourceState, requestedState int) bool {
	return getResourceState(roleMetaResourceState, c.RoleMetaResourceState, requestedState)
}

func (c Client) GetGroupMetaResourceState(groupMetaResourceState, requestedState int) bool {
	return getResourceState(groupMetaResourceState, c.GroupMetaResourceState, requestedState)
}

func getResourceState(resourceState, clientState, requestedState int) bool {
	if resourceState == -1 {
		resourceState = clientState
	}
	if resourceState == -1 {
		return false
	}
	return (resourceState & requestedState) != 0
}

func NewClient(zmsConfig *ZmsConfig) (*Client, error) {
	tlsConfig, err := getTLSConfigFromFiles(zmsConfig.Cert, zmsConfig.Key, zmsConfig.CaCert)
	if err != nil {
		return nil, err
	}
	transport := http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &Client{
		Url:                    zmsConfig.Url,
		Transport:              &transport,
		ResourceOwner:          zmsConfig.ResourceOwner,
		RoleMetaResourceState:  zmsConfig.RoleMetaResourceState,
		GroupMetaResourceState: zmsConfig.GroupMetaResourceState,
	}
	return client, err
}

func getTLSConfigFromFiles(certFile, keyFile string, caCert string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to formulate clientCert from key and cert bytes, error: %v", err)
	}

	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0] = cert

	if caCert != "" {
		caCertPem, err := os.ReadFile(caCert)
		if err != nil {
			return nil, fmt.Errorf("unable to cacert file, error: %v", err)
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCertPem)
		config.RootCAs = certPool
	}

	// Set Renegotiation explicitly
	config.Renegotiation = tls.RenegotiateOnceAsClient

	return config, err
}
