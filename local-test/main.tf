terraform {
  required_version = ">= 1.1"
  required_providers {
    athenz = {
      source = "yahoo/provider/athenz"
      version = ">= 1.0.8"
    }
  }
}

provider "athenz" {
  zms_url = "https://zms.athenz.ouroath.com:4443/zms/v1"
  cert    = fileexists("/tokens/cert") ? "/tokens/cert" : pathexpand("~/.athenz/cert")
  key     = fileexists("/tokens/key") ? "/tokens/key" : pathexpand("~/.athenz/key")
}

resource "athenz_role" "foo" {
  domain  = "<DEV_ATHENZ_DOMAIN>"
  name    = "foo"
}