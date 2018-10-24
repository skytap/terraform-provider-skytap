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
	"net/http"
	"time"
)

func resourceSkytapEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapEnvironmentCreate,
		Read:   resourceSkytapEnvironmentRead,
		Update: resourceSkytapEnvironmentUpdate,
		Delete: resourceSkytapEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				ForceNew:     true,
			},

			"project_id": {
				Type:         schema.TypeInt,
				Optional:     true,
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

			//"owner": {
			//	Type:         schema.TypeString,
			//	Optional:     true,
			//	ValidateFunc: validation.NoZeroValues,
			//},

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
				Type:     schema.TypeInt,
				Optional: true,
				Default:  nil,
			},

			"suspend_at_time": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},

			"shutdown_on_idle": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  nil,
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

	log.Printf("[INFO] preparing arguments for creating the Skytap Environment")

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

	if v, ok := d.GetOk("project_id"); ok {
		opts.ProjectID = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = utils.String(v.(string))
	}

	//if v, ok := d.GetOk("owner"); ok {
	//	opts.Owner = utils.String(v.(string))
	//}

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

	log.Printf("[DEBUG] environment create options: %#v", opts)
	environment, err := client.Create(ctx, &opts)
	if err != nil {
		return errors.Errorf("error creating environment: %v", err)
	}

	d.SetId(*environment.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    environmentPendingCreateRunstates,
		Target:     environmentTargetCreateRunstates,
		Refresh:    environmentCreateRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
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

	d.Set("name", environment.Name)
	d.Set("description", environment.Description)
	//d.Set("owner", environment.OwnerID)
	d.Set("outbound_traffic", environment.OutboundTraffic)
	d.Set("routable", environment.OutboundTraffic)
	d.Set("suspend_on_idle", environment.SuspendOnIdle)
	d.Set("suspend_at_time", environment.SuspendAtTime)
	d.Set("shutdown_on_idle", environment.ShutdownOnIdle)
	d.Set("shutdown_at_time", environment.ShutdownAtTime)

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

	//if v, ok := d.GetOk("owner"); ok {
	//	opts.Owner = utils.String(v.(string))
	//}

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

	log.Printf("[DEBUG] environment update options: %#v", opts)
	_, err := client.Update(ctx, id, &opts)
	if err != nil {
		return errors.Errorf("error updating environment (%s): %v", id, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    environmentPendingUpdateRunstates,
		Target:     environmentTargetUpdateRunstates,
		Refresh:    environmentUpdateRunstateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[INFO] Waiting for environment (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for environment (%s) to complete: %s", d.Id(), err)
	}

	return resourceSkytapEnvironmentRead(d, meta)
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

		log.Printf("[INFO] retrieving environment: %s", id)
		environment, err := client.Get(ctx, id)

		if err != nil {
			log.Printf("[WARN] Error on retrieving environment status (%s) when waiting: %s", id, err)
			return nil, "", err
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

func environmentUpdateRunstateRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).environmentsClient
		ctx := meta.(*SkytapClient).StopContext

		id := d.Id()

		log.Printf("[INFO] retrieving environment: %s", id)
		environment, err := client.Get(ctx, id)

		if err != nil {
			if skytapErr, ok := err.(*skytap.ErrorResponse); ok {
				if http.StatusNotFound == skytapErr.Response.StatusCode {
					return 42, string(skytap.EnvironmentRunstateStopped), nil
				}
			}
			log.Printf("[WARN] Error on retrieving environment status (%s) when waiting: %s", id, err)
			return nil, "", err
		}

		log.Printf("[DEBUG] environment status (%s): %s", id, *environment.Runstate)

		return environment, string(*environment.Runstate), nil
	}
}
