---
page_title: "Role data source - terraform-provider-athenz"
subcategory: ""
description: |-
The role data source provides details about a specific Athenz role.
---

# Data Source: athenz_role

`athenz_role` provides details about a specific Athenz role.

## Example Usage

```hcl
variable "role_name" {
  type = string
}

data "athenz_role" "selected" {
  name= var.role_name
  domain = "some_domain"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available roles in the current Athenz domain. 
The given filters must match exactly one role whose data will be exported as attributes.

- `name` - (Required) The name of the specific Athenz role.

- `domain` - (Required) The Athenz domain name.
