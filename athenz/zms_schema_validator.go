package athenz

import (
	"fmt"
	"regexp"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/ardielle/ardielle-go/rdl"
)

const (
	DOMAIN_NAME               = "DomainName"
	ENTTITY_NAME              = "EntityName"
	SIMPLE_NAME               = "SimpleName"
	MEMBER_NAME               = "MemberName"
	GROUP_MEMBER_NAME         = "GroupMemberName"
	ASSERTION_CONDITION_VALUE = "AssertionConditionValue"
)

var rdlSchema *rdl.Schema

var regexValidatorCache map[string]*regexp.Regexp

func init() {
	rdlSchema = zms.ZMSSchema()
	regexValidatorCache = map[string]*regexp.Regexp{
		DOMAIN_NAME:               buildRegexFromRdlSchema(DOMAIN_NAME),
		ENTTITY_NAME:              buildRegexFromRdlSchema(ENTTITY_NAME),
		SIMPLE_NAME:               buildRegexFromRdlSchema(SIMPLE_NAME),
		MEMBER_NAME:               buildRegexFromRdlSchema(MEMBER_NAME),
		GROUP_MEMBER_NAME:         buildRegexFromRdlSchema(GROUP_MEMBER_NAME),
		ASSERTION_CONDITION_VALUE: buildRegexFromRdlSchema(ASSERTION_CONDITION_VALUE),
	}
}

func buildRegexFromRdlSchema(stringType string) *regexp.Regexp {
	for _, t := range rdlSchema.Types {
		if t.StringTypeDef != nil && string(t.StringTypeDef.Name) == stringType {
			re, _ := regexp.Compile(t.StringTypeDef.Pattern)
			re.Longest()
			return re
		}
	}
	return nil
}

func validatePatternFunc(attribute string) schema.SchemaValidateDiagFunc {
	return func(val interface{}, c cty.Path) diag.Diagnostics {
		re := regexValidatorCache[attribute]
		if re == nil {
			return diag.FromErr(fmt.Errorf("regex not found for attribute %s", attribute))
		}
		if re.FindString(val.(string)) != val.(string) {
			return diag.FromErr(fmt.Errorf("%s must match the pattern %s", attribute, re.String()))
		}
		return nil
	}
}
