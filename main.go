package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/observeinc/terraform-provider-observe/observe"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: observe.Provider})
}
