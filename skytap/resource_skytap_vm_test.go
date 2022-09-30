package skytap

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"

	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

const (
	vmEnvironmentPrefix = "tftest-vm"
	MINUTES             = 1
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
	ctx := context.TODO()

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
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "name = \"test\"", "", ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "test"),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
		},
	})
}

func TestAccSkytapVM_Timeout(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSkytapVMConfig_basic(newEnvTemplateID, acctest.RandInt(), "", templateID, vmID, "name = \"test\"", "", `timeouts { create = "60s" }`),
				ExpectError: regexp.MustCompile(".*?context deadline exceeded.*?"),
			},
		},
	})
}

func TestAccSkytapVM_Update(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	uniqueSuffixVM := acctest.RandInt()
	var vm skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "", ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "name"),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
			{
				// Pause between the steps to avoid an issue where the Skytap console shows a "Guest OS not responding" error
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID,
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
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id 	= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}`, templateID, vmID, "name = \"test\"", `

                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = skytap_network.baz.id
						ip = "192.168.0.10"
						hostname = "bloggs-web"
                  	}

                    network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = skytap_network.baz.id
						ip = "192.168.0.11"
						hostname = "bloggs-web2"
                  	}

				`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 2),
					testAccCheckSkytapInterfaceAttributes(t, "skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.10", "192.168.0.11"}, []string{"bloggs-web", "bloggs-web2"}),
				),
			}, {
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id 	= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}
					`, templateID, vmID, "name = \"test\"",
					`

                 	network_interface {
                   		interface_type = "vmxnet3"
                   		network_id = skytap_network.baz.id
						ip = "192.168.0.20"
						hostname = "bloggs-web3"
                 	}

                    network_interface {
                   		interface_type = "vmxnet3"
                   		network_id = skytap_network.baz.id
						ip = "192.168.0.21"
						hostname = "bloggs-web4"
                 	}
					`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 2),
					testAccCheckSkytapInterfaceAttributes(t, "skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.20", "192.168.0.21"}, []string{"bloggs-web3", "bloggs-web4"}),
				),
			}, {
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id 		= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}
					`, templateID, vmID, "name = \"test\"", `
                 	network_interface {
                   		interface_type = "vmxnet3"
                   		network_id = skytap_network.baz.id
						ip = "192.168.0.22"
						hostname = "bloggs-web5"
                 	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapInterfaceAttributes(t, "skytap_environment.foo", "skytap_network.baz", &vm, skytap.NICTypeVMXNet3, []string{"192.168.0.22"}, []string{"bloggs-web5"}),
				),
			},
		},
	})
}

func TestAccSkytapVM_PublishedService(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id 	= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}`, templateID, vmID, "name = \"test\"", `

                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = skytap_network.baz.id
						ip = "192.168.0.10"
						hostname = "bloggs-web"

						published_service {
							name = "web0"
							internal_port = 8080
						}

						published_service {
							name = "web1"
							internal_port = 8081
						}
                  	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8080, 8081}),
				),
			}, {
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id   	= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}`,
					templateID, vmID, "name = \"test\"",
					`
                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = skytap_network.baz.id
						ip = "192.168.0.10"
						hostname = "bloggs-web"
						published_service {
							name = "web2"
							internal_port = 8082
						}
						published_service {
							name = "web3"
							internal_port = 8083
						}
                  	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8082, 8083}),
				),
			}, {
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id 	= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}`,
					templateID, vmID, "name = \"test\"",
					`
                  	network_interface {
                    	interface_type = "vmxnet3"
                    	network_id = skytap_network.baz.id
						ip = "192.168.0.10"
						hostname = "bloggs-web"
						published_service {
							name = "web4"
							internal_port = 8084
						}
                  	}`, ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapInterfacesExists("skytap_environment.foo", "skytap_vm.bar", "skytap_network.baz", 1),
					testAccCheckSkytapPublishedServiceAttributes(&vm, []int{8084}),
				),
			}, {
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id 	= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}`,
					templateID, vmID, "name = \"test\"",
					`
                  	network_interface {
                    	interface_type = "e1000"
                    	network_id = skytap_network.baz.id
						ip = "192.168.0.10"
						hostname = "bloggs-web"
						published_service {
							name = "web4"
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
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, `
					resource "skytap_network" "baz" {
  						name        		= "tftest-network-1"
						domain      		= "mydomain.com"
  						environment_id 	= skytap_environment.foo.id
  						subnet      		= "192.168.0.0/16"
					}

					`, templateID, vmID, "name = \"test\"",
					`
                  	network_interface {
                    	interface_type = "e1000e"
                    	network_id = skytap_network.baz.id
						ip = "192.168.0.10"
						hostname = "bloggs-web"
                  	}`, ``),
				ExpectError: regexp.MustCompile(`error creating interface: POST (.*?): 422 \(request "(.*?)"\)`),
			},
		},
	})
}

func TestAccExternalPorts(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_typical(newEnvTemplateID, templateID, vmID, uniqueSuffixEnv, 23,
					`published_service {
						name = "web-internal" 
						internal_port = 8080
					 }
					 `,
					`network_interface {
    	              interface_type = "vmxnet3"
        	          network_id     = skytap_network.dev_network.id
        	          ip         = "10.0.3.2"
                      hostname = "myhost2"

        	          published_service {
						name = "ssh"
          	            internal_port = 22
        	          }
        	          published_service {
						name = "web"
          	            internal_port = 80
        	          }
      	            }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapExternalPorts(t, "skytap_vm.cassandra1", "4"),
				),
			},
		},
	})
}

