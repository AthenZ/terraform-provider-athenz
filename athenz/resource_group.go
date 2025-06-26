package athenz

import (
	"context"
	"log"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:             schema.TypeString,
				Description:      "Name of the domain that group belongs to",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the standard group role",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(ENTITY_NAME),
			},
			"members": {
				Type:        schema.TypeSet,
				Description: "Users or services to be added as members",
				Optional:    true,
				Elem: &schema.Schema{Type: schema.TypeString,
					ValidateDiagFunc: validatePatternFunc(GROUP_MEMBER_NAME),
					Set:              schema.HashString,
				},
				ConflictsWith: []string{"member"},
				Deprecated:    "use member attribute instead",
			},
			"member": {
				Type:          schema.TypeSet,
				Description:   "Users or services to be added as members with attribute",
				Optional:      true,
				ConflictsWith: []string{"members"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validatePatternFunc(GROUP_MEMBER_NAME),
						},
						"expiration": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							ValidateDiagFunc: validateDatePatternFunc(DATE_PATTERN, MEMBER_EXPIRATION),
						},
					},
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
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"service_expiry_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"max_members": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"last_reviewed_date": {
				Type:             schema.TypeString,
				Description:      "The last reviewed timestamp for the group",
				Optional:         true,
				ValidateDiagFunc: validateDatePatternFunc(DATE_PATTERN, LAST_REVIEWED_DATE),
			},
			"principal_domain_filter": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
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
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
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
			},
			"user_authority_expiration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notify_roles": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notify_details": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	dn := d.Get("domain").(string)
	gn := d.Get("name").(string)
	fullResourceName := dn + GROUP_SEPARATOR + gn
	groupCheck, err := zmsClient.GetGroup(dn, gn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			group := zms.Group{
				Name:     zms.ResourceName(fullResourceName),
				Modified: nil,
			}
			if v, ok := d.GetOk("members"); ok {
				group.GroupMembers = expandDeprecatedGroupMembers(v.(*schema.Set).List())
			} else if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
				group.GroupMembers = expandGroupMembers(v.(*schema.Set).List())
			}
			if v, ok := d.GetOk("tags"); ok {
				group.Tags = expandTagsMap(v.(map[string]interface{}))
			}
			auditRef := d.Get("audit_ref").(string)
			if v, ok := d.GetOk("last_reviewed_date"); ok {
				group.LastReviewedDate = stringToTimestamp(v.(string))
			}
			if v, ok := d.GetOk("settings"); ok && v.(*schema.Set).Len() > 0 {
				settings, ok := v.(*schema.Set).List()[0].(map[string]interface{})
				if ok {
					userExpiryDays := int32(settings["user_expiry_days"].(int))
					serviceExpiryDays := int32(settings["service_expiry_days"].(int))
					maxMembers := int32(settings["max_members"].(int))

					group.MemberExpiryDays = &userExpiryDays
					group.ServiceExpiryDays = &serviceExpiryDays
					group.MaxMembers = &maxMembers
				}
			}
			selfServe := d.Get("self_serve").(bool)
			group.SelfServe = &selfServe
			reviewEnabled := d.Get("review_enabled").(bool)
			group.ReviewEnabled = &reviewEnabled
			group.NotifyRoles = d.Get("notify_roles").(string)
			group.NotifyDetails = d.Get("notify_details").(string)
			group.PrincipalDomainFilter = d.Get("principal_domain_filter").(string)
			group.UserAuthorityFilter = d.Get("user_authority_filter").(string)
			group.UserAuthorityExpiration = d.Get("user_authority_expiration").(string)
			deleteProtection := d.Get("delete_protection").(bool)
			group.DeleteProtection = &deleteProtection
			selfRenew := d.Get("self_renew").(bool)
			group.SelfRenew = &selfRenew
			selfRenewMins := int32(d.Get("self_renew_mins").(int))
			group.SelfRenewMins = &selfRenewMins
			auditEnabled := d.Get("audit_enabled").(bool)
			group.AuditEnabled = &auditEnabled
			if err = zmsClient.PutGroup(dn, gn, auditRef, &group); err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(err)
		}
	case rdl.Any:
		return diag.FromErr(err)
	case nil:
		if groupCheck != nil {
			return diag.Errorf("the group %s already exists in the domain %s, use terraform import command", gn, dn)
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId(fullResourceName)
	return readAfterWrite(resourceGroupRead, ctx, d, meta)
}

func resourceGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("domain", dn); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("name", gn); err != nil {
		return diag.FromErr(err)
	}

	group, err := zmsClient.GetGroup(dn, gn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			if !d.IsNewResource() {
				log.Printf("[WARN] Athenz Group %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.Errorf("error retrieving Athenz Group %s: %s", d.Id(), v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	if group == nil {
		return diag.Errorf("error retrieving Athenz Group - Make sure your cert/key are valid")
	}

	if len(group.GroupMembers) > 0 {
		if _, ok := d.GetOk("members"); ok {
			if err = d.Set("members", flattenDeprecatedGroupMembers(group.GroupMembers)); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err = d.Set("member", flattenGroupMembers(group.GroupMembers)); err != nil {
				return diag.FromErr(err)
			}
		}
	} else {
		if err = d.Set("members", nil); err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("member", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	if len(group.Tags) > 0 {
		if err = d.Set("tags", flattenTag(group.Tags)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("tags", nil); err != nil {
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

	if len(groupSettings) != 0 {
		if err = d.Set("settings", flattenIntSettings(groupSettings)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if hasNoGroupSettings(d) {
			if err = d.Set("settings", nil); err != nil {
				return diag.FromErr(err)
			}
		} else {
			groupSettings = emptyGroupSettings()
			if err = d.Set("settings", flattenIntSettings(groupSettings)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err = d.Set("user_authority_filter", group.UserAuthorityFilter); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("user_authority_expiration", group.UserAuthorityExpiration); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("notify_roles", group.NotifyRoles); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("notify_details", group.NotifyDetails); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("principal_domain_filter", group.PrincipalDomainFilter); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("self_serve", group.SelfServe); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("self_renew", group.SelfRenew); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("delete_protection", group.DeleteProtection); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("review_enabled", group.ReviewEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("audit_enabled", group.AuditEnabled); err != nil {
		return diag.FromErr(err)
	}
	if group.LastReviewedDate != nil {
		if err = d.Set("last_reviewed_date", timestampToString(group.LastReviewedDate)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	group, err := zmsClient.GetGroup(dn, gn)

	if err != nil {
		return diag.FromErr(err)
	}
	if v, ok := d.GetOk("members"); ok {
		group.GroupMembers = expandDeprecatedGroupMembers(v.(*schema.Set).List())
	} else if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		group.GroupMembers = expandGroupMembers(v.(*schema.Set).List())
	} else {
		group.GroupMembers = nil
	}

	if d.HasChange("tags") {
		_, n := d.GetChange("tags")
		tags := expandTagsMap(n.(map[string]interface{}))
		group.Tags = tags
	}
	if d.HasChange("last_reviewed_date") {
		group.LastReviewedDate = stringToTimestamp(d.Get("last_reviewed_date").(string))
	}
	if d.HasChange("settings") {
		_, n := d.GetChange("settings")
		if len(n.(*schema.Set).List()) != 0 {
			settings := n.(*schema.Set).List()[0].(map[string]interface{})
			userExpiryDays := int32(settings["user_expiry_days"].(int))
			serviceExpiryDays := int32(settings["service_expiry_days"].(int))
			maxMembers := int32(settings["max_members"].(int))

			group.MemberExpiryDays = &userExpiryDays
			group.ServiceExpiryDays = &serviceExpiryDays
			group.MaxMembers = &maxMembers
		} else {
			group.MemberExpiryDays = nil
			group.ServiceExpiryDays = nil
			group.MaxMembers = nil
		}
	}

	if d.HasChange("principal_domain_filter") {
		group.PrincipalDomainFilter = d.Get("principal_domain_filter").(string)
	}
	if d.HasChange("self_serve") {
		selfServe := d.Get("self_serve").(bool)
		group.SelfServe = &selfServe
	}
	if d.HasChange("review_enabled") {
		reviewEnabled := d.Get("review_enabled").(bool)
		group.ReviewEnabled = &reviewEnabled
	}
	if d.HasChange("notify_roles") {
		group.NotifyRoles = d.Get("notify_roles").(string)
	}
	if d.HasChange("notify_details") {
		group.NotifyDetails = d.Get("notify_details").(string)
	}
	if d.HasChange("user_authority_filter") {
		group.UserAuthorityFilter = d.Get("user_authority_filter").(string)
	}
	if d.HasChange("user_authority_expiration") {
		group.UserAuthorityExpiration = d.Get("user_authority_expiration").(string)
	}
	if d.HasChange("delete_protection") {
		deleteProtection := d.Get("delete_protection").(bool)
		group.DeleteProtection = &deleteProtection
	}
	if d.HasChange("self_renew") {
		selfRenew := d.Get("self_renew").(bool)
		group.SelfRenew = &selfRenew
	}
	if d.HasChange("self_renew_mins") {
		selfRenewMins := int32(d.Get("self_renew_mins").(int))
		group.SelfRenewMins = &selfRenewMins
	}
	if d.HasChange("audit_enabled") {
		auditEnabled := d.Get("audit_enabled").(bool)
		group.AuditEnabled = &auditEnabled
	}

	err = zmsClient.PutGroup(dn, gn, auditRef, group)
	if err != nil {
		return diag.Errorf("error updating group: %s", err)
	}

	return readAfterWrite(resourceGroupRead, ctx, d, meta)
}

func resourceGroupDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	err = zmsClient.DeleteGroup(dn, gn, auditRef)

	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return nil
		}
		return diag.FromErr(err)
	case rdl.Any:
		return diag.FromErr(err)
	}

	return nil
}

func hasNoGroupSettings(d *schema.ResourceData) bool {
	isSettingsNotInResourceData := len(d.Get("settings").(*schema.Set).List()) == 0
	isSettingsNotInState := d.GetRawState().IsNull() || d.GetRawState().AsValueMap()["settings"].AsValueSet().Values() == nil

	return isSettingsNotInResourceData && isSettingsNotInState
}

func emptyGroupSettings() map[string]int {
	groupSettings := map[string]int{}
	groupSettings["user_expiry_days"] = 0
	groupSettings["service_expiry_days"] = 0
	groupSettings["max_members"] = 0
	return groupSettings
}
