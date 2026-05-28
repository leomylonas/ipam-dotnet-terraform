---
page_title: "ipam_exclusion Resource - dotnet-ipam"
description: Manages subnet exclusion ranges.
---

# ipam_exclusion

Manages subnet exclusion ranges.

## Schema

- Required: `subnet_id`, `start`, `end`, `description`
- Computed: `id`

## Example

```hcl
resource "ipam_exclusion" "gateway" {
  subnet_id    = ipam_private_subnet.team_a.id
  start        = "10.50.0.1"
  end          = "10.50.0.1"
  description  = "Gateway IP"
}
```

## Import

```bash
terraform import ipam_exclusion.gateway <subnet_id>/<exclusion_id>
```
