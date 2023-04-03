resource "athenz_role" "with_tags" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foo"
  domain = var.sys_test_domain
  member {
    name = var.athenz_provider_foo-member-0-name
  }
  audit_ref = var.athenz_provider_foo-audit_ref
  tags = {
    (var.athenz_provider_foo-tags-0-key) = var.athenz_provider_foo-tags-0-values
  }
}

resource "athenz_role" "without_tags" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_bar"
  domain = var.sys_test_domain
  member {
    name = var.athenz_provider_bar-member-0-name
    expiration = var.athenz_provider_bar-member-0-expiration
  }
}

resource "athenz_role" "with_tags_deprecated" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foo_deprecated"
  domain = var.sys_test_domain
  members = var.athenz_provider_foo-members
  audit_ref = var.athenz_provider_foo-audit_ref
  tags = {
    (var.athenz_provider_foo-tags-0-key) = var.athenz_provider_foo-tags-0-values
  }
}

resource "athenz_role" "without_tags_deprecated" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_bar_deprecated"
  domain = var.sys_test_domain
  members = var.athenz_provider_bar-members
}

resource "athenz_role" "with_all_settings" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foobar"
  domain = var.sys_test_domain
  member {
    name = var.athenz_provider_bar-member-0-name
    expiration = var.athenz_provider_bar-member-0-expiration
    review = var.athenz_provider_bar-member-0-review
  }
  settings {
    token_expiry_mins = 7
    cert_expiry_mins = 8
    user_expiry_days = 14
    user_review_days = 15
    group_expiry_days = 21
    group_review_days = 22
    service_expiry_days = 30
    service_review_days = 31
  }
}
