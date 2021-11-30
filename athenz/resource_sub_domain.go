package athenz

import (
	"fmt"
	"log"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSubDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubDomainCreate,
		Read:   resourceSubDomainRead,
		Delete: resourceSubDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"parent_name": {
				Type:        schema.TypeString,
				Description: "Name of the standard parent domain",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the standard sub domain",
				Required:    true,
				ForceNew:    true,
			},
			"admin_users": {
				Type:        schema.TypeSet,
				Description: "Names of the standard admin users",
				Required:    true,
				ForceNew:    true, // must to be true, because no update method
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // must to be true, because no update method
				Default:  AUDIT_REF,
			},
		},
	}
}

func getSubDomainSchemaAttributes(d *schema.ResourceData) (adminUsers []interface{}, auditRef string) {
	adminUsers = d.Get("admin_users").(*schema.Set).List()
	auditRef = d.Get("audit_ref").(string)
	return
}

func resourceSubDomainCreate(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	parentDomainName := d.Get("parent_name").(string)
	domainName := shortName(parentDomainName, d.Get("name").(string), SUB_DOMAIN_SEPARATOR)
	adminUsers, auditRef := getSubDomainSchemaAttributes(d)
	subDomainDetail := zms.SubDomain{
		Name:       zms.SimpleName(domainName),
		Parent:     zms.DomainName(parentDomainName),
		AdminUsers: convertToZmsResourceNameList(adminUsers),
	}
	subDomainCheck, err := zmsClient.GetDomain(domainName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			subDomain, err := zmsClient.PostSubDomain(parentDomainName, auditRef, &subDomainDetail)
			if err != nil {
				return err
			}
			if subDomain == nil {
				return fmt.Errorf("error creating Sub Domain: %s", err)
			}

		}
	case rdl.Any:
		return err
	case nil:
		if subDomainCheck != nil {
			return fmt.Errorf("the sub-domain %s is already exists, use terraform import command", domainName)
		} else {
			return err
		}
	}
	d.SetId(parentDomainName + SUB_DOMAIN_SEPARATOR + domainName)
	return resourceSubDomainRead(d, meta)
}

func resourceSubDomainRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	fullyQualifiedName := d.Id()
	parentDomainName, domainName := splitServiceId(fullyQualifiedName)

	subDomain, err := zmsClient.GetDomain(fullyQualifiedName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz Sub Domain %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving Athenz Sub Domain: %s", v)
	case rdl.Any:
		return err
	}

	if subDomain == nil {
		return fmt.Errorf("error retrieving Athenz Sub Domain - Make sure your cert/key are valid")
	}

	adminRole, err := zmsClient.GetRole(fullyQualifiedName, "admin")
	if err != nil {
		return err
	}
	adminUsers := flattenRoleMembers(adminRole.RoleMembers)
	if err = d.Set("admin_users", adminUsers); err != nil {
		return err
	}
	if err = d.Set("parent_name", parentDomainName); err != nil {
		return err
	}
	if err = d.Set("name", domainName); err != nil {
		return err
	}
	return nil
}

func resourceSubDomainDelete(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	parentDomainName, subDomainName := splitSubDomainId(d.Id())
	auditRef := d.Get("audit_ref").(string)
	err := zmsClient.DeleteSubDomain(parentDomainName, subDomainName, auditRef)
	if err != nil {
		return err
	}
	return nil
}

func convertToZmsResourceNameList(resourceNames []interface{}) []zms.ResourceName {
	zmsResourceNames := make([]zms.ResourceName, 0, len(resourceNames))
	for _, val := range resourceNames {
		resourceName, ok := val.(string)
		if ok && resourceName != "" {
			zmsResourceNames = append(zmsResourceNames, zms.ResourceName(resourceName))
		}
	}
	return zmsResourceNames
}
