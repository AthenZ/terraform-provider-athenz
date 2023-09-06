variable "domain" {
  type = string
}

variable "service" {
  type = string
}

variable "acl_category" {
  type = string
}

variable "policies" {
  type = list(object({
    identifier        = string
    services          = list(string)
    protocol          = string
    source_ports      = string
    destination_ports = string
    conditions = list(object({
      scope            = list(string)
      enforcementstate = string
      hosts            = string
    }))
  }))
  default = []
  validation {
    condition     = length([for policy in var.policies : policy.identifier]) == length(distinct([for policy in var.policies : policy.identifier]))
    error_message = "Duplicate identifiers found in policies variable"
  }
}

# Check if there are duplicate identifiers


#validation {
#  condition     = alltrue([for policy in var.outbound_policies : can(policy.protocol == "TCP" || policy.protocol == "UDP")])
#  error_message = "Each policy's protocol must be 'TCP' or 'UDP'"
#}


#validation {
#  condition     = alltrue([for policy in var.policies : can(policy.protocol == "TCP" || policy.protocol == "UDP")])
#  error_message = "Each policy's protocol must be 'TCP' or 'UDP'"
#}
#
#validation {
#  condition     = alltrue([for policy in var.policies : can(policy.enforcementstate == "report" || policy.enforcementstate == "enforce")])
#  error_message = "Each policy's enforcementstate must be 'report' or 'enforce'"
#}
#
#validation {
#  condition     = alltrue([for policy in var.policies : alltrue([for s in policy.scope : can(s == "onprem" || s == "aws" || s == "all")])])
#  error_message = "Each policy's scope must be 'onprem', 'aws', or 'gcp'"
#}
#}