package athenz

import (
	"fmt"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func DataSourcePolicyVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePolicyVersionRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Name of the domain that policy belongs to",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the policy",
				Required:    true,
			},
			"active_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy version that will be active",
			},
			"versions": {
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assertion": policyVersionAssertionSchema(),
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

func dataSourcePolicyVersionRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	pn := d.Get("name").(string)
	fullResourceName := dn + POLICY_SEPARATOR + pn
	policyVersionList, err := getAllPolicyVersions(zmsClient, dn, pn)
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
	if policyVersionList == nil {
		return fmt.Errorf("error retrieving Athenz Policy - Make sure your cert/key are valid")
	}

	activeVersion := getActiveVersionName(policyVersionList)
	if err = d.Set("active_version", activeVersion); err != nil {
		return err
	}
	if err = d.Set("versions", flattenPolicyVersions(policyVersionList)); err != nil {
		return err
	}
	return nil
}
