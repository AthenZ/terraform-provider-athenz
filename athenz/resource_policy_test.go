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
				Config: testAccPolicyConfigAddAssertion(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "2"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccPolicyConfigRemoveAssertion(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
			{
				Config: testAccPolicyConfigAddTags(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					testAccCheckCorrectTags(resName, map[string]string{"key1": "a1,a2", "key2": "b1,b2"}),
				),
			},
			{
				Config: testAccPolicyConfigRemoveTags(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					testAccCheckCorrectTags(resName, map[string]string{"key1": "a1,a2"}),
				),
			},
		},
	})
}

func TestAccGroupCreatePolicyWithAssertions(t *testing.T) {
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
				Config: testAccGroupConfigCreatePolicyWithAssertions(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "2"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
		},
	})
}

func TestAccGroupCreatePolicyCaseSensitiveAssertion(t *testing.T) {
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
				Config: testAccGroupConfigCreatePolicyWithCaseSensitiveAssertions(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "1"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resName, "assertion.0.action", "PLAY"),
					resource.TestCheckResourceAttr(resName, "assertion.0.resource", domainName+RESOURCE_SEPARATOR+"OWS"),
					resource.TestCheckResourceAttr(resName, "assertion.0.case_sensitive", "true"),
				),
			},
		},
	})
}

func TestAccGroupCreatePolicyWithAssertionConditions(t *testing.T) {
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
				Config: testAccGroupPolicyWithAssertionConditions(resourceRole, name, domainName, resourceRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(resName, &policy),
					resource.TestCheckResourceAttr(resName, "name", name),
					resource.TestCheckResourceAttr(resName, "assertion.#", "1"),
					resource.TestCheckResourceAttr(resName, "assertion.0.condition.#", "2"),
					resource.TestCheckResourceAttr(resName, "audit_ref", AUDIT_REF),
				),
			},
		},
	})
}

func TestAccGroupPolicyInvalidResource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupPolicyInvalidDomainNameConfig(),
				ExpectError: getPatternErrorRegex(DOMAIN_NAME),
			},
			{
				Config:      testAccGroupPolicyInvalidPolicyNameConfig(),
				ExpectError: getPatternErrorRegex(ENTTITY_NAME),
			},
			{
				Config:      testAccGroupPolicyInvalidResourceNameConfig(),
				ExpectError: getErrorRegex("you must specify the fully qualified name for resource"),
			},
			{
				Config:      testAccGroupPolicyInvalidRoleNameConfig(),
				ExpectError: getErrorRegex("please provide only the role name without the domain prefix"),
			},
			{
				Config:      testAccGroupPolicyInvalidCaseSensitive1Config(),
				ExpectError: getErrorRegex("enabling case_sensitive flag is allowed only if action or resource has capital letters"),
			},
			{
				Config:      testAccGroupPolicyInvalidCaseSensitive2Config(),
				ExpectError: getErrorRegex("capitalized action or resource allowed only when enabling case_sensitive flag"),
			},
			{
				Config:      testAccGroupPolicyInvalidEnforcementState(),
				ExpectError: getErrorRegex("expected value to be one of \\[report enforce\\]"),
			},
			{
				Config:      testAccGroupPolicyDifferentModesWithSameEnforcementState(),
				ExpectError: getErrorRegex("enforcement state can't be same for different conditions in a msd policy"),
			},
			{
				Config:      testAccGroupPolicySharedHostsBetweenModes1(),
				ExpectError: getErrorRegex("the same host can not exist in both \"report\" and \"enforce\" modes"),
			},
			{
				Config:      testAccGroupPolicySharedHostsBetweenModes2(),
				ExpectError: getErrorRegex("the same host can not exist in both \"report\" and \"enforce\" modes"),
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

func testAccPolicyConfigAddAssertion(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion {
    effect="ALLOW"
    action="*"
    role="${athenz_role.forPolicyTest.name}"
    resource="%sservice.ows"
  }
  assertion {
    effect="DENY"
    action="play"
    role="${athenz_role.%s.name}"
    resource="%sservice.ows"
  }
}
`, resourceRole, name, domain, domain+RESOURCE_SEPARATOR, resourceRoleName, domain+RESOURCE_SEPARATOR)
}

func testAccPolicyConfigRemoveAssertion(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion {
    effect="DENY"
    action="*"
    role="${athenz_role.%s.name}"
    resource="%sservice.ows"
  }
}
`, resourceRole, name, domain, resourceRoleName, domain+RESOURCE_SEPARATOR)
}

func testAccPolicyConfigAddTags(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion {
    effect="DENY"
    action="*"
    role="${athenz_role.%s.name}"
    resource="%sservice.ows"
  }
 tags = {
	key1 = "a1,a2"
	key2 = "b1,b2"
	}
}
`, resourceRole, name, domain, resourceRoleName, domain+RESOURCE_SEPARATOR)
}

func testAccPolicyConfigRemoveTags(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion {
    effect="DENY"
    action="*"
    role="${athenz_role.%s.name}"
    resource="%sservice.ows"
  }
tags = {
	key1 = "a1,a2"
  }
}
`, resourceRole, name, domain, resourceRoleName, domain+RESOURCE_SEPARATOR)
}

