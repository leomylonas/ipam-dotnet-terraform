# IPAM Terraform Provider

Terraform provider for the [`ipam-dotnet`](../ipam-dotnet) IP Address Management service.

This repository contains the provider implementation, tests, and examples for managing IPAM entities through Terraform.

## Project Summary

The provider enables infrastructure-as-code workflows for a multi-tenant IPAM API built in .NET. It is designed to support role-scoped operations with separate credentials (aliased providers) for:

- `GlobalAdmin`
- `TenantAdmin`
- `TenantUser`

The implementation is planned in two phases:

1. Add required API update endpoints in `ipam-dotnet`.
2. Implement this Terraform provider using the HashiCorp Plugin Framework.

## Planned Features

### Provider Configuration

- `base_url` for the IPAM API endpoint
- `username` / `password` for HTTP Basic Auth
- Optional TLS skip-verify setting for non-production environments
- Configurable request timeout

### Resources (v1)

- `ipam_tenancy`
- `ipam_user`
- `ipam_private_subnet`
- `ipam_shared_subnet`
- `ipam_shared_subnet_access`
- `ipam_exclusion`
- `ipam_allocation`
- `ipam_allocation_tags`

### Data Sources (v1)

- Tenancy/user/subnet/allocation lookup data sources
- `ipam_subnet_stats`
- `ipam_allocations` (optional tag filters)

Not included in v1:

- `ipam_audit_logs` data source

### Lifecycle Semantics

- Update-in-place where API endpoints support updates.
- ForceNew behavior for immutable fields (for example subnet CIDR, exclusion range bounds, allocation identity fields, and shared access identity tuple).
- `ipam_allocation_tags` maps to full-replace API semantics.

## Usage

### 1. Run the IPAM API

Start the `ipam-dotnet` service and ensure you can authenticate with Basic Auth.

### 2. Configure Provider

```hcl
terraform {
  required_providers {
    ipam = {
      source  = "local/ipam"
      version = "0.1.0"
    }
  }
}

provider "ipam" {
  base_url = "http://localhost:8080"
  username = var.ipam_username
  password = var.ipam_password
}
```

### 3. Use Aliased Providers for Role-Scoped Workflows

```hcl
provider "ipam" {
  alias    = "global"
  base_url = "http://localhost:8080"
  username = var.global_admin_username
  password = var.global_admin_password
}

provider "ipam" {
  alias    = "tenant_admin"
  base_url = "http://localhost:8080"
  username = var.tenant_admin_username
  password = var.tenant_admin_password
}
```

### 4. Import ID Conventions

Planned import formats:

- Single ID resources: `tenancy`, `user`, `shared subnet`, `allocation`
- Composite IDs:
  - private subnet: `{tenancy_id}/{subnet_id}`
  - exclusion: `{subnet_id}/{exclusion_id}`
  - shared access: `{subnet_id}/{tenancy_id}`
  - allocation tags: `{allocation_id}`

## Development

### Prerequisites

- Go (current stable)
- Terraform CLI
- Running `ipam-dotnet` API for acceptance tests

### Expected Validation Workflow

- `go test ./...`
- Provider acceptance tests against a local API instance
- `terraform validate` for examples

## Current Status

This repository currently holds the project skeleton and delivery plan. Provider code, tests, and examples are being implemented according to `AGENTS.md`.
