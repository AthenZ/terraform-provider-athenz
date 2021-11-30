package athenz

import (
	"fmt"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDomainRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	domainName := d.Get("name").(string)
	domain, err := zmsClient.GetDomain(domainName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return fmt.Errorf("athenz domain %s not found, update your data source query", domainName)
		} else {
			return fmt.Errorf("error retrieving Athenz domain: %s", v)
		}
	case rdl.Any:
		return err
	}
	if domain == nil {
		return fmt.Errorf("error retrieving Athenz domain: %s", domainName)
	}
	d.SetId(string(domain.Name))

	return nil
}
