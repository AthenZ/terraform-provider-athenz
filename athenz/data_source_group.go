package athenz

import (
	"fmt"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGroupRead,
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
		},
	}
}

func dataSourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	domainName := d.Get("domain").(string)
	groupName := d.Get("name").(string)
	fullResourceName := domainName + GROUP_SEPARATOR + groupName

	group, err := zmsClient.GetGroup(domainName, groupName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return fmt.Errorf("athenz group %s not found, update your data source query", fullResourceName)
		} else {
			return fmt.Errorf("error retrieving Athenz Group: %s", v)
		}
	case rdl.Any:
		return err
	}
	d.SetId(fullResourceName)

	if len(group.GroupMembers) > 0 {
		d.Set("members", flattenGroupMember(group.GroupMembers))
	}

	return nil
}
