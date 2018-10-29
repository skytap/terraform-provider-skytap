package skytap

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/pkg/errors"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
	"log"
	"sort"
	"time"
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
	{
		// create the VM
		createOpts := skytap.CreateVMRequest{
			TemplateID: templateID,
			VMID:       []string{templateVMID},
		}

		log.Printf("[DEBUG] vm create options: %#v", createOpts)
		environment, err := client.Create(ctx, environmentID, &createOpts)
		if err != nil || len(environment.VMs) == 0 {
			return errors.Errorf("error creating vm: %v", err)
		}

		vm := mostRecentVM(environment.VMs)

		d.SetId(*vm.ID)
	}
	{
		// Update VM to running
		updateOpts := skytap.UpdateVMRequest{
			Runstate: utils.VMRunstate(skytap.VMRunstateRunning),
		}
		log.Printf("[DEBUG] vm update runstate to running")
		_, err := client.Update(ctx, environmentID, d.Id(), &updateOpts)
		if err != nil {
			return errors.Errorf("error updating vm: %v", err)
		}
	}
	stateConf := &resource.StateChangeConf{
		Pending:    VMPendingCreateRunstates,
		Target:     VMTargetCreateRunstates,
		Refresh:    VMRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for vm (%s) to complete", d.Id())
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for vm (%s) to complete: %s", d.Id(), err)
	}

	return resourceSkytapVMRead(d, meta)
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
	d.Set("runstate", vm.Runstate)

	return err
}

func resourceSkytapVMUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).vmsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	environmentID := d.Get("environment_id").(string)

	opts := skytap.UpdateVMRequest{}

	if v, ok := d.GetOk("name"); ok {
		opts.Name = utils.String(v.(string))
	}

	log.Printf("[DEBUG] vm update options: %#v", opts)
	_, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return errors.Errorf("error updating vm (%s): %v", id, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    VMPendingUpdateRunstates,
		Target:     VMTargetUpdateRunstates,
		Refresh:    VMRunstateRefreshFunc(d, meta),
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

	return err
}

var VMPendingCreateRunstates = []string{
	string(skytap.VMRunstateBusy),
}

var VMPendingUpdateRunstates = []string{
	string(skytap.VMRunstateBusy),
}

var VMTargetCreateRunstates = []string{
	string(skytap.VMRunstateRunning),
}

var VMTargetUpdateRunstates = []string{
	string(skytap.VMRunstateRunning),
	string(skytap.VMRunstateStopped),
	string(skytap.VMRunstateReset),
	string(skytap.VMRunstateSuspended),
	string(skytap.VMRunstateHalted),
}

func mostRecentVM(vms []skytap.VM) skytap.VM {
	sort.Slice(vms, func(i, j int) bool {
		time1, _ := time.Parse(timestampFormat, *vms[i].CreatedAt)
		time2, _ := time.Parse(timestampFormat, *vms[j].CreatedAt)
		return time1.After(time2)
	})
	return vms[0]
}

func VMRunstateRefreshFunc(
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
