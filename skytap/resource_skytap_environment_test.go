package skytap

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
	"log"
	"testing"
)

func init() {
	resource.AddTestSweepers("skytap_environment", &resource.Sweeper{
		Name: "skytap_environment",
		F:    testSweepSkytapEnvironment,
	})
}

func testSweepSkytapEnvironment(region string) error {
	meta, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	client := meta.environmentsClient
	ctx := meta.StopContext

	log.Printf("[INFO] Retrieving list of environments")
	environments, err := client.List(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving list of environments: %v", err)
	}

	for _, e := range environments.Value {
		if shouldSweepAcceptanceTestResource(*e.Name) {
			log.Printf("destroying environment %s", *e.Name)
			if err := client.Delete(ctx, *e.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccSkytapEnvironment_Basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "name", fmt.Sprintf("tftest-environment-%d", rInt)),
					resource.TestCheckResourceAttrSet("skytap_environment.foo", "description"),
					resource.TestCheckResourceAttrSet("skytap_environment.foo", "template_id"),
					resource.TestCheckNoResourceAttr("skytap_environment.foo", "project_id"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "outbound_traffic", "false"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "routable", "false"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "suspend_on_idle", "0"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "suspend_at_time", ""),
					resource.TestCheckResourceAttr("skytap_environment.foo", "shutdown_on_idle", "0"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "shutdown_at_time", ""),
				),
			},
		},
	})
}

// Verifies the Environment exists
func testAccCheckSkytapEnvironmentExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %q", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Environment ID is set")
		}

		// retrieve the connection established in Provider configuration
		client := testAccProvider.Meta().(*SkytapClient).environmentsClient
		ctx := testAccProvider.Meta().(*SkytapClient).StopContext

		// Retrieve our environment by referencing it's state ID for API lookup
		_, err := client.Get(ctx, rs.Primary.ID)
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				return errors.Errorf("environment (%s) was not found - does not exist", rs.Primary.ID)
			}

			return fmt.Errorf("error retrieving environment (%s): %v", rs.Primary.ID, err)
		}

		return nil
	}
}

// Verifies the Environment has been destroyed
func testAccCheckSkytapEnvironmentDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).environmentsClient
	ctx := testAccProvider.Meta().(*SkytapClient).StopContext

	// loop through the resources in state, verifying each environment
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "skytap_environment" {
			continue
		}

		// Retrieve our environment by referencing it's state ID for API lookup
		_, err := client.Get(ctx, rs.Primary.ID)
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				return nil
			}

			return fmt.Errorf("error waiting for environment (%s) to be destroyed: %s", rs.Primary.ID, err)
		}

		return fmt.Errorf("environment still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccSkytapEnvironmentConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "skytap_environment" "foo" {
	template_id = "1452333"
	name = "tftest-environment-%d"
	description = "This is an environment created by the skytap terraform provider acceptance test"
}`, rInt)
}
