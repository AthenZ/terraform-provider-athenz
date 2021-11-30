package athenz

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandGroupMembers(configured []interface{}) []*zms.GroupMember {
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
func flattenGroupMember(list []*zms.GroupMember) []interface{} {
	groupMember := make([]interface{}, 0, len(list))
	for _, member := range list {
		groupMember = append(groupMember, string(member.MemberName))
	}
	return groupMember
}
func updateGroupMembers(dn string, gn string, oldVal interface{}, newVal interface{}, zmsClient client.ZmsClient, auditRef string) error {
	if oldVal == nil {
		oldVal = new(schema.Set)
	}
	if newVal == nil {
		newVal = new(schema.Set)
	}

	os := oldVal.(*schema.Set)
	ns := newVal.(*schema.Set)
	remove := expandGroupMembers(os.Difference(ns).List())
	add := expandGroupMembers(ns.Difference(os).List())

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
			member.GroupName = zms.ResourceName(gn)
			err := zmsClient.PutGroupMembership(dn, gn, name, auditRef, &member)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
