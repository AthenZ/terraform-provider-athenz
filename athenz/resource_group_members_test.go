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

func TestAccGroupMembersBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
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
	resName := "athenz_group_members.groupTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	groupName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	err := createTestGroupForMembers(domainName, groupName)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanAllAccTestGroupMembers(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembers(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembersExists(resName),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resName, "member.0.name", member1),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
				),
			},
			{
				Config: testAccGroupMembersChangeAuditRef(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembersExists(resName),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resName, "member.0.name", member1),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
				),
			},
			{
				Config: testAccGroupMembersAddMember(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembersExists(resName),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "2"),
				),
			},
			{
				Config: testAccGroupMembersRemoveMember(groupName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembersExists(resName),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "member.0.name", member2),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", "2022-12-29 23:59:59"),
				),
			},
		},
	})
}

func createTestGroupForMembers(dn, gn string) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	group := zms.Group{
		Name: zms.ResourceName(gn),
	}
	return zmsClient.PutGroup(dn, gn, AUDIT_REF, &group)
}

func cleanAllAccTestGroupMembers(domain string, groups []string) {
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

func testAccCheckGroupMembersExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group ID is set")
		}

		fullResourceName := strings.Split(rs.Primary.ID, GROUP_SEPARATOR)
		dn, gn := fullResourceName[0], fullResourceName[1]

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		_, err := zmsClient.GetGroup(dn, gn)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckGroupMembersDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_group_members" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, GROUP_SEPARATOR)
		dn, gn := fullResourceName[0], fullResourceName[1]

		group, err := zmsClient.GetGroup(dn, gn)
		if err == nil {
			if group.GroupMembers != nil && len(group.GroupMembers) > 0 {
				return fmt.Errorf("athenz Group Members still exists")
			}
			_ = zmsClient.DeleteGroup(dn, gn, AUDIT_REF)
		}
	}

	return nil
}

func testAccGroupMembers(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group_members" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
}
`, name, domain, member1)
}

func testAccGroupMembersChangeAuditRef(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group_members" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  audit_ref = "done by someone"
}
`, name, domain, member1)
}

func testAccGroupMembersAddMember(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group_members" "groupTest" {
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

func testAccGroupMembersRemoveMember(name, domain, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group_members" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
}
`, name, domain, member2)
}
