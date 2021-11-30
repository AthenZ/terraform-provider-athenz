package athenz

import (
	b64 "encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
)

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

func splitServiceId(serviceId string) (string, string) {
	return splitId(serviceId, SERVICE_SEPARATOR)
}

func splitSubDomainId(subDomainId string) (string, string) {
	return splitId(subDomainId, SUB_DOMAIN_SEPARATOR)
}

func splitId(id, separator string) (string, string) {
	indexOfPrefixEnd := strings.LastIndex(id, separator)
	prefix := id[:indexOfPrefixEnd]
	shortName := id[indexOfPrefixEnd+1:]
	return prefix, shortName
}

// adapted from https://github.com/yahoo/athenz/blob/master/libs/go/zmscli/utils.go
// Copyright 2016 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.
func expandRoleMembers(configured []interface{}) []*zms.RoleMember {
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

func flattenRoleMembers(list []*zms.RoleMember) []interface{} {
	roleMembers := make([]interface{}, 0, len(list))
	for _, m := range list {
		roleMembers = append(roleMembers, string(m.MemberName))
	}
	return roleMembers
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

// adapted from https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_autoscaling_group.go
func updateRoleMembers(dn string, rn string, remove []*zms.RoleMember, add []*zms.RoleMember, auditRef string, zmsClient client.ZmsClient) error {
	if len(remove) > 0 {
		for _, m := range remove {
			name := m.MemberName
			err := zmsClient.DeleteMembership(dn, rn, name, auditRef)
			if err != nil {
				return fmt.Errorf("error removing membership: %s", err)
			}
		}
	}
	if len(add) > 0 {
		for _, m := range add {
			var member zms.Membership
			name := m.MemberName
			member.MemberName = name
			member.RoleName = zms.ResourceName(rn)
			err := zmsClient.PutMembership(dn, rn, name, auditRef, &member)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//no double values
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
		role["members"] = flattenRoleMembers(zmsRole.RoleMembers)
	}
	if len(zmsRole.Tags) > 0 {
		role["tags"] = flattenTag(zmsRole.Tags)
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
