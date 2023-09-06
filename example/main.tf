terraform {
  required_providers {
    athenz = {
      source = "yahoo/provider/athenz"
    }
  }
}

#terraform {
#  required_providers {
#    athenz = {
#      source = "AthenZ/athenz"
#      version = "1.0.31"
#    }
#  }
#}

locals {
  user_name = "mshneorson"
}

provider "athenz" {
  zms_url = "https://dev.zms.athens.yahoo.com:4443/zms/v1"
  #  zms_url = "https://stage.zms.athenz.ouroath.com:4443/zms/v1"
  cert = "/Users/${local.user_name}/.athenz/cert" # change to your location
  key  = "/Users/${local.user_name}/.athenz/key"  # change to your location
}

module "mendi_msd_policies" {
  source = "../msd_module/msd_policy"
  service = "mendi"
  domain = "home.mshneorson"
  inbound_policies = [
    {
      identifier = "tf-module-test"
      source_services = ["yamas.api"]
      destination_ports = "4443"
      source_ports = "1024-65535"
      protocol = "TCP"
      conditions = [
        {
          hosts            = "yahoo.host1"
          scope            = ["all"]
          enforcementstate = "enforce"
        },
        {
          hosts            = "yahoo.host12"
          scope            = ["all"]
          enforcementstate = "report"
        }
      ]
    },
    {
      identifier = "tf-module-test2"
      source_services = ["sys.calypso"]
      destination_ports = "8443"
      source_ports = "1024-65535"
      protocol = "TCP"
      conditions = [
        {
          enforcementstate = "report"
          hosts            = "*"
          scope            = ["aws"]
        }
      ]
    },
  ]
  outbound_policies = [
    {
      identifier = "tf-module-test"
      destination_services = ["athens.ci"]
      destination_ports = "4443"
      source_ports = "1024-65535"
      protocol = "TCP"
      conditions = [
        {
          enforcementstate = "report"
          hosts            = "*"
          scope            = ["aws"]
        }
      ]
    },
  ]
}

# `inbound policies`

output "mendi_inbound_roles" {
  description = "The `inbound` Athenz roles."
  value       = module.mendi_msd_policies.inbound_roles
}

output "mendi_inbound_policy" {
  description = "The `inbound` Athenz policy."
  value       = module.mendi_msd_policies.inbound_policy
}

# `outbound policies`

output "mendi_outbound_role" {
  description = "The `outbound` Athenz roles."
  value       = module.mendi_msd_policies.outbound_roles
}

output "mendi_outbound_policy" {
  description = "The `outbound` Athenz policy."
  value       = module.mendi_msd_policies.outbound_policy
}


module "dvir_msd_policies" {
  source = "../msd_module/msd_policy"
  service = "dvir"
  domain = "home.mshneorson"
  inbound_policies = [
    {
      identifier = "tf-module-test"
      source_services = ["yamas.api"]
      destination_ports = "4443"
      source_ports = "1024-65535"
      protocol = "TCP"
      conditions = [
        {
          hosts            = "yahoo.host1"
          scope            = ["all"]
          enforcementstate = "enforce"
        },
        {
          hosts            = "yahoo.host12"
          scope            = ["all"]
          enforcementstate = "report"
        }
      ]
    },
    {
      identifier = "tf-module-test2"
      source_services = ["sys.calypso"]
      destination_ports = "8443"
      source_ports = "1024-65535"
      protocol = "TCP"
      conditions = [
        {
          enforcementstate = "report"
          hosts            = "*"
          scope            = ["aws"]
        }
      ]
    },
  ]
  outbound_policies = [
    {
      identifier = "tf-module-test"
      destination_services = ["athens.ci"]
      destination_ports = "4443"
      source_ports = "1024-65535"
      protocol = "TCP"
      conditions = [
        {
          enforcementstate = "report"
          hosts            = "*"
          scope            = ["aws"]
        }
      ]
    },
  ]
}

#data "athenz_policy" "assertion_condition" {
#    name = "acl.mshneors.inbound"
#    domain = "home.mshneorson"
#}
#
#output "assertion_condition" {
#  value = data.athenz_policy.assertion_condition
#}

#resource "athenz_policy" "foo_policy" {
#  name   = "acl.test.inbound"
#  domain = "home.mshneorson"
#  assertion {
#    effect   = "ALLOW"
#    action   = "some_action"
#    role     = "test"
#    resource = "home.mshneorson:some_resource"
#  }
#  assertion {
#    role           = "acl.mshneors.inbound-test"
#    resource       = "home.mshneorson:mshneors"
#    action         = "TCP-IN:1024-65535:4443-4443"
#    effect         = "ALLOW"
#    case_sensitive = true
#    condition {
#      instances {
#        value = "*"
#      }
#      enforcementstate {
#        value = "report"
#      }
#      scopeaws {
#        value = "true"
#      }
#      scopeonprem {
#        value = "false"
#      }
#      scopeall {
#        value = "false"
#      }
#    }
#    condition {
#      instances {
#        value = "*"
#      }
#      enforcementstate {
#        value = "report"
#      }
#      scopeaws {
#        value = "true"
#      }
#      scopeonprem {
#        value = "false"
#      }
#      scopeall {
#        value = "false"
#      }
#    }
#  }
#  audit_ref = "create policy"
#}

#output "assertion_ids" {
#  value = [
#  for a in athenz_policy.foo_policy.assertion :
#  a.id
#  ]
#}
