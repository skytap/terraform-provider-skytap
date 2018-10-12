package skytap

import (
	"context"

	"github.com/pkg/errors"
	"github.com/skytap/skytap-sdk-go/skytap"
)

type Config struct {
	Username string
	Password string
	ApiToken string
}

type SkytapClient struct {
	StopContext context.Context

	projectsClient     skytap.ProjectsService
	environmentsClient skytap.EnvironmentsService
}

func (c *Config) Client() (*SkytapClient, error) {
	var credentialsProvider skytap.CredentialsProvider
	if c.ApiToken != "" {
		credentialsProvider = skytap.NewApiTokenCredentials(c.Username, c.ApiToken)
	} else if c.Password != "" {
		credentialsProvider = skytap.NewPasswordCredentials(c.Username, c.Password)
	} else {
		return nil, errors.Errorf("Either a password or an Api token must be provided in order to successfully authenticate to SkyTap")
	}

	client, err := skytap.NewClient(skytap.NewDefaultSettings(skytap.WithCredentialsProvider(credentialsProvider)))
	if err != nil {
		return nil, errors.Errorf("Failed to initialize the SkyTap client: %v", err)
	}

	skytapClient := SkytapClient{
		projectsClient:     client.Projects,
		environmentsClient: client.Environments,
	}

	return &skytapClient, nil
}
