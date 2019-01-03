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

func resourceSkytapNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapNetworkCreate,
		Read:   resourceSkytapNetworkRead,
		Update: resourceSkytapNetworkUpdate,
		Delete: resourceSkytapNetworkDelete,

		Schema: map[string]*schema.Schema{
			"environment_id": {
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

			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"subnet": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.CIDRNetwork(16, 29),
			},

			"gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.SingleIP(),
			},

			"tunnelable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceSkytapNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).networksClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	name := d.Get("name").(string)
	domain := d.Get("domain").(string)
	subnet := d.Get("subnet").(string)
	tunnelable := d.Get("tunnelable").(bool)

	opts := skytap.CreateNetworkRequest{
		Name:        &name,
		NetworkType: utils.NetworkType(skytap.NetworkTypeAutomatic),
		Domain:      &domain,
		Subnet:      &subnet,
		Tunnelable:  &tunnelable,
	}

	if v, ok := d.GetOk("gateway"); ok {
		opts.Gateway = utils.String(v.(string))
	}

	log.Printf("[INFO] network create")
	log.Printf("[DEBUG] network create options: %#v", spew.Sdump(opts))
	network, err := client.Create(ctx, environmentID, &opts)
	if err != nil {
		return fmt.Errorf("error creating network: %v", err)
	}

	if network.ID == nil {
		return fmt.Errorf("network ID is not set")
	}
	networkID := *network.ID
	d.SetId(networkID)

	log.Printf("[INFO] network created: %s", *network.ID)
	log.Printf("[DEBUG] network created: %#v", spew.Sdump(network))

	if err = waitForEnvironmentReady(d, meta, environmentID); err != nil {
		return err
	}

	return resourceSkytapNetworkRead(d, meta)
}

func resourceSkytapNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).networksClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	log.Printf("[INFO] retrieving network: %s", id)
	network, err := client.Get(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] network (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error retrieving network (%s): %v", id, err)
	}

	d.Set("environment_id", environmentID)
	d.Set("name", network.Name)
	d.Set("domain", network.Domain)
	d.Set("subnet", network.Subnet)
	d.Set("gateway", network.Gateway)
	d.Set("tunnelable", network.Tunnelable)

	log.Printf("[INFO] network retrieved: %s", id)
	log.Printf("[DEBUG] network retrieved: %#v", spew.Sdump(network))

	return err
}

func resourceSkytapNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).networksClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	environmentID := d.Get("environment_id").(string)
	name := d.Get("name").(string)
	domain := d.Get("domain").(string)
	subnet := d.Get("subnet").(string)
	tunnelable := d.Get("tunnelable").(bool)

	opts := skytap.UpdateNetworkRequest{
		Name:       &name,
		Domain:     &domain,
		Subnet:     &subnet,
		Tunnelable: &tunnelable,
	}

	if v, ok := d.GetOk("gateway"); ok {
		opts.Gateway = utils.String(v.(string))
	}

	log.Printf("[INFO] network update: %s", id)
	log.Printf("[DEBUG] network update options: %#v", spew.Sdump(opts))
	network, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return fmt.Errorf("error updating network (%s): %v", id, err)
	}

	log.Printf("[INFO] network updated: %s", id)
	log.Printf("[DEBUG] network updated: %#v", spew.Sdump(network))

	if err = waitForEnvironmentReady(d, meta, environmentID); err != nil {
		return err
	}

	return resourceSkytapNetworkRead(d, meta)
}

func resourceSkytapNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).networksClient
	ctx := meta.(*SkytapClient).StopContext

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	log.Printf("[INFO] destroying network: %s", id)
	err := client.Delete(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] network (%s) was not found - assuming removed", id)
			return nil
		}

		return fmt.Errorf("error deleting network (%s): %v", id, err)
	}
	if err = waitForEnvironmentReady(d, meta, environmentID); err != nil {
		return err
	}

	log.Printf("[INFO] network destroyed: %s", id)

	return err
}
