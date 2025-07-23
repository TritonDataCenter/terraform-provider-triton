package main

import (
	"github.com/TritonDataCenter/terraform-provider-triton/triton"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: triton.Provider,
			ProviderAddr: "registry.terraform.io/TritonDataCenter/triton",
		},
	)
}
