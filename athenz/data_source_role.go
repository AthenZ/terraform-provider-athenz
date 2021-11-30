package athenz

import (
	"fmt"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"members": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceRoleRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	dn := d.Get("domain").(string)
	rn := d.Get("name").(string)
	fullResourceName := dn + ROLE_SEPARATOR + rn

	role, err := zmsClient.GetRole(dn, rn)

	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return fmt.Errorf("athenz Role %s not found, update your data source query", fullResourceName)
		} else {
			return fmt.Errorf("error retrieving Athenz Role: %s", v)
		}
	case rdl.Any:
		return err
	}
	d.SetId(fullResourceName)

	if len(role.RoleMembers) > 0 {
		d.Set("members", flattenRoleMembers(role.RoleMembers))

	}
	if len(role.Tags) > 0 {
		d.Set("tags", flattenTag(role.Tags))
	}

	return nil
}
