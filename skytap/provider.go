package skytap

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	minTimeout = 10
	delay      = 10
)

// Provider returns a schema.Provider for Skytap.
func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SKYTAP_USERNAME", nil),
				Description: "Username for the skytap account.",
			},
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SKYTAP_API_TOKEN", nil),
				Description: "API Token for the skytap account.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"skytap_project":  dataSourceSkytapProject(),
			"skytap_template": dataSourceSkytapTemplate(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"skytap_project":     resourceSkytapProject(),
			"skytap_environment": resourceSkytapEnvironment(),
			"skytap_network":     resourceSkytapNetwork(),
			"skytap_vm":          resourceSkytapVM(),
		},
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		config := &Config{
			Username: d.Get("username").(string),
			APIToken: d.Get("api_token").(string),
		}

		client, err := config.Client()
		if err != nil {
			return nil, err
		}

		client.StopContext = p.StopContext()

		return client, nil
	}
}
