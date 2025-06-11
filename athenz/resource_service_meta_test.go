package athenz

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccServiceMetaBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	domainName := os.Getenv("DOMAIN")
	serviceName := "test-service-meta"
	resourceName := "athenz_service_meta.test_service_meta"

	t.Cleanup(func() {
		cleanAccTestServiceMeta(domainName, serviceName)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceMetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceMetaConfigBasic(domainName, serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMetaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", domainName),
					resource.TestCheckResourceAttr(resourceName, "name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "provider_endpoint", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "test audit ref"),
				),
			},
		},
	})
}

func cleanAccTestServiceMeta(domainName, serviceName string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	_, err := zmsClient.GetServiceIdentity(domainName, serviceName)
	if err == nil {
		serviceMeta := zms.ServiceIdentitySystemMeta{
			ProviderEndpoint: "",
		}
		if e := zmsClient.PutServiceIdentitySystemMeta(domainName, serviceName, "providerendpoint", AUDIT_REF, &serviceMeta); e != nil {
			log.Printf("unable to reset service meta for %s:service.%s: %v\n", domainName, serviceName, e)
		}
	}
}

func testAccCheckServiceMetaExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Service Meta ID is set")
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
		if service.ProviderEndpoint == "" {
			return fmt.Errorf("service provider endpoint is not set")
		}
		return nil
	}
}

func testAccCheckServiceMetaDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_service_meta" {
			continue
		}
		dn, sn, err := splitServiceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		service, err := zmsClient.GetServiceIdentity(dn, sn)
		if err != nil {
			var v rdl.ResourceError
			if errors.As(err, &v) && v.Code == 404 {
				return nil
			}
			return fmt.Errorf("unexpected error when checking service identity: %v", err)
		}
		if service.ProviderEndpoint != "" {
			return fmt.Errorf("athenz service meta still has provider_endpoint set, not destroyed properly")
		}
	}
	return nil
}

func testAccServiceMetaConfigBasic(domainName, serviceName string) string {
	return fmt.Sprintf(`
resource "athenz_service_meta" "test_service_meta" {
  domain            = "%s"
  name              = "%s"
  provider_endpoint = "https://example.com"
  audit_ref         = "test audit ref"
}
`, domainName, serviceName)
}
