package athenz

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGroupRoleBasic(t *testing.T) {
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
	var role zms.Role
	resourceName := "athenz_role.roleTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	roleName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	t.Cleanup(func() {
		cleanAllAccTestRoles(domainName, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleConfig(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccGroupRoleConfigChangeAuditRef(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupRoleConfigAddTags(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}, "key2": {"b1", "b2"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveTags(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigAddMember(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []string{member1, member2}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveMember(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []string{member1}),
				),
			},
		},
	})
}

func cleanAllAccTestRoles(domain string, roles []string) {
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

func testAccCheckGroupRoleExists(n string, r *zms.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
		}

		fullResourceName := strings.Split(string(rs.Primary.ID), ROLE_SEPARATOR)
		dn, rn := fullResourceName[0], fullResourceName[1]

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		role, err := zmsClient.GetRole(dn, rn)

		if err != nil {
			return err
		}

		*r = *role

		return nil
	}
}

func testAccCheckCorrectnessOfSet(n string, expectedSet []string, keyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
		}
		check := false
		for key, val := range rs.Primary.Attributes {
			theKeyArr := strings.Split(string(key), ".")
			if checkMemberKey := theKeyArr[0]; checkMemberKey == keyName {
				if theKeyArr[1] != "#" {
					for _, checkGroupMember := range expectedSet {
						if val == checkGroupMember {
							check = true
							break
						}
					}
					if !check {
						return fmt.Errorf("the member %s is not found", val)
					}
					check = false
				}
			}
		}
		return nil
	}
}

func makeTestTags(attributes map[string]string) map[string][]string {
	tagsSet := map[string][]string{}
	allVales := map[string][]string{}
	allKeys := map[string]string{}
	for key, val := range attributes {
		theKeyArr := strings.Split(string(key), ".")
		if len(theKeyArr) > 2 {
			if theKeyArr[0] == "tags" {
				if theKeyArr[2] == "values" {
					if theKeyArr[3] != "#" {
						if allVales[theKeyArr[1]] == nil {
							allVales[theKeyArr[1]] = []string{val}
						} else {
							allVales[theKeyArr[1]] = append(allVales[theKeyArr[1]], val)
						}
					}
				} else {
					allKeys[theKeyArr[1]] = val
				}
			}
		}
	}
	for key, val := range allKeys {
		tagsSet[val] = allVales[key]
	}
	return tagsSet
}

func testAccCheckCorrectTags(n string, expectedTagsMap map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
		}
		toCheckTagMaps := makeTestTags(rs.Primary.Attributes)
		for key, valArr := range toCheckTagMaps {
			if !compareStringSets(valArr, expectedTagsMap[key]) {
				return fmt.Errorf("the key %s is not equal for the expected", key)
			}
		}

		return nil
	}
}
func testAccCheckCorrectGroupMembers(n string, groupMembers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
		}
		check := false
		for key, val := range rs.Primary.Attributes {
			theKeyArr := strings.Split(string(key), ".")
			if checkMemberKey := theKeyArr[0]; checkMemberKey == "members" {
				if theKeyArr[1] != "#" {
					for _, checkGroupMember := range groupMembers {
						if val == checkGroupMember {
							check = true
							break
						}
					}
					if !check {
						return fmt.Errorf("the member %s is Not found", val)
					}
					check = false
				}
			}
		}

		return nil
	}
}

func testAccCheckGroupRoleDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_role" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, ROLE_SEPARATOR)
		dn, rn := fullResourceName[0], fullResourceName[1]

		_, err := zmsClient.GetRole(dn, rn)
		if err == nil {
			return fmt.Errorf("athenz Group Role still exists")
		}
	}

	return nil
}

func testAccGroupRoleConfig(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}
func testAccGroupRoleConfigChangeAuditRef(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
}
`, name, domain, member1)
}

func testAccGroupRoleConfigAddTags(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
  tags = {
	key1 = "a1,a2"
	key2 = "b1,b2"
	}
}
`, name, domain, member1)
}
func testAccGroupRoleConfigRemoveTags(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
name = "%s"
domain = "%s"
members = ["%s"]
  tags = {
	key1 = "a1,a2"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigAddMember(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  members = ["%s","%s"]
  tags = {
	key1 = "a1,a2"
	}
}
`, name, domain, member1, member2)
}

func testAccGroupRoleConfigRemoveMember(name, domain string, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
  tags = {
	key1 = "a1,a2"
	}
}
`, name, domain, member1)
}
