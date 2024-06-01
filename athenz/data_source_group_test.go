package athenz

import (
	"fmt"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"os"
	"testing"
	"time"
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
	now := rdl.TimestampNow()
	lastReviewedDate := timestampToString(&now)
	monthExpiry := rdl.Timestamp{Time: time.Now().UTC().Add(time.Duration(720) * time.Hour)}
	memberExpiry := timestampToString(&monthExpiry)
	t.Cleanup(func() {
		cleanAllAccTestGroups(domainName, []string{groupName})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig(groupName, domainName, member1, member2, memberExpiry, lastReviewedDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.#", dataSourceName, "member.#"),
					resource.TestCheckResourceAttrPair(resourceName, "member.0.name", dataSourceName, "member.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.0.expiration", dataSourceName, "member.0.expiration"),
					resource.TestCheckResourceAttrPair(resourceName, "member.1.name", dataSourceName, "member.1.name"),
					resource.TestCheckResourceAttrPair(resourceName, "member.1.expiration", dataSourceName, "member.1.expiration"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "settings.#", dataSourceName, "settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "settings.0.user_expiry_days", dataSourceName, "settings.0.user_expiry_days"),
					resource.TestCheckResourceAttrPair(resourceName, "settings.0.service_expiry_days", dataSourceName, "settings.0.service_expiry_days"),
					resource.TestCheckResourceAttrPair(resourceName, "settings.0.max_members", dataSourceName, "settings.0.max_members"),
					resource.TestCheckResourceAttrPair(resourceName, "last_reviewed_date", dataSourceName, "last_reviewed_date"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_domain_filter", dataSourceName, "principal_domain_filter"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig(name, domain, member1, member2, memberExpiry, lastReviewedDate string) string {
	return fmt.Sprintf(`
resource "athenz_group" "groupTest" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
	expiration = "%s"
  }
  member {
	name = "%s"
	expiration = "%s"
  }
  audit_ref = "done by someone"
  tags = {
	key1 = "v1,v2"
	key2 = "v2,v3"
  }
  settings {
	user_expiry_days = 60
	service_expiry_days = 90
	max_members = 5
  }
  last_reviewed_date = "%s"
  principal_domain_filter = "user,%s"
}

data "athenz_group" "groupTest" {
  domain = athenz_group.groupTest.domain
  name = athenz_group.groupTest.name
}
`, name, domain, member1, memberExpiry, member2, memberExpiry, lastReviewedDate, domain)
}
