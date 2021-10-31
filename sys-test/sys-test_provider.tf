terraform {
  required_providers {
    athenz = {
      source = "dvirguttman/athenz"
      version = "0.0.3"
    }
  }
}

provider "athenz" {
  zms_url = var.zms_url
  cacert = "../docker/sample/CAs/athenz_ca.pem"
  cert = "../docker/sample/domain-admin/domain_admin_cert.pem"
  key = "../docker/sample/domain-admin/domain_admin_key.pem"
}
