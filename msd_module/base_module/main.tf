locals {
  prefix         = "acl."
  identifier_key = "identifier"
  scope_key      = "scope"
  policy_name    = "${local.prefix}${var.service}.${var.acl_category}"
  role_names = [
    for p in var.policies :
    "${local.prefix}${var.service}.${var.acl_category}-${p[local.identifier_key]}"
  ]
  role_members = [
    for p in var.policies :
    p["services"]
  ]
  assertions = [
    for p in var.policies :
    {
      role           = "${local.prefix}${var.service}.${var.acl_category}-${p[local.identifier_key]}"
      resource       = "${var.domain}:${var.service}"
      action         = "${p["protocol"]}-${var.acl_category == "inbound" ? "IN" : "OUT"}:${p["source_ports"]}:${p["destination_ports"]}"
      effect         = "ALLOW"
      case_sensitive = true
      condition = [
        for c in p["conditions"] : {
          instances = [{
            value = c["hosts"]
          }]
          scopeaws = [{
            value = contains(c[local.scope_key], "aws") && !contains(c[local.scope_key], "all") ? "true" : "false"
          }]
          scopeonprem = [{
            value = contains(c[local.scope_key], "onprem") && !contains(c[local.scope_key], "all") ? "true" : "false"
          }]
          scopeall = [{
            value = contains(c[local.scope_key], "all") ? "true" : "false"
          }]
          enforcementstate = [{
            value = c["enforcementstate"]
          }]
      }]
    }
  ]
}


resource "athenz_role" "this" {
  count  = length(local.role_names)
  domain = var.domain
  name   = local.role_names[count.index]

  dynamic "member" {
    for_each = local.role_members[count.index]
    content {
      name = member.value
    }
  }
}

resource "athenz_policy" "this" {
  count  = length(local.assertions) > 0 ? 1 : 0
  name   = local.policy_name
  domain = var.domain

  dynamic "assertion" {
    for_each = local.assertions
    content {
      role           = assertion.value.role
      resource       = assertion.value.resource
      action         = assertion.value.action
      effect         = assertion.value.effect
      case_sensitive = assertion.value.case_sensitive
      dynamic "condition" {
        for_each = assertion.value.condition
        content {
          dynamic "instances" {
            for_each = condition.value.instances
            content {
              value = trim(instances.value.value, " ")
            }
          }
          dynamic "scopeonprem" {
            for_each = condition.value.scopeonprem
            content {
              value = scopeonprem.value.value
            }
          }
          dynamic "scopeaws" {
            for_each = condition.value.scopeaws
            content {
              value = scopeaws.value.value
            }
          }
          dynamic "scopeall" {
            for_each = condition.value.scopeall
            content {
              value = scopeall.value.value
            }
          }
          dynamic "enforcementstate" {
            for_each = condition.value.enforcementstate
            content {
              value = enforcementstate.value.value
            }
          }
        }
      }
    }
  }

  // this is a workaround for make sure that the policy is created after the role
  depends_on = [athenz_role.this]

  // this is a workaround for make sure that the policy update (delete assertions) will perform before the resource role destroy
  lifecycle {
    create_before_destroy = true
  }
}
