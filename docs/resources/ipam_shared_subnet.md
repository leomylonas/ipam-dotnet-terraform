# ipam_shared_subnet

Manages global shared subnets.

## Schema

- Required: `cidr`, `name`, `description`
- Computed: `id`, `created_at`

## Example

```hcl
resource "ipam_shared_subnet" "shared" {
  cidr        = "172.20.0.0/24"
  name        = "shared-services"
  description = "Shared services subnet"
}
```

## Import

```bash
terraform import ipam_shared_subnet.shared <subnet_id>
```
