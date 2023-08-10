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

func TestAccGroupDataSource(t *testing.T) {
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
	var group zms.Group
	resourceName := "athenz_group.groupTest"
	dataSourceName := "data.athenz_group.groupTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	groupName := fmt.Sprintf("test%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	t.Cleanup(func() {
		cleanAllAccTestGroups(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig(groupName, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.#", dataSourceName, "member.#"),
					resource.TestCheckResourceAttrPair(resourceName, "member.0.name", dataSourceName, "member.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.0.expiration", dataSourceName, "member.0.expiration"),
					resource.TestCheckResourceAttrPair(resourceName, "member.1.name", dataSourceName, "member.1.name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.1.expiration", dataSourceName, "member.1.expiration"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig(name, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  member {
	name = "%s"
	expiration = "2022-12-29 23:59:59"
  }
  audit_ref="done by someone"
}

data "athenz_group" "groupTest" {
  domain = athenz_group.groupTest.domain
  name = athenz_group.groupTest.name
}
`, name, domain, member1, member2)
}
