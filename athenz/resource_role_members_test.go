package athenz

import (
	"fmt"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestAccRoleMembersBasic(t *testing.T) {
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
	resourceName := "athenz_role_members.roleTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	roleName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	err := createTestRoleForMembers(domainName, roleName)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanAllAccTestRoleMembers(domainName, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoleMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMembersConfig(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccRoleMembersConfigChangeAuditRef(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccRoleMembersConfigAddMemberWithExpiration(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}, {"name": member2, "expiration": "2022-12-29 23:59:59", "review": ""}}),
				),
			},
			{
				Config: testAccRoleMembersConfigRemoveMember(roleName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectRoleMembers(resourceName, []map[string]string{{"name": member2, "expiration": "2022-12-29 23:59:59", "review": ""}}),
				),
			},
			{
				Config: testAccRoleMembersConfigAddMemberWithReview(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectRoleMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": "2022-12-29 23:59:59"}, {"name": member2, "expiration": "", "review": ""}}),
				),
			},
		},
	})
}

func cleanAllAccTestRoleMembers(domain string, roles []string) {
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

func createTestRoleForMembers(dn, rn string) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	role := zms.Role{
		Name: zms.ResourceName(rn),
	}
	return zmsClient.PutRole(dn, rn, AUDIT_REF, &role)
}

func testAccCheckRoleMembersExists(n string) resource.TestCheckFunc {
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

// we implement this check function since we can't predict the order of the members
func testAccCheckCorrectRoleMembers(n string, lookingForMembers []map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
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

func testAccCheckRoleMembersDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_role_members" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, ROLE_SEPARATOR)
		dn, rn := fullResourceName[0], fullResourceName[1]

		role, err := zmsClient.GetRole(dn, rn)
		if err == nil {
			if role.RoleMembers != nil && len(role.RoleMembers) > 0 {
				return fmt.Errorf("athenz Role Members still exists")
			}
			_ = zmsClient.DeleteRole(dn, rn, AUDIT_REF)
		}
	}

	return nil
}

func testAccRoleMembersConfig(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role_members" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }  
  audit_ref="done by someone"
}
`, name, domain, member1)
}

func testAccRoleMembersConfigChangeAuditRef(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role_members" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }  
}
`, name, domain, member1)
}

func testAccRoleMembersConfigAddMemberWithExpiration(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role_members" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }  
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
}
`, name, domain, member1, member2)
}

func testAccRoleMembersConfigAddMemberWithReview(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role_members" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	review = "2022-12-29 23:59:59"
  }  
  member {
	name = "%s"
	review = ""
  }
}
`, name, domain, member1, member2)
}

func testAccRoleMembersConfigRemoveMember(name, domain string, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role_members" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
}
`, name, domain, member1)
}
