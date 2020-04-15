package skytap

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
	"log"
	"regexp"
	"strconv"
	"testing"
)

func init() {
	resource.AddTestSweepers("skytap_label_category", &resource.Sweeper{
		Name: "skytap_label_category",
		F:    testSweepSkytapLabelCategory,
	})
}

func TestAccSkytapLabelCategory_Basic(t *testing.T) {
	/** This test does not randomize as there are a total of 200 label category per account
	  including label categories that have been deleted. If the test randomize the input it will soon reach
	  an account limit
	*/
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapLabelCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapLabelCategory_basic("tftest-label", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapLabelCategoryExists("skytap_label_category.env_category"),
					resource.TestCheckResourceAttr("skytap_label_category.env_category", "name",
						fmt.Sprintf("tftest-label")),
					resource.TestCheckResourceAttrSet("skytap_label_category.env_category", "single_value"),
					resource.TestCheckResourceAttr("skytap_label_category.env_category", "single_value", "true"),
				),
			},
		},
	})
}

func TestAccSkytapLabelCategory_Update(t *testing.T) {
	/** This test does not randomize as there are a total of 200 label category per account
	  including label categories that have been deleted. If the test randomize the input it will soon reach
	  an account limit
	*/
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapLabelCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapLabelCategory_basic("tftest-label", true),
				Check:  testAccCheckSkytapLabelCategoryExists("skytap_label_category.env_category"),
			},
			{
				ExpectNonEmptyPlan: true,
				Config:             testAccSkytapLabelCategory_basic("tftest-label", false),
				Check:              testAccCheckSkytapLabelCategoryExists("skytap_label_category.env_category"),
				ExpectError:        regexp.MustCompile(`can not be created with this single value property as it is recreated from a existing label category`),
			},
		},
	})
}

func TestAccSkytapLabelCategory_MultiValueBasic(t *testing.T) {
	/** This test does not randomize as there are a total of 200 label category per account
	  including label categories that have been deleted. If the test randomize the input it will soon reach
	  an account limit
	*/
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapLabelCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapLabelCategory_basic("tftest-label-multi", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapLabelCategoryExists("skytap_label_category.env_category"),
					resource.TestCheckResourceAttr("skytap_label_category.env_category", "name",
						fmt.Sprintf("tftest-label-multi")),
					resource.TestCheckResourceAttrSet("skytap_label_category.env_category", "single_value"),
					resource.TestCheckResourceAttr("skytap_label_category.env_category", "single_value", "false"),
				),
			},
		},
	})
}

func TestAccSkytapLabelCategory_Duplicated(t *testing.T) {
	labelCategoryDuplicated := `
		resource skytap_label_category "duplabel1" {
		  name = "tftest-dup"
		  single_value = true
		}
		
		resource skytap_label_category "duplabel2" {
		  // making sure the first category is created before trying to create the second
		  // to avoid flakiness.
          depends_on = [skytap_label_category.duplabel1]
		  name = "tftest-dup"
		  single_value = true
		}
	`
	error, _ := regexp.Compile(".* Validation failed: Name has already been taken")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapLabelCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      labelCategoryDuplicated,
				ExpectError: error,
			},
		},
	})
}

func testAccSkytapLabelCategory_basic(labelCategoryName string, singleValue bool) string {
	return fmt.Sprintf(`
      resource "skytap_label_category" "env_category" {
	    name =  "%s"
		single_value = %t
      }`, labelCategoryName, singleValue)
}

func testAccCheckSkytapLabelCategoryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %q", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no  label category set ")
		}

		// retrieve the connection established in Provider configuration
		client := testAccProvider.Meta().(*SkytapClient).labelCategoryClient
		ctx := testAccProvider.Meta().(*SkytapClient).StopContext

		// Retrieve our label category by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("label category (%s) is not an integer: %v", rs.Primary.ID, err)
		}

		labelCategory, err := client.Get(ctx, id)
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				return fmt.Errorf("label category (%d) was not found - does not exist", id)
			}
			return fmt.Errorf("error retrieving label category (%d): %v", id, err)
		}

		if !*labelCategory.Enabled {
			return fmt.Errorf("error category %d is dissabled", id)
		}
		return nil
	}
}

func testAccCheckSkytapLabelCategoryDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).labelCategoryClient
	ctx := testAccProvider.Meta().(*SkytapClient).StopContext

	// loop through the resources in state, verifying each label category
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "skytap_label_category" {
			continue
		}

		// Retrieve our label category by referencing it's state ID for API lookup
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("label category (%s) is not an integer: %v", rs.Primary.ID, err)
		}

		labelCategory, err := client.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("error waiting for label category (%d) to be destroyed: %s", id, err)
		} else if *labelCategory.Enabled {
			return fmt.Errorf("label category still exists: %d", id)
		}
	}

	return nil
}

func testSweepSkytapLabelCategory(region string) error {
	meta, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	client := meta.labelCategoryClient
	ctx := meta.StopContext

	log.Printf("[INFO] Retrieving list of label category")
	labelCategories, err := client.List(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving list label categories: %v", err)
	}

	for _, l := range labelCategories {
		if shouldSweepAcceptanceTestResource(*l.Name) {
			log.Printf("destroying label category %s", *l.Name)
			if err := client.Delete(ctx, *l.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
