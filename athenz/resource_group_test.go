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

	"github.com/ardielle/ardielle-go/rdl"

	"github.com/stretchr/testify/assert"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGroupConflictArgumentError(t *testing.T) {
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
				Config:      testAccGroupMembersConflictingMember(),
				ExpectError: r,
			},
		},
	})
}

func testAccGroupMembersConflictingMember() string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "test"
  domain = "sys.auth"
  members = ["user.jone"]
  member {
	name = "user.jone"
  }
}
`)
}

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
	now := rdl.TimestampNow()
	lastReviewedDate := timestampToString(&now)
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
					resource.TestCheckResourceAttr(resName, "principal_domain_filter", domainName),
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
			{
				Config: testAccGroupConfigAddTags(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", "2022-12-29 23:59:59"),
					testAccCheckCorrectTags(resName, map[string]string{"key1": "a1,a2", "key2": "b1,b2"}),
				),
			},
			{
				Config: testAccGroupConfigRemoveTags(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", "2022-12-29 23:59:59"),
					testAccCheckCorrectTags(resName, map[string]string{"key1": "a1,a2"}),
				),
			},
			{
				Config: testAccGroupConfigSettings(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", "2022-12-29 23:59:59"),
					resource.TestCheckResourceAttr(resName, "settings.#", "1"),
					testAccCheckCorrectGroupSettings(resName, map[string]string{"user_expiry_days": "10", "service_expiry_days": "20", "max_members": "30"}),
				),
			},
			{
				Config: testAccGroupConfigLastReviewedDate(groupName, domainName, member1, lastReviewedDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", "2022-12-29 23:59:59"),
					testAccCheckCorrectTags(resName, map[string]string{"key1": "a1,a2"}),
					resource.TestCheckResourceAttr(resName, "last_reviewed_date", lastReviewedDate),
				),
			},
		},
	})
}

func TestAccGroupAllAttributes(t *testing.T) {
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
				Config: testAccGroupConfigAllAttributes(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resName, "member.0.name", member1),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
					resource.TestCheckResourceAttr(resName, "principal_domain_filter", "user,"+domainName),
					resource.TestCheckResourceAttr(resName, "self_serve", "true"),
					resource.TestCheckResourceAttr(resName, "self_renew", "true"),
					resource.TestCheckResourceAttr(resName, "self_renew_mins", "100"),
					resource.TestCheckResourceAttr(resName, "delete_protection", "true"),
					resource.TestCheckResourceAttr(resName, "review_enabled", "false"),
					resource.TestCheckResourceAttr(resName, "notify_roles", "admin"),
					resource.TestCheckResourceAttr(resName, "notify_details", "notify details"),
					testAccCheckCorrectGroupSettings(resName, map[string]string{"user_expiry_days": "20", "max_members": "30"}),
				),
			},
			{
				Config: testAccGroupConfigAllAttributesChanged(groupName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resName, "member.0.name", member1),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
					resource.TestCheckResourceAttr(resName, "principal_domain_filter", "user,"+domainName),
					resource.TestCheckResourceAttr(resName, "self_serve", "false"),
					resource.TestCheckResourceAttr(resName, "self_renew", "false"),
					resource.TestCheckResourceAttr(resName, "self_renew_mins", "50"),
					resource.TestCheckResourceAttr(resName, "delete_protection", "false"),
					resource.TestCheckResourceAttr(resName, "review_enabled", "false"),
					resource.TestCheckResourceAttr(resName, "notify_roles", "admin"),
					resource.TestCheckResourceAttr(resName, "notify_details", "notify details"),
					testAccCheckCorrectGroupSettings(resName, map[string]string{"user_expiry_days": "15", "max_members": "20"}),
				),
			},
			{
				Config: testAccGroupConfigAllAttributesAddMember(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "2"),
					resource.TestCheckResourceAttr(resName, "principal_domain_filter", "user,"+domainName),
					resource.TestCheckResourceAttr(resName, "self_serve", "true"),
					resource.TestCheckResourceAttr(resName, "self_renew", "false"),
					resource.TestCheckResourceAttr(resName, "self_renew_mins", "50"),
					resource.TestCheckResourceAttr(resName, "delete_protection", "true"),
					resource.TestCheckResourceAttr(resName, "review_enabled", "false"),
					resource.TestCheckResourceAttr(resName, "notify_roles", ""),
					resource.TestCheckResourceAttr(resName, "notify_details", ""),
					resource.TestCheckResourceAttr(resName, "settings.#", "0"),
				),
			},
			{
				Config: testAccGroupConfigAllAttributesRemoveMember(groupName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "1"),
					resource.TestCheckResourceAttr(resName, "member.0.name", member2),
					resource.TestCheckResourceAttr(resName, "member.0.expiration", ""),
					resource.TestCheckResourceAttr(resName, "principal_domain_filter", "user,"+domainName),
					resource.TestCheckResourceAttr(resName, "self_serve", "true"),
					resource.TestCheckResourceAttr(resName, "self_renew", "false"),
					resource.TestCheckResourceAttr(resName, "self_renew_mins", "50"),
					resource.TestCheckResourceAttr(resName, "delete_protection", "true"),
					resource.TestCheckResourceAttr(resName, "review_enabled", "false"),
					resource.TestCheckResourceAttr(resName, "notify_roles", ""),
					resource.TestCheckResourceAttr(resName, "notify_details", ""),
					resource.TestCheckResourceAttr(resName, "settings.#", "0"),
				),
			},
		},
	})
}

func TestAccGroupReviewEnabled(t *testing.T) {
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
	t.Cleanup(func() {
		cleanAllAccTestGroups(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfigReviewEnabled(groupName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resName, &group),
					resource.TestCheckResourceAttr(resName, "name", groupName),
					resource.TestCheckResourceAttr(resName, "member.#", "0"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resName, "review_enabled", "true"),
					resource.TestCheckResourceAttr(resName, "notify_roles", "admin"),
					resource.TestCheckResourceAttr(resName, "notify_details", "notify details"),
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
				ExpectError: getPatternErrorRegex(ENTITY_NAME),
			},
			{
				Config:      testAccGroupInvalidMemberNameConfig(),
				ExpectError: getPatternErrorRegex(GROUP_MEMBER_NAME),
			},
			{
				Config:      testAccGroupInvalidExpirationConfig(),
				ExpectError: getPatternErrorRegex(MEMBER_EXPIRATION),
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
  tags = {
	key1 = "s1,s2"
	key2 = "s3,s4"
  }
  principal_domain_filter = "%s"
}
`, name, domain, member1, domain)
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
	member {
		name = "user.jone"
	}
	member {
		name = "sys.auth:group.test"
	}
}
`)
}

func testAccGroupInvalidExpirationConfig() string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "user.jone"
		expiration = "2022-01-01 13:56"
	}
}
`)
}

