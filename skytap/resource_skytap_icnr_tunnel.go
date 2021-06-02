package skytap

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapICNRTunnel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSkytapICNRTunnelCreate,
		ReadContext:   resourceSkytapICNRTunnelRead,
		DeleteContext: resourceSkytapICNRTunnelDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"source": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Source network from where the connection is initiated. This network does not need to be 'tunnelable' (visible to other networks)",
				ForceNew:    true,
			},
			"target": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Target network to which the connection is made. The network does need to be 'tunnelable' (visible to other networks)",
				ForceNew:    true,
			},
		},
	}
}

func resourceSkytapICNRTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).icnrTunnelClient

	source := d.Get("source").(int)
	target := d.Get("target").(int)

	log.Printf("[INFO] ICNR tunnel created create")
	tunnel, err := client.Create(ctx, source, target)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*tunnel.ID)
	return resourceSkytapICNRTunnelRead(ctx, d, meta)
}

func resourceSkytapICNRTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).icnrTunnelClient

	id := d.Id()
	_, err := client.Get(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] ICNR tunnel (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}
	return nil
}

func resourceSkytapICNRTunnelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).icnrTunnelClient

	log.Printf("[INFO] destroying ICNR tunnel: %s", d.Id())
	err := client.Delete(ctx, d.Id())
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] ICNR tunnel (%s) was not found - assuming removed", d.Id())
			return nil
		}

		return diag.Errorf("error deleting ICNR tunnel (%s): %v", d.Id(), err)
	}
	log.Printf("[INFO] environment destroyed: %s", d.Id())

	return nil
}
