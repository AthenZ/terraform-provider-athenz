package athenz

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGroupSubDomainBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	var domain zms.Domain
	if v := os.Getenv("PARENT_DOMAIN"); v == "" {
		t.Fatal("PARENT_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("SUB_DOMAIN"); v == "" {
		t.Fatal("SUB_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("ADMIN_USER"); v == "" {
		t.Fatal("ADMIN_USER must be set for acceptance tests")
	}
	parentDomain := os.Getenv("PARENT_DOMAIN")
	subDomainName := os.Getenv("SUB_DOMAIN")
	adminUser := os.Getenv("ADMIN_USER")
	resourceName := "athenz_sub_domain.testSubDomain"
	t.Cleanup(func() {
		cleanAccTestSubDomain(parentDomain, subDomainName)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupSubDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupSubDomainConfigBasic(subDomainName, parentDomain, adminUser),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupSubDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "name", subDomainName),
					resource.TestCheckResourceAttr(resourceName, "admin_users.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
		},
	})
}

func cleanAccTestSubDomain(parentName, domainName string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	fullName := parentName + SUB_DOMAIN_SEPARATOR + domainName
	_, err := zmsClient.GetDomain(fullName)
	if err == nil {
		if err = zmsClient.DeleteSubDomain(parentName, domainName, AUDIT_REF); err != nil {
			log.Fatalf("fail to delete Sub Domain %s. error: %s", fullName, err.Error())
		}
	}
}

func testAccCheckGroupSubDomainExists(resource string, d *zms.Domain) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz SubDomain ID is set")
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

func testAccCheckGroupSubDomainDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_sub_domain" {
			continue
		}

		_, err := zmsClient.GetDomain(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("athenz Sub Domain still exists")
		}
	}

	return nil
}

func testAccGroupSubDomainConfigBasic(name, parentName, adminUser string) string {
	return fmt.Sprintf(`
resource "athenz_sub_domain" "testSubDomain" {
  name = "%s"
  parent_name = "%s"
  admin_users = ["%s"]
}
`, name, parentName, adminUser)
}
