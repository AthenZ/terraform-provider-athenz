---
page_title: "Group data source - terraform-provider-athenz"
subcategory: ""
description: |-
The group data source provides details about a specific Athenz group.
---

# Data Source `athenz_group`

`athenz_group` provides details about a specific Athenz group.

### Example Usage

```hcl
variable "group_name" {
  type = string
}

data "athenz_group" "selected" {
  name = var.group_name
  domain = "some_domain"
}
```

### Argument Reference

The arguments of this data source act as filters for querying the available groups in the current Athenz domain.
The given filters must match exactly one group whose data will be exported as attributes.

- `name` - (Required) The name of the specific Athenz group.

- `domain` - (Required) The Athenz domain name.
