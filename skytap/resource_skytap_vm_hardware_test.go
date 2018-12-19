package skytap

import (
	"log"
	"os"
	"regexp"
	"sort"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"
)

func TestAccSkytapVMCPURam_Create(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=136409")
		err := os.Setenv("SKYTAP_TEMPLATE_ID", "136409")
		assert.NoError(t, err)
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=849656")
		err := os.Setenv("SKYTAP_VM_ID", "849656")
		assert.NoError(t, err)
	}
	newEnvTemplateID := os.Getenv("SKYTAP_TEMPLATE_ID")
	if os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID") != "" {
		newEnvTemplateID = os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID")
		log.Printf("[DEBUG] SKYTAP_TEMPLATE_NEW_ENV_ID=%s", newEnvTemplateID)
	}
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "cpus"),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "ram"),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"cpus" = 8`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "8"),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMCPU(t, &vmUpdated, 8),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"ram" = 8192`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "8192"),
					testAccCheckSkytapVMRAM(t, &vm, 8192),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"cpus" = 4
									"ram" = 4096`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "4"),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "4096"),
					testAccCheckSkytapVMCPU(t, &vm, 4),
					testAccCheckSkytapVMRAM(t, &vm, 4096),
				),
			},
		},
	})
}

func TestAccSkytapVMCPU_Invalid(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		err := os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
		assert.NoError(t, err)
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		err := os.Setenv("SKYTAP_VM_ID", "37865463")
		assert.NoError(t, err)
	}
	newEnvTemplateID := os.Getenv("SKYTAP_TEMPLATE_ID")
	if os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID") != "" {
		newEnvTemplateID = os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID")
		log.Printf("[DEBUG] SKYTAP_TEMPLATE_NEW_ENV_ID=%s", newEnvTemplateID)
	}
	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"cpus" = 121
									"ram" = 819000000002`),
				ExpectError: regexp.MustCompile(`config is invalid: 2 problems:*`),
			},
		},
	})
}

func TestAccSkytapVMCPU_OutOfRange(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=136409")
		err := os.Setenv("SKYTAP_TEMPLATE_ID", "136409")
		assert.NoError(t, err)
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=849656")
		err := os.Setenv("SKYTAP_VM_ID", "849656")
		assert.NoError(t, err)
	}
	newEnvTemplateID := os.Getenv("SKYTAP_TEMPLATE_ID")
	if os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID") != "" {
		newEnvTemplateID = os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID")
		log.Printf("[DEBUG] SKYTAP_TEMPLATE_NEW_ENV_ID=%s", newEnvTemplateID)
	}
	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"cpus" = 12
									"ram" = 131072`),
				ExpectError: regexp.MustCompile(`the 'cpus' argument has been assigned 12 which is more than the maximum allowed \(8\) as defined by this VM`),
			},
		},
	})
}

func TestAccSkytapVMDisks_Create(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=136409")
		err := os.Setenv("SKYTAP_TEMPLATE_ID", "136409")
		assert.NoError(t, err)
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=849656")
		err := os.Setenv("SKYTAP_VM_ID", "849656")
		assert.NoError(t, err)
	}
	newEnvTemplateID := os.Getenv("SKYTAP_TEMPLATE_ID")
	if os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID") != "" {
		newEnvTemplateID = os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID")
		log.Printf("[DEBUG] SKYTAP_TEMPLATE_NEW_ENV_ID=%s", newEnvTemplateID)
	}
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"disk" = {
										"size" = 2048
										"name" = "smaller"  
									}
									"disk" = {
										"size" = 2048
										"name" = "smaller2"  
									}
									"disk" = {
										"size" = 2049
										"name" = "bigger"
									}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "3"),
					testAccCheckSkytapVMDisks(t, &vm, []int{2048, 2048, 2049}),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"disk" = {
										"size" = 2048
										"name" = "smaller2"  # stays the same
									}
									"disk" = {
										"size" = 2049
										"name" = "bigger2" # new
									}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "2"),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{2048, 2049}),
				),
			},
		},
	})
}

func TestAccSkytapVMDisk_Invalid(t *testing.T) {
	//t.Parallel()

	if os.Getenv("SKYTAP_TEMPLATE_ID") == "" {
		log.Printf("[WARN] SKYTAP_TEMPLATE_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_TEMPLATE_ID=1473407")
		err := os.Setenv("SKYTAP_TEMPLATE_ID", "1473407")
		assert.NoError(t, err)
	}
	if os.Getenv("SKYTAP_VM_ID") == "" {
		log.Printf("[WARN] SKYTAP_VM_ID required to run skytap_vm_resource acceptance tests. Setting: SKYTAP_VM_ID=37865463")
		err := os.Setenv("SKYTAP_VM_ID", "37865463")
		assert.NoError(t, err)
	}
	newEnvTemplateID := os.Getenv("SKYTAP_TEMPLATE_ID")
	if os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID") != "" {
		newEnvTemplateID = os.Getenv("SKYTAP_TEMPLATE_NEW_ENV_ID")
		log.Printf("[DEBUG] SKYTAP_TEMPLATE_NEW_ENV_ID=%s", newEnvTemplateID)
	}
	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"disk" = {
										"size" = 2047
										"name" = "too small"  
									}`),
				ExpectError: regexp.MustCompile(`config is invalid: skytap_vm.bar: expected disk.0.size to be in the range \(2048 - 2096128\), got 2047`),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", os.Getenv("SKYTAP_TEMPLATE_ID"), os.Getenv("SKYTAP_VM_ID"), "", "",
					`"disk" = {
										"size" = 2096129
										"name" = "too big"  
									}`),
				ExpectError: regexp.MustCompile(`config is invalid: skytap_vm.bar: expected disk.0.size to be in the range \(2048 - 2096128\), got 2096129`),
			},
		},
	})
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

func testAccCheckSkytapVMDiskResource(t *testing.T, name string, disks string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsVM, err := getResource(s, name)
		assert.NotNil(t, rsVM)
		if err != nil {
			return err
		}
		assert.Equal(t, disks, rsVM.Primary.Attributes["disk.#"])
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
