package skytap

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/pkg/errors"
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
				ValidateFunc: validation.NoZeroValues,
				ForceNew:     true,
			},

			"template_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				ForceNew:     true,
			},

			"vm_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				ForceNew:     true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
				Computed:     true,
			},

			"network_interface": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interface_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(skytap.NICTypeDefault),
								string(skytap.NICTypeE1000),
								string(skytap.NICTypeE1000E),
								string(skytap.NICTypePCNet32),
								string(skytap.NICTypeVMXNet),
								string(skytap.NICTypeVMXNet3),
							}, false),
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
							ValidateFunc: validation.SingleIP(),
							Computed:     true,
							ForceNew:     true,
						},
						"hostname": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.NoZeroValues,
							Computed:     true,
							ForceNew:     true,
						},

						"published_service": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"internal_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.NoZeroValues,
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

	log.Printf("[INFO] preparing arguments for creating the Skytap VM")

	environmentID := d.Get("environment_id").(string)
	templateID := d.Get("template_id").(string)
	templateVMID := d.Get("vm_id").(string)

	// create the VM
	createOpts := skytap.CreateVMRequest{
		TemplateID: templateID,
		VMID:       templateVMID,
	}

	log.Printf("[DEBUG] vm create options: %#v", createOpts)
	vm, err := client.Create(ctx, environmentID, &createOpts)
	if err != nil {
		return errors.Errorf("error creating vm: %v", err)
	}

	d.SetId(*vm.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    vmPendingCreateRunstates,
		Target:     vmTargetCreateRunstates,
		Refresh:    vmRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for vm (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for vm (%s) to complete: %s", d.Id(), err)
	}

	// create network interfaces if necessary
	err = addNetworkAdapters(d, meta, *vm.ID)

	return update(d, meta, true)
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
				return errors.Errorf("error resolving vm network interfaces: %v", err)
			}
			for _, iface := range vmIfaces.Value {
				err = client.Delete(ctx, environmentID, vmID, *iface.ID)
				if err != nil {
					return errors.Errorf("Error removing the default interface from vm: %v", err)
				}
			}
		}
		for i := 0; i < networkIfaceCount; i++ {
			log.Printf("[INFO] preparing arguments for creating the Skytap vm network interface")
			networkInterface := d.Get(fmt.Sprintf("network_interface.%d", i)).(map[string]interface{})
			var id string
			{
				// create
				nicType := skytap.CreateInterfaceRequest{
					NICType: utils.NICType(skytap.NICType(networkInterface["interface_type"].(string))),
				}
				log.Printf("[DEBUG] vm network interface create options: %#v", nicType)
				networkInterface, err := client.Create(ctx, environmentID, vmID, &nicType)
				if err != nil {
					return errors.Errorf("error creating vm network interface: %v", err)
				}
				id = *networkInterface.ID
			}
			{
				// attach
				networkID := skytap.AttachInterfaceRequest{
					NetworkID: utils.String(networkInterface["network_id"].(string)),
				}
				log.Printf("[DEBUG] vm network interface attachment : %#v", networkID)
				_, err := client.Attach(ctx, environmentID, vmID, id, &networkID)
				if err != nil {
					return errors.Errorf("error attaching vm network interface: %v", err)
				}
			}
			{
				// update
				opts := skytap.UpdateInterfaceRequest{}
				if v, ok := networkInterface["ip"]; ok {
					opts.IP = utils.String(v.(string))
				}
				if v, ok := networkInterface["hostname"]; ok {
					opts.Hostname = utils.String(v.(string))
				}
				log.Printf("[DEBUG] vm network interface attachment : %#v", opts)
				_, err := client.Update(ctx, environmentID, vmID, id, &opts)
				if err != nil {
					return errors.Errorf("error attaching vm network interface: %v", err)
				}
			}
			{
				// create network interfaces if necessary
				err := addPublishedServices(d, meta, environmentID, vmID, id, networkInterface)
				if err != nil {
					return errors.Errorf("error attaching vm network interface: %v", err)
				}
			}
		}
	}

	return nil
}

