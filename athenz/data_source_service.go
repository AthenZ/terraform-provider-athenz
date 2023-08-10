package athenz

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServiceRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the service",
				Optional:    true,
			},
			"public_keys": {
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Optional:   true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(client.ZmsClient)

	domainName := d.Get("domain").(string)
	serviceName := d.Get("name").(string)
	shortServiceName := getShortName(domainName, serviceName, SERVICE_SEPARATOR)
	fullResourceName := domainName + SERVICE_SEPARATOR + shortServiceName

	service, err := client.GetServiceIdentity(domainName, shortServiceName)
	if err := d.Set("description", service.Description); err != nil {
		return nil
	}
	if err := d.Set("public_keys", flattenPublicKeyEntryList(service.PublicKeys)); err != nil {
		return nil
	}
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return diag.Errorf("athenz Service %s not found, update your data source query", fullResourceName)
		} else {
			return diag.Errorf("error retrieving Athenz Service: %s", v)
		}
	case rdl.Any:
		return diag.FromErr(err)
	}
	d.SetId(fullResourceName)
	if len(service.Tags) > 0 {
		if err = d.Set("tags", flattenTag(service.Tags)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
