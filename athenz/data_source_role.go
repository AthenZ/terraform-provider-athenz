package athenz

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleRead,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"trust": {
				Type:        schema.TypeString,
				Description: "The domain, which this role is trusted to",
				Optional:    true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	dn := d.Get("domain").(string)
	rn := d.Get("name").(string)
	fullResourceName := dn + ROLE_SEPARATOR + rn

	role, err := zmsClient.GetRole(dn, rn)

	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return diag.Errorf("athenz Role %s not found, update your data source query", fullResourceName)
		} else {
			return diag.Errorf("error retrieving Athenz Role: %s", v)
		}
	case rdl.Any:
		return diag.FromErr(err)
	}
	d.SetId(fullResourceName)

	if len(role.RoleMembers) > 0 {
		if err = d.Set("members", flattenRoleMembers(role.RoleMembers)); err != nil {
			return diag.FromErr(err)
		}
	}
	if len(role.Tags) > 0 {
		if err = d.Set("tags", flattenTag(role.Tags)); err != nil {
			return diag.FromErr(err)
		}
	}
	if role.Trust != "" {
		if err = d.Set("trust", string(role.Trust)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
