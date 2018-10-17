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

// Skytap is the Skytap client implementation
type Skytap struct {
	StopContext context.Context

	projectsClient     skytap.ProjectsService
	environmentsClient skytap.EnvironmentsService
}

// Client creates a Skytap client
func (c *Config) Client() (*Skytap, error) {
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

	skytapClient := Skytap{
		projectsClient:     client.Projects,
		environmentsClient: client.Environments,
	}

	return &skytapClient, nil
}
