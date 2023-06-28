---
page_title: "All domain details data source - terraform-provider-athenz"
subcategory: ""
description: |-
The all domain details data source provides full details about a specific Athenz domain.
---

# Data Source: athenz_all_domain_details

`athenz_all_domain_details` provides full details about a specific Athenz domain (top-level domain/ sub domain/ personal domain)

Note: It can be useful for import  

## Example Usage

```hcl
variable "domain_name" {
  type = string
}

data "athenz_all_domain_details" "domain-test" {
  name = var.domain_name
}
```

## Argument Reference

The arguments of this data source act as filters for querying the current Athenz domain.

- `name` - (Required) The name of the specific Athenz domain. Must be fully qualified name.

## Attribute Reference

The following attributes are exported in addition to the `name`

- `aws_account_id` - The accound id from aws if present for the domain
- `gcp_project_name` - GCP project name if present for the domain
- `gcp_project_number` - GCP project number if it is present for the domain
- `azure_subscription` - Azure subscription if present for the domain
- `role_list` - List of roles for the domain
- `policy_list` - List of policies in the domain
- `service_list` - List of services present in the domain
- `group_list` - List of groups in the domain
