package athenz

import (
	"fmt"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"os"
	"strings"
	"testing"
)

func TestAccGroupRolesDataSource(t *testing.T) {
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
	resourceName1 := "athenz_role.roleTest1"
	resourceName2 := "athenz_role.roleTest2"
	dataSourceName := "data.athenz_roles.rolesTest"
	rInt := acctest.RandInt()
	domainName := os.Getenv("DOMAIN")
	role1Name := fmt.Sprintf("test1_%d", rInt)
	role2Name := fmt.Sprintf("test2_%d", rInt)
	member1 := os.Getenv("MEMBER_1")
	member2 := os.Getenv("MEMBER_2")
	t.Cleanup(func() {
		cleanAllAccTestRoles(domainName, []string{role1Name, role2Name})
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGroupRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRolesDataSourceConfig(role1Name, role2Name, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName1, &role),
					testAccCheckGroupRoleExists(resourceName2, &role),
					testAccCheckRolesNames(dataSourceName, domainName, []string{role1Name, role2Name, "admin"}),
				),
			},
			{
				Config: testAccGroupRolesDataSourceConfigFilterByKey(role1Name, role2Name, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName1, &role),
					testAccCheckGroupRoleExists(resourceName2, &role),
					resource.TestCheckResourceAttr(dataSourceName, "roles.#", "2"),
					testAccCheckRolesNames(dataSourceName, domainName, []string{role1Name, role2Name}),
				),
			},
			{
				Config: testAccGroupRolesDataSourceConfigFilterByKeyAndVal(role1Name, role2Name, domainName, member1, member2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupRoleExists(resourceName1, &role),
					testAccCheckGroupRoleExists(resourceName2, &role),
					resource.TestCheckResourceAttr(dataSourceName, "roles.#", "1"),
					testAccCheckRolesNames(dataSourceName, domainName, []string{role2Name}),
				),
			},
		},
	})
}

func testAccCheckRolesNames(n string, domainName string, lookingForRoles []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Athenz Group Role ID is set")
		}
		lookingForRolesMap := make(map[string]bool)

		// for build the expected members, we look for all attribute from the following pattern: member.<index>.<attribute> (e.g. member.0.expiration)
		for key, val := range rs.Primary.Attributes {
			if !strings.HasPrefix(key, "roles.") {
				continue
			}
			theKeyArr := strings.Split(key, ".")
			if len(theKeyArr) == 3 && theKeyArr[2] == "name" {
				lookingForRolesMap[val] = true
			}
		}

		for _, lookingForRole := range lookingForRoles {
			_, found := lookingForRolesMap[domainName+ROLE_SEPARATOR+lookingForRole]
			if !found {
				return fmt.Errorf("looking for role %s in domain %s but not found", lookingForRole, domainName)
			}
		}

		return nil
	}
}

func testAccGroupRolesDataSourceConfig(name1, name2, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest1" {
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
  tags = {
	acc_test_tf_k1 = "acc_test_tf_v1,acc_test_tf_v2"
	acc_test_tf_k2 = "acc_test_tf_v2,acc_test_tf_v3"
	}
}
resource "athenz_role" "roleTest2" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  tags = {
	acc_test_tf_k2 = "acc_test_tf_v3,acc_test_tf_v4"
	acc_test_tf_k3 = "acc_test_tf_v1,acc_test_tf_v2"
  } 
}

data "athenz_roles" "rolesTest" {
	domain = athenz_role.roleTest1.domain
}
`, name1, domain, member1, member2, name2, domain, member1)
}

func testAccGroupRolesDataSourceConfigFilterByKey(name1, name2, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest1" {
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
  tags = {
	acc_test_tf_k1 = "acc_test_tf_v1,acc_test_tf_v2"
	acc_test_tf_k2 = "acc_test_tf_v2,acc_test_tf_v3"
  }
}
resource "athenz_role" "roleTest2" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  audit_ref="done by someone"
  tags = {
	acc_test_tf_k2 = "acc_test_tf_v3,acc_test_tf_v4"
	acc_test_tf_k3 = "acc_test_tf_v1,acc_test_tf_v2"
  }
  settings {
	token_expiry_mins = 5
	cert_expiry_mins = 10
  }  
}

data "athenz_roles" "rolesTest" {
	domain = athenz_role.roleTest1.domain
	tag_key = "acc_test_tf_k2"	
}
`, name1, domain, member1, member2, name2, domain, member1)
}

func testAccGroupRolesDataSourceConfigFilterByKeyAndVal(name1, name2, domain, member1, member2 string) string {
	return fmt.Sprintf(`
resource "athenz_role" "roleTest1" {
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
	acc_test_tf_k1 = "acc_test_tf_v1,acc_test_tf_v2"
	acc_test_tf_k2 = "acc_test_tf_v2,acc_test_tf_v3"
  } 
}
resource "athenz_role" "roleTest2" {
  name = "%s"
  domain = "%s"
  member {
	name = "%s"
  }
  audit_ref="done by someone"
  tags = {
	acc_test_tf_k2 = "acc_test_tf_v3,acc_test_tf_v4"
	acc_test_tf_k3 = "acc_test_tf_v1,acc_test_tf_v2"
  }
}

data "athenz_roles" "rolesTest" {
	domain = athenz_role.roleTest1.domain
	tag_key = "acc_test_tf_k2"	
	tag_value = "acc_test_tf_v4"	
}
`, name1, domain, member1, member2, name2, domain, member1)
}
