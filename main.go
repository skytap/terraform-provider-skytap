package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/opencredo/terraform-provider-skytap-internal/skytap"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: skytap.Provider})
}
