package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/skytap/terraform-provider-skytap/skytap"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: skytap.Provider})
}
