package skytap

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapVMCreate,
		Read:   resourceSkytapVMRead,
		Update: resourceSkytapVMUpdate,
		Delete: resourceSkytapVMDelete,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"template_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"vm_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"cpus": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 12),
			},

			"max_cpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ram": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(256, 131072),
			},

			"max_ram": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"os_disk_size": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntBetween(2048, 2096128),
			},

			"disk": {
				Type:     schema.TypeSet,
				Set:      diskHash,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(2048, 2096128),
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"controller": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lun": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"network_interface": {
				Type:     schema.TypeSet,
				Set:      networkInterfaceHash,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interface_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateNICType(),
						},
						"network_id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.SingleIP(),
						},
						"hostname": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"published_service": {
							Type:     schema.TypeSet,
							Set:      publishedServiceHash,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"internal_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"external_ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"external_port": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"service_ips": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"service_ports": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func resourceSkytapVMCreate(d *schema.ResourceData, meta interface{}) error {
	environmentID := d.Get("environment_id").(string)

	id, err := vmCreate(d, meta, environmentID)
	if err != nil {
		return err
	}
	d.SetId(id)

	// create network interfaces if necessary
	var vmNetworks interface{}
	if _, ok := d.GetOk("network_interface"); ok {
		vmNetworks, err = addNetworkAdapters(d, meta, id)
		if err != nil {
			return err
		}
	}
	vmDisks, err := addVMHardware(d, meta, environmentID, id)
	if err != nil {
		return err
	}

	if err = forceRunning(meta, environmentID, id); err != nil {
		return err
	}

	return resourceSkytapVMReadAfterCreateUpdate(d, meta, vmNetworks, vmDisks)
}

func resourceSkytapVMRead(d *schema.ResourceData, meta interface{}) error {
	return resourceSkytapVMReadAfterCreateUpdate(d, meta, nil, nil)
}

func resourceSkytapVMReadAfterCreateUpdate(d *schema.ResourceData, meta interface{}, vmNetworks interface{}, vmDisks interface{}) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	log.Printf("[INFO] retrieving VM with ID: %s", id)
	vm, err := client.Get(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] VM (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error retrieving VM (%s): %v", id, err)
	}

	// templateID and vmID are not set, as they are not returned by the VM response.
	// If any of these attributes are changed, this VM will be rebuilt.
	d.Set("environment_id", environmentID)
	d.Set("name", vm.Name)
	d.Set("cpus", vm.Hardware.CPUs)
	d.Set("ram", vm.Hardware.RAM)
	d.Set("max_cpus", vm.Hardware.MaxCPUs)
	d.Set("max_ram", vm.Hardware.MaxRAM)

	if len(vm.Interfaces) > 0 {
		var networkSetFlattened *schema.Set
		if vmNetworks == nil {
			// add the names
			networkInterfaceSet := d.Get("network_interface").(*schema.Set)
			for _, networkInterface := range networkInterfaceSet.List() {
				networkInterfaceMap := networkInterface.(map[string]interface{})
				vmInterface, err := getVMNetworkInterface(networkInterfaceMap["id"].(string), vm)
				if err != nil {
					return err
				}
				if _, ok := networkInterfaceMap["published_service"]; ok {
					publishedServiceSet := networkInterfaceMap["published_service"].(*schema.Set)
					for _, publishedService := range publishedServiceSet.List() {
						publishedServiceMap := publishedService.(map[string]interface{})
						for idx := range vmInterface.Services {
							if *vmInterface.Services[idx].InternalPort == publishedServiceMap["internal_port"].(int) {
								vmInterface.Services[idx].Name = utils.String(publishedServiceMap["name"].(string))
								break
							}
						}
					}
				}
			}
			networkSetFlattened = flattenNetworkInterfaces(vm.Interfaces)
		} else {
			networkSetFlattened = vmNetworks.(*schema.Set)
		}

		if err := d.Set("network_interface", networkSetFlattened); err != nil {
			log.Printf("[ERROR] error flattening network interfaces: %v", err)
			return err
		}
		ports, ips := buildServices(networkSetFlattened)
		err = d.Set("service_ports", ports)
		if err != nil {
			return err
		}
		err = d.Set("service_ips", ips)
		if err != nil {
			return err
		}
	}

	if len(vm.Hardware.Disks) > 0 {
		d.Set("os_disk_size", *vm.Hardware.Disks[0].Size)
	}

	if len(vm.Hardware.Disks) > 1 {
		var diskSetFlattened *schema.Set
		if vmDisks == nil {
			// add the names
			diskSet := d.Get("disk").(*schema.Set)
			for _, disk := range diskSet.List() {
				diskMap := disk.(map[string]interface{})
				for idx := range vm.Hardware.Disks {
					if *vm.Hardware.Disks[idx].ID == diskMap["id"].(string) {
						vm.Hardware.Disks[idx].Name = utils.String(diskMap["name"].(string))
						break
					}
				}
			}
			diskSetFlattened = flattenDisks(vm.Hardware.Disks)
		} else {
			diskSetFlattened = vmDisks.(*schema.Set)
		}

		if err := d.Set("disk", diskSetFlattened); err != nil {
			log.Printf("[ERROR] error flattening disks: %v", err)
			return err
		}
	}
	log.Printf("[INFO] retrieved VM: %s", id)
	log.Printf("[DEBUG] retrieved VM: %#v", spew.Sdump(vm))

	return nil
}

func resourceSkytapVMUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	opts := skytap.UpdateVMRequest{}

	if v, ok := d.GetOk("name"); ok && d.HasChange("name") {
		opts.Name = utils.String(v.(string))
	}

	hardware, err := updateHardware(d)
	if err != nil {
		return err
	}
	opts.Hardware = hardware

	log.Printf("[INFO] VM update: %s", id)
	log.Printf("[DEBUG] VM update options: %#v", spew.Sdump(opts))
	vm, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return fmt.Errorf("error updating vm (%s): %v", id, err)
	}
	log.Printf("[INFO] updated VM: %s", id)
	log.Printf("[DEBUG] updated VM: %#v", spew.Sdump(vm))

	// Have to do this here in order to capture `name`
	vmDisks := flattenDisks(vm.Hardware.Disks)

	return resourceSkytapVMReadAfterCreateUpdate(d, meta, nil, vmDisks)
}

func resourceSkytapVMDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	log.Printf("[INFO] destroying VM ID: %s", id)
	err := client.Delete(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] VM (%s) was not found - assuming removed", id)
			return nil
		}

		return fmt.Errorf("error deleting VM (%s): %v", id, err)
	}

	return err
}

func addNetworkAdapters(d *schema.ResourceData, meta interface{}, vmID string) (interface{}, error) {
	client := meta.(*SkytapClient).interfacesClient
	ctx := meta.(*SkytapClient).StopContext
	environmentID := d.Get("environment_id").(string)
	networkIfaceCount := d.Get("network_interface.#").(int)

	// In case there is network interface defined
	// we remove the default networks from the VM before create the network defined
	if networkIfaceCount > 0 {
		vmIfaces, err := client.List(ctx, environmentID, vmID)
		if err != nil {
			return nil, fmt.Errorf("error resolving VM network interfaces: %v", err)
		}
		for _, iface := range vmIfaces.Value {
			log.Printf("[INFO] deleting network interface: %s", *iface.ID)
			err = client.Delete(ctx, environmentID, vmID, *iface.ID)
			if err != nil {
				return nil, fmt.Errorf("error removing the default interface from VM: %v", err)
			}
			log.Printf("[INFO] deleted network interface: %s", *iface.ID)
		}
	}
	networkInterfaceSet := d.Get("network_interface").(*schema.Set)
	vmNetworkInterfaces := make([]skytap.Interface, networkInterfaceSet.Len())
	log.Printf("[INFO] creating %d network interfaces", networkInterfaceSet.Len())
	for idx, networkInterface := range networkInterfaceSet.List() {
		networkInterfaceMap := networkInterface.(map[string]interface{})
		nicType := skytap.CreateInterfaceRequest{
			NICType: utils.NICType(skytap.NICType(networkInterfaceMap["interface_type"].(string))),
		}
		networkID := skytap.AttachInterfaceRequest{
			NetworkID: utils.String(networkInterfaceMap["network_id"].(string)),
		}

		opts := skytap.UpdateInterfaceRequest{}
		requiresUpdate := false
		if v, ok := networkInterfaceMap["ip"]; ok && networkInterfaceMap["ip"] != "" {
			opts.IP = utils.String(v.(string))
			requiresUpdate = true
		}
		if v, ok := networkInterfaceMap["hostname"]; ok && networkInterfaceMap["hostname"] != "" {
			opts.Hostname = utils.String(v.(string))
			requiresUpdate = true
		}

		var id string
		{
			log.Printf("[INFO] creating interface")
			log.Printf("[DEBUG] creating interface: %#v", spew.Sdump(nicType))
			networkInterface, err := client.Create(ctx, environmentID, vmID, &nicType)
			if err != nil {
				return nil, fmt.Errorf("error creating interface: %v", err)
			}
			id = *networkInterface.ID

			log.Printf("[INFO] created interface: %s", id)
			log.Printf("[DEBUG] created interface: %#v", spew.Sdump(networkInterface))
		}
		{
			log.Printf("[INFO] attaching interface: %s", id)
			log.Printf("[DEBUG] attaching interface: %#v", spew.Sdump(networkID))
			_, err := client.Attach(ctx, environmentID, vmID, id, &networkID)
			if err != nil {
				return nil, fmt.Errorf("error attaching interface: %v", err)
			}

			log.Printf("[INFO] attached interface: %s", id)
			log.Printf("[DEBUG] attached interface: %#v", spew.Sdump(networkInterface))
		}
		{
			// if the user define a hostname or ip we need an interface update.
			if requiresUpdate {
				log.Printf("[INFO] updating interface: %s", id)
				log.Printf("[DEBUG] updating interface options: %#v", spew.Sdump(opts))
				vmInterface, err := client.Update(ctx, environmentID, vmID, id, &opts)
				if err != nil {
					return nil, fmt.Errorf("error updating interface: %v", err)
				}
				vmNetworkInterfaces[idx] = *vmInterface
				log.Printf("[INFO] updated interface: %s", id)
				log.Printf("[DEBUG] updated interface: %#v", spew.Sdump(networkInterface))
			}
		}
		if _, ok := networkInterfaceMap["published_service"]; ok {
			// create network interfaces if necessary
			err := addPublishedServices(meta, environmentID, vmID, id, networkInterfaceMap, &vmNetworkInterfaces[idx])
			if err != nil {
				return nil, err
			}
		}
	}
	// Have to do this here in order to capture `published_service` name
	return flattenNetworkInterfaces(vmNetworkInterfaces), nil
}

