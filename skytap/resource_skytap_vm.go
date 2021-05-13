package skytap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapVM() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSkytapVMCreate,
		ReadContext:   resourceSkytapVMRead,
		UpdateContext: resourceSkytapVMUpdate,
		DeleteContext: resourceSkytapVMDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

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
							ValidateFunc: validation.IsIPAddress,
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
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"label": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"category": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceSkytapVMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	environmentID := d.Get("environment_id").(string)
	client := meta.(*SkytapClient).vmsClient

	// Give it some more breathing space. Might reject request if straight after a destroy.
	if err := waitForEnvironmentReady(ctx, d, meta, environmentID, schema.TimeoutCreate); err != nil {
		return diag.FromErr(err)
	}

	id, err := vmCreate(ctx, d, meta, environmentID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err = waitForVMStopped(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}
	if err = waitForEnvironmentReady(ctx, d, meta, environmentID, schema.TimeoutCreate); err != nil {
		return diag.FromErr(err)
	}

	// create network interfaces if necessary
	if _, ok := d.GetOk("network_interface"); ok {
		vmNetworks, err := addNetworkAdapters(ctx, d, meta, id)
		if err != nil {
			return diag.FromErr(err)
		}

		err = d.Set("network_interface", vmNetworks)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	vmDisks, err := addVMHardware(ctx, d, meta, environmentID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("disk", vmDisks); err != nil {
		log.Printf("[ERROR] error flattening disks: %v", err)
		return diag.FromErr(err)
	}

	if userData, ok := d.GetOk("user_data"); ok {
		if err := client.UpdateUserData(ctx, environmentID, id, utils.String(userData.(string))); err != nil {
			return diag.FromErr(err)
		}
	}

	if label, ok := d.GetOk("label"); ok {
		labels := vmCreateLabels(label.(*schema.Set))
		for _, l := range labels {
			if err = client.CreateLabel(ctx, environmentID, id, l); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err = forceRunning(ctx, meta, environmentID, id); err != nil {
		return diag.FromErr(err)
	}

	stateConfUpdate := &resource.StateChangeConf{
		Pending:    getVMPendingUpdateRunstates(true),
		Target:     getVMTargetUpdateRunstates(true),
		Refresh:    vmRunstateRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for VM (%s) to complete", d.Id())
	_, err = stateConfUpdate.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for VM (%s) to complete: %s", d.Id(), err)
	}

	return resourceSkytapVMRead(ctx, d, meta)
}

func resourceSkytapVMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).vmsClient

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

		return diag.Errorf("error retrieving VM (%s): %v", id, err)
	}

	// templateID and vmID are not set, as they are not returned by the VM response.
	// If any of these attributes are changed, this VM will be rebuilt.
	err = d.Set("environment_id", environmentID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("name", vm.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("cpus", vm.Hardware.CPUs)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("ram", vm.Hardware.RAM)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("max_cpus", vm.Hardware.MaxCPUs)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("max_ram", vm.Hardware.MaxRAM)
	if err != nil {
		return diag.FromErr(err)
	}

	userData, err := client.GetUserData(ctx, environmentID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("user_data", userData)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("label", flattenLabels(vm.Labels)); err != nil {
		return diag.FromErr(err)
	}

	if len(vm.Interfaces) > 0 {
		// add the names
		networkInterfaceSet := d.Get("network_interface").(*schema.Set)
		for _, networkInterface := range networkInterfaceSet.List() {
			networkInterfaceMap := networkInterface.(map[string]interface{})
			vmInterface, err := getVMNetworkInterface(networkInterfaceMap["id"].(string), vm)
			if err != nil {
				return diag.FromErr(err)
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
		networkSetFlattened := flattenNetworkInterfaces(vm.Interfaces)

		if err := d.Set("network_interface", networkSetFlattened); err != nil {
			log.Printf("[ERROR] error flattening network interfaces: %v", err)
			return diag.FromErr(err)
		}
		ports, ips := buildServices(networkSetFlattened)
		err = d.Set("service_ports", ports)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set("service_ips", ips)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if len(vm.Hardware.Disks) > 0 {
		err = d.Set("os_disk_size", *vm.Hardware.Disks[0].Size)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	for _, disk := range vm.Hardware.Disks {
		log.Printf("[INFO] disks: %#v, %#v", disk.Name, disk.Size)
	}
	if len(vm.Hardware.Disks) > 1 {
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
		diskSetFlattened := flattenDisks(vm.Hardware.Disks)

		if err := d.Set("disk", diskSetFlattened); err != nil {
			log.Printf("[ERROR] error flattening disks: %v", err)
			return diag.FromErr(err)
		}
	}
	log.Printf("[INFO] retrieved VM: %s", id)
	log.Printf("[TRACE] retrieved VM: %v", spew.Sdump(vm))

	return nil
}

func resourceSkytapVMUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).vmsClient

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	opts := skytap.UpdateVMRequest{}

	if v, ok := d.GetOk("name"); ok && d.HasChange("name") {
		opts.Name = utils.String(v.(string))
	}

	hardware, err := updateHardware(d)
	if err != nil {
		return diag.FromErr(err)
	}
	opts.Hardware = hardware

	var vmDisks []interface{}
	if d.HasChange("disk") || d.HasChange("name") || d.HasChange("ram") ||
		d.HasChange("cpus") || d.HasChange("os_disk_size") {

		log.Printf("[INFO] VM update: %s", id)
		log.Printf("[TRACE] VM update options: %v", spew.Sdump(opts))
		vm, err := client.Update(ctx, environmentID, id, &opts)
		if err != nil {
			return diag.Errorf("error updating vm (%s): %v", id, err)
		}

		log.Printf("[INFO] updated VM: %s", id)
		log.Printf("[TRACE] updated VM: %v", spew.Sdump(vm))

		// Have to do this here in order to capture `name`
		vmDisks = flattenDisks(vm.Hardware.Disks)

		if err := d.Set("disk", vmDisks); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("user_data") {
		if userData, ok := d.GetOk("user_data"); ok {
			if err := client.UpdateUserData(ctx, environmentID, id, utils.String(userData.(string))); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("label") {
		old, new := d.GetChange("label")
		remove := old.(*schema.Set).Difference(new.(*schema.Set))
		add := new.(*schema.Set).Difference(old.(*schema.Set))

		for _, l := range remove.List() {
			label := l.(map[string]interface{})
			if err = client.DeleteLabel(ctx, environmentID, id, label["id"].(string)); err != nil {
				return diag.FromErr(err)
			}
		}
		labelsToAdd := vmCreateLabels(add)
		for _, l := range labelsToAdd {
			if err = client.CreateLabel(ctx, environmentID, id, l); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    getVMPendingUpdateRunstates(false),
		Target:     getVMTargetUpdateRunstates(false),
		Refresh:    vmRunstateRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for VM (%s) to complete", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for VM (%s) to complete: %s", d.Id(), err)
	}

	return resourceSkytapVMRead(ctx, d, meta)
}

func resourceSkytapVMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).vmsClient

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	log.Printf("[INFO] destroying VM ID: %s", id)
	err := client.Delete(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] VM (%s) was not found - assuming removed", id)
			return nil
		}

		return diag.Errorf("error deleting VM (%s): %v", id, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"false"},
		Target:     []string{"true"},
		Refresh:    vmDeleteRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for VM (%s) to complete", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for VM (%s) to complete: %s", d.Id(), err)
	}
	if err = waitForEnvironmentReady(ctx, d, meta, environmentID, schema.TimeoutDelete); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] destroyed VM ID: %s", id)

	return nil
}

func addNetworkAdapters(ctx context.Context, d *schema.ResourceData, meta interface{}, vmID string) (interface{}, error) {
	client := meta.(*SkytapClient).interfacesClient
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
			log.Printf("[TRACE] creating interface: %v", spew.Sdump(nicType))
			networkInterface, err := client.Create(ctx, environmentID, vmID, &nicType)
			if err != nil {
				return nil, fmt.Errorf("error creating interface: %v", err)
			}
			id = *networkInterface.ID

			log.Printf("[INFO] created interface: %s", id)
			log.Printf("[TRACE] created interface: %v", spew.Sdump(networkInterface))
		}
		{
			log.Printf("[INFO] attaching interface: %s", id)
			log.Printf("[TRACE] attaching interface: %v", spew.Sdump(networkID))
			_, err := client.Attach(ctx, environmentID, vmID, id, &networkID)
			if err != nil {
				return nil, fmt.Errorf("error attaching interface: %v", err)
			}

			log.Printf("[INFO] attached interface: %s", id)
			log.Printf("[TRACE] attached interface: %v", spew.Sdump(networkInterface))
		}
		{
			// if the user define a hostname or ip we need an interface update.
			if requiresUpdate {
				log.Printf("[INFO] updating interface: %s", id)
				log.Printf("[TRACE] updating interface options: %v", spew.Sdump(opts))
				vmInterface, err := client.Update(ctx, environmentID, vmID, id, &opts)
				if err != nil {
					return nil, fmt.Errorf("error updating interface: %v", err)
				}
				vmNetworkInterfaces[idx] = *vmInterface
				log.Printf("[INFO] updated interface: %s", id)
				log.Printf("[TRACE] updated interface: %v", spew.Sdump(networkInterface))
			}
		}
		if _, ok := networkInterfaceMap["published_service"]; ok {
			// create network interfaces if necessary
			err := addPublishedServices(ctx, meta, environmentID, vmID, id, networkInterfaceMap, &vmNetworkInterfaces[idx])
			if err != nil {
				return nil, err
			}
		}
	}
	// Have to do this here in order to capture `published_service` name
	return flattenNetworkInterfaces(vmNetworkInterfaces), nil
}

// create the public service for a specific interface
func addPublishedServices(ctx context.Context, meta interface{}, environmentID string, vmID string, nicID string, networkInterface map[string]interface{},
	vmInterface *skytap.Interface) error {
	client := meta.(*SkytapClient).publishedServicesClient
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
		log.Printf("[TRACE] creating published service: %v", spew.Sdump(internalPort))
		createdService, err := client.Create(ctx, environmentID, vmID, nicID, &internalPort)
		if err != nil {
			return fmt.Errorf("error creating published service: %v", err)
		}

		log.Printf("[INFO] created published service: %s", *createdService.ID)
		log.Printf("[TRACE] created published service: %v", spew.Sdump(createdService))

		// Have to do this here in order to capture `published_service` name
		createdService.Name = utils.String(publishedServiceMap["name"].(string))
		vmInterface.Services[idx] = *createdService
	}
	return nil
}

