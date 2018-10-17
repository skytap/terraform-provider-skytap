package skytap

import (
	"context"

	"github.com/pkg/errors"
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

	projectsClient     skytap.ProjectsService
	environmentsClient skytap.EnvironmentsService
}

// Client creates a SkytapClient client
func (c *Config) Client() (*SkytapClient, error) {
	var credentialsProvider skytap.CredentialsProvider
	if c.APIToken != "" {
		credentialsProvider = skytap.NewAPITokenCredentials(c.Username, c.APIToken)
	} else {
		return nil, errors.Errorf("An API token must be provided in order to successfully authenticate to Skytap")
	}

	client, err := skytap.NewClient(skytap.NewDefaultSettings(skytap.WithCredentialsProvider(credentialsProvider)))
	if err != nil {
		return nil, errors.Errorf("failed to initialize the Skytap client: %v", err)
	}

	skytapClient := SkytapClient{
		projectsClient:     client.Projects,
		environmentsClient: client.Environments,
	}

	return &skytapClient, nil
}
