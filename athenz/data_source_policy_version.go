package athenz

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourcePolicyVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePolicyVersionRead,
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
				Required:    true,
				Description: "The policy version that will be active",
			},
			"version": {
				Required: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"assertion": dataSourceAssertionSchema(),
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

func dataSourcePolicyVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	pn := d.Get("name").(string)
	fullResourceName := dn + POLICY_SEPARATOR + pn
	policyVersionList, err := getAllPolicyVersions(zmsClient, dn, pn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return diag.Errorf("athenz Policy %s not found, update your data source query", fullResourceName)
		} else {
			return diag.Errorf("error retrieving Athenz Policy: %s", v)
		}
	case rdl.Any:
		return diag.FromErr(err)
	}

	d.SetId(fullResourceName)
	if policyVersionList == nil {
		return diag.Errorf("error retrieving Athenz Policy - Make sure your cert/key are valid")
	}

	activeVersion := getActiveVersionName(policyVersionList)
	if err = d.Set("active_version", activeVersion); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("version", flattenPolicyVersions(policyVersionList)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
