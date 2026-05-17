# data.ipam_bulk_allocations

List allocations by `bulk_id`.

## Example

```hcl
data "ipam_bulk_allocations" "nodes" {
  bulk_id = ipam_bulk_allocation.nodes.bulk_id
}
```
