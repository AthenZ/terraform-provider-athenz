# `inbound policies`

output "inbound_roles" {
  description = "The `inbound` Athenz roles."
  value       = module.inbound.athenz_roles
}

output "inbound_policy" {
  description = "The `inbound` Athenz policy."
  value       = module.inbound.athenz_policy
}

# `outbound policies`

output "outbound_roles" {
  description = "The `outbound` Athenz roles."
  value       = module.outbound.athenz_roles
}

output "outbound_policy" {
  description = "The `outbound` Athenz policy."
  value       = module.outbound.athenz_policy
}
