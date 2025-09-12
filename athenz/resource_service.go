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

func ResourceService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceCreate,
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
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
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the service",
				Optional:    true,
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
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
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	domainName := d.Get("domain").(string)
	serviceName := d.Get("name").(string)
	auditRef := d.Get("audit_ref").(string)
	description := d.Get("description").(string)
	publicKeys := d.Get("public_keys").(*schema.Set).List()
	shortName := getShortName(domainName, serviceName, SERVICE_SEPARATOR)
	longName := domainName + SERVICE_SEPARATOR + shortName
	publicKeyList := convertToPublicKeyEntryList(publicKeys)

	serviceCheck, err := zmsClient.GetServiceIdentity(domainName, serviceName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			service := zms.ServiceIdentity{
				Name:        zms.ServiceName(longName),
				Description: description,
				PublicKeys:  publicKeyList,
			}
			if v, ok := d.GetOk("tags"); ok {
				service.Tags = expandTagsMap(v.(map[string]interface{}))
			}
			err = zmsClient.PutServiceIdentity(domainName, shortName, auditRef, &service)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(err)
		}
	case rdl.Any:
		return diag.FromErr(err)
	case nil:
		if serviceCheck != nil {
			return diag.Errorf("the service %s already exists in the domain %s, use terraform import command", serviceName, domainName)
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId(longName)
	return readAfterWrite(resourceServiceRead, ctx, d, meta)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			if !d.IsNewResource() {
				log.Printf("[WARN] Athenz Service %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.Errorf("error retrieving Athenz Service: %s", v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	if service == nil {
		return diag.Errorf("error retrieving Athenz Service - Make sure your cert/key are valid")
	}
	if err = d.Set("description", service.Description); err != nil {
		return diag.FromErr(err)
	}
	if len(service.PublicKeys) > 0 {
		if err = d.Set("public_keys", flattenPublicKeyEntryList(service.PublicKeys)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("public_keys", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	if len(service.Tags) > 0 {
		if err = d.Set("tags", flattenTag(service.Tags)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("tags", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	domainName, serviceName, err := splitServiceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	description := d.Get("description").(string)
	shortName := getShortName(domainName, serviceName, SERVICE_SEPARATOR)
	auditRef := d.Get("audit_ref").(string)
	service, err := zmsClient.GetServiceIdentity(domainName, serviceName)
	if err != nil {
		return diag.Errorf("error retrieving service %s: %s", d.Id(), err)
	}
	service.Description = description

	if d.HasChange("public_keys") {
		_, newVal := d.GetChange("public_keys")
		if newVal == nil {
			newVal = new(schema.Set)
		}
		newPublicKeyList := convertToPublicKeyEntryList(newVal.(*schema.Set).List())
		service.PublicKeys = newPublicKeyList
	} else {
		publicKeyList := d.Get("public_keys").(*schema.Set).List()
		service.PublicKeys = convertToPublicKeyEntryList(publicKeyList)
	}

	if d.HasChange("tags") {
		_, n := d.GetChange("tags")
		service.Tags = expandTagsMap(n.(map[string]interface{}))
	}

	err = zmsClient.PutServiceIdentity(domainName, shortName, auditRef, service)
	if err != nil {
		return diag.Errorf("error updating service membership: %s", err)
	}

	return readAfterWrite(resourceServiceRead, ctx, d, meta)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domainName, serviceName, err := splitServiceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	err = zmsClient.DeleteServiceIdentity(domainName, serviceName, auditRef)

	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return nil
		}
		return diag.FromErr(err)
	case rdl.Any:
		return diag.FromErr(err)
	}
	return nil
}
