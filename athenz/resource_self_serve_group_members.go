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

func ResourceSelfServeGroupMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSelfServeGroupMembersCreate,
		ReadContext:   resourceSelfServeGroupMembersRead,
		UpdateContext: resourceSelfServeGroupMembersUpdate,
		DeleteContext: resourceSelfServeGroupMembersDelete,
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
				Description:      "Name of the self-serve group",
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
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
		},
	}
}

func resourceSelfServeGroupMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	gn := d.Get("name").(string)
	fullResourceName := dn + GROUP_SEPARATOR + gn

	_, err := zmsClient.GetGroup(dn, gn)
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)

	if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		groupMembers := expandGroupMembers(v.(*schema.Set).List())
		for _, member := range groupMembers {
			membership := zms.GroupMembership{
				MemberName: member.MemberName,
				Expiration: member.Expiration,
			}
			err = zmsClient.PutGroupMembership(dn, gn, member.MemberName, auditRef, &membership)
			if err != nil {
				return diag.Errorf("error adding self-serve group member: %v", err)
			}
		}
	}

	d.SetId(fullResourceName)
	return readAfterWrite(resourceSelfServeGroupMembersRead, ctx, d, meta)
}

func resourceSelfServeGroupMembersRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	_, err = zmsClient.GetGroup(dn, gn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			if !d.IsNewResource() {
				log.Printf("[WARN] Athenz self-serve group %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.Errorf("error retrieving Athenz self-serve group %s: %s", d.Id(), v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	// For self-serve group members, we only read the members that are in the Terraform state
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

func resourceSelfServeGroupMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	membersToDelete := make([]*zms.GroupMember, 0)
	membersToAdd := make([]*zms.GroupMember, 0)

	_, err = zmsClient.GetGroup(dn, gn)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("member") {
		os, ns := handleChange(d, "member")
		membersToDelete = append(membersToDelete, expandGroupMembers(os.Difference(ns).List())...)
		membersToAdd = append(membersToAdd, expandGroupMembers(ns.Difference(os).List())...)
	}

	err = updateGroupMembers(dn, gn, membersToDelete, membersToAdd, zmsClient, auditRef)
	if err != nil {
		return diag.Errorf("error updating self-serve group membership: %s", err)
	}

	return readAfterWrite(resourceSelfServeGroupMembersRead, ctx, d, meta)
}

func resourceSelfServeGroupMembersDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)

	// For self-serve group members, we only delete the members that are in the Terraform state
	// We don't delete all members from the system to avoid affecting externally managed members
	if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		groupMembers := expandGroupMembers(v.(*schema.Set).List())
		for _, member := range groupMembers {
			err = zmsClient.DeleteGroupMembership(dn, gn, member.MemberName, auditRef)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}
