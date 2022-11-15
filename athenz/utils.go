package athenz

import (
	b64 "encoding/base64"
	"fmt"
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
				},
			},
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

func shortName(domainName string, en string, separator string) string {
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
		member := map[string]interface{}{
			"name":       name,
			"expiration": expiration,
		}
		roleMembers = append(roleMembers, member)
	}
	return roleMembers
}

func timestampToString(timeStamp *rdl.Timestamp) string {
	if timeStamp == nil {
		return ""
	}
	str := timeStamp.Time.String()
	// <yyyy>-<mm>-<dd> <hh>:<mm>:<ss> +0000 UTC
	lastIndex := strings.Index(str, "+") - 1
	return str[0:lastIndex]
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
			"key_value": convertToDecodedKey(val.Key) + "\n",
		}
		publicKeys = append(publicKeys, publicKey)

	}
	return publicKeys
}

func convertToKeyBase64(keyValue string) string {
	keyBytes := []byte(strings.Trim(keyValue, "\n"))
	encodeChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._"
	return b64.NewEncoding(encodeChars).WithPadding('-').EncodeToString(keyBytes)
}

func convertToDecodedKey(keyValue string) string {
	keyBytes, _ := b64.StdEncoding.DecodeString(keyValue)
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

func deleteRoleMembers(dn string, rn string, members []*zms.RoleMember, auditRef string, zmsClient client.ZmsClient) error {
	if members != nil {
		for _, m := range members {
			if err := deleteRoleMember(dn, rn, m, auditRef, zmsClient); err != nil {
				return fmt.Errorf("error removing membership: %s", err)
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

// no double values
func compareStringSets(set1 []string, set2 []string) bool {
	if len(set1) != len(set2) {
		return false
	}
	check := false
	for _, val := range set1 {
		for _, val2 := range set2 {
			if val == val2 {
				check = true
				break
			}
		}
		if !check {
			return false
		}
		check = false
	}
	return true
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
	if zmsRole.Trust != "" {
		role["trust"] = string(zmsRole.Trust)
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

func validateExpirationPatternFunc(validPattern string, attribute string) schema.SchemaValidateDiagFunc {
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

func getPatternErrorRegex(attribute string) *regexp.Regexp {
	r, _ := regexp.Compile(fmt.Sprintf("Error: %s must match the pattern", attribute))
	return r
}
