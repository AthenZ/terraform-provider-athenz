package athenz

import (
	"fmt"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func policyVersionAssertionSchema() *schema.Schema {
	return &schema.Schema{
		Type:       schema.TypeSet,
		ConfigMode: schema.SchemaConfigModeAttr,
		Optional:   true,
		Elem: &schema.Schema{
			Type: schema.TypeMap,
			ValidateFunc: func(i interface{}, s string) (ws []string, errors []error) {
				assertionMap := i.(map[string]interface{})
				if len(assertionMap) != 4 {
					errors = append(errors, fmt.Errorf("assertion: %v is invalid. each assertion must be exactly 4 items", assertionMap))
				}
				validKeys := []string{"effect", "action", "role", "resource"}
				var valid bool
				for key := range assertionMap {
					valid = false
					for _, validKay := range validKeys {
						if key == validKay {
							valid = true
							break
						}
					}
					if !valid {
						errors = append(errors, fmt.Errorf("assertion: %v is invalid. the asserion key must matchs one of the follwoing: %v", assertionMap, validKeys))
					}
				}
				return
			},
			Elem: &schema.Schema{Type: schema.TypeString, Required: true},
		},
	}
}

func expandPolicyAssertions(dn string, configured []interface{}) []*zms.Assertion {
	assertions := make([]*zms.Assertion, 0, len(configured))
	for _, aRaw := range configured {
		data := aRaw.(map[string]interface{})
		role := data["role"].(string)
		if !strings.Contains(role, ROLE_SEPARATOR) {
			role = dn + ROLE_SEPARATOR + role
		}
		resource := data["resource"].(string)
		if !strings.Contains(resource, RESOURCE_SEPARATOR) {
			resource = dn + RESOURCE_SEPARATOR + resource
		}

		var effect = zms.NewAssertionEffect(strings.ToUpper(data["effect"].(string)))

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
		resource := strings.Split(a.Resource, RESOURCE_SEPARATOR)[1]
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
