package athenz

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAssertionSchema() *schema.Schema {
	return &schema.Schema{
		Type:       schema.TypeSet,
		ConfigMode: schema.SchemaConfigModeAttr,
		Optional:   true,
		Computed:   false,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"effect": {
					Type:     schema.TypeString,
					Required: true,
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
	}
}

func resourceAssertionSchema() *schema.Schema {
	return &schema.Schema{
		Type:       schema.TypeList,
		ConfigMode: schema.SchemaConfigModeAttr,
		Optional:   true,
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
					ValidateFunc: func(i interface{}, s string) (ws []string, errors []error) {
						err := validateRoleNameWithinAssertion(i.(string))
						if err != nil {
							errors = append(errors, err)
						}
						return
					},
				},
				"resource": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: func(i interface{}, s string) (ws []string, errors []error) {
						err := validateResourceNameWithinAssertion(i.(string))
						if err != nil {
							errors = append(errors, err)
						}
						return
					},
				},
			},
		},
	}
}

func expandPolicyAssertions(dn string, configured []interface{}) []*zms.Assertion {
	assertions := make([]*zms.Assertion, 0, len(configured))
	for _, aRaw := range configured {
		data := aRaw.(map[string]interface{})
		role := dn + ROLE_SEPARATOR + data["role"].(string)
		resource := data["resource"].(string)
		effect := zms.NewAssertionEffect(strings.ToUpper(data["effect"].(string)))

		a := &zms.Assertion{
			Role:     role,
			Resource: resource,
			Action:   data["action"].(string),
			Effect:   &effect,
		}

		assertions = append(assertions, a)
	}

	return assertions
}

func flattenPolicyAssertion(list []*zms.Assertion) []interface{} {
	policyAssertions := make([]interface{}, 0, len(list))
	for _, a := range list {
		role := strings.Split(a.Role, ROLE_SEPARATOR)[1]
		resource := a.Resource
		effect := a.Effect.String()
		action := a.Action

		a := map[string]interface{}{
			"role":     role,
			"resource": resource,
			"action":   action,
			"effect":   effect,
		}
		policyAssertions = append(policyAssertions, a)
	}
	return policyAssertions
}
