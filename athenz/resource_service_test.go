package athenz

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGroupServiceBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	var service zms.ServiceIdentity
	rInt := acctest.RandInt()
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	resourceName := "athenz_service.serviceTest"
	domain := os.Getenv("DOMAIN")
	serviceName := fmt.Sprintf("test%d", rInt)
	t.Cleanup(func() {
		cleanAllAccTestServices(domain, []string{serviceName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupServiceConfigBasic(serviceName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccGroupServiceConfigAddPublicKey(serviceName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccGroupServiceConfigRemovePublicKey(serviceName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccGroupServiceConfigChangeDescription(serviceName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "description", "this service is for acc test"),
				),
			},
			{
				Config: testAccServiceConfigAddTags(serviceName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}, "key2": {"b1", "b2"}}),
				),
			},
			{
				Config: testAccServiceConfigRemoveTags(serviceName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					testAccCheckCorrectTags(resourceName, map[string][]string{"key1": {"a1", "a2"}}),
				),
			},
		},
	})
}

func TestAccGroupServiceInvalidResource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupServiceInvalidDomainNameConfig(),
				ExpectError: getPatternErrorRegex(DOMAIN_NAME),
			},
			{
				Config:      testAccGroupServiceInvalidServiceNameConfig(),
				ExpectError: getPatternErrorRegex(SIMPLE_NAME),
			},
		},
	})
}

func cleanAllAccTestServices(domain string, services []string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	for _, serviceName := range services {
		_, err := zmsClient.GetServiceIdentity(domain, serviceName)
		if err == nil {
			if err = zmsClient.DeleteServiceIdentity(domain, serviceName, AUDIT_REF); err != nil {
				log.Printf("error deleting Service %s: %s", serviceName, err)
			}
		}
	}
}

func testAccCheckGroupServiceExists(n string, s *zms.ServiceIdentity) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group ID is set")
		}

		dn, sn, err := splitServiceId(rs.Primary.ID)
		if err != nil {
			return err
		}

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		service, err := zmsClient.GetServiceIdentity(dn, sn)

		if err != nil {
			return err
		}

		*s = *service

		return nil
	}
}

func testAccCheckGroupServiceDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_service" {
			continue
		}

		dn, sn, err := splitServiceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = zmsClient.GetServiceIdentity(dn, sn)

		if err == nil {
			return fmt.Errorf("athenz Group still exists")
		}
	}

	return nil
}

func testAccGroupServiceConfigBasic(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
  name = "%s"
  domain = "%s"
}
`, name, domain)
}

func testAccGroupServiceConfigAddPublicKey(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
  name = "%s"
  domain = "%s"
  public_keys = [{
		key_id = "v0"
		key_value = <<EOK
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzZCUhLc3TpvObhjdY8Hb
/0zkfWAYSXLXaC9O1S8AXoM7/L70XY+9KL+1Iy7xYDTrbZB0tcolLwnnWHq5giZm
Uw3u6FGSl5ld4xpyqB02iK+cFSqS7KOLLH0p9gXRfxXiaqRiV2rKF0ThzrGox2cm
Df/QoZllNdwIFGqkuRcEDvBnRTLWlEVV+1U12fyEsA1yvVb4F9RscZDYmiPRbhA+
cLzqHKxX51dl6ek1x7AvUIM8js6WPIEfelyTRiUzXwOgIZbqvRHSPmFG0ZgZDjG3
Llfy/E8K0QtCk3ki1y8Tga2I5k2hffx3DrHMnr14Zj3Br0T9RwiqJD7FoyTiD/ti
xQIDAQAB
-----END PUBLIC KEY-----
EOK
	}]
}
`, name, domain)
}

func testAccGroupServiceConfigRemovePublicKey(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
  name = "%s"
  domain = "%s"
}
`, name, domain)
}

func testAccGroupServiceConfigChangeDescription(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
  name = "%s"
  domain = "%s"
  description = "this service is for acc test"
}
`, name, domain)
}

func testAccGroupServiceInvalidDomainNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
	domain = "sys.au@th"
	name = "service_test"
}
`)
}

func testAccGroupServiceInvalidServiceNameConfig() string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
	domain = "sys.auth"
	name = "service.test"
}
`)
}

func testAccServiceConfigAddTags(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
  name = "%s"
  domain = "%s"
tags = {
	key1 = "a1,a2"
	key2 = "b1,b2"
	}
}
`, name, domain)
}

func testAccServiceConfigRemoveTags(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
  name = "%s"
  domain = "%s"
tags = {
	key1 = "a1,a2"
  }
}
`, name, domain)
}
