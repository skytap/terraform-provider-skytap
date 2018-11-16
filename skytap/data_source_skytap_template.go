package skytap

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
)

const timestampFormat = "2006/01/02 15:04:05 -0700"

func dataSourceSkytapTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSkytapTemplateRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "regex expression to name of the template",
				ValidateFunc: validation.NoZeroValues,
			},
			"most_recent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "used when multiple items will be returned",
			},
		},
	}
}

func dataSourceSkytapTemplateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).templatesClient
	ctx := meta.(*SkytapClient).StopContext

	log.Printf("[INFO] preparing arguments for finding the Skytap Template")

	name := d.Get("name").(string)

	templatesResult, err := client.List(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving templates: %s", err)
	}

	templates := filterDataSourceSkytapTemplatesByName(templatesResult.Value, name)

	if len(templates) == 0 {
		return fmt.Errorf("no template found with name %s", name)
	}

	var template skytap.Template

	if len(templates) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] template datasource - multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			template = mostRecentTemplate(templates)
		} else {
			return fmt.Errorf("your query returned more than one result. Please try a more " +
				"specific search criteria, or set `most_recent` attribute to true")
		}
	} else {
		template = templates[0]
	}

	if template.ID == nil {
		return fmt.Errorf("template ID is not set")
	}
	templateID := *template.ID
	d.SetId(templateID)
	d.Set("name", template.Name)

	return nil
}

func filterDataSourceSkytapTemplatesByName(templates []skytap.Template, name string) []skytap.Template {
	var result []skytap.Template
	for _, p := range templates {
		re := regexp.MustCompile(name)
		if re.FindString(*p.Name) != "" {
			result = append(result, p)
		}
	}
	return result
}

func mostRecentTemplate(templates []skytap.Template) skytap.Template {
	sort.Slice(templates, func(i, j int) bool {
		time1, _ := time.Parse(timestampFormat, *templates[i].CreatedAt)
		time2, _ := time.Parse(timestampFormat, *templates[j].CreatedAt)
		return time1.After(time2)
	})
	return templates[0]
}
