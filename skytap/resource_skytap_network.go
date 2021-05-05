package skytap

import (
	"context"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"

	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSkytapNetworkCreate,
		ReadContext:   resourceSkytapNetworkRead,
		UpdateContext: resourceSkytapNetworkUpdate,
		DeleteContext: resourceSkytapNetworkDelete,

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
				ValidateFunc: validation.IsCIDRNetwork(16, 29),
			},

			"gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPAddress,
			},

			"tunnelable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceSkytapNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).networksClient

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
	log.Printf("[TRACE] network create options: %v", spew.Sdump(opts))
	network, err := client.Create(ctx, environmentID, &opts)
	if err != nil {
		return diag.Errorf("error creating network: %v", err)
	}

	if network.ID == nil {
		return diag.Errorf("network ID is not set")
	}
	networkID := *network.ID
	d.SetId(networkID)

	log.Printf("[INFO] network created: %s", *network.ID)
	log.Printf("[TRACE] network created: %v", spew.Sdump(network))

	if err = waitForEnvironmentReady(ctx, d, meta, environmentID, schema.TimeoutCreate); err != nil {
		return diag.FromErr(err)
	}

	return resourceSkytapNetworkRead(ctx, d, meta)
}

func resourceSkytapNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).networksClient

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

		return diag.Errorf("error retrieving network (%s): %v", id, err)
	}

	err = d.Set("environment_id", environmentID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("name", network.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("domain", network.Domain)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("subnet", network.Subnet)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("gateway", network.Gateway)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("tunnelable", network.Tunnelable)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] network retrieved: %s", id)
	log.Printf("[TRACE] network retrieved: %v", spew.Sdump(network))

	return nil
}

func resourceSkytapNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).networksClient

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
	log.Printf("[TRACE] network update options: %v", spew.Sdump(opts))
	network, err := client.Update(ctx, environmentID, id, &opts)
	if err != nil {
		return diag.Errorf("error updating network (%s): %v", id, err)
	}

	log.Printf("[INFO] network updated: %s", id)
	log.Printf("[TRACE] network updated: %v", spew.Sdump(network))

	if err = waitForEnvironmentReady(ctx, d, meta, environmentID, schema.TimeoutUpdate); err != nil {
		return diag.FromErr(err)
	}

	return resourceSkytapNetworkRead(ctx, d, meta)
}

func resourceSkytapNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).networksClient

	environmentID := d.Get("environment_id").(string)
	id := d.Id()

	log.Printf("[INFO] destroying network: %s", id)
	err := client.Delete(ctx, environmentID, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] network (%s) was not found - assuming removed", id)
			return nil
		}

		return diag.Errorf("error deleting network (%s): %v", id, err)
	}
	if err = waitForEnvironmentReady(ctx, d, meta, environmentID, schema.TimeoutDelete); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] network destroyed: %s", id)

	return nil
}
