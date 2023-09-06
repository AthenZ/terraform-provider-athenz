output "athenz_roles" {
  value = length(athenz_role.this) > 0 ? athenz_role.this : null
}

output "athenz_policy" {
  value = length(athenz_policy.this) == 1 ? athenz_policy.this[0] : null
}