package skytap

import (
	"context"
	"fmt"
	"log"

	"github.com/skytap/skytap-sdk-go/skytap"
)

// Config describes the configuration
type Config struct {
	Username string
	APIToken string
}

// SkytapClient is the Skytap client implementation
type SkytapClient struct {
	StopContext context.Context

	projectsClient          skytap.ProjectsService
	environmentsClient      skytap.EnvironmentsService
	templatesClient         skytap.TemplatesService
	networksClient          skytap.NetworksService
	vmsClient               skytap.VMsService
	interfacesClient        skytap.InterfacesService
	publishedServicesClient skytap.PublishedServicesService
	labelCategoryClient     skytap.LabelCategoryService
}

// Client creates a SkytapClient client
func (c *Config) Client() (*SkytapClient, error) {
	client, err := c.createClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the Skytap client: %v", err)
	}

	skytapClient := SkytapClient{
		projectsClient:          client.Projects,
		environmentsClient:      client.Environments,
		templatesClient:         client.Templates,
		networksClient:          client.Networks,
		vmsClient:               client.VMs,
		interfacesClient:        client.Interfaces,
		publishedServicesClient: client.PublishedServices,
		labelCategoryClient:	 client.LabelCategory,
	}

	return &skytapClient, nil
}

func (c *Config) createClient() (*skytap.Client, error) {
	var credentialsProvider skytap.CredentialsProvider
	if c.APIToken != "" {
		credentialsProvider = skytap.NewAPITokenCredentials(c.Username, c.APIToken)
	} else {
		return nil, fmt.Errorf("an API token must be provided in order to successfully authenticate to Skytap")
	}

	userAgent, err := getUserAgent()
	if err != nil {
		return nil, err
	}

	return skytap.NewClient(skytap.NewDefaultSettings(
		skytap.WithCredentialsProvider(credentialsProvider),
		skytap.WithUserAgent(userAgent)))
}

func getUserAgent() (string, error) {
	log.Printf("[DEBUG] user agent version (version.go): %s", userAgentVersion)
	return fmt.Sprintf("terraform-provider-skytap/%s", userAgentVersion), nil
}
