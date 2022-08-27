---
page_title: "Policy resource - terraform-provider-athenz"
subcategory: ""
description: |-
The policy resource provides an Athenz policy resource.
---

# Resource `athenz_policy`

`athenz_policy` provides an Athenz policy resource.

### Example Usage

```hcl
resource "athenz_policy" "foo_policy" {
  name = "foo"
  domain = "some_domain"
  assertion = [
    {
      effect = "ALLOW"
      action = "some_action"
      role = "some_role_name"
      resource = "some_domain:some_resource"
  }]
  audit_ref = "create policy"
}
```

### Argument Reference

The following arguments are supported:

- `name` - (Required) The policy name.


- `domain` - (Required) The Athenz domain name.


- `assertion` (Optional) - A set of assertions that govern usage of resources. where <assertion\> is <effect\> <action\> to <role\> on <resource\>.
  
    - `effect` - (Required) The value effect must be either ALLOW or DENY.
      
    - `role` - (Required) The name of the role this assertion applies to. MUST be the role name only (without the prefix `<domain name>:role`)
      
    - `action` - (Required) The action is the domain administrator defined action available for the resource (e.g. read, write, delete).
      
    - `resource` - (Required) The resource is the YRN of the resource this assertion applies to. MUST provide fully qualified name: `<domain name>:<resource name>`
      
    - `case_sensitive` - (Optional) If true, action and resource will be case-sensitive.


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


### Import
Policy resource can be imported using the policy id: `<domain>:policy.<policy name>`, e.g.

```hcl
#1. Define empty resource in your <somefile>.tf

    resource "athenz_policy" "import_policy" {
    }

#2. In the directory where the file is located, run this command:
        
   ֿ$ terraform import athenz_policy.import_policy <domain>:policy.<policy name> 

#3. Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html    