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
  password = "changeme"
}
