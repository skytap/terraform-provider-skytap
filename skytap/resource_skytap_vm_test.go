package skytap

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

const (
	vmEnvironmentPrefix = "tftest-vm"
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

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		os.Setenv("SKYTAP_VM_ID", "37865463")
	}

	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", "", ``),
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

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		os.Setenv("SKYTAP_VM_ID", "37865463")
	}

	uniqueSuffixEnv := acctest.RandInt()
	uniqueSuffixVM := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "", ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "name"),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"),
					fmt.Sprintf("name = \"tftest-vm-%d\"", uniqueSuffixVM), "", ``),
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

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		os.Setenv("SKYTAP_VM_ID", "37865463")
	}

	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
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
                  	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 2),
					testAccCheckSkytapInterfaceAttributes("skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.10", "192.168.0.11"}, []string{"bloggs-web", "bloggs-web2"}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
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
                 	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 2),
					testAccCheckSkytapInterfaceAttributes("skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.20", "192.168.0.21"}, []string{"bloggs-web3", "bloggs-web4"}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
                 	network_interface {
                   	interface_type = "vmxnet3"
                   		network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.22"
						hostname = "bloggs-web5"
                 	}`, ``),
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

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		os.Setenv("SKYTAP_VM_ID", "37865463")
	}

	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
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
                  	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8080, 8081}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
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
                  	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8082, 8083}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.10"
						hostname = "bloggs-web"
						published_service {
							internal_port = 8084
						}
                  	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8084}),
				),
			}, {
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
                  	network_interface {
                    	interface_type = "e1000"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.10"
						hostname = "bloggs-web"
						published_service {
							internal_port = 8084
						}
                  	}`, ``),
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

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		os.Setenv("SKYTAP_VM_ID", "37865463")
	}

	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						"name"        		= "tftest-network-1"
						"domain"      		= "mydomain.com"
  						"environment_id" 	= "${skytap_environment.foo.id}"
  						"subnet"      		= "192.168.0.0/16"}`, os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "name = \"test\"", `
                  	network_interface {
                    	interface_type = "e1000e"
                    	network_id = "${skytap_network.baz.id}"
						ip = "192.168.0.10"
						hostname = "bloggs-web"
                  	}`, ``),
				ExpectError: regexp.MustCompile(`error creating interface: POST (.*?): 422 \(request "(.*?)"\)`),
			},
		},
	})
}

func TestAccCasandra(t *testing.T) {

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		os.Setenv("SKYTAP_VM_ID", "37865463")
	}

	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccSkytapVMConfig_cassandra(os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), uniqueSuffixEnv, 22, "", ""),
				ExpectNonEmptyPlan: false,
			}, {
				Config: testAccSkytapVMConfig_cassandra(os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), uniqueSuffixEnv, 23, `"published_service" = {"internal_port" = 8080}`,
					`"network_interface" = {
    	              "interface_type" = "vmxnet3"
        	          "network_id"     = "${skytap_network.dev_network.id}"
        	          "ip"         = "10.0.3.2"
                      "hostname" = "myhost2"

        	          "published_service" = {
          	            "internal_port" = 22
        	          }
        	          "published_service" = {
          	            "internal_port" = 80
        	          }
      	            }`),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccSkytapVMCPURam_Create(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1497575")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1497575")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=39084167")
		os.Setenv("SKYTAP_VM_ID", "39084167")
	}

	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"cpus" = 12`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "12"),
					testAccCheckSkytapVMCPU(t, &vm, 12),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"ram" = 8192`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "8192"),
					testAccCheckSkytapVMRAM(t, &vm, 8192),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"cpus" = 12
									"ram" = 8192`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "12"),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "8192"),
					testAccCheckSkytapVMCPU(t, &vm, 12),
					testAccCheckSkytapVMRAM(t, &vm, 8192),
				),
			},
		},
	})
}

func TestAccSkytapVMCPU_Invalid(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1496747")
		os.Setenv("SKYTAP_TEMPLATE_ID", "1496747")
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=39055511")
		os.Setenv("SKYTAP_VM_ID", "39055511")
	}

	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(os.Getenv("SKYTAP_TEMPLATE_ID"), uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"cpus" = 121
									"ram" = 819000000002`),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccSkytapVMConfig_cassandra(templateID string, vmID string, uniqueSuffixEnv int, existingPort int, extraPublishedService string, extraNIC string) string {
	config := fmt.Sprintf(`

    resource "skytap_environment" "my_new_environment" {
      "name" = "%s-environment-%d"
      "template_id" = "%s"
      "description" = "An enviroment"
    }

    resource "skytap_network" "dev_network" {
      "environment_id" = "${skytap_environment.my_new_environment.id}"
      "name" = "tftest-network-1"
      "domain" = "dev.skytap.io"
      "subnet" = "10.0.3.0/24"
    }

    resource "skytap_vm" "cassandra1" {
      "environment_id" = "${skytap_environment.my_new_environment.id}"
      "template_id" = "%s"
      "vm_id" = "%s"
      "name" = "cassandra1"
      "network_interface" = {
        "interface_type" = "vmxnet3"
        "network_id" = "${skytap_network.dev_network.id}"
        "ip" = "10.0.3.1"
        "hostname" = "myhost"

        "published_service" = {
          "internal_port" = %d
        }
        %s
      }
      %s
    }`, vmEnvironmentPrefix, uniqueSuffixEnv, templateID, templateID, vmID, existingPort, extraPublishedService, extraNIC)
	return config
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
			return *vm.Interfaces[i].IP < *vm.Interfaces[j].IP
		})

		for i := 0; i < len(ips); i++ {
			adapter := vm.Interfaces[i]

			if *adapter.IP != ips[i] {
				return fmt.Errorf("the interface ip (%s) is not configured as expected (%s)", *adapter.IP, ips[i])
			}
			if len(hostnames) > i {
				if *adapter.Hostname != hostnames[i] {
					return fmt.Errorf("the interface hostname (%s) is not configured as expected (%s)", *adapter.Hostname, hostnames[i])
				}
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

func testAccSkytapVMConfig_basic(envTemplateID string, uniqueSuffixEnv int, network string, VMTemplateID string, VMID string, name string, networkInterface string, hardware string) string {
	config := fmt.Sprintf(`
 	resource "skytap_environment" "foo" {
 		template_id = "%s"
 		name 		= "%s-environment-%d"
 		description = "This is an environment to support a vm skytap terraform provider acceptance test"
 	}

	%s

 	resource "skytap_vm" "bar" {
		environment_id    = "${skytap_environment.foo.id}"
   		template_id       = "%s"
 		vm_id      		  = "%s"
		%s
        %s
		%s
 	}
 `, envTemplateID, vmEnvironmentPrefix, uniqueSuffixEnv, network, VMTemplateID, VMID, name, networkInterface, hardware)
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

func testAccCheckSkytapVMCPU(t *testing.T, vm *skytap.VM, cpus int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		assert.Equal(t, cpus, *vm.Hardware.CPUs, "cpus")
		return nil
	}
}

func testAccCheckSkytapVMRAM(t *testing.T, vm *skytap.VM, ram int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		assert.Equal(t, ram, *vm.Hardware.RAM, "ram")
		return nil
	}
}
