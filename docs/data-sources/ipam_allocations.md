# data.ipam_allocations

List allocations, optional filters: `tag_key`, `tag_value`.

## Example: all allocations

```hcl
data "ipam_allocations" "all" {}
```

## Example: filtered by tag

```hcl
data "ipam_allocations" "dev" {
  tag_key   = "env"
  tag_value = "dev"
}
```
