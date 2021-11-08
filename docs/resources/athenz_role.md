---
page_title: "Role resource - terraform-provider-athenz"
subcategory: ""
description: |-
The role resource provides an Athenz role resource.
---

# Resource `athenz_role`

`athenz_role` provides an Athenz role resource.

### Example Usage

```hcl
resource "athenz_role" "foo_role" {
  name = "some_name"
  domain = "some_domain"
  members = ["domain1.user1", "domain2.user2"]
  audit_ref = "create role"
  tags = {
    key1 = "val1,val2"
    key2 = "val3,val4"
  }
  
}
```

### Argument Reference

The following arguments are supported:

- `name` - (Required) The role name.
    

- `domain` - (Required) The Athenz domain name. 
    

- `members` - (Optional) List of Athenz principal members. must be in this format: `user.<userid> or <domain>.<service> or <domain>:group.<group>`.


- `tags` - (Optional) Map of tags. The kay is the tag-name and value is the tag-values are represented as a string with a comma separator. e.g. key1 = "val1,val2", this will be converted to: key1 = ["val1", "val2"]


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


### Import
Role resource can be imported using the role id: `<domain>:role.<role name>`, e.g.

```hcl
#1. Define empty resource in your <somefile>.tf

    resource "athenz_role" "import_role" {
    }

#2. In the directory where the file is located, run this command:
        
   ֿ$ terraform import athenz_role.import_role <domain>:role.<role name> 

#3. Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html    