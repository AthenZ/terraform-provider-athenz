package athenz

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleRead,
		Schema:      dataSourceRoleSchema(),
	}
}

func dataSourceRoleRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	dn := d.Get("domain").(string)
	rn := d.Get("name").(string)
	fullResourceName := dn + ROLE_SEPARATOR + rn

	role, err := zmsClient.GetRole(dn, rn)

	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return diag.Errorf("athenz Role %s not found, update your data source query", fullResourceName)
		} else {
			return diag.Errorf("error retrieving Athenz Role: %s", v)
		}
	case rdl.Any:
		return diag.FromErr(err)
	}
	d.SetId(fullResourceName)

	if len(role.RoleMembers) > 0 {
		if err = d.Set("member", flattenRoleMembers(role.RoleMembers)); err != nil {
			return diag.FromErr(err)
		}
	}
	if len(role.Tags) > 0 {
		if err = d.Set("tags", flattenTag(role.Tags)); err != nil {
			return diag.FromErr(err)
		}
	}
	roleSettings := map[string]int{}
	if role.TokenExpiryMins != nil {
		roleSettings["token_expiry_mins"] = int(*role.TokenExpiryMins)
	}
	if role.CertExpiryMins != nil {
		roleSettings["cert_expiry_mins"] = int(*role.CertExpiryMins)
	}
	if role.MemberExpiryDays != nil {
		roleSettings["user_expiry_days"] = int(*role.MemberExpiryDays)
	}
	if role.MemberReviewDays != nil {
		roleSettings["user_review_days"] = int(*role.MemberReviewDays)
	}
	if role.GroupExpiryDays != nil {
		roleSettings["group_expiry_days"] = int(*role.GroupExpiryDays)
	}
	if role.GroupReviewDays != nil {
		roleSettings["group_review_days"] = int(*role.GroupReviewDays)
	}
	if role.ServiceExpiryDays != nil {
		roleSettings["service_expiry_days"] = int(*role.ServiceExpiryDays)
	}
	if role.ServiceReviewDays != nil {
		roleSettings["service_review_days"] = int(*role.ServiceReviewDays)
	}
	if role.MaxMembers != nil {
		roleSettings["max_members"] = int(*role.MaxMembers)
	}
	if len(roleSettings) > 0 {
		if err = d.Set("settings", flattenIntSettings(roleSettings)); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.SelfServe != nil {
		if err = d.Set("self_serve", *role.SelfServe); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.SelfRenew != nil {
		if err = d.Set("self_renew", *role.SelfRenew); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.SelfRenewMins != nil {
		if err = d.Set("self_renew_mins", int(*role.SelfRenewMins)); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.DeleteProtection != nil {
		if err = d.Set("delete_protection", *role.DeleteProtection); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.AuditEnabled != nil {
		if err = d.Set("audit_enabled", *role.AuditEnabled); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.Description != "" {
		if err = d.Set("description", role.Description); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.ReviewEnabled != nil {
		if err = d.Set("review_enabled", *role.ReviewEnabled); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.UserAuthorityFilter != "" {
		if err = d.Set("user_authority_filter", role.UserAuthorityFilter); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.UserAuthorityExpiration != "" {
		if err = d.Set("user_authority_expiration", role.UserAuthorityExpiration); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.SignAlgorithm != "" {
		if err = d.Set("sign_algorithm", role.SignAlgorithm); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.NotifyRoles != "" {
		if err = d.Set("notify_roles", role.NotifyRoles); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.NotifyDetails != "" {
		if err = d.Set("notify_details", role.NotifyDetails); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.Trust != "" {
		if err = d.Set("trust", string(role.Trust)); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.LastReviewedDate != nil {
		if err = d.Set("last_reviewed_date", timestampToString(role.LastReviewedDate)); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.PrincipalDomainFilter != "" {
		if err = d.Set("principal_domain_filter", role.PrincipalDomainFilter); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
