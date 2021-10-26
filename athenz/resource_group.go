package athenz

import (
	"fmt"
	"log"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/AthenZ/terraform-provider-athenz/client"

	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Name of the domain that group belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the standard group role",
				Required:    true,
				ForceNew:    true,
			},
			"members": {
				Type:        schema.TypeSet,
				Description: "Users or services to be added as members",
				Optional:    true,
				Computed:    false,
				Elem: &schema.Schema{Type: schema.TypeString,
					ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
						value := v.(string)
						if strings.Contains(value, ":group.") {
							errors = append(errors, fmt.Errorf("%q. A group can't be a member of another group", v))
						}
						return
					},
					Set: schema.HashString,
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

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
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

			if v, ok := d.GetOk("members"); ok && v.(*schema.Set).Len() > 0 {
				group.GroupMembers = expandGroupMembers(v.(*schema.Set).List())
			}

			auditRef := d.Get("audit_ref").(string)
			if err = zmsClient.PutGroup(dn, gn, auditRef, &group); err != nil {
				return err
			}
		}
	case rdl.Any:
		return err
	case nil:
		if groupCheck != nil {
			return fmt.Errorf("the group %s is already exists in the domain %s use terraform import command", gn, dn)
		} else {
			return err
		}
	}
	d.SetId(fullResourceName)

	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	fullResourceName := strings.Split(d.Id(), GROUP_SEPARATOR)
	dn, gn := fullResourceName[0], fullResourceName[1]
	if err := d.Set("domain", dn); err != nil {
		return err
	}
	if err := d.Set("name", gn); err != nil {
		return err
	}

	group, err := zmsClient.GetGroup(dn, gn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz Group %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving Athenz Group %s: %s", d.Id(), v)
	case rdl.Any:
		return err
	}

	if group == nil {
		return fmt.Errorf("error retrieving Athenz Group - Make sure your cert/key are valid")
	}

	if len(group.GroupMembers) > 0 {
		d.Set("members", flattenGroupMember(group.GroupMembers))
	}

	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	fullResourceName := strings.Split(d.Id(), GROUP_SEPARATOR)
	dn, gn := fullResourceName[0], fullResourceName[1]

	auditRef := d.Get("audit_ref").(string)
	if d.HasChange("members") {
		oldVal, newVal := d.GetChange("members")
		err := updateGroupMembers(dn, gn, oldVal, newVal, zmsClient, auditRef)
		if err != nil {
			return fmt.Errorf("error updating group membership: %s", err)
		}
	}
	return resourceGroupRead(d, meta)
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	fullResourceName := strings.Split(d.Id(), GROUP_SEPARATOR)
	dn, gn := fullResourceName[0], fullResourceName[1]
	auditRef := d.Get("audit_ref").(string)
	err := zmsClient.DeleteGroup(dn, gn, auditRef)
	if err != nil {
		return err
	}
	return nil
}
