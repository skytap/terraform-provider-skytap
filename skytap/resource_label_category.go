package skytap

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
	"log"
	"strconv"
)

func resourceSkytapLabelCategory() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapLabelCategoryCreate,
		Read:   resourceSkytapLabelCategoryRead,
		Delete: resourceSkytapLabelCategoryDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.NoZeroValues,
					validation.StringLenBetween(1, 128),
					validateNoSubString(";"),
					validateNoSubString(","),
					validateNoStartWith("Skytap"),
				),
				ForceNew: true,
			},

			"single_value": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSkytapLabelCategoryCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).labelCategoryClient
	ctx := meta.(*SkytapClient).StopContext

	name := d.Get("name").(string)
	singleValue := d.Get("single_value").(bool)

	newLabelCategory := skytap.LabelCategory{
		Name:        &name,
		SingleValue: &singleValue,
	}

	createdLabelCategory, err := client.Create(ctx, &newLabelCategory)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(*createdLabelCategory.ID))
	log.Printf("[INFO] Label category created: %d", *createdLabelCategory.ID)
	log.Printf("[TRACE] Label category created: %v", spew.Sdump(createdLabelCategory))

	return resourceSkytapLabelCategoryRead(d, meta)
}

func resourceSkytapLabelCategoryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).labelCategoryClient
	ctx := meta.(*SkytapClient).StopContext

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("label category (%s) is not an integer: %v", d.Id(), err)
	}

	log.Printf("[INFO] retrieving project category: %d", id)
	labelCategory, err := client.Get(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] label category (%d) was not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving label category (%d): %v", id, err)
	} else {
		if ! *labelCategory.Enabled {
			log.Printf("[DEBUG] label category (%d) was not found - removing from state", id)
			d.SetId("")
			return nil
		}
	}

	d.Set("name", labelCategory.Name)
	d.Set("single_value", labelCategory.SingleValue)

	log.Printf("[INFO] label category retrieved: %d", id)
	log.Printf("[TRACE] label category retrieved: %v", spew.Sdump(labelCategory))

	return err
}

func resourceSkytapLabelCategoryDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).labelCategoryClient
	ctx := meta.(*SkytapClient).StopContext

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("label category (%s) is not an integer: %v", d.Id(), err)
	}

	log.Printf("[INFO] destroying label category: %d", id)
	err = client.Delete(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] label category (%d) was not found - assuming removed", id)
			return nil
		}
		return fmt.Errorf("error deleting label category (%d): %v", id, err)
	}

	log.Printf("[INFO] label category destroyed: %d", id)
	return err
}
