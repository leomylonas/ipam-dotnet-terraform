# AGENTS: Current Implementation Guide

## Summary

`leomylonas/dotnet-ipam` is a Terraform provider for the `ipam-dotnet` API.

Current implementation includes:

- Full provider/resource/data-source scaffolding with HashiCorp Plugin Framework.
- Role-aware workflows via provider aliases (`GlobalAdmin`, `TenantAdmin`, `TenantUser`).
- HTTP Basic Auth client with retry/backoff controls.
- Resource and data-source docs with runnable examples.
- Acceptance test harness for real API environments.
- CI and release scaffolding.

## Provider Behavior

Provider schema:

- `base_url` (required)
- `username` (required)
- `password` (required, sensitive)
- `timeout_seconds` (optional, default `30`)
- `insecure_skip_tls_verify` (optional, default `false`)
- `max_retries` (optional, default `2`)
- `retry_wait_min_ms` (optional, default `200`)
- `retry_wait_max_ms` (optional, default `2000`)

Client behavior:

- Uses Basic Auth on every request.
- Retries transient failures (`429`, `5xx`, retryable network errors).
- Exponential backoff bounded by configured retry wait min/max.
- Returns normalized API errors with status/body context.

## Implemented Resources

- `ipam_tenancy`
- `ipam_user`
- `ipam_private_subnet`
- `ipam_shared_subnet`
- `ipam_shared_subnet_access`
- `ipam_exclusion`
- `ipam_allocation`
- `ipam_bulk_allocation`
- `ipam_allocation_tags`

## Implemented Data Sources

- `ipam_tenancy`
- `ipam_users`
- `ipam_shared_subnets`
- `ipam_private_subnets`
- `ipam_allocation`
- `ipam_allocations`
- `ipam_bulk_allocations`
- `ipam_subnet_stats`

Not implemented:

- `ipam_audit_logs`

## Lifecycle and Import Semantics

Update-in-place resources:

- `ipam_tenancy` (`name`, `description`)
- `ipam_user` (`username`, `role`, `tenancy_id`)
- `ipam_private_subnet` (`name`, `description`)
- `ipam_shared_subnet` (`name`, `description`)
- `ipam_exclusion` (`description`)
- `ipam_allocation_tags` (full replace)

Replace-on-change resources/fields:

- `ipam_shared_subnet_access` identity (`subnet_id`, `tenancy_id`)
- `ipam_allocation` request fields (`subnet_id`, `description`)
- `ipam_bulk_allocation` request fields (`subnet_id`, `count`, `description`)
- Private/shared subnet `cidr`
- Exclusion `start` / `end`
- Tenancy bootstrap admin credentials (`admin_username`, `admin_password`)

Import formats:

- Single ID: tenancy, user, shared subnet, allocation, bulk allocation
- Composite IDs:
  - private subnet: `{tenancy_id}/{subnet_id}`
  - exclusion: `{subnet_id}/{exclusion_id}`
  - shared subnet access: `{subnet_id}/{tenancy_id}`
  - allocation tags: `{allocation_id}`

## API Alignment Assumptions

Provider update behavior assumes these API endpoints exist:

- `PUT /api/tenancies/{id}`
- `PUT /api/users/{id}`
- `PUT /api/tenancies/{tenancyId}/subnets/{subnetId}`
- `PUT /api/subnets/shared/{id}`
- `PUT /api/subnets/{subnetId}/exclusions/{id}`

If API contracts change, re-evaluate lifecycle settings and acceptance tests.

## Testing

Unit/build checks:

- `go test ./...`

Acceptance tests (real API):

- Environment variables required:
  - `IPAM_ACC=1`
  - `IPAM_BASE_URL`
  - `IPAM_USERNAME`
  - `IPAM_PASSWORD`
- Run:
  - `make testacc`
  - or `go test ./internal/provider -v -run 'TestAcc' -count=1`

Acceptance coverage includes resource CRUD/import paths and allocation data-source checks.

## Examples and Docs

Examples:

- `examples/basic`
- `examples/multi-role`

Reference docs:

- `docs/resources/*`
- `docs/data-sources/*`
- `docs/TROUBLESHOOTING.md`
- `docs/COMPATIBILITY.md`
- `docs/RELEASE.md`

## CI and Release

CI workflows:

- `.github/workflows/ci.yml`:
  - `go test ./...`
  - provider build
  - `terraform validate` for `examples/basic` and `examples/multi-role` via dev overrides
- `.github/workflows/acceptance.yml`:
  - manual acceptance test run using repository secrets

Release scaffolding:

- `.goreleaser.yaml` for cross-platform archives and checksums
- `Makefile` targets for build/test/acceptance/dev-install
