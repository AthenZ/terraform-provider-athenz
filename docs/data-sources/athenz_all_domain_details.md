##Data Source: athenz_all_domain_details

`athenz_all_domain_details` provides full details about a specific Athenz domain (top-level domain/ sub domain/ personal domain)

Note: It can be useful for import  

### Example Usage

```hcl
variable "domain_name" {
  type = string
}

data "athenz_all_domain_details" "domain-test" {
  name = var.domain_name
}
```

### Argument Reference

The arguments of this data source act as filters for querying the current Athenz domain.

- `name` - (Required) The name of the specific Athenz domain. Must be fully qualified name.
