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

func TestAccRoleMetaBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	domainName := os.Getenv("DOMAIN")
	roleName := "test-role-meta"
	resourceName := "athenz_role_meta.test_role_meta"
	t.Cleanup(func() {
		cleanAccTestRoleMeta(domainName, roleName)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoleMetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMetaConfigBasic(domainName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleMetaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "test role"),
					resource.TestCheckResourceAttr(resourceName, "token_expiry_mins", "10"),
					resource.TestCheckResourceAttr(resourceName, "cert_expiry_mins", "20"),
					resource.TestCheckResourceAttr(resourceName, "user_expiry_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "user_review_days", "40"),
					resource.TestCheckResourceAttr(resourceName, "group_expiry_days", "50"),
					resource.TestCheckResourceAttr(resourceName, "group_review_days", "60"),
					resource.TestCheckResourceAttr(resourceName, "service_expiry_days", "70"),
					resource.TestCheckResourceAttr(resourceName, "service_review_days", "80"),
					resource.TestCheckResourceAttr(resourceName, "max_members", "90"),
					resource.TestCheckResourceAttr(resourceName, "self_serve", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew_mins", "100"),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "true"),
					resource.TestCheckResourceAttr(resourceName, "review_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notify_roles", "admin,security"),
					resource.TestCheckResourceAttr(resourceName, "notify_details", "notify details"),
					resource.TestCheckResourceAttr(resourceName, "principal_domain_filter", "user,sys.auth"),
					resource.TestCheckResourceAttr(resourceName, "tags.zms.DisableExpirationNotifications", "4"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "test audit ref"),
				),
			},
		},
	})
}

func cleanAccTestRoleMeta(domainName, roleName string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	_, err := zmsClient.GetRole(domainName, roleName)
	if err == nil {
		var zero int32
		zero = 0
		disabled := false
		roleMeta := zms.RoleMeta{
			SelfServe:               &disabled,
			MemberExpiryDays:        &zero,
			TokenExpiryMins:         &zero,
			CertExpiryMins:          &zero,
			SignAlgorithm:           "",
			ServiceExpiryDays:       &zero,
			MemberReviewDays:        &zero,
			ServiceReviewDays:       &zero,
			ReviewEnabled:           &disabled,
			NotifyRoles:             "",
			NotifyDetails:           "",
			UserAuthorityFilter:     "",
			UserAuthorityExpiration: "",
			GroupExpiryDays:         &zero,
			GroupReviewDays:         &zero,
			Tags:                    make(map[zms.TagKey]*zms.TagValueList),
			Description:             "",
			DeleteProtection:        &disabled,
			SelfRenew:               &disabled,
			SelfRenewMins:           &zero,
			MaxMembers:              &zero,
			AuditEnabled:            &disabled,
			PrincipalDomainFilter:   "",
		}
		if err = zmsClient.PutRoleMeta(domainName, roleName, AUDIT_REF, &roleMeta); err != nil {
			log.Printf("unable to reset role meta for %s:role.%s: %v\n", domainName, roleName, err)
		}
	}
}

func testAccCheckRoleMetaExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Role ID is set")
		}
		dn, rn, err := splitRoleId(rs.Primary.ID)
		if err != nil {
			return err
		}
		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		role, err := zmsClient.GetRole(dn, rn)
		if err != nil {
			return err
		}
		if role.Description == "" {
			return fmt.Errorf("does not have description set")
		}
		return nil
	}
}

func testAccCheckRoleMetaDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_role_meta" {
			continue
		}
		dn, rn, err := splitRoleId(rs.Primary.ID)
		if err != nil {
			return err
		}
		role, err := zmsClient.GetRole(dn, rn)
		if err != nil {
			return err
		}
		if role.Description != "" {
			return fmt.Errorf("athenz role meta still exists")
		}
		_ = zmsClient.DeleteRole(dn, rn, AUDIT_REF)
	}

	return nil
}

