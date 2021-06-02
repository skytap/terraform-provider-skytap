package skytap

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	minTimeout = 10
	delay      = 10
)

// Provider returns a schema.Provider for Skytap.
func Provider() *schema.Provider {
	p := &schema.Provider{
		ConfigureContextFunc: providerConfigure,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SKYTAP_USERNAME", nil),
				Description: "The Skytap username. May also be specified by the `SKYTAP_USERNAME` shell environment variable",
			},
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SKYTAP_API_TOKEN", nil),
				Description: "The Skytap API token. May also be specified by the `SKYTAP_API_TOKEN` shell environment variable",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"skytap_project":  dataSourceSkytapProject(),
			"skytap_template": dataSourceSkytapTemplate(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"skytap_project":        resourceSkytapProject(),
			"skytap_environment":    resourceSkytapEnvironment(),
			"skytap_network":        resourceSkytapNetwork(),
			"skytap_vm":             resourceSkytapVM(),
			"skytap_label_category": resourceSkytapLabelCategory(),
			"skytap_icnr_tunnel":    resourceSkytapICNRTunnel(),
		},
	}

	return p
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := &Config{
		Username: d.Get("username").(string),
		APIToken: d.Get("api_token").(string),
	}

	client, err := config.Client()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}
