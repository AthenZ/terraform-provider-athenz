package athenz

import (
	"errors"
	"fmt"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/ardielle/ardielle-go/rdl"
	"log"
	"os"
	"testing"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGroupMetaBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	domainName := os.Getenv("DOMAIN")
	groupName := "test-group-meta"
	resourceName := "athenz_group_meta.test_group_meta"
	t.Cleanup(func() {
		cleanAccTestGroupMeta(domainName, groupName)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupMetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMetaConfigBasic(domainName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMetaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", domainName),
					resource.TestCheckResourceAttr(resourceName, "user_expiry_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "service_expiry_days", "70"),
					resource.TestCheckResourceAttr(resourceName, "max_members", "90"),
					resource.TestCheckResourceAttr(resourceName, "self_serve", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew_mins", "100"),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "true"),
					resource.TestCheckResourceAttr(resourceName, "review_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notify_roles", "admin,security"),
					resource.TestCheckResourceAttr(resourceName, "tags.zms.DisableExpirationNotifications", "4"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "test audit ref"),
				),
			},
		},
	})
}

func cleanAccTestGroupMeta(domainName, groupName string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	_, err := zmsClient.GetGroup(domainName, groupName)
	if err == nil {
		var zero int32
		zero = 0
		disabled := false
		groupMeta := zms.GroupMeta{
			SelfServe:               &disabled,
			MemberExpiryDays:        &zero,
			ServiceExpiryDays:       &zero,
			ReviewEnabled:           &disabled,
			NotifyRoles:             "",
			UserAuthorityFilter:     "",
			UserAuthorityExpiration: "",
			Tags:                    make(map[zms.TagKey]*zms.TagValueList),
			DeleteProtection:        &disabled,
			SelfRenew:               &disabled,
			SelfRenewMins:           &zero,
			MaxMembers:              &zero,
			AuditEnabled:            &disabled,
		}
		if err = zmsClient.PutGroupMeta(domainName, groupName, AUDIT_REF, &groupMeta); err != nil {
			log.Printf("unable to reset group meta for %s:group.%s: %v\n", domainName, groupName, err)
		}
	}
}

func testAccCheckGroupMetaExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group ID is set")
		}
		dn, gn, err := splitGroupId(rs.Primary.ID)
		if err != nil {
			return err
		}
		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		group, err := zmsClient.GetGroup(dn, gn)
		if err != nil {
			return err
		}
		if group.NotifyRoles == "" {
			return fmt.Errorf("does not have notify roles set")
		}
		return nil
	}
}

func testAccCheckGroupMetaDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_group_meta" {
			continue
		}
		dn, gn, err := splitGroupId(rs.Primary.ID)
		if err != nil {
			return err
		}
		group, err := zmsClient.GetGroup(dn, gn)
		if err != nil {
			return err
		}
		if group.NotifyRoles != "" {
			return fmt.Errorf("athenz group meta still exists")
		}
		_ = zmsClient.DeleteGroup(dn, gn, AUDIT_REF)
	}

	return nil
}

func testAccGroupMetaConfigBasic(domainName, groupName string) string {
	return fmt.Sprintf(`
resource "athenz_group_meta" "test_group_meta" {
  domain = "%s"
  name = "%s"
  user_expiry_days = 30
  service_expiry_days = 70
  max_members = 90
  self_serve = true
  self_renew = true
  self_renew_mins = 100
  delete_protection = true
  review_enabled = true
  notify_roles = "admin,security"
  tags = {
    "zms.DisableExpirationNotifications" = "4"
  }
  audit_ref = "test audit ref"
}
`, domainName, groupName)
}

func TestAccGroupMetaResourceStateDelete(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	domainName := os.Getenv("DOMAIN")
	groupName := "test-group-meta-delete"
	resourceName := "athenz_group_meta.test_group_meta_delete"
	t.Cleanup(func() {
		cleanAccTestGroupMeta(domainName, groupName)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupMetaResourceStateDeleteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMetaConfigResourceStateDelete(domainName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMetaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", domainName),
					resource.TestCheckResourceAttr(resourceName, "user_expiry_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "service_expiry_days", "70"),
					resource.TestCheckResourceAttr(resourceName, "max_members", "90"),
					resource.TestCheckResourceAttr(resourceName, "self_serve", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew_mins", "100"),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "true"),
					resource.TestCheckResourceAttr(resourceName, "review_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notify_roles", "admin,security"),
					resource.TestCheckResourceAttr(resourceName, "tags.zms.DisableExpirationNotifications", "4"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "test audit ref"),
				),
			},
		},
	})
}

func testAccCheckGroupMetaResourceStateDeleteDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_group_meta" {
			continue
		}
		dn, gn, err := splitGroupId(rs.Primary.ID)
		if err != nil {
			return err
		}
		// make sure our group is deleted and 404 is returned
		_, err = zmsClient.GetGroup(dn, gn)
		if err == nil {
			_ = zmsClient.DeleteGroup(dn, gn, AUDIT_REF)
			return fmt.Errorf("athenz group still exists")
		}
		var v rdl.ResourceError
		switch {
		case errors.As(err, &v):
			if v.Code == 404 {
				return nil
			}
		}
		return fmt.Errorf("unexpected error: %v", err)
	}

	return nil
}

func testAccGroupMetaConfigResourceStateDelete(domainName, groupName string) string {
	return fmt.Sprintf(`
resource "athenz_group_meta" "test_group_meta_delete" {
  domain = "%s"
  name = "%s"
  user_expiry_days = 30
  service_expiry_days = 70
  max_members = 90
  self_serve = true
  self_renew = true
  self_renew_mins = 100
  delete_protection = true
  review_enabled = true
  notify_roles = "admin,security"
  tags = {
    "zms.DisableExpirationNotifications" = "4"
  }
  resource_state = 3
  audit_ref = "test audit ref"
}
`, domainName, groupName)
}
