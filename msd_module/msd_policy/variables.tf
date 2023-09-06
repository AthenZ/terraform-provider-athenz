variable "domain" {
  type = string
}

variable "service" {
  type = string
}


variable "inbound_policies" {
  type = list(object({
    identifier        = string
    source_services   = list(string)
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
}

variable "outbound_policies" {
  type = list(object({
    identifier           = string
    destination_services = list(string)
    protocol             = string
    source_ports         = string
    destination_ports    = string
    conditions = list(object({
      scope            = list(string)
      enforcementstate = string
      hosts            = string
    }))
  }))
  default = []
}

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