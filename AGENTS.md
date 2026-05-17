# AGENTS Plan: IPAM Terraform Provider + API Prerequisites

## Summary

Build this as a two-phase delivery:

1. Extend the IPAM API to add missing update operations needed for stable Terraform lifecycle semantics.
2. Implement a Terraform provider (Plugin Framework) with aliased-provider workflows for multi-role operation (`GlobalAdmin`, `TenantAdmin`, `TenantUser`).

V1 scope includes core resources plus allocations and tags, plus read-focused data sources, excluding audit logs.

## Key Changes

### 1. API prerequisites in `ipam-dotnet` (Phase 1)

Add update endpoints:

- `PATCH/PUT /api/tenancies/{id}` for `name`, `description`.
- `PATCH/PUT /api/tenancies/{tenancyId}/subnets/{subnetId}` for private subnet `name`, `description` only.
- `PATCH/PUT /api/subnets/shared/{id}` for shared subnet `name`, `description` only.
- `PATCH/PUT /api/subnets/{subnetId}/exclusions/{id}` for exclusion `description` only.
- `PATCH/PUT /api/users/{id}` for non-password mutable user profile fields (minimum: `username`; role/tenancy optional based on business rules).

Keep immutable-by-design fields create-only:

- subnet `cidr`
- exclusion `start` / `end`
- allocation identity fields
- shared access identity tuple

### 2. Provider implementation in `leomylonas/dotnet-ipam-terraform` (Phase 2)

Use `terraform-plugin-framework` with standard provider architecture.

Provider config:

- `base_url`
- `username`
- `password` (sensitive)
- optional insecure TLS handling for non-prod
- request timeout

HTTP client layer requirements:

- Basic Auth on every request
- consistent error mapping
- retry/backoff for transient `5xx` / `429`

Resources (v1):

- `ipam_tenancy`
- `ipam_user`
- `ipam_private_subnet`
- `ipam_shared_subnet`
- `ipam_shared_subnet_access`
- `ipam_exclusion`
- `ipam_allocation`
- `ipam_allocation_tags`

Data sources (v1):

- tenancy/users/subnets/allocation lookups
- `ipam_subnet_stats`
- `ipam_allocations` (with optional tag filters)
- **no `ipam_audit_logs` data source**

Lifecycle rules:

- Use Update where supported by API.
- ForceNew for immutable attributes:
  - subnet `cidr`
  - exclusion `start` / `end`
  - allocation request attributes
  - shared access identity
- `ipam_allocation_tags` should use full replace behavior to match API `PUT` semantics.

Import ID conventions:

- Single ID: tenancy, user, shared subnet, allocation
- Composite IDs:
  - private subnet: `{tenancy_id}/{subnet_id}`
  - exclusion: `{subnet_id}/{exclusion_id}`
  - shared access: `{subnet_id}/{tenancy_id}`
  - allocation tags: `{allocation_id}`

### 3. Multi-role usage model

Document and support aliased providers (for example `ipam.global`, `ipam.tenant_admin`, `ipam.tenant_user`).

Provider should return clear diagnostics for insufficient permissions (`403`).

### 4. Quality and release setup

Acceptance test setup:

- Run local API (`docker-compose` or scripted `dotnet run`)
- Seed identities for each role

CI pipeline expectations:

- `go test`
- acceptance tests
- `terraform validate` on examples
- docs generation

## Test Plan

### 1. API phase tests (`ipam-dotnet`)

- Integration tests for each new update endpoint:
  - authorized success
  - forbidden cross-tenancy / insufficient role
  - not found
  - conflict / validation failures
- Regression coverage for existing create/delete/list behavior.

### 2. Provider unit/integration tests

- CRUD + import for each resource.
- ForceNew checks for immutable attributes.
- Update-in-place checks for newly updatable attributes.
- Auth/authorization diagnostics (`401` / `403`).
- Drift/read checks for aliased provider tenancy scoping.

### 3. End-to-end Terraform scenarios

- Global admin creates tenancy/shared subnet/access grants.
- Tenant admin creates private subnet/exclusions/users.
- Tenant user allocates/releases IP and manages tags.
- Stats/allocation data sources return expected scoped results.

## Public Interfaces / Types Affected

- New API DTOs and update endpoints in `ipam-dotnet`.
- Provider schema surface:
  - provider arguments (`base_url`, `username`, `password`, etc.)
  - resource/data source schemas
  - documented import formats

## Assumptions and Defaults

- Provider stack: `terraform-plugin-framework`.
- Aliased provider pattern is the primary multi-role workflow model.
- V1 includes allocations and tags.
- Audit logs are not exposed as a Terraform data source in v1.
- Any field still lacking API update support after Phase 1 remains ForceNew.
