package skytap

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapVMCreate,
		Read:   resourceSkytapVMRead,
		Update: resourceSkytapVMUpdate,
		Delete: resourceSkytapVMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

			"network_interface": {
				Type:     schema.TypeList,
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
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.SingleIP(),
						},
						"hostname": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},

						"published_service": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"internal_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"external_ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"id": {
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
		},
	}
}

func resourceSkytapVMCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	templateID := d.Get("template_id").(string)
	templateVMID := d.Get("vm_id").(string)

	// create the VM
	createOpts := skytap.CreateVMRequest{
		TemplateID: templateID,
		VMID:       templateVMID,
	}

	log.Printf("[INFO] VM create options: %#v", spew.Sdump(createOpts))
	vm, err := client.Create(ctx, environmentID, &createOpts)
	if err != nil {
		return fmt.Errorf("error creating VM: %v", err)
	}

	if vm.ID == nil {
		return fmt.Errorf("VM ID is not set")
	}
	vmID := *vm.ID
	d.SetId(vmID)

	log.Printf("[INFO] created VM: %#v", spew.Sdump(vm))

	if err = waitForVMReady(d, meta); err != nil {
		return err
	}

	// create network interfaces if necessary
	if err = addNetworkAdapters(d, meta, *vm.ID); err != nil {
		return err
	}

	return updateVMResource(d, meta, true)
}

func waitForVMReady(d *schema.ResourceData, meta interface{}) error {
	stateConf := &resource.StateChangeConf{
		Pending:    vmPendingCreateRunstates,
		Target:     vmTargetCreateRunstates,
		Refresh:    vmRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for VM (%s) to complete", d.Id())
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for VM (%s) to complete: %s", d.Id(), err)
	}
	return nil
}

