package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/joyent/terraform-provider-triton/triton"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: triton.Provider})
}
