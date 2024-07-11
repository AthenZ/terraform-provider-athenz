package athenz

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceAllDomainDetails() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAllDomainDetailsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"gcp_project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"gcp_project_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"azure_subscription": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"azure_client": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"azure_tenant": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "description for the domain",
				Optional:    true,
			},
			"org": {
				Type:        schema.TypeString,
				Description: "audit organization name for the domain",
				Optional:    true,
			},
			"application_id": {
				Type:        schema.TypeString,
				Description: "associated application id",
				Optional:    true,
			},
			"user_expiry_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "all user members in the domain will have specified max expiry days",
			},
			"token_expiry_mins": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "tokens issued for this domain will have specified max timeout in mins",
			},
			"service_cert_expiry_mins": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "service identity certs issued for this domain will have specified max timeout in mins",
			},
			"role_cert_expiry_mins": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "role certs issued for this domain will have specified max timeout in mins",
			},
			"service_expiry_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "all services in the domain roles will have specified max expiry days",
			},
			"group_expiry_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "all groups in the domain roles will have specified max expiry days",
			},
			"member_purge_expiry_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "purge role/group members with expiry date configured days in the past",
			},
			"user_authority_filter": {
				Type:        schema.TypeString,
				Description: "membership filtered based on user authority configured attributes",
				Optional:    true,
			},
			"business_service": {
				Type:        schema.TypeString,
				Description: "associated business service with domain",
				Optional:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: "string specifying the environment this domain is used in (production, staging, etc.)",
				Optional:    true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"contacts": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"role_list": {
				Type:        schema.TypeSet,
				Description: "set of all roles",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"policy_list": {
				Type:        schema.TypeSet,
				Description: "set of all policies",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"service_list": {
				Type:        schema.TypeSet,
				Description: "set of all services",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"group_list": {
				Type:        schema.TypeSet,
				Description: "set of all groups",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAllDomainDetailsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName := d.Get("name").(string)
	domain, err := zmsClient.GetDomain(domainName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return diag.Errorf("athenz domain %s not found, update your data source query", domainName)
		} else {
			return diag.Errorf("error retrieving Athenz domain: %s", v)
		}
	case rdl.Any:
		return diag.FromErr(err)
	}
	if domain == nil {
		return diag.Errorf("error retrieving Athenz domain: %s", domainName)
	}
	d.SetId(string(domain.Name))

	if domain.Account != "" {
		d.Set("aws_account_id", domain.Account)
	}
	if domain.GcpProject != "" && domain.GcpProjectNumber != "" {
		d.Set("gcp_project_id", domain.GcpProject)
		d.Set("gcp_project_number", domain.GcpProjectNumber)
	}
	if domain.AzureSubscription != "" {
		d.Set("azure_subscription", domain.AzureSubscription)
	}
	if domain.AzureTenant != "" {
		d.Set("azure_tenant", domain.AzureTenant)
	}
	if domain.AzureClient != "" {
		d.Set("azure_client", domain.AzureClient)
	}
	if domain.Org != "" {
		d.Set("org", domain.Org)
	}
	if domain.Description != "" {
		d.Set("description", domain.Description)
	}
	if domain.MemberExpiryDays != nil {
		d.Set("user_expiry_days", domain.MemberExpiryDays)
	}
	if domain.ApplicationId != "" {
		d.Set("application_id", domain.ApplicationId)
	}
	if domain.TokenExpiryMins != nil {
		d.Set("token_expiry_mins", domain.TokenExpiryMins)
	}
	if domain.ServiceCertExpiryMins != nil {
		d.Set("service_cert_expiry_mins", domain.ServiceCertExpiryMins)
	}
	if domain.RoleCertExpiryMins != nil {
		d.Set("role_cert_expiry_mins", domain.RoleCertExpiryMins)
	}
	if domain.ServiceExpiryDays != nil {
		d.Set("service_expiry_days", domain.ServiceExpiryDays)
	}
	if domain.GroupExpiryDays != nil {
		d.Set("group_expiry_days", domain.GroupExpiryDays)
	}
	if domain.MemberPurgeExpiryDays != nil {
		d.Set("member_purge_expiry_days", domain.MemberPurgeExpiryDays)
	}
	if domain.UserAuthorityFilter != "" {
		d.Set("user_authority_filter", domain.UserAuthorityFilter)
	}
	if domain.BusinessService != "" {
		d.Set("business_service", domain.BusinessService)
	}
	if domain.Environment != "" {
		d.Set("environment", domain.Environment)
	}
	if domain.Tags != nil {
		d.Set("tags", flattenTag(domain.Tags))
	}
	if domain.Contacts != nil {
		d.Set("contacts", domain.Contacts)
	}
	roleList, err := zmsClient.GetRoleList(domainName, nil, "")
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("role_list", convertEntityNameListToStringList(roleList.Names)); err != nil {
		return diag.FromErr(err)
	}
	policyList, err := zmsClient.GetPolicyList(domainName, nil, "")
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("policy_list", convertEntityNameListToStringList(policyList.Names)); err != nil {
		return diag.FromErr(err)
	}
	serviceList, err := zmsClient.GetServiceIdentityList(domainName, nil, "")
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("service_list", convertEntityNameListToStringList(serviceList.Names)); err != nil {
		return diag.FromErr(err)
	}
	groupList, err := zmsClient.GetGroups(domainName, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("group_list", getGroupsNames(groupList.List)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