func TestAccSkytapVM_Typical(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_typical(newEnvTemplateID, templateID, vmID, uniqueSuffixEnv, 22, "", ""),
			}, {
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_typical(newEnvTemplateID, templateID, vmID, uniqueSuffixEnv, 23,
					`published_service  {
						name = "web-internal"
						internal_port = 8080
					}`,
					`network_interface  {
    	              interface_type = "vmxnet3"
        	          network_id     = skytap_network.dev_network.id
        	          ip         = "10.0.3.2"
                      hostname = "myhost2"

        	          published_service  {
						name = "ssh"
          	            internal_port = 22
        	          }

        	          published_service {
						name = "web"
          	            internal_port = 80
        	          }

      	            }`),
			},
		},
	})
}

func TestAccSkytapVMCPURam_Create(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "", ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "name"),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config:    testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "name = \"test\"", "", ``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "test"),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "name = \"test\"", "",
					`cpus = 1`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "1"),
					testAccCheckSkytapVMCPU(t, &vm, 1),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "name = \"test\"", "",
					`ram = 8192`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "8192"),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMRAM(t, &vmUpdated, 8192),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "name = \"test\"", "",
					` cpus = 2
                               ram = 4096`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "2"),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "4096"),
					testAccCheckSkytapVMCPU(t, &vmUpdated, 2),
					testAccCheckSkytapVMRAM(t, &vmUpdated, 4096),
				),
			},
		},
	})
}

// To ensure the presence of a disk works unchanged
func TestAccSkytapVMCPU_DiskIntact(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`	cpus = 2
								ram = 2048
								disk  {
									size = 2048
									name = "disk1"
								}

 							 `),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "2"),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"disk1"}),
					testAccCheckSkytapVMCPU(t, &vm, 2),
					testAccCheckSkytapVMDisks(t, &vm, []int{2048}),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`  cpus = 1
								ram = 2048
								disk {
									size = 2048
									name = "disk1"
								}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "1"),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"disk1"}),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMCPU(t, &vmUpdated, 1),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{2048}),
				),
			},
		},
	})
}

func TestAccSkytapVMCPURAM_UpdateNPECheck(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "cpus"),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "ram"),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`cpus = 2
							  ram = 2048`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
				),
			},
		},
	})
}

func TestAccSkytapVMCPURAM_Invalid(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`cpus = 121
                              ram = 819000000002`),
				ExpectError: regexp.MustCompile(`expected cpus to be in the range \(1 - 12\), got 121`),
			},
		},
	})
}

func TestAccSkytapVMCPU_OutOfRange(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupNonDefaultEnvironment("SKYTAP_TEMPLATE_OUTOFRANGE_ID", "1877151", "SKYTAP_VM_OUTOFRANGE_ID", "66413705")
	uniqueSuffixEnv := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`cpus = 12
                              ram = 131072`),
				ExpectError: regexp.MustCompile(`the 'cpus' argument has been assigned \(12\) which is more than the maximum allowed \(8\) as defined by this VM`),
			},
		},
	})
}

func TestAccSkytapVMCPU_OutOfRangeAfterUpdate(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupNonDefaultEnvironment("SKYTAP_TEMPLATE_OUTOFRANGE_ID", "1877151", "SKYTAP_VM_OUTOFRANGE_ID", "66413705")
	uniqueSuffixEnv := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					``),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`cpus = 12
                              ram = 131072`),
				ExpectError: regexp.MustCompile(`the 'cpus' argument has been assigned \(12\) which is more than the maximum allowed \(8\) as defined by this VM`),
			},
		},
	})
}

func TestAccSkytapVMDisks_Create(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk  {
								size = 2048
								name = "smaller"
                             }

							disk {
								size = 2048
								name = "smaller2"
							}

							disk {
								size = 2049
								name = "bigger"
							}
								
							`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "3", []string{"smaller", "smaller2", "bigger"}),
					testAccCheckSkytapVMDisks(t, &vm, []int{2048, 2048, 2049}),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk {
								size = 2048
								name = "smaller2"  # stays the same
							  }
			
							  disk {
								size = 2049
								name = "bigger2" # new
 							  }
					`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "2", []string{"smaller2", "bigger2"}),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{2048, 2049}),
				),
			},
		},
	})
}

// NPE checks
func TestAccSkytapVMDisks_UpdateNPECheck(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk  {
								size = 8000
								name = "smaller"
								}
				`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"smaller"}),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{8000}),
				),
			},
		},
	})
}

