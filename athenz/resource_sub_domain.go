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

func ResourceSubDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubDomainCreate,
		ReadContext:   resourceSubDomainRead,
		DeleteContext: resourceSubDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"parent_name": {
				Type:             schema.TypeString,
				Description:      "Name of the standard parent domain",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the standard sub domain",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(SIMPLE_NAME),
			},
			"admin_users": {
				Type:        schema.TypeSet,
				Description: "Names of the standard admin users",
				Required:    true,
				ForceNew:    true, // must be set as true, since no update method
				Elem:        &schema.Schema{Type: schema.TypeString},
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

func getSubDomainSchemaAttributes(d *schema.ResourceData) (adminUsers []interface{}, auditRef string) {
	adminUsers = d.Get("admin_users").(*schema.Set).List()
	auditRef = d.Get("audit_ref").(string)
	return
}

func resourceSubDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
				return diag.FromErr(err)
			}
			if subDomain == nil {
				return diag.Errorf("error creating Sub Domain: %s", err)
			}
		} else {
			return diag.FromErr(err)
		}
	case rdl.Any:
		return diag.FromErr(err)
	case nil:
		if subDomainCheck != nil {
			return diag.Errorf("the sub-domain %s is already exists, use terraform import command", domainName)
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId(parentDomainName + SUB_DOMAIN_SEPARATOR + domainName)
	return resourceSubDomainRead(ctx, d, meta)
}

func resourceSubDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	fullyQualifiedName := d.Id()
	parentDomainName, domainName, err := splitSubDomainId(fullyQualifiedName)
	if err != nil {
		return diag.FromErr(err)
	}
	subDomain, err := zmsClient.GetDomain(fullyQualifiedName)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz Sub Domain %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving Athenz Sub Domain: %s", v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	if subDomain == nil {
		return diag.Errorf("error retrieving Athenz Sub Domain - Make sure your cert/key are valid")
	}

	adminRole, err := zmsClient.GetRole(fullyQualifiedName, "admin")
	if err != nil {
		return diag.FromErr(err)
	}
	adminUsers := flattenDeprecatedRoleMembers(adminRole.RoleMembers)
	if err = d.Set("admin_users", adminUsers); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("parent_name", parentDomainName); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("name", domainName); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSubDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	parentDomainName, subDomainName, err := splitSubDomainId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	if err = zmsClient.DeleteSubDomain(parentDomainName, subDomainName, auditRef); err != nil {
		return diag.FromErr(err)
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
