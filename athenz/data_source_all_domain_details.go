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
