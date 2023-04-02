package athenz

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

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
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccGroupRoleConfigChangeAuditRef(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupRoleConfigAddTags(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}, "key2": {"b1", "b2"}}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveTags(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
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
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}, {"name": member2, "expiration": "2022-12-29 23:59:59", "review": ""}}),
				),
			},
			{
				Config: testAccGroupRoleConfigRemoveMember(roleName, domainName, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member2, "expiration": "2022-12-29 23:59:59", "review": ""}}),
				),
			},
			{
				Config: testAccGroupRoleConfigAddMemberWithReview(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": "2022-12-29 23:59:59"}, {"name": member2, "expiration": "", "review": ""}}),
				),
			},
			{
				Config: testAccGroupRoleConfigAddCertExpirySettings(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": "2022-12-29 23:59:59"}, {"name": member2, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"cert_expiry_mins": "75"}),
				),
			},
			{
				Config: testAccGroupRoleConfigAddUserReviewDaysSettings(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": "2022-12-29 23:59:59"}, {"name": member2, "expiration": "", "review": "2023-01-29 23:59:59"}}),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"cert_expiry_mins": "75", "user_review_days": "45"}),
				),
			},
		},
	})
}

func TestAccRoleSettings(t *testing.T) {
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
	var role zms.Role
	resourceName := "athenz_role.roleTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	roleName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	t.Cleanup(func() {
		cleanAllAccTestRoles(domainName, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleConfigWithAllSetting(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "2022-12-29 23:59:59", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"token_expiry_mins": "5", "cert_expiry_mins": "10", "user_expiry_days": "90"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithAllSettingChanged(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"token_expiry_mins": "30", "cert_expiry_mins": "75"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithTokenExpirySetting(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"token_expiry_mins": "30"}),
				),
			},
			{
				Config: testAccGroupRoleConfig(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "0"),
				),
			},
		},
	})
}

func TestAccRoleSettingsStartWithOneEditAndReplace(t *testing.T) {
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
	var role zms.Role
	resourceName := "athenz_role.roleTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	roleName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	t.Cleanup(func() {
		cleanAllAccTestRoles(domainName, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleConfigWithTokenExpirySetting(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"token_expiry_mins": "30"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithTokenExpirySettingChanged(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"token_expiry_mins": "5"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithCertExpirySetting(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"cert_expiry_mins": "75"}),
				),
			},
		},
	})
}

