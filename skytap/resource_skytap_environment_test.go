package skytap

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	ctx := context.TODO()

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
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	var environment skytap.Environment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_basic(uniqueSuffix, templateID, `["integration_test"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					resource.TestCheckResourceAttr("skytap_environment.foo", "name", fmt.Sprintf("tftest-environment-%d", uniqueSuffix)),
					resource.TestCheckResourceAttr("skytap_environment.foo", "description", "This is an environment created by the skytap terraform provider acceptance test"),
					resource.TestCheckResourceAttrSet("skytap_environment.foo", "template_id"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "disable_internet", "false"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "outbound_traffic", "false"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "routable", "false"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "suspend_on_idle", "0"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "suspend_at_time", ""),
					resource.TestCheckResourceAttr("skytap_environment.foo", "shutdown_on_idle", "0"),
					resource.TestCheckResourceAttr("skytap_environment.foo", "shutdown_at_time", ""),
					resource.TestCheckResourceAttr("skytap_environment.foo", "tags.#", "1"),
				),
			},
		},
	})
}

func TestAccSkytapEnvironment_Update(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	var environment skytap.Environment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_advanced(uniqueSuffix, templateID, `["integration_test"]`, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
				),
			},
			{
				PreConfig: func() {
					pause(MINUTES)()
					// This should only be used until stopping the environment for disableInternet/routable update is implemented
					err := suspendEnvironment(&environment)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccSkytapEnvironmentConfig_advanced(uniqueSuffix, templateID, `["integration_test"]`, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
				),
			},
		},
	})
}

func TestAccSkytapEnvironment_UpdateTemplate(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	template2ID := utils.GetEnv("SKYTAP_TEMPLATE_ID2", "1877151")
	rInt := acctest.RandInt()
	var environment skytap.Environment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_basic(rInt, templateID, `[]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
				),
			},
			{
				Config: testAccSkytapEnvironmentConfig_basic(rInt, template2ID, `[]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentAfterTemplateChanged("skytap_environment.foo", &environment),
				),
			},
		},
	})
}

func TestAccSkytapEnvironment_UpdateTags(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()

	var environment skytap.Environment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_basic(uniqueSuffix, templateID, `["foo", "bar"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("skytap_environment.foo", "tags.#", "2"),
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					testAccCheckSkytapEnvironmentContainsTag(&environment, "foo"),
					testAccCheckSkytapEnvironmentContainsTag(&environment, "bar"),
				),
			},
			{
				Config: testAccSkytapEnvironmentConfig_basic(uniqueSuffix, templateID, `["foozzz", "Bar", "foobar"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("skytap_environment.foo", "tags.#", "3"),
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					testAccCheckSkytapEnvironmentContainsTag(&environment, "foozzz"),
					testAccCheckSkytapEnvironmentContainsTag(&environment, "Bar"),
					testAccCheckSkytapEnvironmentContainsTag(&environment, "foobar"),
				),
			},
		},
	})
}

func TestAccSkytapEnvironment_UserData(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	var environment skytap.Environment

	userData := `<<EOF
				cat /proc/cpu_info
				EOF
				`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				PreventDiskCleanup: true,
				Config:             testAccSkytapEnvironmentConfig_UserData(uniqueSuffix, templateID, userData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					resource.TestCheckResourceAttrSet("skytap_environment.foo", "user_data"),
				),
			},
		},
	})
}

func TestAccSkytapEnvironment_UserDataUpdate(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	var environment skytap.Environment

	userData := `<<EOF
				cat /proc/cpu_info
				EOF
				`
	userDataRe := regexp.MustCompile(`\s*cat /proc/cpu_info\n`)
	userDataUpdate := `<<EOF
				cat /proc/cpu_info  > /temp/acc
				EOF
				`
	userDataUpdateRe := regexp.MustCompile(`\s*cat /proc/cpu_info\s*>\s*/temp/acc\n`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapEnvironmentConfig_UserData(uniqueSuffix, templateID, userData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					resource.TestCheckResourceAttrSet("skytap_environment.foo", "user_data"),
					resource.TestMatchResourceAttr("skytap_environment.foo", "user_data", userDataRe),
				),
			},
			{
				PreventDiskCleanup: true,
				Config:             testAccSkytapEnvironmentConfig_UserData(uniqueSuffix, templateID, userDataUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					resource.TestCheckResourceAttrSet("skytap_environment.foo", "user_data"),
					resource.TestMatchResourceAttr("skytap_environment.foo", "user_data", userDataUpdateRe),
				),
			},
		},
	})
}

func labelRequirements(uniqueSuffix int) string {
	return fmt.Sprintf(`
		resource skytap_label_category "environment_label" {
			name = "tftest-Environment-%d"
			single_value = true
		}

		resource skytap_label_category "owners_label" {
			name = "tftest-Owners-%d"
			single_value = false
		}
	`, uniqueSuffix, uniqueSuffix)
}

func TestAccSkytapEnvironment_Labels(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	var environment skytap.Environment

	labels := `
		label {
			category = skytap_label_category.environment_label.name
			value = "Prod"
		}
		label {
			category = skytap_label_category.owners_label.name
			value = "Finance"
		}
		label {
			category = skytap_label_category.owners_label.name
			value = "Accounting"
		}
	`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				PreventDiskCleanup: true,
				Config:             testAccSkytapEnvironmentConfigBlock(uniqueSuffix, templateID, labelRequirements(uniqueSuffix), labels),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
				),
			},
		},
	})
}

func TestAccSkytapEnvironment_LabelsUpdate(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	var environment skytap.Environment

	labels := `
		label {
			category = skytap_label_category.environment_label.name
			value = "Prod"
		}
		label {
			category = skytap_label_category.owners_label.name
			value = "Finance"
		}
		label {
			category = skytap_label_category.owners_label.name
			value = "Accounting"
		}
	`

	labelsUpdated := `
		label {
			category = skytap_label_category.environment_label.name
			value = "UAT"
		}
		label {
			category = skytap_label_category.owners_label.name
			value = "Accounting"
		}
	`

	labelEnv := fmt.Sprintf("tftest-Environment-%d", uniqueSuffix)
	labelOwners := fmt.Sprintf("tftest-Owners-%d", uniqueSuffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				PreventDiskCleanup: true,
				Config:             testAccSkytapEnvironmentConfigBlock(uniqueSuffix, templateID, labelRequirements(uniqueSuffix), labels),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					testAccCheckSkytapEnvironmentContainsLabel(&environment, labelEnv, "Prod"),
					testAccCheckSkytapEnvironmentContainsLabel(&environment, labelOwners, "Finance"),
					testAccCheckSkytapEnvironmentContainsLabel(&environment, labelOwners, "Accounting"),
				),
			},
			{
				PreventDiskCleanup: true,
				Config:             testAccSkytapEnvironmentConfigBlock(uniqueSuffix, templateID, labelRequirements(uniqueSuffix), labelsUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapEnvironmentExists("skytap_environment.foo", &environment),
					testAccCheckSkytapEnvironmentContainsLabel(&environment, labelEnv, "UAT"),
					testAccCheckSkytapEnvironmentContainsLabel(&environment, labelOwners, "Accounting"),
				),
			},
		},
	})
}

func TestAccSkytapEnvironment_DisableInternetConflict(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	resourceWithConflict := fmt.Sprintf(`
		resource skytap_environment "env" {
		  name = "tftest-environment-%d"
          description = "This is an environment created by the skytap terraform provider acceptance test"
		  template_id = %s
		  disable_internet = true
          outbound_traffic = true
		}
	`, uniqueSuffix, templateID)
	expectedError := regexp.MustCompile("\"disable_internet\": conflicts with outbound_traffic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapLabelCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      resourceWithConflict,
				ExpectError: expectedError,
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

// Suspends the Environment exists
func suspendEnvironment(env *skytap.Environment) error {
	client := testAccProvider.Meta().(*SkytapClient).environmentsClient

	ctx := context.TODO()

	stopped := skytap.EnvironmentRunstateSuspended
	_, err := client.Update(ctx, *env.ID, &skytap.UpdateEnvironmentRequest{Runstate: &stopped})
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			err = fmt.Errorf("environment (%s) was not found - does not exist", *env.ID)
		}

		err = fmt.Errorf("error retrieving environment (%s): %v", *env.ID, err)
	}
	return nil
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

// Verifies the Environment have a specific tag
func testAccCheckSkytapEnvironmentContainsTag(enviroment *skytap.Environment, tag string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, t := range enviroment.Tags {
			// tags are stored in lowercase
			if strings.ToLower(tag) == *t.Value {
				return nil
			}
		}
		return fmt.Errorf("tag not found: %s", tag)
	}
}

// Verifies the Environment have a specific label
func testAccCheckSkytapEnvironmentContainsLabel(enviroment *skytap.Environment, category string, label string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, t := range enviroment.Labels {
			if strings.ToLower(category) == strings.ToLower(*t.LabelCategory) &&
				strings.ToLower(label) == strings.ToLower(*t.Value) {
				return nil
			}
		}
		return fmt.Errorf("label  (%s : %s) not found", category, label)
	}
}

// Verifies the Environment has been destroyed
func testAccCheckSkytapEnvironmentDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).environmentsClient
	ctx := context.TODO()

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

func testAccSkytapEnvironmentConfig_basic(uniqueSuffix int, templateID string, tags string) string {
	return fmt.Sprintf(`
      resource "skytap_environment" "foo" {
	    template_id = "%s"
		tags = %s
	    name = "tftest-environment-%d"
	    description = "This is an environment created by the skytap terraform provider acceptance test"
      }`, templateID, tags, uniqueSuffix)
}

func testAccSkytapEnvironmentConfig_advanced(uniqueSuffix int, templateID string, tags string, disableInternet bool, routable bool) string {
	return fmt.Sprintf(`
      resource "skytap_environment" "foo" {
	    template_id = "%s"
		tags = %s
	    name = "tftest-environment-%d"
	    description = "This is an environment created by the skytap terraform provider acceptance test"
		disable_internet = %t
		routable = %t
      }`, templateID, tags, uniqueSuffix, disableInternet, routable)
}

func testAccSkytapEnvironmentConfig_UserData(uniqueSuffix int, templateID string, userData string) string {
	return fmt.Sprintf(`
      resource "skytap_environment" "foo" {
	    template_id = "%s"
		user_data = %s
	    name = "tftest-environment-%d"
	    description = "This is an environment created by the skytap terraform provider acceptance test"
      }`, templateID, userData, uniqueSuffix)
}

func testAccSkytapEnvironmentConfigBlock(uniqueSuffix int, templateID string, requirements string, block string) string {
	return fmt.Sprintf(`
		%s

      resource "skytap_environment" "foo" {
	    template_id = "%s"
		%s
	    name = "tftest-environment-%d"
	    description = "This is an environment created by the skytap terraform provider acceptance test"
      }`, requirements, templateID, block, uniqueSuffix)
}

func getEnvironment(rs *terraform.ResourceState) (*skytap.Environment, error) {
	var err error
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).environmentsClient
	ctx := context.TODO()

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
