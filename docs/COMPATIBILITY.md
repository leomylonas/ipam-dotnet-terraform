# Compatibility Matrix

| Component | Supported |
|---|---|
| Terraform CLI | >= 1.6 |
| Go toolchain (for development) | 1.24.x |
| IPAM API | Current `ipam-dotnet` API surface used by this provider |

## Immutable (ForceNew) Fields

These fields are intentionally modeled as replace-on-change:

- Private/shared subnet `cidr`
- Exclusion `start` / `end`
- Allocation request fields (`subnet_id`, `description`)
- Shared access identity (`subnet_id`, `tenancy_id`)
- Tenancy admin bootstrap credentials (`admin_username`, `admin_password`)

If API behavior changes, re-evaluate lifecycle settings and acceptance tests.
