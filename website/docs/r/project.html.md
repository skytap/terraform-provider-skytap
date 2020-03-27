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
* `auto_add_role_name` - (Optional) If this field is set to `viewer`, `participant`, `editor`, or `manager`, new users added to your Skytap account are automatically added to this project with the specified project role. Existing users aren’t affected by this setting. If the field is set to `null`, new users aren’t automatically added to the project. For additional details, see [Automatically adding new users to a project](https://help.skytap.com/csh-project-automatic-role.html).
* `show_project_members` - (Optional) Determines whether projects members can view a list of other project members. False by default.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 mins) Used when creating the project
* `update` - (Defaults to 10 mins) Used when updating the project
* `delete` - (Defaults to 10 mins) destroying the project

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the project.
