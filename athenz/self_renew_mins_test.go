package athenz

import (
	"context"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ast "gotest.tools/assert"
)

// TestReadSelfRenewMins verifies that the role/group resources refresh
// self_renew_mins from ZMS, so out-of-band changes are detected as drift and
// the value is populated on terraform import.
func TestReadSelfRenewMins(t *testing.T) {
	selfRenew := true
	selfRenewMins := int32(120)

	tests := []struct {
		name     string
		resource *schema.Resource
		id       string
		read     func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	}{
		{"role", ResourceRole(), dName + ROLE_SEPARATOR + "role1", resourceRoleRead},
		{"role_meta", ResourceRoleMeta(), dName + ROLE_SEPARATOR + "role1", resourceRoleMetaRead},
		{"group", ResourceGroup(), dName + GROUP_SEPARATOR + "group1", resourceGroupRead},
		{"group_meta", ResourceGroupMeta(), dName + GROUP_SEPARATOR + "group1", resourceGroupMetaRead},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			clientMock := client.NewMockZmsClient(mockCtrl)
			clientMock.EXPECT().GetRole(dName, "role1", gomock.Any(), gomock.Any()).Return(&zms.Role{
				Name:          zms.ResourceName(dName + ROLE_SEPARATOR + "role1"),
				SelfRenew:     &selfRenew,
				SelfRenewMins: &selfRenewMins,
			}, nil).AnyTimes()
			clientMock.EXPECT().GetGroup(dName, "group1", gomock.Any()).Return(&zms.Group{
				Name:          zms.ResourceName(dName + GROUP_SEPARATOR + "group1"),
				SelfRenew:     &selfRenew,
				SelfRenewMins: &selfRenewMins,
			}, nil).AnyTimes()

			// state as it looks right after terraform import, or after the
			// value was changed outside of terraform
			d := schema.TestResourceDataRaw(t, tt.resource.Schema, map[string]interface{}{})
			d.SetId(tt.id)

			diags := tt.read(context.Background(), d, clientMock)
			ast.Assert(t, !diags.HasError(), diags)
			ast.Equal(t, d.Get("self_renew").(bool), true)
			ast.Equal(t, d.Get("self_renew_mins").(int), 120)
		})
	}
}
