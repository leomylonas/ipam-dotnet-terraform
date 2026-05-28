---
page_title: "ipam_tenancy Resource - dotnet-ipam"
description: Manages tenancy lifecycle.
---

# ipam_tenancy

Manages tenancy lifecycle.

## Schema

- Required: `name`, `description`, `admin_username`, `admin_password`
- Computed: `id`, `created_at`

## Example

```hcl
resource "ipam_tenancy" "team_a" {
  name           = "team-a"
  description    = "Team A tenancy"
  admin_username = "team-a-admin"
  admin_password = "Example1234"
}
```

## Import

```bash
terraform import ipam_tenancy.team_a <tenancy_id>
```
