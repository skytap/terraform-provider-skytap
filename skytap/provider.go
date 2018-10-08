package skytap

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/opencredo/skytap-sdk-go-internal"
	"github.com/opencredo/skytap-sdk-go-internal/options"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SKYTAP_USER", nil),
				Description: "Username for the skytap account.",
			},
			"access_token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SKYTAP_ACCESS_TOKEN", nil),
				Description: "Token for the skytap account.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"skytap_project": resourceProject(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	username := d.Get("username").(string)
	token := d.Get("access_token").(string)

	client, err := skytap.NewClient(context.Background(),
		options.WithUser(username),
		options.WithAPIToken(token),
		options.WithScheme("https"),
		options.WithHost("cloud.skytap.com"))

	if err != nil {
		return nil, err
	}

	return client, nil
}
