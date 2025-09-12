package athenz

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type MemberType uint8

const (
	USER = iota
	GROUP
	SERVICE
)

func (s MemberType) String() string {
	switch s {
	case USER:
		return "user"
	case GROUP:
		return "group"
	case SERVICE:
		return "service"
	default:
		return "Invalid member type"
	}
}

type SettingType uint8

const (
	EXPIRATION = iota
	REVIEW
)

func (s SettingType) String() string {
	switch s {
	case EXPIRATION:
		return "expiration"
	case REVIEW:
		return "review"
	default:
		return "Invalid setting type"
	}
}

func dataSourceRoleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain": {
			Type:     schema.TypeString,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"member": {
			Type:        schema.TypeSet,
			Description: "Athenz principal to be added as members",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"expiration": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},
					"review": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},
				},
			},
		},
		"settings": {
			Type:        schema.TypeSet,
			Description: "Advanced settings",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"token_expiry_mins": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"cert_expiry_mins": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"user_expiry_days": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"user_review_days": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"group_expiry_days": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"group_review_days": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"service_expiry_days": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"service_review_days": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"max_members": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
		},
		"self_serve": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"audit_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"self_renew": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"self_renew_mins": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"delete_protection": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"review_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"user_authority_filter": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"user_authority_expiration": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"notify_roles": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"notify_details": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"sign_algorithm": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"last_reviewed_date": {
			Type:        schema.TypeString,
			Description: "Last reviewed date for the role",
			Optional:    true,
		},
		"trust": {
			Type:        schema.TypeString,
			Description: "The domain, which this role is trusted to",
			Optional:    true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"principal_domain_filter": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
	}
}

func getGroupsNames(zmsGroupList []*zms.Group) []string {
	groupList := make([]string, 0, len(zmsGroupList))
	for _, group := range zmsGroupList {
		groupList = append(groupList, string(group.Name))
	}
	return groupList
}
func convertEntityNameListToStringList(entityNames []zms.EntityName) []string {
	rolesNames := make([]string, 0, len(entityNames))
	for _, entityName := range entityNames {
		rolesNames = append(rolesNames, string(entityName))
	}
	return rolesNames
}

func getShortName(domainName string, en string, separator string) string {
	shortName := en
	if strings.HasPrefix(shortName, domainName+separator) {
		shortName = shortName[len(domainName)+len(separator):]
	}
	return shortName
}

func splitServiceId(serviceId string) (string, string, error) {
	return splitId(serviceId, SERVICE_SEPARATOR)
}

func splitSubDomainId(subDomainId string) (string, string, error) {
	return splitId(subDomainId, SUB_DOMAIN_SEPARATOR)
}

func splitRoleId(roleId string) (string, string, error) {
	return splitId(roleId, ROLE_SEPARATOR)
}

func splitPolicyId(policyId string) (string, string, error) {
	return splitId(policyId, POLICY_SEPARATOR)
}

func splitGroupId(policyId string) (string, string, error) {
	return splitId(policyId, GROUP_SEPARATOR)
}

func splitId(id, separator string) (string, string, error) {
	indexOfPrefixEnd := strings.LastIndex(id, separator) // it used for all resource id (e.g. service), so we're looking for last index
	if indexOfPrefixEnd == -1 {
		return "", "", fmt.Errorf("id pattern mismatch. expected: <domain_name>%s<resource_name>", separator)
	}
	prefix := id[:indexOfPrefixEnd]
	shortName := id[indexOfPrefixEnd+len(separator):]
	return prefix, shortName, nil
}

func expandDeprecatedRoleMembers(configured []interface{}) []*zms.RoleMember {
	roleMembers := make([]*zms.RoleMember, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			roleMember := zms.NewRoleMember()
			roleMember.MemberName = zms.MemberName(val)
			roleMembers = append(roleMembers, roleMember)
		}
	}
	return roleMembers
}

func flattenDeprecatedRoleMembers(list []*zms.RoleMember) []interface{} {
	roleMembers := make([]interface{}, 0, len(list))
	for _, m := range list {
		roleMembers = append(roleMembers, string(m.MemberName))
	}
	return roleMembers
}

