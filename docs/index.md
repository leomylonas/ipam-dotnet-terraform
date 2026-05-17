---
page_title: "dotnet-ipam Provider"
description: |-
  The dotnet-ipam provider manages IP address management resources exposed by the dotnet-ipam REST API.
---

# dotnet-ipam Provider

The `dotnet-ipam` provider manages resources in a [dotnet-ipam](https://github.com/leomylonas/dotnet-ipam) server: tenancies, subnets, users, exclusions, allocations, and allocation tags.

The provider uses Basic Auth and supports multi-role workflows via [aliased provider configurations](#multi-role-usage).

## Example Usage

```hcl
terraform {
  required_providers {
    dotnet-ipam = {
      source  = "leomylonas/dotnet-ipam"
      version = "~> 0.0"
    }
  }
}

provider "dotnet-ipam" {
  base_url = "https://ipam.example.com"
  username = "admin"
  password = var.ipam_password
}
```

## Multi-Role Usage

dotnet-ipam enforces role-based access control (`GlobalAdmin`, `TenantAdmin`, `TenantUser`). The recommended pattern is to declare one aliased provider per role so that each resource is managed under the appropriate identity.

```hcl
provider "dotnet-ipam" {
  alias    = "global"
  base_url = var.ipam_base_url
  username = var.global_admin_username
  password = var.global_admin_password
}

provider "dotnet-ipam" {
  alias    = "tenant_admin"
  base_url = var.ipam_base_url
  username = var.tenant_admin_username
  password = var.tenant_admin_password
}

provider "dotnet-ipam" {
  alias    = "tenant_user"
  base_url = var.ipam_base_url
  username = var.tenant_user_username
  password = var.tenant_user_password
}

# GlobalAdmin creates tenancy and subnet
resource "dotnet-ipam_ipam_tenancy" "team" {
  provider    = dotnet-ipam.global
  name        = "team-a"
  description = "Team A tenancy"
}

# TenantUser allocates an IP
resource "dotnet-ipam_ipam_allocation" "host" {
  provider    = dotnet-ipam.tenant_user
  subnet_id   = dotnet-ipam_ipam_private_subnet.team_subnet.id
  description = "app-server-01"
}
```

A `403` response is surfaced as a Terraform error diagnostic identifying the resource and operation that was denied.

## Environment Variables

Credentials can be supplied via environment variables instead of (or in addition to) the provider block. Environment variables take effect when the corresponding provider argument is absent or empty.

| Argument                  | Environment variable              |
|---------------------------|-----------------------------------|
| `base_url`                | `IPAM_BASE_URL`                   |
| `username`                | `IPAM_USERNAME`                   |
| `password`                | `IPAM_PASSWORD`                   |
| `timeout_seconds`         | `IPAM_TIMEOUT_SECONDS`            |
| `insecure_skip_tls_verify`| `IPAM_INSECURE_SKIP_TLS_VERIFY`   |
| `max_retries`             | `IPAM_MAX_RETRIES`                |
| `retry_wait_min_ms`       | `IPAM_RETRY_WAIT_MIN_MS`          |
| `retry_wait_max_ms`       | `IPAM_RETRY_WAIT_MAX_MS`          |

Example using env vars for credentials:

```shell
export IPAM_BASE_URL="https://ipam.example.com"
export IPAM_USERNAME="admin"
export IPAM_PASSWORD="changeme"
```

```hcl
provider "dotnet-ipam" {}
```

## Schema

### Required

`base_url`, `username`, and `password` are required — either in the provider block or via the corresponding environment variables above.

- `base_url` (String) — Base URL of the dotnet-ipam API, e.g. `https://ipam.example.com`.
- `username` (String) — Username for Basic Auth.
- `password` (String, Sensitive) — Password for Basic Auth.

### Optional

- `timeout_seconds` (Number) — HTTP request timeout in seconds. Defaults to `30`.
- `insecure_skip_tls_verify` (Boolean) — Skip TLS certificate verification. Use only in non-production environments. Defaults to `false`.
- `max_retries` (Number) — Maximum number of retries for transient `5xx` and `429` responses. Defaults to `2`.
- `retry_wait_min_ms` (Number) — Minimum wait between retries in milliseconds. Defaults to `200`.
- `retry_wait_max_ms` (Number) — Maximum wait between retries in milliseconds. Defaults to `2000`.
