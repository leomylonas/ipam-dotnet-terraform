# ipam_allocation

Allocates one IP from a subnet.

## Schema

- Required: `subnet_id`, `description`
- Computed: `id`, `ip_address`, `user_id`, `allocated_at`, `bulk_id`

## Example

```hcl
resource "ipam_allocation" "app01" {
  subnet_id   = ipam_private_subnet.team_a.id
  description = "app-01"
}
```

## Import

```bash
terraform import ipam_allocation.app01 <allocation_id>
```
