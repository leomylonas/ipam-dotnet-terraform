# leomylonas/terraform-provider-dotnet-ipam

Terraform provider for the [`ipam-dotnet`](https://github.com/leomylonas/ipam-dotnet) API.

## What This Project Provides

- Terraform provider built with HashiCorp Plugin Framework.
- HTTP Basic Auth integration to the IPAM API.
- Retry/backoff for transient `429`/`5xx` API failures.
- Core IPAM resources for tenancy, users, subnets, exclusions, allocations, and tags.
- Read-focused data sources for tenancy, users, subnets, allocations, and subnet stats.
- Aliased-provider workflow support for role-scoped access (`GlobalAdmin`, `TenantAdmin`, `TenantUser`).

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

## Provider Configuration

```hcl
provider "ipam" {
  base_url                 = "http://localhost:8080"
  username                 = var.ipam_username
  password                 = var.ipam_password
  timeout_seconds          = 30
  max_retries              = 2
  retry_wait_min_ms        = 200
  retry_wait_max_ms        = 2000
  insecure_skip_tls_verify = false
}
```

## Import ID Formats

- Single ID: tenancy, user, shared subnet, allocation
- Composite IDs:
  - private subnet: `{tenancy_id}/{subnet_id}`
  - exclusion: `{subnet_id}/{exclusion_id}`
  - shared access: `{subnet_id}/{tenancy_id}`
  - allocation tags: `{allocation_id}`
  - bulk allocation: `{bulk_id}`

## Development

Prerequisites:

- Go `1.24+`
- Terraform CLI
- Running `ipam-dotnet` API endpoint

### Makefile workflow

The `Makefile` is the primary local developer interface for repeatable commands.
It standardizes provider build/test/install operations and keeps local + CI workflows consistent.
Run `make help` to see available targets.

Targets:

- `make tidy`: run `go mod tidy`
- `make build`: build provider binary (`terraform-provider-dotnet-ipam`)
- `make test`: run unit/compile tests (`go test ./...`)
- `make testacc`: run acceptance tests (`TestAcc*`) against a real API
- `make dev-install`: build and move provider binary into `.terraform-dev/`
- `make lint`: currently aliases test pass (`go test ./...`)

Recommended day-to-day commands:

```bash
make tidy
make test
```

Example configuration is in `examples/basic` and `examples/multi-role`.

## Run Locally with a Terraform Module

If you want a separate Terraform module/repo to use your local build of this provider, use Terraform CLI `dev_overrides`.

### 1. Build the provider binary

From this repository:

```bash
make tidy
make dev-install
```

### 2. Create a Terraform CLI config with `dev_overrides`

Create a file (for example `~/.terraformrc` or a temporary file) with:

```hcl
provider_installation {
  dev_overrides {
    "leomylonas/terraform-provider-dotnet-ipam" = "/absolute/path/to/terraform-provider-dotnet-ipam/.terraform-dev"
  }
  direct {}
}
```

Use an absolute path to the directory that contains the built provider binary.

### 3. Point Terraform to that CLI config

If you are using a temporary config file:

```bash
export TF_CLI_CONFIG_FILE=/absolute/path/to/terraformrc
```

### 4. Reference the provider in your Terraform module

In the consuming module:

```hcl
terraform {
  required_providers {
    ipam = {
      source  = "leomylonas/terraform-provider-dotnet-ipam"
      version = "0.1.0"
    }
  }
}
```

```hcl
provider "ipam" {
  base_url = "http://localhost:8080"
  username = var.ipam_username
  password = var.ipam_password
}
```

Then run:

```bash
terraform init
terraform plan
```

Terraform will use the local provider build from your `dev_overrides` path.

### 5. Rebuild workflow while developing

When you change provider code:

```bash
make dev-install
```

Then re-run `terraform plan` in the consuming module.

## CI

### `ci` workflow (`.github/workflows/ci.yml`)

Triggers:

- `push`
- `pull_request`

Execution order:

1. `make tidy` to normalize module dependencies.
2. `make test` to run provider unit/compile tests.
3. `make dev-install` to build and stage a local provider binary in `.terraform-dev`.
4. Terraform validation for both examples with a `dev_overrides` CLI config:
   - `examples/basic`
   - `examples/multi-role`

This ensures code compiles/tests, the provider binary builds, and published examples remain valid.

### `release` workflow (`.github/workflows/release.yml`)

Trigger:

- `push` tag matching `v*` (for example `v0.1.0`)

Execution order:

1. `make tidy`
2. GoReleaser `release --clean` using `.goreleaser.yaml`
3. Publish release artifacts/checksums to GitHub Releases

## Additional Docs

- `docs/TROUBLESHOOTING.md`
- `docs/COMPATIBILITY.md`
- `docs/RELEASE.md`
- `docs/resources/*`
- `docs/data-sources/*`

## Acceptance Tests

Acceptance tests run against a real IPAM API instance.

Required environment variables:

- `IPAM_ACC=1`
- `IPAM_BASE_URL`
- `IPAM_USERNAME`
- `IPAM_PASSWORD`

Run locally:

```bash
make testacc
```

Or directly:

```bash
IPAM_ACC=1 IPAM_BASE_URL=http://localhost:8080 IPAM_USERNAME=admin IPAM_PASSWORD=Admin1234! go test ./internal/provider -v -run 'TestAcc' -count=1
```
