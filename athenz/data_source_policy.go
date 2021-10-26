package athenz

import (
	"fmt"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func DataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePolicyRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"assertion": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"effect": {
							Type:     schema.TypeString,
							Required: true,
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	pn := d.Get("name").(string)
	fullResourceName := dn + POLICY_SEPARATOR + pn
	policy, err := zmsClient.GetPolicy(dn, pn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return fmt.Errorf("athenz Policy %s not found, update your data source query", fullResourceName)
		} else {
			return fmt.Errorf("error retrieving Athenz Policy: %s", v)
		}
	case rdl.Any:
		return err
	}
	d.SetId(fullResourceName)
	if len(policy.Assertions) > 0 {
		d.Set("assertion", flattenPolicyAssertion(policy.Assertions))
	}
	return nil
}
