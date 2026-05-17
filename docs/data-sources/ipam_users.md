# data.ipam_users

List users visible to caller; optionally filter by `tenancy_id`.

## Example

```hcl
data "ipam_users" "team_a" {
  tenancy_id = ipam_tenancy.team_a.id
}
```
