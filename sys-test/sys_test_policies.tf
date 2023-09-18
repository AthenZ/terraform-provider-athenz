resource "athenz_policy" "add-policy-test" {
  domain = var.sys_test_domain
  name   = "athenz_provider_foo"
  assertion {
    effect   = var.athenz_provider_foo-assertion-0-effect
    action   = var.athenz_provider_foo-assertion-0-action
    role     = var.athenz_provider_foo-assertion-0-role
    resource = "${var.sys_test_domain}:${var.athenz_provider_foo-assertion-0-resource}"
  }
  assertion {
    effect         = var.athenz_provider_foo-assertion-1-effect
    action         = var.athenz_provider_foo-assertion-1-action
    role           = var.athenz_provider_foo-assertion-1-role
    resource       = "${var.sys_test_domain}:${var.athenz_provider_foo-assertion-1-resource}"
    case_sensitive = true
  }
  assertion {
    effect         = var.athenz_provider_foo-assertion-2-effect
    action         = var.athenz_provider_foo-assertion-2-action
    role           = var.athenz_provider_foo-assertion-2-role
    resource       = "${var.sys_test_domain}:${var.athenz_provider_foo-assertion-2-resource}"
    case_sensitive = true
    condition {
      instances {
        value = "yahoo.host1,yahoo.host2"
      }
      enforcementstate {
        value = "enforce"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
    condition {
      instances {
        value = "yahoo.host3,yahoo.host4"
      }
      enforcementstate {
        value = "report"
      }
      scopeaws {
        value = "true"
      }
      scopeonprem {
        value = "false"
      }
      scopeall {
        value = "false"
      }
    }
  }
  // IMPORTANT: The roles "athenz_provider_foo" and "athenz_provider_bar" must be defined since they were used in the assertions.
  depends_on = [athenz_top_level_domain.test_domain, athenz_role.with_tags, athenz_role.without_tags]
}