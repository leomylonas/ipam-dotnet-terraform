---
page_title: "ipam_subnet_stats Data Source - dotnet-ipam"
description: Fetch live subnet utilization stats for `subnet_id`.
---

# data.ipam_subnet_stats

Fetch live subnet utilization stats for `subnet_id`.

## Example

```hcl
data "ipam_subnet_stats" "team_a" {
  subnet_id = ipam_private_subnet.team_a.id
}
```
