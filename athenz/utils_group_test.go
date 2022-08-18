package athenz

import (
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ast "gotest.tools/assert"
)

// example for mocking client
func Test_updateGroupMembers(t *testing.T) {
	type args struct {
		dn        string
		gn        string
		d         *schema.ResourceData
		zmsClient client.ZmsClient
	}
	mockCtrl := gomock.NewController(t)
	clientMock := client.NewMockZmsClient(mockCtrl)
	clientMock.EXPECT().GetRole(gomock.Any(), gomock.Any()).Return(&zms.Role{Name: "test"}, nil).AnyTimes()
	clientMock.EXPECT().PutRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// _ = args{
	// 	zmsClient: clientMock,
	// }
}

func getFlattedGroupMembers() []interface{} {
	return []interface{}{"member1", "member2"}
}
func getZmsGroupMembers() []*zms.GroupMember {
	return []*zms.GroupMember{
		zms.NewGroupMember(&zms.GroupMember{MemberName: "member1"}),
		zms.NewGroupMember(&zms.GroupMember{MemberName: "member2"})}
}

func Test_expandGroupMembers(t *testing.T) {
	// case: regular test
	ast.DeepEqual(t, expandGroupMembers(getFlattedGroupMembers()), getZmsGroupMembers())

	// case: empty string test
	ast.DeepEqual(t, expandGroupMembers([]interface{}{""}), []*zms.GroupMember{})
}

func Test_flattenGroupMember(t *testing.T) {
	ast.DeepEqual(t, flattenGroupMember(getZmsGroupMembers()), getFlattedGroupMembers())
}
