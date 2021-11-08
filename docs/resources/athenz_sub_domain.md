---
page_title: "Sub Domain resource - terraform-provider-athenz"
subcategory: ""
description: |-
The sub-domain resource provides an Athenz sub-domain resource.
---

# Resource `athenz_sub_domain`

`athenz_sub_domain` provides an Athenz sub-domain resource.

Important Note: Use this resource only for create new sub-domain, update not supported. For import existing one, pls use terraform import.

### Example Usage

```hcl
resource "athenz_sub_domain" "sub_domain-test" {
  parent_name="home.some_user"
  name = "test"
  admin_users = ["user.someone"]
  audit_ref = "create domain"
}
```

### Argument Reference

The following arguments are supported:

- `parnet_name` - (Required) name of the parent domain.


- `name` - (Required) name of the domain.


- `admin_users` - (Required) list of domain administrators. must be in this format: `user.<userid> or <domain>.<service>`.


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


### Import
Sub-Domain resource can be imported using the Sub-Domain id: `<parent domain>.<domain name>`, e.g.

```hcl
#1. Define empty resource in your <somefile>.tf

    resource "athenz_sub_domain" "import_sub_domain" {
    }

#2. In the directory where the file is located, run this command:
        
   ֿ$ terraform import athenz_sub_domain.import_sub_domain <parent domain>.<domain name>

#3. Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html