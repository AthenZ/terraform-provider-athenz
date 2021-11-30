package athenz

import (
	"fmt"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceServiceRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(client.ZmsClient)

	domainName := d.Get("domain").(string)
	serviceName := d.Get("name").(string)
	shortServiceName := shortName(domainName, serviceName, SERVICE_SEPARATOR)
	fullResourceName := domainName + SERVICE_SEPARATOR + shortServiceName

	_, err := client.GetServiceIdentity(domainName, shortServiceName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return fmt.Errorf("athenz Service %s not found, update your data source query", fullResourceName)
		} else {
			return fmt.Errorf("error retrieving Athenz Service: %s", v)
		}
	case rdl.Any:
		return err
	}
	d.SetId(fullResourceName)

	return nil
}
