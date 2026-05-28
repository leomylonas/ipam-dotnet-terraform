---
page_title: "ipam_private_subnet Resource - dotnet-ipam"
description: Manages tenant private subnets.
---

# ipam_private_subnet

Manages tenant private subnets.

## Schema

- Required: `tenancy_id`, `cidr`, `name`, `description`
- Computed: `id`, `created_at`

## Example

```hcl
resource "ipam_private_subnet" "team_a" {
  tenancy_id  = ipam_tenancy.team_a.id
  cidr        = "10.50.0.0/24"
  name        = "team-a-subnet"
  description = "Team A private subnet"
}
```

## Import

```bash
terraform import ipam_private_subnet.team_a <tenancy_id>/<subnet_id>
```
