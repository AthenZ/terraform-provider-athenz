---
page_title: "Role resource - terraform-provider-athenz"
subcategory: ""
description: |-
The role resource provides an Athenz role resource.
---

# Resource: athenz_role

`athenz_role` provides an Athenz role resource.

## Example Usage \*\*Deprecated** (please use as explained in the second example)

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

## Example Usage

IMPORTANT NOTE: please do NOT use json syntax but only hcl syntax

```hcl
resource "athenz_role" "foo_role" {
  name = "some_name"
  domain = "some_domain"
  member {
    name = "domain1.user1"
  }
  member {
    name = "domain2.user2"
    expiration = "2022-12-29 23:59:59"
  }
  member {
    name = "domain3.user3"
    review = "2023-12-29 23:59:59"
  }
  settings {
    token_expiry_mins = 60
    cert_expiry_mins = 60
    user_expiry_days = 7
    user_review_days = 7
    group_expiry_days = 14
    group_review_days = 14
    service_expiry_days = 21
    service_review_days = 21
  }
  audit_ref = "create role"
  tags = {
    key1 = "val1,val2"
    key2 = "val3,val4"
  }
  
}
```

## Example Delegated Role Usage 

```hcl
resource "athenz_role" "foo_role" {
  name = "some_name"
  domain = "some_domain"
  trust = "some_delegated_domain"
  audit_ref = "create delegated role"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The role name.
    

- `domain` - (Required) The Athenz domain name. 
    

- `members` - \*\*Deprecated** use member attribute instead (Optional) List of Athenz principal members. must be in this format: `user.<userid> or <domain>.<service> or <domain>:group.<group>`.


- `member` - (Optional) A set of Athenz principal members. Each member consists the following arguments:

  - `name` - (Required) The name of the Athenz principal member. must be in this format: `user.<userid> or <domain>.<service> or <domain>:group.<group>`.

  - `expiration` - (Optional) The expiration time in UTC of the Athenz principal member. must be in this format: `<yyyy>-<mm>-<dd> <hh>:<MM>:<ss>`

  - `review` - (Optional) The review time in UTC of the Athenz principal member. must be in this format: `<yyyy>-<mm>-<dd> <hh>:<MM>:<ss>`


- `settings` - (Optional) A map of advanced settings with the following options:
  - `token_expiry_mins` - (Optional) Tokens issued for this role will have specified max timeout in mins
  - `cert_expiry_mins` - (Optional) Certs issued for this role will have specified max timeout in mins
  - `user_expiry_days` - (Optional) All user members in the role will have specified max expiry days
  - `user_review_days` - (Optional) All user members in the role will have specified max review days
  - `group_expiry_days` - (Optional) All group members in the role will have specified max expiry days
  - `group_review_days` - (Optional) All groups in the role will have specified max review days
  - `service_expiry_days` - (Optional) All services in the role will have specified max expiry days
  - `service_review_days` - (Optional) All service members in the role will have specified review days


- `tags` - (Optional) Map of tags. The kay is the tag-name and value is the tag-values are represented as a string with a comma separator. e.g. key1 = "val1,val2", this will be converted to: key1 = ["val1", "val2"]

- `trust` - (Optional) The domain, which this role is trusted to.

- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


## Import
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