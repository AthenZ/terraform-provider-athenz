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
				ValidateDiagFunc: validatePatternFunc(ENTTITY_NAME),
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

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return nil
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	isGroupChanged := false
	dn, gn, err := splitGroupId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	membersToDelete := make([]*zms.GroupMember, 0)
	membersToAdd := make([]*zms.GroupMember, 0)
	currentGroup, err := zmsClient.GetGroup(dn, gn)

	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("members") {
		oldVal, newVal := handleChange(d, "members")
		membersToDelete = expandDeprecatedGroupMembers(oldVal.Difference(newVal).List())
		membersToAdd = expandDeprecatedGroupMembers(newVal.Difference(oldVal).List())
	}
	if d.HasChange("member") {
		oldVal, newVal := handleChange(d, "member")
		membersToDelete = append(membersToDelete, expandGroupMembers(oldVal.Difference(newVal).List())...)
		membersToAdd = append(membersToAdd, expandGroupMembers(newVal.Difference(oldVal).List())...)
	}

	if d.HasChange("tags") {
		isGroupChanged = true
		_, n := d.GetChange("tags")
		tags := expandTagsMap(n.(map[string]interface{}))
		currentGroup.Tags = tags
	}

	if isGroupChanged {
		err := zmsClient.PutGroup(dn, gn, auditRef, currentGroup)
		if err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	err = updateGroupMembers(dn, gn, membersToDelete, membersToAdd, zmsClient, auditRef)
	if err != nil {
		return diag.Errorf("error updating group membership: %s", err)
	}

	return readAfterWrite(resourceGroupRead, ctx, d, meta)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