func expandRoleMembers(configured []interface{}) []*zms.RoleMember {
	roleMembers := make([]*zms.RoleMember, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(map[string]interface{})
		if ok {
			roleMember := zms.NewRoleMember()
			roleMember.MemberName = zms.MemberName(val["name"].(string))
			roleMember.Expiration = stringToTimestamp(val["expiration"].(string))
			roleMember.ReviewReminder = stringToTimestamp(val["review"].(string))
			roleMembers = append(roleMembers, roleMember)
		}
	}
	return roleMembers
}

func stringToTimestamp(val string) *rdl.Timestamp {
	if val == "" {
		return nil
	}
	expiration, _ := time.ParseInLocation(EXPIRATION_LAYOUT, val, time.UTC)
	return &rdl.Timestamp{Time: expiration}
}

func flattenRoleMembers(list []*zms.RoleMember) []interface{} {
	roleMembers := make([]interface{}, 0, len(list))
	for _, m := range list {
		name := string(m.MemberName)
		expiration := timestampToString(m.Expiration)
		review := timestampToString(m.ReviewReminder)
		member := map[string]interface{}{
			"name":       name,
			"expiration": expiration,
			"review":     review,
		}
		roleMembers = append(roleMembers, member)
	}
	return roleMembers
}

func flattenIntSettings(values map[string]int) []interface{} {
	settingsSchemaSet := make([]interface{}, 0, 1)
	settings := map[string]interface{}{}

	for key, value := range values {
		settings[key] = value
	}

	settingsSchemaSet = append(settingsSchemaSet, settings)
	return settingsSchemaSet
}

func timestampToString(timeStamp *rdl.Timestamp) string {
	if timeStamp == nil {
		return ""
	}
	str := timeStamp.Time.String()
	// <yyyy>-<mm>-<dd> <hh>:<mm>:<ss> +0000 UTC - without nanoseconds
	// <yyyy>-<mm>-<dd> <hh>:<mm>:<ss>.<SSSSSS> +0000 UTC - with nanoseconds
	lastIndex := strings.LastIndex(str, ".")
	if lastIndex == -1 {
		lastIndex = strings.Index(str, " +")
	}
	if lastIndex == -1 {
		return str
	} else {
		return str[0:lastIndex]
	}
}

func convertToPublicKeyEntryList(publicKeys []interface{}) []*zms.PublicKeyEntry {
	publicKeyEntryList := make([]*zms.PublicKeyEntry, 0, len(publicKeys))
	for _, val := range publicKeys {
		m := val.(map[string]interface{})
		publicKeyEntry := &zms.PublicKeyEntry{
			Id:  m["key_id"].(string),
			Key: convertToKeyBase64(m["key_value"].(string)),
		}
		publicKeyEntryList = append(publicKeyEntryList, publicKeyEntry)
	}
	return publicKeyEntryList
}

func flattenPublicKeyEntryList(publicKeysEntry []*zms.PublicKeyEntry) []interface{} {
	publicKeys := make([]interface{}, 0, len(publicKeysEntry))
	for _, val := range publicKeysEntry {
		publicKey := map[string]interface{}{
			"key_id":    val.Id,
			"key_value": convertToDecodedKey(val.Key),
		}
		publicKeys = append(publicKeys, publicKey)
	}
	return publicKeys
}

func convertToKeyBase64(keyValue string) string {
	keyBytes := []byte(keyValue)
	encodeChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._"
	return b64.NewEncoding(encodeChars).WithPadding('-').EncodeToString(keyBytes)
}

func convertToDecodedKey(keyValue string) string {
	encodeChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._"
	keyBytes, _ := b64.NewEncoding(encodeChars).WithPadding('-').DecodeString(keyValue)
	return string(keyBytes)
}

func deleteRoleMember(dn string, rn string, member *zms.RoleMember, auditRef string, zmsClient client.ZmsClient) error {
	name := member.MemberName
	err := zmsClient.DeleteMembership(dn, rn, name, auditRef)
	if err != nil {
		return fmt.Errorf("error removing membership: %s", err)
	}
	return nil
}

func deleteRoleMembers(dn string, rn string, members []*zms.RoleMember, auditRef string, zmsClient client.ZmsClient, membersToNotDelete stringSet) error {
	if members != nil {
		for _, m := range members {
			if !membersToNotDelete.contains(string(m.MemberName)) {
				if err := deleteRoleMember(dn, rn, m, auditRef, zmsClient); err != nil {
					return fmt.Errorf("error removing membership: %s", err)
				}
			}
		}
	}
	return nil
}

