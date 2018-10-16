package skytap

import (
	"fmt"
)

const (
	DefaultBaseURL   = "https://cloud.skytap.com/"
	DefaultUserAgent = "skytap-sdk-go/" + version
)

type Settings struct {
	baseUrl   string
	userAgent string

	credentials CredentialsProvider
}

func (s *Settings) Validate() error {
	if s.baseUrl == "" {
		return fmt.Errorf("The base URL must be provided")
	}
	if s.userAgent == "" {
		return fmt.Errorf("The user agent must be provided")
	}
	if s.credentials == nil {
		return fmt.Errorf("The credential provider must be provided")
	}

	return nil
}

func NewDefaultSettings(clientSettings ...ClientSetting) Settings {
	settings := Settings{
		baseUrl:     DefaultBaseURL,
		userAgent:   DefaultUserAgent,
		credentials: NewNoOpCredentials(),
	}

	// Apply any custom settings
	for _, c := range clientSettings {
		c.Apply(&settings)
	}

	return settings
}

type ClientSetting interface {
	Apply(*Settings)
}

type withBaseUrl string
type withUserAgent string
type withCredentialsProvider struct{ cp CredentialsProvider }

func (w withBaseUrl) Apply(s *Settings) {
	s.baseUrl = string(w)
}

func (w withUserAgent) Apply(s *Settings) {
	s.userAgent = string(w)
}

func (w withCredentialsProvider) Apply(s *Settings) {
	s.credentials = w.cp
}

func WithBaseUrl(BaseUrl string) ClientSetting {
	return withBaseUrl(BaseUrl)
}

func WithUserAgent(BaseUrl string) ClientSetting {
	return withUserAgent(BaseUrl)
}

func WithCredentialsProvider(credentialsProvider CredentialsProvider) ClientSetting {
	return withCredentialsProvider{credentialsProvider}
}
