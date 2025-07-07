package athenz

import (
	"context"
	"log"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/AthenZ/terraform-provider-athenz/client"

	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceGroupMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupMembersCreate,
		ReadContext:   resourceGroupMembersRead,
		UpdateContext: resourceGroupMembersUpdate,
		DeleteContext: resourceGroupMembersDelete,
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
			"member": {
				Type:        schema.TypeSet,
				Description: "Users or services to be added as members with attribute",
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

func resourceGroupMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

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
				return diag.Errorf("error adding group member: %v", err)
			}
		}
	}

	d.SetId(fullResourceName)
	return readAfterWrite(resourceGroupMembersRead, ctx, d, meta)
}

func resourceGroupMembersRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if len(group.GroupMembers) > 0 {
		if err = d.Set("member", flattenGroupMembers(group.GroupMembers)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("member", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceGroupMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		oldVal, newVal := handleChange(d, "member")
		membersToDelete = append(membersToDelete, expandGroupMembers(oldVal.Difference(newVal).List())...)
		membersToAdd = append(membersToAdd, expandGroupMembers(newVal.Difference(oldVal).List())...)
	}

	err = updateGroupMembers(dn, gn, membersToDelete, membersToAdd, zmsClient, auditRef)
	if err != nil {
		return diag.Errorf("error updating group membership: %s", err)
	}

	return readAfterWrite(resourceGroupMembersRead, ctx, d, meta)
}

func resourceGroupMembersDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	group, err := zmsClient.GetGroup(dn, gn)
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
	for _, member := range group.GroupMembers {
		err = zmsClient.DeleteGroupMembership(dn, gn, member.MemberName, auditRef)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
