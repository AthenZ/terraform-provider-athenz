package athenz

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
)

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

	// we don't want to delete a member that should be added right after
	membersToNotDelete := stringSet{}
	for _, member := range add {
		membersToNotDelete.add(string(member.MemberName))
	}

	if len(remove) > 0 {
		for _, member := range remove {
			if !membersToNotDelete.contains(string(member.MemberName)) {
				name := member.MemberName
				err := zmsClient.DeleteGroupMembership(dn, gn, name, auditRef)
				if err != nil {
					return fmt.Errorf("Error removing membership: %s", err)
				}
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
