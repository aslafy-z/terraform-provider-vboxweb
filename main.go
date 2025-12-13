package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "example.com/local/vboxweb",
	})
	if err != nil {
		log.Fatal(err)
	}
}