func testAccGroupConfigCreatePolicyWithAssertions(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion {
    effect="ALLOW"
    action="*"
    role="${athenz_role.forPolicyTest.name}"
    resource="%sservice.ows"
  }
 assertion {
    effect="DENY"
    action="play"
    role="${athenz_role.%s.name}"
    resource="%sservice.ows"
  }
}
`, resourceRole, name, domain, domain+RESOURCE_SEPARATOR, resourceRoleName, domain+RESOURCE_SEPARATOR)
}

func testAccGroupConfigCreatePolicyWithCaseSensitiveAssertions(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
name = "%s"
  domain = "%s"
  assertion {
    effect="DENY"
    action="PLAY"
    role="${athenz_role.%s.name}"
    resource="%sOWS"
    case_sensitive=true
  }
}
`, resourceRole, name, domain, resourceRoleName, domain+RESOURCE_SEPARATOR)
}

func testAccGroupPolicyInvalidDomainNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_policy" "policyTest" {
	domain = "sys.au@th"
	name = "acc.test"
}
`)
}

func testAccGroupPolicyInvalidPolicyNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_policy" "policyTest" {
	domain = "sys.auth"
	name = "acc:test"
}
`)
}

func testAccGroupPolicyInvalidResourceNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name = "test"
  domain = "sys.auth"
  assertion {
    effect="DENY"
    action="play"
    role="test"
    resource="ows"
  }
}
`)
}

func testAccGroupPolicyInvalidRoleNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name = "policy_test"
  domain = "sys.auth"
  assertion {
    effect="DENY"
    action="play"
    role="sys.auth:role.test"
    resource="sys.auth:ows"
  }
}
`)
}

func testAccGroupPolicyInvalidCaseSensitive1Config() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name = "test"
  domain = "sys.auth"
  assertion {
    effect="DENY"
    action="play"
    role="test"
    resource="sys.auth:ows"
    case_sensitive=true
  }
}
`)
}

func testAccGroupPolicyInvalidCaseSensitive2Config() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name = "policy_test"
  domain = "sys.auth"
  assertion {
    effect="DENY"
    action="PLAY"
    role="role_test"
    resource="sys.auth:ows"
  }
}
`)
}

func testAccGroupPolicyInvalidEnforcementState() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name = "policy_test"
  domain = "sys.auth"
  assertion {
    effect="DENY"
    action="TCP-IN:1024-65535:4443-4443"
    role="role_test"
    resource="sys.auth:ows"
	case_sensitive=true
    condition {
      instances {
        value = "*"
      }
      enforcementstate {
        value = "no_valid"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
	}
  }
}
`)
}

func testAccGroupPolicyDifferentModesWithSameEnforcementState() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name   = "policy_test"
  domain = "sys.auth"
  assertion {
    effect         = "DENY"
    action         = "TCP-IN:1024-65535:4443-4443"
    role           = "role_test"
    resource       = "sys.auth:ows"
    case_sensitive = true
    condition {
      instances {
        value = "yahoo.host1,yahoo.host2"
      }
      enforcementstate {
        value = "report"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
    condition {
      instances {
        value = "yahoo.host3,yahoo.host4"
      }
      enforcementstate {
        value = "report"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
  }
}
`)
}

func testAccGroupPolicySharedHostsBetweenModes1() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name   = "policy_test"
  domain = "sys.auth"
  assertion {
    effect         = "DENY"
    action         = "TCP-IN:1024-65535:4443-4443"
    role           = "role_test"
    resource       = "sys.auth:ows"
    case_sensitive = true
    condition {
      instances {
        value = "yahoo.host1,yahoo.host2,yahoo.host3"
      }
      enforcementstate {
        value = "report"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
    condition {
      instances {
        value = "yahoo.host3,yahoo.host4"
      }
      enforcementstate {
        value = "enforce"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
  }
}
`)
}

func testAccGroupPolicySharedHostsBetweenModes2() string {
	return fmt.Sprintf(`
resource "athenz_policy" "invalid" {
  name   = "policy_test"
  domain = "sys.auth"
  assertion {
    effect         = "DENY"
    action         = "TCP-IN:1024-65535:4443-4443"
    role           = "role_test"
    resource       = "sys.auth:ows"
    case_sensitive = true
    condition {
      instances {
        value = "yahoo.host1,yahoo.host2"
      }
      enforcementstate {
        value = "report"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
    condition {
      instances {
        value = "*"
      }
      enforcementstate {
        value = "enforce"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
  }
}
`)
}

func testAccGroupPolicyWithAssertionConditions(resourceRole, name, domain, resourceRoleName string) string {
	return fmt.Sprintf(`
%s
resource "athenz_policy" "policyTest" {
  name   = "%s"
  domain = "%s"
  assertion {
    role           = "${athenz_role.%s.name}"
    resource       = "%s"
    action         = "TCP-IN:1024-65535:4443-4443"
    effect         = "ALLOW"
    case_sensitive = true
    condition {
      instances {
        value = "host1,host2"
      }
      enforcementstate {
        value = "report"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
    condition {
      instances {
        value = "host3,host4"
      }
      enforcementstate {
        value = "enforce"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
  }
}
`, resourceRole, name, domain, resourceRoleName, domain+RESOURCE_SEPARATOR+"service")
}
