package athenz

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceUserDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserDomainCreate,
		ReadContext:   resourceUserDomainRead,
		DeleteContext: resourceUserDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the standard user domain",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // must be set as true, since no update method
				Default:  AUDIT_REF,
			},
		},
	}
}

func resourceUserDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName := d.Get("name").(string)
	auditRef := d.Get("audit_ref").(string)
	userDomainDetail := zms.UserDomain{
		Name: zms.SimpleName(domainName),
	}
	userDomain, err := zmsClient.PostUserDomain(domainName, auditRef, &userDomainDetail)
	if err != nil {
		return diag.FromErr(err)
	}
	if userDomain == nil {
		return diag.Errorf("error creating User Domain: %s", err)
	}
	d.SetId(PREFIX_USER_DOMAIN + domainName)
	return readAfterWrite(resourceUserDomainRead, ctx, d, meta)
}

func resourceUserDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName := d.Id()
	shortDomainName := shortName("", domainName, PREFIX_USER_DOMAIN)
	userDomain, err := zmsClient.GetDomain(domainName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz User Domain %s not found, removing from state", d.Id())
			return diag.FromErr(err)
		}
		return diag.Errorf("error retrieving Athenz User Domain: %s", v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	if userDomain == nil {
		return diag.Errorf("error retrieving Athenz User Domain - Make sure your cert/key are valid")
	}
	if err = d.Set("name", shortDomainName); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceUserDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName := shortName("", d.Id(), PREFIX_USER_DOMAIN)
	auditRef := d.Get("audit_ref").(string)
	err := zmsClient.DeleteUserDomain(domainName, auditRef)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
