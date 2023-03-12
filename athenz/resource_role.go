package athenz

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"

	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				ValidateDiagFunc: validatePatternFunc(ENTTITY_NAME),
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
							ValidateFunc: validation.IntAtLeast(1),
						},
						"cert_expiry_mins": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"review": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							ValidateDiagFunc: validateReviewReminderPatternFunc(REVIEW_REMINDER_PATTERN, MEMBER_REVIEW_REMINDER),
						},
					},
				},
			},
			"trust": {
				Type:             schema.TypeString,
				Description:      "The domain, which this role is trusted to",
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
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
				role.Tags = expandRoleTags(v.(map[string]interface{}))
			}
			if v, ok := d.GetOk("trust"); ok {
				if len(role.RoleMembers) != 0 {
					return diag.Errorf("delegated roles cannot have members")
				}
				role.Trust = zms.DomainName(v.(string))
			}
			if v, ok := d.GetOk("settings"); ok && v.(*schema.Set).Len() > 0 {
				settings, ok := v.(*schema.Set).List()[0].(map[string]interface{})
				if ok {
					tokenExpiryMins := int32(settings["token_expiry_mins"].(int))
					certExpiryMins := int32(settings["cert_expiry_mins"].(int))

					role.TokenExpiryMins = &tokenExpiryMins
					role.CertExpiryMins = &certExpiryMins
				}
			}
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
			return diag.Errorf("the role %s is already exists in the domain %s use terraform import command", rn, dn)
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId(fullResourceName)

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
			log.Printf("[WARN] Athenz Role %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
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
	}
	// added for role tag
	if len(role.Tags) > 0 {
		if err = d.Set("tags", flattenTag(role.Tags)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("tags", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	zmsSettings := map[string]int{}
	if role.TokenExpiryMins != nil {
		zmsSettings["token_expiry_mins"] = int(*role.TokenExpiryMins)
	}
	if role.CertExpiryMins != nil {
		zmsSettings["cert_expiry_mins"] = int(*role.CertExpiryMins)
	}
	if len(zmsSettings) != 0 {
		if err = d.Set("settings", flattenRoleSettings(zmsSettings)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("settings", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	membersToDelete := make([]*zms.RoleMember, 0)
	membersToAdd := make([]*zms.RoleMember, 0)
	if d.HasChange("members") {
		if _, ok := d.GetOk("trust"); ok {
			return diag.Errorf("delegated roles cannot change members")
		}
		os, ns := handleChange(d, "members")
		membersToDelete = expandDeprecatedRoleMembers(os.Difference(ns).List())
		membersToAdd = expandDeprecatedRoleMembers(ns.Difference(os).List())
	}
	if d.HasChange("member") {
		if _, ok := d.GetOk("trust"); ok {
			return diag.Errorf("delegated roles cannot change members")
		}
		os, ns := handleChange(d, "member")
		membersToDelete = append(membersToDelete, expandRoleMembers(os.Difference(ns).List())...)
		membersToAdd = append(membersToAdd, expandRoleMembers(ns.Difference(os).List())...)
	}

	// we don't want to delete a member that should be added right after
	membersToNotDelete := stringSet{}
	for _, member := range membersToAdd {
		membersToNotDelete.add(string(member.MemberName))
	}

	err = deleteRoleMembers(dn, rn, membersToDelete, auditRef, zmsClient, membersToNotDelete)
	if err != nil {
		return diag.Errorf("error updating group membership: %s", err)
	}

	err = addRoleMembers(dn, rn, membersToAdd, auditRef, zmsClient)
	if err != nil {
		return diag.Errorf("error updating group membership: %s", err)
	}

	role, err := zmsClient.GetRole(dn, rn)
	if err != nil {
		return diag.FromErr(err)
	}
	isRoleChanged := false

	if d.HasChange("settings") {
		isRoleChanged = true
		_, n := d.GetChange("settings")
		if len(n.(*schema.Set).List()) != 0 {
			settings := n.(*schema.Set).List()[0].(map[string]interface{})
			tokenExpiryMins := int32(settings["token_expiry_mins"].(int))
			certExpiryMins := int32(settings["cert_expiry_mins"].(int))

			role.TokenExpiryMins = &tokenExpiryMins
			role.CertExpiryMins = &certExpiryMins
		} else {
			role.TokenExpiryMins = nil
			role.CertExpiryMins = nil
		}
	}

	if d.HasChange("tags") {
		isRoleChanged = true
		_, n := d.GetChange("tags")
		tags := expandRoleTags(n.(map[string]interface{}))
		role.Tags = tags
	}

	if isRoleChanged {
		err = zmsClient.PutRole(dn, rn, auditRef, role)
		if err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	err = zmsClient.DeleteRole(dn, rn, auditRef)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
