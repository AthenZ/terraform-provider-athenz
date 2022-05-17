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

func TestAccGroupPolicyBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	var policy zms.Policy
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	resName := "athenz_policy.policyTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	name := fmt.Sprintf("test%d", rInt)
	resourceRoleName := "forPolicyTest"
	resourceRole := fmt.Sprintf(`resource "athenz_role" "%s" {
  			name = "%s"
  			domain = "%s"
		}`, resourceRoleName, name, domainName)
	t.Cleanup(func() {
		cleanAllAccTestPolicies(domainName, []string{name}, []string{name})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfigBasic(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "0"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupPolicyConfigChangeAuditRef(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "0"),
					resource.TestCheckResourceAttr(resName, "audit_ref", "done by someone"),
				),
			},
			{
				Config: testAccGroupConfigAddAssertion(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "2"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccGroupConfigRemoveAssertion(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
		},
	})
}

func cleanAllAccTestPolicies(domain string, policies, roles []string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	for _, policyName := range policies {
		_, err := zmsClient.GetPolicy(domain, policyName)
		if err == nil {
			if err = zmsClient.DeletePolicy(domain, policyName, AUDIT_REF); err != nil {
				log.Printf("error deleting Policy %s: %s", policyName, err)
			}
		}
	}
	for _, roleName := range roles {
		_, err := zmsClient.GetRole(domain, roleName)
		if err == nil {
			if err = zmsClient.DeleteRole(domain, roleName, AUDIT_REF); err != nil {
				log.Printf("error deleting Role %s: %s", roleName, err)
			}
		}
	}
}

func testAccCheckGroupPolicyExists(n string, p *zms.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Policy ID is set")
		}

		fullResourceName := strings.Split(rs.Primary.ID, POLICY_SEPARATOR)
		dn, pn := fullResourceName[0], fullResourceName[1]

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		policy, err := zmsClient.GetPolicy(dn, pn)

		if err != nil {
			return err
		}

		*p = *policy

		return nil
	}
}

func testAccCheckGroupPolicyDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_policy" {
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

func testAccGroupPolicyConfigBasic(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
}
`, name, domain)
}

func testAccGroupPolicyConfigChangeAuditRef(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_policy" "policyTest" {
  name = "%s"
  domain = "%s"
  audit_ref = "done by someone"
}
`, name, domain)
}

func testAccGroupConfigAddAssertion(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion = [{
    effect="ALLOW"
    action="*"
    role="${athenz_role.forPolicyTest.name}"
    resource="%sservice.ows"
  },{
    effect="DENY"
    action="play"
    role="${athenz_role.%s.name}"
    resource="%sservice.ows"
  }]
}
`, resourceRole, name, domain, domain+RESOURCE_SEPARATOR, resourceRoleName, domain+RESOURCE_SEPARATOR)
}

func testAccGroupConfigRemoveAssertion(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion = [{
    effect="DENY"
    action="*"
    role="${athenz_role.%s.name}"
    resource="%sservice.ows"
  }]
}
`, resourceRole, name, domain, resourceRoleName, domain+RESOURCE_SEPARATOR)
}
