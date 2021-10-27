package athenz

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccGroupTopLevelDomainBasic(t *testing.T) {
	t.Skip("Skipping domains creation in acc tests")
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	var domain zms.Domain
	if v := os.Getenv("TOP_LEVEL_DOMAIN"); v == "" {
		t.Fatal("TOP_LEVEL_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("ADMIN_USER"); v == "" {
		t.Fatal("ADMIN_USER must be set for acceptance tests")
	}
	rInt := acctest.RandInt()
	ypmId := rInt % 1000000
	topLevelDomainName := os.Getenv("TOP_LEVEL_DOMAIN")
	adminUser := os.Getenv("ADMIN_USER")
	resourceName := "athenz_top_level_domain.test_domain"
	t.Cleanup(func() {
		cleanAccTestDomain(topLevelDomainName)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupTopLevelDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTopLevelDomainConfigBasic(topLevelDomainName, adminUser, ypmId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTopLevelDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "name", topLevelDomainName),
					resource.TestCheckResourceAttr(resourceName, "admin_users.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
		},
	})
}

func cleanAccTestDomain(domainName string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	_, err := zmsClient.GetDomain(domainName)
	if err == nil {
		if err = zmsClient.DeleteTopLevelDomain(domainName, AUDIT_REF); err != nil {
			log.Fatalf("fail to delete Top Level Domain %s. error: %s", domainName, err.Error())
		}
	}
}

func testAccCheckGroupTopLevelDomainExists(resource string, d *zms.Domain) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Top Level Domain ID is set")
		}

		zmsClient := testAccProvider.Meta().(client.ZmsClient)
		domain, err := zmsClient.GetDomain(rs.Primary.ID)

		if err != nil {
			return err
		}
		*d = *domain
		return nil
	}
}

func testAccCheckGroupTopLevelDomainDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_top_level_domain" {
			continue
		}

		_, err := zmsClient.GetDomain(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("athenz Top Level Domain still exists")
		}
	}

	return nil
}

func testAccGroupTopLevelDomainConfigBasic(name, adminUser string, ypmId int) string {
	return fmt.Sprintf(`
resource "athenz_top_level_domain" "test_domain" {
  name = "%s"
  admin_users = ["%s"]
  ypm_id = %d
}
`, name, adminUser, ypmId)
}
