package main

import (
	"github.com/TritonDataCenter/terraform-provider-triton/triton"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: triton.Provider})
}