func addRoleMember(dn string, rn string, m *zms.RoleMember, auditRef string, zmsClient client.ZmsClient) error {
	var member zms.Membership
	name := m.MemberName
	member.MemberName = name
	member.RoleName = zms.ResourceName(rn)
	member.Expiration = m.Expiration
	member.ReviewReminder = m.ReviewReminder
	err := zmsClient.PutMembership(dn, rn, name, auditRef, &member)
	if err != nil {
		return err
	}
	return nil
}

func addRoleMembers(dn string, rn string, members []*zms.RoleMember, auditRef string, zmsClient client.ZmsClient) error {
	if members != nil {
		for _, m := range members {
			if err := addRoleMember(dn, rn, m, auditRef, zmsClient); err != nil {
				return fmt.Errorf("error removing membership: %s", err)
			}
		}
	}
	return nil
}

func flattenRoles(zmsRoles []*zms.Role, domainName string) []interface{} {
	roles := make([]interface{}, 0, len(zmsRoles))
	for _, role := range zmsRoles {
		roles = append(roles, flattenRole(role, domainName))
	}
	return roles
}

func flattenRole(zmsRole *zms.Role, domainName string) map[string]interface{} {
	role := make(map[string]interface{})
	role["domain"] = domainName
	role["name"] = zmsRole.Name
	if len(zmsRole.RoleMembers) > 0 {
		members := flattenRoleMembers(zmsRole.RoleMembers)
		if len(members) > 0 {
			role["member"] = members
		}
	}
	if len(zmsRole.Tags) > 0 {
		role["tags"] = flattenTag(zmsRole.Tags)
	}
	roleSettings := map[string]int{}
	if zmsRole.TokenExpiryMins != nil {
		roleSettings["token_expiry_mins"] = int(*zmsRole.TokenExpiryMins)
	}
	if zmsRole.CertExpiryMins != nil {
		roleSettings["cert_expiry_mins"] = int(*zmsRole.CertExpiryMins)
	}
	if zmsRole.MemberExpiryDays != nil {
		roleSettings["user_expiry_days"] = int(*zmsRole.MemberExpiryDays)
	}
	if zmsRole.MemberReviewDays != nil {
		roleSettings["user_review_days"] = int(*zmsRole.MemberReviewDays)
	}
	if zmsRole.GroupExpiryDays != nil {
		roleSettings["group_expiry_days"] = int(*zmsRole.GroupExpiryDays)
	}
	if zmsRole.GroupReviewDays != nil {
		roleSettings["group_review_days"] = int(*zmsRole.GroupReviewDays)
	}
	if zmsRole.ServiceExpiryDays != nil {
		roleSettings["service_expiry_days"] = int(*zmsRole.ServiceExpiryDays)
	}
	if zmsRole.ServiceReviewDays != nil {
		roleSettings["service_review_days"] = int(*zmsRole.ServiceReviewDays)
	}
	if zmsRole.MaxMembers != nil {
		roleSettings["max_members"] = int(*zmsRole.MaxMembers)
	}
	if len(roleSettings) > 0 {
		role["settings"] = flattenIntSettings(roleSettings)
	}
	if zmsRole.Trust != "" {
		role["trust"] = string(zmsRole.Trust)
	}
	if zmsRole.PrincipalDomainFilter != "" {
		role["principal_domain_filter"] = zmsRole.PrincipalDomainFilter
	}
	return role
}

// input - the schema and the key that you want to get the changes from
// output - os- the old set , ns the new set
func handleChange(d *schema.ResourceData, key string) (*schema.Set, *schema.Set) {
	o, n := d.GetChange(key)
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	return os, ns
}

// validate that provides only role name (not fully qualified name)
func validateRoleNameWithinAssertion(roleName string) error {
	if strings.Contains(roleName, ROLE_SEPARATOR) {
		return fmt.Errorf("please provide only the role name without the domain prefix. the role is: %s", roleName)
	}
	return nil
}

// validate that provides fully qualified name
func validateResourceNameWithinAssertion(resourceName string) error {
	if !strings.Contains(resourceName, RESOURCE_SEPARATOR) {
		return fmt.Errorf("you must specify the fully qualified name for resource: %s", resourceName)
	}
	return nil
}

