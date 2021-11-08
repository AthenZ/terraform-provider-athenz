---
page_title: "Service resource - terraform-provider-athenz"
subcategory: ""
description: |-
The role resource provides an Athenz service resource.
---

# Resource `athenz_service`

`athenz_service` provides an Athenz service resource.

### Example Usage

```hcl
resource "athenz_service" "foo_service" {
  name = "foo"
  domain = "some_domain"
  audit_ref = "create service"
  public_keys = [{
    key_id = "v0"
    key_value = <<EOK
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzZCUhLc3TpvObhjdY8Hb
/0zkfWAYSXLXaC9O1S8AXoM7/L70XY+9KL+1Iy7xYDTrbZB0tcolLwnnWHq5giZm
Uw3u6FGSl5ld4xpyqB02iK+cFSqS7KOLLH0p9gXRfxXiaqRiV2rKF0ThzrGox2cm
Df/QoZllNdwIFGqkuRcEDvBnRTLWlEVV+1U12fyEsA1yvVb4F9RscZDYmiPRbhA+
cLzqHKxX51dl6ek1x7AvUIM8js6WPIEfelyTRiUzXwOgIZbqvRHSPmFG0ZgZDjG3
Llfy/E8K0QtCk3ki1y8Tga2I5k2hffx3DrHMnr14Zj3Br0T9RwiqJD7FoyTiD/ti
xQIDAQAB
-----END PUBLIC KEY-----
EOK
  }]
}
```

### Argument Reference

The following arguments are supported:

- `name` - (Required) The service name.


- `domain` - (Required) The Athenz domain name.


- `public_keys` - (Optional) Set of maps of public keys. Each map consists the following arguments:  

    - `key_id` - (Required) The key id.
      
    - `key_value` - (Required) The Key Value which must be a PEM encoded public key.


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


### Import
Service resource can be imported using the service id: `<domain>.<service name>`, e.g.

```hcl
#1. Define empty resource in your <somefile>.tf

    resource "athenz_service" "import_service" {
    }

#2. In the directory where the file is located, run this command:
        
   ֿ$ terraform import athenz_service.import_service <domain>.<service name> 

#3. Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html