// create the public service for a specific interface
func addPublishedServices(meta interface{}, environmentID string, vmID string, nicID string, networkInterface map[string]interface{},
	vmInterface *skytap.Interface) error {
	client := meta.(*SkytapClient).publishedServicesClient
	ctx := meta.(*SkytapClient).StopContext
	publishedServiceSet := networkInterface["published_service"].(*schema.Set)
	vmInterface.Services = make([]skytap.PublishedService, publishedServiceSet.Len())
	log.Printf("[INFO] creating %d published services", publishedServiceSet.Len())
	for idx, publishedService := range publishedServiceSet.List() {
		publishedServiceMap := publishedService.(map[string]interface{})
		// create
		internalPort := skytap.CreatePublishedServiceRequest{
			InternalPort: utils.Int(publishedServiceMap["internal_port"].(int)),
		}
		log.Printf("[INFO] creating published service")
		log.Printf("[DEBUG] creating published service: %#v", spew.Sdump(internalPort))
		createdService, err := client.Create(ctx, environmentID, vmID, nicID, &internalPort)
		if err != nil {
			return fmt.Errorf("error creating published service: %v", err)
		}

		log.Printf("[INFO] created published service: %s", *createdService.ID)
		log.Printf("[DEBUG] created published service: %#v", spew.Sdump(createdService))

		// Have to do this here in order to capture `published_service` name
		createdService.Name = utils.String(publishedServiceMap["name"].(string))
		vmInterface.Services[idx] = *createdService
	}
	return nil
}

