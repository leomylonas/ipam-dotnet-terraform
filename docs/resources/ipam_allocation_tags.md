# ipam_allocation_tags

Replaces all tags on an allocation.

## Schema

- Required: `allocation_id`, `tags`
- Computed: `id`

## Example

```hcl
resource "ipam_allocation_tags" "app01" {
  allocation_id = ipam_allocation.app01.id
  tags = {
    env   = "dev"
    owner = "platform"
    app   = "api"
  }
}
```

## Import

```bash
terraform import ipam_allocation_tags.app01 <allocation_id>
```
