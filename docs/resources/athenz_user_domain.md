---
page_title: "User resource - terraform-provider-athenz"
subcategory: ""
description: |-
The user-domain resource provides an Athenz user-domain resource.
---

# Resource: athenz_user_domain

`athenz_user_domain` provides an Athenz user-domain resource.

Important Note: Use this resource only for create new user domain, update not supported. For import existing one, pls use terraform import.

## Example Usage

```hcl
resource "athenz_user_domain" "athenz_user_domain-test" {
  name = "some_user_id"
  audit_ref = "create domain"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) user id which will be the domain name.

- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


## Import
User-Domain resource can be imported using the User-Domain name: `<domain name>`, e.g.

```hcl
#1. Define empty resource in your <somefile>.tf

    resource "athenz_user_domain" "import_user_domain" {
    }

#2. In the directory where the file is located, run this command:
        
   ֿ$ terraform import athenz_user_domain.import_user_domain <domain name>

#3. Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html