func testAccGroupConfigAddTags(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
   member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
  tags = {
	key1 = "a1,a2"
	key2 = "b1,b2"
  }
}
`, name, domain, member1)
}

func testAccGroupConfigRemoveTags(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
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

func testAccGroupConfigLastReviewedDate(name, domain, member1, lastReviewedDate string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
  tags = {
	key1 = "a1,a2"
  }
  last_reviewed_date = "%s"
}
`, name, domain, member1, lastReviewedDate)
}

func testAccGroupConfigSettings(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
  settings {
	user_expiry_days = 10
	service_expiry_days = 20
	max_members = 30
  }
}
`, name, domain, member1)
}

func testAccGroupConfigReviewEnabled(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  review_enabled = true
  notify_roles = "admin"
  notify_details = "notify details"
}
`, name, domain)
}

func testAccGroupConfigAllAttributes(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  tags = {
	key1 = "s1,s2"
	key2 = "s3,s4"
  }
  settings {
	user_expiry_days = 20
	max_members = 30
  }
  principal_domain_filter = "user,%s"
  self_serve = true
  self_renew = true
  self_renew_mins = 100
  delete_protection = true
  review_enabled = false
  notify_roles = "admin"
  notify_details = "notify details"
}
`, name, domain, member1, domain)
}

func testAccGroupConfigAllAttributesChanged(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  tags = {
	key1 = "s1,s2"
	key2 = "s3,s4"
  }
  settings {
	user_expiry_days = 15
	max_members = 20
  }
  principal_domain_filter = "user,%s"
  self_serve = false
  self_renew = false
  self_renew_mins = 50
  delete_protection = false
  review_enabled = false
  notify_roles = "admin"
  notify_details = "notify details"
  audit_ref = "done by someone"
}
`, name, domain, member1, domain)
}

func testAccGroupConfigAllAttributesAddMember(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  member {
	name = "%s"
  }
  tags = {
	key1 = "s1,s2"
	key2 = "s3,s4"
  }
  principal_domain_filter = "user,%s"
  self_serve = true
  self_renew = false
  self_renew_mins = 50
  delete_protection = true
  review_enabled = false
}
`, name, domain, member1, member2, domain)
}

func testAccGroupConfigAllAttributesRemoveMember(name, domain, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  tags = {
	key1 = "s1,s2"
	key2 = "s3,s4"
  }
  principal_domain_filter = "user,%s"
  self_serve = true
  self_renew = false
  self_renew_mins = 50
  delete_protection = true
  review_enabled = false
}
`, name, domain, member2, domain)
}

func testAccCheckCorrectGroupSettings(n string, lookingForSettings map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group ID is set")
		}
		expectedSettings := make([]map[string]string, 1)
		// for build the expected members, we look for all attribute from the following
		// pattern: member.<index>.<attribute> (e.g. member.0.expiration)
		for key, val := range rs.Primary.Attributes {
			if !strings.HasPrefix(key, "settings.") {
				continue
			}
			theKeyArr := strings.Split(key, ".")
			if len(theKeyArr) == 3 && theKeyArr[2] != "%" {
				_, err := strconv.Atoi(theKeyArr[1])
				if err != nil {
					return err
				}
				attributeKey := theKeyArr[2]
				attributeVal := val
				if attributeVal != "0" {
					if expectedSettings[0] == nil {
						settingsSchema := map[string]string{
							attributeKey: attributeVal,
						}
						expectedSettings[0] = settingsSchema
					} else {
						expectedSettings[0][attributeKey] = attributeVal
					}
				}
			}
		}

		if !reflect.DeepEqual(lookingForSettings, expectedSettings[0]) {
			return fmt.Errorf("the settings %v is Not found", lookingForSettings)
		}

		return nil
	}
}
