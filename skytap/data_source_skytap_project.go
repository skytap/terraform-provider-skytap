package skytap

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
)

func dataSourceSkytapProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSkytapProjectRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the project",
				ValidateFunc: validation.NoZeroValues,
			},

			// computed attributes
			"summary": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The summary description of the project",
			},

			"auto_add_role_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role automatically assigned to every new user added to the project",
			},

			"show_project_members": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether project members can view a list of the other project members",
			},
		},
	}
}

func dataSourceSkytapProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*SkytapClient).projectsClient

	log.Printf("[INFO] preparing arguments for finding the Skytap Project")

	name := d.Get("name").(string)

	projectsResult, err := client.List(ctx)
	if err != nil {
		return diag.Errorf("error retrieving projects: %s", err)
	}

	projects := filterDataSourceSkytapProjectsByName(projectsResult.Value, name)

	if len(projects) == 0 {
		return diag.Errorf("no project found with name %s", name)
	}

	if len(projects) > 1 {
		return diag.Errorf("too many projects found with name %s (found %d, expected 1)", name, len(projects))
	}

	project := projects[0]
	if project.ID == nil {
		return diag.Errorf("project ID is not set")
	}
	projectID := strconv.Itoa(*project.ID)
	d.SetId(projectID)

	err = d.Set("name", project.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("summary", project.Summary)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("auto_add_role_name", project.AutoAddRoleName)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("show_project_members", project.ShowProjectMembers)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func filterDataSourceSkytapProjectsByName(projects []skytap.Project, name string) []skytap.Project {
	var result []skytap.Project
	for _, p := range projects {
		if *p.Name == name {
			result = append(result, p)
		}
	}
	return result
}
