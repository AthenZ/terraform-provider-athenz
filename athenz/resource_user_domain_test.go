package athenz

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccGroupUserDomainBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("SKIP_USER_DOMAIN_TEST"); v == "true" {
		return
	}
	var domain zms.Domain
	if v := os.Getenv("SHORT_ID"); v == "" {
		t.Fatal("SHORT_ID must be set for acceptance tests")
	}
	shortId := os.Getenv("SHORT_ID")
	resourceName := "athenz_user_domain.testUserDomain"
	t.Cleanup(func() {
		cleanAccTestUserDomain(shortId)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupUserDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupUserDomainConfigBasic(shortId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupUserDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "name", shortId),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", AUDIT_REF),
				),
			},
		},
	})
}

func cleanAccTestUserDomain(shortId string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	_, err := zmsClient.GetDomain(PREFIX_USER_DOMAIN + shortId)
	if err == nil {
		if err = zmsClient.DeleteUserDomain(shortId, AUDIT_REF); err != nil {
			log.Fatalf("fail to delete User Domain %s. error: %s", shortId, err.Error())
		}
	}
}

func testAccCheckGroupUserDomainExists(resource string, d *zms.Domain) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz User Domain ID is set")
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

func testAccCheckGroupUserDomainDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_user_domain" {
			continue
		}

		_, err := zmsClient.GetDomain(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("athenz User Domain still exists")
		}
	}

	return nil
}

func testAccGroupUserDomainConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "athenz_user_domain" "testUserDomain" {
  name = "%s"
}
`, name)
}
