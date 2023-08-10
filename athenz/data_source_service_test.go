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

func TestAccGroupServiceDataSource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	var service zms.ServiceIdentity
	resourceName := "athenz_service.serviceTest"
	dataSourceName := "data.athenz_service.serviceTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	serviceName := fmt.Sprintf("test%d", rInt)
	t.Cleanup(func() {
		cleanAllAccTestServices(domainName, []string{serviceName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupServiceDataSourceConfig(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupServiceExists(resourceName, &service),
					resource.TestCheckResourceAttrPair(resourceName, "domain", dataSourceName, "domain"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttr(dataSourceName, "public_keys.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "public_keys.0.key_id", dataSourceName, "public_keys.0.key_id"),
					resource.TestCheckResourceAttrPair(resourceName, "public_keys.0.key_value", dataSourceName, "public_keys.0.key_value"),
				),
			},
		},
	})
}

func testAccGroupServiceDataSourceConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "athenz_service" "serviceTest" {
  name = "%s"
  domain = "%s"
  description = "test service"
  audit_ref = "done by someone"
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

data "athenz_service" "serviceTest" {
  domain = athenz_service.serviceTest.domain
  name = athenz_service.serviceTest.name
}
`, name, domain)
}