func validateDatePatternFunc(validPattern string, attribute string) schema.SchemaValidateDiagFunc {
	return func(val interface{}, c cty.Path) diag.Diagnostics {
		r, e := regexp.Compile(validPattern)
		if e != nil {
			return diag.FromErr(e)
		}
		if r.FindString(val.(string)) != val.(string) {
			return diag.FromErr(fmt.Errorf("%s must match the pattern %s", attribute, validPattern))
		}
		return nil
	}
}

func validateRoleMember(members []interface{}, settings map[string]interface{}) error {
	for _, mRaw := range members {
		data := mRaw.(map[string]interface{})
		name := data["name"].(string)

		expirationDays := 0
		reviewDays := 0
		var memberType MemberType

		if strings.HasPrefix(name, "user.") {
			memberType = USER
			if settings["user_expiry_days"] != nil {
				expirationDays = settings["user_expiry_days"].(int)
			}
			if settings["user_review_days"] != nil {
				reviewDays = settings["user_review_days"].(int)
			}
		} else if strings.Contains(name, ":group.") || strings.HasPrefix(name, "unix.") {
			memberType = GROUP
			if settings["group_expiry_days"] != nil {
				expirationDays = settings["group_expiry_days"].(int)
			}
			if settings["group_review_days"] != nil {
				reviewDays = settings["group_review_days"].(int)
			}
		} else {
			memberType = SERVICE
			if settings["service_expiry_days"] != nil {
				expirationDays = settings["service_expiry_days"].(int)
			}
			if settings["service_review_days"] != nil {
				reviewDays = settings["service_review_days"].(int)
			}
		}

		if err := validateMemberReviewAndExpiration(data, expirationDays, reviewDays, memberType); err != nil {
			return err
		}
	}
	return nil
}

func validateMemberReviewAndExpiration(memberData map[string]interface{}, expirationDays int, reviewDays int, memberType MemberType) error {
	expiration := memberData["expiration"].(string)
	review := memberData["review"].(string)

	if err := validateMemberDate(expirationDays, expiration, memberType, SettingType(EXPIRATION)); err != nil {
		return err
	}

	if err := validateMemberDate(reviewDays, review, memberType, SettingType(REVIEW)); err != nil {
		return err
	}

	return nil
}

func validateMemberDate(days int, dateString string, memberType MemberType, settingType SettingType) error {
	current := time.Now()

	settingAttr := fmt.Sprintf("%v_expiry_days", memberType)
	if settingType == REVIEW {
		settingAttr = fmt.Sprintf("%v_review_days", memberType)
	}

	if days > 0 {
		limit := current.AddDate(0, 0, days)
		if dateString == "" {
			return fmt.Errorf("settings.%s is defined but for one or more %v isn't set", settingAttr, memberType)
		}

		date, err := time.Parse(EXPIRATION_LAYOUT, dateString)
		if err != nil {
			return err
		}

		if limit.Before(date) {
			return fmt.Errorf("one or more %v is set past the %s limit: %s", memberType, settingAttr, limit)
		}
	}

	return nil
}

func readAfterWrite(readFunc func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics, ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := readFunc(ctx, d, meta)

	// 2 more retries after 5 and 10 seconds
	for i := 1; diags.HasError() && i < 3; i++ {
		time.Sleep(time.Duration(i) * time.Duration(5) * time.Second)
		log.Print("[WARN] resource did not found, about to try again")
		diags = readFunc(ctx, d, meta)
	}

	// if is a 404 error (not found)
	if diags.HasError() && strings.HasPrefix(diags[0].Summary, "404") {
		log.Printf("[WARN] Resource %s not found, removing from state", d.Id())
		d.SetId("")
	}
	return diags
}

func getTheValueFromCondition(condition map[string]interface{}, key string) string {
	return condition[key].(*schema.Set).List()[0].(map[string]interface{})["value"].(string)
}

func isAllHosts(instances string) bool {
	// If the instances is empty, it means all hosts.
	if len(instances) == 0 {
		return true
	}
	for _, host := range strings.Split(instances, ",") {
		if host == "*" {
			return true
		}
	}
	return false
}

func isSharedHostsBetweenConditionInstances(instances1, instances2 string) bool {
	if isAllHosts(instances1) || isAllHosts(instances2) {
		// If one of the instances includes all host, any host listed on the other condition will be shared with it.
		return true
	}

	set1 := stringSet{}
	for _, host := range strings.Split(instances1, ",") {
		set1.add(host)
	}
	for _, host := range strings.Split(instances2, ",") {
		if set1.contains(host) {
			return true
		}
	}
	return false
}
