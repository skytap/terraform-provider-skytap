package skytap

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSkytapProject_Basic(t *testing.T) {
	templateID, _, _ := setupEnvironment()
	uniqueSuffix := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSkytapProjectConfig_basic(templateID, uniqueSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapProjectExists("data.skytap_project.bar"),
					resource.TestCheckResourceAttr("data.skytap_project.bar", "name", fmt.Sprintf("tftest-project-data-%d", uniqueSuffix)),
					resource.TestCheckResourceAttrSet("data.skytap_project.bar", "summary"),
					resource.TestCheckResourceAttr("data.skytap_project.bar", "auto_add_role_name", ""),
					resource.TestCheckResourceAttr("data.skytap_project.bar", "show_project_members", "true"),
					resource.TestCheckTypeSetElemAttrPair("data.skytap_project.bar", "environment_ids.*", "skytap_environment.foo", "id"),
				),
			},
		},
	})
}

func testAccDataSourceSkytapProjectConfig_basic(envTemplateID string, uniqueSuffix int) string {
	return fmt.Sprintf(`
resource "skytap_project" "foo" {
	name = "tftest-project-data-%d"
	summary = "This is a project created by the skytap terraform provider acceptance test"
	environment_ids = [skytap_environment.foo.id]
}

resource "skytap_environment" "foo" {
	template_id = "%s"
	name 		= "%s-environment-%d"
	description = "This is an environment to support a skytap project terraform provider acceptance test"
}

data "skytap_project" "bar" {
	name = skytap_project.foo.name
}`, uniqueSuffix, envTemplateID, vmEnvironmentPrefix, uniqueSuffix)
}
