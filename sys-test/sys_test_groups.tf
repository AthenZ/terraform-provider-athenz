resource "athenz_group" "new_group" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foo"
  domain = var.sys_test_domain
  member {
    name = var.athenz_provider_foo-group_member-0-name
    expiration = var.athenz_provider_foo-group_member-0-expiration
  }
}

resource "athenz_group" "new_group_deprecated" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foo_deprecated"
  domain = var.sys_test_domain
  members = var.athenz_provider_foo-group_members
}