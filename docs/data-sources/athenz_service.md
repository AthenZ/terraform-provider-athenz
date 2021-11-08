---
page_title: "Service data source - terraform-provider-athenz"
subcategory: ""
description: |-
The role data source provides details about a specific Athenz service.
---

# Data Source `athenz_service`

`athenz_service` provides details about a specific Athenz service.

### Example Usage

```hcl
variable "service_name" {
  type = string
}

data "athenz_service" "selected" {
  name = var.service_name
  domain = some_domain
}
```

### Argument Reference

The arguments of this data source act as filters for querying the available services in the current Athenz domain.
The given filters must match exactly one service whose data will be exported as attributes.

- `name` - (Required) The name of the specific Athenz service.

- `domain` - (Required) The Athenz domain name.
