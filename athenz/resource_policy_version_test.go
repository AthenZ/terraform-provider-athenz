package athenz

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/AthenZ/terraform-provider-athenz/client"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccGroupPolicyVersionBasic(t *testing.T) {
	t.Skip("Skipping testing until docker image will be updated")
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	resName := "athenz_policy_version.policy_version_test"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	name := fmt.Sprintf("test%d", rInt)
	version1 := "test_version_1"
	version2 := "test_version_2"
	version3 := "test_version_3"
	t.Cleanup(func() {
		cleanAllAccTestPoliciesVersion(domainName, []string{name})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupPolicyVersionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyVersionConfigBasic(name, domainName, version1, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resName, []string{version1}),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "active_version", version1),
					resource.TestCheckResourceAttr(resName, "versions.#", "1"),
				),
			},
			{
				Config: testAccGroupPolicyVersionConfigAddAssertion(name, domainName, version1, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resName, []string{version1}),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "active_version", version1),
					resource.TestCheckResourceAttr(resName, "versions.#", "1"),
				),
			},
			{
				Config: testAccGroupPolicyVersionConfigAddNonActiveVersion(name, domainName, version1, version1, version2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resName, []string{version1, version2}),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "active_version", version1),
					resource.TestCheckResourceAttr(resName, "versions.#", "2"),
				),
			},
			{
				Config: testAccGroupPolicyVersionConfigChangeActiveVersion(name, domainName, version2, version1, version2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resName, []string{version1, version2}),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "active_version", version2),
					resource.TestCheckResourceAttr(resName, "versions.#", "2"),
				),
			},
			{
				Config: testAccGroupPolicyVersionConfigAddActiveVersion(name, domainName, version3, version1, version2, version3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resName, []string{version1, version2, version3}),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "active_version", version3),
					resource.TestCheckResourceAttr(resName, "versions.#", "3"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupPolicyVersionConfigRemoveNonActiveVersion(name, domainName, version3, version1, version3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resName, []string{version1, version3}),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "active_version", version3),
					resource.TestCheckResourceAttr(resName, "versions.#", "2"),
				),
			},
			{
				Config: testAccGroupPolicyVersionConfigRemovePreviousActiveVersion(name, domainName, version1, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resName, []string{version1}),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "active_version", version1),
					resource.TestCheckResourceAttr(resName, "versions.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
		},
	})
}

func cleanAllAccTestPoliciesVersion(domain string, policies []string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	for _, policyName := range policies {
		_, err := zmsClient.GetPolicy(domain, policyName)
		if err == nil {
			if err = zmsClient.DeletePolicy(domain, policyName, AUDIT_REF); err != nil {
				log.Printf("error deleting Policy %s: %s", policyName, err)
			}
		}
	}
}

func testAccCheckGroupPolicyVersionsExists(resourceName string, policyVersionNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Policy ID is set")
		}

		fullResourceName := strings.Split(rs.Primary.ID, POLICY_SEPARATOR)
		dn, pn := fullResourceName[0], fullResourceName[1]

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		for _, versionName := range policyVersionNames {
			_, err := zmsClient.GetPolicyVersion(dn, pn, versionName)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckGroupPolicyVersionsDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_policy_version" {
			continue
		}

		fullResourceName := strings.Split(rs.Primary.ID, POLICY_SEPARATOR)
		dn, pn := fullResourceName[0], fullResourceName[1]

		_, err := zmsClient.GetPolicy(dn, pn)

		if err == nil {
			return fmt.Errorf("athenz Policy still exists")
		}
	}

	return nil
}

