package athenz

import (
	"fmt"
	"log"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceService() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreate,
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Name of the domain that service belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the service to be added to the domain",
				Required:    true,
				ForceNew:    true,
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
		},
	}
}

func resourceServiceCreate(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	domainName := d.Get("domain").(string)
	serviceName := d.Get("name").(string)
	auditRef := d.Get("audit_ref").(string)
	description := d.Get("description").(string)
	publicKeys := d.Get("public_keys").(*schema.Set).List()
	shortName := shortName(domainName, serviceName, SERVICE_SEPARATOR)
	longName := domainName + SERVICE_SEPARATOR + shortName
	publicKeyList := convertToPublicKeyEntryList(publicKeys)

	serviceCheck, err := zmsClient.GetServiceIdentity(domainName, serviceName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			detail := zms.ServiceIdentity{
				Name:        zms.ServiceName(longName),
				Description: description,
				PublicKeys:  publicKeyList,
			}

			err = zmsClient.PutServiceIdentity(domainName, shortName, auditRef, &detail)

			if err != nil {
				return err
			}
		}
	case rdl.Any:
		return err
	case nil:
		if serviceCheck != nil {
			return fmt.Errorf("the service %s is already exists in the domain %s use terraform import command", serviceName, domainName)
		} else {
			return err
		}
	}
	d.SetId(longName)

	return resourceServiceRead(d, meta)
}

func resourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	domainName, shortName := splitServiceId(d.Id())

	if err := d.Set("domain", domainName); err != nil {
		return err
	}
	if err := d.Set("name", shortName); err != nil {
		return err
	}
	service, err := zmsClient.GetServiceIdentity(domainName, shortName)

	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz Service %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving Athenz Service: %s", v)
	case rdl.Any:
		return err
	}

	if service == nil {
		return fmt.Errorf("error retrieving Athenz Service - Make sure your cert/key are valid")
	}
	if err = d.Set("description", service.Description); err != nil {
		return err
	}
	if len(service.PublicKeys) > 0 {
		if err = d.Set("public_keys", flattenPublicKeyEntryList(service.PublicKeys)); err != nil {
			return err
		}
	}

	return nil
}

func resourceServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	domainName := d.Get("domain").(string)
	serviceName := d.Get("name").(string)
	description := d.Get("description").(string)
	shortName := shortName(domainName, serviceName, SERVICE_SEPARATOR)
	longName := domainName + SERVICE_SEPARATOR + shortName
	auditRef := d.Get("audit_ref").(string)
	if d.HasChange("public_keys") {
		_, newVal := d.GetChange("public_keys")
		if newVal == nil {
			newVal = new(schema.Set)
		}
		newPublicKeyList := convertToPublicKeyEntryList(newVal.(*schema.Set).List())
		detail := zms.NewServiceIdentity()
		detail.Name = zms.ServiceName(longName)
		detail.PublicKeys = newPublicKeyList
		detail.Description = description
		err := zmsClient.PutServiceIdentity(domainName, shortName, auditRef, detail)
		if err != nil {
			return fmt.Errorf("error updating service membership: %s", err)
		}
	}
	return resourceServiceRead(d, meta)
}

func resourceServiceDelete(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	domainName, serviceName := splitServiceId(d.Id())
	auditRef := d.Get("audit_ref").(string)
	err := zmsClient.DeleteServiceIdentity(domainName, serviceName, auditRef)
	if err != nil {
		return err
	}

	return nil
}
