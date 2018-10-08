package skytap

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/opencredo/skytap-sdk-go-internal"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"summary": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*skytap.Client)
	name := d.Get("name").(string)
	summary := d.Get("summary").(string)
	project, err := client.CreateProject(context.Background(), name, summary)
	if err != nil {
		return err
	}
	d.SetId(project.Id)
	return err
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*skytap.Client)
	project, err := client.ReadProject(context.Background(), d.Id())

	if err != nil {
		return err
	}

	d.Set("name", project.Name)
	d.Set("summary", project.Summary)
	return err
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*skytap.Client)
	name := d.Get("name").(string)
	summary := d.Get("summary").(string)
	_, err := client.UpdateProject(context.Background(), &skytap.Project{
		d.Id(),
		name,
		summary,
	})
	return err
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*skytap.Client)
	err := client.DeleteProject(context.Background(), d.Id())
	return err
}
