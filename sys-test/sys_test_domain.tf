resource "athenz_top_level_domain" "test_domain" {
  name        = var.sys_test_domain
  admin_users = ["user.github-7654321"]
  ypm_id      = 1
}

resource "athenz_top_level_domain" "test_delegate_domain" {
  name        = var.sys_test_delegated_domain
  admin_users = ["user.github-7654321"]
  ypm_id      = 2
}
