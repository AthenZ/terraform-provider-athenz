---
page_title: "Domain data source - terraform-provider-athenz"
subcategory: ""
description: |-
The domain data source provides details about a specific Athenz domain.
---

# Data Source: athenz_domain

`athenz_domain` provides details about a specific Athenz domain (top-level domain/ sub domain/ personal domain)

## Example Usage

```hcl
variable "domain_name" {
  type = string
}

data "athenz_domain" "domain-test" {
  name = var.domain_name
}
```

## Argument Reference

The arguments of this data source act as filters for querying the current Athenz domain.

- `name` - (Required) The name of the specific Athenz domain. must be fully qualified name.
