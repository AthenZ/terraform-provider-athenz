terraform {
  required_providers {
    athenz = {
      source = "yahoo/provider/athenz"
      version = "9.9.9"
    }
  }
}


#terraform {
#  required_providers {
#    athenz = {
#      source = "AthenZ/athenz"
#      version = "1.0.21"
#    }
#  }
#}


provider "athenz" {
  zms_url = "https://stage.zms.athenz.ouroath.com:4443/zms/v1"
  cert = "/Users/nsegal/.athenz/cert" # change to your location
  key = "/Users/nsegal/.athenz/key" # change to your location
}