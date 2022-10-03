package athenz

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
)

// return true iff the members is includes in the members list
func isGroupMemberIncludes(member *zms.GroupMember, members []*zms.GroupMember) bool {
	for _, m := range members {
		if string(member.MemberName) == string(m.MemberName) {
			return true
		}
	}
	return false
}

// return all group members that appears in list1 but not includes in list2
func unifyGroupMembers(list1, list2 []*zms.GroupMember) []*zms.GroupMember {
	toReturn := make([]*zms.GroupMember, 0)
	for _, m := range list1 {
		if !isGroupMemberIncludes(m, list2) {
			toReturn = append(toReturn, m)
		}
	}
	return toReturn
}

/*
since we not allow configuring both members (deprecated) and member attributes at the same time, state changes CAN'T be one of the following:
1. remove members from both attributes member and members.
2. add members in both attributes member and members.
*/
func handleGroupMembersChange(removeDeprecatedRoleMembers, addDeprecatedRoleMembers, removeRoleMembers, addRoleMembers []*zms.GroupMember) ([]*zms.GroupMember, []*zms.GroupMember) {
	var removeMembers []*zms.GroupMember
	if len(removeDeprecatedRoleMembers) == 0 {
		removeMembers = unifyGroupMembers(removeRoleMembers, addDeprecatedRoleMembers)
	} else {
		removeMembers = unifyGroupMembers(removeDeprecatedRoleMembers, addRoleMembers)
	}
	if len(addDeprecatedRoleMembers) == 0 {
		return removeMembers, addRoleMembers
	}
	return removeMembers, addDeprecatedRoleMembers
}

func expandDeprecatedGroupMembers(configured []interface{}) []*zms.GroupMember {
	groupMembers := make([]*zms.GroupMember, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			groupMember := zms.NewGroupMember()
			groupMember.MemberName = zms.GroupMemberName(val)
			groupMembers = append(groupMembers, groupMember)
		}
	}
	return groupMembers
}

func flattenDeprecatedGroupMembers(list []*zms.GroupMember) []interface{} {
	groupMember := make([]interface{}, 0, len(list))
	for _, member := range list {
		groupMember = append(groupMember, string(member.MemberName))
	}
	return groupMember
}

func expandGroupMembers(configured []interface{}) []*zms.GroupMember {
	groupMembers := make([]*zms.GroupMember, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(map[string]interface{})
		if ok {
			groupMember := zms.NewGroupMember()
			groupMember.MemberName = zms.GroupMemberName(val["name"].(string))
			groupMember.Expiration = stringToTimestamp(val["expiration"].(string))
			groupMembers = append(groupMembers, groupMember)
		}
	}
	return groupMembers
}

func flattenGroupMembers(list []*zms.GroupMember) []interface{} {
	groupMembers := make([]interface{}, 0, len(list))
	for _, m := range list {
		name := string(m.MemberName)
		expiration := timestampToString(m.Expiration)
		member := map[string]interface{}{
			"name":       name,
			"expiration": expiration,
		}
		groupMembers = append(groupMembers, member)
	}
	return groupMembers
}

func updateGroupMembers(dn string, gn string, remove []*zms.GroupMember, add []*zms.GroupMember, zmsClient client.ZmsClient, auditRef string) error {
	if len(remove) > 0 {
		for _, member := range remove {
			name := member.MemberName
			err := zmsClient.DeleteGroupMembership(dn, gn, name, auditRef)
			if err != nil {
				return fmt.Errorf("Error removing membership: %s", err)
			}
		}
	}

	if len(add) > 0 {
		for _, m := range add {
			var member zms.GroupMembership
			name := m.MemberName
			member.MemberName = name
			member.Expiration = m.Expiration
			member.GroupName = zms.ResourceName(gn)
			err := zmsClient.PutGroupMembership(dn, gn, name, auditRef, &member)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
