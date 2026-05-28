terraform {
  required_providers {
    ipam = {
      source  = "registry.terraform.io/leomylonas/dotnet-ipam"
      version = "0.1.0"
    }
  }
}

provider "ipam" {
  alias    = "global"
  base_url = var.ipam_base_url
  username = var.global_admin_username
  password = var.global_admin_password
}

provider "ipam" {
  alias    = "tenant_admin"
  base_url = var.ipam_base_url
  username = var.tenant_admin_username
  password = var.tenant_admin_password
}

provider "ipam" {
  alias    = "tenant_user"
  base_url = var.ipam_base_url
  username = var.tenant_user_username
  password = var.tenant_user_password
}

resource "ipam_tenancy" "team" {
  provider       = ipam.global
  name           = "team-a"
  description    = "Team A tenancy"
  admin_username = "team-a-admin"
  admin_password = "Example1234"
}

resource "ipam_private_subnet" "team_subnet" {
  provider    = ipam.global
  tenancy_id  = ipam_tenancy.team.id
  cidr        = "10.60.0.0/24"
  name        = "team-a-subnet"
  description = "Subnet for team-a"
}

resource "ipam_user" "team_user" {
  provider   = ipam.tenant_admin
  username   = "team-a-user"
  password   = "Example1234"
  role       = "TenantUser"
  tenancy_id = ipam_tenancy.team.id
}

resource "ipam_allocation" "endpoint" {
  provider    = ipam.tenant_user
  subnet_id   = ipam_private_subnet.team_subnet.id
  description = "endpoint-01"
}

resource "ipam_allocation_tags" "endpoint_tags" {
  provider      = ipam.tenant_user
  allocation_id = ipam_allocation.endpoint.id
  tags = {
    env  = "dev"
    app  = "api"
    team = "team-a"
  }
}