func TestAccSkytapVMDisks_Resize(t *testing.T) {

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk  {
										size = 8000
										name = "smaller"
									}
					`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"smaller"}),
					testAccCheckSkytapVMDisks(t, &vm, []int{8000}),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk {
										size = 8192
										name = "smaller"
									}
							`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"smaller"}),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{8192}),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk  {
								size = 6000
								name = "smaller"
							}`),
				ExpectError: regexp.MustCompile(`cannot shrink volume \(smaller\) from size \(8192\) to size \(6000\)`),
			},
		},
	})
}

func TestAccSkytapVMDisk_Invalid(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk  {
										size = 2047
										name = "too small"
									}
				`),
				ExpectError: regexp.MustCompile(`expected disk.0.size to be in the range \(2048 - 2096128\), got 2047`),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`disk  {
								size = 2096129
								name = "too big"
							}`),
				ExpectError: regexp.MustCompile(`expected disk.0.size to be in the range \(2048 - 2096128\), got 2096129`),
			},
		},
	})
}

func TestAccSkytapVMDisk_OS(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`os_disk_size = 30721`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "os_disk_size", "30721"),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "max_ram"),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "max_cpus"),
					testAccCheckSkytapVMOSDisk(t, &vm, 30721),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`os_disk_size = 30722`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "os_disk_size", "30722"),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMOSDisk(t, &vmUpdated, 30722),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`os_disk_size = 3000`),
				ExpectError: regexp.MustCompile(`cannot shrink volume \(OS\) from size \(30722\) to size \(3000\)`),
			},
		},
	})
}

func TestAccSkytapVMDisk_OSChangeAfter(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "os_disk_size"),
				),
			},
			{
				PreConfig: pause(MINUTES),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`os_disk_size = 30721`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "os_disk_size", "30721"),
				),
			},
		},
	})
}

func TestAccSkytapVM_Concurrent(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_concurrent(newEnvTemplateID, uniqueSuffixEnv, templateID, vmID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.webservers.0", &vm),
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.webservers.1", &vm),
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.webservers.2", &vm),
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.webservers.3", &vm),
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.webservers.4", &vm),
				),
			},
		},
	})
}

func TestAccSkytapVM_UserData(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

	userData := `user_data = <<EOF
		cat /proc/cpu_info
	EOF
	`
	userDataRe, _ := regexp.Compile("\\s*cat /proc/cpu_info\\n")

	userDataUpdated := `user_data = <<EOF
		less /proc/cpu_info
	EOF
	`
	userDataUpdatedRe, _ := regexp.Compile("\\s*less /proc/cpu_info\\n")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfigBlock(newEnvTemplateID, uniqueSuffixEnv, templateID, vmID, "test",
					"", userData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "test"),
					resource.TestMatchResourceAttr("skytap_vm.bar", "user_data", userDataRe),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
			{
				Config: testAccSkytapVMConfigBlock(newEnvTemplateID, uniqueSuffixEnv, templateID, vmID, "test",
					"", userDataUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "test"),
					resource.TestMatchResourceAttr("skytap_vm.bar", "user_data", userDataUpdatedRe),
					testAccCheckSkytapVMRunning(&vm),
				),
			},
		},
	})
}

