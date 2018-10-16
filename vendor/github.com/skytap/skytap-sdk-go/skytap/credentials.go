package skytap

import (
	"context"
	"encoding/base64"
	"fmt"
)

// A CredentialsProvider is the interface for any component which will provide credentials.
// A CredentialsProvider is required to manage its own state
type CredentialsProvider interface {
	// Retrieve returns the authorization header value to be used in the request
	// Error is returned if the value were not obtainable, or empty.
	Retrieve(ctx context.Context) (string, error)
}

type NoOpCredentials struct{}

func (c *NoOpCredentials) Retrieve(ctx context.Context) (string, error) {
	return "", nil
}

func NewNoOpCredentials() *NoOpCredentials {
	return &NoOpCredentials{}
}

type PasswordCredentials struct {
	Username string
	Password string
}

func (c *PasswordCredentials) Retrieve(ctx context.Context) (string, error) {
	return buildBasicAuth(c.Username, c.Password), nil
}

func NewPasswordCredentials(username, password string) *PasswordCredentials {
	return &PasswordCredentials{
		Username: username,
		Password: password,
	}
}

type ApiTokenCredentials struct {
	Username string
	ApiToken string
}

func (c *ApiTokenCredentials) Retrieve(ctx context.Context) (string, error) {
	return buildBasicAuth(c.Username, c.ApiToken), nil
}

func NewApiTokenCredentials(username, apiToken string) *ApiTokenCredentials {
	return &ApiTokenCredentials{
		Username: username,
		ApiToken: apiToken,
	}
}

// Helper functions

func buildBasicAuth(username, secret string) string {
	auth := username + ":" + secret
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
}
