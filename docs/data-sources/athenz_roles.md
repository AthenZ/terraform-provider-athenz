##Data Source: athenz_roles

`athenz_roles` This Data Source you can get the list of all roles in a domain with optional flag whether or not include members

### Example Usage

```hcl
variable "tag_kay" {
  type = string
}

variable "tag_value" {
  type = string
}

data "athenz_roles" "selected" {
  domain = "some_domain"
  tag_kay =  var.tag_key
  tag_value = var.tag_value
  include_members = false
}
```

### Argument Reference

The arguments of this data source act as filters for querying the available roles in the current Athenz domain.

- `domain` - (Required) The Athenz domain name.

- `tag_key` - (Optional. Required if tag_value presented) Query all roles that have a given tag_kay.

- `tag_value` - (Optional) Query all roles that have a given tag_key AND tag_value.

- `include_members` - (Optional Default = true) If true - return list of members in the role.