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
	GetPolicies(domainName string, assertions bool, includeNonActive bool) (*zms.Policies, error)
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
	return zmsClient.GetPolicies(zms.DomainName(domainName), &assertions, &includeNonActive)
}

func (c Client) DeletePolicyVersion(domainName string, policyName string, version string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeletePolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), auditRef)
}

func (c Client) SetActivePolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.SetActivePolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), policyOptions, auditRef)
}

func (c Client) PutPolicyVersion(domainName string, policyName string, policyOptions *zms.PolicyOptions, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), policyOptions, auditRef)
}
func (c Client) GetPolicyVersion(domainName string, policyName string, version string) (*zms.Policy, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version))
}
func (c Client) GetPolicyVersionList(domainName string, policyName string) (*zms.PolicyList, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicyVersionList(zms.DomainName(domainName), zms.EntityName(policyName))
}

func (c Client) DeleteAssertionPolicyVersion(domainName string, policyName string, version string, assertionId int64, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteAssertionPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), assertionId, auditRef)
}
func (c Client) PutAssertionPolicyVersion(domainName string, policyName string, version string, auditRef string, assertion *zms.Assertion) (*zms.Assertion, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutAssertionPolicyVersion(zms.DomainName(domainName), zms.EntityName(policyName), zms.SimpleName(version), auditRef, assertion)
}

func (c Client) GetGroups(domainName string, members *bool) (*zms.Groups, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetGroups(zms.DomainName(domainName), members)
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
	return zmsClient.GetRoles(zms.DomainName(domainName), members, zms.CompoundName(tagKey), zms.CompoundName(tagValue))
}

func (c Client) GetRoleList(domainName string, limit *int32, skip string) (*zms.RoleList, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetRoleList(zms.DomainName(domainName), limit, skip)
}
func (c Client) PutDomainMeta(name string, auditRef string, detail *zms.DomainMeta) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutDomainMeta(zms.DomainName(name), auditRef, detail)
}
func (c Client) PostTopLevelDomain(auditRef string, detail *zms.TopLevelDomain) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PostTopLevelDomain(auditRef, detail)
}

func (c Client) DeleteTopLevelDomain(name string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteTopLevelDomain(zms.SimpleName(name), auditRef)
}

func (c Client) DeleteSubDomain(parentDomain string, subDomainName string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteSubDomain(zms.DomainName(parentDomain), zms.SimpleName(subDomainName), auditRef)
}
func (c Client) PostSubDomain(parentDomain string, auditRef string, detail *zms.SubDomain) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PostSubDomain(zms.DomainName(parentDomain), auditRef, detail)
}
func (c Client) DeleteUserDomain(domainName string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteUserDomain(zms.SimpleName(domainName), auditRef)
}

func (c Client) PostUserDomain(domainName string, auditRef string, detail *zms.UserDomain) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PostUserDomain(zms.SimpleName(domainName), auditRef, detail)
}

func (c Client) GetDomain(domainName string) (*zms.Domain, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetDomain(zms.DomainName(domainName))
}

func (c Client) PutServiceIdentity(domain string, serviceName string, auditRef string, detail *zms.ServiceIdentity) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName), auditRef, detail)
}

func (c Client) DeleteServiceIdentity(domain string, serviceName string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName), auditRef)
}

func (c Client) GetServiceIdentity(domain string, serviceName string) (*zms.ServiceIdentity, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(serviceName))
}
func (c Client) PutGroupMembership(domain string, groupName string, memberName zms.GroupMemberName, auditRef string, membership *zms.GroupMembership) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutGroupMembership(zms.DomainName(domain), zms.EntityName(groupName), memberName, auditRef, membership)
}

func (c Client) DeleteGroupMembership(domain string, groupName string, member zms.GroupMemberName, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteGroupMembership(zms.DomainName(domain), zms.EntityName(groupName), member, auditRef)
}

func (c Client) PutGroup(domain string, groupName string, auditRef string, group *zms.Group) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutGroup(zms.DomainName(domain), zms.EntityName(groupName), auditRef, group)
}

func (c Client) DeleteGroup(domain string, groupName string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteGroup(zms.DomainName(domain), zms.EntityName(groupName), auditRef)
}

func (c Client) GetGroup(domain string, groupName string) (*zms.Group, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetGroup(zms.DomainName(domain), zms.EntityName(groupName), nil, nil)
}

func (c Client) GetPolicy(domain string, policy string) (*zms.Policy, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetPolicy(zms.DomainName(domain), zms.EntityName(policy))
}

func (c Client) PutPolicy(domain string, policyName string, auditRef string, policy *zms.Policy) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutPolicy(zms.DomainName(domain), zms.EntityName(policyName), auditRef, policy)
}

func (c Client) DeletePolicy(domain string, policyName string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeletePolicy(zms.DomainName(domain), zms.EntityName(policyName), auditRef)
}

func (c Client) GetRole(domain string, roleName string) (*zms.Role, error) {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.GetRole(zms.DomainName(domain), zms.EntityName(roleName), nil, nil, nil)
}

func (c Client) PutRole(domain string, roleName string, auditRef string, role *zms.Role) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutRole(zms.DomainName(domain), zms.EntityName(roleName), auditRef, role)
}

func (c Client) DeleteRole(domain string, roleName string, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteRole(zms.DomainName(domain), zms.EntityName(roleName), auditRef)
}

func (c Client) PutMembership(domain string, roleName string, memberName zms.MemberName, auditRef string, membership *zms.Membership) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.PutMembership(zms.DomainName(domain), zms.EntityName(roleName), memberName, auditRef, membership)
}

func (c Client) DeleteMembership(domain string, roleMember string, member zms.MemberName, auditRef string) error {
	zmsClient := zms.NewClient(c.Url, c.Transport)
	return zmsClient.DeleteMembership(zms.DomainName(domain), zms.EntityName(roleMember), member, auditRef)
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
