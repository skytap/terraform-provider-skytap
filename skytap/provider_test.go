package skytap

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"skytap": func() (*schema.Provider, error) { return testAccProvider, nil },
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	required := []string{
		"SKYTAP_USERNAME",
		"SKYTAP_API_TOKEN",
	}

	for _, prop := range required {
		if os.Getenv(prop) == "" {
			t.Fatalf("%s must be set for acceptance test", prop)
		}
	}
}
