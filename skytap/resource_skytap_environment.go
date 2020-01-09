package skytap

import (
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapEnvironmentCreate,
		Read:   resourceSkytapEnvironmentRead,
		Update: resourceSkytapEnvironmentUpdate,
		Delete: resourceSkytapEnvironmentDelete,

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"outbound_traffic": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"routable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"suspend_on_idle": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(300, 86400),
			},

			"suspend_at_time": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},

			"shutdown_on_idle": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(300, 86400),
			},

			"shutdown_at_time": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceSkytapEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx := meta.(*SkytapClient).StopContext

	templateID := d.Get("template_id").(string)
	name := d.Get("name").(string)
	outboundTraffic := d.Get("outbound_traffic").(bool)
	routable := d.Get("routable").(bool)

	opts := skytap.CreateEnvironmentRequest{
		TemplateID:      &templateID,
		Name:            &name,
		OutboundTraffic: &outboundTraffic,
		Routable:        &routable,
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = utils.String(v.(string))
	}

	if v, ok := d.GetOk("suspend_on_idle"); ok {
		opts.SuspendOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("suspend_at_time"); ok {
		opts.SuspendAtTime = utils.String(v.(string))
	}

	if v, ok := d.GetOk("shutdown_on_idle"); ok {
		opts.ShutdownOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("shutdown_at_time"); ok {
		opts.ShutdownAtTime = utils.String(v.(string))
	}

	log.Printf("[INFO] environment create")
	log.Printf("[TRACE] environment create options: %v", spew.Sdump(opts))
	environment, err := client.Create(ctx, &opts)
	if err != nil {
		return fmt.Errorf("error creating environment: %v", err)
	}

	if environment.ID == nil {
		return fmt.Errorf("environment ID is not set")
	}
	environmentID := *environment.ID
	d.SetId(environmentID)

	log.Printf("[INFO] environment created: %s", *environment.ID)
	log.Printf("[TRACE] environment created: %v", spew.Sdump(environment))

	stateConf := &resource.StateChangeConf{
		Pending:    environmentPendingCreateRunstates,
		Target:     environmentTargetCreateRunstates,
		Refresh:    environmentCreateRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for environment (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for environment (%s) to complete: %s", d.Id(), err)
	}

	return resourceSkytapEnvironmentRead(d, meta)
}

func resourceSkytapEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	log.Printf("[INFO] retrieving environment: %s", id)
	environment, err := client.Get(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] environment (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error retrieving environment (%s): %v", id, err)
	}

	var routable bool
	if environment.OutboundTraffic != nil {
		routable = *environment.OutboundTraffic
	}

	// The templateID is not set as it is used to build the environment and is not returned by the environment response.
	// If this attribute is changed, this environment will be rebuilt
	d.Set("name", environment.Name)
	d.Set("description", environment.Description)
	d.Set("outbound_traffic", environment.OutboundTraffic)
	d.Set("routable", routable)
	d.Set("suspend_on_idle", environment.SuspendOnIdle)
	d.Set("suspend_at_time", environment.SuspendAtTime)
	d.Set("shutdown_on_idle", environment.ShutdownOnIdle)
	d.Set("shutdown_at_time", environment.ShutdownAtTime)

	log.Printf("[INFO] environment retrieved: %s", id)
	log.Printf("[TRACE] environment retrieved: %v", spew.Sdump(environment))

	return err
}

func resourceSkytapEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	name := d.Get("name").(string)
	outboundTraffic := d.Get("outbound_traffic").(bool)
	routable := d.Get("routable").(bool)

	opts := skytap.UpdateEnvironmentRequest{
		Name:            &name,
		OutboundTraffic: &outboundTraffic,
		Routable:        &routable,
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = utils.String(v.(string))
	}

	if v, ok := d.GetOk("suspend_on_idle"); ok {
		opts.SuspendOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("suspend_at_time"); ok {
		opts.SuspendAtTime = utils.String(v.(string))
	}

	if v, ok := d.GetOk("shutdown_on_idle"); ok {
		opts.ShutdownOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("shutdown_at_time"); ok {
		opts.ShutdownAtTime = utils.String(v.(string))
	}

	log.Printf("[INFO] environment update: %s", id)
	log.Printf("[TRACE] environment update options: %v", spew.Sdump(opts))
	environment, err := client.Update(ctx, id, &opts)
	if err != nil {
		return fmt.Errorf("error updating environment (%s): %v", id, err)
	}

	log.Printf("[INFO] environment updated: %s", id)
	log.Printf("[TRACE] environment updated: %v", spew.Sdump(environment))

	if err = waitForEnvironmentReady(d, meta, *environment.ID); err != nil {
		return err
	}

	return resourceSkytapEnvironmentRead(d, meta)
}

func waitForEnvironmentReady(d *schema.ResourceData, meta interface{}, environmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    environmentPendingUpdateRunstates,
		Target:     environmentTargetUpdateRunstates,
		Refresh:    environmentUpdateRunstateRefreshFunc(meta, environmentID),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for environment (%s) to complete", environmentID)
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for environment (%s) to complete: %s", environmentID, err)
	}
	return nil
}

func resourceSkytapEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	log.Printf("[INFO] destroying environment: %s", id)
	err := client.Delete(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] environment (%s) was not found - assuming removed", id)
			return nil
		}

		return fmt.Errorf("error deleting environment (%s): %v", id, err)
	}

	log.Printf("[INFO] environment destroyed: %s", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"false"},
		Target:     []string{"true"},
		Refresh:    environmentDeleteRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for environment (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for environment (%s) to complete: %s", d.Id(), err)
	}

	return err
}