func addPublishedServices(d *schema.ResourceData, meta interface{}, environmentID string, vmID string, nicID string, networkInterfaces map[string]interface{}) error {
	if _, ok := networkInterfaces["published_service"]; ok {
		client := meta.(*SkytapClient).publishedServicesClient
		ctx := meta.(*SkytapClient).StopContext

		publishedServices := networkInterfaces["published_service"].([]interface{})
		for _, v := range publishedServices {

			log.Printf("[INFO] preparing arguments for creating the Skytap vm network interface service")

			publishedService := v.(map[string]interface{})
			// create
			internalPort := skytap.CreatePublishedServiceRequest{
				InternalPort: utils.Int(publishedService["internal_port"].(int)),
			}
			log.Printf("[DEBUG] vm network interface service : %#v", internalPort)
			_, err := client.Create(ctx, environmentID, vmID, nicID, &internalPort)
			if err != nil {
				return errors.Errorf("error creating vm network interface service: %v", err)
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

	log.Printf("[INFO] retrieving vm: %s", id)
	vm, err := client.Get(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] vm (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error retrieving vm (%s): %v", id, err)
	}

	d.Set("environment_id", environmentID)
	d.Set("name", vm.Name)
	d.Set("network_interface", flattenInterfaces(vm.Interfaces))

	return err
}

func flattenInterfaces(interfaces []skytap.Interface) interface{} {
	results := make([]map[string]interface{}, 0)

	for _, v := range interfaces {
		result := make(map[string]interface{})
		result["interface_type"] = *v.NICType
		result["network_id"] = *v.NetworkID
		result["ip"] = *v.IP
		result["hostname"] = *v.Hostname
		result["published_service"] = flattenPublishedServices(v.Services)

		results = append(results, result)
	}

	return results
}

func flattenPublishedServices(publishedServices []skytap.PublishedService) interface{} {
	results := make([]map[string]interface{}, 0)

	for _, v := range publishedServices {
		result := make(map[string]interface{})
		result["id"] = *v.ID
		result["internal_port"] = *v.InternalPort
		result["external_ip"] = *v.ExternalPort
		result["external_port"] = *v.ExternalIP

		results = append(results, result)
	}

	return results
}

func resourceSkytapVMUpdate(d *schema.ResourceData, meta interface{}) error {
	return update(d, meta, false)
}

func resourceSkytapVMDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	log.Printf("[INFO] destroying vm: %s", id)
	err := client.Delete(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] vm (%s) was not found - assuming removed", id)
			return nil
		}

		return fmt.Errorf("error deleting vm (%s): %v", id, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"false"},
		Target:     []string{"true"},
		Refresh:    vmDeleteRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for vm (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for vm (%s) to complete: %s", d.Id(), err)
	}

	return err
}

func update(d *schema.ResourceData, meta interface{}, forceRunning bool) error {
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

	log.Printf("[DEBUG] vm update options: %#v", opts)
	_, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return errors.Errorf("error updating vm (%s): %v", id, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    vmPendingUpdateRunstates,
		Target:     getVMTargetUpdateRunstates(forceRunning),
		Refresh:    vmRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for vm (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for vm (%s) to complete: %s", d.Id(), err)
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

		log.Printf("[INFO] retrieving VM: %s", id)
		vm, err := client.Get(ctx, environmentID, id)

		if err != nil {
			log.Printf("[WARN] Error on retrieving VM status (%s) when waiting: %s", id, err)
			return nil, "", err
		}

		log.Printf("[DEBUG] environment status (%s): %s", id, *vm.Runstate)

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

		log.Printf("[INFO] retrieving VM: %s", id)
		vm, err := client.Get(ctx, environmentID, id)

		var removed = "false"
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				log.Printf("[DEBUG] vm (%s) has been removed.", id)
				removed = "true"
			} else {
				return nil, "", fmt.Errorf("error retrieving vm (%s): %v", id, err)
			}
		}

		return vm, removed, nil
	}
}