func testAccGroupPolicyVersionConfigBasic(name, domain, activeVersion, version1 string) string {
	return fmt.Sprintf(`
resource "athenz_policy_version" "policy_version_test" {
name = "%s"
domain = "%s"
active_version = "%s"
versions{
	version_name = "%s"
	}
}`, name, domain, activeVersion, version1)
}
func testAccGroupPolicyVersionConfigAddAssertion(name, domain, activeVersion, version1 string) string {
	return fmt.Sprintf(`
resource "athenz_policy_version" "policy_version_test" {
name = "%s"
domain = "%s"
active_version = "%s"
versions {
  version_name = "%s"
  assertion = [{
    effect = "ALLOW"
    action = "*"
    role = "mendi_role1"
    resource = "mendi_resource1"
  }]
}
}`, name, domain, activeVersion, version1)
}
func testAccGroupPolicyVersionConfigAddNonActiveVersion(name, domain, activeVersion, version1, version2 string) string {
	return fmt.Sprintf(`
resource "athenz_policy_version" "policy_version_test" {
name = "%s"
domain = "%s"
active_version = "%s"
versions {
  version_name = "%s"
  assertion = [{
    effect = "ALLOW"
    action = "*"
    role = "mendi_role1"
    resource = "mendi_resource1"
  }]
}

versions {
  version_name = "%s"
  assertion  = [{
    effect = "ALLOW"
    action = "*"
    role = "mendi_role2"
    resource = "mendi_resource2"
  },
	{
    role = "mendi_role2"
    effect = "DENY"
    action = "play"
    resource = "mendi_resource2"
  }]
}
}
`, name, domain, activeVersion, version1, version2)
}

func testAccGroupPolicyVersionConfigChangeActiveVersion(name, domain, activeVersion, version1, version2 string) string {
	return fmt.Sprintf(`
resource "athenz_policy_version" "policy_version_test" {
name = "%s"
domain = "%s"
active_version = "%s"
versions {
 version_name = "%s"
 assertion = [{
   effect = "ALLOW"
   action = "*"
   role = "mendi_role1"
   resource = "mendi_resource1"
 }]
}

versions {
 version_name = "%s"
 assertion  = [{
   effect = "ALLOW"
   action = "*"
   role = "mendi_role2"
   resource = "mendi_resource2"
 },
 {
   role = "mendi_role2"
   effect = "DENY"
   action = "play"
   resource = "mendi_resource2"
 }]
}
}
`, name, domain, activeVersion, version1, version2)
}

func testAccGroupPolicyVersionConfigAddActiveVersion(name, domain, activeVersion, version1, version2, version3 string) string {
	return fmt.Sprintf(`
resource "athenz_policy_version" "policy_version_test" {
name = "%s"
domain = "%s"
active_version = "%s"
versions = [
  {
    version_name = "%s"
    assertion = [
      {
        effect = "ALLOW"
        action = "*"
        role = "mendi_role1"
        resource = "mendi_resource1"
      }]
  },
  {
    version_name = "%s"
    assertion = [
      {
        effect = "ALLOW"
        action = "*"
        role = "mendi_role2"
        resource = "mendi_resource2"
      },
		{
		 role = "mendi_role2"
		 effect = "DENY"
		 action = "play"
		 resource = "mendi_resource2"
      }
]
  },
  {
    version_name = "%s"
    assertion = [
      {
        effect = "ALLOW"
        action = "*"
        role = "mendi_role3"
        resource = "mendi_resource3"
      },
      {
        role = "mendi_role3"
        effect = "DENY"
        action = "play"
        resource = "mendi_resource3"
      }]
  }
]
}
`, name, domain, activeVersion, version1, version2, version3)
}

func testAccGroupPolicyVersionConfigRemoveNonActiveVersion(name, domain, activeVersion, version1, version3 string) string {
	return fmt.Sprintf(`
resource "athenz_policy_version" "policy_version_test" {
name = "%s"
domain = "%s"
active_version = "%s"
versions = [
  {
    version_name = "%s"
    assertion = [
      {
        effect = "ALLOW"
        action = "*"
        role = "mendi_role1"
        resource = "mendi_resource1"
      }]
  },
  {
    version_name = "%s"
    assertion = [
      {
        effect = "ALLOW"
        action = "*"
        role = "mendi_role3"
        resource = "mendi_resource3"
      },
      {
        role = "mendi_role3"
        effect = "DENY"
        action = "play"
        resource = "mendi_resource3"
      }]
  }
]
}
`, name, domain, activeVersion, version1, version3)
}

func testAccGroupPolicyVersionConfigRemovePreviousActiveVersion(name, domain, activeVersion, version1 string) string {
	return fmt.Sprintf(`
resource "athenz_policy_version" "policy_version_test" {
name = "%s"
domain = "%s"
active_version = "%s"
versions = [
  {
    version_name = "%s"
    assertion = [
      {
        effect = "ALLOW"
        action = "*"
        role = "mendi_role1"
        resource = "mendi_resource1"
      }]
  }
]
}
`, name, domain, activeVersion, version1)
}
