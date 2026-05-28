# Troubleshooting

## 401 Unauthorized

- Verify `username` and `password` in provider config.
- Ensure credentials are valid for HTTP Basic Auth.
- Confirm `base_url` points to the right IPAM instance.

## 403 Forbidden

- Verify role permissions:
  - `GlobalAdmin`: tenancy/shared subnet lifecycle and global operations.
  - `TenantAdmin`: tenant-scoped operations.
  - `TenantUser`: allocation/release/tag operations only.
- Use aliased providers for mixed-role workflows.

## 404 Not Found

- Resource may have been deleted out-of-band.
- Tenancy-scoped lookups can fail when using wrong tenancy credentials.

## 409 Conflict

- Duplicate tenancy names or overlapping CIDRs.
- Bulk/allocation exhaustion scenarios.

## Terraform not using local provider build

- Ensure `TF_CLI_CONFIG_FILE` points to a config with `dev_overrides`.
- Ensure the override address matches: `registry.terraform.io/leomylonas/dotnet-ipam`.