func TestAccSkytapVM_Labels(t *testing.T) {
	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM

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
			value = "Marketing"
		}
		label {
			category = skytap_label_category.owners_label.name
			value = "Accounting"
		}
	`

	labelEnv := fmt.Sprintf("tftest-Environment-%d", uniqueSuffixEnv)
	labelOwners := fmt.Sprintf("tftest-Owners-%d", uniqueSuffixEnv)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfigBlock(newEnvTemplateID, uniqueSuffixEnv, templateID, vmID, "test",
					labelRequirements(uniqueSuffixEnv), labels),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "test"),
					testAccCheckSkytapVMRunning(&vm),
					testAccCheckSkytapLabelExists(&vm, labelEnv, "Prod"),
					testAccCheckSkytapLabelExists(&vm, labelOwners, "Finance"),
					testAccCheckSkytapLabelExists(&vm, labelOwners, "Accounting"),
				),
			},
			{
				Config: testAccSkytapVMConfigBlock(newEnvTemplateID, uniqueSuffixEnv, templateID, vmID, "test",
					labelRequirements(uniqueSuffixEnv), labelsUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "name", "test"),
					testAccCheckSkytapVMRunning(&vm),
					testAccCheckSkytapLabelExists(&vm, labelEnv, "UAT"),
					testAccCheckSkytapLabelExists(&vm, labelOwners, "Marketing"),
					testAccCheckSkytapLabelExists(&vm, labelOwners, "Accounting"),
				),
			},
		},
	})
}

func testAccCheckSkytapExternalPorts(t *testing.T, vmName string, count string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsVM, err := getResource(s, vmName)
		if err != nil {
			return err
		}
		assert.Equal(t, count, rsVM.Primary.Attributes["service_ports.%"], "empty map entry")
		assert.Equal(t, count, rsVM.Primary.Attributes["service_ips.%"], "empty map entry")

		return nil
	}
}

func setupNonDefaultEnvironment(templateKey string, templateIDFallback string, vmKey string, vmIDFallback string) (templateID, vmID, newEnvTemplateID string) {
	templateID = utils.GetEnv(templateKey, templateIDFallback)
	vmID = utils.GetEnv(vmKey, vmIDFallback)
	newEnvTemplateID = utils.GetEnv("SKYTAP_TEMPLATE_NEW_ENV_ID", templateID)
	return
}

func setupEnvironment() (string, string, string) {
	return setupNonDefaultEnvironment("SKYTAP_TEMPLATE_ID", "1469947", "SKYTAP_VM_ID", "37715265")
}

func testAccSkytapVMConfig_typical(envTemplateID string, templateID string, vmID string, uniqueSuffixEnv int, existingPort int, extraPublishedService string, extraNIC string) string {
	config := fmt.Sprintf(`

    resource "skytap_environment" "my_new_environment" {
      name = "%s-environment-%d"
      template_id = "%s"
      description = "An enviroment"
    }

    resource "skytap_network" "dev_network" {
      environment_id = skytap_environment.my_new_environment.id
      name = "tftest-network-1"
      domain = "dev.skytap.io"
      subnet = "10.0.3.0/24"
    }

    resource "skytap_vm" "cassandra1" {
      environment_id = skytap_environment.my_new_environment.id
      template_id = "%s"
      vm_id = "%s"
      name = "cassandra1"
      network_interface  {
        interface_type = "vmxnet3"
        network_id = skytap_network.dev_network.id
        ip = "10.0.3.1"
        hostname = "myhost"

        published_service  {
          name = "service"
          internal_port = %d
        }
        %s
      }
      %s
    }`, vmEnvironmentPrefix, uniqueSuffixEnv, envTemplateID, templateID, vmID, existingPort, extraPublishedService, extraNIC)
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

func testAccCheckSkytapInterfaceAttributes(t *testing.T, environmentName string, networkName string, vm *skytap.VM, nicType skytap.NICType, ips []string, hostnames []string) resource.TestCheckFunc {
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
			assert.NotNil(t, adapter.ID)
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

func testAccCheckSkytapLabelExists(vm *skytap.VM, category string, label string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, l := range vm.Labels {
			if strings.ToLower(*l.LabelCategory) == strings.ToLower(category) &&
				strings.ToLower(*l.Value) == strings.ToLower(label) {
				return nil
			}
		}
		return fmt.Errorf("VM with id %d, do not contain label (%s: %s)", vm.ID, category, label)
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
		environment_id    = skytap_environment.foo.id
   		template_id       = "%s"
 		vm_id      		  = "%s"
		%s
        %s
		%s
 	}
 `, envTemplateID, vmEnvironmentPrefix, uniqueSuffixEnv, network, VMTemplateID, VMID, name, networkInterface, hardware)
	return config
}

