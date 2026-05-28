terraform {
  required_providers {
    ipam = {
      source  = "registry.terraform.io/leomylonas/ipam"
      version = "~> 0.0"
    }
  }
}

provider "ipam" {
  base_url = "https://ipam.example.com"
  username = "admin"
  password = "changeme"
}