func addVMHardware(ctx context.Context, d *schema.ResourceData, meta interface{}, environmentID string, vmID string) (interface{}, error) {
	client := meta.(*SkytapClient).vmsClient

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

	// Check vCPUs does not exceed RAM
	var ramGBs = mbToGb(*vm.Hardware.RAM)
	if opts.Hardware.RAM != nil {
		ramGBs = mbToGb(*opts.Hardware.RAM)
	}
	var vCPUs = *vm.Hardware.CPUs
	if opts.Hardware.CPUs != nil {
		vCPUs = *opts.Hardware.CPUs
	}
	if vCPUs > ramGBs {
		return nil, cpusExceedsRamError(vCPUs, ramGBs)
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
	log.Printf("[TRACE] VM create update options: %v", spew.Sdump(opts))
	vmUpdated, err := client.Update(ctx, environmentID, *vm.ID, &opts)
	if err != nil {
		return nil, fmt.Errorf("error updating vm (%s): %v", *vm.ID, err)
	}
	log.Printf("[INFO] updated VM after create: %s", *vm.ID)
	log.Printf("[TRACE] updated VM after create: %v", spew.Sdump(vmUpdated))

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
	if cpus, ok := d.GetOk("cpus"); ok && d.HasChange("cpus") {
		hardware.CPUs = utils.Int(cpus.(int))
		if maxCPUs, maxOK := d.GetOk("max_cpus"); maxOK {
			if *hardware.CPUs > maxCPUs.(int) {
				return nil, outOfRangeError("cpus", *hardware.CPUs, maxCPUs.(int))
			}
		} else {
			return nil, fmt.Errorf("unable to read the 'max_cpus' element")
		}

		if ram, ok := d.GetOk("ram"); ok && cpus.(int) > mbToGb(ram.(int)) {
			return nil, cpusExceedsRamError(cpus.(int), mbToGb(ram.(int)))
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

func mbToGb(mb int) int {
	return mb / 1024
}

func cpusExceedsRamError(cpus, ram int) error {
	return fmt.Errorf("the 'cpus' argument has been assigned (%d) which is more than the maximum allowed (%d), the number of GB of RAM", cpus, ram)
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

var vmPendingCreateRunstates = []string{
	string(skytap.VMRunstateBusy),
}

var vmTargetCreateRunstates = []string{
	string(skytap.VMRunstateStopped),
}

var vmPendingUpdateRunstates = []string{
	string(skytap.VMRunstateBusy),
}

var vmPendingUpdateRunstateAfterCreate = []string{
	string(skytap.VMRunstateBusy),
	string(skytap.VMRunstateStopped),
}

var vmTargetUpdateRunstateAfterCreate = []string{
	string(skytap.VMRunstateRunning),
}

var vmTargetUpdateRunstates = []string{
	string(skytap.VMRunstateRunning),
	string(skytap.VMRunstateStopped),
	string(skytap.VMRunstateReset),
	string(skytap.VMRunstateSuspended),
	string(skytap.VMRunstateHalted),
}

func getVMPendingUpdateRunstates(running bool) []string {
	if running {
		return vmPendingUpdateRunstateAfterCreate
	}
	return vmPendingUpdateRunstates
}

func getVMTargetUpdateRunstates(running bool) []string {
	if running {
		return vmTargetUpdateRunstateAfterCreate
	}
	return vmTargetUpdateRunstates
}

func vmRunstateRefreshFunc(
	ctx context.Context, d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).vmsClient

		id := d.Id()
		environmentID := d.Get("environment_id").(string)

		log.Printf("[DEBUG] retrieving VM: %s", id)
		vm, err := client.Get(ctx, environmentID, id)

		if err != nil {
			return nil, "", fmt.Errorf("error retrieving VM (%s) when waiting: (%s)", id, err)
		}

		log.Printf("[DEBUG] VM status (%s): %s", id, *vm.Runstate)

		return vm, string(*vm.Runstate), nil
	}
}

func waitForVMStopped(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	stateConf := &resource.StateChangeConf{
		Pending:    vmPendingCreateRunstates,
		Target:     vmTargetCreateRunstates,
		Refresh:    vmRunstateRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for VM (%s) to complete", d.Id())
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for VM (%s) to complete: %s", d.Id(), err)
	}
	return nil
}

func vmDeleteRefreshFunc(
	ctx context.Context, d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).vmsClient

		id := d.Id()
		environmentID := d.Get("environment_id").(string)

		log.Printf("[DEBUG] retrieving VM: %s", id)
		vm, err := client.Get(ctx, environmentID, id)

		var removed = "false"
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				log.Printf("[DEBUG] VM (%s) has been removed.", id)
				removed = "true"
			} else {
				return nil, "", fmt.Errorf("error retrieving VM (%s) when waiting: (%s)", id, err)
			}
		}

		return vm, removed, nil
	}
}

