package skytap

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
)

func dataSourceSkytapProject() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSkytapProjectRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "name of the project",
				ValidateFunc: validation.NoZeroValues,
			},

			// computed attributes
			"summary": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "summary description of the project",
			},

			"auto_add_role_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "role to automatically assign to every new user added to the project",
			},

			"show_project_members": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "whether project members can view a list of other project members",
			},
		},
	}
}

func dataSourceSkytapProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Skytap).projectsClient
	ctx := meta.(*Skytap).StopContext

	log.Printf("[INFO] preparing arguments for finding the Skytap Project")

	name := d.Get("name").(string)

	projectsResult, err := client.List(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving projects: %s", err)
	}

	projects := filterDataSourceSkytapSnapshotsByName(projectsResult.Value, name)

	if len(projects) == 0 {
		return fmt.Errorf("no project found with name %s", name)
	}

	if len(projects) > 1 {
		return fmt.Errorf("too many projects found with name %s (found %d, expected 1)", name, len(projects))
	}

	project := projects[0]

	d.SetId(*project.ID)
	d.Set("name", project.Name)
	d.Set("summary", project.Summary)
	d.Set("auto_add_role_name", project.AutoAddRoleName)
	d.Set("show_project_members", project.ShowProjectMembers)

	return nil
}

func filterDataSourceSkytapSnapshotsByName(projects []skytap.Project, name string) []skytap.Project {
	var result []skytap.Project
	for _, p := range projects {
		if *p.Name == name {
			result = append(result, p)
		}
	}
	return result
}
