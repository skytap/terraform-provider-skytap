package skytap

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
	"log"
	"time"
)

func resourceSkytapICNRTunnel() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapICNRTunnelCreate,
		Read:   resourceSkytapICNRTunnelRead,
		Delete: resourceSkytapICNRTunnelDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"source": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSkytapICNRTunnelCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).icnrTunnelClient
	ctx, cancel := stopContextForCreate(d, meta.(*SkytapClient))
	defer cancel()

	source := d.Get("source").(int)
	target := d.Get("target").(int)

	log.Printf("[INFO] ICNR tunnel created create")
	tunnel, err := client.Create(ctx, source, target)
	if err != nil {
		return err
	}

	d.SetId(*tunnel.ID)
	return resourceSkytapICNRTunnelRead(d, meta)
}

func resourceSkytapICNRTunnelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).icnrTunnelClient
	ctx, cancel := stopContextForRead(d, meta.(*SkytapClient))
	defer cancel()

	id := d.Id()
	_, err := client.Get(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] ICNR tunnel (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return err
	}
	return nil
}

func resourceSkytapICNRTunnelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).icnrTunnelClient
	ctx, cancel := stopContextForDelete(d, meta.(*SkytapClient))
	defer cancel()

	log.Printf("[INFO] destroying ICNR tunnel: %s", d.Id())
	err := client.Delete(ctx, d.Id())
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] ICNR tunnel (%s) was not found - assuming removed", d.Id())
			return nil
		}

		return fmt.Errorf("error deleting ICNR tunnel (%s): %v", d.Id(), err)
	}
	log.Printf("[INFO] environment destroyed: %s", d.Id())

	return nil
}
