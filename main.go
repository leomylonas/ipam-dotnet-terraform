package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/leomylonas/dotnet-ipam-terraform/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New("dev"), providerserver.ServeOpts{
		Address: "registry.terraform.io/leomylonas/dotnet-ipam-terraform",
	})
	if err != nil {
		log.Fatal(err)
	}
}
