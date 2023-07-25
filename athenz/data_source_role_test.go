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

func TestAccGroupRoleDataSource(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" && v != "true" {
		log.Printf("TF_ACC must be set for acceptance tests, value is: %s", v)
		return
	}
	if v := os.Getenv("DOMAIN"); v == "" {
		t.Fatal("DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("MEMBER_1"); v == "" {
		t.Fatal("MEMBER_1 must be set for acceptance tests")
	}
	if v := os.Getenv("MEMBER_2"); v == "" {
		t.Fatal("MEMBER_2 must be set for acceptance tests")
	}
	var role zms.Role
	resourceName := "athenz_role.roleTest"
	dataSourceName := "data.athenz_role.roleTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	roleName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	t.Cleanup(func() {
		cleanAllAccTestRoles(domainName, []string{roleName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRoleDataSourceConfig(roleName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName, &role),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.#", dataSourceName, "member.#"),
					resource.TestCheckResourceAttrPair(resourceName, "member.0.name", dataSourceName, "member.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.0.expiration", dataSourceName, "member.0.expiration"),
					resource.TestCheckResourceAttrPair(resourceName, "member.0.review", dataSourceName, "member.0.review"),
					resource.TestCheckResourceAttrPair(resourceName, "member.1.name", dataSourceName, "member.1.name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.1.expiration", dataSourceName, "member.1.expiration"),
					resource.TestCheckResourceAttrPair(resourceName, "member.1.review", dataSourceName, "member.1.review"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "settings.#", dataSourceName, "settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "settings.0.token_expiry_mins", dataSourceName, "settings.0.token_expiry_mins"),
					resource.TestCheckResourceAttrPair(resourceName, "settings.0.cert_expiry_mins", dataSourceName, "settings.0.cert_expiry_mins"),
				),
			},
		},
	})
}

func testAccGroupRoleDataSourceConfig(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
	review = "2022-12-29 23:59:59"
  }
  audit_ref="done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
	}
  settings {
	token_expiry_mins = 5
	cert_expiry_mins = 10
  }  
}

data "athenz_role" "roleTest" {
  domain = athenz_role.roleTest.domain
  name = athenz_role.roleTest.name
}
`, name, domain, member1, member2)
}
