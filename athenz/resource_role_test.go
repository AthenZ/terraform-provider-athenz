package athenz

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGroupRoleConflictArgumentError(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	r, e := regexp.Compile("Error: Conflicting configuration arguments")
	if e != nil {
		assert.Fail(t, e.Error())
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupRoleMembersConflictingMember(),
				ExpectError: r,
			},
		},
	})
}

func testAccGroupRoleMembersConflictingMember() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "test"
  domain = "sys.auth"
  members = ["user.jone"]
  member {
	name = "user.jone"
  }
}
`)
}
func TestAccGroupRoleBasicDeprecated(t *testing.T) {
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleConfigDeprecated(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccGroupRoleConfigChangeAuditRefDeprecated(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupRoleConfigAddTagsDeprecated(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}, "key2": {"b1", "b2"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveTagsDeprecated(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigAddMemberDeprecated(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembersDeprecated(resourceName, []string{member1, member2}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveMemberDeprecated(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembersDeprecated(resourceName, []string{member1}),
				),
			},
		},
	})
}

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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleConfig(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccGroupRoleConfigChangeAuditRef(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupRoleConfigAddTags(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}}),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}, "key2": {"b1", "b2"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveTags(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}}),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigAddMemberWithExpiration(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}, {"name": member2, "expiration": "2022-12-29 23:59:59"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveMember(roleName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member2, "expiration": "2022-12-29 23:59:59"}}),
				),
			},
		},
	})
}

func TestAccGroupRoleDelegation(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("DELEGATED_DOMAIN"); v == "" {
		t.Fatal("DELEGATED_DOMAIN must be set for acceptance tests")
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
	delegatedDomain := os.Getenv("DELEGATED_DOMAIN")
	t.Cleanup(func() {
		cleanAllAccTestRoles(domainName, []string{roleName})
	})

	// Switch between delegated and non-delegated, then back to delegated
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleConfigDelegated(roleName, domainName, delegatedDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resourceName, "trust", delegatedDomain),
				),
			},
			{
				Config: testAccGroupRoleConfigAddMemberWithExpiration(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					resource.TestCheckNoResourceAttr(resourceName, "trust"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}, {"name": member2, "expiration": "2022-12-29 23:59:59"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveMember(roleName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					resource.TestCheckNoResourceAttr(resourceName, "trust"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member2, "expiration": "2022-12-29 23:59:59"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigDelegated(roleName, domainName, delegatedDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resourceName, "trust", delegatedDomain),
				),
			},
		},
	})
}

func TestAccGroupRoleTransitionFromMembersToMember(t *testing.T) {
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleConfigUsingMembers(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembersDeprecated(resourceName, []string{member1, member2}),
				),
			},
			{
				Config: testAccGroupRoleConfigMoveToMember(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resourceName, "members.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}, {"name": member2, "expiration": "2022-12-29 23:59:59"}}),
				),
			},
		},
	})
}

func TestAccGroupRoleInvalidResource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupRoleInvalidDomainNameConfig(),
				ExpectError: getPatternErrorRegex(DOMAIN_NAME),
			},
			{
				Config:      testAccGroupRoleInvalidRoleNameConfig(),
				ExpectError: getPatternErrorRegex(ENTTITY_NAME),
			},
			{
				Config:      testAccGroupRoleInvalidMemberNameConfig(),
				ExpectError: getPatternErrorRegex(MEMBER_NAME),
			},
			{
				Config:      testAccGroupRoleInvalidExpirationConfig(),
				ExpectError: getPatternErrorRegex(MEMBER_EXPIRATION),
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

func testAccCheckCorrectGroupMembersDeprecated(n string, groupMembers []string) resource.TestCheckFunc {
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

// we implement this check function since we can't predict the order of the members
func testAccCheckCorrectGroupMembers(n string, lookingForMembers []map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
		}
		expectedMembers := make([]map[string]string, len(lookingForMembers))
		// for build the expected members, we look for all attribute from the following pattern: member.<index>.<attribute> (e.g. member.0.expiration)
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

func testAccGroupRoleConfigDeprecated(name, domain, member1 string) string {
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
func testAccGroupRoleConfigChangeAuditRefDeprecated(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
}
`, name, domain, member1)
}

func testAccGroupRoleConfigAddTagsDeprecated(name, domain, member1 string) string {
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
func testAccGroupRoleConfigRemoveTagsDeprecated(name, domain, member1 string) string {
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

func testAccGroupRoleConfigAddMemberDeprecated(name, domain, member1, member2 string) string {
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

func testAccGroupRoleConfigRemoveMemberDeprecated(name, domain string, member1 string) string {
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

func testAccGroupRoleConfig(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }  
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
  member {
	name = "%s"
  }  
}
`, name, domain, member1)
}

func testAccGroupRoleConfigAddTags(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }  
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
  member {
	name = "%s"
  }  
  tags = {
	key1 = "a1,a2"
  }
}
`, name, domain, member1)
}

func testAccGroupRoleConfigAddMemberWithExpiration(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }  
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
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
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
  tags = {
	key1 = "a1,a2"
  }
}
`, name, domain, member1)
}

func testAccGroupRoleConfigDelegated(name, domain, trust string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  trust = "%s"
  tags = {
	key1 = "a1,a2"
	}
}
`, name, domain, trust)
}

func testAccGroupRoleConfigUsingMembers(name, domain, member1, member2 string) string {
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
func testAccGroupRoleConfigMoveToMember(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
  tags = {
	key1 = "a1,a2"
	}
}
`, name, domain, member1, member2)
}

func testAccGroupRoleInvalidDomainNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.au@th"
	name = "acc.test"
}
`)
}

func testAccGroupRoleInvalidRoleNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc:test"
}
`)
}
func testAccGroupRoleInvalidMemberNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "user.jone"
	}
	member {
		name = "sys.auth:group.test"
	}
	member {
		name = "user:bob"
	}
}
`)
}

func testAccGroupRoleInvalidExpirationConfig() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "user.jone"
		expiration = "2022-01-01 13:56"
	}
}
`)
}
