---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "athenz_group_members Resource - terraform-provider-athenz"
subcategory: ""
description: |-
  The athenz_group_members resource provides support for managing members of an existing athenz group
---
  
---

# athenz_group_members (Resource)

`athenz_group_members` provides support for managing members of an existing athenz group

## Example Usage

IMPORTANT NOTE: please do NOT use json syntax but only hcl syntax

```hcl
resource "athenz_group_members" "newgrp" {
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
}
```

## Argument Reference

### Required

- `domain` (String) Name of the domain that group belongs to
- `name` (String) Name of the standard group role

### Optional

- `audit_ref` (String, Default = "done by terraform provider")  string containing audit specification or ticket number.
- `member` (Block Set) Users or services to be added as members with attribute (see [below for nested schema](#nestedblock--member))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--member"></a>
### Nested Schema for `member`

Required:

- `name` (String) - The name of the Athenz principal member. must be in this format: `user.<userid> or <domain>.<service>`.

Optional:

- `expiration` (String) - The expiration of the Athenz principal member. must be in this format: `<yyyy>-<mm>-<dd> <hh>:<MM>:<ss>`