package skytap

import (
	"fmt"
)

const (
	// DefaultBaseURL is the base URL if not explicitly set
	DefaultBaseURL = "https://cloud.skytap.com/"
	// DefaultUserAgent is the default user agent if not explicitly set
	DefaultUserAgent = "skytap-sdk-go/" + version
)

// Settings holds the base URL, user agent and credential data
type Settings struct {
	baseURL   string
	userAgent string

	credentials   CredentialsProvider
	maxRetryCount int
}

// Validate the settings
func (s *Settings) Validate() error {
	if s.baseURL == "" {
		return fmt.Errorf("the base URL must be provided")
	}
	if s.userAgent == "" {
		return fmt.Errorf("the user agent must be provided")
	}
	if s.credentials == nil {
		return fmt.Errorf("the credential provider must be provided")
	}

	return nil
}

// NewDefaultSettings creates a new Settings based upon the input clientSettings
func NewDefaultSettings(clientSettings ...ClientSetting) Settings {
	settings := Settings{
		baseURL:     DefaultBaseURL,
		userAgent:   DefaultUserAgent,
		credentials: NewNoOpCredentials(),
	}

	// Apply any custom settings
	for _, c := range clientSettings {
		c.Apply(&settings)
	}

	return settings
}

// ClientSetting abstracts an individual setting
type ClientSetting interface {
	Apply(*Settings)
}

type withBaseURL string
type withUserAgent string
type withMaxRetryCount int
type withCredentialsProvider struct{ cp CredentialsProvider }

func (w withBaseURL) Apply(s *Settings) {
	s.baseURL = string(w)
}

func (w withUserAgent) Apply(s *Settings) {
	s.userAgent = string(w)
}

func (w withMaxRetryCount) Apply(s *Settings) {
	s.maxRetryCount = int(w)
}

func (w withCredentialsProvider) Apply(s *Settings) {
	s.credentials = w.cp
}

// WithBaseURL accepts a base URL
func WithBaseURL(BaseURL string) ClientSetting {
	return withBaseURL(BaseURL)
}

// WithUserAgent accepts a user agent
func WithUserAgent(UserAgent string) ClientSetting {
	return withUserAgent(UserAgent)
}

func WithMaxRetryCount(retryCount int) ClientSetting {
	return withMaxRetryCount(retryCount)
}

// WithCredentialsProvider accepts an abstracted set of credentials
func WithCredentialsProvider(credentialsProvider CredentialsProvider) ClientSetting {
	return withCredentialsProvider{credentialsProvider}
}
