package skytap

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/opencredo/skytap-sdk-go-internal"
	"strings"
	"testing"
)

func TestAccCreateProject(t *testing.T) {

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateProject(rName),
				Check:  testAccCheckProjectExists,
			},
		},
	})
}

func testAccCheckProjectExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*skytap.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "skytap_project" {
			continue
		}
		// Retrieve our widget by referencing it's state ID for API lookup
		_, err := client.ReadProject(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCreateProject(rName string) string {
	config := `
    resource "skytap_project" "terraform-project-acc-test" {
  		name = %q
  		summary = "This is a project created by the skytap terraform provider acceptance test"
	}`
	return fmt.Sprintf(config, rName)
}

// testAccCheckProjectResourceDestroy verifies the Project
// has been destroyed
func testAccCheckProjectResourceDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testAccProvider.Meta().(*skytap.Client)

	// loop through the resources in state, verifying each widget
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "skytap_project" {
			continue
		}

		// Retrieve our widget by referencing it's state ID for API lookup
		response, err := conn.ReadProject(context.Background(), rs.Primary.ID)
		if err == nil {
			if response.Id == rs.Primary.ID {
				return fmt.Errorf("project (%s) still exists", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the widget is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}
