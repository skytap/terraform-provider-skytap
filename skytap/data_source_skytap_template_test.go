package skytap

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceSkytapTemplate_Basic(t *testing.T) {
	//t.Parallel()

	if setEnv(t, "SKYTAP_TEMPLATE_NAME", "Advanced Import Appliance on Ubuntu 18.04.1") {
		defer unsetEnv("SKYTAP_TEMPLATE_NAME")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSkytapTemplateConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.skytap_template.foo", "id"),
				),
			},
		},
	})
}

func testAccDataSourceSkytapTemplateConfig_basic() string {
	return fmt.Sprintf(`
data "skytap_template" "foo" {
	name = "%s"
}

output "id" {
  value = "${data.skytap_template.foo.id}"
}`, os.Getenv("SKYTAP_TEMPLATE_NAME"))
}

func TestAccDataSourceSkytapTemplate_RegexMostRecent(t *testing.T) {
	//t.Parallel()

	if setEnv(t, "SKYTAP_TEMPLATE_NAME", "Advanced Import Appliance on Ubuntu 18.04.1") {
		defer unsetEnv("SKYTAP_TEMPLATE_NAME")
	}
	if setEnv(t, "SKYTAP_TEMPLATE_NAME_PARTIAL", "Appliance on Ubuntu 18.04") {
		defer unsetEnv("SKYTAP_TEMPLATE_NAME_PARTIAL")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSkytapTemplateConfig_regexMostRecent(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.skytap_template.foo", "id"),
					resource.TestCheckResourceAttr("data.skytap_template.foo", "name", os.Getenv("SKYTAP_TEMPLATE_NAME")),
				),
			},
		},
	})
}

func testAccDataSourceSkytapTemplateConfig_regexMostRecent() string {
	return fmt.Sprintf(`
data "skytap_template" "foo" {
	name = "%s"
    most_recent = true
}

output "id" {
  value = "${data.skytap_template.foo.id}"
}`, os.Getenv("SKYTAP_TEMPLATE_NAME_PARTIAL"))
}