func testAccRoleMetaConfigBasic(domainName, roleName string) string {
	return fmt.Sprintf(`
resource "athenz_role_meta" "test_role_meta" {
  domain = "%s"
  name = "%s"
  description = "test role"
  token_expiry_mins = 10
  cert_expiry_mins = 20
  user_expiry_days = 30
  user_review_days = 40
  group_expiry_days = 50
  group_review_days = 60
  service_expiry_days = 70
  service_review_days = 80
  max_members = 90
  self_serve = true
  self_renew = true
  self_renew_mins = 100
  delete_protection = true
  review_enabled = true
  notify_roles = "admin,security"
  notify_details = "notify details"
  principal_domain_filter = "user,sys.auth"
  tags = {
    "zms.DisableExpirationNotifications" = "4"
  }
  audit_ref = "test audit ref"
}
`, domainName, roleName)
}

func TestAccRoleMetaResourceStateDelete(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	domainName := os.Getenv("DOMAIN")
	roleName := "test-role-meta-delete"
	resourceName := "athenz_role_meta.test_role_meta_delete"
	t.Cleanup(func() {
		cleanAccTestRoleMeta(domainName, roleName)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoleMetaResourceStateDeleteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMetaConfigResourceStateDelete(domainName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleMetaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "test role"),
					resource.TestCheckResourceAttr(resourceName, "token_expiry_mins", "10"),
					resource.TestCheckResourceAttr(resourceName, "cert_expiry_mins", "20"),
					resource.TestCheckResourceAttr(resourceName, "user_expiry_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "user_review_days", "40"),
					resource.TestCheckResourceAttr(resourceName, "group_expiry_days", "50"),
					resource.TestCheckResourceAttr(resourceName, "group_review_days", "60"),
					resource.TestCheckResourceAttr(resourceName, "service_expiry_days", "70"),
					resource.TestCheckResourceAttr(resourceName, "service_review_days", "80"),
					resource.TestCheckResourceAttr(resourceName, "max_members", "90"),
					resource.TestCheckResourceAttr(resourceName, "self_serve", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_renew_mins", "100"),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "true"),
					resource.TestCheckResourceAttr(resourceName, "review_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notify_roles", "admin,security"),
					resource.TestCheckResourceAttr(resourceName, "notify_details", "notify details"),
					resource.TestCheckResourceAttr(resourceName, "principal_domain_filter", "user"),
					resource.TestCheckResourceAttr(resourceName, "tags.zms.DisableExpirationNotifications", "4"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "test audit ref"),
				),
			},
		},
	})
}

func testAccCheckRoleMetaResourceStateDeleteDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_role_meta" {
			continue
		}
		dn, rn, err := splitRoleId(rs.Primary.ID)
		if err != nil {
			return err
		}
		// make sure our role is deleted and 404 is returned
		_, err = zmsClient.GetRole(dn, rn)
		if err == nil {
			_ = zmsClient.DeleteRole(dn, rn, AUDIT_REF)
			return fmt.Errorf("athenz role still exists")
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

func testAccRoleMetaConfigResourceStateDelete(domainName, roleName string) string {
	return fmt.Sprintf(`
resource "athenz_role_meta" "test_role_meta_delete" {
  domain = "%s"
  name = "%s"
  description = "test role"
  token_expiry_mins = 10
  cert_expiry_mins = 20
  user_expiry_days = 30
  user_review_days = 40
  group_expiry_days = 50
  group_review_days = 60
  service_expiry_days = 70
  service_review_days = 80
  max_members = 90
  self_serve = true
  self_renew = true
  self_renew_mins = 100
  delete_protection = true
  review_enabled = true
  notify_roles = "admin,security"
  notify_details = "notify details"
  principal_domain_filter = "user"
  tags = {
    "zms.DisableExpirationNotifications" = "4"
  }
  resource_state = 3
  audit_ref = "test audit ref"
}
`, domainName, roleName)
}
