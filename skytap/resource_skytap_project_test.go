package skytap

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
)

func init() {
	resource.AddTestSweepers("skytap_project", &resource.Sweeper{
		Name: "skytap_project",
		F:    testSweepSkytapProject,
	})
}

func testSweepSkytapProject(region string) error {
	meta, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	client := meta.projectsClient
	ctx := meta.StopContext

	log.Printf("[INFO] Retrieving list of project")
	projects, err := client.List(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving list of project: %v", err)
	}

	for _, p := range projects.Value {
		if shouldSweepAcceptanceTestResource(*p.Name) {
			log.Printf("destroying project %s", *p.Name)
			if err := client.Delete(ctx, *p.Id); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccSkytapProject_Basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapProjectConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapProjectExists("skytap_project.foo"),
					resource.TestCheckResourceAttr("skytap_project.foo", "name", fmt.Sprintf("tftest-project-%d", rInt)),
					resource.TestCheckResourceAttrSet("skytap_project.foo", "summary"),
					resource.TestCheckResourceAttr("skytap_project.foo", "auto_add_role_name", ""),
					resource.TestCheckResourceAttr("skytap_project.foo", "show_project_members", "true"),
				),
			},
		},
	})
}

// Verifies the Project exists
func testAccCheckSkytapProjectExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %q", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Project ID is set")
		}

		// retrieve the connection established in Provider configuration
		client := testAccProvider.Meta().(*SkytapClient).projectsClient
		ctx := testAccProvider.Meta().(*SkytapClient).StopContext

		// Retrieve our project by referencing it's state ID for API lookup
		_, err := client.Get(ctx, rs.Primary.ID)
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				return errors.Errorf("project (%s) was not found - does not exist", rs.Primary.ID)
			}

			return fmt.Errorf("error retrieving project (%s): %v", rs.Primary.ID, err)
		}

		return nil
	}
}

// Verifies the Project has been destroyed
func testAccCheckSkytapProjectDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).projectsClient
	ctx := testAccProvider.Meta().(*SkytapClient).StopContext

	// loop through the resources in state, verifying each project
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "skytap_project" {
			continue
		}

		// Retrieve our project by referencing it's state ID for API lookup
		_, err := client.Get(ctx, rs.Primary.ID)
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				return nil
			}

			return fmt.Errorf("error waiting for project (%s) to be destroyed: %s", rs.Primary.ID, err)
		}

		return fmt.Errorf("project still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccSkytapProjectConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "skytap_project" "foo" {
	name = "tftest-project-%d"
	summary = "This is a project created by the skytap terraform provider acceptance test"
}`, rInt)
}
