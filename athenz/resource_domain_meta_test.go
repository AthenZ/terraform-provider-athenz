package athenz

import (
	"fmt"
	"github.com/AthenZ/athenz/clients/go/zms"
	"log"
	"os"
	"testing"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGroupDomainMetaBasic(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Print("TF_ACC must be set for acceptance tests")
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	domainName := os.Getenv("DOMAIN")
	resourceName := "athenz_domain_meta.test_domain_meta"
	t.Cleanup(func() {
		cleanAccTestDomainMeta(domainName)
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupDomainMetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDomainMetaConfigBasic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupDomainMetaExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "test domain"),
					resource.TestCheckResourceAttr(resourceName, "application_id", "app-id-01"),
					resource.TestCheckResourceAttr(resourceName, "token_expiry_mins", "60"),
					resource.TestCheckResourceAttr(resourceName, "role_cert_expiry_mins", "20"),
					resource.TestCheckResourceAttr(resourceName, "service_cert_expiry_mins", "10"),
					resource.TestCheckResourceAttr(resourceName, "member_purge_expiry_days", "25"),
					resource.TestCheckResourceAttr(resourceName, "business_service", "test-service"),
					resource.TestCheckResourceAttr(resourceName, "contacts.security-contact", "user.joe"),
					resource.TestCheckResourceAttr(resourceName, "contacts.pe-contact", "user.jack"),
					resource.TestCheckResourceAttr(resourceName, "tags.zms.DisableExpirationNotifications", "4"),
					resource.TestCheckResourceAttr(resourceName, "audit_ref", "test audit ref"),
				),
			},
		},
	})
}

func cleanAccTestDomainMeta(domainName string) {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)
	_, err := zmsClient.GetDomain(domainName)
	if err == nil {
		var zero int32
		zero = 0
		domainMeta := zms.DomainMeta{
			Description:           "",
			ApplicationId:         "",
			MemberExpiryDays:      &zero,
			TokenExpiryMins:       &zero,
			ServiceCertExpiryMins: &zero,
			RoleCertExpiryMins:    &zero,
			ServiceExpiryDays:     &zero,
			GroupExpiryDays:       &zero,
			MemberPurgeExpiryDays: &zero,
			UserAuthorityFilter:   "",
			BusinessService:       "",
			Tags:                  make(map[zms.TagKey]*zms.TagValueList),
			Contacts:              make(map[zms.SimpleName]string),
		}
		if err = zmsClient.PutDomainMeta(domainName, AUDIT_REF, &domainMeta); err != nil {
			log.Printf("unable to reset domain name for %s: %v\n", domainName, err)
		}
	}
}

func testAccCheckGroupDomainMetaExists(resource string) resource.TestCheckFunc {
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
		if domain.Description == "" {
			return fmt.Errorf("does not have description set")
		}
		return nil
	}
}

func testAccCheckGroupDomainMetaDestroy(s *terraform.State) error {
	zmsClient := testAccProvider.Meta().(client.ZmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "athenz_domain_meta" {
			continue
		}

		domain, _ := zmsClient.GetDomain(rs.Primary.ID)

		if domain != nil && domain.Description != "" {
			return fmt.Errorf("athenz meta still exists")
		}
	}

	return nil
}

func testAccGroupDomainMetaConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "athenz_domain_meta" "test_domain_meta" {
  domain = "%s"
  description = "test domain"
  application_id = "app-id-01"
  token_expiry_mins = 60
  role_cert_expiry_mins = 20
  service_cert_expiry_mins = 10
  member_purge_expiry_days = 25
  business_service = "test-service"
  contacts = {
    "security-contact" = "user.joe",
    "pe-contact" = "user.jack"
  }
  tags = {
    "zms.DisableExpirationNotifications" = "4"
  }
  audit_ref = "test audit ref"
}
`, name)
}
