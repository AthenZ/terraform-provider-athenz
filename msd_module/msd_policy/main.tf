module "inbound" {
  source       = "../base_module"
  acl_category = "inbound"
  domain       = var.domain
  service      = var.service
  policies = [for p in var.inbound_policies :
    {
      identifier        = p["identifier"]
      services          = p["source_services"]
      protocol          = p["protocol"]
      source_ports      = p["source_ports"]
      destination_ports = p["destination_ports"]
      conditions        = p["conditions"]
    }
  ]
}

module "outbound" {
  source       = "../base_module"
  acl_category = "outbound"
  domain       = var.domain
  service      = var.service
  policies = [for p in var.outbound_policies :
    {
      identifier        = p["identifier"]
      services          = p["destination_services"]
      protocol          = p["protocol"]
      source_ports      = p["source_ports"]
      destination_ports = p["destination_ports"]
      conditions        = p["conditions"]
    }
  ]
}