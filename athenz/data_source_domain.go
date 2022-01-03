package athenz

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName := d.Get("name").(string)
	domain, err := zmsClient.GetDomain(domainName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return diag.Errorf("athenz domain %s not found, update your data source query", domainName)
		} else {
			return diag.Errorf("error retrieving Athenz domain: %s", v)
		}
	case rdl.Any:
		return diag.FromErr(err)
	}
	if domain == nil {
		return diag.Errorf("error retrieving Athenz domain: %s", domainName)
	}
	d.SetId(string(domain.Name))

	return nil
}
