terraform {
  required_providers {
    ipam = {
      source  = "leomylonas/dotnet-ipam-terraform"
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

resource "ipam_tenancy" "example" {
  provider       = ipam.global
  name           = "example-tenancy"
  description    = "Created by Terraform"
  admin_username = "example-tadmin"
  admin_password = "Example1234!"
}

resource "ipam_private_subnet" "tenant_subnet" {
  provider    = ipam.global
  tenancy_id  = ipam_tenancy.example.id
  cidr        = "10.50.0.0/24"
  name        = "tenant-subnet"
  description = "Tenant private subnet"
}

resource "ipam_allocation" "host" {
  provider    = ipam.tenant_admin
  subnet_id   = ipam_private_subnet.tenant_subnet.id
  description = "app-server-01"
}

resource "ipam_allocation_tags" "host_tags" {
  provider      = ipam.tenant_admin
  allocation_id = ipam_allocation.host.id
  tags = {
    env   = "dev"
    owner = "platform"
  }
}

data "ipam_subnet_stats" "tenant_subnet" {
  provider  = ipam.tenant_admin
  subnet_id = ipam_private_subnet.tenant_subnet.id
}

output "allocated_ip" {
  value = ipam_allocation.host.ip_address
}
