package skytap

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const PREFIX = "tftest"

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedClientForRegion(region string) (*SkytapClient, error) {
	username := os.Getenv("SKYTAP_USERNAME")
	apiToken := os.Getenv("SKYTAP_API_TOKEN")

	if username == "" || apiToken == "" {
		return nil, fmt.Errorf("SKYTAP_USERNAME and SKYTAP_API_TOKEN must be set for acceptance tests")
	}

	config := &Config{
		Username: username,
		APIToken: apiToken,
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
	return shouldSweepAcceptanceTestResourceWithPrefix(name, PREFIX)
}

func shouldSweepAcceptanceTestResourceWithPrefix(name string, prefix string) bool {
	loweredName := strings.ToLower(name)

	if !strings.HasPrefix(loweredName, prefix) {
		log.Printf("ignoring Resource %q as it doesn't start with `%s`", name, prefix)
		return false
	}

	return true
}
