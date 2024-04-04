package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/AthenZ/athenz/clients/go/zms"
)

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
}

type Client struct {
	Url       string
	Transport *http.Transport
}

type ZmsConfig struct {
	Url    string
	Cert   string
	Key    string
	CaCert string
}

func (c Client) GetPolicies(domainName string, assertions bool, includeNonActive bool) (*zms.Policies, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicies(zms.DomainName(domainName), &assertions, &includeNonActive, zms.TagKey(""), zms.TagCompoundValue(""))
}

func (c Client) DeletePolicyVersion(domainName string, policyName string, version string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeletePolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), auditRef, resourceOwner)
}

func (c Client) SetActivePolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.SetActivePolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), policyOptions, auditRef, resourceOwner)
}

func (c Client) PutPolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	retObject := false
	_, err := zmsClient.PutPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), policyOptions, auditRef, &retObject, resourceOwner)
	return err
}

func (c Client) GetPolicyVersion(domainName string, policyName string, version string) (*zms.Policy, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version))
}

func (c Client) GetPolicyVersionList(domainName string, policyName string) (*zms.PolicyList, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicyVersionList(zms.DomainName(domainName), zms.EntityName(policyName))
}

func (c Client) DeleteAssertionPolicyVersion(domainName string, policyName string, version string, assertionId int64, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteAssertionPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), assertionId, auditRef, resourceOwner)
}

func (c Client) PutAssertionPolicyVersion(domainName string, policyName string, version string, auditRef string, resourceOwner string, assertion *zms.Assertion) (*zms.Assertion, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutAssertionPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), auditRef, resourceOwner, assertion)
}

func (c Client) GetGroups(domainName string, members *bool) (*zms.Groups, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetGroups(zms.DomainName(domainName), members, "", "")
}

func (c Client) GetServiceIdentityList(domainName string, limit *int32, skip string) (*zms.ServiceIdentityList, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetServiceIdentityList(zms.DomainName(domainName), limit, skip)
}

func (c Client) GetPolicyList(domainName string, limit *int32, skip string) (*zms.PolicyList, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicyList(zms.DomainName(domainName), limit, skip)
}

func (c Client) GetRoles(domainName string, members *bool, tagKey string, tagValue string) (*zms.Roles, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetRoles(zms.DomainName(domainName), members, zms.TagKey(tagKey), zms.TagCompoundValue(tagValue))
}

func (c Client) GetRoleList(domainName string, limit *int32, skip string) (*zms.RoleList, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetRoleList(zms.DomainName(domainName), limit, skip)
}

func (c Client) PutDomainMeta(name string, auditRef string, resourceOwner string, detail *zms.DomainMeta) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutDomainMeta(zms.DomainName(name), auditRef, resourceOwner, detail)
}

func (c Client) PostTopLevelDomain(auditRef string, resourceOwner string, detail *zms.TopLevelDomain) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PostTopLevelDomain(auditRef, resourceOwner, detail)
}

func (c Client) DeleteTopLevelDomain(name string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteTopLevelDomain(zms.SimpleName(name), auditRef, resourceOwner)
}

func (c Client) DeleteSubDomain(parentDomain string, subDomainName string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteSubDomain(zms.DomainName(parentDomain), zms.SimpleName(subDomainName), auditRef, resourceOwner)
}

func (c Client) PostSubDomain(parentDomain string, auditRef string, resourceOwner string, detail *zms.SubDomain) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PostSubDomain(zms.DomainName(parentDomain), auditRef, resourceOwner, detail)
}

func (c Client) DeleteUserDomain(domainName string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteUserDomain(zms.SimpleName(domainName), auditRef, resourceOwner)
}

func (c Client) PostUserDomain(domainName string, auditRef string, resourceOwner string, detail *zms.UserDomain) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PostUserDomain(zms.SimpleName(domainName), auditRef, resourceOwner, detail)
}

func (c Client) GetDomain(domainName string) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetDomain(zms.DomainName(domainName))
}

func (c Client) PutServiceIdentity(domain string, serviceName string, auditRef string, resourceOwner string, detail *zms.ServiceIdentity) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	retObject := false
	_, err := zmsClient.PutServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName), auditRef, &retObject, resourceOwner, detail)
	return err
}

func (c Client) DeleteServiceIdentity(domain string, serviceName string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName), auditRef, resourceOwner)
}

func (c Client) GetServiceIdentity(domain string, serviceName string) (*zms.ServiceIdentity, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName))
}