func addVMHardware(d *schema.ResourceData, meta interface{}, environmentID string, vmID string) (interface{}, error) {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	vm, err := client.Get(ctx, environmentID, vmID)
	if err != nil {
		return nil, err
	}

	opts := skytap.UpdateVMRequest{
		Hardware: &skytap.UpdateHardware{
			UpdateDisks: &skytap.UpdateDisks{},
		},
	}

	if v, ok := d.GetOk("name"); ok {
		opts.Name = utils.String(v.(string))
	}
	if v, ok := d.GetOk("cpus"); ok {
		opts.Hardware.CPUs = utils.Int(v.(int))
		if *opts.Hardware.CPUs > *vm.Hardware.MaxCPUs {
			return nil, outOfRangeError("cpus", *opts.Hardware.CPUs, *vm.Hardware.MaxCPUs)
		}
	}
	if v, ok := d.GetOk("ram"); ok {
		opts.Hardware.RAM = utils.Int(v.(int))
		if *opts.Hardware.RAM > *vm.Hardware.MaxRAM {
			return nil, outOfRangeError("ram", *opts.Hardware.RAM, *vm.Hardware.MaxRAM)
		}
	}
	if v, ok := d.GetOk("os_disk_size"); ok {
		sizeNew := v.(int)
		opts.Hardware.UpdateDisks.OSSize = utils.Int(sizeNew)
		err := checkDiskNotShrunk(*vm.Hardware.Disks[0].Size, sizeNew, "OS")
		if err != nil {
			return nil, err
		}
	}

	if v, ok := d.GetOk("disk"); ok {
		diskSet := v.(*schema.Set)
		log.Printf("[INFO] creating %d disks", diskSet.Len())
		opts.Hardware.UpdateDisks.NewDisks = make([]int, d.Get("disk.#").(int))
		opts.Hardware.UpdateDisks.DiskIdentification = make([]skytap.DiskIdentification, d.Get("disk.#").(int))
		for idx, disk := range diskSet.List() {
			diskMap := disk.(map[string]interface{})
			opts.Hardware.UpdateDisks.NewDisks[idx] = diskMap["size"].(int)
			opts.Hardware.UpdateDisks.DiskIdentification[idx] = skytap.DiskIdentification{
				ID: nil, Name: utils.String(diskMap["name"].(string)), Size: utils.Int(diskMap["size"].(int)),
			}
		}
	} else {
		opts.Hardware.UpdateDisks.DiskIdentification = make([]skytap.DiskIdentification, 0)
	}

	log.Printf("[INFO] VM create update: %s", *vm.ID)
	log.Printf("[DEBUG] VM create update options: %#v", spew.Sdump(opts))
	vmUpdated, err := client.Update(ctx, environmentID, *vm.ID, &opts)
	if err != nil {
		return nil, fmt.Errorf("error updating vm (%s): %v", *vm.ID, err)
	}
	log.Printf("[INFO] updated VM after create: %s", *vm.ID)
	log.Printf("[DEBUG] updated VM after create: %#v", spew.Sdump(vmUpdated))

	// Have to do this in order to capture `name`
	return flattenDisks(vmUpdated.Hardware.Disks), nil
}

func outOfRangeError(field string, value int, max int) error {
	return fmt.Errorf("the '%s' argument has been assigned (%d) which is more "+
		"than the maximum allowed (%d) as defined by this VM",
		field, value, max)
}

