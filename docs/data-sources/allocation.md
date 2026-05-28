---
page_title: "ipam_allocation Data Source - dotnet-ipam"
description: Lookup allocation by `id`.
---

# data.ipam_allocation

Lookup allocation by `id`.

## Example

```hcl
data "ipam_allocation" "app01" {
  id = ipam_allocation.app01.id
}
```
