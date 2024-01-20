package athenz

import (
	"context"
	"log"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"

	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRoleMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleMembersCreate,
		ReadContext:   resourceRoleMembersRead,
		UpdateContext: resourceRoleMembersUpdate,
		DeleteContext: resourceRoleMembersDelete,
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
			"member": {
				Type:        schema.TypeSet,
				Description: "Athenz principal to be added as members",
				Optional:    true,
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
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
		},
	}
}

func resourceRoleMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	rn := d.Get("name").(string)
	fullResourceName := dn + ROLE_SEPARATOR + rn

	_, err := zmsClient.GetRole(dn, rn)
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)

	if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		roleMembers := expandRoleMembers(v.(*schema.Set).List())
		for _, member := range roleMembers {
			membership := zms.Membership{
				MemberName:     member.MemberName,
				Expiration:     member.Expiration,
				ReviewReminder: member.ReviewReminder,
			}
			err = zmsClient.PutMembership(dn, rn, member.MemberName, auditRef, &membership)
			if err != nil {
				return diag.Errorf("error adding role member: %v", err)
			}
		}
	}

	d.SetId(fullResourceName)
	return readAfterWrite(resourceRoleMembersRead, ctx, d, meta)
}

func resourceRoleMembersRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if len(role.RoleMembers) > 0 {
		if err = d.Set("member", flattenRoleMembers(role.RoleMembers)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("member", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRoleMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	membersToDelete := make([]*zms.RoleMember, 0)
	membersToAdd := make([]*zms.RoleMember, 0)

	_, err = zmsClient.GetRole(dn, rn)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("member") {
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
		return diag.Errorf("error updating role membership: %s", err)
	}

	err = addRoleMembers(dn, rn, membersToAdd, auditRef, zmsClient)
	if err != nil {
		return diag.Errorf("error updating role membership: %s", err)
	}

	return readAfterWrite(resourceRoleMembersRead, ctx, d, meta)
}

func resourceRoleMembersDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	role, err := zmsClient.GetRole(dn, rn)
	if err != nil {
		switch v := err.(type) {
		case rdl.ResourceError:
			if v.Code == 404 {
				return nil
			}
			return diag.FromErr(err)
		case rdl.Any:
			return diag.FromErr(err)
		}
	}
	for _, member := range role.RoleMembers {
		err = zmsClient.DeleteMembership(dn, rn, member.MemberName, auditRef)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
