package skytap

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/opencredo/skytap-sdk-go-internal"
	"github.com/opencredo/skytap-sdk-go-internal/options"
	"os"
	"testing"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"skytap": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	required := []string{"SKYTAP_USER", "SKYTAP_ACCESS_TOKEN"}

	for _, prop := range required {
		if os.Getenv(prop) == "" {
			t.Fatalf("%s must be set for acceptance test", prop)
		}
	}

	_, err := skytap.NewClient(context.Background(),
		options.WithUser(os.Getenv("SKYTAP_USER")),
		options.WithAPIToken(os.Getenv("SKYTAP_ACCESS_TOKEN")),
		options.WithScheme("https"),
		options.WithHost("cloud.skytap.com"),
	)
	if err != nil {
		t.Fatal(fmt.Sprintf("%+v", err))
	}
}
