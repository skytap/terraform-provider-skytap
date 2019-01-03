package skytap

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
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
	//t.Parallel()

	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1473407")
	uniqueSuffix := acctest.RandInt()
	var environment skytap.Environment

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_basic(uniqueSuffix, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					resource.TestCheckResourceAttr("skytap_environment.foo", "name", fmt.Sprintf("tftest-environment-%d", uniqueSuffix)),
					resource.TestCheckResourceAttr("skytap_environment.foo", "description", "This is an environment created by the skytap terraform provider acceptance test"),
					resource.TestCheckResourceAttrSet("skytap_environment.foo", "template_id"),
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

func TestAccSkytapEnvironment_UpdateTemplate(t *testing.T) {
	//t.Parallel()

	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1473407")
	template2ID := utils.GetEnv("SKYTAP_TEMPLATE_ID2", "1473347")
	rInt := acctest.RandInt()
	var environment skytap.Environment

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_basic(rInt, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
				),
			},
			{
				Config: testAccSkytapEnvironmentConfig_basic(rInt, template2ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentAfterTemplateChanged("skytap_environment.foo", &environment),
				),
			},
		},
	})
}

// Verifies the Environment exists
func testAccCheckSkytapEnvironmentExists(name string, environment *skytap.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, err := getResource(s, name)
		if err != nil {
			return err
		}

		// Get the environment
		env, err := getEnvironment(rs)
		if err != nil {
			return err
		}

		*environment = *env
		log.Printf("[DEBUG] environment (%s)\n", *environment.ID)

		return nil
	}
}

// Verifies the Environment is brand new
func testAccCheckSkytapEnvironmentAfterTemplateChanged(name string, environmentOld *skytap.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, err := getResource(s, name)
		if err != nil {
			return err
		}

		// Get the environment
		environmentNew, err := getEnvironment(rs)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] old environment (%s) and new environment (%s)\n", *environmentOld.ID, *environmentNew.ID)
		if *environmentOld.ID == *environmentNew.ID {
			return fmt.Errorf("the old environment (%s) has been updated", rs.Primary.ID)
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

func testAccSkytapEnvironmentConfig_basic(uniqueSuffix int, templateID string) string {
	return fmt.Sprintf(`
      resource "skytap_environment" "foo" {
	    template_id = "%s"
	    name = "tftest-environment-%d"
	    description = "This is an environment created by the skytap terraform provider acceptance test"
      }`, templateID, uniqueSuffix)
}

func getEnvironment(rs *terraform.ResourceState) (*skytap.Environment, error) {
	var err error
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).environmentsClient
	ctx := testAccProvider.Meta().(*SkytapClient).StopContext

	// Retrieve our environment by referencing it's state ID for API lookup
	environment, errClient := client.Get(ctx, rs.Primary.ID)
	if errClient != nil {
		if utils.ResponseErrorIsNotFound(err) {
			err = fmt.Errorf("environment (%s) was not found - does not exist", rs.Primary.ID)
		}

		err = fmt.Errorf("error retrieving environment (%s): %v", rs.Primary.ID, err)
	}
	return environment, err
}

func getResource(s *terraform.State, name string) (*terraform.ResourceState, error) {
	rs, ok := s.RootModule().Resources[name]
	var err error
	if !ok {
		err = fmt.Errorf("not found: %q", name)
	}

	if ok && rs.Primary.ID == "" {
		err = fmt.Errorf("no resource ID is set")
	}
	return rs, err
}
