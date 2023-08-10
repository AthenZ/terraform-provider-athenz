package athenz

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"os"
	"testing"
)

func TestAccGroupPolicyVersionDataSource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	resourceName := "athenz_policy_version.policyVersionTest"
	dataSourceName := "data.athenz_policy_version.policyVersionTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	policyName := fmt.Sprintf("test%d", rInt)
	roleName := fmt.Sprintf("test%d", rInt)
	t.Cleanup(func() {
		cleanAllAccTestPoliciesVersion(domainName, []string{policyName}, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupPolicyVersionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyVersionDataSourceConfig(policyName, domainName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyVersionsExists(resourceName, []string{"1", "2"}),
					resource.TestCheckResourceAttrPair(resourceName, "domain", dataSourceName, "domain"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "active_version", dataSourceName, "active_version"),
					resource.TestCheckResourceAttrPair(resourceName, "version.#", dataSourceName, "version.#"),
					resource.TestCheckResourceAttrPair(resourceName, "version.0.assertion.#", dataSourceName, "version.0.assertion.#"),
					resource.TestCheckResourceAttrPair(resourceName, "version.0.assertion.0.effect", dataSourceName, "version.0.assertion.0.effect"),
					resource.TestCheckResourceAttrPair(resourceName, "version.0.assertion.0.action", dataSourceName, "version.0.assertion.0.action"),
					resource.TestCheckResourceAttrPair(resourceName, "version.0.assertion.0.role", dataSourceName, "version.0.assertion.0.role"),
					resource.TestCheckResourceAttrPair(resourceName, "version.0.assertion.0.resource", dataSourceName, "version.0.assertion.0.resource"),
					resource.TestCheckResourceAttrPair(resourceName, "version.0.assertion.0.case_sensitive", dataSourceName, "version.0.assertion.0.case_sensitive"),
					resource.TestCheckResourceAttrPair(resourceName, "version.1.assertion.#", dataSourceName, "version.1.assertion.#"),
					resource.TestCheckResourceAttrPair(resourceName, "version.1.assertion.0.effect", dataSourceName, "version.1.assertion.0.effect"),
					resource.TestCheckResourceAttrPair(resourceName, "version.1.assertion.0.action", dataSourceName, "version.1.assertion.0.action"),
					resource.TestCheckResourceAttrPair(resourceName, "version.1.assertion.0.role", dataSourceName, "version.1.assertion.0.role"),
					resource.TestCheckResourceAttrPair(resourceName, "version.1.assertion.0.resource", dataSourceName, "version.1.assertion.0.resource"),
					resource.TestCheckResourceAttrPair(resourceName, "version.1.assertion.0.case_sensitive", dataSourceName, "version.1.assertion.0.case_sensitive"),
				),
			},
		},
	})
}

func testAccGroupPolicyVersionDataSourceConfig(name, domain, roleName string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	name = "%s"
	domain = "%s"
}

resource "athenz_policy_version" "policyVersionTest" {
	name = "%s"
	domain = "%s"
	active_version = "1"
	version {
		version_name = "1"
		assertion {
			effect = "DENY"
			action = "PLAY"
			role = athenz_role.roleTest.name
			resource = "sys.auth:ows"
			case_sensitive = true
		} 
	}
	version {
		version_name = "2"
		assertion {
			effect = "ALLOW"
			action = "*"
			role = athenz_role.roleTest.name
			resource = "sys.auth:ows"
		}
	}
}

data "athenz_policy_version" "policyVersionTest" {
  domain = athenz_policy_version.policyVersionTest.domain
  name = athenz_policy_version.policyVersionTest.name
}
`, roleName, domain, name, domain)
}
