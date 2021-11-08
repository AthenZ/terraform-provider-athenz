terraform {
  required_providers {
    athenz = {
      source = "AthenZ/athenz"
      version = ">= 0.0.3"
    }
  }
}

provider "athenz" {
  zms_url = var.zms_url
  cacert = var.cacert
  cert = var.cert
  key = var.key
}
