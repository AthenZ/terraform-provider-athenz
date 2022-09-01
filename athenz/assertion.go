package athenz

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
				"case_sensitive": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
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
				"case_sensitive": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
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
		action := data["action"].(string)
		caseSensitive := data["case_sensitive"].(bool)

		a := &zms.Assertion{
			Role:          role,
			Resource:      resource,
			Action:        action,
			Effect:        &effect,
			CaseSensitive: &caseSensitive,
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
		caseSensitive := inferCaseSensitiveValue(action, resource)

		a := map[string]interface{}{
			"role":           role,
			"resource":       resource,
			"action":         action,
			"effect":         effect,
			"case_sensitive": caseSensitive,
		}
		policyAssertions = append(policyAssertions, a)

	}

	return policyAssertions
}

// enabling case_sensitive flag is allowed only if action or resource has capital letters
func validateCaseSensitiveValue(caseSensitive bool, action string, resourceName string) error {
	if caseSensitive {
		if strings.ToLower(resourceName) == resourceName && action == strings.ToLower(action) {
			return fmt.Errorf("enabling case_sensitive flag is allowed only if action or resource has capital letters")
		}
	}
	return nil
}

// case-sensitive value is inferred by analyzing assertion action and assertion resource
func inferCaseSensitiveValue(action, resourceName string) bool {
	return strings.ToLower(resourceName) != resourceName || strings.ToLower(action) != action
}

// utilized CustomizeDiff method to achieve multi-attribute validation at terraform plan stage
func validateAssertion() schema.CustomizeDiffFunc {
	return customdiff.All(
		customdiff.ValidateChange("assertion", func(ctx context.Context, old, new, meta any) error {
			assertions := new.([]interface{})
			for _, aRaw := range assertions {
				data := aRaw.(map[string]interface{})
				resource := data["resource"].(string)
				action := data["action"].(string)
				caseSensitive := data["case_sensitive"].(bool)
				if err := validateCaseSensitiveValue(caseSensitive, action, resource); err != nil {
					return err
				}
			}
			return nil
		}),
	)
}
