package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/leomylonas/terraform-provider-dotnet-ipam/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New("dev"), providerserver.ServeOpts{
		Address: "registry.terraform.io/leomylonas/ipam",
	})
	if err != nil {
		log.Fatal(err)
	}
}
