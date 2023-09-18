# ------------------ global variables --------------------
variable "zms_url" {
  type        = string
  description = "The zms to use"
}
variable "sys_test_domain" {
  type        = string
  description = "The domain name to use"
}
variable "sys_test_delegated_domain" {
  type        = string
  description = "The domain name to use for delegating roles"
}
variable "cacert" {
  type        = string
  description = "cacert for systest"
}
variable "cert" {
  type        = string
  description = "cert for systest"
}
variable "key" {
  type        = string
  description = "key for systest"
}
# ------------------ roles variables --------------------
variable "athenz_provider_foo-members" {
  type        = list(string)
  description = "The role members to use"
}
variable "athenz_provider_foo-member-0-name" {
  type        = string
  description = "The name to use"
}
variable "athenz_provider_foo-audit_ref" {
  type        = string
  description = "The auditRef to use"
}
variable "athenz_provider_foo-tags-0-key" {
  type        = string
  description = "The tag key to use"
}
variable "athenz_provider_foo-tags-0-values" {
  type        = string
  description = "The tag values to use"
}
variable "athenz_provider_bar-members" {
  type        = list(string)
  description = "The role members to use"
}
variable "athenz_provider_bar-member-0-name" {
  type        = string
  description = "The name to use"
}
variable "athenz_provider_bar-member-0-expiration" {
  type        = string
  description = "The expiration to use"
}

variable "athenz_provider_bar-member-0-review" {
  type        = string
  description = "The review to use"
}
# ------------------ groups variables --------------------
variable "athenz_provider_foo-group_members" {
  type        = list(string)
  description = "The group members to use"
}
variable "athenz_provider_foo-group_member-0-name" {
  type        = string
  description = "The name to use"
}
variable "athenz_provider_foo-group_member-0-expiration" {
  type        = string
  description = "The expiration to use"
}
# ------------------ services variables --------------------
variable "athenz_provider_foo-keys-0-id" {
  type        = string
  description = "The public key id to use"
}
variable "athenz_provider_foo-keys-0-value" {
  type        = string
  description = "The public key value to use"
}
variable "athenz_provider_foo-service_audit_ref" {
  type        = string
  description = "The auditRef to use"
}

# ------------------ policies variables --------------------
variable "athenz_provider_foo-assertion-0-effect" {
  type        = string
  description = "The effect to use (ALLOW/DENY)"
}
variable "athenz_provider_foo-assertion-0-action" {
  type        = string
  description = "The action to ALLOW/DENY"
}
variable "athenz_provider_foo-assertion-0-role" {
  type        = string
  description = "The role name to use"
}
variable "athenz_provider_foo-assertion-0-resource" {
  type        = string
  description = "The resource name to use"
}
variable "athenz_provider_foo-assertion-1-effect" {
  type        = string
  description = "The effect to use (ALLOW/DENY)"
}
variable "athenz_provider_foo-assertion-1-action" {
  type        = string
  description = "The action to ALLOW/DENY"
}
variable "athenz_provider_foo-assertion-1-role" {
  type        = string
  description = "The role name to use"
}
variable "athenz_provider_foo-assertion-1-resource" {
  type        = string
  description = "The resource name to use"
}

variable "athenz_provider_foo-assertion-2-effect" {
  type        = string
  description = "The effect to use (ALLOW/DENY)"
}
variable "athenz_provider_foo-assertion-2-action" {
  type        = string
  description = "The action to ALLOW/DENY"
}
variable "athenz_provider_foo-assertion-2-role" {
  type        = string
  description = "The role name to use"
}
variable "athenz_provider_foo-assertion-2-resource" {
  type        = string
  description = "The resource name to use"
}

# ------------------ policies versions variables --------------------
variable "athenz_provider_with_versions-active_version" {
  type        = string
  description = "The active version to use"
}
variable "athenz_provider_with_versions-assertion-version1-0-effect" {
  type        = string
  description = "The effect to use (ALLOW/DENY)"
}
variable "athenz_provider_with_versions-assertion-version1-0-action" {
  type        = string
  description = "The action to ALLOW/DENY"
}
variable "athenz_provider_with_versions-assertion-version1-0-role" {
  type        = string
  description = "The role name to use"
}
variable "athenz_provider_with_versions-assertion-version1-0-resource" {
  type        = string
  description = "The resource name to use"
}
variable "athenz_provider_with_versions-assertion-version2-0-effect" {
  type        = string
  description = "The effect to use (ALLOW/DENY)"
}
variable "athenz_provider_with_versions-assertion-version2-0-action" {
  type        = string
  description = "The action to ALLOW/DENY"
}
variable "athenz_provider_with_versions-assertion-version2-0-role" {
  type        = string
  description = "The role name to use"
}
variable "athenz_provider_with_versions-assertion-version2-0-resource" {
  type        = string
  description = "The resource name to use"
}
variable "athenz_provider_with_versions-assertion-version2-1-effect" {
  type        = string
  description = "The effect to use (ALLOW/DENY)"
}
variable "athenz_provider_with_versions-assertion-version2-1-action" {
  type        = string
  description = "The action to ALLOW/DENY"
}
variable "athenz_provider_with_versions-assertion-version2-1-role" {
  type        = string
  description = "The role name to use"
}
variable "athenz_provider_with_versions-assertion-version2-1-resource" {
  type        = string
  description = "The resource name to use"
}