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
	"sort"
	"strings"
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
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, "", testVMTemplateID, testVMID, "name = \"test\"", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "test"),
					testAccCheckSkytapVMRunning(&vm),
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
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, "", testVMTemplateID, testVMID, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "CentOS 6 Desktop x64"),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, "", testVMTemplateID, testVMID,
					fmt.Sprintf("name = \"tftest-vm-%d\"", uniqueSuffixVM), ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", fmt.Sprintf("tftest-vm-%d", uniqueSuffixVM)),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
		},
	})
}

func TestAccSkytapVM_Interface(t *testing.T) {
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, testVMTemplateID, testVMID, "name = \"test\"", `
                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.10"
						hostname = "bloggs-web"
                  	}
                    network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.11"
						hostname = "bloggs-web2"
                  	}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfaceExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz"),
					testAccCheckSkytapInterfaceAttributes("skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, "192.168.0.10", "192.168.0.11", "bloggs-web", "bloggs-web2"),
				),
			},
		},
	})
}
func testAccCheckSkytapInterfaceAttributes(environmentName string, networkName string, vm *skytap.VM, nicType skytap.NICType, ip1 string, ip2 string, hostname1 string, hostname2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rsEnvironment, err := getResource(s, environmentName)
		if err != nil {
			return err
		}
		environmentID := rsEnvironment.Primary.ID

		// Get the network
		rsNetwork, err := getResource(s, networkName)
		if err != nil {
			return err
		}

		// Get the network
		net, err := getNetwork(rsNetwork, environmentID)
		if err != nil {
			return err
		}

		sort.Slice(vm.Interfaces, func(i, j int) bool {
			return strings.Compare(*vm.Interfaces[i].ID, *vm.Interfaces[j].ID) == 0
		})

		adapter1 := vm.Interfaces[1]
		adapter2 := vm.Interfaces[2]

		if *adapter1.IP != ip1 {
			return fmt.Errorf("the interface ip (%s) is not configured as expected (%s)", *adapter1.IP, ip1)
		}
		if *adapter2.IP != ip2 {
			return fmt.Errorf("the interface ip (%s) is not configured as expected (%s)", *adapter2.IP, ip2)
		}
		if *adapter1.Hostname != hostname1 {
			return fmt.Errorf("the interface hostname (%s) is not configured as expected (%s)", *adapter1.Hostname, hostname1)
		}
		if *adapter2.Hostname != hostname2 {
			return fmt.Errorf("the interface hostname (%s) is not configured as expected (%s)", *adapter2.Hostname, hostname2)
		}
		if *adapter1.NICType != nicType || *adapter2.NICType != nicType {
			return fmt.Errorf("the interface NIC types (%s,%s) are not configured as expected (%s)", *adapter1.NICType, *adapter2.NICType, nicType)
		}

		if *adapter1.NetworkID != *net.ID || *adapter2.NetworkID != *net.ID {
			return fmt.Errorf("the interface network IDs (%s,%s) are not configured as expected (%s)", *adapter1.NetworkID, *adapter2.NetworkID, *net.ID)
		}

		return nil
	}
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

func testAccCheckSkytapInterfaceExists(environmentName string, vmName string, networkName string) resource.TestCheckFunc {
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

		// Get the network
		rsNetwork, err := getResource(s, networkName)
		if err != nil {
			return err
		}

		// Get the network
		net, err := getNetwork(rsNetwork, environmentID)
		if err != nil {
			return err
		}

		count := 0
		for i := 0; i < len(createdVM.Interfaces); i++ {
			if *createdVM.Interfaces[i].NetworkID == *net.ID {
				count++
			}
		}

		if count < 2 {
			return errors.New("not all interfaces were attached to the network")
		}

		return nil
	}
}

func testAccSkytapVMConfig_basic(envTemplateID string, uniqueSuffixEnv int, network string, VMTemplateID string, VMID string, name string, networkInterface string) string {
	config := fmt.Sprintf(`
 	resource "skytap_environment" "foo" {
 		template_id = %s
 		name 		= "%s-environment-%d"
 		description = "This is an environment to support a vm skytap terraform provider acceptance test"
 	}

	%s

 	resource "skytap_vm" "bar" {
		environment_id    = "${skytap_environment.foo.id}"
   		template_id       = %s
 		vm_id      		  = %s
		%s
        %s
 	}
 `, envTemplateID, vmEnvironmentPrefix, uniqueSuffixEnv, network, VMTemplateID, VMID, name, networkInterface)
	return config
}

func getVM(rs *terraform.ResourceState, environmentID string) (*skytap.VM, error) {
	var err error
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).vmsClient
	ctx := testAccProvider.Meta().(*SkytapClient).StopContext

	// Retrieve our vm by referencing it's state ID for API lookup
	vm, errClient := client.Get(ctx, environmentID, rs.Primary.ID)
	if errClient != nil {
		if utils.ResponseErrorIsNotFound(err) {
			err = errors.Errorf("vm (%s) was not found - does not exist", rs.Primary.ID)
		}

		err = fmt.Errorf("error retrieving vm (%s): %v", rs.Primary.ID, err)
	}
	return vm, err
}

func testAccCheckSkytapVMRunning(vm *skytap.VM) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource.TestCheckResourceAttr("skytap_vm.bar", "runstate", string(skytap.VMRunstateRunning))
		if skytap.VMRunstateRunning != *vm.Runstate {
			return errors.Errorf("vm (%s) is not running as expected", *vm.ID)
		}
		return nil
	}
}
