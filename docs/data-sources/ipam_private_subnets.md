# data.ipam_private_subnets

List private subnets for required `tenancy_id`.

## Example

```hcl
data "ipam_private_subnets" "team_a" {
  tenancy_id = ipam_tenancy.team_a.id
}
```
