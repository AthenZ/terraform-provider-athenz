package athenz

import (
	"context"
	"log"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourcePolicyRead,
		CreateContext: resourcePolicyCreate,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:             schema.TypeString,
				Description:      "Name of the domain that policy belongs to",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the standard policy",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(ENTTITY_NAME),
			},
			"assertion": resourceAssertionSchema(),
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
		},
		// utilized CustomizeDiff method to achieve multi-attribute validation at terraform plan stage
		CustomizeDiff: validatePolicySchema(),
	}
}

// utilized CustomizeDiff method to achieve multi-attribute validation at terraform plan stage
func validatePolicySchema() schema.CustomizeDiffFunc {
	return customdiff.All(
		customdiff.ValidateChange("assertion", func(ctx context.Context, old, new, meta any) error {
			assertions := new.(*schema.Set).List()
			return validateAssertion(assertions)
		}),
	)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, pn, err := splitPolicyId(d.Id())
	d.Get("assertion")
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("domain", dn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", pn); err != nil {
		return diag.FromErr(err)
	}
	policy, err := zmsClient.GetPolicy(dn, pn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz Policy %s not found, removing from state", d.Id())
			return diag.Errorf(NOT_FOUNT_ERR)
		}
		return diag.Errorf("error retrieving Athenz Policy %s: %s", d.Id(), v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	if policy == nil {
		return diag.Errorf("error retrieving Athenz Policy - Make sure your cert/key are valid")
	}
	if len(policy.Assertions) > 0 {
		if err = d.Set("assertion", flattenPolicyAssertion(policy.Assertions)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("assertion", nil); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(err)
		}
	case rdl.Any:
		return diag.FromErr(err)
	case nil:
		if policyCheck != nil {
			return diag.Errorf("the policy %s is already exists in the domain %s use terraform import command", pn, dn)
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId(fullResourceName)

	return readAfterWrite(resourcePolicyRead, ctx, d, meta)
}
func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, pn, err := splitPolicyId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := zmsClient.GetPolicy(dn, pn)
	if err != nil {
		return diag.Errorf("error retrieving Athenz Policy: %s", err)
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
			return diag.FromErr(err)
		}
	}
	return readAfterWrite(resourcePolicyRead, ctx, d, meta)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, pn, err := splitPolicyId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	auditRef := d.Get("audit_ref").(string)
	if err := zmsClient.DeletePolicy(dn, pn, auditRef); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
