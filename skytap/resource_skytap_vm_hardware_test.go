package skytap

import (
	"fmt"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"
)

func TestAccSkytapVMCPURam_Create(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 8`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "8"),
					testAccCheckSkytapVMCPU(t, &vm, 8),
				),
			},
			{
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"ram" = 8192`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "8192"),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMRAM(t, &vmUpdated, 8192),
				),
			},
			{
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 4
									"ram" = 4096`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "4"),
					resource.TestCheckResourceAttr("skytap_vm.bar", "ram", "4096"),
					testAccCheckSkytapVMCPU(t, &vmUpdated, 4),
					testAccCheckSkytapVMRAM(t, &vmUpdated, 4096),
				),
			},
		},
	})
}

// To ensure the presence of a disk works unchanged
func TestAccSkytapVMCPU_DiskIntact(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 8
									"disk" = {
										"size" = 2048
										"name" = "disk1"
									}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "8"),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"disk1"}),
					testAccCheckSkytapVMCPU(t, &vm, 8),
					testAccCheckSkytapVMDisks(t, &vm, []int{2048}),
				),
			},
			{
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 4
									"disk" = {
										"size" = 2048
										"name" = "disk1"
									}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "cpus", "4"),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"disk1"}),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMCPU(t, &vmUpdated, 4),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{2048}),
				),
			},
		},
	})
}

func TestAccSkytapVMCPURAM_UpdateNPECheck(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
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
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 8
									"ram" = 2048`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
				),
			},
		},
	})
}

func TestAccSkytapVMCPURAM_Invalid(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 121
									"ram" = 819000000002`),
				ExpectError: regexp.MustCompile(`config is invalid: 2 problems:*`),
			},
		},
	})
}

func TestAccSkytapVMCPU_OutOfRange(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupNonDefaultEnvironment("SKYTAP_TEMPLATE_OUTOFRANGE_ID", "136409", "SKYTAP_VM_OUTOFRANGE_ID", "849656")
	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 12
									"ram" = 131072`),
				ExpectError: regexp.MustCompile(`the 'cpus' argument has been assigned \(12\) which is more than the maximum allowed \(8\) as defined by this VM`),
			},
		},
	})
}

func TestAccSkytapVMCPU_OutOfRangeAfterUpdate(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupNonDefaultEnvironment("SKYTAP_TEMPLATE_OUTOFRANGE_ID", "136409", "SKYTAP_VM_OUTOFRANGE_ID", "849656")
	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					``),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"cpus" = 12
									"ram" = 131072`),
				ExpectError: regexp.MustCompile(`the 'cpus' argument has been assigned \(12\) which is more than the maximum allowed \(8\) as defined by this VM`),
			},
		},
	})
}

func TestAccSkytapVMDisks_Create(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
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
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "3", []string{"smaller", "smaller2", "bigger"}),
					testAccCheckSkytapVMDisks(t, &vm, []int{2048, 2048, 2049}),
				),
			},
			{
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
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
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "2", []string{"smaller2", "bigger2"}),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{2048, 2049}),
				),
			},
		},
	})
}

// NPE checks
func TestAccSkytapVMDisks_UpdateNPECheck(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					``),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
				),
			},
			{
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"disk" = {
										"size" = 8000
										"name" = "smaller"
									}`),
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
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"disk" = {
										"size" = 8000
										"name" = "smaller"
									}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"smaller"}),
					testAccCheckSkytapVMDisks(t, &vm, []int{8000}),
				),
			},
			{
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"disk" = {
										"size" = 8192
										"name" = "smaller"
									}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMDiskResource(t, "skytap_vm.bar", "1", []string{"smaller"}),
					testAccCheckSkytapVMDisks(t, &vmUpdated, []int{8192}),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"disk" = {
										"size" = 6000
										"name" = "smaller"
									}`),
				ExpectError: regexp.MustCompile(`cannot shrink volume \(smaller\) from size \(8192\) to size \(6000\)`),
			},
		},
	})
}

func TestAccSkytapVMDisk_Invalid(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"disk" = {
										"size" = 2047
										"name" = "too small"
									}`),
				ExpectError: regexp.MustCompile(`config is invalid: skytap_vm.bar: expected disk.0.size to be in the range \(2048 - 2096128\), got 2047`),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"disk" = {
										"size" = 2096129
										"name" = "too big"
									}`),
				ExpectError: regexp.MustCompile(`config is invalid: skytap_vm.bar: expected disk.0.size to be in the range \(2048 - 2096128\), got 2096129`),
			},
		},
	})
}

func TestAccSkytapVMDisk_OS(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"os_disk_size" = 30721`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vm),
					resource.TestCheckResourceAttr("skytap_vm.bar", "os_disk_size", "30721"),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "max_ram"),
					resource.TestCheckResourceAttrSet("skytap_vm.bar", "max_cpus"),
					testAccCheckSkytapVMOSDisk(t, &vm, 30721),
				),
			},
			{
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"os_disk_size" = 30722`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "os_disk_size", "30722"),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					testAccCheckSkytapVMOSDisk(t, &vmUpdated, 30722),
				),
			},
			{
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"os_disk_size" = 3000`),
				ExpectError: regexp.MustCompile(`cannot shrink volume \(OS\) from size \(30722\) to size \(3000\)`),
			},
		},
	})
}

func TestAccSkytapVMDisk_OSChangeAfter(t *testing.T) {
	//t.Parallel()

	templateID, vmID, newEnvTemplateID := setupEnvironment()
	uniqueSuffixEnv := acctest.RandInt()
	var vm skytap.VM
	var vmUpdated skytap.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
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
				PreConfig: pause(),
				Config: testAccSkytapVMConfig_basic(newEnvTemplateID, uniqueSuffixEnv, "", templateID, vmID, "", "",
					`"os_disk_size" = 30721`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapVMExists("skytap_environment.foo", "skytap_vm.bar", &vmUpdated),
					testAccCheckSkytapVMUpdated(t, &vm, &vmUpdated),
					resource.TestCheckResourceAttr("skytap_vm.bar", "os_disk_size", "30721"),
				),
			},
		},
	})
}

func pause() func() {
	return func() { time.Sleep(time.Duration(2) * time.Minute) }
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

func testAccCheckSkytapVMDiskResource(t *testing.T, name string, disks string, names []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsVM, err := getResource(s, name)
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
