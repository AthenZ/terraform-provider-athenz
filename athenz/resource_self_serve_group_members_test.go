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

func TestAccSelfServeGroupMembersBasic(t *testing.T) {
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
	resourceName := "athenz_self_serve_group_members.groupTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	groupName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	err := createTestSelfServeGroupForMembers(domainName, groupName)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanAllAccTestSelfServeGroupMembers(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSelfServeGroupMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSelfServeGroupMembersConfig(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeGroupMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectSelfServeGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccSelfServeGroupMembersConfigChangeAuditRef(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeGroupMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectSelfServeGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccSelfServeGroupMembersConfigAddMemberWithExpiration(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeGroupMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectSelfServeGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}, {"name": member2, "expiration": "2025-12-29 23:59:59"}}),
				),
			},
			{
				Config: testAccSelfServeGroupMembersConfigRemoveMember(groupName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeGroupMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectSelfServeGroupMembers(resourceName, []map[string]string{{"name": member2, "expiration": "2025-12-29 23:59:59"}}),
				),
			},
		},
	})
}

func TestAccSelfServeGroupMembersExternalMembers(t *testing.T) {
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
	resourceName := "athenz_self_serve_group_members.groupTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	groupName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")

	err := createTestSelfServeGroupForMembers(domainName, groupName)
	if err != nil {
		t.Fatal(err)
	}

	// Add an external member outside of Terraform
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	membership := zms.GroupMembership{
		MemberName: zms.GroupMemberName(member2),
	}
	err = zmsClient.PutGroupMembership(domainName, groupName, zms.GroupMemberName(member2), AUDIT_REF, &membership)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanAllAccTestSelfServeGroupMembers(domainName, []string{groupName})
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSelfServeGroupExternalMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSelfServeGroupMembersConfig(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelfServeGroupMembersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectSelfServeGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": ""}}),
					// Verify that the external member is not affected by Terraform
					testAccCheckExternalGroupMemberStillExists(domainName, groupName, member2),
				),
			},
		},
	})
}

func cleanAllAccTestSelfServeGroupMembers(domain string, groups []string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	for _, groupName := range groups {
		_, err := zmsClient.GetGroup(domain, groupName)
		if err == nil {
			if err = zmsClient.DeleteGroup(domain, groupName, AUDIT_REF); err != nil {
				log.Printf("error deleting Group %s: %s", groupName, err)
			}
		}
	}
}

func createTestSelfServeGroupForMembers(dn, gn string) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	group := zms.Group{
		Name: zms.ResourceName(gn),
	}
	return zmsClient.PutGroup(dn, gn, AUDIT_REF, &group)
}

func testAccCheckSelfServeGroupMembersExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group ID is set")
		}

		fullResourceName := strings.Split(rs.Primary.ID, GROUP_SEPARATOR)
		dn, gn := fullResourceName[0], fullResourceName[1]

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		_, err := zmsClient.GetGroup(dn, gn)
		if err != nil {
			group := zms.Group{
				Name: zms.ResourceName(gn),
			}
			_ = zmsClient.PutGroup(dn, gn, AUDIT_REF, &group)
			return err
		}

		return nil
	}
}

func testAccCheckExternalGroupMemberStillExists(domain, group, member string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		groupData, err := zmsClient.GetGroup(domain, group)
		if err != nil {
			return err
		}

		for _, groupMember := range groupData.GroupMembers {
			if string(groupMember.MemberName) == member {
				return nil // External member still exists
			}
		}
		return fmt.Errorf("external member %s not found in group", member)
	}
}

// we implement this check function since we can't predict the order of the members
func testAccCheckCorrectSelfServeGroupMembers(n string, lookingForMembers []map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group ID is set")
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

func testAccCheckSelfServeGroupMembersDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_self_serve_group_members" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, GROUP_SEPARATOR)
		dn, gn := fullResourceName[0], fullResourceName[1]

		group, err := zmsClient.GetGroup(dn, gn)
		if err == nil {
			if len(group.GroupMembers) > 0 {
				return fmt.Errorf("athenz Self Serve Group Members still exists")
			}
			_ = zmsClient.DeleteGroup(dn, gn, AUDIT_REF)
		}
	}

	return nil
}

func testAccCheckSelfServeGroupExternalMembersDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_self_serve_group_members" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, GROUP_SEPARATOR)
		dn, gn := fullResourceName[0], fullResourceName[1]

		group, err := zmsClient.GetGroup(dn, gn)
		if err == nil {
			if len(group.GroupMembers) == 0 {
				return fmt.Errorf("athenz ext group member should be present")
			}
			_ = zmsClient.DeleteGroup(dn, gn, AUDIT_REF)
		}
	}

	return nil
}

func testAccSelfServeGroupMembersConfig(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_group_members" "groupTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
  }
  audit_ref = "done by someone"
}
`, domain, name, member1)
}

func testAccSelfServeGroupMembersConfigChangeAuditRef(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_group_members" "groupTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
  }
}
`, domain, name, member1)
}

func testAccSelfServeGroupMembersConfigAddMemberWithExpiration(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_group_members" "groupTest" {
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

func testAccSelfServeGroupMembersConfigRemoveMember(name, domain string, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_self_serve_group_members" "groupTest" {
  domain = "%s"
  name = "%s"
  member {
    name = "%s"
    expiration = "2025-12-29 23:59:59"
  }
}
`, domain, name, member1)
}