func testAccSkytapVMConfigBlock(envTemplateID string, uniqueSuffixEnv int, VMTemplateID string,
	VMID string, name string, requirements string, block string) string {

	config := fmt.Sprintf(`
 	resource "skytap_environment" "foo" {
 		template_id = "%s"
 		name 		= "%s-environment-%d"
 		description = "This is an environment to support a vm skytap terraform provider acceptance test"
 	}

	%s

 	resource "skytap_vm" "bar" {
		environment_id    = skytap_environment.foo.id
   		template_id       = "%s"
 		vm_id      		  = "%s"
		name 			  = "%s"
		%s
 	}
 	`,
		envTemplateID, vmEnvironmentPrefix, uniqueSuffixEnv, requirements, VMTemplateID, VMID, name, block)
	return config
}

func testAccSkytapVMConfig_concurrent(envTemplateID string, uniqueSuffixEnv int, VMTemplateID string, VMID string) string {
	config := fmt.Sprintf(`
 	resource "skytap_environment" "foo" {
 		template_id = "%s"
 		name 		= "%s-environment-%d"
 		description = "This is an environment to support a vm skytap terraform provider acceptance test"
 	}

	resource "skytap_network" "network" {
	  environment_id = skytap_environment.foo.id
	  name = "Network"
	  domain = "skytap.services"
	  subnet = "192.168.1.0/24"
	  gateway = "192.168.1.254"
	  tunnelable = true
	}
	
	variable "ip_addreses" {
	  default = {
		"0" = "192.168.1.100"
		"1" = "192.168.1.101"
		"2" = "192.168.1.102"
		"3" = "192.168.1.103"
		"4" = "192.168.1.104"
	  }
	}
	
	resource "skytap_vm" "webservers" {
	  count = 5
	  template_id = "%s"
	  vm_id = "%s"
	  environment_id = skytap_environment.foo.id
	  name = "web-${count.index}"

	  network_interface {
		interface_type = "vmxnet3"
		network_id = skytap_network.network.id
		ip = lookup(var.ip_addreses, count.index)
		hostname = "web-${count.index}"
	  }

	}
 `, envTemplateID, vmEnvironmentPrefix, uniqueSuffixEnv, VMTemplateID, VMID)
	return config
}

func getVM(rs *terraform.ResourceState, environmentID string) (*skytap.VM, error) {
	var err error
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).vmsClient
	ctx := context.TODO()

	// Retrieve our vm by referencing its state ID for API lookup
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
		if skytap.VMRunstateRunning == *vm.Runstate {
			return nil
		}
		return fmt.Errorf("vm not running but in runstate (%s)", string(*vm.Runstate))
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

func testAccCheckSkytapVMDiskResource(t *testing.T, vmName string, disks string, names []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsVM, err := getResource(s, vmName)
		assert.NotNil(t, rsVM)
		if err != nil {
			return err
		}
		assert.Equal(t, disks, rsVM.Primary.Attributes["disk.#"])
		for key := range rsVM.Primary.Attributes {
			re := regexp.MustCompile("\\d+")
			hash := re.FindString(key)
			nameKey := fmt.Sprintf("disk.%s.name", hash)
			if v, ok := rsVM.Primary.Attributes[nameKey]; ok {
				found := false
				for _, name := range names {
					if name == v {
						found = true
						break
					}
				}
				assert.True(t, found, "locating name")
			}
		}

		return nil
	}
}

func testAccCheckSkytapVMDisks(t *testing.T, vm *skytap.VM, sizes []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		sort.Slice(vm.Hardware.Disks, func(i, j int) bool {
			return *vm.Hardware.Disks[i].Size < *vm.Hardware.Disks[j].Size
		})
		ok := assert.Equal(t, len(sizes)+1, len(vm.Hardware.Disks)) // os disk means +1
		if ok {
			for idx, size := range sizes {
				assert.Equal(t, size, *vm.Hardware.Disks[idx].Size, "disk size")
			}
		}
		return nil
	}
}

func testAccCheckSkytapVMUpdated(t *testing.T, vm *skytap.VM, vm2 *skytap.VM) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		assert.Equal(t, *vm.ID, *vm2.ID, "vm ids")
		return nil
	}
}

func testAccCheckSkytapVMOSDisk(t *testing.T, vm *skytap.VM, size int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		assert.Equal(t, size, *vm.Hardware.Disks[0].Size)
		return nil
	}
}

func pause(minutes int) func() {
	return func() {
		log.Printf("[INFO] pausing for %d minutes", minutes)
		time.Sleep(time.Duration(minutes) * time.Minute)
	}
}
