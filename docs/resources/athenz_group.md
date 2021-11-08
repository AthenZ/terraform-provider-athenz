---
page_title: "Group resource - terraform-provider-athenz"
subcategory: ""
description: |-
The group resource provides an Athenz group resource.
---

# Resource `athenz_group`

`athenz_group` provides an Athenz group resource.

### Example Usage

```hcl
resource "athenz_group" "newgrp" {
  name = "some_group"
  domain = "some_domain"
  members = ["user.<user-id>", "<domain>.<service-name>"]
  audit_ref = "create group"
}
```

### Argument Reference

The following arguments are supported:

- `name` - (Required) The group name.


- `domain` - (Required) The Athenz domain name.


- `members` - (Optional) List of Athenz principal members. must be in this format: `user.<user id> or <domain>.<service>`


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number


### Import
Group resource can be imported using the group id: `<domain>:group.<group name>`, e.g.

```hcl
1. Define empty resource in your <somefile>.tf

    resource "athenz_group" "import_group" {
    }

2. In the directory where the file is located, run this command:
        
    terraform import athenz_group.import_group <domain>:group.<group name> 

3.  Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html