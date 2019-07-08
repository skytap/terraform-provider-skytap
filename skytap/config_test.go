package skytap

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgent(t *testing.T) {
	config := &Config{
		APIToken: "abc123",
	}

	userAgent, err := getUserAgent()

	matched, err := regexp.Match(`terraform-provider-skytap/\d+\.\d+\.\d+`, []byte(userAgent))

	assert.True(t, matched, "Not matched")

	client, err := config.createClient()

	assert.NoError(t, err)
	assert.Equal(t, userAgent, client.UserAgent)
}