func vmCreate(ctx context.Context, d *schema.ResourceData, meta interface{}, environmentID string) (string, error) {
	client := meta.(*SkytapClient).vmsClient

	templateID := d.Get("template_id").(string)
	templateVMID := d.Get("vm_id").(string)

	// create the VM
	createOpts := skytap.CreateVMRequest{
		TemplateID: templateID,
		VMID:       templateVMID,
	}

	log.Printf("[INFO] VM create")
	log.Printf("[TRACE] VM create options: %v", spew.Sdump(createOpts))
	vm, err := client.Create(ctx, environmentID, &createOpts)
	if err != nil {
		return "", fmt.Errorf("error creating VM: %v with template ID: %s and VM ID: %s", err, createOpts.TemplateID, createOpts.VMID)
	}
	log.Printf("[INFO] created VM: %s", *vm.ID)
	log.Printf("[TRACE] created VM: %v", spew.Sdump(vm))

	return *vm.ID, nil
}

func forceRunning(ctx context.Context, meta interface{}, environmentID string, id string) error {
	client := meta.(*SkytapClient).vmsClient

	opts := skytap.UpdateVMRequest{}
	opts.Runstate = utils.VMRunstate(skytap.VMRunstateRunning)
	log.Printf("[INFO] VM starting: %s", id)
	log.Printf("[TRACE] VM starting: %v", spew.Sdump(opts))
	vm, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return fmt.Errorf("error starting VM (%s): %v", id, err)
	}
	log.Printf("[INFO] started VM: %s", id)
	log.Printf("[TRACE] started VM: %v", spew.Sdump(vm))
	return nil
}

func vmCreateLabels(vs *schema.Set) []*skytap.CreateVMLabelRequest {
	createLabelsRequest := make([]*skytap.CreateVMLabelRequest, vs.Len())
	for i, v := range vs.List() {
		elem := v.(map[string]interface{})
		createLabelsRequest[i] = &skytap.CreateVMLabelRequest{
			Category: utils.String(elem["category"].(string)),
			Value:    utils.String(elem["value"].(string)),
		}
	}
	return createLabelsRequest
}
