package athenz

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSelfServeRoleMembersBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("MEMBER_1"); v == "" {
		t.Fatal("MEMBER_1 must be set for acceptance tests")
	}
	if v := os.Getenv("MEMBER_2"); v == "" {
		t.Fatal("MEMBER_2 must be set for acceptance tests")
	}
	resourceName := "athenz_self_serve_role_members.roleTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	roleName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	err := createTestSelfServeRoleForMembers(domainName, roleName)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanAllAccTestSelfServeRoleMembers(domainName, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSelfServeRoleMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSelfServeRoleMembersConfig(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectSelfServeRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccSelfServeRoleMembersConfigChangeAuditRef(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectSelfServeRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccSelfServeRoleMembersConfigAddMemberWithExpiration(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectSelfServeRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}, {"name": member2, "expiration": "2025-12-29 23:59:59", "review": ""}}),
				),
			},
			{
				Config: testAccSelfServeRoleMembersConfigRemoveMember(roleName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectSelfServeRoleMembers(resourceName, []map[string]string{{"name": member2, "expiration": "2025-12-29 23:59:59", "review": ""}}),
				),
			},
			{
				Config: testAccSelfServeRoleMembersConfigAddMemberWithReview(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectSelfServeRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": "2025-12-29 23:59:59"}, {"name": member2, "expiration": "", "review": ""}}),
				),
			},
		},
	})
}

func TestAccSelfServeRoleMembersExternalMembers(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("MEMBER_1"); v == "" {
		t.Fatal("MEMBER_1 must be set for acceptance tests")
	}
	if v := os.Getenv("MEMBER_2"); v == "" {
		t.Fatal("MEMBER_2 must be set for acceptance tests")
	}
	resourceName := "athenz_self_serve_role_members.roleTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	roleName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")

	err := createTestSelfServeRoleForMembers(domainName, roleName)
	if err != nil {
		t.Fatal(err)
	}

	// Add an external member outside of Terraform
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	membership := zms.Membership{
		MemberName: zms.MemberName(member2),
	}
	err = zmsClient.PutMembership(domainName, roleName, zms.MemberName(member2), AUDIT_REF, &membership)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanAllAccTestSelfServeRoleMembers(domainName, []string{roleName})
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSelfServeRoleExternalMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSelfServeRoleMembersConfig(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectSelfServeRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					// Verify that the external member is not affected by Terraform
					testAccCheckExternalMemberStillExists(domainName, roleName, member2),
				),
			},
		},
	})
}

func cleanAllAccTestSelfServeRoleMembers(domain string, roles []string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	for _, roleName := range roles {
		_, err := zmsClient.GetRole(domain, roleName)
		if err == nil {
			if err = zmsClient.DeleteRole(domain, roleName, AUDIT_REF); err != nil {
				log.Printf("error deleting Role %s: %s", roleName, err)
			}
		}
	}
}

func createTestSelfServeRoleForMembers(dn, rn string) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	role := zms.Role{
		Name: zms.ResourceName(rn),
	}
	return zmsClient.PutRole(dn, rn, AUDIT_REF, &role)
}

func testAccCheckSelfServeRoleMembersExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Role ID is set")
		}

		fullResourceName := strings.Split(rs.Primary.ID, ROLE_SEPARATOR)
		dn, rn := fullResourceName[0], fullResourceName[1]

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		_, err := zmsClient.GetRole(dn, rn)
		if err != nil {
			role := zms.Role{
				Name: zms.ResourceName(rn),
			}
			_ = zmsClient.PutRole(dn, rn, AUDIT_REF, &role)
			return err
		}

		return nil
	}
}

func testAccCheckExternalMemberStillExists(domain, role, member string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		roleData, err := zmsClient.GetRole(domain, role)
		if err != nil {
			return err
		}

		for _, roleMember := range roleData.RoleMembers {
			if string(roleMember.MemberName) == member {
				return nil // External member still exists
			}
		}
		return fmt.Errorf("external member %s not found in role", member)
	}
}

// we implement this check function since we can't predict the order of the members
func testAccCheckCorrectSelfServeRoleMembers(n string, lookingForMembers []map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Role ID is set")
		}
		expectedMembers := make([]map[string]string, len(lookingForMembers))
		// for build the expected members, we look for all attribute from the following
		// pattern: member.<index>.<attribute> (e.g. member.0.expiration)
		for key, val := range rs.Primary.Attributes {
			if !strings.HasPrefix(key, "member.") {
				continue
			}
			theKeyArr := strings.Split(key, ".")
			if len(theKeyArr) == 3 && theKeyArr[2] != "%" {
				index, err := strconv.Atoi(theKeyArr[1])
				if err != nil {
					return err
				}
				attributeKey := theKeyArr[2]
				attributeVal := val
				if expectedMembers[index] == nil {
					member := map[string]string{
						attributeKey: attributeVal,
					}
					expectedMembers[index] = member
				} else {
					expectedMembers[index][attributeKey] = attributeVal
				}
			}
		}

		for _, lookingForMember := range lookingForMembers {
			found := false
			for _, expectedMember := range expectedMembers {
				if reflect.DeepEqual(lookingForMember, expectedMember) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("the member %v is Not found", lookingForMember)
			}
		}

		return nil
	}
}

func testAccCheckSelfServeRoleMembersDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_self_serve_role_members" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, ROLE_SEPARATOR)
		dn, rn := fullResourceName[0], fullResourceName[1]

		role, err := zmsClient.GetRole(dn, rn)
		if err == nil {
			if len(role.RoleMembers) > 0 {
				return fmt.Errorf("athenz Self Serve Role Members still exists")
			}
			_ = zmsClient.DeleteRole(dn, rn, AUDIT_REF)
		}
	}

	return nil
}

func testAccCheckSelfServeRoleExternalMembersDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_self_serve_role_members" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, ROLE_SEPARATOR)
		dn, rn := fullResourceName[0], fullResourceName[1]

		role, err := zmsClient.GetRole(dn, rn)
		if err == nil {
			if len(role.RoleMembers) == 0 {
				return fmt.Errorf("athenz ext role member should be present")
			}
			_ = zmsClient.DeleteRole(dn, rn, AUDIT_REF)
		}
	}

	return nil
}

func testAccSelfServeRoleMembersConfig(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_role_members" "roleTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
  }
  audit_ref = "done by someone"
}
`, domain, name, member1)
}

func testAccSelfServeRoleMembersConfigChangeAuditRef(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_role_members" "roleTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
  }
}
`, domain, name, member1)
}

func testAccSelfServeRoleMembersConfigAddMemberWithExpiration(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_role_members" "roleTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
  }
  member {
    name = "%s"
    expiration = "2025-12-29 23:59:59"
  }
}
`, domain, name, member1, member2)
}

func testAccSelfServeRoleMembersConfigAddMemberWithReview(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_role_members" "roleTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
    review = "2025-12-29 23:59:59"
  }
  member {
    name = "%s"
  }
}
`, domain, name, member1, member2)
}

func testAccSelfServeRoleMembersConfigRemoveMember(name, domain string, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_role_members" "roleTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
    expiration = "2025-12-29 23:59:59"
  }
}
`, domain, name, member1)
}