func TestAccRoleSettingsMemberExpiryAndReview(t *testing.T) {
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
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "0"),
				),
			},
			{
				Config: testAccGroupRoleConfigWithUserExpiryDaysSettingChangeMemberExpiration(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "2023-03-29 23:59:59", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"user_expiry_days": "30"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithUserExpiryDaysSetting(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "2022-12-29 23:59:59", "review": ""}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"user_expiry_days": "30"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithUserExpiryDaysAndUserReviewDays(roleName, domainName, member1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "2022-12-29 23:59:59", "review": "2021-12-29 23:59:59"}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"user_expiry_days": "30", "user_review_days": "70"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithUserExpiryDaysAndUserReviewDaysRemoveMember(roleName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"user_expiry_days": "30", "user_review_days": "70"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithUserReviewDaysTwoMembers(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": "2021-12-29 23:59:59"}, {"name": member2, "expiration": "2022-12-29 23:59:59", "review": "2020-12-29 23:59:59"}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"user_review_days": "35"}),
				),
			},
			{
				Config: testAccGroupRoleConfigWithUserExpiryDaysAndReviewDaysTwoMembers(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "2022-12-29 23:59:59", "review": "2021-12-29 23:59:59"}, {"name": member2, "expiration": "2022-12-29 23:59:59", "review": "2020-12-29 23:59:59"}}),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "done by someone"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					testAccCheckCorrectSettings(resourceName, map[string]string{"user_review_days": "35", "user_expiry_days": "7"}),
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
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}, {"name": member2, "expiration": "2022-12-29 23:59:59", "review": ""}}),
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
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member2, "expiration": "2022-12-29 23:59:59", "review": ""}}),
				),
			},
			{
				Config: testAccGroupRoleConfigAddMemberWithReview(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", roleName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": "2022-12-29 23:59:59"}, {"name": member2, "expiration": "", "review": ""}}),
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
					testAccCheckCorrectGroupMembers(resourceName, []map[string]string{{"name": member1, "expiration": "", "review": ""}, {"name": member2, "expiration": "2022-12-29 23:59:59", "review": "2022-12-29 23:59:59"}}),
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

	date := time.Now().AddDate(0, 0, 14).Format(EXPIRATION_LAYOUT)

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
			{
				Config:      testAccGroupRoleInvalidReviewConfig(),
				ExpectError: getPatternErrorRegex(MEMBER_REVIEW_REMINDER),
			},
			{
				Config:      testAccGroupRoleInvalidSettingsConfig(),
				ExpectError: regexp.MustCompile("expected settings.0.token_expiry_mins to be at least \\(1\\), got 0"),
			},
			{
				Config:      testAccGroupRoleUserExpirationAfterSettingUserExpirationDays(date),
				ExpectError: getErrorRegex("one or more user is set past the user_expiry_days limit: "),
			},
			{
				Config:      testAccGroupRoleUserExpirationNotSetButSettingUserExpirationDaysDefined(),
				ExpectError: getErrorRegex("settings.user_expiry_days is defined but for one or more user isn't set"),
			},
			{
				Config:      testAccGroupRoleGroupReviewAfterSettingGroupReviewDays(date),
				ExpectError: getErrorRegex("one or more group is set past the group_review_days limit: "),
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

func testAccCheckCorrectSettings(n string, lookingForSettings map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
		}
		expectedSettings := make([]map[string]string, 1)
		// for build the expected members, we look for all attribute from the following pattern: member.<index>.<attribute> (e.g. member.0.expiration)
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

func testAccGroupRoleConfigWithTokenExpirySetting(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  settings {
	token_expiry_mins = 30
  }  
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigWithTokenExpirySettingChanged(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  settings {
	token_expiry_mins = 5
  }  
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigWithCertExpirySetting(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  settings {
	cert_expiry_mins = 75
  }  
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigWithAllSettingChanged(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  settings {
	token_expiry_mins = 30
	cert_expiry_mins = 75
  }  
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigWithAllSetting(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
  settings {
	token_expiry_mins = 5
	cert_expiry_mins = 10
	user_expiry_days = 90
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

func testAccGroupRoleConfigAddMemberWithReview(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
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
  tags = {
	key1 = "a1,a2"
	}
}
`, name, domain, member1, member2)
}

func testAccGroupRoleConfigAddCertExpirySettings(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
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
  settings {
	cert_expiry_mins = 75
  }  
  tags = {
	key1 = "a1,a2"
	}
}
`, name, domain, member1, member2)
}

func testAccGroupRoleConfigAddUserReviewDaysSettings(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	review = "2022-12-29 23:59:59"
  }  
  member {
	name = "%s"
	review = "2023-01-29 23:59:59"
  }
  settings {
	cert_expiry_mins = 75
	user_review_days = 45
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
	review = "2022-12-29 23:59:59"
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

func testAccGroupRoleInvalidReviewConfig() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "user.jone"
		review = "2022-01-01 13:56"
	}
}
`)
}

func testAccGroupRoleInvalidSettingsConfig() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "user.jone"
	}
	settings {
		token_expiry_mins = 0
		cert_expiry_mins = 60
  	} 
}
`)
}

func testAccGroupRoleUserExpirationAfterSettingUserExpirationDays(date string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "user.jone"
		expiration = "%s"
	}
	settings {
		user_expiry_days = 2
  	} 
}
`, date)
}

func testAccGroupRoleUserExpirationNotSetButSettingUserExpirationDaysDefined() string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "user.jone"
	}
	settings {
		user_expiry_days = 25
  	} 
}
`)
}

func testAccGroupRoleGroupReviewAfterSettingGroupReviewDays(date string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	domain = "sys.auth"
	name = "acc.test"
	member {
		name = "dummy:group.jone"
		review = "%s"
	}
	settings {
		group_review_days = 7
  	} 
}
`, date)
}

func testAccGroupRoleConfigWithUserExpiryDaysSetting(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }  
  settings {
	user_expiry_days = 30
  }
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigWithUserExpiryDaysSettingChangeMemberExpiration(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2023-03-29 23:59:59"
  }  
  settings {
	user_expiry_days = 30
  }
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigWithUserExpiryDaysAndUserReviewDays(name, domain, member1 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
	review = "2021-12-29 23:59:59"
  }  
  settings {
	user_expiry_days = 30
	user_review_days = 70
  }
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1)
}

func testAccGroupRoleConfigWithUserExpiryDaysAndUserReviewDaysRemoveMember(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  settings {
	user_expiry_days = 30
	user_review_days = 70
  }
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain)
}

func testAccGroupRoleConfigWithUserReviewDaysTwoMembers(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	review = "2021-12-29 23:59:59"
  }
  member {
	name = "%s"
	review = "2020-12-29 23:59:59"
	expiration = "2022-12-29 23:59:59"
  }
  settings {
	user_review_days = 35
  }
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1, member2)
}

func testAccGroupRoleConfigWithUserExpiryDaysAndReviewDaysTwoMembers(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	review = "2021-12-29 23:59:59"
	expiration = "2022-12-29 23:59:59"
  }
  member {
	name = "%s"
	review = "2020-12-29 23:59:59"
	expiration = "2022-12-29 23:59:59"
  }
  settings {
	user_review_days = 35
	user_expiry_days = 7
  }
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
}
`, name, domain, member1, member2)
}
