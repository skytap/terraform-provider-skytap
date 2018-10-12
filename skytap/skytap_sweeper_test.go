package skytap

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/pkg/errors"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedClientForRegion(region string) (*SkytapClient, error) {
	username := os.Getenv("SKYTAP_USERNAME")
	apiToken := os.Getenv("SKYTAP_API_TOKEN")

	if username == "" || apiToken == "" {
		return nil, errors.Errorf("SKYTAP_USERNAME and SKYTAP_API_TOKEN must be set for acceptance tests")
	}

	config := &Config{
		Username: username,
		ApiToken: apiToken,
	}

	// configures a default client for the region, using the above env vars
	client, err := config.Client()
	if err != nil {
		return nil, err
	}

	client.StopContext = context.Background()

	return client, nil
}

func shouldSweepAcceptanceTestResource(name string) bool {
	loweredName := strings.ToLower(name)

	if !strings.HasPrefix(loweredName, "tftest") {
		log.Printf("Ignoring Resource %q as it doesn't start with `tftest`", name)
		return false
	}

	return true
}
