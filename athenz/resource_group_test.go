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

func TestAccGroupBasicDeprecated(t *testing.T) {
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
	var group zms.Group
	resName := "athenz_group.groupTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	groupName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	t.Cleanup(func() {
		cleanAllAccTestGroups(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfigBasicDeprecated(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "members.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupConfigBasicChangeAuditRefDeprecated(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "members.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccGroupConfigAddMemberDeprecated(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "members.#", "2"),
				),
			},
			{
				Config: testAccGroupConfigRemoveMemberDeprecated(groupName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "members.#", "1"),
				),
			},
		},
	})
}

func TestAccGroupBasic(t *testing.T) {
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
	var group zms.Group
	resName := "athenz_group.groupTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	groupName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	t.Cleanup(func() {
		cleanAllAccTestGroups(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfigBasic(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resName, "member.0.name", member1),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
				),
			},
			{
				Config: testAccGroupConfigBasicChangeAuditRef(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resName, "member.0.name", member1),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
				),
			},
			{
				Config: testAccGroupConfigAddMember(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "2"),
				),
			},
			{
				Config: testAccGroupConfigRemoveMember(groupName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "member.0.name", member2),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", "2022-12-29 23:59:59"),
				),
			},
		},
	})
}

func TestAccGroupTransitionFromMembersToMember(t *testing.T) {
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
	var group zms.Group
	resName := "athenz_group.groupTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	groupName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	t.Cleanup(func() {
		cleanAllAccTestGroups(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfigUsingMembers(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "members.#", "2"),
					resource.TestCheckResourceAttr(resName, "member.#", "0"),
				),
			},
			{
				Config: testAccGroupConfigMoveToMember(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "members.#", "0"),
					resource.TestCheckResourceAttr(resName, "member.#", "2"),
					resource.TestCheckResourceAttr(resName, "member.0.name", member1),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
					resource.TestCheckResourceAttr(resName, "member.1.name", member2),
					resource.TestCheckResourceAttr(resName, "member.1.expiration", "2022-12-29 23:59:59"),
				),
			},
		},
	})
}

func TestAccGroupInvalidResource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupInvalidDomainNameConfig(),
				ExpectError: getPatternErrorRegex(DOMAIN_NAME),
			},
			{
				Config:      testAccGroupInvalidGroupNameConfig(),
				ExpectError: getPatternErrorRegex(ENTTITY_NAME),
			},
			{
				Config:      testAccGroupInvalidMemberNameConfig(),
				ExpectError: getPatternErrorRegex(GROUP_MEMBER_NAME),
			},
		},
	})
}

func cleanAllAccTestGroups(domain string, groups []string) {
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

func testAccCheckGroupExists(resourceName string, g *zms.Group) resource.TestCheckFunc {
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
		group, err := zmsClient.GetGroup(dn, gn)

		if err != nil {
			return err
		}

		*g = *group

		return nil
	}
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_group" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, GROUP_SEPARATOR)
		dn, gn := fullResourceName[0], fullResourceName[1]

		_, err := zmsClient.GetGroup(dn, gn)

		if err == nil {
			return fmt.Errorf("athenz Group still exists")
		}
	}

	return nil
}

func testAccGroupConfigBasicDeprecated(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
}
`, name, domain, member1)
}

func testAccGroupConfigBasicChangeAuditRefDeprecated(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
 audit_ref = "done by someone"
}
`, name, domain, member1)
}

func testAccGroupConfigAddMemberDeprecated(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  members = ["%s", "%s"]
}
`, name, domain, member1, member2)
}

func testAccGroupConfigRemoveMemberDeprecated(name, domain, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  members = ["%s"]
}
`, name, domain, member2)
}

func testAccGroupConfigBasic(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
}
`, name, domain, member1)
}

func testAccGroupConfigBasicChangeAuditRef(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  audit_ref = "done by someone"
}
`, name, domain, member1)
}

func testAccGroupConfigAddMember(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
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

func testAccGroupConfigRemoveMember(name, domain, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
}
`, name, domain, member2)
}

func testAccGroupConfigUsingMembers(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  members = ["%s", "%s"]
}
`, name, domain, member1, member2)
}
func testAccGroupConfigMoveToMember(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
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

func testAccGroupInvalidDomainNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
	domain = "sys.au@th"
	name = "acc.test"
}
`)
}

func testAccGroupInvalidGroupNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
	domain = "sys.auth"
	name = "acc:test"
}
`)
}

func testAccGroupInvalidMemberNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
	domain = "sys.auth"
	name = "acc.test"
    members = ["user.jone", "sys.auth:group.test", "user:bob"]
}
`)
}