func addNetworkAdapters(d *schema.ResourceData, meta interface{}, vmID string) error {
	if _, ok := d.GetOk("network_interface"); ok {

		client := meta.(*SkytapClient).interfacesClient
		ctx := meta.(*SkytapClient).StopContext
		environmentID := d.Get("environment_id").(string)
		networkIfaceCount := d.Get("network_interface.#").(int)

		// In case there is network interface defined
		// we remove the default networks from the VM before create the network defined
		if networkIfaceCount > 0 {
			vmIfaces, err := client.List(ctx, environmentID, vmID)
			if err != nil {
				return fmt.Errorf("error resolving VM network interfaces: %v", err)
			}
			for _, iface := range vmIfaces.Value {
				log.Printf("[INFO] deleting network interface: %s", *iface.ID)
				err = client.Delete(ctx, environmentID, vmID, *iface.ID)
				if err != nil {
					return fmt.Errorf("error removing the default interface from VM: %v", err)
				}
				log.Printf("[INFO] deleted network interface: %s", *iface.ID)
			}
			if err = waitForVMReady(d, meta); err != nil {
				return err
			}
		}
		for i := 0; i < networkIfaceCount; i++ {
			networkInterface := d.Get(fmt.Sprintf("network_interface.%d", i)).(map[string]interface{})
			nicType := skytap.CreateInterfaceRequest{
				NICType: utils.NICType(skytap.NICType(networkInterface["interface_type"].(string))),
			}
			networkID := skytap.AttachInterfaceRequest{
				NetworkID: utils.String(networkInterface["network_id"].(string)),
			}

			opts := skytap.UpdateInterfaceRequest{}
			requiresUpdate := false
			if v, ok := networkInterface["ip"]; ok {
				opts.IP = utils.String(v.(string))
				requiresUpdate = true
			}
			if v, ok := networkInterface["hostname"]; ok {
				opts.Hostname = utils.String(v.(string))
				requiresUpdate = true
			}

			var id string
			{
				log.Printf("[INFO] creating interface: %#v", spew.Sdump(nicType))
				networkInterface, err := client.Create(ctx, environmentID, vmID, &nicType)
				if err != nil {
					return fmt.Errorf("error creating interface: %v", err)
				}
				id = *networkInterface.ID

				log.Printf("[INFO] created interface: %#v", spew.Sdump(networkInterface))

				if err = waitForVMReady(d, meta); err != nil {
					return err
				}
			}
			{
				log.Printf("[INFO] attaching interface: %#v", spew.Sdump(networkID))
				_, err := client.Attach(ctx, environmentID, vmID, id, &networkID)
				if err != nil {
					return fmt.Errorf("error attaching interface: %v", err)
				}

				log.Printf("[INFO] attached interface: %#v", spew.Sdump(networkInterface))

				if err = waitForVMReady(d, meta); err != nil {
					return err
				}
			}
			{
				// if the user define a hostname or ip we need an interface update.
				if requiresUpdate {
					log.Printf("[INFO] updating interface options: %#v", spew.Sdump(opts))
					networkInterface, err := client.Update(ctx, environmentID, vmID, id, &opts)
					if err != nil {
						return fmt.Errorf("error attaching interface: %v", err)
					}
					log.Printf("[INFO] updated interface: %#v", spew.Sdump(networkInterface))

					if err = waitForVMReady(d, meta); err != nil {
						return err
					}
				}
			}
			{
				// create network interfaces if necessary
				err := addPublishedServices(d, meta, environmentID, vmID, id, networkInterface)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// create the public service for a specific interface
func addPublishedServices(d *schema.ResourceData, meta interface{}, environmentID string, vmID string, nicID string, networkInterface map[string]interface{}) error {
	if _, ok := networkInterface["published_service"]; ok {
		client := meta.(*SkytapClient).publishedServicesClient
		ctx := meta.(*SkytapClient).StopContext
		publishedServices := networkInterface["published_service"].([]interface{})
		log.Printf("[INFO] creating %d published services", len(publishedServices))
		for _, v := range publishedServices {
			publishedService := v.(map[string]interface{})
			// create
			internalPort := skytap.CreatePublishedServiceRequest{
				InternalPort: utils.Int(publishedService["internal_port"].(int)),
			}
			log.Printf("[INFO] creating published service: %#v", spew.Sdump(internalPort))
			createdService, err := client.Create(ctx, environmentID, vmID, nicID, &internalPort)
			if err != nil {
				return fmt.Errorf("error creating published service: %v", err)
			}

			log.Printf("[INFO] created published service: %#v", spew.Sdump(createdService))

			if err = waitForVMReady(d, meta); err != nil {
				return err
			}
		}
	}
	return nil
}

func resourceSkytapVMRead(d *schema.ResourceData, meta interface{}) error {
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
	if err := d.Set("network_interface", flattenInterfaces(vm.Interfaces)); err != nil {
		log.Printf("[ERROR] error flattening network interfaces: %v", err)
		return err
	}

	log.Printf("[INFO] retrieved VM: %#v", spew.Sdump(vm))

	return nil
}

func resourceSkytapVMUpdate(d *schema.ResourceData, meta interface{}) error {
	return updateVMResource(d, meta, false)
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

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"false"},
		Target:     []string{"true"},
		Refresh:    vmDeleteRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for VM (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for VM (%s) to complete: %s", d.Id(), err)
	}

	log.Printf("[INFO] destroyed VM ID: %s", id)

	return err
}

func updateVMResource(d *schema.ResourceData, meta interface{}, forceRunning bool) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	environmentID := d.Get("environment_id").(string)

	opts := skytap.UpdateVMRequest{}

	if forceRunning {
		opts.Runstate = utils.VMRunstate(skytap.VMRunstateRunning)
	}
	if v, ok := d.GetOk("name"); ok {
		opts.Name = utils.String(v.(string))
	}

	log.Printf("[INFO] VM update options: %#v", spew.Sdump(opts))
	vm, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return fmt.Errorf("error updating vm (%s): %v", id, err)
	}

	log.Printf("[INFO] updated VM: %#v", spew.Sdump(vm))

	stateConf := &resource.StateChangeConf{
		Pending:    getVMPendingUpdateRunstates(forceRunning),
		Target:     getVMTargetUpdateRunstates(forceRunning),
		Refresh:    vmRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for VM (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for VM (%s) to complete: %s", d.Id(), err)
	}

	return resourceSkytapVMRead(d, meta)
}

func vmRunstateRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).vmsClient
		ctx := meta.(*SkytapClient).StopContext

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

func vmDeleteRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).vmsClient
		ctx := meta.(*SkytapClient).StopContext

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
