resource "athenz_group" "new_group" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foo"
  domain = var.sys_test_domain
  members = var.athenz_provider_foo-group_members
}