---
layout: "skytap"
page_title: "Skytap: skytap_project"
sidebar_current: "docs-skytap-datasource-project"
description: |-
  Get information on a project.
---

# skytap_project

Get information on a project. This data source provides the id, name, summary, auto_add_role_name and 
show_project_members properties of a project as configured on your Skytap account.
This is useful in order to retrieve a project's id via its name.

An error is triggered if:
 1. No projects can be retrieved.
 2. The project does not exist.
 3. More than one project matches the name.

## Example Usage

Get the project:

```hcl
data "skytap_project" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of project.

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the project.
* `summary`: The summary description of the project.
* `auto_add_role_name`: The role automatically assigned to every new user added to the project.
* `show_project_members`: Whether project members can view a list of the other project members.
