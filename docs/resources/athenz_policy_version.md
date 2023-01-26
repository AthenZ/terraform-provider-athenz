---
page_title: "Policy version resource - terraform-provider-athenz"
subcategory: ""
description: |-
The policy version resource provides an Athenz policy with all its versions resource.
---

# Resource: athenz_policy_version

`athenz_policy_version` provides an Athenz policy with all its versions resource.

## Example Usage

IMPORTANT NOTE: please do NOT use json syntax but only hcl syntax

```hcl
resource "athenz_policy_version" "policy_with_version" {
  name = "with_version"
  domain = "some_domain"
  active_version = "version1"
  version {
      version_name = "version1"
      assertion {
          effect = "ALLOW"
          action = "*"
          role = "role1"
          resource = "some_domain:resource1"
        }
    }
  version {
      version_name = "version2"
      assertion {
          effect = "ALLOW"
          action = "*"
          role = "role2"
          resource = "some_domain:RESOURCE2"
          case_sensitive = true
      }
      assertion {
          effect = "DENY"
          action = "PLAY"
          role = "role2"
          resource = "some_domain:resource2"
          case_sensitive = true
      }
    }
  audit_ref = "create policy"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The policy name.

- `domain` - (Required) The Athenz domain name.

- `active_version` - (Required) The active version of the policy. Must match one of the version name defined un the resource

- `version` - (Required) A set of policy versions. Each version consists the following arguments:

    - `version_name` - (Required) The version name.

    - `assertion` - (Optional) A set of assertions that govern usage of resources. where <assertion\> is <effect\> <action\> to <role\> on <resource\>.

        - `effect` - (Required) The value effect must be either ALLOW or DENY.

        - `role` - (Required) The name of the role this assertion applies to. MUST be the role name only (without the prefix `<domain name>:role`).

        - `action` - (Required) The action is the domain administrator defined action available for the resource (e.g. read, write, delete).

        - `resource` - (Required) The resource is the YRN of the resource this assertion applies to. MUST provide fully qualified name: `<domain name>:<resource name>`

        - `case_sensitive` - (Optional Default = false) If true, action and resource will be case-sensitive.


- `audit_ref` - (Optional Default = "done by terraform provider")  string containing audit specification or ticket number.


## Import
Policy with all its versions resource can be imported using the policy id: `<domain>:policy.<policy name>`, e.g.

```hcl
#1. Define empty resource in your <somefile>.tf

    resource "athenz_policy_version" "import_policy_versions" {
    }

#2. In the directory where the file is located, run this command:

   ֿ$ terraform import athenz_policy_version.import_policy_versions <domain>:policy.<policy name>

#3. Make any adjustments to the configuration to align with the current (or desired) state of the imported object.
```
For more information: https://www.terraform.io/docs/cli/import/index.html