var environmentPendingCreateRunstates = []string{
	string(skytap.EnvironmentRunstateBusy),
}

var environmentTargetCreateRunstates = []string{
	string(skytap.EnvironmentRunstateRunning),
}

var environmentPendingUpdateRunstates = []string{
	string(skytap.EnvironmentRunstateBusy),
}

var environmentTargetUpdateRunstates = []string{
	string(skytap.EnvironmentRunstateRunning),
	string(skytap.EnvironmentRunstateStopped),
	string(skytap.EnvironmentRunstateSuspended),
}

func environmentCreateRunstateRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).environmentsClient
		ctx := meta.(*SkytapClient).StopContext

		id := d.Id()

		log.Printf("[DEBUG] retrieving environment: %s", id)
		environment, err := client.Get(ctx, id)

		if err != nil {
			return nil, "", fmt.Errorf("error retrieving environment (%s) when waiting: %v", id, err)
		}

		computedRunstate := skytap.EnvironmentRunstateRunning
		for i := 0; i < *environment.VMCount; i++ {
			if *environment.VMs[i].Runstate != skytap.VMRunstateRunning {
				computedRunstate = skytap.EnvironmentRunstateBusy
				break
			}
		}

		log.Printf("[DEBUG] environment status (%s): %s", id, *environment.Runstate)

		return environment, string(computedRunstate), nil
	}
}

func environmentUpdateRunstateRefreshFunc(meta interface{}, environmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).environmentsClient
		ctx := meta.(*SkytapClient).StopContext

		log.Printf("[DEBUG] retrieving environment: %s", environmentID)
		environment, err := client.Get(ctx, environmentID)

		if err != nil {
			return nil, "", fmt.Errorf("error retrieving environment (%s) when waiting: %v", environmentID, err)
		}

		log.Printf("[DEBUG] environment (%s): %s", environmentID, *environment.Runstate)

		return environment, string(*environment.Runstate), nil
	}
}

func environmentDeleteRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).environmentsClient
		ctx := meta.(*SkytapClient).StopContext

		id := d.Id()

		log.Printf("[DEBUG] retrieving environment: %s", id)
		environment, err := client.Get(ctx, id)

		var removed = "false"
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				log.Printf("[DEBUG] environment (%s) has been removed.", id)
				removed = "true"
			} else {
				return nil, "", fmt.Errorf("error retrieving environment (%s) when waiting: %v", id, err)
			}
		}

		return environment, removed, nil
	}
}
