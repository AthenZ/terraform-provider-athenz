package athenz

import (
	"fmt"
	"github.com/AthenZ/athenz/clients/go/msd"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConditionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operator": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"value": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func dataSourceAssertionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
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
				"condition": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"instances": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     dataSourceConditionSchema(),
							},
							"id": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     dataSourceConditionSchema(),
							},
							"enforcementstate": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     dataSourceConditionSchema(),
							},
							"scopeonprem": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     dataSourceConditionSchema(),
							},
							"scopeaws": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     dataSourceConditionSchema(),
							},
							"scopeall": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     dataSourceConditionSchema(),
							},
						},
					},
				},
			},
		},
	}
}

func resourceConditionSchema(validateDiagFuncForScope schema.SchemaValidateDiagFunc) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operator": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1, // "EQUALS"
			},
			"value": {
				Required:         true,
				Type:             schema.TypeString,
				ValidateDiagFunc: validateDiagFuncForScope,
			},
		},
	}
}

func validateDiagFuncForScope(v any, p cty.Path) diag.Diagnostics {
	value := v.(string)
	return validation.ToDiagFunc(validation.StringInSlice([]string{"false", "true"}, false))(value, p)
}

func resourceAssertionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
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
				"id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"condition": {
					Type:     schema.TypeSet,
					MaxItems: 2, /* each assertion represent acl policy. Since for a given service,
					  			  you can make the acl policy enforced on some hosts and not on others,
								  therefore, to apply more than 2 conditions per assertion doesn't make any sense */
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Computed: true,
								Type:     schema.TypeInt,
							},
							"instances": {
								Type:     schema.TypeSet,
								Required: true,
								MaxItems: 1,
								Elem:     resourceConditionSchema(validatePatternFunc(ASSERTION_CONDITION_VALUE)),
							},
							"enforcementstate": {
								Type:     schema.TypeSet,
								Required: true,
								MaxItems: 1,
								Elem: resourceConditionSchema(
									validation.ToDiagFunc(
										validation.StringInSlice([]string{strings.ToLower(msd.REPORT.String()), strings.ToLower(msd.ENFORCE.String())}, false)),
								),
							},
							"scopeonprem": {
								Type:     schema.TypeSet,
								Required: true,
								MaxItems: 1,
								Elem:     resourceConditionSchema(validateDiagFuncForScope),
							},
							"scopeaws": {
								Type:     schema.TypeSet,
								Required: true,
								MaxItems: 1,
								Elem:     resourceConditionSchema(validateDiagFuncForScope),
							},
							"scopeall": {
								Type:     schema.TypeSet,
								Required: true,
								MaxItems: 1,
								Elem:     resourceConditionSchema(validateDiagFuncForScope),
							},
						},
					},
				},
			},
		},
	}
}

func expandAssertionConditions(configured []interface{}) *zms.AssertionConditions {
	conditionsList := make([]*zms.AssertionCondition, 0, len(configured))
	keys := []string{Instances, EnforcementState, ScopeONPREM, ScopeAWS, ScopeALL}
	for _, cRaw := range configured {
		data := cRaw.(map[string]interface{})
		conditionsMap := make(map[zms.AssertionConditionKey]*zms.AssertionConditionData, len(keys))
		for _, key := range keys {
			conditionsData := (data[key].(*schema.Set).List())[0].(map[string]interface{})
			operator := conditionsData["operator"].(int)
			value := conditionsData["value"].(string)
			conditionsMap[zms.AssertionConditionKey(key)] = &zms.AssertionConditionData{
				Operator: zms.AssertionConditionOperator(operator),
				Value:    zms.AssertionConditionValue(value),
			}
		}
		assertionCondition := &zms.AssertionCondition{
			ConditionsMap: conditionsMap,
		}
		conditionsList = append(conditionsList, assertionCondition)
	}
	return &zms.AssertionConditions{
		ConditionsList: conditionsList,
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
		if conditions, ok := data["condition"]; ok {
			a.Conditions = expandAssertionConditions(conditions.(*schema.Set).List())
		}
		assertions = append(assertions, a)
	}

	return assertions
}

func flattenAssertionConditions(list []*zms.AssertionCondition) []interface{} {
	assertionConditions := make([]interface{}, 0, len(list))
	for _, condition := range list {
		keys := []string{Instances, EnforcementState, ScopeONPREM, ScopeAWS, ScopeALL}

		c := make(map[string]interface{}, len(keys))
		c["id"] = (int)(*condition.Id)
		for _, key := range keys {
			c[key] = []map[string]interface{}{
				{
					"operator": (int)(condition.ConditionsMap[zms.AssertionConditionKey(key)].Operator),
					"value":    (string)(condition.ConditionsMap[zms.AssertionConditionKey(key)].Value),
				},
			}
		}
		assertionConditions = append(assertionConditions, c)
	}
	return assertionConditions
}

func flattenPolicyAssertion(list []*zms.Assertion) []interface{} {
	policyAssertions := make([]interface{}, 0, len(list))

	for _, assertion := range list {
		role := strings.Split(assertion.Role, ROLE_SEPARATOR)[1]
		resource := assertion.Resource
		effect := assertion.Effect.String()
		action := assertion.Action
		caseSensitive := inferCaseSensitiveValue(action, resource)

		a := map[string]interface{}{
			"role":           role,
			"resource":       resource,
			"action":         action,
			"effect":         effect,
			"case_sensitive": caseSensitive,
			"id":             (int)(*assertion.Id),
		}
		if assertion.Conditions != nil {
			a["condition"] = flattenAssertionConditions(assertion.Conditions.ConditionsList)
		}
		policyAssertions = append(policyAssertions, a)
	}

	return policyAssertions
}

// enabling case_sensitive flag is allowed if and only if action or resource has capital letters
func validateCaseSensitiveValue(caseSensitive bool, action string, resourceName string) error {
	if caseSensitive {
		if strings.ToLower(resourceName) == resourceName && action == strings.ToLower(action) {
			return fmt.Errorf("enabling case_sensitive flag is allowed only if action or resource has capital letters")
		}
	} else {
		if strings.ToLower(resourceName) != resourceName || action != strings.ToLower(action) {
			return fmt.Errorf("capitalized action or resource allowed only when enabling case_sensitive flag")
		}
	}
	return nil
}

// case-sensitive value is inferred by analyzing assertion action and assertion resource
func inferCaseSensitiveValue(action, resourceName string) bool {
	return strings.ToLower(resourceName) != resourceName || strings.ToLower(action) != action
}

func validateAssertion(assertions []interface{}) error {
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
}

func validateAssertionConditions(assertionConditions interface{}) error {
	conditions := assertionConditions.(*schema.Set).List()
	if len(conditions) <= 1 {
		return nil
	}
	c1 := conditions[0].(map[string]interface{})
	c2 := conditions[1].(map[string]interface{})
	enforcementState1 := getTheValueFromCondition(c1, EnforcementState)
	enforcementState2 := getTheValueFromCondition(c2, EnforcementState)
	if enforcementState1 == enforcementState2 {
		return fmt.Errorf("enforcement state can't be same for different conditions in a msd policy")
	}

	instances1 := getTheValueFromCondition(c1, Instances)
	instances2 := getTheValueFromCondition(c2, Instances)
	if isSharedHostsBetweenConditionInstances(instances1, instances2) {
		return fmt.Errorf("the same host can not exist in both \"report\" and \"enforce\" modes")
	}
	return nil
}
