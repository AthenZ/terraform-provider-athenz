resource "athenz_service" "new-service" {
  depends_on = [athenz_top_level_domain.test_domain]
  name = "athenz_provider_foo"
  domain = var.sys_test_domain
  audit_ref = var.athenz_provider_foo-service_audit_ref
  public_keys = [{
    key_id = var.athenz_provider_foo-keys-0-id
    key_value = var.athenz_provider_foo-keys-0-value
    }]
}