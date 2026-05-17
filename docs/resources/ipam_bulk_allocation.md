# ipam_bulk_allocation

Allocates a contiguous block of IPs.

## Schema

- Required: `subnet_id`, `allocation_count`, `description`
- Computed: `id`, `bulk_id`, `allocation_ids`, `ip_addresses`

## Example

```hcl
resource "ipam_bulk_allocation" "nodes" {
  subnet_id   = ipam_private_subnet.team_a.id
  allocation_count = 3
  description = "worker-nodes"
}
```

## Import

```bash
terraform import ipam_bulk_allocation.nodes <bulk_id>
```
