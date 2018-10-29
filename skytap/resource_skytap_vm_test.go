package skytap

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
	"log"
	"testing"
)

const (
	vmEnvironmentPrefix = "tftest-vm"
	testEnvTemplateID   = "1448141"
	testVMTemplateID    = "1448141"
	testVMID            = "36628546"
)

func init() {
	resource.AddTestSweepers("skytap_vm", &resource.Sweeper{
		Name: "skytap_vm",
		F:    testSweepSkytapVM,
	})
}

func testSweepSkytapVM(region string) error {
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
		if shouldSweepAcceptanceTestResourceWithPrefix(*e.Name, vmEnvironmentPrefix) {
			log.Printf("destroying environment %s", *e.Name)
			if err := client.Delete(ctx, *e.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccSkytapVM_Basic(t *testing.T) {
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, testVMTemplateID, testVMID, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "CentOS 6 Desktop x64"),
					resource.TestCheckResourceAttr("skytap_vm.bar", "runstate", string(skytap.VMRunstateRunning)),
				),
			},
		},
	})
}

func TestAccSkytapVM_Update(t *testing.T) {
	uniqueSuffixEnv := acctest.RandInt()
	uniqueSuffixVM := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, testVMTemplateID, testVMID, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "CentOS 6 Desktop x64"),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, testVMTemplateID, testVMID,
					fmt.Sprintf("name = \"tftest-vm-%d\"", uniqueSuffixVM)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", fmt.Sprintf("tftest-vm-%d", uniqueSuffixVM)),
				),
			},
		},
	})
}

func testAccCheckSkytapVMExists(environmentName string, vmName string, vm *skytap.VM) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rsEnvironment, err := getResource(s, environmentName)
		if err != nil {
			return err
		}
		environmentID := rsEnvironment.Primary.ID

		rsVM, err := getResource(s, vmName)
		if err != nil {
			return err
		}

		// Get the vm
		createdVM, err := getVM(rsVM, environmentID)
		if err != nil {
			return err
		}

		*vm = *createdVM
		log.Printf("[DEBUG] vm (%s)\n", *vm.ID)

		return nil
	}
}

func testAccSkytapVMConfig_basic(envTemplateID string, uniqueSuffixEnv int, VMTemplateID string, VMID string, name string) string {
	config := fmt.Sprintf(`
 	resource "skytap_environment" "foo" {
 		template_id = %s
 		name 		= "%s-environment-%d"
 		description = "This is an environment to support a vm skytap terraform provider acceptance test"
 	}

 	resource "skytap_vm" "bar" {
		environment_id    = "${skytap_environment.foo.id}"
   		template_id       = %s
 		vm_id      		  = %s
		%s
 	}
 `, envTemplateID, vmEnvironmentPrefix, uniqueSuffixEnv, VMTemplateID, VMID, name)
	return config
}

func getVM(rs *terraform.ResourceState, environmentId string) (*skytap.VM, error) {
	var err error
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).vmsClient
	ctx := testAccProvider.Meta().(*SkytapClient).StopContext

	// Retrieve our vm by referencing it's state ID for API lookup
	vm, errClient := client.Get(ctx, environmentId, rs.Primary.ID)
	if errClient != nil {
		if utils.ResponseErrorIsNotFound(err) {
			err = errors.Errorf("vm (%s) was not found - does not exist", rs.Primary.ID)
		}

		err = fmt.Errorf("error retrieving vm (%s): %v", rs.Primary.ID, err)
	}
	return vm, err
}
