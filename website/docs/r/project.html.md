---
layout: "skytap"
page_title: "Skytap: skytap_project"
sidebar_current: "docs-skytap-resource-project"
description: |-
  Provides a Skytap Project resource.
---

# skytap\_project

Provides a Skytap Project resource. Projects are an access permissions model used to share environments, 
templates, and assets with other users.

## Example Usage


```hcl
# Create a new project
resource "skytap_project" "project" {
  name = "Terraform Example"
  summary = "Skytap terraform provider example project."
  show_project_members = false
  auto_add_role_name = "participant"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) User-defined project name.
* `summary` - (Optional) User-defined description of project.
* `auto_add_role_name` - (Optional) If set to `viewer`, `participant`, `editor`, or `manager`, Skytap will assign the specified project role to every new user added to the project. The project role roles of existing project members will be unchanged. If auto-add is disabled, this field will be null.
* `show_project_members` - (Optional) Determines whether projects members can view a list of other project members. False by default.

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the project.
