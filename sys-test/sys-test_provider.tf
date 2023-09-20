terraform {
  required_providers {
    athenz = {
      source  = "yahoo/provider/athenz"
      version = "x.x.x"
    }
  }
}

provider "athenz" {
  zms_url = var.zms_url
  cacert  = var.cacert
  cert    = var.cert
  key     = var.key
}
