---
page_title: "Policy version data source - terraform-provider-athenz"
subcategory: ""
description: |-
The policy version data source provides details about a specific Athenz policy.
---

# Data Source `athenz_policy_version`

`athenz_policy_version` provides details about a specific Athenz policy with all its versions.

### Example Usage


```hcl
variable "policy_with_versions_name" {
  type = string
}

data "athenz_policy_versions" "selected" {
  name = var.policy_with_versions_name
  domain = "some_domain"
}
```

### Argument Reference

The arguments of this data source act as filters for querying the available policies versions in the current Athenz domain.
The given filters must match exactly one policy with all its versions whose data will be exported as attributes.

- `name` - (Required) The name of the specific Athenz policy.

- `domain` - (Required) The Athenz domain name.
