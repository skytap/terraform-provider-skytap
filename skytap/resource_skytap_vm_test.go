package skytap

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
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
	//t.Parallel()
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
	//t.Parallel()
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
	//t.Parallel()
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
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 2),
					testAccCheckSkytapInterfaceAttributes("skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.11", "192.168.0.10"}, []string{"bloggs-web2", "bloggs-web"}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, testVMTemplateID, testVMID, "name = \"test\"", `
                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.20"
						hostname = "bloggs-web3"
                  	}
                    network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.21"
						hostname = "bloggs-web4"
                  	}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 2),
					testAccCheckSkytapInterfaceAttributes("skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.21", "192.168.0.20"}, []string{"bloggs-web4", "bloggs-web3"}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, testVMTemplateID, testVMID, "name = \"test\"", `
                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.22"
						hostname = "bloggs-web5"
                  	}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapInterfaceAttributes("skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.22"}, []string{"bloggs-web5"}),
				),
			},
		},
	})
}

func TestAccSkytapVM_PublishedService(t *testing.T) {
	//t.Parallel()
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
						published_service {
							internal_port = 8080
						}
						published_service {
							internal_port = 8081
						}
                  	}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8080, 8081}),
				),
			}, {
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
						published_service {
							internal_port = 8082
						}
						published_service {
							internal_port = 8083
						}
                  	}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8082, 8083}),
				),
			}, {
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
						published_service {
							internal_port = 8084
						}
                  	}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8084}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(testEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, testVMTemplateID, testVMID, "name = \"test\"", `
                  	network_interface {
                    	interface_type = "e1000"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.10"
						hostname = "bloggs-web"
						published_service {
							internal_port = 8084
						}
                  	}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8084}),
				),
			},
		},
	})
}

// the interface type is wrong and will be rejected by the API. This tests the SDK error handling.
func TestAccSkytapVM_PublishedServiceBadNic(t *testing.T) {
	//t.Parallel()
	uniqueSuffixEnv := acctest.RandInt()

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
                    	interface_type = "e1000e"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.10"
						hostname = "bloggs-web"
                  	}`),
				ExpectError: regexp.MustCompile(`error creating interface: POST (.*?): 422 \(request "(.*?)"\)`),
			},
		},
	})
}

func testAccCheckSkytapPublishedServiceAttributes(vm *skytap.VM, ports []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		sort.Slice(vm.Interfaces, func(i, j int) bool {
			return *vm.Interfaces[i].ID > *vm.Interfaces[j].ID
		})

		adapter := vm.Interfaces[0]

		sort.Slice(adapter.Services, func(i, j int) bool {
			return *adapter.Services[i].ID < *adapter.Services[j].ID
		})

		for i := 0; i < len(adapter.Services); i++ {
			publishedService := adapter.Services[i]
			if *publishedService.InternalPort != ports[i] {
				return fmt.Errorf("the publishedService port (%d) is not configured as expected (%d)", *publishedService.InternalPort, ports[i])
			}
			if *publishedService.ID != strconv.Itoa(ports[i]) {
				return fmt.Errorf("the publishedService ID (%s) is not configured as expected (%d)", *publishedService.ID, ports[i])
			}
			if publishedService.ExternalPort == nil {
				return fmt.Errorf("the publishedService ExternalPort is not configured")
			}
			if publishedService.ExternalIP == nil {
				return fmt.Errorf("the publishedService ExternalIP is not configured")
			}
		}
		return nil
	}
}

func testAccCheckSkytapInterfaceAttributes(environmentName string, networkName string, vm *skytap.VM, nicType skytap.NICType, ips []string, hostnames []string) resource.TestCheckFunc {
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
			return *vm.Interfaces[i].ID > *vm.Interfaces[j].ID
		})

		if len(vm.Interfaces) != len(ips) {
			return fmt.Errorf("invalid number of interfaces, expected (%d)", len(vm.Interfaces))
		}

		for i := 0; i < len(ips); i++ {
			adapter := vm.Interfaces[i]

			if *adapter.IP != ips[i] {
				return fmt.Errorf("the interface ip (%s) is not configured as expected (%s)", *adapter.IP, ips[i])
			}
			if *adapter.Hostname != hostnames[i] {
				return fmt.Errorf("the interface hostname (%s) is not configured as expected (%s)", *adapter.Hostname, hostnames[i])
			}
			if *adapter.NICType != nicType {
				return fmt.Errorf("the interface NIC types (%s) are not configured as expected (%s)", *adapter.NICType, nicType)
			}
			if *adapter.NetworkID != *net.ID {
				return fmt.Errorf("the interface network IDs (%s) are not configured as expected (%s)", *adapter.NetworkID, *net.ID)
			}
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

func testAccCheckSkytapInterfacesExists(environmentName string, vmName string, networkName string, interfaceCount int) resource.TestCheckFunc {
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

		if count != interfaceCount {
			return fmt.Errorf("expecting %d networks but found %d", interfaceCount, count)
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
			err = fmt.Errorf("vm (%s) was not found - does not exist", rs.Primary.ID)
		}

		err = fmt.Errorf("error retrieving vm (%s): %v", rs.Primary.ID, err)
	}
	return vm, err
}

func testAccCheckSkytapVMRunning(vm *skytap.VM) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource.TestCheckResourceAttr("skytap_vm.bar", "runstate", string(skytap.VMRunstateRunning))
		if skytap.VMRunstateRunning != *vm.Runstate {
			return fmt.Errorf("vm (%s) is not running as expected", *vm.ID)
		}
		return nil
	}
}
