package athenz

import (
	"context"
	"errors"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/AthenZ/terraform-provider-athenz/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRoleMeta() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleMetaCreate,
		ReadContext:   resourceRoleMetaRead,
		UpdateContext: resourceRoleMetaUpdate,
		DeleteContext: resourceRoleMetaDelete,
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
				Description:      "Name of the standard role",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(ENTTITY_NAME),
			},
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
			"sign_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func createNewRoleIfNecessary(zmsClient client.ZmsClient, dn, rn string) error {
	// if role exists already, we don't need to create it
	_, err := zmsClient.GetRole(dn, rn)
	if err == nil {
		return nil
	}
	// only create the role if the return code was 404 - not found
	var v rdl.ResourceError
	switch {
	case errors.As(err, &v):
		if v.Code == 404 {
			role := zms.Role{
				Name: zms.ResourceName(rn),
			}
			return zmsClient.PutRole(dn, rn, AUDIT_REF, &role)
		}
	}
	return err
}

func resourceRoleMetaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	rn := d.Get("name").(string)

	// if the role doesn't exist, we need to create it first
	err := createNewRoleIfNecessary(zmsClient, dn, rn)
	if err != nil {
		return diag.FromErr(err)
	}

	// update our role meta data
	resp := updateRoleMeta(zmsClient, dn, rn, d)
	if resp != nil {
		return resp
	}
	fullResourceName := dn + ROLE_SEPARATOR + rn
	d.SetId(fullResourceName)
	return readAfterWrite(resourceRoleMetaRead, ctx, d, meta)
}

