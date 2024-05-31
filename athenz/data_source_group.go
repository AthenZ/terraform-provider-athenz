package athenz

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"member": {
				Type:        schema.TypeSet,
				Description: "Users or services to be added as members",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"expiration": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"settings": {
				Type:        schema.TypeSet,
				Description: "Advanced settings",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_expiry_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"service_expiry_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_members": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"self_serve": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"audit_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"self_renew": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"self_renew_mins": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"review_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"user_authority_filter": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"user_authority_expiration": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"notify_roles": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"last_reviewed_date": {
				Type:        schema.TypeString,
				Description: "Last reviewed date for the group",
				Optional:    true,
			},
			"principal_domain_filter": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func dataSourceGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	domainName := d.Get("domain").(string)
	groupName := d.Get("name").(string)
	fullResourceName := domainName + GROUP_SEPARATOR + groupName

	group, err := zmsClient.GetGroup(domainName, groupName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return diag.Errorf("athenz group %s not found, update your data source query", fullResourceName)
		} else {
			return diag.Errorf("error retrieving Athenz Group: %s", v)
		}
	case rdl.Any:
		return diag.FromErr(err)
	}
	d.SetId(fullResourceName)

	if len(group.GroupMembers) > 0 {
		if err = d.Set("member", flattenGroupMembers(group.GroupMembers)); err != nil {
			return diag.FromErr(err)
		}
	}

	if len(group.Tags) > 0 {
		if err = d.Set("tags", flattenTag(group.Tags)); err != nil {
			return diag.FromErr(err)
		}
	}
	groupSettings := map[string]int{}
	if group.MemberExpiryDays != nil {
		groupSettings["user_expiry_days"] = int(*group.MemberExpiryDays)
	}
	if group.ServiceExpiryDays != nil {
		groupSettings["service_expiry_days"] = int(*group.ServiceExpiryDays)
	}
	if group.MaxMembers != nil {
		groupSettings["max_members"] = int(*group.MaxMembers)
	}
	if len(groupSettings) > 0 {
		if err = d.Set("settings", flattenIntSettings(groupSettings)); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.SelfServe != nil {
		if err = d.Set("self_serve", *group.SelfServe); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.AuditEnabled != nil {
		if err = d.Set("audit_enabled", *group.AuditEnabled); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.SelfRenew != nil {
		if err = d.Set("self_renew", *group.SelfRenew); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.SelfRenewMins != nil {
		if err = d.Set("self_renew_mins", *group.SelfRenewMins); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.DeleteProtection != nil {
		if err = d.Set("delete_protection", *group.DeleteProtection); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.ReviewEnabled != nil {
		if err = d.Set("review_enabled", *group.ReviewEnabled); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.UserAuthorityFilter != "" {
		if err = d.Set("user_authority_filter", group.UserAuthorityFilter); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.UserAuthorityExpiration != "" {
		if err = d.Set("user_authority_expiration", group.UserAuthorityExpiration); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.NotifyRoles != "" {
		if err = d.Set("notify_roles", group.NotifyRoles); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.LastReviewedDate != nil {
		if err = d.Set("last_reviewed_date", timestampToString(group.LastReviewedDate)); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.PrincipalDomainFilter != "" {
		if err = d.Set("principal_domain_filter", group.NotifyRoles); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
