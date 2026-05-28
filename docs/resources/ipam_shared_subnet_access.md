---
page_title: "ipam_shared_subnet_access Resource - dotnet-ipam"
description: Manages shared subnet tenancy access grants.
---

# ipam_shared_subnet_access

Manages shared subnet tenancy access grants.

## Schema

- Required: `subnet_id`, `tenancy_id`
- Computed: `id`

## Example

```hcl
resource "ipam_shared_subnet_access" "team_a" {
  subnet_id  = ipam_shared_subnet.shared.id
  tenancy_id = ipam_tenancy.team_a.id
}
```

## Import

```bash
terraform import ipam_shared_subnet_access.team_a <subnet_id>/<tenancy_id>
```
