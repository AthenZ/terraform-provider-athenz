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

func ResourceGroupMeta() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupMetaCreate,
		ReadContext:   resourceGroupMetaRead,
		UpdateContext: resourceGroupMetaUpdate,
		DeleteContext: resourceGroupMetaDelete,
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
				Description:      "Name of the standard group",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(ENTITY_NAME),
			},
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
			"resource_state": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"principal_domain_filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func createNewGroupIfNecessary(zmsClient client.ZmsClient, dn, gn string) error {
	// if group exists already, we don't need to create it
	_, err := zmsClient.GetGroup(dn, gn)
	if err == nil {
		return nil
	}
	// only create the group if the return code was 404 - not found
	var v rdl.ResourceError
	switch {
	case errors.As(err, &v):
		if v.Code == 404 {
			group := zms.Group{
				Name: zms.ResourceName(gn),
			}
			return zmsClient.PutGroup(dn, gn, AUDIT_REF, &group)
		}
	}
	return err
}

func resourceGroupMetaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	gn := d.Get("name").(string)

	// if the group doesn't exist, we need to create it first
	// but only if the object_state is set to create if necessary
	if zmsClient.GetGroupMetaResourceState(d.Get("resource_state").(int), client.StateCreateIfNecessary) {
		err := createNewGroupIfNecessary(zmsClient, dn, gn)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// update our group meta data
	resp := updateGroupMeta(zmsClient, dn, gn, d)
	if resp != nil {
		return resp
	}
	fullResourceName := dn + GROUP_SEPARATOR + gn
	d.SetId(fullResourceName)
	return readAfterWrite(resourceGroupMetaRead, ctx, d, meta)
}

func updateGroupMeta(zmsClient client.ZmsClient, dn, gn string, d *schema.ResourceData) diag.Diagnostics {

	group, err := zmsClient.GetGroup(dn, gn)
	if err != nil {
		return diag.Errorf("unable to fetch group %s in domain %s: %v", gn, dn, err)
	}
	groupMeta := zms.GroupMeta{
		SelfServe:               group.SelfServe,
		MemberExpiryDays:        group.MemberExpiryDays,
		ServiceExpiryDays:       group.ServiceExpiryDays,
		ReviewEnabled:           group.ReviewEnabled,
		NotifyRoles:             group.NotifyRoles,
		NotifyDetails:           group.NotifyDetails,
		UserAuthorityFilter:     group.UserAuthorityFilter,
		UserAuthorityExpiration: group.UserAuthorityExpiration,
		Tags:                    group.Tags,
		DeleteProtection:        group.DeleteProtection,
		SelfRenew:               group.SelfRenew,
		SelfRenewMins:           group.SelfRenewMins,
		MaxMembers:              group.MaxMembers,
		AuditEnabled:            group.AuditEnabled,
	}
	selfServe := d.Get("self_serve").(bool)
	groupMeta.SelfServe = &selfServe
	if d.HasChange("user_expiry_days") {
		memberExpiryDays := int32(d.Get("user_expiry_days").(int))
		groupMeta.MemberExpiryDays = &memberExpiryDays
	}
	if d.HasChange("service_expiry_days") {
		serviceExpiryDays := int32(d.Get("service_expiry_days").(int))
		groupMeta.ServiceExpiryDays = &serviceExpiryDays
	}
	reviewEnabled := d.Get("review_enabled").(bool)
	groupMeta.ReviewEnabled = &reviewEnabled
	groupMeta.NotifyRoles = d.Get("notify_roles").(string)
	groupMeta.NotifyDetails = d.Get("notify_details").(string)
	groupMeta.PrincipalDomainFilter = d.Get("principal_domain_filter").(string)
	groupMeta.UserAuthorityFilter = d.Get("user_authority_filter").(string)
	groupMeta.UserAuthorityExpiration = d.Get("user_authority_expiration").(string)
	if d.HasChange("tags") {
		_, n := d.GetChange("tags")
		groupMeta.Tags = expandTagsMap(n.(map[string]interface{}))
	}
	deleteProtection := d.Get("delete_protection").(bool)
	groupMeta.DeleteProtection = &deleteProtection
	selfRenew := d.Get("self_renew").(bool)
	groupMeta.SelfRenew = &selfRenew
	if d.HasChange("self_renew_mins") {
		selfRenewMins := int32(d.Get("self_renew_mins").(int))
		groupMeta.SelfRenewMins = &selfRenewMins
	}
	if d.HasChange("max_members") {
		maxMembers := int32(d.Get("max_members").(int))
		groupMeta.MaxMembers = &maxMembers
	}
	auditEnabled := d.Get("audit_enabled").(bool)
	groupMeta.AuditEnabled = &auditEnabled
	auditRef := d.Get("audit_ref").(string)
	err = zmsClient.PutGroupMeta(dn, gn, auditRef, &groupMeta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceGroupMetaRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	if err != nil {
		return diag.FromErr(err)
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
	if group.MemberExpiryDays != nil {
		if err = d.Set("user_expiry_days", group.MemberExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.ServiceExpiryDays != nil {
		if err = d.Set("service_expiry_days", group.ServiceExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if group.MaxMembers != nil {
		if err = d.Set("max_members", group.MaxMembers); err != nil {
			return diag.FromErr(err)
		}
	}
	if err = d.Set("tags", flattenTag(group.Tags)); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("audit_enabled", group.AuditEnabled); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceGroupMetaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	resp := updateGroupMeta(zmsClient, dn, gn, d)
	if resp != nil {
		return resp
	}
	return readAfterWrite(resourceGroupMetaRead, ctx, d, meta)
}

func resourceGroupMetaDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zmsClient := meta.(client.ZmsClient)
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	if zmsClient.GetGroupMetaResourceState(d.Get("resource_state").(int), client.StateAlwaysDelete) {
		err = zmsClient.DeleteGroup(dn, gn, auditRef)
	} else {
		var zero int32
		zero = 0
		disabled := false
		groupMeta := zms.GroupMeta{
			SelfServe:               &disabled,
			MemberExpiryDays:        &zero,
			ServiceExpiryDays:       &zero,
			ReviewEnabled:           &disabled,
			NotifyRoles:             "",
			NotifyDetails:           "",
			UserAuthorityFilter:     "",
			UserAuthorityExpiration: "",
			Tags:                    make(map[zms.TagKey]*zms.TagValueList),
			DeleteProtection:        &disabled,
			SelfRenew:               &disabled,
			SelfRenewMins:           &zero,
			MaxMembers:              &zero,
			AuditEnabled:            &disabled,
			PrincipalDomainFilter:   "",
		}
		if v, ok := d.GetOk("tags"); ok {
			for key := range v.(map[string]interface{}) {
				groupMeta.Tags[zms.TagKey(key)] = &zms.TagValueList{List: []zms.TagCompoundValue{}}
			}
		}
		err = zmsClient.PutGroupMeta(dn, gn, auditRef, &groupMeta)
	}
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
