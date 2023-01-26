---
page_title: "Top Level Domain resource - terraform-provider-athenz"
subcategory: ""
description: |-
The top-level-domain resource provides an Athenz top-level domain resource.
---

# Resource: athenz_top_level_domain

`athenz_top_level_domain` provides an Athenz top-level domain resource.

**Important Note: Use this resource only for create new top-level domain, update not supported. For import existing one, pls use terraform import.**

## Example Usage

```hcl
resource "athenz_top_level_domain" "athenz_top_level_domain-test" {
  name = "test"
  admin_users = ["user.someone"]
  ypm_id = "some_positive_integer"
  audit_ref = "create domain"
}
```

## Argument Reference

The following arguments are supported:


- `name` - (Required) name of the domain.


- `admin_users` - (Required) list of domain administrators. must be in this format: `user.<userid> or <domain>.<service>`.


- `ypm_id` - (Required) associated product id. must be a positive integer.


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


## Import
Top-Level-Domain resource can be imported using the Top-Level-Domain name: `<domain name>`, e.g.

```hcl
#1. Define empty resource in your <somefile>.tf

    resource "athenz_top_level_domain" "import_top_level_domain" {
    }

#2. In the directory where the file is located, run this command:
        
   ֿ$ terraform import athenz_top_level_domain.import_top_level_domain <domain name>

#3. Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html