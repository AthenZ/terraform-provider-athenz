package athenz

import (
	"fmt"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"os"
	"testing"
)

func TestAccGroupPolicyDataSource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	var policy zms.Policy
	resourceName := "athenz_policy.policyTest"
	dataSourceName := "data.athenz_policy.policyTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	policyName := fmt.Sprintf("test%d", rInt)
	roleName := fmt.Sprintf("test%d", rInt)
	t.Cleanup(func() {
		cleanAllAccTestPolicies(domainName, []string{policyName}, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyDataSourceConfig(policyName, domainName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttrPair(resourceName, "domain", dataSourceName, "domain"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.#", dataSourceName, "assertion.#"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.0.effect", dataSourceName, "assertion.0.effect"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.0.action", dataSourceName, "assertion.0.action"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.0.role", dataSourceName, "assertion.0.role"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.0.resource", dataSourceName, "assertion.0.resource"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.0.case_sensitive", dataSourceName, "assertion.0.case_sensitive"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.1.effect", dataSourceName, "assertion.1.effect"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.1.action", dataSourceName, "assertion.1.action"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.1.role", dataSourceName, "assertion.1.role"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.1.resource", dataSourceName, "assertion.1.resource"),
					resource.TestCheckResourceAttrPair(resourceName, "assertion.1.case_sensitive", dataSourceName, "assertion.1.case_sensitive"),
				),
			},
		},
	})
}

func testAccGroupPolicyDataSourceConfig(name, domain, roleName string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
	name = "%s"
	domain = "%s"
}

resource "athenz_policy" "policyTest" {
  name = "%s"
  domain = "%s"
  assertion {
    effect = "DENY"
    action = "PLAY"
    role = athenz_role.roleTest.name
    resource = "sys.auth:ows"
    case_sensitive = true
  }  
 assertion {
    effect = "ALLOW"
    action = "*"
    role = athenz_role.roleTest.name
    resource = "sys.auth:ows"
  }
}

data "athenz_policy" "policyTest" {
  domain = athenz_policy.policyTest.domain
  name = athenz_policy.policyTest.name
}
`, roleName, domain, name, domain)
}
