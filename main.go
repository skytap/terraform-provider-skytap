package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/terraform-providers/terraform-provider-skytap/skytap"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: skytap.Provider})
}
