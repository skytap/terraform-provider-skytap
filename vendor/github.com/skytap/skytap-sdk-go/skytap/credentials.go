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

// NoOpCredentials is used when no credentials are required
type NoOpCredentials struct{}

// Retrieve the credentials
func (c *NoOpCredentials) Retrieve(ctx context.Context) (string, error) {
	return "", nil
}

// NewNoOpCredentials creates a new no op credentials instance
func NewNoOpCredentials() *NoOpCredentials {
	return &NoOpCredentials{}
}

// APITokenCredentials is ued when the credentials used are the username and api token data
type APITokenCredentials struct {
	Username string
	APIToken string
}

// Retrieve the username and api token data
func (c *APITokenCredentials) Retrieve(ctx context.Context) (string, error) {
	return buildBasicAuth(c.Username, c.APIToken), nil
}

// NewAPITokenCredentials creates a new username and api token instance
func NewAPITokenCredentials(username, apiToken string) *APITokenCredentials {
	return &APITokenCredentials{
		Username: username,
		APIToken: apiToken,
	}
}

// Helper functions
func buildBasicAuth(username, secret string) string {
	auth := username + ":" + secret
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
}
