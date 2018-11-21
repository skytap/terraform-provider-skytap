package skytap

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceSkytapTemplate_Basic(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_NAME") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_NAME required to run skytap_template_datasource acceptance tests. Setting: SKYTAP_TEMPLATE_NAME=Advanced Import Appliance on Ubuntu 18.04.1")
		os.Setenv("SKYTAP_TEMPLATE_NAME", "Advanced Import Appliance on Ubuntu 18.04.1")
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

	if os.Getenv("SKYTAP_TEMPLATE_NAME") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_NAME required to run skytap_template_datasource acceptance tests. Setting: SKYTAP_TEMPLATE_NAME=Advanced Import Appliance on Ubuntu 18.04.1")
		os.Setenv("SKYTAP_TEMPLATE_NAME", "Advanced Import Appliance on Ubuntu 18.04.1")
	}
	if os.Getenv("SKYTAP_TEMPLATE_NAME_PARTIAL") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_NAME_PARTIAL required to run skytap_template_datasource acceptance tests. Setting: SKYTAP_TEMPLATE_NAME_PARTIAL=18.04")
		os.Setenv("SKYTAP_TEMPLATE_NAME_PARTIAL", "18.04")
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
