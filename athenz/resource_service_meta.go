package athenz

import (
	"context"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceServiceMeta() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceMetaCreate,
		ReadContext:   resourceServiceMetaRead,
		UpdateContext: resourceServiceMetaUpdate,
		DeleteContext: resourceServiceMetaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:             schema.TypeString,
				Description:      "Name of the domain that service belongs to",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the service to be added to the domain",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(SIMPLE_NAME),
			},
			"provider_endpoint": {
				Type:        schema.TypeString,
				Description: "provider endpoint",
				Required:    true,
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
		},
	}
}

func resourceServiceMetaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName := d.Get("domain").(string)
	serviceName := d.Get("name").(string)
	shortName := getShortName(domainName, serviceName, SERVICE_SEPARATOR)
	longName := domainName + SERVICE_SEPARATOR + shortName

	_, err := zmsClient.GetServiceIdentity(domainName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	resp := updateServiceMeta(zmsClient, domainName, shortName, d)
	if resp != nil {
		return resp
	}
	d.SetId(longName)

	return readAfterWrite(resourceServiceMetaRead, ctx, d, meta)
}

func resourceServiceMetaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	domainName, serviceName, err := splitServiceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("domain", domainName); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("name", serviceName); err != nil {
		return diag.FromErr(err)
	}

	service, err := zmsClient.GetServiceIdentity(domainName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("provider_endpoint", service.ProviderEndpoint); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceServiceMetaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName, serviceName, err := splitServiceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	shortName := getShortName(domainName, serviceName, SERVICE_SEPARATOR)
	resp := updateServiceMeta(zmsClient, domainName, shortName, d)
	if resp != nil {
		return resp
	}

	return readAfterWrite(resourceServiceMetaRead, ctx, d, meta)
}

func updateServiceMeta(zmsClient client.ZmsClient, dn, sn string, d *schema.ResourceData) diag.Diagnostics {
	auditRef := d.Get("audit_ref").(string)
	pe := d.Get("provider_endpoint").(string)
	serviceMeta := zms.ServiceIdentitySystemMeta{
		ProviderEndpoint: pe,
	}
	err := zmsClient.PutServiceIdentitySystemMeta(dn, sn, "providerendpoint", auditRef, &serviceMeta)
	if err != nil {
		return diag.Errorf("error updating service provider endpoint: %s", err)
	}
	return nil
}

func resourceServiceMetaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName, serviceName, err := splitServiceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	shortName := getShortName(domainName, serviceName, SERVICE_SEPARATOR)
	//longName := domainName + SERVICE_SEPARATOR + shortName
	auditRef := d.Get("audit_ref").(string)
	serviceMeta := zms.ServiceIdentitySystemMeta{
		ProviderEndpoint: "",
	}
	err = zmsClient.PutServiceIdentitySystemMeta(domainName, shortName, "providerendpoint", auditRef, &serviceMeta)
	if err != nil {
		return diag.Errorf("error updating service provider endpoint: %s", err)
	}

	return nil
}
