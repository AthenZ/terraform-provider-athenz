resource "athenz_role" "with_tags" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foo"
  domain = var.sys_test_domain
  members = var.athenz_provider_foo-members
  audit_ref = var.athenz_provider_foo-audit_ref
  tags = {
    (var.athenz_provider_foo-tags-0-key) = var.athenz_provider_foo-tags-0-values
  }
}

resource "athenz_role" "without_tags" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_bar"
  domain = var.sys_test_domain
  members = var.athenz_provider_bar-members
}