func updateHardware(d *schema.ResourceData) (*skytap.UpdateHardware, error) {
	var hardware = &skytap.UpdateHardware{
		UpdateDisks: &skytap.UpdateDisks{},
	}
	if _, ok := d.GetOk("disk"); ok {
		oldDisks, newDisks := d.GetChange("disk")
		diskSet := newDisks.(*schema.Set)
		diskIDs := make([]skytap.DiskIdentification, 0)
		disksNew := make([]int, 0)
		// adds and initialises disk identification struct
		for _, disk := range diskSet.List() {
			diskMap := disk.(map[string]interface{})
			name := diskMap["name"].(string)
			sizeNew := diskMap["size"].(int)
			id, sizeOld := retrieveIDsFromOldState(oldDisks.(*schema.Set), name)
			if id == "" { // new
				disksNew = append(disksNew, sizeNew)
			} else {
				err := checkDiskNotShrunk(sizeOld, sizeNew, name)
				if err != nil {
					return nil, err
				}
			}
			diskID := skytap.DiskIdentification{
				ID: utils.String(id), Name: utils.String(name), Size: utils.Int(sizeNew),
			}
			diskIDs = append(diskIDs, diskID)
		}
		hardware.UpdateDisks.DiskIdentification = diskIDs
		if len(disksNew) > 0 {
			log.Printf("[INFO] creating %d disk(s)", len(disksNew))
			hardware.UpdateDisks.NewDisks = disksNew
		}
	} else {
		hardware.UpdateDisks.DiskIdentification = make([]skytap.DiskIdentification, 0)
	}

	if cpus, ok := d.GetOk("cpus"); ok && d.HasChange("cpus") {
		hardware.CPUs = utils.Int(cpus.(int))
		if maxCPUs, maxOK := d.GetOk("max_cpus"); maxOK {
			if *hardware.CPUs > maxCPUs.(int) {
				return nil, outOfRangeError("cpus", *hardware.CPUs, maxCPUs.(int))
			}
		} else {
			return nil, fmt.Errorf("unable to read the 'max_cpus' element")
		}
	}
	if ram, ok := d.GetOk("ram"); ok && d.HasChange("ram") {
		hardware.RAM = utils.Int(ram.(int))
		if maxRAM, maxOK := d.GetOk("max_ram"); maxOK {
			if *hardware.RAM > maxRAM.(int) {
				return nil, outOfRangeError("ram", *hardware.RAM, maxRAM.(int))
			}
		} else {
			return nil, fmt.Errorf("unable to read the 'max_ram' element")
		}
	}

	if _, ok := d.GetOk("os_disk_size"); ok && d.HasChange("os_disk_size") {
		sizeOld, sizeNew := d.GetChange("os_disk_size")
		sizeOldInt := sizeOld.(int)
		sizeNewInt := sizeNew.(int)
		hardware.UpdateDisks.OSSize = utils.Int(sizeNewInt)
		err := checkDiskNotShrunk(sizeOldInt, sizeNewInt, "OS")
		if err != nil {
			return nil, err
		}
	}

	return hardware, nil
}

// Confirm size not shrunk
func checkDiskNotShrunk(sizeOld int, sizeNew int, name string) error {
	if sizeOld > sizeNew {
		return fmt.Errorf("cannot shrink volume (%s) from size (%d) to size (%d)", name, sizeOld, sizeNew)
	}
	return nil
}

func retrieveIDsFromOldState(d *schema.Set, name string) (string, int) {
	for _, disk := range d.List() {
		diskMap := disk.(map[string]interface{})
		if diskMap["name"] == name {
			return diskMap["id"].(string), diskMap["size"].(int)
		}
	}
	return "", 0
}

func vmCreate(d *schema.ResourceData, meta interface{}, environmentID string) (string, error) {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	templateID := d.Get("template_id").(string)
	templateVMID := d.Get("vm_id").(string)

	// create the VM
	createOpts := skytap.CreateVMRequest{
		TemplateID: templateID,
		VMID:       templateVMID,
	}

	log.Printf("[INFO] VM create")
	log.Printf("[DEBUG] VM create options: %#v", spew.Sdump(createOpts))
	vm, err := client.Create(ctx, environmentID, &createOpts)
	if err != nil {
		return "", fmt.Errorf("error creating VM: %v with options: %#v", err, spew.Sdump(createOpts))
	}
	log.Printf("[INFO] created VM: %s", *vm.ID)
	log.Printf("[DEBUG] created VM: %#v", spew.Sdump(vm))

	return *vm.ID, nil
}

func forceRunning(meta interface{}, environmentID string, id string) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext
	opts := skytap.UpdateVMRequest{}
	opts.Runstate = utils.VMRunstate(skytap.VMRunstateRunning)
	log.Printf("[INFO] VM starting: %s", id)
	log.Printf("[DEBUG] VM starting: %#v", spew.Sdump(opts))
	vm, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return fmt.Errorf("error starting VM (%s): %v", id, err)
	}
	log.Printf("[INFO] started VM: %s", id)
	log.Printf("[DEBUG] started VM: %#v", spew.Sdump(vm))
	return nil
}
