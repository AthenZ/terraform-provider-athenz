##Data Source: athenz_policy

`athenz_policy` provides details about a specific Athenz policy.

### Example Usage


```hcl
variable "policy_name" {
  type = string
}

data "athenz_policy" "selected" {
  name = var.policy_name
  domain = "some_domain"
}
```

### Argument Reference

The arguments of this data source act as filters for querying the available policies in the current Athenz domain.
The given filters must match exactly one policy whose data will be exported as attributes.

- `name` - (Required) The name of the specific Athenz policy.

- `domain` - (Required) The Athenz domain name.
