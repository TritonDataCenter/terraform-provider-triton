package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/terraform-providers/terraform-provider-triton/triton"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: triton.Provider})
}