func (c Client) PutGroupMembership(domain string, groupName string, memberName zms.GroupMemberName, auditRef string, resourceOwner string, membership *zms.GroupMembership) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	retObject := false
	_, err := zmsClient.PutGroupMembership(zms.DomainName(domain), zms.EntityName(groupName), memberName, auditRef, &retObject, resourceOwner, membership)
	return err
}

func (c Client) DeleteGroupMembership(domain string, groupName string, member zms.GroupMemberName, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteGroupMembership(zms.DomainName(domain), zms.EntityName(groupName), member, auditRef, resourceOwner)
}

func (c Client) PutGroup(domain string, groupName string, auditRef string, resourceOwner string, group *zms.Group) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	retObject := false
	_, err := zmsClient.PutGroup(zms.DomainName(domain), zms.EntityName(groupName), auditRef, &retObject, resourceOwner, group)
	return err
}

func (c Client) DeleteGroup(domain string, groupName string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteGroup(zms.DomainName(domain), zms.EntityName(groupName), auditRef, resourceOwner)
}

func (c Client) GetGroup(domain string, groupName string) (*zms.Group, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetGroup(zms.DomainName(domain), zms.EntityName(groupName), nil, nil)
}

func (c Client) GetPolicy(domain string, policy string) (*zms.Policy, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicy(zms.DomainName(domain), zms.EntityName(policy))
}

func (c Client) PutPolicy(domain string, policyName string, auditRef string, resourceOwner string, policy *zms.Policy) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	retObject := false
	_, err := zmsClient.PutPolicy(zms.DomainName(domain), zms.EntityName(policyName), auditRef, &retObject, resourceOwner, policy)
	return err
}

func (c Client) DeletePolicy(domain string, policyName string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeletePolicy(zms.DomainName(domain), zms.EntityName(policyName), auditRef, resourceOwner)
}

func (c Client) PutAssertionConditions(domainName string, policyName string, assertionId int64, auditRef string, resourceOwner string, assertionConditions *zms.AssertionConditions) (*zms.AssertionConditions, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutAssertionConditions(zms.DomainName(domainName), zms.EntityName(policyName), assertionId, auditRef, resourceOwner, assertionConditions)
}

func (c Client) GetRole(domain string, roleName string) (*zms.Role, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetRole(zms.DomainName(domain), zms.EntityName(roleName), nil, nil, nil)
}

func (c Client) PutRole(domain string, roleName string, auditRef string, resourceOwner string, role *zms.Role) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	retObject := false
	_, err := zmsClient.PutRole(zms.DomainName(domain), zms.EntityName(roleName), auditRef, &retObject, resourceOwner, role)
	return err
}

func (c Client) DeleteRole(domain string, roleName string, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteRole(zms.DomainName(domain), zms.EntityName(roleName), auditRef, resourceOwner)
}

func (c Client) PutMembership(domain string, roleName string, memberName zms.MemberName, auditRef string, resourceOwner string, membership *zms.Membership) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	retObject := false
	_, err := zmsClient.PutMembership(zms.DomainName(domain), zms.EntityName(roleName), memberName, auditRef, &retObject, resourceOwner, membership)
	return err
}

func (c Client) DeleteMembership(domain string, roleMember string, member zms.MemberName, auditRef string, resourceOwner string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteMembership(zms.DomainName(domain), zms.EntityName(roleMember), member, auditRef, resourceOwner)
}

func NewClient(url string, certFile string, keyFile string, caCert string) (*Client, error) {
	tlsConfig, err := getTLSConfigFromFiles(certFile, keyFile, caCert)
	if err != nil {
		return nil, err
	}
	transport := http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &Client{
		Url:       url,
		Transport: &transport,
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
		caCertPem, err := ioutil.ReadFile(caCert)
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

func (c Client) PutGroupMeta(domain string, groupName string, auditRef string, resourceOwner string, groupMeta *zms.GroupMeta) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	err := zmsClient.PutGroupMeta(zms.DomainName(domain), zms.EntityName(groupName), auditRef, resourceOwner, groupMeta)
	return err
}

func (c Client) PutRoleMeta(domain string, roleName string, auditRef string, resourceOwner string, roleMeta *zms.RoleMeta) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	err := zmsClient.PutRoleMeta(zms.DomainName(domain), zms.EntityName(roleName), auditRef, resourceOwner, roleMeta)
	return err
}
