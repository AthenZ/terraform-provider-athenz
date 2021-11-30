package athenz

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Read:   resourcePolicyRead,
		Create: resourcePolicyCreate,
		Update: resourcePolicyUpdate,
		Delete: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Name of the domain that policy belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the standard policy",
				Required:    true,
				ForceNew:    true,
			},
			"assertion": {
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Optional:   true,
				Computed:   false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"effect": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"ALLOW",
								"DENY",
							}, false),
							StateFunc: func(v interface{}) string {
								return strings.ToUpper(v.(string))
							},
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
		},
	}
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	fullResourceName := strings.Split(d.Id(), POLICY_SEPARATOR)
	dn := fullResourceName[0]
	pn := fullResourceName[1]

	if err := d.Set("domain", dn); err != nil {
		return err
	}
	if err := d.Set("name", pn); err != nil {
		return err
	}
	policy, err := zmsClient.GetPolicy(dn, pn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz Policy %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving Athenz Policy %s: %s", d.Id(), v)
	case rdl.Any:
		return err
	}

	if policy == nil {
		return fmt.Errorf("error retrieving Athenz Policy - Make sure your cert/key are valid")
	}
	if len(policy.Assertions) > 0 {
		if err = d.Set("assertion", flattenPolicyAssertion(policy.Assertions)); err != nil {
			return err
		}
	}
	return nil
}

func resourcePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	pn := d.Get("name").(string)
	fullResourceName := dn + POLICY_SEPARATOR + pn
	policyCheck, err := zmsClient.GetPolicy(dn, pn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			policy := zms.Policy{
				Name:     zms.ResourceName(fullResourceName),
				Modified: nil,
			}
			if v, ok := d.GetOk("assertion"); ok && v.(*schema.Set).Len() > 0 {
				policy.Assertions = expandPolicyAssertions(dn, v.(*schema.Set).List())
			} else {
				policy.Assertions = make([]*zms.Assertion, 0)
			}

			auditRef := d.Get("audit_ref").(string)
			err = zmsClient.PutPolicy(dn, pn, auditRef, &policy)
			if err != nil {
				return err
			}
		}
	case rdl.Any:
		return err
	case nil:
		if policyCheck != nil {
			return fmt.Errorf("the policy %s is already exists in the domain %s use terraform import command", pn, dn)
		} else {
			return err
		}
	}
	d.SetId(fullResourceName)

	return resourcePolicyRead(d, meta)
}
func resourcePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	fullResourceName := strings.Split(d.Id(), POLICY_SEPARATOR)
	dn := fullResourceName[0]
	pn := fullResourceName[1]

	policy, err := zmsClient.GetPolicy(dn, pn)
	if err != nil {
		return fmt.Errorf("error retrieving Athenz Policy: %s", err)
	}
	if d.HasChange("assertion") {
		_, newVal := d.GetChange("assertion")
		if newVal == nil {
			newVal = new(schema.Set)
		}
		ns := newVal.(*schema.Set).List()
		policy.Assertions = expandPolicyAssertions(dn, ns)
		auditRef := d.Get("audit_ref").(string)
		err = zmsClient.PutPolicy(dn, pn, auditRef, policy)
		if err != nil {
			return err
		}
	}
	return resourcePolicyRead(d, meta)
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)
	fullResourceName := strings.Split(d.Id(), POLICY_SEPARATOR)
	dn := fullResourceName[0]
	pn := fullResourceName[1]

	auditRef := d.Get("audit_ref").(string)
	err := zmsClient.DeletePolicy(dn, pn, auditRef)
	if err != nil {
		return err
	}
	return nil
}
