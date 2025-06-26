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

func ResourceSelfServeRoleMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSelfServeRoleMembersCreate,
		ReadContext:   resourceSelfServeRoleMembersRead,
		UpdateContext: resourceSelfServeRoleMembersUpdate,
		DeleteContext: resourceSelfServeRoleMembersDelete,
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
				Description:      "Name of the self-serve role",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(ENTITY_NAME),
			},
			"member": {
				Type:        schema.TypeSet,
				Description: "Athenz principal to be added as members (only manages members defined in Terraform)",
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

func resourceSelfServeRoleMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
				return diag.Errorf("error adding self-serve role member: %v", err)
			}
		}
	}

	d.SetId(fullResourceName)
	return readAfterWrite(resourceSelfServeRoleMembersRead, ctx, d, meta)
}

func resourceSelfServeRoleMembersRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	_, err = zmsClient.GetRole(dn, rn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			if !d.IsNewResource() {
				log.Printf("[WARN] Athenz self-serve role %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.Errorf("error retrieving Athenz self-serve role %s: %s", d.Id(), v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	// For self-serve role members, we only read the members that are in the Terraform state
	// We don't sync with the actual system members to avoid conflicts with externally managed members
	if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		// Keep the existing state as is - don't sync with system
		if err = d.Set("member", v); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("member", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceSelfServeRoleMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("error updating self-serve role membership: %s", err)
	}

	err = addRoleMembers(dn, rn, membersToAdd, auditRef, zmsClient)
	if err != nil {
		return diag.Errorf("error updating self-serve role membership: %s", err)
	}

	return readAfterWrite(resourceSelfServeRoleMembersRead, ctx, d, meta)
}

func resourceSelfServeRoleMembersDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, rn, err := splitRoleId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)

	// For self-serve role members, we only delete the members that are in the Terraform state
	// We don't delete all members from the system to avoid affecting externally managed members
	if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		roleMembers := expandRoleMembers(v.(*schema.Set).List())
		for _, member := range roleMembers {
			err = zmsClient.DeleteMembership(dn, rn, member.MemberName, auditRef)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}
