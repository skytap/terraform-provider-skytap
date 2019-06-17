package skytap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgent(t *testing.T) {
	config := &Config{
		APIToken: "abc123",
	}

	userAgent, err := getUserAgent()
	assert.True(t, len(userAgent) > 3)

	client, err := config.createClient()

	assert.NoError(t, err)
	assert.Equal(t, userAgent, client.UserAgent)
}
