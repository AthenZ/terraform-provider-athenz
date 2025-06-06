---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "athenz_group_meta Resource - terraform-provider-athenz"
subcategory: ""
description: |-
  Group Meta Attribute resource.
---

## Example Usage

IMPORTANT NOTE: please do NOT use json syntax but only hcl syntax

```hcl
resource "athenz_group_meta" "group_meta" {
  name = "some_group"
  domain = "some_domain"

  user_expiry_days = 90
  service_expiry_days = 120
  max_members = 0
  self_serve = true
  self_renew = false
  self_renew_mins = 90
  delete_protection = false
  review_enabled = false
  user_authority_filter = "OnShore-US"
  user_authority_expiration = "ElevatedClearance"
  notify_roles = "role1,role2"
  notify_roles = "notify details"
  principal_domain_filter = "user,home,+sports,-sports.dev"
  tags = {
    key1 = "val1,val2"
    key2 = "val3,val4"
  }
  audit_ref = "update group meta"
}
```

# athenz_group_meta (Resource)

`athenz_group_meta` provides an Athenz group meta resource.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain` (String) name of the domain
- `name` (String) Name of the group

### Optional

- `audit_enabled` (Bool) audit enabled flag for the group
- `audit_ref` (String, Default = "done by terraform provider")  string containing audit specification or ticket number.
- `delete_protection` (Bool) If true, ask for delete confirmation in audit and review enabled groups
- `max_members` (Number) maximum number of members allowed in the group
- `notify_details` (String) Set of instructions included in notifications for review and audit enabled groups
- `notify_roles` (String) comma seperated list of roles whose members should be notified for member review/approval
- `principal_domain_filter` (String) comma seperated list of domains to enforce principal membership
- `resource_state` (Number) Bitmask of resource state flags controlling group behavior when creating or destroying the resource. 0x01: create the group if not already present, 0x02: always delete the group when destroying the resource. Default value is -1 indicating to inherit the value defined at the provider configuration level.
- `review_enabled` (Bool) Flag indicates whether group updates require another review and approval
- `self_renew` (Bool) Flag indicates whether to allow expired members to renew their membership
- `self_renew_mins` (Number) Number of minutes members can renew their membership if self review option is enabled
- `self_serve` (Bool) Flag indicates whether group allows self-service. Users can add themselves in the group, but it has to be approved by domain admins to be effective.
- `service_expiry_days` (Number) all services in the group will have specified max expiry days
- `tags` (Map of String) map of group tags
- `user_authority_filter` (String) membership filtered based on user authority configured attributes
- `user_authority_expiration` (String) expiration enforced by a user authority configured attribute
- `user_expiry_days` (Number) all user members in the group will have specified max expiry days

### Read-Only

- `id` (String) The ID of this resource.
