---
page_title: "ipam_tenancy Data Source - dotnet-ipam"
description: Lookup a tenancy by `id` or `name`.
---

# data.ipam_tenancy

Lookup a tenancy by `id` or `name`.

## Example: by name

```hcl
data "ipam_tenancy" "team_a" {
  name = "team-a"
}
```

## Example: by id

```hcl
data "ipam_tenancy" "team_a" {
  id = "00000000-0000-0000-0000-000000000000"
}
```