func updateRoleMeta(zmsClient client.ZmsClient, dn, rn string, d *schema.ResourceData) diag.Diagnostics {

	role, err := zmsClient.GetRole(dn, rn)
	if err != nil {
		return diag.Errorf("unable to fetch role %s in domain %s: %v", rn, dn, err)
	}
	roleMeta := zms.RoleMeta{
		SelfServe:               role.SelfServe,
		MemberExpiryDays:        role.MemberExpiryDays,
		TokenExpiryMins:         role.TokenExpiryMins,
		CertExpiryMins:          role.CertExpiryMins,
		SignAlgorithm:           role.SignAlgorithm,
		ServiceExpiryDays:       role.ServiceExpiryDays,
		MemberReviewDays:        role.MemberReviewDays,
		ServiceReviewDays:       role.ServiceReviewDays,
		ReviewEnabled:           role.ReviewEnabled,
		NotifyRoles:             role.NotifyRoles,
		UserAuthorityFilter:     role.UserAuthorityFilter,
		UserAuthorityExpiration: role.UserAuthorityExpiration,
		GroupExpiryDays:         role.GroupExpiryDays,
		GroupReviewDays:         role.GroupReviewDays,
		Tags:                    role.Tags,
		Description:             role.Description,
		DeleteProtection:        role.DeleteProtection,
		SelfRenew:               role.SelfRenew,
		SelfRenewMins:           role.SelfRenewMins,
		MaxMembers:              role.MaxMembers,
		AuditEnabled:            role.AuditEnabled,
	}
	selfServe := d.Get("self_serve").(bool)
	roleMeta.SelfServe = &selfServe
	if d.HasChange("user_expiry_days") {
		memberExpiryDays := int32(d.Get("user_expiry_days").(int))
		roleMeta.MemberExpiryDays = &memberExpiryDays
	}
	if d.HasChange("token_expiry_mins") {
		tokenExpiryMins := int32(d.Get("token_expiry_mins").(int))
		roleMeta.TokenExpiryMins = &tokenExpiryMins
	}
	if d.HasChange("cert_expiry_mins") {
		certExpiryMins := int32(d.Get("cert_expiry_mins").(int))
		roleMeta.CertExpiryMins = &certExpiryMins
	}
	roleMeta.SignAlgorithm = zms.SimpleName(d.Get("sign_algorithm").(string))
	if d.HasChange("service_expiry_days") {
		serviceExpiryDays := int32(d.Get("service_expiry_days").(int))
		roleMeta.ServiceExpiryDays = &serviceExpiryDays
	}
	if d.HasChange("user_review_days") {
		memberReviewDays := int32(d.Get("user_review_days").(int))
		roleMeta.MemberReviewDays = &memberReviewDays
	}
	if d.HasChange("service_review_days") {
		serviceReviewDays := int32(d.Get("service_review_days").(int))
		roleMeta.ServiceReviewDays = &serviceReviewDays
	}
	reviewEnabled := d.Get("review_enabled").(bool)
	roleMeta.ReviewEnabled = &reviewEnabled
	roleMeta.NotifyRoles = d.Get("notify_roles").(string)
	roleMeta.UserAuthorityFilter = d.Get("user_authority_filter").(string)
	roleMeta.UserAuthorityExpiration = d.Get("user_authority_expiration").(string)
	if d.HasChange("group_expiry_days") {
		groupExpiryDays := int32(d.Get("group_expiry_days").(int))
		roleMeta.GroupExpiryDays = &groupExpiryDays
	}
	if d.HasChange("group_review_days") {
		groupReviewDays := int32(d.Get("group_review_days").(int))
		roleMeta.GroupReviewDays = &groupReviewDays
	}
	if d.HasChange("tags") {
		_, n := d.GetChange("tags")
		roleMeta.Tags = expandTagsMap(n.(map[string]interface{}))
	}
	roleMeta.Description = d.Get("description").(string)
	deleteProtection := d.Get("delete_protection").(bool)
	roleMeta.DeleteProtection = &deleteProtection
	selfRenew := d.Get("self_renew").(bool)
	roleMeta.SelfRenew = &selfRenew
	if d.HasChange("self_renew_mins") {
		selfRenewMins := int32(d.Get("self_renew_mins").(int))
		roleMeta.SelfRenewMins = &selfRenewMins
	}
	if d.HasChange("max_members") {
		maxMembers := int32(d.Get("max_members").(int))
		roleMeta.MaxMembers = &maxMembers
	}
	auditEnabled := d.Get("audit_enabled").(bool)
	roleMeta.AuditEnabled = &auditEnabled
	auditRef := d.Get("audit_ref").(string)
	err = zmsClient.PutRoleMeta(dn, rn, auditRef, &roleMeta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRoleMetaRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	if err != nil {
		return diag.FromErr(err)
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
	if role.TokenExpiryMins != nil {
		if err = d.Set("token_expiry_mins", role.TokenExpiryMins); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.CertExpiryMins != nil {
		if err = d.Set("cert_expiry_mins", role.CertExpiryMins); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.MemberExpiryDays != nil {
		if err = d.Set("user_expiry_days", role.MemberExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.MemberReviewDays != nil {
		if err = d.Set("user_review_days", role.MemberReviewDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.GroupExpiryDays != nil {
		if err = d.Set("group_expiry_days", role.GroupExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.GroupReviewDays != nil {
		if err = d.Set("group_review_days", role.GroupReviewDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.ServiceExpiryDays != nil {
		if err = d.Set("service_expiry_days", role.ServiceExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.ServiceReviewDays != nil {
		if err = d.Set("service_review_days", role.ServiceReviewDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.MaxMembers != nil {
		if err = d.Set("max_members", role.MaxMembers); err != nil {
			return diag.FromErr(err)
		}
	}
	if err = d.Set("tags", flattenTag(role.Tags)); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("audit_enabled", role.AuditEnabled); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRoleMetaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	resp := updateRoleMeta(zmsClient, dn, rn, d)
	if resp != nil {
		return resp
	}
	return readAfterWrite(resourceRoleMetaRead, ctx, d, meta)
}

func resourceRoleMetaDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	var zero int32
	zero = 0
	disabled := false
	roleMeta := zms.RoleMeta{
		SelfServe:               &disabled,
		MemberExpiryDays:        &zero,
		TokenExpiryMins:         &zero,
		CertExpiryMins:          &zero,
		SignAlgorithm:           "",
		ServiceExpiryDays:       &zero,
		MemberReviewDays:        &zero,
		ServiceReviewDays:       &zero,
		ReviewEnabled:           &disabled,
		NotifyRoles:             "",
		UserAuthorityFilter:     "",
		UserAuthorityExpiration: "",
		GroupExpiryDays:         &zero,
		GroupReviewDays:         &zero,
		Tags:                    make(map[zms.TagKey]*zms.TagValueList),
		Description:             "",
		DeleteProtection:        &disabled,
		SelfRenew:               &disabled,
		SelfRenewMins:           &zero,
		MaxMembers:              &zero,
		AuditEnabled:            &disabled,
	}
	if v, ok := d.GetOk("tags"); ok {
		for key := range v.(map[string]interface{}) {
			roleMeta.Tags[zms.TagKey(key)] = &zms.TagValueList{List: []zms.TagCompoundValue{}}
		}
	}
	err = zmsClient.PutRoleMeta(dn, rn, auditRef, &roleMeta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
