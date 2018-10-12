package skytap

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceSkytapProject_Basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSkytapProjectConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapProjectExists("data.skytap_project.foo"),
					resource.TestCheckResourceAttr("data.skytap_project.foo", "name", fmt.Sprintf("tftest-project-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.skytap_project.foo", "summary"),
					resource.TestCheckResourceAttr("data.skytap_project.foo", "auto_add_role_name", ""),
					resource.TestCheckResourceAttr("data.skytap_project.foo", "show_project_members", "true"),
				),
			},
		},
	})
}

func testAccDataSourceSkytapProjectConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "skytap_project" "foo" {
	name = "tftest-project-%d"
	summary = "This is a project created by the skytap terraform provider acceptance test"
}

data "skytap_project" "foo" {
	name = "${skytap_project.foo.name}"
}`, rInt)
}
