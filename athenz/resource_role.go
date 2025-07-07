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

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:             schema.TypeString,
				Description:      "Name of the domain that role belongs to",
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
				Description: "Athenz principal to be added as members",
				Optional:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validatePatternFunc(MEMBER_NAME),
				},
				Set:           schema.HashString,
				ConflictsWith: []string{"member"},
				Deprecated:    "use member attribute instead",
			},
			"member": {
				Type:          schema.TypeSet,
				Description:   "Athenz principal to be added as members",
				Optional:      true,
				ConflictsWith: []string{"members"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validatePatternFunc(MEMBER_NAME),
						},
						"expiration": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							ValidateDiagFunc: validateDatePatternFunc(DATE_PATTERN, MEMBER_EXPIRATION),
						},
						"review": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							ValidateDiagFunc: validateDatePatternFunc(DATE_PATTERN, MEMBER_REVIEW_REMINDER),
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
						"token_expiry_mins": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"cert_expiry_mins": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"user_expiry_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"user_review_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"group_expiry_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"group_review_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"service_expiry_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"service_review_days": {
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
			"trust": {
				Type:             schema.TypeString,
				Description:      "The domain, which this role is trusted to",
				Optional:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"last_reviewed_date": {
				Type:             schema.TypeString,
				Description:      "The last reviewed timestamp for the role",
				Optional:         true,
				ValidateDiagFunc: validateDatePatternFunc(DATE_PATTERN, LAST_REVIEWED_DATE),
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
			"sign_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"principal_domain_filter": {
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
		CustomizeDiff: validateRoleSchema,
	}
}

func validateRoleSchema(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	_, mNew := d.GetChange("member")
	members := mNew.(*schema.Set).List()

	_, sNew := d.GetChange("settings")
	if len(sNew.(*schema.Set).List()) == 0 {
		return nil
	}
	settings := sNew.(*schema.Set).List()[0].(map[string]interface{})

	return validateRoleMember(members, settings)
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	rn := d.Get("name").(string)
	fullResourceName := dn + ROLE_SEPARATOR + rn

	roleCheck, err := zmsClient.GetRole(dn, rn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			role := zms.Role{
				Name:     zms.ResourceName(fullResourceName),
				Modified: nil,
			}
			if v, ok := d.GetOk("members"); ok {
				role.RoleMembers = expandDeprecatedRoleMembers(v.(*schema.Set).List())
			} else if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
				role.RoleMembers = expandRoleMembers(v.(*schema.Set).List())
			}
			auditRef := d.Get("audit_ref").(string)
			if v, ok := d.GetOk("tags"); ok {
				role.Tags = expandTagsMap(v.(map[string]interface{}))
			}
			if v, ok := d.GetOk("trust"); ok {
				if len(role.RoleMembers) != 0 {
					return diag.Errorf("delegated roles cannot have members")
				}
				role.Trust = zms.DomainName(v.(string))
			}
			if v, ok := d.GetOk("last_reviewed_date"); ok {
				role.LastReviewedDate = stringToTimestamp(v.(string))
			}
			if v, ok := d.GetOk("settings"); ok && v.(*schema.Set).Len() > 0 {
				settings, ok := v.(*schema.Set).List()[0].(map[string]interface{})
				if ok {
					tokenExpiryMins := int32(settings["token_expiry_mins"].(int))
					certExpiryMins := int32(settings["cert_expiry_mins"].(int))
					userExpiryDays := int32(settings["user_expiry_days"].(int))
					userReviewDays := int32(settings["user_review_days"].(int))
					groupExpiryDays := int32(settings["group_expiry_days"].(int))
					groupReviewDays := int32(settings["group_review_days"].(int))
					serviceExpiryDays := int32(settings["service_expiry_days"].(int))
					serviceReviewDays := int32(settings["service_review_days"].(int))
					maxMembers := int32(settings["max_members"].(int))

					role.TokenExpiryMins = &tokenExpiryMins
					role.CertExpiryMins = &certExpiryMins
					role.MemberExpiryDays = &userExpiryDays
					role.MemberReviewDays = &userReviewDays
					role.GroupExpiryDays = &groupExpiryDays
					role.GroupReviewDays = &groupReviewDays
					role.ServiceExpiryDays = &serviceExpiryDays
					role.ServiceReviewDays = &serviceReviewDays
					role.MaxMembers = &maxMembers
				}
			}
			role.PrincipalDomainFilter = d.Get("principal_domain_filter").(string)
			selfServe := d.Get("self_serve").(bool)
			role.SelfServe = &selfServe
			role.SignAlgorithm = d.Get("sign_algorithm").(string)
			reviewEnabled := d.Get("review_enabled").(bool)
			role.ReviewEnabled = &reviewEnabled
			role.NotifyRoles = d.Get("notify_roles").(string)
			role.NotifyDetails = d.Get("notify_details").(string)
			role.UserAuthorityFilter = d.Get("user_authority_filter").(string)
			role.UserAuthorityExpiration = d.Get("user_authority_expiration").(string)
			role.Description = d.Get("description").(string)
			deleteProtection := d.Get("delete_protection").(bool)
			role.DeleteProtection = &deleteProtection
			selfRenew := d.Get("self_renew").(bool)
			role.SelfRenew = &selfRenew
			if d.HasChange("self_renew_mins") {
				selfRenewMins := int32(d.Get("self_renew_mins").(int))
				role.SelfRenewMins = &selfRenewMins
			}
			auditEnabled := d.Get("audit_enabled").(bool)
			role.AuditEnabled = &auditEnabled
			err = zmsClient.PutRole(dn, rn, auditRef, &role)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(err)
		}
	case rdl.Any:
		return diag.FromErr(err)
	case nil:
		if roleCheck != nil {
			return diag.Errorf("the role %s already exists in the domain %s, use terraform import command", rn, dn)
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId(fullResourceName)
	return readAfterWrite(resourceRoleRead, ctx, d, meta)
}

func resourceRoleRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("domain", dn); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("name", rn); err != nil {
		return diag.FromErr(err)
	}
	role, err := zmsClient.GetRole(dn, rn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			if !d.IsNewResource() {
				log.Printf("[WARN] Athenz Role %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.Errorf("error retrieving Athenz Role %s: %s", d.Id(), v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	if role == nil {
		return diag.Errorf("error retrieving Athenz Role - Make sure your cert/key are valid")
	}
	if len(role.RoleMembers) > 0 {
		if _, ok := d.GetOk("members"); ok {
			if err = d.Set("members", flattenDeprecatedRoleMembers(role.RoleMembers)); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err = d.Set("member", flattenRoleMembers(role.RoleMembers)); err != nil {
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

	if role.Trust != "" {
		if err = d.Set("trust", string(role.Trust)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("trust", nil); err != nil {
			return diag.FromErr(err)
		}
	}
	// added for role tag
	if len(role.Tags) > 0 {
		if err = d.Set("tags", flattenTag(role.Tags)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		tags := d.Get("tags").(map[string]interface{})
		// if no tags in zms and there are tags configured, we have a drift,
		// so we set tags to empty map to let terraform know that tags need to be re added
		if len(tags) > 0 {
			if err = d.Set("tags", nil); err != nil {
				return diag.FromErr(err)
			}
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
	if len(roleSettings) != 0 {
		if err = d.Set("settings", flattenIntSettings(roleSettings)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if hasNoRoleSettings(d) {
			if err = d.Set("settings", nil); err != nil {
				return diag.FromErr(err)
			}
		} else {
			roleSettings = emptyRoleSettings()
			if err = d.Set("settings", flattenIntSettings(roleSettings)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err = d.Set("user_authority_filter", role.UserAuthorityFilter); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("user_authority_expiration", role.UserAuthorityExpiration); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("notify_roles", role.NotifyRoles); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("notify_details", role.NotifyDetails); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("principal_domain_filter", role.PrincipalDomainFilter); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("sign_algorithm", role.SignAlgorithm); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("description", role.Description); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("self_serve", role.SelfServe); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("self_renew", role.SelfRenew); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("delete_protection", role.DeleteProtection); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("review_enabled", role.ReviewEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("audit_enabled", role.AuditEnabled); err != nil {
		return diag.FromErr(err)
	}
	if role.LastReviewedDate != nil {
		if err = d.Set("last_reviewed_date", timestampToString(role.LastReviewedDate)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func hasNoRoleSettings(d *schema.ResourceData) bool {
	isSettingsNotInResourceData := len(d.Get("settings").(*schema.Set).List()) == 0
	isSettingsNotInState := d.GetRawState().IsNull() || d.GetRawState().AsValueMap()["settings"].AsValueSet().Values() == nil

	return isSettingsNotInResourceData && isSettingsNotInState
}

func emptyRoleSettings() map[string]int {
	roleSettings := map[string]int{}
	roleSettings["token_expiry_mins"] = 0
	roleSettings["cert_expiry_mins"] = 0
	roleSettings["user_expiry_days"] = 0
	roleSettings["user_review_days"] = 0
	roleSettings["group_expiry_days"] = 0
	roleSettings["group_review_days"] = 0
	roleSettings["service_expiry_days"] = 0
	roleSettings["service_review_days"] = 0
	roleSettings["max_members"] = 0
	return roleSettings
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)

	role, err := zmsClient.GetRole(dn, rn)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("settings") {
		_, n := d.GetChange("settings")
		if len(n.(*schema.Set).List()) != 0 {
			settings := n.(*schema.Set).List()[0].(map[string]interface{})
			tokenExpiryMins := int32(settings["token_expiry_mins"].(int))
			certExpiryMins := int32(settings["cert_expiry_mins"].(int))
			userExpiryDays := int32(settings["user_expiry_days"].(int))
			userReviewDays := int32(settings["user_review_days"].(int))
			groupExpiryDays := int32(settings["group_expiry_days"].(int))
			groupReviewDays := int32(settings["group_review_days"].(int))
			serviceExpiryDays := int32(settings["service_expiry_days"].(int))
			serviceReviewDays := int32(settings["service_review_days"].(int))
			maxMembers := int32(settings["max_members"].(int))

			role.TokenExpiryMins = &tokenExpiryMins
			role.CertExpiryMins = &certExpiryMins
			role.MemberExpiryDays = &userExpiryDays
			role.MemberReviewDays = &userReviewDays
			role.GroupExpiryDays = &groupExpiryDays
			role.GroupReviewDays = &groupReviewDays
			role.ServiceExpiryDays = &serviceExpiryDays
			role.ServiceReviewDays = &serviceReviewDays
			role.MaxMembers = &maxMembers
		} else {
			role.TokenExpiryMins = nil
			role.CertExpiryMins = nil
			role.MemberExpiryDays = nil
			role.MemberReviewDays = nil
			role.GroupExpiryDays = nil
			role.GroupReviewDays = nil
			role.ServiceExpiryDays = nil
			role.ServiceReviewDays = nil
			role.MaxMembers = nil
		}
	}

	if d.HasChange("tags") {
		_, n := d.GetChange("tags")
		tags := expandTagsMap(n.(map[string]interface{}))
		role.Tags = tags
	}

	if v, ok := d.GetOk("members"); ok {
		role.RoleMembers = expandDeprecatedRoleMembers(v.(*schema.Set).List())
	} else if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		role.RoleMembers = expandRoleMembers(v.(*schema.Set).List())
	} else {
		role.RoleMembers = nil
	}
	if v, ok := d.GetOk("trust"); ok {
		if len(role.RoleMembers) != 0 {
			return diag.Errorf("delegated roles cannot have members")
		}
		role.Trust = zms.DomainName(v.(string))
	} else {
		role.Trust = ""
	}
	if d.HasChange("last_reviewed_date") {
		role.LastReviewedDate = stringToTimestamp(d.Get("last_reviewed_date").(string))
	}
	if d.HasChange("principal_domain_filter") {
		role.PrincipalDomainFilter = d.Get("principal_domain_filter").(string)
	}
	if d.HasChange("self_serve") {
		selfServe := d.Get("self_serve").(bool)
		role.SelfServe = &selfServe
	}
	if d.HasChange("sign_algorithm") {
		role.SignAlgorithm = d.Get("sign_algorithm").(string)
	}
	if d.HasChange("review_enabled") {
		reviewEnabled := d.Get("review_enabled").(bool)
		role.ReviewEnabled = &reviewEnabled
	}
	if d.HasChange("notify_roles") {
		role.NotifyRoles = d.Get("notify_roles").(string)
	}
	if d.HasChange("notify_details") {
		role.NotifyDetails = d.Get("notify_details").(string)
	}
	if d.HasChange("user_authority_filter") {
		role.UserAuthorityFilter = d.Get("user_authority_filter").(string)
	}
	if d.HasChange("user_authority_expiration") {
		role.UserAuthorityExpiration = d.Get("user_authority_expiration").(string)
	}
	if d.HasChange("description") {
		role.Description = d.Get("description").(string)
	}
	if d.HasChange("delete_protection") {
		deleteProtection := d.Get("delete_protection").(bool)
		role.DeleteProtection = &deleteProtection
	}
	if d.HasChange("self_renew") {
		selfRenew := d.Get("self_renew").(bool)
		role.SelfRenew = &selfRenew
	}
	if d.HasChange("self_renew_mins") {
		selfRenewMins := int32(d.Get("self_renew_mins").(int))
		role.SelfRenewMins = &selfRenewMins
	}
	if d.HasChange("audit_enabled") {
		auditEnabled := d.Get("audit_enabled").(bool)
		role.AuditEnabled = &auditEnabled
	}

	err = zmsClient.PutRole(dn, rn, auditRef, role)
	if err != nil {
		return diag.Errorf("error updating role: %s", err)
	}

	return readAfterWrite(resourceRoleRead, ctx, d, meta)
}

func resourceRoleDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	err = zmsClient.DeleteRole(dn, rn, auditRef)

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
