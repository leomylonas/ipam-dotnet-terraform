# ipam_user

Manages user lifecycle excluding password updates post-create.

## Schema

- Required: `username`, `password`, `role`
- Optional: `tenancy_id`
- Computed: `id`

## Example

```hcl
resource "ipam_user" "tenant_user" {
  username   = "team-a-user"
  password   = "Example1234"
  role       = "TenantUser"
  tenancy_id = ipam_tenancy.team_a.id
}
```

## Import

```bash
terraform import ipam_user.tenant_user <user_id>
```
