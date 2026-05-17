# leomylonas/dotnet-ipam-terraform

Terraform provider for the `ipam-dotnet` API.

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
- `ipam_allocation_tags`

## Implemented Data Sources

- `ipam_tenancy`
- `ipam_users`
- `ipam_shared_subnets`
- `ipam_private_subnets`
- `ipam_allocation`
- `ipam_allocations`
- `ipam_subnet_stats`

Not included:

- `ipam_audit_logs`

## Provider Configuration

```hcl
provider "ipam" {
  base_url                 = "http://localhost:8080"
  username                 = var.ipam_username
  password                 = var.ipam_password
  timeout_seconds          = 30
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

## Development

Prerequisites:

- Go `1.24+`
- Terraform CLI
- Running `ipam-dotnet` API endpoint

Commands:

```bash
go mod tidy
go test ./...
```

Example configuration is in `examples/basic`.

## Run Locally with a Terraform Module

If you want a separate Terraform module/repo to use your local build of this provider, use Terraform CLI `dev_overrides`.

### 1. Build the provider binary

From this repository:

```bash
go mod tidy
go build -o terraform-provider-dotnet-ipam-terraform .
mkdir -p ./.terraform-dev
mv terraform-provider-dotnet-ipam-terraform ./.terraform-dev/
```

### 2. Create a Terraform CLI config with `dev_overrides`

Create a file (for example `~/.terraformrc` or a temporary file) with:

```hcl
provider_installation {
  dev_overrides {
    "leomylonas/dotnet-ipam-terraform" = "/absolute/path/to/dotnet-ipam-terraform/.terraform-dev"
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
      source  = "leomylonas/dotnet-ipam-terraform"
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
go build -o ./.terraform-dev/terraform-provider-dotnet-ipam-terraform .
```

Then re-run `terraform plan` in the consuming module.

## CI

GitHub Actions workflow (`.github/workflows/ci.yml`) runs:

- `go test ./...`
- `terraform validate` for `examples/basic`
