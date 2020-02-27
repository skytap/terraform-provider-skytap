package skytap

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapProjectCreate,
		Read:   resourceSkytapProjectRead,
		Update: resourceSkytapProjectUpdate,
		Delete: resourceSkytapProjectDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"summary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"auto_add_role_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRoleType(),
			},

			"show_project_members": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceSkytapProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx, cancel := stopContextForCreate(d, meta.(*SkytapClient))
	defer cancel()

	name := d.Get("name").(string)
	showProjectMembers := d.Get("show_project_members").(bool)

	opts := skytap.Project{
		Name:               &name,
		ShowProjectMembers: &showProjectMembers,
	}

	if v, ok := d.GetOk("summary"); ok {
		opts.Summary = utils.String(v.(string))
	}

	if v, ok := d.GetOk("auto_add_role_name"); ok {
		autoAddRoleName := skytap.ProjectRole(v.(string))
		opts.AutoAddRoleName = &autoAddRoleName
	}

	log.Printf("[INFO] project create")
	log.Printf("[TRACE] project create options: %v", spew.Sdump(opts))
	project, err := client.Create(ctx, &opts)
	if err != nil {
		return fmt.Errorf("error creating project: %v", err)
	}

	if project.ID == nil {
		return fmt.Errorf("project ID is not set")
	}
	projectID := strconv.Itoa(*project.ID)
	d.SetId(projectID)

	log.Printf("[INFO] project created: %d", *project.ID)
	log.Printf("[TRACE] project created: %v", spew.Sdump(project))

	return resourceSkytapProjectRead(d, meta)
}

func resourceSkytapProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx, cancel := stopContextForRead(d, meta.(*SkytapClient))
	defer cancel()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("project (%s) is not an integer: %v", d.Id(), err)
	}

	log.Printf("[INFO] retrieving project: %d", id)
	project, err := client.Get(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] project (%d) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error retrieving project (%d): %v", id, err)
	}

	d.Set("name", project.Name)
	d.Set("summary", project.Summary)
	d.Set("auto_add_role_name", project.AutoAddRoleName)
	d.Set("show_project_members", project.ShowProjectMembers)

	log.Printf("[INFO] project retrieved: %d", id)
	log.Printf("[TRACE] project retrieved: %v", spew.Sdump(project))

	return err
}

func resourceSkytapProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx, cancel := stopContextForUpdate(d, meta.(*SkytapClient))
	defer cancel()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("project (%s) is not an integer: %v", d.Id(), err)
	}

	name := d.Get("name").(string)
	showProjectMembers := d.Get("show_project_members").(bool)

	opts := skytap.Project{
		Name:               &name,
		ShowProjectMembers: &showProjectMembers,
	}

	if v, ok := d.GetOk("summary"); ok {
		opts.Summary = utils.String(v.(string))
	}

	if v, ok := d.GetOk("auto_add_role_name"); ok {
		autoAddRoleName := skytap.ProjectRole(v.(string))
		opts.AutoAddRoleName = &autoAddRoleName
	}

	log.Printf("[INFO] project update: %d", id)
	log.Printf("[TRACE] project update options: %v", spew.Sdump(opts))
	project, err := client.Update(ctx, id, &opts)
	if err != nil {
		return fmt.Errorf("error updating project (%d): %v", id, err)
	}

	log.Printf("[INFO] project updated: %d", id)
	log.Printf("[TRACE] project updated: %v", spew.Sdump(project))

	return resourceSkytapProjectRead(d, meta)
}

func resourceSkytapProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx, cancel := stopContextForDelete(d, meta.(*SkytapClient))
	defer cancel()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("project (%s) is not an integer: %v", d.Id(), err)
	}

	log.Printf("[INFO] destroying project: %d", id)
	err = client.Delete(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] project (%d) was not found - assuming removed", id)
			return nil
		}

		return fmt.Errorf("error deleting project (%d): %v", id, err)
	}

	log.Printf("[INFO] project destroyed: %d", id)

	return err
}
