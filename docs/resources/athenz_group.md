---
page_title: "Group resource - terraform-provider-athenz"
subcategory: ""
description: |-
The group resource provides an Athenz group resource.
---

# Resource: athenz_group

`athenz_group` provides an Athenz group resource.

## Example Usage

IMPORTANT NOTE: please do NOT use json syntax but only hcl syntax

```hcl
resource "athenz_group" "newgrp" {
  name = "some_group"
  domain = "some_domain"
  member {
    name = "user.<user-id>"
  }
  member {
    name = "<domain>.<service-name>"
    expiration = "2022-12-29 23:59:59"
  }
  audit_ref = "create group"
  tags = {
    key1 = "val1,val2"
    key2 = "val3,val4"
  }
}
```

## Example Usage \*\*Deprecated**

```hcl
resource "athenz_group" "newgrp" {
  name = "some_group"
  domain = "some_domain"
  members = ["user.<user-id>", "<domain>.<service-name>"]
  audit_ref = "create group"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The group name.


- `domain` - (Required) The Athenz domain name.


- `members` - \*\*Deprecated** use member attribute instead (Optional) List of Athenz principal members. must be in this format: `user.<user id> or <domain>.<service>`


- `member` - (Optional) A set of Athenz principal members. Each member consists the following arguments:

    - `name` - (Required) The name of the Athenz principal member. must be in this format: `user.<userid> or <domain>.<service>`.

    - `expiration` - (Optional) The expiration of the Athenz principal member. must be in this format: `<yyyy>-<mm>-<dd> <hh>:<MM>:<ss>`


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number

- `tags` - (Optional) Map of tags. The kay is the tag-name and value is the tag-values are represented as a string with a comma separator. e.g. key1 = "val1,val2", this will be converted to: key1 = ["val1", "val2"]

## Import
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